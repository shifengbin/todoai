package main

import "testing"

func TestStartTerminalInjectsClaudeStatusDir(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellClaudeStatusDir("/some/status-dir"),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.EnsureProjectTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("EnsureProjectTerminal: %v", err)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(starter.requests))
	}
	if got := envValueFromList(starter.requests[0].Env, "TODOAI_STATUS_DIR"); got != "/some/status-dir" {
		t.Fatalf("TODOAI_STATUS_DIR = %q, want /some/status-dir", got)
	}
}

func TestStartTerminalOmitsClaudeStatusDirWhenUnset(t *testing.T) {
	t.Setenv("TODOAI_STATUS_DIR", "/tmp/inherited-status-dir")

	starter := newFakeShellStarter()
	manager := NewShellSessionManager(starter.Start, ShellSessionCallbacks{})
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.EnsureProjectTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("EnsureProjectTerminal: %v", err)
	}
	if got := envValueFromList(starter.requests[0].Env, "TODOAI_STATUS_DIR"); got != "" {
		t.Fatalf("TODOAI_STATUS_DIR should be absent, got %q", got)
	}
}
