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

const defaultClaudeStatusDir = "/tmp/claude"

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
}

func NewClaudeStatusWatcher(dir string, terminals func() []ProjectTerminal, emit func(TerminalAgentStatusEvent)) *ClaudeStatusWatcher {
	return &ClaudeStatusWatcher{
		dir:      dir,
		terminal: terminals,
		emit:     emit,
		seen:     map[string]ClaudeStatus{},
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
		current[status.Session] = status
		previous, seen := watcher.seen[status.Session]
		if seen && previous.Status == status.Status && previous.Event == status.Event && previous.UpdatedAt == status.UpdatedAt {
			continue
		}
		terminal, ok := matchClaudeStatusTerminal(status, terminals)
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
		terminal, ok := matchClaudeStatusTerminal(status, terminals)
		if ok {
			watcher.emit(claudeSessionEndEvent(status, terminal))
		}
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

func matchClaudeStatusTerminal(status ClaudeStatus, terminals []ProjectTerminal) (ProjectTerminal, bool) {
	if status.TerminalID != "" {
		for _, terminal := range terminals {
			if terminal.ID == status.TerminalID && terminal.State == ShellStateRunning {
				return terminal, true
			}
		}
	}
	if status.CWD == "" {
		return ProjectTerminal{}, false
	}
	var matched ProjectTerminal
	matches := 0
	for _, terminal := range terminals {
		if terminal.State != ShellStateRunning {
			continue
		}
		if filepath.Clean(terminal.projectPath) == filepath.Clean(status.CWD) {
			matched = terminal
			matches++
		}
	}
	return matched, matches == 1
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
