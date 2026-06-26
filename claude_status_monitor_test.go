package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClaudeStatusFileToAgentStatusEventMapsDocumentedStatuses(t *testing.T) {
	now := time.Unix(1718450010, 0)
	terminal := ProjectTerminal{
		ID:        "terminal-a",
		ProjectID: "project-a",
		ShellName: "zsh",
		State:     ShellStateRunning,
	}

	tests := []struct {
		name       string
		statusJSON string
		wantPhase  string
		wantReason string
		wantLabel  string
	}{
		{
			name:       "prompt submitted",
			statusJSON: `{"session":"a1","status":"busy","event":"UserPromptSubmit","cwd":"/work/a","ts":1718450010}`,
			wantPhase:  "busy",
			wantReason: "user-prompt-submit",
		},
		{
			name:       "tool running",
			statusJSON: `{"session":"a1","status":"tool:Bash","event":"PreToolUse","cwd":"/work/a","tool":"Bash","ts":1718450010}`,
			wantPhase:  "busy",
			wantReason: "pre-tool-use",
			wantLabel:  "Bash",
		},
		{
			name:       "waiting notification",
			statusJSON: `{"session":"a1","status":"waiting","event":"Notification","cwd":"/work/a","ts":1718450010}`,
			wantPhase:  "needs-input",
			wantReason: "notification",
		},
		{
			name:       "stop returns idle",
			statusJSON: `{"session":"a1","status":"idle","event":"Stop","cwd":"/work/a","ts":1718450010}`,
			wantPhase:  "idle",
			wantReason: "stop",
		},
		{
			name:       "post tool use remains busy until stop",
			statusJSON: `{"session":"a1","status":"tool_done","event":"PostToolUse","cwd":"/work/a","tool":"Bash","ts":1718450010}`,
			wantPhase:  "busy",
			wantReason: "post-tool-use",
			wantLabel:  "Bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := parseClaudeStatusFile([]byte(tt.statusJSON), now)
			if err != nil {
				t.Fatalf("parseClaudeStatusFile() error = %v", err)
			}
			event := claudeStatusToAgentStatusEvent(status, terminal)

			if event.TerminalID != "terminal-a" || event.ProjectID != "project-a" {
				t.Fatalf("event identity = %#v, want project-a/terminal-a", event)
			}
			if event.Phase != tt.wantPhase {
				t.Fatalf("Phase = %q, want %q", event.Phase, tt.wantPhase)
			}
			if event.Source != "claude-hook" || event.Confidence != "structured" {
				t.Fatalf("source/confidence = %q/%q, want claude-hook/structured", event.Source, event.Confidence)
			}
			if event.Reason != tt.wantReason {
				t.Fatalf("Reason = %q, want %q", event.Reason, tt.wantReason)
			}
			if event.Label != tt.wantLabel {
				t.Fatalf("Label = %q, want %q", event.Label, tt.wantLabel)
			}
			if event.UpdatedAt != 1718450010000 {
				t.Fatalf("UpdatedAt = %d, want 1718450010000", event.UpdatedAt)
			}
		})
	}
}

func TestClaudeStatusMatchesTerminalByTerminalIDOrCwd(t *testing.T) {
	projectPath := t.TempDir()
	otherPath := t.TempDir()
	terminals := []ProjectTerminal{
		{ID: "terminal-a", ProjectID: "project-a", ShellName: "zsh", State: ShellStateRunning, projectPath: projectPath},
		{ID: "terminal-b", ProjectID: "project-b", ShellName: "zsh", State: ShellStateRunning, projectPath: otherPath},
	}

	matched, ok := matchClaudeStatusTerminal(ClaudeStatus{TerminalID: "terminal-b", CWD: projectPath}, terminals, map[string]string{})
	if !ok || matched.ID != "terminal-b" {
		t.Fatalf("match by terminal ID = %#v/%v, want terminal-b", matched, ok)
	}

	matched, ok = matchClaudeStatusTerminal(ClaudeStatus{CWD: projectPath}, terminals, map[string]string{})
	if !ok || matched.ID != "terminal-a" {
		t.Fatalf("match by cwd = %#v/%v, want terminal-a", matched, ok)
	}

	_, ok = matchClaudeStatusTerminal(ClaudeStatus{CWD: t.TempDir()}, terminals, map[string]string{})
	if ok {
		t.Fatal("match unknown cwd succeeded, want no match")
	}
}

// TestClaudeStatusStickyBindingPinsEachSessionToItsTerminal is the regression
// test for multi-claude status bleed-through: two claudes running in two
// different terminals that happen to share a cwd must each stay pinned to its
// own terminal, instead of both driving whichever terminal the cwd fallback
// happens to resolve to first.
func TestClaudeStatusStickyBindingPinsEachSessionToItsTerminal(t *testing.T) {
	sharedPath := t.TempDir()
	terminals := []ProjectTerminal{
		{ID: "terminal-a", ProjectID: "project-a", State: ShellStateRunning, projectPath: sharedPath},
		{ID: "terminal-b", ProjectID: "project-b", State: ShellStateRunning, projectPath: sharedPath},
	}
	bindings := map[string]string{}

	matched, ok := matchClaudeStatusTerminal(ClaudeStatus{Session: "s1", TerminalID: "terminal-a", CWD: sharedPath}, terminals, bindings)
	if !ok || matched.ID != "terminal-a" {
		t.Fatalf("s1 = %#v/%v, want terminal-a", matched, ok)
	}
	matched, ok = matchClaudeStatusTerminal(ClaudeStatus{Session: "s2", TerminalID: "terminal-b", CWD: sharedPath}, terminals, bindings)
	if !ok || matched.ID != "terminal-b" {
		t.Fatalf("s2 = %#v/%v, want terminal-b", matched, ok)
	}
	if bindings["s1"] != "terminal-a" || bindings["s2"] != "terminal-b" {
		t.Fatalf("bindings = %#v, want s1→terminal-a, s2→terminal-b", bindings)
	}

	// A subsequent update for s1 must still land on terminal-a even though both
	// terminals match its cwd, because the sticky binding wins.
	matched, ok = matchClaudeStatusTerminal(ClaudeStatus{Session: "s1", CWD: sharedPath}, terminals, bindings)
	if !ok || matched.ID != "terminal-a" {
		t.Fatalf("sticky s1 = %#v/%v, want terminal-a", matched, ok)
	}
}

// TestClaudeStatusCwdFallbackDoesNotPileTwoClaudesOnOneTerminal covers the
// external-claude pollution case: when one session already claimed the only
// terminal at a cwd, a second sessionless/external claude at the same cwd must
// NOT also latch onto that terminal (which would overwrite its badge).
func TestClaudeStatusCwdFallbackDoesNotPileTwoClaudesOnOneTerminal(t *testing.T) {
	sharedPath := t.TempDir()
	terminals := []ProjectTerminal{
		{ID: "terminal-a", ProjectID: "project-a", State: ShellStateRunning, projectPath: sharedPath},
	}
	bindings := map[string]string{}

	if _, ok := matchClaudeStatusTerminal(ClaudeStatus{Session: "s1", CWD: sharedPath}, terminals, bindings); !ok {
		t.Fatal("first claude at cwd should match the single terminal")
	}
	if _, ok := matchClaudeStatusTerminal(ClaudeStatus{Session: "s2", CWD: sharedPath}, terminals, bindings); ok {
		t.Fatal("second claude at the same cwd must not pile onto the occupied terminal")
	}
}

// TestClaudeStatusTerminalIdEvictsCwdFallbackSquatter: a real terminalId match
// is authoritative, so it reclaims a terminal that an earlier cwd-fallback
// match (e.g. an external claude without a terminalId) had grabbed.
func TestClaudeStatusTerminalIdEvictsCwdFallbackSquatter(t *testing.T) {
	sharedPath := t.TempDir()
	terminals := []ProjectTerminal{
		{ID: "terminal-a", ProjectID: "project-a", State: ShellStateRunning, projectPath: sharedPath},
	}
	bindings := map[string]string{}

	if _, ok := matchClaudeStatusTerminal(ClaudeStatus{Session: "s1", CWD: sharedPath}, terminals, bindings); !ok {
		t.Fatal("cwd fallback should claim the single terminal")
	}
	if bindings["s1"] != "terminal-a" {
		t.Fatalf("s1 should hold terminal-a, bindings=%#v", bindings)
	}

	// s2 arrives with the authoritative terminalId for terminal-a and must
	// evict s1's cwd-fallback claim.
	matched, ok := matchClaudeStatusTerminal(ClaudeStatus{Session: "s2", TerminalID: "terminal-a", CWD: sharedPath}, terminals, bindings)
	if !ok || matched.ID != "terminal-a" {
		t.Fatalf("s2 = %#v/%v, want terminal-a", matched, ok)
	}
	if _, squatter := bindings["s1"]; squatter {
		t.Fatalf("s1 should have been evicted, bindings=%#v", bindings)
	}
	if bindings["s2"] != "terminal-a" {
		t.Fatalf("s2 should now hold terminal-a, bindings=%#v", bindings)
	}
}

func TestClaudeStatusWatcherEmitsUpdatesAndExitsOnDelete(t *testing.T) {
	dir := t.TempDir()
	events := make(chan TerminalAgentStatusEvent, 4)
	watcher := NewClaudeStatusWatcher(dir, func() []ProjectTerminal {
		return []ProjectTerminal{
			{
				ID:          "terminal-a",
				ProjectID:   "project-a",
				ShellName:   "zsh",
				State:       ShellStateRunning,
				projectPath: "/work/a",
			},
		}
	}, func(event TerminalAgentStatusEvent) {
		events <- event
	})

	if err := os.WriteFile(filepath.Join(dir, "a1.status"), []byte(`{"session":"a1","status":"busy","event":"UserPromptSubmit","cwd":"/work/a","ts":1718450010}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	watcher.Poll()

	select {
	case event := <-events:
		if event.TerminalID != "terminal-a" || event.Phase != "busy" {
			t.Fatalf("event = %#v, want terminal-a busy", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for busy event")
	}

	if err := os.Remove(filepath.Join(dir, "a1.status")); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	watcher.Poll()

	select {
	case event := <-events:
		if event.TerminalID != "terminal-a" || event.Phase != "exited" || event.Reason != "session-end" {
			t.Fatalf("delete event = %#v, want terminal-a exited/session-end", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for delete event")
	}
}

func TestClaudeStatusWatcherAgesOutStaleBusyStatus(t *testing.T) {
	dir := t.TempDir()
	events := make(chan TerminalAgentStatusEvent, 4)
	watcher := NewClaudeStatusWatcher(dir, func() []ProjectTerminal {
		return []ProjectTerminal{
			{
				ID:          "terminal-a",
				ProjectID:   "project-a",
				ShellName:   "zsh",
				State:       ShellStateRunning,
				projectPath: "/work/a",
			},
		}
	}, func(event TerminalAgentStatusEvent) {
		events <- event
	})

	stale := time.Now().Add(-2 * claudeStatusStaleThreshold)
	path := filepath.Join(dir, "a1.status")
	if err := os.WriteFile(path, []byte(`{"session":"a1","status":"busy","event":"UserPromptSubmit","cwd":"/work/a","ts":1718450010}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	watcher.Poll()

	select {
	case event := <-events:
		if event.TerminalID != "terminal-a" || event.Phase != "idle" || event.Reason != "stale-timeout" {
			t.Fatalf("stale event = %#v, want terminal-a idle/stale-timeout", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stale-timeout idle event")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stale .status should be removed, err=%v", err)
	}
}

func TestClaudeStatusWatcherDoesNotAgeOutIdleStatus(t *testing.T) {
	dir := t.TempDir()
	events := make(chan TerminalAgentStatusEvent, 4)
	watcher := NewClaudeStatusWatcher(dir, func() []ProjectTerminal {
		return []ProjectTerminal{
			{
				ID:          "terminal-a",
				ProjectID:   "project-a",
				ShellName:   "zsh",
				State:       ShellStateRunning,
				projectPath: "/work/a",
			},
		}
	}, func(event TerminalAgentStatusEvent) {
		events <- event
	})

	stale := time.Now().Add(-2 * claudeStatusStaleThreshold)
	path := filepath.Join(dir, "a1.status")
	if err := os.WriteFile(path, []byte(`{"session":"a1","status":"idle","event":"Stop","cwd":"/work/a","ts":1718450010}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	watcher.Poll()

	select {
	case event := <-events:
		if event.Phase != "idle" || event.Reason == "stale-timeout" {
			t.Fatalf("idle event = %#v, want idle without stale-timeout", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for normal idle event")
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("idle .status should remain, err=%v", err)
	}

	// Second poll: status unchanged → seen dedup suppresses further events.
	watcher.Poll()
	select {
	case event := <-events:
		t.Fatalf("unexpected second event = %#v", event)
	case <-time.After(100 * time.Millisecond):
	}
}
