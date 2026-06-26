package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// claudeStatusStaleThreshold is how long a non-idle .status file may sit
// without being refreshed before the watcher treats its owner as a zombie
// (e.g. claude killed mid-turn) and emits idle + removes the stale file. An
// active claude refreshes mtime on every tool call, so long-running turns are
// never mistaken for zombies.
const claudeStatusStaleThreshold = 15 * time.Minute

type ClaudeStatus struct {
	Session    string `json:"session"`
	Status     string `json:"status"`
	Event      string `json:"event"`
	CWD        string `json:"cwd"`
	Tool       string `json:"tool"`
	Transcript string `json:"transcript"`
	TerminalID string `json:"terminalId"`
	TS         int64  `json:"ts"`
	UpdatedAt  int64
}

type ClaudeStatusWatcher struct {
	dir      string
	terminal func() []ProjectTerminal
	emit     func(TerminalAgentStatusEvent)
	seen     map[string]ClaudeStatus
	// sessionTerminal pins each claude session to a terminal across polls so
	// that two claudes sharing a cwd don't both drive the same terminal's
	// status badge. See matchClaudeStatusTerminal for the resolution rules.
	sessionTerminal map[string]string
}

func NewClaudeStatusWatcher(dir string, terminals func() []ProjectTerminal, emit func(TerminalAgentStatusEvent)) *ClaudeStatusWatcher {
	return &ClaudeStatusWatcher{
		dir:             dir,
		terminal:        terminals,
		emit:            emit,
		seen:            map[string]ClaudeStatus{},
		sessionTerminal: map[string]string{},
	}
}

func (watcher *ClaudeStatusWatcher) Poll() {
	if watcher == nil || watcher.emit == nil || watcher.terminal == nil {
		return
	}
	entries, err := os.ReadDir(watcher.dir)
	if err != nil {
		return
	}

	current := map[string]ClaudeStatus{}
	terminals := watcher.terminal()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".status") {
			continue
		}
		path := filepath.Join(watcher.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		status, err := parseClaudeStatusFile(data, time.Now())
		if err != nil || status.Session == "" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if isStaleStatus(status.Status) && time.Since(info.ModTime()) > claudeStatusStaleThreshold {
			terminal, ok := matchClaudeStatusTerminal(status, terminals, watcher.sessionTerminal)
			if ok {
				watcher.emit(claudeStaleIdleEvent(status, terminal))
			}
			_ = os.Remove(path)
			delete(watcher.seen, status.Session)
			delete(watcher.sessionTerminal, status.Session)
			continue
		}
		current[status.Session] = status
		previous, seen := watcher.seen[status.Session]
		if seen && previous.Status == status.Status && previous.Event == status.Event && previous.UpdatedAt == status.UpdatedAt {
			continue
		}
		terminal, ok := matchClaudeStatusTerminal(status, terminals, watcher.sessionTerminal)
		if !ok {
			continue
		}
		watcher.emit(claudeStatusToAgentStatusEvent(status, terminal))
	}

	deletedSessions := make([]string, 0)
	for session := range watcher.seen {
		if _, ok := current[session]; !ok {
			deletedSessions = append(deletedSessions, session)
		}
	}
	sort.Strings(deletedSessions)
	for _, session := range deletedSessions {
		status := watcher.seen[session]
		terminal, ok := matchClaudeStatusTerminal(status, terminals, watcher.sessionTerminal)
		if ok {
			watcher.emit(claudeSessionEndEvent(status, terminal))
		}
		delete(watcher.sessionTerminal, session)
	}
	watcher.seen = current
}

func parseClaudeStatusFile(data []byte, fallback time.Time) (ClaudeStatus, error) {
	var status ClaudeStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return ClaudeStatus{}, err
	}
	var aliases struct {
		SessionID  string `json:"session_id"`
		Terminal   string `json:"terminal"`
		TerminalID string `json:"terminal_id"`
	}
	_ = json.Unmarshal(data, &aliases)
	if status.Session == "" {
		status.Session = aliases.SessionID
	}
	if status.TerminalID == "" {
		if aliases.TerminalID != "" {
			status.TerminalID = aliases.TerminalID
		} else {
			status.TerminalID = aliases.Terminal
		}
	}
	if status.Session == "" {
		return ClaudeStatus{}, errors.New("missing claude session")
	}
	if status.UpdatedAt == 0 {
		if status.TS > 0 {
			status.UpdatedAt = status.TS * 1000
		} else if !fallback.IsZero() {
			status.UpdatedAt = fallback.UnixMilli()
		}
	}
	return status, nil
}

// matchClaudeStatusTerminal binds a Claude session to a terminal. bindings is a
// session→terminalID map carried across polls: once a session is bound it stays
// bound while that terminal keeps running, so each claude's status stays pinned
// to its own terminal instead of being re-guessed by cwd on every poll.
//
// Resolution order:
//  1. Sticky: the session is already bound and that terminal is still running.
//  2. TerminalID: exact id match. This is authoritative — a hook-supplied
//     terminalId means the claude really runs in that terminal, so any other
//     session that grabbed it via the cwd fallback is evicted.
//  3. Cwd fallback: the single running terminal at that cwd, excluding
//     terminals already claimed by another session. An ambiguous or contested
//     cwd resolves to no match rather than piling two claudes onto one
//     terminal (which is what caused multi-claude status bleed-through).
//
// bindings is mutated as a side effect to record (or evict) the chosen binding.
func matchClaudeStatusTerminal(status ClaudeStatus, terminals []ProjectTerminal, bindings map[string]string) (ProjectTerminal, bool) {
	session := status.Session

	// 1. Sticky binding from a previous poll.
	if session != "" && bindings != nil {
		if boundID, ok := bindings[session]; ok {
			for _, terminal := range terminals {
				if terminal.ID == boundID && terminal.State == ShellStateRunning {
					return terminal, true
				}
			}
			// Bound terminal no longer running/present — drop and re-resolve.
			delete(bindings, session)
		}
	}

	// 2. Authoritative terminalId match.
	if status.TerminalID != "" {
		for _, terminal := range terminals {
			if terminal.ID == status.TerminalID && terminal.State == ShellStateRunning {
				if bindings != nil {
					// Evict any other session that claimed this terminal via the
					// cwd fallback; the terminalId is the ground truth.
					for otherSession, boundID := range bindings {
						if boundID == terminal.ID && otherSession != session {
							delete(bindings, otherSession)
						}
					}
					if session != "" {
						bindings[session] = terminal.ID
					}
				}
				return terminal, true
			}
		}
	}

	// 3. Cwd fallback, excluding terminals already claimed by another session.
	if status.CWD == "" {
		return ProjectTerminal{}, false
	}
	occupied := make(map[string]bool, len(bindings))
	if bindings != nil {
		for otherSession, boundID := range bindings {
			if otherSession != session {
				occupied[boundID] = true
			}
		}
	}
	target := filepath.Clean(status.CWD)
	var matched ProjectTerminal
	matches := 0
	for _, terminal := range terminals {
		if terminal.State != ShellStateRunning || occupied[terminal.ID] {
			continue
		}
		if filepath.Clean(terminal.projectPath) == target {
			matched = terminal
			matches++
		}
	}
	if matches == 1 {
		if bindings != nil && session != "" {
			bindings[session] = matched.ID
		}
		return matched, true
	}
	return ProjectTerminal{}, false
}

func claudeStatusToAgentStatusEvent(status ClaudeStatus, terminal ProjectTerminal) TerminalAgentStatusEvent {
	phase, reason := claudePhaseAndReason(status)
	return TerminalAgentStatusEvent{
		ProjectID:     terminal.ProjectID,
		TodoID:        terminal.TodoID,
		TodoProjectID: terminal.TodoProjectID,
		TerminalID:    terminal.ID,
		Phase:         phase,
		Source:        "claude-hook",
		Confidence:    "structured",
		Reason:        reason,
		Label:         status.Tool,
		UpdatedAt:     status.UpdatedAt,
	}
}

func claudeSessionEndEvent(status ClaudeStatus, terminal ProjectTerminal) TerminalAgentStatusEvent {
	return TerminalAgentStatusEvent{
		ProjectID:     terminal.ProjectID,
		TodoID:        terminal.TodoID,
		TodoProjectID: terminal.TodoProjectID,
		TerminalID:    terminal.ID,
		Phase:         "exited",
		Source:        "claude-hook",
		Confidence:    "structured",
		Reason:        "session-end",
		UpdatedAt:     time.Now().UnixMilli(),
	}
}

// isStaleStatus reports whether a status token represents an active (non-idle)
// phase that a zombie file could be stuck on. idle is excluded because a long
// idle .status is the normal steady state of a finished-but-open claude.
func isStaleStatus(status string) bool {
	return status != "" && status != "idle"
}

// claudeStaleIdleEvent emits an idle transition for a .status file that exceeded
// the stale threshold, signalling the terminal is no longer busy.
func claudeStaleIdleEvent(status ClaudeStatus, terminal ProjectTerminal) TerminalAgentStatusEvent {
	return TerminalAgentStatusEvent{
		ProjectID:     terminal.ProjectID,
		TodoID:        terminal.TodoID,
		TodoProjectID: terminal.TodoProjectID,
		TerminalID:    terminal.ID,
		Phase:         "idle",
		Source:        "claude-hook",
		Confidence:    "structured",
		Reason:        "stale-timeout",
		Label:         status.Tool,
		UpdatedAt:     time.Now().UnixMilli(),
	}
}

func claudePhaseAndReason(status ClaudeStatus) (string, string) {
	event := normalizeStatusToken(status.Event)
	switch normalizeStatusToken(status.Status) {
	case "busy":
		return "busy", eventOrStatusReason(event, "user-prompt-submit")
	case "waiting":
		return "needs-input", eventOrStatusReason(event, "notification")
	case "idle":
		return "idle", eventOrStatusReason(event, "stop")
	case "tool-done":
		return "busy", eventOrStatusReason(event, "post-tool-use")
	}
	if strings.HasPrefix(status.Status, "tool:") {
		return "busy", eventOrStatusReason(event, "pre-tool-use")
	}
	return "busy", eventOrStatusReason(event, normalizeStatusToken(status.Status))
}

func eventOrStatusReason(event string, fallback string) string {
	if event != "" {
		return event
	}
	return fallback
}

func normalizeStatusToken(value string) string {
	camelBoundary := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	withBoundaries := camelBoundary.ReplaceAllString(value, `${1}-${2}`)
	return strings.Trim(strings.ToLower(strings.ReplaceAll(withBoundaries, "_", "-")), " ")
}
