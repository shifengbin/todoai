package main

import (
	"os"
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

func TestAppUsesSavedTerminalShellForNewTerminals(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()
	shellPath := executableFile(t, "zsh")
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(configDir, "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	if _, err := app.SelectProject(state.Projects[0].ID); err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}

	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != shellPath {
		t.Fatalf("ShellPath = %q, want %q", starter.requests[0].ShellPath, shellPath)
	}
}

func TestAppKeepsExistingTerminalShellAfterSettingChanges(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()
	shellA := executableFile(t, "bash")
	shellB := executableFile(t, "zsh")
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(configDir, "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1", "terminal-2")),
	)
	if _, err := app.SaveTerminalShell(shellA, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell(shellA) error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	if _, err := app.SelectProject(projectID); err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}

	if _, err := app.SaveTerminalShell(shellB, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell(shellB) error = %v", err)
	}
	if _, err := app.CreateTerminal(projectID, 80, 24); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if len(starter.requests) != 2 {
		t.Fatalf("shell start count = %d, want 2", len(starter.requests))
	}
	if starter.requests[0].ShellPath != shellA {
		t.Fatalf("first ShellPath = %q, want %q", starter.requests[0].ShellPath, shellA)
	}
	if starter.requests[1].ShellPath != shellB {
		t.Fatalf("second ShellPath = %q, want %q", starter.requests[1].ShellPath, shellB)
	}
}

func TestAppFallsBackWhenSavedTerminalShellIsUnavailable(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()
	fallbackShell := executableFile(t, "sh")
	t.Setenv("SHELL", fallbackShell)
	writeSettingsFile(t, filepath.Join(configDir, "settings.json"), filepath.Join(t.TempDir(), "missing-zsh"), "manual")
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(configDir, "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	if _, err := app.SelectProject(state.Projects[0].ID); err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}

	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != fallbackShell {
		t.Fatalf("ShellPath = %q, want fallback %q", starter.requests[0].ShellPath, fallbackShell)
	}
}

func TestAppDeletesProjectAndOwnedTerminals(t *testing.T) {
	projectDirA := t.TempDir()
	projectDirB := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a1", "terminal-a2", "terminal-b1")),
	)

	state, err := app.AddProjectFromPath(projectDirA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectAID := state.Projects[0].ID
	if _, err := app.SelectProject(projectAID); err != nil {
		t.Fatalf("SelectProject(A) error = %v", err)
	}
	if _, err := app.CreateTerminal(projectAID, 80, 24); err != nil {
		t.Fatalf("CreateTerminal(A) error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectDirB)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	projectBID := state.ActiveProjectID
	if _, err := app.SelectProject(projectBID); err != nil {
		t.Fatalf("SelectProject(B) error = %v", err)
	}

	state, err = app.DeleteProject(projectAID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].ID != projectBID {
		t.Fatalf("Projects = %#v, want only project B", state.Projects)
	}
	if state.ActiveProjectID != projectBID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, projectBID)
	}
	if state.ActiveTerminalID != "terminal-b1" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-b1", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 1 || state.Terminals[0].ID != "terminal-b1" {
		t.Fatalf("Terminals = %#v, want only terminal-b1", state.Terminals)
	}
	if !starter.processes[0].closed || !starter.processes[1].closed {
		t.Fatal("deleted project terminal processes were not closed")
	}
	if starter.processes[2].closed {
		t.Fatal("remaining project terminal process was closed")
	}
}

func TestAppDeletesTerminalAndReturnsUpdatedState(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	if _, err := app.SelectProject(projectID); err != nil {
		t.Fatalf("SelectProject() error = %v", err)
	}
	state, err = app.CreateTerminal(projectID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-b" {
		t.Fatalf("ActiveTerminalID setup = %q, want terminal-b", state.ActiveTerminalID)
	}

	state, err = app.DeleteTerminal("terminal-b")
	if err != nil {
		t.Fatalf("DeleteTerminal(terminal-b) error = %v", err)
	}

	if state.ActiveTerminalID != "terminal-a" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-a", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 1 || state.Terminals[0].ID != "terminal-a" {
		t.Fatalf("Terminals = %#v, want only terminal-a", state.Terminals)
	}
	if !starter.processes[1].closed {
		t.Fatal("deleted terminal process was not closed")
	}

	state, err = app.DeleteTerminal("terminal-a")
	if err != nil {
		t.Fatalf("DeleteTerminal(terminal-a) error = %v", err)
	}
	if state.ActiveTerminalID != "" {
		t.Fatalf("ActiveTerminalID after deleting last terminal = %q, want empty", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 0 {
		t.Fatalf("Terminals length after deleting last terminal = %d, want 0", len(state.Terminals))
	}
	if len(starter.requests) != 2 {
		t.Fatalf("start count = %d, want no replacement terminal", len(starter.requests))
	}
}

func TestAppGetsProjectGitStatusForAvailableProject(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	app.gitStatus = func(path string) (GitStatus, error) {
		if path != projectDir {
			t.Fatalf("git status path = %q, want %q", path, projectDir)
		}
		return GitStatus{IsRepo: true, Branch: "main", ChangedCount: 3}, nil
	}

	status, err := app.GetProjectGitStatus(projectID)
	if err != nil {
		t.Fatalf("GetProjectGitStatus() error = %v", err)
	}

	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want %q", status.ProjectID, projectID)
	}
	if !status.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if status.Branch != "main" {
		t.Fatalf("Branch = %q, want main", status.Branch)
	}
	if status.ChangedCount != 3 {
		t.Fatalf("ChangedCount = %d, want 3", status.ChangedCount)
	}
}

func TestAppGetsProjectGitStatusWithoutQueryingUnavailableProject(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("RemoveAll(projectDir) error = %v", err)
	}
	gitStatusCalls := 0
	app.gitStatus = func(path string) (GitStatus, error) {
		gitStatusCalls++
		return GitStatus{}, nil
	}

	status, err := app.GetProjectGitStatus(projectID)
	if err != nil {
		t.Fatalf("GetProjectGitStatus() error = %v", err)
	}

	if gitStatusCalls != 0 {
		t.Fatalf("git status calls = %d, want 0", gitStatusCalls)
	}
	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want %q", status.ProjectID, projectID)
	}
	if !status.PathUnavailable {
		t.Fatal("PathUnavailable = false, want true")
	}
}

func TestAppGetProjectGitStatusReturnsErrorWhenProjectIsMissing(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	if _, err := app.GetProjectGitStatus("missing-project"); err == nil {
		t.Fatal("GetProjectGitStatus() error = nil, want error")
	}
}
