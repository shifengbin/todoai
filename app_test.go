package main

import (
	"path/filepath"
	"testing"
)

func TestAppAddsAndSelectsProjectsThroughPublicAPI(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellPathResolver(func() string { return "/bin/zsh" }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	if len(state.Projects) != 1 {
		t.Fatalf("Projects length = %d, want 1", len(state.Projects))
	}
	if state.ActiveProjectID != state.Projects[0].ID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, state.Projects[0].ID)
	}
	if len(state.Terminals) != 0 {
		t.Fatalf("Terminals length after add = %d, want 0 before project selection", len(state.Terminals))
	}

	state, err = app.SelectProject(state.Projects[0].ID)
	if err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}
	if state.ActiveProjectID != state.Projects[0].ID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, state.Projects[0].ID)
	}
	if state.ActiveTerminalID != "terminal-1" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-1", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 1 {
		t.Fatalf("Terminals length = %d, want 1", len(state.Terminals))
	}
	if state.Terminals[0].ProjectID != state.Projects[0].ID {
		t.Fatalf("Terminal ProjectID = %q, want %q", state.Terminals[0].ProjectID, state.Projects[0].ID)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].TerminalID != "terminal-1" {
		t.Fatalf("TerminalID = %q, want terminal-1", starter.requests[0].TerminalID)
	}
}

func TestAppCreatesAndSelectsAdditionalProjectTerminals(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1", "terminal-2")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.SelectProject(projectID)
	if err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}

	state, err = app.CreateTerminal(projectID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if state.ActiveProjectID != projectID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, projectID)
	}
	if state.ActiveTerminalID != "terminal-2" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-2", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 2 {
		t.Fatalf("Terminals length = %d, want 2", len(state.Terminals))
	}
	if len(starter.requests) != 2 {
		t.Fatalf("shell start count = %d, want 2", len(starter.requests))
	}

	state, err = app.SelectTerminal("terminal-1")
	if err != nil {
		t.Fatalf("SelectTerminal() error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-1" {
		t.Fatalf("ActiveTerminalID after SelectTerminal = %q, want terminal-1", state.ActiveTerminalID)
	}
	if state.ActiveProjectID != projectID {
		t.Fatalf("ActiveProjectID after SelectTerminal = %q, want %q", state.ActiveProjectID, projectID)
	}
}
