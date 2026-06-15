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

	matched, ok := matchClaudeStatusTerminal(ClaudeStatus{TerminalID: "terminal-b", CWD: projectPath}, terminals)
	if !ok || matched.ID != "terminal-b" {
		t.Fatalf("match by terminal ID = %#v/%v, want terminal-b", matched, ok)
	}

	matched, ok = matchClaudeStatusTerminal(ClaudeStatus{CWD: projectPath}, terminals)
	if !ok || matched.ID != "terminal-a" {
		t.Fatalf("match by cwd = %#v/%v, want terminal-a", matched, ok)
	}

	_, ok = matchClaudeStatusTerminal(ClaudeStatus{CWD: t.TempDir()}, terminals)
	if ok {
		t.Fatal("match unknown cwd succeeded, want no match")
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
