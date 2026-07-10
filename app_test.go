package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAppAddsAndSelectsProjectsThroughPublicAPI(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellPathResolver(func() string { return "/bin/zsh" }),
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
	if state.ActiveTerminalID != "" {
		t.Fatalf("ActiveTerminalID = %q, want empty for project-library selection", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 0 {
		t.Fatalf("Terminals length = %d, want 0 after project-library selection", len(state.Terminals))
	}
	if len(starter.requests) != 0 {
		t.Fatalf("shell start count = %d, want 0 after project-library selection", len(starter.requests))
	}
}

func TestAppWorkspaceOpenScopesProjectsTodosAndHistoryButKeepsSettingsGlobal(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()
	shellPath := executableFile(t, "zsh-global")
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.OpenWorkspaceFromPath(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(A) error = %v", err)
	}
	if state.CurrentWorkspace == nil || state.CurrentWorkspace.Path != mustAbs(t, workspaceA) {
		t.Fatalf("CurrentWorkspace = %#v, want workspace A", state.CurrentWorkspace)
	}
	state, err = app.AddProjectFromPath(projectA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	if _, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{state.Projects[0].ID}}); err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell(global) error = %v", err)
	}

	state, err = app.OpenWorkspaceFromPath(workspaceB)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(B) error = %v", err)
	}
	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectA) || len(state.Todos) != 0 {
		t.Fatalf("workspace B initial state = projects %#v todos %#v, want shared project A and empty todos", state.Projects, state.Todos)
	}
	state, err = app.AddProjectFromPath(projectB)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	if _, err := app.CreateTodo(CreateTodoRequest{Title: "升级依赖", ProjectIDs: []string{state.ActiveProjectID}}); err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}

	state, err = app.OpenWorkspaceFromPath(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(A again) error = %v", err)
	}
	if len(state.Projects) != 2 || !containsProjectPath(state.Projects, mustAbs(t, projectA)) || !containsProjectPath(state.Projects, mustAbs(t, projectB)) {
		t.Fatalf("workspace A projects = %#v, want shared project candidates", state.Projects)
	}
	if len(state.Todos) != 1 || state.Todos[0].Title != "修复登录问题" {
		t.Fatalf("workspace A todos = %#v, want A todo only", state.Todos)
	}
	settings, err := app.LoadTerminalSettings()
	if err != nil {
		t.Fatalf("LoadTerminalSettings(A) error = %v", err)
	}
	if settings.Selected.Path != shellPath {
		t.Fatalf("workspace A shell = %q, want global %q", settings.Selected.Path, shellPath)
	}

	state, err = app.OpenWorkspaceFromPath(workspaceB)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(B again) error = %v", err)
	}
	if len(state.Projects) != 2 || !containsProjectPath(state.Projects, mustAbs(t, projectA)) || !containsProjectPath(state.Projects, mustAbs(t, projectB)) {
		t.Fatalf("workspace B projects = %#v, want shared project candidates", state.Projects)
	}
	if len(state.Todos) != 1 || state.Todos[0].Title != "升级依赖" {
		t.Fatalf("workspace B todos = %#v, want B todo only", state.Todos)
	}
	settings, err = app.LoadTerminalSettings()
	if err != nil {
		t.Fatalf("LoadTerminalSettings(B) error = %v", err)
	}
	if settings.Selected.Path != shellPath {
		t.Fatalf("workspace B shell = %q, want global %q", settings.Selected.Path, shellPath)
	}
}

func TestAppSharesGlobalProjectCandidatesAcrossWorkspacesButScopesTodos(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	projectA := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.OpenWorkspaceFromPath(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(A) error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectID := state.Projects[0].ID
	if _, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{projectID}}); err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}

	state, err = app.OpenWorkspaceFromPath(workspaceB)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(B) error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectA) {
		t.Fatalf("workspace B candidates = %#v, want shared project A candidate", state.Projects)
	}
	if len(state.Todos) != 0 || len(state.TodoProjects) != 0 {
		t.Fatalf("workspace B TODO state = todos %#v todoProjects %#v, want isolated empty", state.Todos, state.TodoProjects)
	}
}

func TestAppRequiresWorkspaceForWorkspaceScopedOperations(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
	)

	state, err := app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if state.CurrentWorkspace != nil {
		t.Fatalf("CurrentWorkspace = %#v, want nil", state.CurrentWorkspace)
	}
	if len(state.Projects) != 0 || len(state.Todos) != 0 || len(state.Terminals) != 0 {
		t.Fatalf("empty workspace state = %#v, want no projects/todos/terminals", state)
	}

	if _, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"}); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("CreateTodo() error = %v, want ErrWorkspaceRequired", err)
	}
	if _, err := app.AddProjectFromPath(t.TempDir()); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("AddProjectFromPath() error = %v, want ErrWorkspaceRequired", err)
	}
	if _, err := app.LoadTerminalSettings(); err != nil {
		t.Fatalf("LoadTerminalSettings() error = %v, want global settings available without workspace", err)
	}
	if _, err := app.GetProjectGitStatus("missing"); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("GetProjectGitStatus() error = %v, want ErrWorkspaceRequired", err)
	}
}

func TestAppSavesTerminalSettingsWithoutWorkspace(t *testing.T) {
	shellPath := executableFile(t, "zsh-global")
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
	)

	state, err := app.SaveTerminalShell(shellPath, ShellSourceManual)
	if err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
}

func TestAppStartupRestoresMostRecentWorkspace(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	state, err := app.OpenWorkspaceFromPath(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(A) error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectAID := state.Projects[0].ID

	state, err = app.OpenWorkspaceFromPath(workspaceB)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath(B) error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectB)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	projectBID := state.ActiveProjectID

	restarted := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	restarted.startup(nil)
	defer restarted.shutdown(nil)

	state, err = restarted.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if state.CurrentWorkspace == nil || state.CurrentWorkspace.Path != mustAbs(t, workspaceB) {
		t.Fatalf("CurrentWorkspace = %#v, want most recent workspace B", state.CurrentWorkspace)
	}
	if len(state.Projects) != 2 || !containsProjectPath(state.Projects, mustAbs(t, projectA)) || !containsProjectPath(state.Projects, mustAbs(t, projectB)) {
		t.Fatalf("restored projects = %#v, want shared candidates", state.Projects)
	}
	if projectAID == projectBID {
		t.Fatal("test setup produced identical project IDs")
	}
}

func TestAppStartupKeepsNoWorkspaceWhenMostRecentWorkspaceUnavailable(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	projectA := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceA); err != nil {
		t.Fatalf("OpenWorkspaceFromPath(A) error = %v", err)
	}
	if _, err := app.AddProjectFromPath(projectA); err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	if _, err := app.OpenWorkspaceFromPath(workspaceB); err != nil {
		t.Fatalf("OpenWorkspaceFromPath(B) error = %v", err)
	}
	if err := os.RemoveAll(workspaceB); err != nil {
		t.Fatalf("RemoveAll(workspaceB) error = %v", err)
	}

	restarted := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	restarted.startup(nil)
	defer restarted.shutdown(nil)

	state, err := restarted.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if state.CurrentWorkspace != nil {
		t.Fatalf("CurrentWorkspace = %#v, want nil when most recent workspace is unavailable", state.CurrentWorkspace)
	}
	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectA) {
		t.Fatalf("Projects = %#v, want global candidates without fallback workspace", state.Projects)
	}
	if len(state.RecentWorkspaces) != 2 {
		t.Fatalf("RecentWorkspaces = %#v, want both recent workspaces retained", state.RecentWorkspaces)
	}
	if state.RecentWorkspaces[0].Path != mustAbs(t, workspaceB) || state.RecentWorkspaces[0].Available {
		t.Fatalf("most recent workspace = %#v, want unavailable workspace B retained first", state.RecentWorkspaces[0])
	}
}

func TestAppOpenWorkspaceFailureKeepsPreviousWorkspaceState(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	missingWorkspace := filepath.Join(t.TempDir(), "missing")

	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID

	state, err = app.OpenWorkspaceFromPath(missingWorkspace)
	if err == nil {
		t.Fatal("OpenWorkspaceFromPath(missing) error = nil, want error")
	}
	if state.CurrentWorkspace == nil || state.CurrentWorkspace.Path != mustAbs(t, workspaceDir) {
		t.Fatalf("CurrentWorkspace after failed open = %#v, want previous workspace", state.CurrentWorkspace)
	}
	if len(state.Projects) != 1 || state.Projects[0].ID != projectID {
		t.Fatalf("Projects after failed open = %#v, want previous project", state.Projects)
	}
}

func TestAppCloseWorkspaceClearsRuntimeStateAndPreservesData(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	state, err := app.OpenWorkspaceFromPath(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	dataPath := state.CurrentWorkspace.DataPath
	state, err = app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	state, err = app.CloseWorkspace()
	if err != nil {
		t.Fatalf("CloseWorkspace() error = %v", err)
	}

	if state.CurrentWorkspace != nil {
		t.Fatalf("CurrentWorkspace = %#v, want nil", state.CurrentWorkspace)
	}
	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectDir) || len(state.Todos) != 0 || len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("closed workspace state = %#v, want global candidates and empty todo/terminal state", state)
	}
	if !starter.processes[0].closed {
		t.Fatal("workspace terminal process was not closed")
	}
	if _, err := os.Stat(filepath.Join(dataPath, "projects.json")); err != nil {
		t.Fatalf("workspace projects data missing after close: %v", err)
	}
	workspaceState, err := app.WorkspaceState()
	if err != nil {
		t.Fatalf("WorkspaceState() error = %v", err)
	}
	if len(workspaceState.RecentWorkspaces) != 1 || workspaceState.RecentWorkspaces[0].Path != mustAbs(t, workspaceDir) {
		t.Fatalf("RecentWorkspaces = %#v, want closed workspace retained", workspaceState.RecentWorkspaces)
	}
}

func TestAppCreatesAndSelectsTodoProjectTerminals(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1", "terminal-2")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	state, err = app.AddProjectToTodo(todoID, projectID)
	if err != nil {
		t.Fatalf("AddProjectToTodo() error = %v", err)
	}
	todoProjectID := state.ActiveTodoProjectID
	if todoProjectID == "" {
		t.Fatal("ActiveTodoProjectID = empty, want associated todo project")
	}

	if _, err := app.CreateTodoTerminal(todoProjectID, 100, 32); err == nil {
		t.Fatal("CreateTodoTerminal(not-started) error = nil, want invalid status error")
	}
	if len(starter.requests) != 0 {
		t.Fatalf("shell start count after rejected terminal = %d, want 0", len(starter.requests))
	}
	state, err = app.ChangeTodoStatus(todoID, "in-progress")
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	state, err = app.CreateTodoTerminal(todoProjectID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	if state.ActiveProjectID != projectID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, projectID)
	}
	if state.ActiveTodoID != todoID {
		t.Fatalf("ActiveTodoID = %q, want %q", state.ActiveTodoID, todoID)
	}
	if state.ActiveTerminalID != "terminal-1" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-1", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 1 {
		t.Fatalf("Terminals length = %d, want 1", len(state.Terminals))
	}
	if state.Terminals[0].TodoID != todoID || state.Terminals[0].TodoProjectID != todoProjectID {
		t.Fatalf("terminal TODO context = %q/%q, want %q/%q", state.Terminals[0].TodoID, state.Terminals[0].TodoProjectID, todoID, todoProjectID)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].TodoID != todoID || starter.requests[0].TodoProjectID != todoProjectID {
		t.Fatalf("shell request TODO context = %q/%q, want %q/%q", starter.requests[0].TodoID, starter.requests[0].TodoProjectID, todoID, todoProjectID)
	}

	state, err = app.CreateTodoTerminal(todoProjectID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTodoTerminal(second) error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-2" {
		t.Fatalf("ActiveTerminalID after second terminal = %q, want terminal-2", state.ActiveTerminalID)
	}
	state, err = app.SelectTerminal("terminal-1")
	if err != nil {
		t.Fatalf("SelectTerminal() error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-1" {
		t.Fatalf("ActiveTerminalID after SelectTerminal = %q, want terminal-1", state.ActiveTerminalID)
	}
}

func TestAppRecordsClearedWorktreeAndCreatesTodoProjectTerminalInSourceProject(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	todoProject := findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
		t.Fatalf("todoProject worktree = %#v, want ready worktree", todoProject)
	}
	removedWorktreePath := todoProject.WorktreePath
	if err := os.RemoveAll(removedWorktreePath); err != nil {
		t.Fatalf("RemoveAll(worktree) error = %v", err)
	}

	state, err = app.CreateTodoTerminal(todoProjectID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].WorkingDir != mustAbs(t, projectDir) {
		t.Fatalf("todo project terminal cwd = %q, want source project %q", starter.requests[0].WorkingDir, mustAbs(t, projectDir))
	}
	if _, err := os.Stat(removedWorktreePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed worktree path stat error = %v, want path to stay removed", err)
	}
	todoProject = findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusCleared {
		t.Fatalf("WorktreeStatus = %q, want %q", todoProject.WorktreeStatus, WorktreeStatusCleared)
	}
}

func TestAppListProjectsRepairsFalseClearedWorktreeWhenPathAndBranchStillExist(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	todoProject := findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" || todoProject.WorktreeBranch == "" {
		t.Fatalf("todoProject worktree = %#v, want ready worktree", todoProject)
	}
	for index := range state.TodoProjects {
		if state.TodoProjects[index].ID == todoProjectID {
			state.TodoProjects[index].WorktreeStatus = WorktreeStatusCleared
		}
	}
	if err := app.projects.saveLocked(state); err != nil {
		t.Fatalf("save false cleared todo project: %v", err)
	}
	app.gitBranches = func(path string) ([]string, error) {
		if path != todoProject.Path {
			t.Fatalf("git branches path = %q, want source project path %q", path, todoProject.Path)
		}
		return []string{todoProject.WorktreeBranch}, nil
	}

	state, err = app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	todoProject = findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady {
		t.Fatalf("WorktreeStatus = %q, want %q", todoProject.WorktreeStatus, WorktreeStatusReady)
	}
}

func TestAppCreatesTaskTerminalInTaskWorkspaceAndSelectsTaskContext(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("project-terminal", "task-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	state, err = app.CreateTaskTerminal(todoID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTaskTerminal() error = %v", err)
	}

	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	if len(starter.requests) != 2 {
		t.Fatalf("shell start count = %d, want 2", len(starter.requests))
	}
	if starter.requests[1].WorkingDir != taskDir {
		t.Fatalf("task terminal cwd = %q, want %q", starter.requests[1].WorkingDir, taskDir)
	}
	if state.ActiveTerminalID != "task-terminal" {
		t.Fatalf("ActiveTerminalID = %q, want task-terminal", state.ActiveTerminalID)
	}
	if state.ActiveTodoID != todoID || state.ActiveTodoProjectID != "" || state.ActiveProjectID != "" {
		t.Fatalf("active task context = todo:%q todoProject:%q project:%q, want todo:%q with empty project context", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, todoID)
	}
	terminal := findTerminalByID(state.Terminals, "task-terminal")
	if terminal.TodoID != todoID || terminal.TodoProjectID != "" || terminal.ProjectID != "" {
		t.Fatalf("task terminal identity = %#v, want todo-only terminal", terminal)
	}

	state, err = app.SelectTerminal("project-terminal")
	if err != nil {
		t.Fatalf("SelectTerminal(project-terminal) error = %v", err)
	}
	if state.ActiveTerminalID != "project-terminal" || state.ActiveTodoProjectID != todoProjectID || state.ActiveProjectID != state.Projects[0].ID {
		t.Fatalf("project terminal context = terminal:%q todoProject:%q project:%q, want project-terminal/%q/%q", state.ActiveTerminalID, state.ActiveTodoProjectID, state.ActiveProjectID, todoProjectID, state.Projects[0].ID)
	}

	state, err = app.SelectTerminal("task-terminal")
	if err != nil {
		t.Fatalf("SelectTerminal(task-terminal) error = %v", err)
	}
	if state.ActiveTerminalID != "task-terminal" {
		t.Fatalf("ActiveTerminalID after select = %q, want task-terminal", state.ActiveTerminalID)
	}
	if state.ActiveTodoID != todoID || state.ActiveTodoProjectID != "" || state.ActiveProjectID != "" {
		t.Fatalf("selected task terminal context = todo:%q todoProject:%q project:%q, want todo:%q with empty project context", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, todoID)
	}
}

func TestAppStartsTodoProjectBackgroundCommandInPreparedWorktreeWithoutAddingTerminal(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	shellPath := executableFile(t, "zsh")
	background := newFakeBackgroundCommandRunner()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithBackgroundCommandRunner(background.Start),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	if err := app.StartTodoProjectBackgroundCommand(todoProjectID, "npm run sync"); err != nil {
		t.Fatalf("StartTodoProjectBackgroundCommand() error = %v", err)
	}

	state, err = app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	wantWorktreeDir := filepath.Join(taskDir, worktreeDirectoryName(filepath.Base(projectDir)))
	if len(background.requests) != 1 {
		t.Fatalf("background command count = %d, want 1", len(background.requests))
	}
	request := background.requests[0]
	if request.Command != "npm run sync" {
		t.Fatalf("background command = %q, want npm run sync", request.Command)
	}
	if request.WorkingDir != wantWorktreeDir {
		t.Fatalf("background working dir = %q, want %q", request.WorkingDir, wantWorktreeDir)
	}
	if request.ShellPath != shellPath {
		t.Fatalf("background shell path = %q, want %q", request.ShellPath, shellPath)
	}
	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("terminal state after background project command = %#v active %q, want unchanged empty terminals", state.Terminals, state.ActiveTerminalID)
	}
	if state.ActiveTodoID != todoID || state.ActiveTodoProjectID != todoProjectID {
		t.Fatalf("active TODO context = %q/%q, want %q/%q", state.ActiveTodoID, state.ActiveTodoProjectID, todoID, todoProjectID)
	}
}

func TestAppStartsTodoProjectBackgroundCommandInSourceProjectForClearedWorktree(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	shellPath := executableFile(t, "zsh")
	background := newFakeBackgroundCommandRunner()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithBackgroundCommandRunner(background.Start),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for index := range state.TodoProjects {
		if state.TodoProjects[index].ID == todoProjectID {
			state.TodoProjects[index].WorktreeStatus = WorktreeStatusCleared
		}
	}
	if err := app.projects.saveLocked(state); err != nil {
		t.Fatalf("save cleared todo project: %v", err)
	}

	if err := app.StartTodoProjectBackgroundCommand(todoProjectID, "npm run sync"); err != nil {
		t.Fatalf("StartTodoProjectBackgroundCommand() error = %v", err)
	}

	if len(background.requests) != 1 {
		t.Fatalf("background command count = %d, want 1", len(background.requests))
	}
	if background.requests[0].WorkingDir != mustAbs(t, projectDir) {
		t.Fatalf("background working dir = %q, want source project %q", background.requests[0].WorkingDir, mustAbs(t, projectDir))
	}
}

func TestAppStartsTaskBackgroundCommandInTaskWorkspaceWithoutAddingTerminal(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	shellPath := executableFile(t, "bash")
	background := newFakeBackgroundCommandRunner()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithBackgroundCommandRunner(background.Start),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	todoID, _ := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	if err := app.StartTaskBackgroundCommand(todoID, "npm run prepare"); err != nil {
		t.Fatalf("StartTaskBackgroundCommand() error = %v", err)
	}

	state, err = app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	if len(background.requests) != 1 {
		t.Fatalf("background command count = %d, want 1", len(background.requests))
	}
	request := background.requests[0]
	if request.Command != "npm run prepare" {
		t.Fatalf("background command = %q, want npm run prepare", request.Command)
	}
	if request.WorkingDir != taskDir {
		t.Fatalf("background working dir = %q, want %q", request.WorkingDir, taskDir)
	}
	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("terminal state after background task command = %#v active %q, want unchanged empty terminals", state.Terminals, state.ActiveTerminalID)
	}
}

func TestAppRejectsInvalidBackgroundCommandContextsWithoutAddingTerminal(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	background := newFakeBackgroundCommandRunner()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(readyWorktreePreparerFunc(func(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult {
			return WorktreePrepareResult{Status: WorktreeStatusFailed, Error: "git failed"}
		})),
		WithBackgroundCommandRunner(background.Start),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	state, err = app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	state, err = app.AddProjectToTodo(todoID, state.Projects[0].ID)
	if err != nil {
		t.Fatalf("AddProjectToTodo() error = %v", err)
	}
	todoProjectID := state.ActiveTodoProjectID

	if err := app.StartTodoProjectBackgroundCommand(todoProjectID, "npm run sync"); err == nil || !strings.Contains(err.Error(), "todo is not in progress") {
		t.Fatalf("StartTodoProjectBackgroundCommand(not-started) error = %v, want todo status error", err)
	}
	if _, err := app.ChangeTodoStatus(todoID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	if err := app.StartTodoProjectBackgroundCommand(todoProjectID, "npm run sync"); err == nil || !strings.Contains(err.Error(), "project worktree preparation failed") {
		t.Fatalf("StartTodoProjectBackgroundCommand(failed worktree) error = %v, want worktree error", err)
	}
	state, err = app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(background.requests) != 0 {
		t.Fatalf("background command count = %d, want 0", len(background.requests))
	}
	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("terminal state after rejected background command = %#v active %q, want unchanged empty terminals", state.Terminals, state.ActiveTerminalID)
	}
}

func TestAppReportsBackgroundCommandStartFailureWithoutAddingTerminal(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	background := newFakeBackgroundCommandRunner()
	background.err = errors.New("start failed")
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithBackgroundCommandRunner(background.Start),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	if err := app.StartTodoProjectBackgroundCommand(todoProjectID, "npm run sync"); err == nil || !strings.Contains(err.Error(), "start failed") {
		t.Fatalf("StartTodoProjectBackgroundCommand(start failure) error = %v, want start failed", err)
	}
	state, err = app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(background.requests) != 1 {
		t.Fatalf("background command count = %d, want 1 attempted start", len(background.requests))
	}
	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("terminal state after failed background start = %#v active %q, want unchanged empty terminals", state.Terminals, state.ActiveTerminalID)
	}
}

func TestAppWritesTodoInitializationFilesWhenTaskWorkspaceIsPrepared(t *testing.T) {
	workspaceDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		InitializationFiles: []TodoInitializationFileSnapshot{
			{Name: "Agent Rules", FileName: "AGENTS.md", Content: "请先阅读任务说明"},
			{Name: "Prompt", FileName: "prompt.md", Content: "模板内容"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID

	state, err = app.ChangeTodoStatus(todoID, "in-progress")
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	agentsPath := filepath.Join(taskDir, "AGENTS.md")
	agents, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(agents) != "请先阅读任务说明" {
		t.Fatalf("AGENTS.md = %q, want snapshot content", string(agents))
	}

	if err := os.WriteFile(agentsPath, []byte("用户修改内容"), 0o644); err != nil {
		t.Fatalf("modify AGENTS.md: %v", err)
	}
	if err := app.OpenTodoFolder(todoID); err != nil {
		t.Fatalf("OpenTodoFolder() error = %v", err)
	}
	agents, err = os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md after reopen: %v", err)
	}
	if string(agents) != "用户修改内容" {
		t.Fatalf("AGENTS.md after reopen = %q, want existing content preserved", string(agents))
	}
	readme, err := os.ReadFile(filepath.Join(taskDir, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(readme), "# 任务: 修复登录问题") {
		t.Fatalf("README.md = %q, want generated readme", string(readme))
	}
}

func TestAppDelaysTodoInitializationFilesUntilAllWorktreesReady(t *testing.T) {
	workspaceDir := t.TempDir()
	parentDir := t.TempDir()
	projectDir := filepath.Join(parentDir, "frontend-app")
	otherProjectDir := filepath.Join(parentDir, "api-service")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(projectDir) error = %v", err)
	}
	if err := os.MkdirAll(otherProjectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(otherProjectDir) error = %v", err)
	}

	readyPreparer := newReadyWorktreePreparer()
	attempts := map[string]int{}
	preparer := readyWorktreePreparerFunc(func(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult {
		attempts[projectName]++
		if projectName == "api-service" && attempts[projectName] == 1 {
			return WorktreePrepareResult{Status: WorktreeStatusFailed, Error: "worktree failed"}
		}
		return readyPreparer.PrepareWorktree(repoPath, requestedBranch, projectName, taskWorkspaceDir)
	})
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(preparer),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(frontend) error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.AddProjectFromPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(api) error = %v", err)
	}
	otherProjectID := state.ActiveProjectID
	state, err = app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		Projects: []TodoProjectSelection{
			{ProjectID: projectID, BaseBranch: "main"},
			{ProjectID: otherProjectID, BaseBranch: "main"},
		},
		InitializationFiles: []TodoInitializationFileSnapshot{
			{Name: "Agent Rules", FileName: "AGENTS.md", Content: "请先阅读任务说明"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID

	state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	agentsPath := filepath.Join(taskDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md exists before all worktrees are ready, stat error = %v", err)
	}

	app.prepareTodoWorkspace(todoID)
	agents, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md after all worktrees ready: %v", err)
	}
	if string(agents) != "请先阅读任务说明" {
		t.Fatalf("AGENTS.md = %q, want snapshot content", string(agents))
	}
}

func TestAppProjectTerminalUsesPreparedWorktreeDirectory(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("project-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		Projects: []TodoProjectSelection{
			{ProjectID: projectID, BaseBranch: "feature/login-fix"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	todoProjectID := state.TodoProjects[0].ID
	state, err = app.ChangeTodoStatus(todoID, "in-progress")
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	state, err = app.CreateTodoTerminal(todoProjectID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	todoProject := state.TodoProjects[0]
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
		t.Fatalf("todoProject worktree = %#v, want ready path", todoProject)
	}
	wantBranch := worktreeBranchName(todoProject.Name, state.Todos[0].WorkspaceDirName)
	if todoProject.BaseBranch != "feature/login-fix" || todoProject.WorktreeBranch != wantBranch {
		t.Fatalf("todoProject branches = %#v, want feature/login-fix -> %s", todoProject, wantBranch)
	}
	readmePath := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName, "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	wantReadmeLine := "1. " + todoProject.Name + ": base分支为feature/login-fix, 当前worktree分支为" + wantBranch + ";"
	if !strings.Contains(string(readme), wantReadmeLine) {
		t.Fatalf("README.md = %q, want project branch line %q", string(readme), wantReadmeLine)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].WorkingDir != todoProject.WorktreePath {
		t.Fatalf("project terminal cwd = %q, want %q", starter.requests[0].WorkingDir, todoProject.WorktreePath)
	}
}

func TestAppProjectTerminalPreparesMissingWorktreeBeforeStarting(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	app := NewAppWithConfigAndShellStarter(
		configPath,
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("project-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for index := range state.TodoProjects {
		if state.TodoProjects[index].ID == todoProjectID {
			state.TodoProjects[index].WorktreeStatus = ""
			state.TodoProjects[index].WorktreePath = ""
			state.TodoProjects[index].WorktreeBranch = ""
			state.TodoProjects[index].WorktreeError = ""
		}
	}
	if err := app.projects.saveLocked(state); err != nil {
		t.Fatalf("save state without worktree metadata: %v", err)
	}

	state, err = app.CreateTodoTerminal(todoProjectID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	todoProject := findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
		t.Fatalf("todoProject worktree = %#v, want ready path", todoProject)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].WorkingDir != todoProject.WorktreePath {
		t.Fatalf("project terminal cwd = %q, want %q", starter.requests[0].WorkingDir, todoProject.WorktreePath)
	}
}

func TestAppCloseWorkspaceStopsTaskAndProjectTerminalsButKeepsTaskDirectory(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("project-terminal", "task-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}
	state, err = app.CreateTaskTerminal(todoID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTaskTerminal() error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	if _, err := os.Stat(taskDir); err != nil {
		t.Fatalf("task dir missing before close: %v", err)
	}

	state, err = app.CloseWorkspace()
	if err != nil {
		t.Fatalf("CloseWorkspace() error = %v", err)
	}

	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("closed workspace terminals = %#v active %q, want none", state.Terminals, state.ActiveTerminalID)
	}
	if !starter.processes[0].closed || !starter.processes[1].closed {
		t.Fatal("workspace close did not stop task and project terminal processes")
	}
	if _, err := os.Stat(taskDir); err != nil {
		t.Fatalf("task dir missing after close: %v", err)
	}
}

func TestAppCreatesWorkspaceGlobalTerminalsWithoutChangingTodoProjectContext(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("todo-terminal", "global-a", "global-b")),
	)

	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	state, err = app.CreateTodoTerminal(todoProjectID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}
	if state.ActiveTodoProjectID != todoProjectID || state.ActiveProjectID != projectID {
		t.Fatalf("active context before global terminal = %q/%q, want %q/%q", state.ActiveTodoProjectID, state.ActiveProjectID, todoProjectID, projectID)
	}

	state, err = app.CreateWorkspaceTerminal(100, 32)
	if err != nil {
		t.Fatalf("CreateWorkspaceTerminal(A) error = %v", err)
	}
	if state.ActiveTerminalID != "global-a" {
		t.Fatalf("ActiveTerminalID after global A = %q, want global-a", state.ActiveTerminalID)
	}
	if state.ActiveTodoID != todoID || state.ActiveTodoProjectID != todoProjectID || state.ActiveProjectID != projectID {
		t.Fatalf("active context after global A = todo %q todoProject %q project %q, want unchanged", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID)
	}
	if len(state.Terminals) != 2 || !findTerminalByID(state.Terminals, "global-a").WorkspaceTerminal {
		t.Fatalf("Terminals after global A = %#v, want global-a workspace terminal", state.Terminals)
	}

	state, err = app.CreateWorkspaceTerminal(120, 40)
	if err != nil {
		t.Fatalf("CreateWorkspaceTerminal(B) error = %v", err)
	}
	if state.ActiveTerminalID != "global-b" {
		t.Fatalf("ActiveTerminalID after global B = %q, want global-b", state.ActiveTerminalID)
	}
	if state.ActiveTodoProjectID != todoProjectID || state.ActiveProjectID != projectID {
		t.Fatalf("active context after global B = %q/%q, want unchanged", state.ActiveTodoProjectID, state.ActiveProjectID)
	}
	if len(starter.requests) != 3 {
		t.Fatalf("shell start count = %d, want 3", len(starter.requests))
	}
	if !starter.requests[1].WorkspaceTerminal || starter.requests[1].WorkingDir != mustAbs(t, workspaceDir) {
		t.Fatalf("global A request = %#v, want workspace terminal in workspace dir", starter.requests[1])
	}
	if !starter.requests[2].WorkspaceTerminal || starter.requests[2].WorkingDir != mustAbs(t, workspaceDir) {
		t.Fatalf("global B request = %#v, want workspace terminal in workspace dir", starter.requests[2])
	}

	state, err = app.SelectTerminal("global-a")
	if err != nil {
		t.Fatalf("SelectTerminal(global-a) error = %v", err)
	}
	if state.ActiveTerminalID != "global-a" || state.ActiveTodoProjectID != todoProjectID || state.ActiveProjectID != projectID {
		t.Fatalf("state after selecting global A = activeTerminal %q todoProject %q project %q, want global active and context unchanged", state.ActiveTerminalID, state.ActiveTodoProjectID, state.ActiveProjectID)
	}

	state, err = app.AddProjectFromPath(t.TempDir())
	if err != nil {
		t.Fatalf("AddProjectFromPath(after global select) error = %v", err)
	}
	if state.ActiveTerminalID != "global-a" {
		t.Fatalf("ActiveTerminalID after project import while global selected = %q, want global-a", state.ActiveTerminalID)
	}
	if state.ImportSummary == nil || state.ImportSummary.AddedCount != 1 || len(state.ImportSummary.Added) != 1 {
		t.Fatalf("single project import summary = %#v, want one added project", state.ImportSummary)
	}

	state, err = app.SelectTodoProject(todoProjectID)
	if err != nil {
		t.Fatalf("SelectTodoProject(after global select) error = %v", err)
	}
	if state.ActiveTerminalID != "todo-terminal" {
		t.Fatalf("ActiveTerminalID after selecting todo project = %q, want todo-terminal", state.ActiveTerminalID)
	}
}

func TestAppKeepsWorkspaceGlobalTerminalsWhenDeletingTodoAndProjectCandidate(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("global-terminal", "todo-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.CreateWorkspaceTerminal(80, 24)
	if err != nil {
		t.Fatalf("CreateWorkspaceTerminal() error = %v", err)
	}
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	state, err = app.DeleteTodo(todoID)
	if err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if len(state.Terminals) != 1 || !state.Terminals[0].WorkspaceTerminal || state.Terminals[0].ID != "global-terminal" {
		t.Fatalf("Terminals after todo delete = %#v, want only global terminal", state.Terminals)
	}
	if starter.processes[0].closed {
		t.Fatal("global terminal process was closed by todo delete")
	}
	if !starter.processes[1].closed {
		t.Fatal("todo terminal process was not closed by todo delete")
	}

	state, err = app.DeleteProject(projectID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	if len(state.Terminals) != 1 || !state.Terminals[0].WorkspaceTerminal {
		t.Fatalf("Terminals after project delete = %#v, want global terminal preserved", state.Terminals)
	}
	if starter.processes[0].closed {
		t.Fatal("global terminal process was closed by project delete")
	}
}

func TestAppCloseWorkspaceClosesWorkspaceGlobalTerminals(t *testing.T) {
	workspaceDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("global-terminal")),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	if _, err := app.CreateWorkspaceTerminal(80, 24); err != nil {
		t.Fatalf("CreateWorkspaceTerminal() error = %v", err)
	}

	state, err := app.CloseWorkspace()
	if err != nil {
		t.Fatalf("CloseWorkspace() error = %v", err)
	}

	if len(state.Terminals) != 0 || state.ActiveTerminalID != "" {
		t.Fatalf("closed workspace terminals = %#v active %q, want none", state.Terminals, state.ActiveTerminalID)
	}
	if !starter.processes[0].closed {
		t.Fatal("global terminal process was not closed by workspace close")
	}
}

func TestAppPollsClaudeStatusFilesAndEmitsAgentStatus(t *testing.T) {
	projectDir := t.TempDir()
	statusDir := t.TempDir()
	starter := newFakeShellStarter()
	events := make(chan TerminalAgentStatusEvent, 1)
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
		WithClaudeStatusDir(statusDir),
		WithTerminalAgentStatusEmitter(func(event TerminalAgentStatusEvent) {
			events <- event
		}),
	)
	app.startClaudeStatusWatcher()
	defer app.stopClaudeStatusWatcher()

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	state, err = app.CreateTerminal(state.Projects[0].ID, 80, 24)
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-1" {
		t.Fatalf("ActiveTerminalID = %q, want terminal-1", state.ActiveTerminalID)
	}

	statusPayload, err := json.Marshal(map[string]any{
		"session":    "session-a",
		"terminalId": "terminal-1",
		"status":     "waiting",
		"event":      "Notification",
		"cwd":        projectDir,
		"ts":         1718450010,
	})
	if err != nil {
		t.Fatalf("marshal status payload: %v", err)
	}
	if err := os.WriteFile(filepath.Join(statusDir, "session-a.status"), statusPayload, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	app.pollClaudeStatus()

	select {
	case event := <-events:
		if event.TerminalID != "terminal-1" || event.ProjectID != state.Projects[0].ID {
			t.Fatalf("event identity = %#v, want active terminal", event)
		}
		if event.Phase != "needs-input" || event.Source != "claude-hook" {
			t.Fatalf("event status = %#v, want needs-input claude-hook", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for claude agent status event")
	}
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Abs(%q) error = %v", path, err)
	}
	return filepath.Clean(absolutePath)
}

func TestAppCreateTodoAcceptsStructuredRequestAndOptionalProjects(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.AddProjectFromPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(other) error = %v", err)
	}
	otherProjectID := state.ActiveProjectID

	state, err = app.CreateTodo(CreateTodoRequest{
		Title:       "修复登录问题",
		Description: "登录后跳回首页",
		Priority:    TodoPriorityHigh,
		ProjectIDs:  []string{projectID, otherProjectID},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1", len(state.Todos))
	}
	if state.Todos[0].Description != "登录后跳回首页" || state.Todos[0].Priority != TodoPriorityHigh {
		t.Fatalf("Todo = %#v, want description and high priority", state.Todos[0])
	}
	if len(state.TodoProjects) != 2 || state.TodoProjects[0].ProjectID != projectID || state.TodoProjects[1].ProjectID != otherProjectID {
		t.Fatalf("TodoProjects = %#v, want optional project associations in request order", state.TodoProjects)
	}
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" || state.ActiveProjectID != otherProjectID {
		t.Fatalf("active context = %q/%q/%q, want unchanged project context %q", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, otherProjectID)
	}
	if len(state.Terminals) != 0 {
		t.Fatalf("Terminals length = %d, want no terminal during TODO creation", len(state.Terminals))
	}
}

func TestAppChangesTodoStatusThroughPublicAPI(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1", len(state.Todos))
	}

	state, err = app.ChangeTodoStatus(state.Todos[0].ID, "in-progress")
	if err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}

	if len(state.Todos) != 1 || state.Todos[0].Status != "in-progress" {
		t.Fatalf("Todos = %#v, want in-progress todo", state.Todos)
	}
	if state.Terminals == nil {
		t.Fatalf("Terminals = nil, want shell state included")
	}
}

func TestAppStartsTodoAndRunsInitializationLifecycleScriptAsynchronously(t *testing.T) {
	requests := make(chan TodoLifecycleScriptRunRequest, 1)
	release := make(chan struct{})
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithTodoLifecycleScriptRunner(func(ctx context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			requests <- request
			select {
			case <-release:
				return TodoLifecycleScriptRunResult{}
			case <-ctx.Done():
				return TodoLifecycleScriptRunResult{Err: ctx.Err(), ExitCode: -1}
			}
		}),
	)
	defer close(release)

	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		LifecycleScript: &TodoLifecycleScriptSnapshot{
			Name:       "Node setup",
			InitScript: "git checkout -b {{branch_name}}",
			Parameters: []TodoLifecycleScriptParameter{
				{Name: "branch_name", Label: "分支名", DefaultValue: "feature/demo", Required: true},
			},
			ParameterValues: map[string]string{"branch_name": "feature/login"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID

	state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
	if err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}

	if state.Todos[0].Status != TodoStatusInProgress {
		t.Fatalf("Status = %q, want in-progress", state.Todos[0].Status)
	}
	status := lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseInit)
	if status == nil || status.Status != TodoLifecycleScriptStatusRunning {
		t.Fatalf("LifecycleScriptStatuses = %#v, want running init status", state.LifecycleScriptStatuses)
	}
	request := receiveLifecycleScriptRequest(t, requests)
	if request.TodoID != todoID || request.Phase != TodoLifecycleScriptPhaseInit || request.Script != "git checkout -b {{branch_name}}" {
		t.Fatalf("request = %#v, want init script for todo", request)
	}
	if !reflect.DeepEqual(request.Parameters, []TodoLifecycleScriptParameter{{Name: "branch_name", Label: "分支名", DefaultValue: "feature/demo", Required: true}}) {
		t.Fatalf("request.Parameters = %#v, want todo parameter snapshot", request.Parameters)
	}
	if !reflect.DeepEqual(request.ParameterValues, map[string]string{"branch_name": "feature/login"}) {
		t.Fatalf("request.ParameterValues = %#v, want todo parameter values", request.ParameterValues)
	}
	if request.WorkingDir == "" || !strings.Contains(request.WorkingDir, string(os.PathSeparator)+"tasks"+string(os.PathSeparator)) {
		t.Fatalf("WorkingDir = %q, want todo task workspace", request.WorkingDir)
	}
}

func TestAppCompletionLifecycleScriptCompletesTodoAfterSuccess(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithTodoLifecycleScriptRunner(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			return TodoLifecycleScriptRunResult{}
		}),
	)
	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		LifecycleScript: &TodoLifecycleScriptSnapshot{
			Name:           "Node setup",
			CompleteScript: "npm test",
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	if _, err := app.ChangeTodoStatus(todoID, TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}

	state, err = app.CompleteTodo(todoID)
	if err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}
	if todo := findTodo(state.Todos, todoID); todo == nil || todo.Status != TodoStatusInProgress {
		t.Fatalf("todo after CompleteTodo = %#v, want still in-progress while script runs", todo)
	}

	state = waitForAppTodoStatus(t, app, todoID, TodoStatusCompleted)
	if lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseComplete) != nil {
		t.Fatalf("LifecycleScriptStatuses = %#v, want completion status cleared", state.LifecycleScriptStatuses)
	}
}

func TestAppFailedCompletionLifecycleScriptKeepsTodoRetryable(t *testing.T) {
	attempts := 0
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithTodoLifecycleScriptRunner(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			attempts++
			if attempts == 1 {
				return TodoLifecycleScriptRunResult{Output: "lint failed", ExitCode: 2, Err: errors.New("exit status 2")}
			}
			return TodoLifecycleScriptRunResult{}
		}),
	)
	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		LifecycleScript: &TodoLifecycleScriptSnapshot{
			Name:           "Node setup",
			CompleteScript: "npm test",
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	if _, err := app.ChangeTodoStatus(todoID, TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}
	if _, err := app.CompleteTodo(todoID); err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	state = waitForAppLifecycleScriptStatus(t, app, todoID, TodoLifecycleScriptPhaseComplete, TodoLifecycleScriptStatusFailed)
	if todo := findTodo(state.Todos, todoID); todo == nil || todo.Status != TodoStatusInProgress {
		t.Fatalf("todo after failed completion script = %#v, want in-progress", todo)
	}
	status := lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseComplete)
	if status == nil || status.ExitCode != 2 || status.OutputTail != "lint failed" {
		t.Fatalf("completion status = %#v, want failed status with output", status)
	}

	if _, err := app.RetryTodoLifecycleScript(todoID, TodoLifecycleScriptPhaseComplete); err != nil {
		t.Fatalf("RetryTodoLifecycleScript() error = %v", err)
	}
	state = waitForAppTodoStatus(t, app, todoID, TodoStatusCompleted)
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseComplete) != nil {
		t.Fatalf("LifecycleScriptStatuses = %#v, want status cleared after retry success", state.LifecycleScriptStatuses)
	}
}

func TestAppDeleteTodoClearsLifecycleScriptStatus(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithTodoLifecycleScriptRunner(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			return TodoLifecycleScriptRunResult{Output: "setup failed", ExitCode: 1, Err: errors.New("exit status 1")}
		}),
	)
	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		LifecycleScript: &TodoLifecycleScriptSnapshot{
			Name:       "Node setup",
			InitScript: "npm install",
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	if _, err := app.ChangeTodoStatus(todoID, TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}
	state = waitForAppLifecycleScriptStatus(t, app, todoID, TodoLifecycleScriptPhaseInit, TodoLifecycleScriptStatusFailed)
	if lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseInit) == nil {
		t.Fatalf("LifecycleScriptStatuses = %#v, want failed init status before delete", state.LifecycleScriptStatuses)
	}

	state, err = app.DeleteTodo(todoID)
	if err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, TodoLifecycleScriptPhaseInit) != nil {
		t.Fatalf("LifecycleScriptStatuses = %#v, want deleted todo status cleared", state.LifecycleScriptStatuses)
	}
}

func TestAppAddsMultipleProjectsToExistingTodo(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.AddProjectFromPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(other) error = %v", err)
	}
	otherProjectID := state.ActiveProjectID
	state, err = app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID

	state, err = app.AddProjectsToTodo(todoID, []string{projectID, otherProjectID})
	if err != nil {
		t.Fatalf("AddProjectsToTodo() error = %v", err)
	}

	if len(state.TodoProjects) != 2 || state.TodoProjects[0].ProjectID != projectID || state.TodoProjects[1].ProjectID != otherProjectID {
		t.Fatalf("TodoProjects = %#v, want both projects associated", state.TodoProjects)
	}
	if state.ActiveTodoProjectID != state.TodoProjects[0].ID || state.ActiveProjectID != projectID {
		t.Fatalf("active context = %q/%q, want first added project", state.ActiveTodoProjectID, state.ActiveProjectID)
	}
}

func TestAppCompletesTodoAndClosesOnlyOwnedTerminals(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	state, err = app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	todoAID := state.Todos[0].ID
	state, err = app.AddProjectToTodo(todoAID, projectID)
	if err != nil {
		t.Fatalf("AddProjectToTodo(A) error = %v", err)
	}
	todoProjectAID := state.ActiveTodoProjectID
	if _, err := app.ChangeTodoStatus(todoAID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(A) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A) error = %v", err)
	}
	state, err = app.CreateTodo(CreateTodoRequest{Title: "升级依赖"})
	if err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}
	todoBID := state.Todos[1].ID
	state, err = app.AddProjectToTodo(todoBID, projectID)
	if err != nil {
		t.Fatalf("AddProjectToTodo(B) error = %v", err)
	}
	todoProjectBID := state.ActiveTodoProjectID
	if _, err := app.ChangeTodoStatus(todoBID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(B) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
	}

	state, err = app.CompleteTodo(todoAID)
	if err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	if !starter.processes[0].closed {
		t.Fatal("completed todo terminal process was not closed")
	}
	if starter.processes[1].closed {
		t.Fatal("other todo terminal process was closed")
	}
	if len(state.Terminals) != 1 || state.Terminals[0].TodoID != todoBID {
		t.Fatalf("Terminals = %#v, want only todo B terminal", state.Terminals)
	}
	completed := findTodo(state.Todos, todoAID)
	if completed == nil || completed.Status != TodoStatusCompleted {
		t.Fatalf("completed todo = %#v, want completed status", completed)
	}
}

func TestAppCompletesTodoClosesTerminalsBeforeReadingSnapshotBranches(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(readyWorktreePreparerFunc(func(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult {
			worktreePath := filepath.Join(taskWorkspaceDir, worktreeDirectoryName(projectName))
			if err := os.MkdirAll(worktreePath, 0o755); err != nil {
				t.Fatalf("MkdirAll(%q) error = %v", worktreePath, err)
			}
			return WorktreePrepareResult{
				BaseBranch:   "main",
				WorktreePath: worktreePath,
				Status:       WorktreeStatusReady,
			}
		})),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	todoID, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}
	app.projects.gitBranch = func(path string) (string, error) {
		if len(starter.processes) != 1 || !starter.processes[0].closed {
			t.Fatal("git branch snapshot read before completed todo terminal was closed")
		}
		return "feature/login", nil
	}

	state, err = app.CompleteTodo(todoID)
	if err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	completed := findTodo(state.Todos, todoID)
	if completed == nil || len(completed.ProjectSnapshots) != 1 {
		t.Fatalf("completed todo = %#v, want one project snapshot", completed)
	}
	if completed.ProjectSnapshots[0].WorktreeBranch != "feature/login" {
		t.Fatalf("WorktreeBranch = %q, want feature/login", completed.ProjectSnapshots[0].WorktreeBranch)
	}
}

func TestAppDeletesCompletedTodosThroughPublicAPI(t *testing.T) {
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
	)

	state, err := app.CreateTodo(CreateTodoRequest{Title: "完成任务 A"})
	if err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	todoAID := state.Todos[0].ID
	if _, err := app.ChangeTodoStatus(todoAID, TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(A) error = %v", err)
	}
	if _, err := app.CompleteTodo(todoAID); err != nil {
		t.Fatalf("CompleteTodo(A) error = %v", err)
	}
	state, err = app.CreateTodo(CreateTodoRequest{Title: "完成任务 B"})
	if err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}
	todoBID := state.Todos[1].ID
	if _, err := app.ChangeTodoStatus(todoBID, TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(B) error = %v", err)
	}
	if _, err := app.CompleteTodo(todoBID); err != nil {
		t.Fatalf("CompleteTodo(B) error = %v", err)
	}

	state, err = app.DeleteCompletedTodos([]string{todoAID, todoBID})
	if err != nil {
		t.Fatalf("DeleteCompletedTodos() error = %v", err)
	}

	if len(state.Todos) != 0 {
		t.Fatalf("Todos = %#v, want completed todos removed", state.Todos)
	}
	if state.Terminals == nil {
		t.Fatalf("Terminals = nil, want shell state included")
	}
	if len(starter.requests) != 0 {
		t.Fatalf("shell start count = %d, want no shell processes started", len(starter.requests))
	}
}

func TestAppUpdateTodoRemovesProjectAndClosesOnlyThatTodoProjectTerminals(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b", "terminal-c")),
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectAID := state.Projects[0].ID
	state, err = app.AddProjectFromPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	projectBID := state.ActiveProjectID
	state, err = app.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{projectAID, projectBID}})
	if err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	todoAID := state.Todos[0].ID
	todoProjectAProjectAID := state.TodoProjects[0].ID
	todoProjectAProjectBID := state.TodoProjects[1].ID
	if _, err := app.ChangeTodoStatus(todoAID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(A) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectAProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A/projectA) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectAProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A/projectB) error = %v", err)
	}
	state, err = app.CreateTodo(CreateTodoRequest{Title: "升级依赖", ProjectIDs: []string{projectAID}})
	if err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}
	todoBID := state.Todos[1].ID
	todoProjectBProjectAID := state.TodoProjects[2].ID
	if _, err := app.ChangeTodoStatus(todoBID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(B) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectBProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B/projectA) error = %v", err)
	}
	if _, err := app.SelectTodoProject(todoProjectAProjectAID); err != nil {
		t.Fatalf("SelectTodoProject(A/projectA) error = %v", err)
	}

	state, err = app.UpdateTodo(UpdateTodoRequest{
		ID:          todoAID,
		Title:       "修复登录跳转",
		Description: "登录后跳回首页",
		Priority:    TodoPriorityHigh,
		ProjectIDs:  []string{projectBID},
	})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}

	if !starter.processes[0].closed {
		t.Fatal("removed TODO-project terminal process was not closed")
	}
	if starter.processes[1].closed {
		t.Fatal("remaining project terminal under edited TODO was closed")
	}
	if starter.processes[2].closed {
		t.Fatal("same project terminal under other TODO was closed")
	}
	if len(state.Terminals) != 2 {
		t.Fatalf("Terminals length = %d, want 2 preserved terminals", len(state.Terminals))
	}
	if state.ActiveTodoID != todoAID || state.ActiveTodoProjectID != todoProjectAProjectBID || state.ActiveProjectID != projectBID {
		t.Fatalf("active context = %q/%q/%q, want edited TODO remaining project", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID)
	}
	updated := findTodo(state.Todos, todoAID)
	if updated == nil || updated.Title != "修复登录跳转" || updated.Priority != TodoPriorityHigh {
		t.Fatalf("updated todo = %#v, want metadata changes", updated)
	}
}

func TestAppRemoveTodoProjectClosesOnlyThatTodoProjectTerminals(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	_, todoProjectAID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	if _, err := app.CreateTodoTerminal(todoProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A) error = %v", err)
	}
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectID)
	if _, err := app.CreateTodoTerminal(todoProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
	}

	state, err = app.RemoveTodoProject(todoProjectAID)
	if err != nil {
		t.Fatalf("RemoveTodoProject() error = %v", err)
	}

	if !starter.processes[0].closed {
		t.Fatal("removed todo-project terminal process was not closed")
	}
	if starter.processes[1].closed {
		t.Fatal("other todo-project terminal process was closed")
	}
	if len(state.TodoProjects) != 1 || state.TodoProjects[0].ID != todoProjectBID {
		t.Fatalf("TodoProjects = %#v, want only todo-project B", state.TodoProjects)
	}
	if len(state.Terminals) != 1 || state.Terminals[0].TodoProjectID != todoProjectBID {
		t.Fatalf("Terminals = %#v, want only todo-project B terminal", state.Terminals)
	}
}

func TestAppTodoProjectUIStatePersistsUnderWorkspaceData(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
	)
	state, err := app.OpenWorkspaceFromPath(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	dataPath := state.CurrentWorkspace.DataPath
	state, err = app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	if err := app.SaveTodoProjectUIState(todoProjectID, TodoProjectUIState{TodoView: "completed"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState() error = %v", err)
	}
	if err := app.SaveTodoSidebarWidth(360); err != nil {
		t.Fatalf("SaveTodoSidebarWidth() error = %v", err)
	}
	loaded, err := app.LoadTodoProjectUIState()
	if err != nil {
		t.Fatalf("LoadTodoProjectUIState() error = %v", err)
	}

	if loaded.TodoProjects[todoProjectID].TodoView != "completed" {
		t.Fatalf("TodoView = %q, want completed", loaded.TodoProjects[todoProjectID].TodoView)
	}
	if loaded.SidebarWidth != 360 {
		t.Fatalf("SidebarWidth = %d, want 360", loaded.SidebarWidth)
	}
	if _, err := os.Stat(filepath.Join(dataPath, "todo-project-ui-state.json")); err != nil {
		t.Fatalf("todo project ui state file missing under .data: %v", err)
	}
}

func TestAppStartupRestoresTodoProjectUIStateFromWorkspaceData(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	state, err := app.OpenWorkspaceFromPath(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if err := app.SaveTodoProjectUIState(todoProjectID, TodoProjectUIState{TodoView: "in-progress"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState() error = %v", err)
	}
	if err := app.SaveTodoSidebarWidth(420); err != nil {
		t.Fatalf("SaveTodoSidebarWidth() error = %v", err)
	}

	restarted := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
		WithRestoreLastWorkspaceOnStartup(),
		WithClaudeStatusDir(""),
	)
	restarted.startup(nil)
	defer restarted.shutdown(nil)

	loaded, err := restarted.LoadTodoProjectUIState()
	if err != nil {
		t.Fatalf("LoadTodoProjectUIState() error = %v", err)
	}
	if loaded.TodoProjects[todoProjectID].TodoView != "in-progress" {
		t.Fatalf("TodoView = %q, want in-progress", loaded.TodoProjects[todoProjectID].TodoView)
	}
	if loaded.SidebarWidth != 420 {
		t.Fatalf("SidebarWidth = %d, want 420", loaded.SidebarWidth)
	}
}

func TestAppTodoProjectUIStateRequiresWorkspace(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
	)

	if _, err := app.LoadTodoProjectUIState(); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("LoadTodoProjectUIState() error = %v, want ErrWorkspaceRequired", err)
	}
	if err := app.SaveTodoProjectUIState("todo-project-a", TodoProjectUIState{TodoView: "completed"}); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("SaveTodoProjectUIState() error = %v, want ErrWorkspaceRequired", err)
	}
	if err := app.SaveTodoSidebarWidth(360); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("SaveTodoSidebarWidth() error = %v, want ErrWorkspaceRequired", err)
	}
	if err := app.DeleteTodoProjectUIState([]string{"todo-project-a"}); !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("DeleteTodoProjectUIState() error = %v, want ErrWorkspaceRequired", err)
	}
}

func TestAppRemoveTodoProjectDeletesTodoProjectUIState(t *testing.T) {
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
	_, todoProjectAID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectID)
	if err := app.SaveTodoProjectUIState(todoProjectAID, TodoProjectUIState{TodoView: "completed"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState(A) error = %v", err)
	}
	if err := app.SaveTodoProjectUIState(todoProjectBID, TodoProjectUIState{TodoView: "in-progress"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState(B) error = %v", err)
	}
	if err := app.SaveTodoSidebarWidth(420); err != nil {
		t.Fatalf("SaveTodoSidebarWidth() error = %v", err)
	}

	if _, err := app.RemoveTodoProject(todoProjectAID); err != nil {
		t.Fatalf("RemoveTodoProject() error = %v", err)
	}
	loaded, err := app.LoadTodoProjectUIState()
	if err != nil {
		t.Fatalf("LoadTodoProjectUIState() error = %v", err)
	}

	if _, ok := loaded.TodoProjects[todoProjectAID]; ok {
		t.Fatalf("removed todo-project UI state still exists: %#v", loaded.TodoProjects)
	}
	if loaded.TodoProjects[todoProjectBID].TodoView != "in-progress" {
		t.Fatalf("remaining todo-project UI state = %#v, want B preserved", loaded.TodoProjects[todoProjectBID])
	}
	if loaded.SidebarWidth != 420 {
		t.Fatalf("SidebarWidth = %d, want 420", loaded.SidebarWidth)
	}
}

func TestAppDeleteTodoDeletesTodoProjectUIState(t *testing.T) {
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
	todoAID, todoProjectAID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectID)
	if err := app.SaveTodoProjectUIState(todoProjectAID, TodoProjectUIState{TodoView: "completed"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState(A) error = %v", err)
	}
	if err := app.SaveTodoProjectUIState(todoProjectBID, TodoProjectUIState{TodoView: "in-progress"}); err != nil {
		t.Fatalf("SaveTodoProjectUIState(B) error = %v", err)
	}
	if err := app.SaveTodoSidebarWidth(420); err != nil {
		t.Fatalf("SaveTodoSidebarWidth() error = %v", err)
	}

	if _, err := app.DeleteTodo(todoAID); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	loaded, err := app.LoadTodoProjectUIState()
	if err != nil {
		t.Fatalf("LoadTodoProjectUIState() error = %v", err)
	}

	if _, ok := loaded.TodoProjects[todoProjectAID]; ok {
		t.Fatalf("deleted TODO project UI state still exists: %#v", loaded.TodoProjects)
	}
	if loaded.TodoProjects[todoProjectBID].TodoView != "in-progress" {
		t.Fatalf("remaining todo-project UI state = %#v, want B preserved", loaded.TodoProjects[todoProjectBID])
	}
	if loaded.SidebarWidth != 420 {
		t.Fatalf("SidebarWidth = %d, want 420", loaded.SidebarWidth)
	}
}

func TestAppImportsProjectsFromParentDirectory(t *testing.T) {
	parentDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(parentDir, "frontend-app"), 0o755); err != nil {
		t.Fatalf("mkdir frontend-app: %v", err)
	}
	if err := os.Mkdir(filepath.Join(parentDir, "frontend-app", ".git"), 0o755); err != nil {
		t.Fatalf("mkdir frontend-app .git: %v", err)
	}
	if err := os.Mkdir(filepath.Join(parentDir, "api-service"), 0o755); err != nil {
		t.Fatalf("mkdir api-service: %v", err)
	}
	if err := os.Mkdir(filepath.Join(parentDir, "api-service", ".git"), 0o755); err != nil {
		t.Fatalf("mkdir api-service .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(parentDir, "README.md"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write README: %v", err)
	}
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.ImportProjectsFromParentDirectory(parentDir)
	if err != nil {
		t.Fatalf("ImportProjectsFromParentDirectory() error = %v", err)
	}

	if len(state.Projects) != 2 {
		t.Fatalf("Projects length = %d, want 2", len(state.Projects))
	}
	if state.ImportSummary == nil || state.ImportSummary.AddedCount != 2 || state.ImportSummary.SkippedCount != 0 {
		t.Fatalf("ImportSummary = %#v, want 2 added", state.ImportSummary)
	}
}

func TestAppImportProjectFromPathRequiresGitInitializationForNonGitDirectory(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	result, err := app.ImportProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("ImportProjectFromPath() error = %v", err)
	}

	if !result.RequiresGitInitialization {
		t.Fatalf("RequiresGitInitialization = false, want true")
	}
	if result.Path != mustAbs(t, projectDir) {
		t.Fatalf("Path = %q, want %q", result.Path, mustAbs(t, projectDir))
	}
	state, err := app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(state.Projects) != 0 {
		t.Fatalf("Projects = %#v, want no import before initialization", state.Projects)
	}
}

func TestAppImportProjectFromPathImportsGitDirectory(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	result, err := app.ImportProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("ImportProjectFromPath() error = %v", err)
	}

	if result.RequiresGitInitialization {
		t.Fatalf("RequiresGitInitialization = true, want false")
	}
	if len(result.State.Projects) != 1 || result.State.Projects[0].Path != mustAbs(t, projectDir) {
		t.Fatalf("State.Projects = %#v, want imported Git directory", result.State.Projects)
	}
}

func TestAppInitializeGitRepositoryAndImportProjectInitializesThenAdds(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	calls := 0
	app.gitInit = func(path string) error {
		calls++
		if path != mustAbs(t, projectDir) {
			t.Fatalf("git init path = %q, want %q", path, mustAbs(t, projectDir))
		}
		if err := os.WriteFile(filepath.Join(path, ".git"), []byte("gitdir: initialized"), 0o600); err != nil {
			t.Fatalf("write .git file: %v", err)
		}
		return nil
	}

	state, err := app.InitializeGitRepositoryAndImportProject(projectDir)
	if err != nil {
		t.Fatalf("InitializeGitRepositoryAndImportProject() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("git init calls = %d, want 1", calls)
	}
	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectDir) {
		t.Fatalf("Projects = %#v, want initialized project imported", state.Projects)
	}
}

func TestAppInitializeGitRepositoryAndImportProjectCreatesInitialCommitBeforeImport(t *testing.T) {
	if err := gitCommandAvailable(); err != nil {
		t.Skipf("git not available: %v", err)
	}
	t.Setenv("GIT_AUTHOR_NAME", "Todo Helper Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "todo-helper-test@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Todo Helper Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "todo-helper-test@example.invalid")

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "app.txt"), []byte("source file\n"), 0o600); err != nil {
		t.Fatalf("write app.txt: %v", err)
	}
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	state, err := app.InitializeGitRepositoryAndImportProject(projectDir)
	if err != nil {
		t.Fatalf("InitializeGitRepositoryAndImportProject() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].Path != mustAbs(t, projectDir) {
		t.Fatalf("Projects = %#v, want initialized project imported", state.Projects)
	}
	if output, err := exec.Command("git", "-C", projectDir, "rev-parse", "--verify", "HEAD").CombinedOutput(); err != nil {
		t.Fatalf("git rev-parse HEAD error = %v output=%s", err, string(output))
	}
	output, err := exec.Command("git", "-C", projectDir, "ls-tree", "--name-only", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-tree HEAD error = %v output=%s", err, string(output))
	}
	if !strings.Contains(string(output), "app.txt") {
		t.Fatalf("HEAD files = %q, want app.txt", string(output))
	}
}

func TestAppNonGitInitializationCommitSupportsTodoWorktreeStart(t *testing.T) {
	if err := gitCommandAvailable(); err != nil {
		t.Skipf("git not available: %v", err)
	}
	t.Setenv("GIT_AUTHOR_NAME", "Todo Helper Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "todo-helper-test@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Todo Helper Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "todo-helper-test@example.invalid")

	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "app.txt"), []byte("source file\n"), 0o600); err != nil {
		t.Fatalf("write app.txt: %v", err)
	}
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.InitializeGitRepositoryAndImportProject(projectDir)
	if err != nil {
		t.Fatalf("InitializeGitRepositoryAndImportProject() error = %v", err)
	}
	projectID := state.Projects[0].ID

	state, err = app.CreateTodo(CreateTodoRequest{
		Title: "修复初始化 worktree",
		Projects: []TodoProjectSelection{
			{ProjectID: projectID},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
	if err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}

	if len(state.TodoProjects) != 1 {
		t.Fatalf("TodoProjects = %#v, want one associated project", state.TodoProjects)
	}
	todoProject := state.TodoProjects[0]
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
		t.Fatalf("todoProject worktree = %#v, want ready worktree", todoProject)
	}
	if _, err := os.Stat(filepath.Join(todoProject.WorktreePath, "app.txt")); err != nil {
		t.Fatalf("worktree app.txt stat error = %v, want initial file in worktree", err)
	}
}

func TestAppInitializeGitRepositoryAndImportProjectDoesNotImportWhenInitializationFails(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	app.gitInit = func(path string) error {
		if path != mustAbs(t, projectDir) {
			t.Fatalf("git init path = %q, want %q", path, mustAbs(t, projectDir))
		}
		return errors.New("git commit failed: Author identity unknown")
	}

	if _, err := app.InitializeGitRepositoryAndImportProject(projectDir); err == nil || !strings.Contains(err.Error(), "git commit failed") {
		t.Fatalf("InitializeGitRepositoryAndImportProject() error = %v, want git commit failure", err)
	}
	state, err := app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(state.Projects) != 0 {
		t.Fatalf("Projects = %#v, want no import after initialization failure", state.Projects)
	}
}

func TestAppInitializeGitRepositoryAndImportProjectRemovesCreatedMetadataWhenInitializationFails(t *testing.T) {
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	app.gitInit = func(path string) error {
		if err := os.Mkdir(filepath.Join(path, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir .git: %v", err)
		}
		return errors.New("git commit failed: Author identity unknown")
	}

	if _, err := app.InitializeGitRepositoryAndImportProject(projectDir); err == nil {
		t.Fatal("InitializeGitRepositoryAndImportProject() error = nil, want commit failure")
	}
	if pathHasGitRepositoryMetadata(projectDir) {
		t.Fatal("pathHasGitRepositoryMetadata() = true, want failed initialization cleanup")
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
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != shellPath {
		t.Fatalf("ShellPath = %q, want %q", starter.requests[0].ShellPath, shellPath)
	}
}

func TestAppSavesTerminalLaunchProfiles(t *testing.T) {
	configDir := t.TempDir()
	shellPath := executableFile(t, "zsh")
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(configDir, "projects.json"),
		newFakeShellStarter().Start,
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}

	state, err := app.SaveTerminalLaunchProfiles([]TerminalLaunchProfileSetting{
		{Name: "Codex", Command: "codex --model gpt-5", Enabled: true},
	})
	if err != nil {
		t.Fatalf("SaveTerminalLaunchProfiles() error = %v", err)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "Codex", Command: "codex --model gpt-5", Enabled: true},
	})

	loaded, err := app.LoadTerminalSettings()
	if err != nil {
		t.Fatalf("LoadTerminalSettings() error = %v", err)
	}
	assertLaunchProfiles(t, loaded.LaunchProfiles, state.LaunchProfiles)
}

func TestAppSavesTodoInitializationFilesGlobally(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	files := []TodoInitializationFileTemplate{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "请先阅读任务说明", DefaultSelected: true},
		{Name: "Prompt", Description: "可选提示词", FileName: "prompt.md", Content: "生成实现计划"},
	}

	state, err := app.SaveTodoInitializationFiles(files)
	if err != nil {
		t.Fatalf("SaveTodoInitializationFiles() error = %v", err)
	}
	assertTodoInitializationFiles(t, state.TodoInitializationFiles, files)

	loaded, err := app.LoadTodoInitializationFiles()
	if err != nil {
		t.Fatalf("LoadTodoInitializationFiles() error = %v", err)
	}
	assertTodoInitializationFiles(t, loaded, files)
}

func TestAppSavesTodoLifecycleScriptsGlobally(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	scripts := []TodoLifecycleScriptTemplate{
		{Name: "Node setup", Description: "安装依赖", InitScript: "npm install", CompleteScript: "npm test", DefaultSelected: true},
		{Name: "Cleanup", Description: "清理缓存", CompleteScript: "rm -rf tmp"},
	}

	state, err := app.SaveTodoLifecycleScripts(scripts)
	if err != nil {
		t.Fatalf("SaveTodoLifecycleScripts() error = %v", err)
	}
	assertTodoLifecycleScripts(t, state.TodoLifecycleScripts, scripts)

	loaded, err := app.LoadTodoLifecycleScripts()
	if err != nil {
		t.Fatalf("LoadTodoLifecycleScripts() error = %v", err)
	}
	assertTodoLifecycleScripts(t, loaded, scripts)
}

func TestAppCreateTodoStoresInitializationFileSnapshots(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	initialFiles := []TodoInitializationFileTemplate{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "旧内容", DefaultSelected: true},
	}
	if _, err := app.SaveTodoInitializationFiles(initialFiles); err != nil {
		t.Fatalf("SaveTodoInitializationFiles(initial) error = %v", err)
	}

	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		InitializationFiles: []TodoInitializationFileSnapshot{
			{Name: initialFiles[0].Name, Description: initialFiles[0].Description, FileName: initialFiles[0].FileName, Content: initialFiles[0].Content},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := app.SaveTodoInitializationFiles([]TodoInitializationFileTemplate{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "新内容", DefaultSelected: true},
	}); err != nil {
		t.Fatalf("SaveTodoInitializationFiles(updated) error = %v", err)
	}

	reloaded, err := app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(state.Todos) != 1 || len(reloaded.Todos) != 1 {
		t.Fatalf("todos = initial %#v reloaded %#v, want one todo", state.Todos, reloaded.Todos)
	}
	assertTodoInitializationFileSnapshots(t, reloaded.Todos[0].InitializationFiles, []TodoInitializationFileSnapshot{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "旧内容"},
	})
}

func TestAppCreateTodoStoresLifecycleScriptSnapshot(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	initialScripts := []TodoLifecycleScriptTemplate{
		{Name: "Node setup", Description: "安装依赖", InitScript: "npm install", CompleteScript: "npm test", DefaultSelected: true},
	}
	if _, err := app.SaveTodoLifecycleScripts(initialScripts); err != nil {
		t.Fatalf("SaveTodoLifecycleScripts(initial) error = %v", err)
	}

	state, err := app.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		LifecycleScript: &TodoLifecycleScriptSnapshot{
			Name:           initialScripts[0].Name,
			Description:    initialScripts[0].Description,
			InitScript:     initialScripts[0].InitScript,
			CompleteScript: initialScripts[0].CompleteScript,
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := app.SaveTodoLifecycleScripts([]TodoLifecycleScriptTemplate{
		{Name: "Node setup", Description: "安装依赖", InitScript: "npm ci", CompleteScript: "npm run lint", DefaultSelected: true},
	}); err != nil {
		t.Fatalf("SaveTodoLifecycleScripts(updated) error = %v", err)
	}

	reloaded, err := app.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(state.Todos) != 1 || len(reloaded.Todos) != 1 {
		t.Fatalf("todos = initial %#v reloaded %#v, want one todo", state.Todos, reloaded.Todos)
	}
	assertTodoLifecycleScriptSnapshot(t, reloaded.Todos[0].LifecycleScript, &TodoLifecycleScriptSnapshot{
		Name:           "Node setup",
		Description:    "安装依赖",
		InitScript:     "npm install",
		CompleteScript: "npm test",
	})
}

func TestAppSavesTerminalThemeWithoutChangingProjectShellBehavior(t *testing.T) {
	configDir := t.TempDir()
	projectDir := t.TempDir()
	shellPath := executableFile(t, "zsh")
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(configDir, "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	if _, err := app.SaveTerminalShell(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell() error = %v", err)
	}

	settings, err := app.SaveTerminalTheme(AppearanceThemeDark)
	if err != nil {
		t.Fatalf("SaveTerminalTheme() error = %v", err)
	}
	if settings.Theme != AppearanceThemeDark {
		t.Fatalf("Theme = %q, want %q", settings.Theme, AppearanceThemeDark)
	}

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != shellPath {
		t.Fatalf("ShellPath = %q, want %q", starter.requests[0].ShellPath, shellPath)
	}

	loaded, err := app.LoadTerminalSettings()
	if err != nil {
		t.Fatalf("LoadTerminalSettings() error = %v", err)
	}
	if loaded.Theme != AppearanceThemeDark {
		t.Fatalf("loaded Theme = %q, want %q", loaded.Theme, AppearanceThemeDark)
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
		WithWorktreePreparer(newReadyWorktreePreparer()),
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
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(first) error = %v", err)
	}

	if _, err := app.SaveTerminalShell(shellB, ShellSourceManual); err != nil {
		t.Fatalf("SaveTerminalShell(shellB) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(second) error = %v", err)
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
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)
	if _, err := app.CreateTodoTerminal(todoProjectID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal() error = %v", err)
	}

	if len(starter.requests) != 1 {
		t.Fatalf("shell start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != fallbackShell {
		t.Fatalf("ShellPath = %q, want fallback %q", starter.requests[0].ShellPath, fallbackShell)
	}
}

func TestAppDeletesProjectCandidateWithoutClosingTodoProjectTerminals(t *testing.T) {
	projectDirA := t.TempDir()
	projectDirB := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a1", "terminal-a2", "terminal-b1")),
	)

	state, err := app.AddProjectFromPath(projectDirA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectAID := state.Projects[0].ID
	_, todoProjectAID := createTodoProjectForApp(t, app, "修复登录问题", projectAID)
	if _, err := app.CreateTodoTerminal(todoProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A1) error = %v", err)
	}
	if _, err := app.CreateTodoTerminal(todoProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A2) error = %v", err)
	}
	state, err = app.AddProjectFromPath(projectDirB)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	projectBID := findProjectByPathForApp(t, state.Projects, projectDirB).ID
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectBID)
	if _, err := app.CreateTodoTerminal(todoProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
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
		t.Fatalf("ActiveTerminalID = %q, want current todo project terminal preserved", state.ActiveTerminalID)
	}
	if len(state.Terminals) != 3 {
		t.Fatalf("Terminals = %#v, want all TODO project terminals preserved", state.Terminals)
	}
	if starter.processes[0].closed || starter.processes[1].closed {
		t.Fatal("deleted candidate closed TODO project terminal processes")
	}
	if starter.processes[2].closed {
		t.Fatal("remaining project terminal process was closed")
	}
}

func TestAppDeletesProjectCandidatesWithoutClosingTodoProjectTerminals(t *testing.T) {
	projectDirA := t.TempDir()
	projectDirB := t.TempDir()
	projectDirC := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a1", "terminal-b1", "terminal-c1")),
	)

	state, err := app.AddProjectFromPath(projectDirA)
	if err != nil {
		t.Fatalf("AddProjectFromPath(A) error = %v", err)
	}
	projectAID := state.Projects[0].ID
	_, todoProjectAID := createTodoProjectForApp(t, app, "修复登录问题", projectAID)
	if _, err := app.CreateTodoTerminal(todoProjectAID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(A) error = %v", err)
	}

	state, err = app.AddProjectFromPath(projectDirB)
	if err != nil {
		t.Fatalf("AddProjectFromPath(B) error = %v", err)
	}
	projectBID := findProjectByPathForApp(t, state.Projects, projectDirB).ID
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectBID)
	if _, err := app.CreateTodoTerminal(todoProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
	}

	state, err = app.AddProjectFromPath(projectDirC)
	if err != nil {
		t.Fatalf("AddProjectFromPath(C) error = %v", err)
	}
	projectCID := findProjectByPathForApp(t, state.Projects, projectDirC).ID
	_, todoProjectCID := createTodoProjectForApp(t, app, "整理项目", projectCID)
	if _, err := app.CreateTodoTerminal(todoProjectCID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(C) error = %v", err)
	}
	state, err = app.SelectTerminal("terminal-b1")
	if err != nil {
		t.Fatalf("SelectTerminal(terminal-b1) error = %v", err)
	}
	if state.ActiveTerminalID != "terminal-b1" {
		t.Fatalf("ActiveTerminalID setup = %q, want terminal-b1", state.ActiveTerminalID)
	}

	state, err = app.DeleteProjects([]string{projectAID, projectCID})
	if err != nil {
		t.Fatalf("DeleteProjects() error = %v", err)
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
	if len(state.Terminals) != 3 {
		t.Fatalf("Terminals = %#v, want all TODO project terminals preserved", state.Terminals)
	}
	if starter.processes[0].closed || starter.processes[2].closed {
		t.Fatal("deleted candidates closed TODO project terminal processes")
	}
	if starter.processes[1].closed {
		t.Fatal("remaining project terminal process was closed")
	}
}

func TestAppDeletesTerminalAndReturnsUpdatedState(t *testing.T) {
	projectDir := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)

	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	state, err = app.CreateTodoTerminal(todoProjectID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTodoTerminal(A) error = %v", err)
	}
	state, err = app.CreateTodoTerminal(todoProjectID, 100, 32)
	if err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
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

func TestAppGetsTodoProjectGitStatusFromPreparedWorktree(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	todoProject := findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
		t.Fatalf("todoProject worktree = %#v, want ready worktree", todoProject)
	}
	app.gitStatus = func(path string) (GitStatus, error) {
		if path != todoProject.WorktreePath {
			t.Fatalf("git status path = %q, want worktree path %q", path, todoProject.WorktreePath)
		}
		if path == projectDir {
			t.Fatalf("git status used source project path %q", path)
		}
		return GitStatus{IsRepo: true, Branch: "todo/fix-login/frontend-app", ChangedCount: 2}, nil
	}

	status, err := app.GetTodoProjectGitStatus(todoProjectID)
	if err != nil {
		t.Fatalf("GetTodoProjectGitStatus() error = %v", err)
	}

	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want source project %q", status.ProjectID, projectID)
	}
	if status.Branch != "todo/fix-login/frontend-app" {
		t.Fatalf("Branch = %q, want worktree branch", status.Branch)
	}
	if status.ChangedCount != 2 {
		t.Fatalf("ChangedCount = %d, want 2", status.ChangedCount)
	}
}

func TestAppGetTodoProjectGitStatusRequiresReadyAvailableWorktree(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for index := range state.TodoProjects {
		if state.TodoProjects[index].ID == todoProjectID {
			state.TodoProjects[index].WorktreeStatus = ""
			state.TodoProjects[index].WorktreePath = ""
			state.TodoProjects[index].WorktreeBranch = ""
		}
	}
	if err := app.projects.saveLocked(state); err != nil {
		t.Fatalf("save todo project without ready worktree: %v", err)
	}
	gitStatusCalls := 0
	app.gitStatus = func(path string) (GitStatus, error) {
		gitStatusCalls++
		return GitStatus{}, nil
	}

	status, err := app.GetTodoProjectGitStatus(todoProjectID)
	if err != nil {
		t.Fatalf("GetTodoProjectGitStatus() error = %v", err)
	}

	if gitStatusCalls != 0 {
		t.Fatalf("git status calls = %d, want 0", gitStatusCalls)
	}
	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want source project %q", status.ProjectID, projectID)
	}
	if !status.PathUnavailable {
		t.Fatal("PathUnavailable = false, want true")
	}
}

func TestAppGetTodoProjectGitStatusRecordsClearedWhenWorktreeBranchIsMissing(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithWorktreePreparer(newReadyWorktreePreparer()),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath() error = %v", err)
	}
	projectID := state.Projects[0].ID
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", projectID)
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	todoProject := findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreeBranch == "" {
		t.Fatalf("todoProject worktree = %#v, want ready branch", todoProject)
	}
	app.gitBranches = func(path string) ([]string, error) {
		if path != todoProject.Path {
			t.Fatalf("git branches path = %q, want source project path %q", path, todoProject.Path)
		}
		return []string{"main", "develop"}, nil
	}
	app.gitStatus = func(path string) (GitStatus, error) {
		t.Fatalf("git status called for cleared worktree path %q", path)
		return GitStatus{}, nil
	}

	status, err := app.GetTodoProjectGitStatus(todoProjectID)
	if err != nil {
		t.Fatalf("GetTodoProjectGitStatus() error = %v", err)
	}

	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want source project %q", status.ProjectID, projectID)
	}
	if !status.WorktreeCleared {
		t.Fatal("WorktreeCleared = false, want true")
	}
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load() after git status error = %v", err)
	}
	todoProject = findTodoProjectByIDForApp(t, state.TodoProjects, todoProjectID)
	if todoProject.WorktreeStatus != WorktreeStatusCleared {
		t.Fatalf("WorktreeStatus = %q, want %q", todoProject.WorktreeStatus, WorktreeStatusCleared)
	}
}

func TestAppGetTodoGitStatusUsesTaskWorkspaceRootRepository(t *testing.T) {
	workspaceDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	if err := os.Mkdir(filepath.Join(taskDir, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir(.git) error = %v", err)
	}
	app.gitStatus = func(path string) (GitStatus, error) {
		if path != taskDir {
			t.Fatalf("git status path = %q, want task workspace root %q", path, taskDir)
		}
		return GitStatus{IsRepo: true, Branch: "todo/root", ChangedCount: 1}, nil
	}

	status, err := app.GetTodoGitStatus(todoID)
	if err != nil {
		t.Fatalf("GetTodoGitStatus() error = %v", err)
	}

	if !status.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if status.Branch != "todo/root" {
		t.Fatalf("Branch = %q, want todo/root", status.Branch)
	}
	if status.ChangedCount != 1 {
		t.Fatalf("ChangedCount = %d, want 1", status.ChangedCount)
	}
}

func TestAppGetTodoGitStatusDoesNotSearchNestedRepositories(t *testing.T) {
	workspaceDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	taskDir := filepath.Join(mustAbs(t, workspaceDir), "tasks", state.Todos[0].WorkspaceDirName)
	if err := os.MkdirAll(filepath.Join(taskDir, "nested-project", ".git"), 0o755); err != nil {
		t.Fatalf("MkdirAll(nested .git) error = %v", err)
	}
	gitStatusCalls := 0
	app.gitStatus = func(path string) (GitStatus, error) {
		gitStatusCalls++
		return GitStatus{}, nil
	}

	status, err := app.GetTodoGitStatus(todoID)
	if err != nil {
		t.Fatalf("GetTodoGitStatus() error = %v", err)
	}

	if gitStatusCalls != 0 {
		t.Fatalf("git status calls = %d, want 0", gitStatusCalls)
	}
	if status.IsRepo {
		t.Fatal("IsRepo = true, want false")
	}
	if status.PathUnavailable {
		t.Fatal("PathUnavailable = true, want false")
	}
}

func TestAppListsProjectBranchesForAvailableProject(t *testing.T) {
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
	app.gitBranches = func(path string) ([]string, error) {
		if path != projectDir {
			t.Fatalf("git branches path = %q, want %q", path, projectDir)
		}
		return []string{"main", "origin/main", "origin/feature/login"}, nil
	}

	branches, err := app.ListProjectBranches(projectID)
	if err != nil {
		t.Fatalf("ListProjectBranches() error = %v", err)
	}

	want := []string{"main", "origin/main", "origin/feature/login"}
	if strings.Join(branches, ",") != strings.Join(want, ",") {
		t.Fatalf("branches = %#v, want %#v", branches, want)
	}
}

func TestAppGetsProjectGitStatusFromTodoProjectCopyAfterCandidateRemoval(t *testing.T) {
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
	if _, err := app.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{projectID}}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := app.DeleteProject(projectID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	app.gitStatus = func(path string) (GitStatus, error) {
		if path != projectDir {
			t.Fatalf("git status path = %q, want %q", path, projectDir)
		}
		return GitStatus{IsRepo: true, Branch: "main"}, nil
	}

	status, err := app.GetProjectGitStatus(projectID)
	if err != nil {
		t.Fatalf("GetProjectGitStatus() error = %v", err)
	}

	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want %q", status.ProjectID, projectID)
	}
	if status.Branch != "main" {
		t.Fatalf("Branch = %q, want main", status.Branch)
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

func TestAppGetProjectGitStatusReturnsGitUnavailableStatus(t *testing.T) {
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
		return GitStatus{GitUnavailable: true}, nil
	}

	status, err := app.GetProjectGitStatus(projectID)
	if err != nil {
		t.Fatalf("GetProjectGitStatus() error = %v", err)
	}

	if status.ProjectID != projectID {
		t.Fatalf("ProjectID = %q, want %q", status.ProjectID, projectID)
	}
	if !status.GitUnavailable {
		t.Fatal("GitUnavailable = false, want true")
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

func TestAppChecksCompletedTodoProjectMergeStatus(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	calls := 0
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		calls++
		if path != "/work/frontend" || worktreeBranch != "todo/fix-login" || baseBranch != "main" {
			t.Fatalf("merge check args = %q/%q/%q, want snapshot args", path, worktreeBranch, baseBranch)
		}
		return true, nil
	}

	statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		{ID: "todo-a:project-a", Path: "/work/frontend", WorktreeBranch: "todo/fix-login", BaseBranch: "main"},
	})
	if err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("merge check calls = %d, want 1", calls)
	}
	if len(statuses) != 1 {
		t.Fatalf("statuses length = %d, want 1", len(statuses))
	}
	if statuses[0].ID != "todo-a:project-a" || statuses[0].Status != CompletedTodoProjectMergeStatusMerged {
		t.Fatalf("status = %#v, want merged result for request", statuses[0])
	}
}

func TestAppCompletedTodoProjectMergeStatusReturnsUnknownForInvalidInputs(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		t.Fatal("gitBranchMerged called for invalid request")
		return false, nil
	}

	statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		{ID: "missing-path", WorktreeBranch: "todo/fix-login", BaseBranch: "main"},
		{ID: "missing-worktree", Path: "/work/frontend", BaseBranch: "main"},
		{ID: "missing-base", Path: "/work/frontend", WorktreeBranch: "todo/fix-login"},
	})
	if err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	for _, status := range statuses {
		if status.Status != CompletedTodoProjectMergeStatusUnknown || status.Reason == "" {
			t.Fatalf("status = %#v, want unknown with reason", status)
		}
	}
}

func TestAppCompletedTodoProjectMergeStatusMapsUnmergedAndErrors(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		switch path {
		case "/work/unmerged":
			return false, nil
		case "/work/error":
			return false, errors.New("branch missing")
		default:
			return true, nil
		}
	}

	statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		{ID: "unmerged", Path: "/work/unmerged", WorktreeBranch: "todo/fix-login", BaseBranch: "main"},
		{ID: "error", Path: "/work/error", WorktreeBranch: "todo/fix-login", BaseBranch: "main"},
	})
	if err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	if statuses[0].Status != CompletedTodoProjectMergeStatusUnmerged {
		t.Fatalf("first status = %#v, want unmerged", statuses[0])
	}
	if statuses[1].Status != CompletedTodoProjectMergeStatusUnknown || statuses[1].Reason == "" {
		t.Fatalf("second status = %#v, want unknown with reason", statuses[1])
	}
}

func TestAppCompletedTodoProjectMergeStatusPersistsRemovedWorktreePath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	app := NewAppWithConfigAndShellStarter(
		configPath,
		newFakeShellStarter().Start,
	)
	writeProjectStateForTest(t, configPath, ProjectState{
		Version: projectConfigVersion,
		Todos: []Todo{completedTodoWithSnapshotForTest(TodoProjectSnapshot{
			ProjectID:      "project-a",
			Name:           "frontend",
			Path:           "/work/frontend",
			WorktreeBranch: "todo/fix-login",
			BaseBranch:     "main",
		})},
	})
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		return false, errGitWorktreePathMissing
	}

	statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "main"),
	})
	if err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	if len(statuses) != 1 || statuses[0].Status != CompletedTodoProjectMergeStatusMerged {
		t.Fatalf("statuses = %#v, want merged status for removed worktree path", statuses)
	}
	state, err := app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	snapshot := state.Todos[0].ProjectSnapshots[0]
	if snapshot.MergeStatus != CompletedTodoProjectMergeStatusConfirmed ||
		snapshot.MergeStatusReason != CompletedTodoProjectMergeStatusReasonWorktreeRemoved {
		t.Fatalf("snapshot merge status = %q/%q, want confirmed/worktree removed", snapshot.MergeStatus, snapshot.MergeStatusReason)
	}
}

func TestAppCompletedTodoProjectMergeStatusPersistsRemovedWorktreeBranch(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	app := NewAppWithConfigAndShellStarter(
		configPath,
		newFakeShellStarter().Start,
	)
	writeProjectStateForTest(t, configPath, ProjectState{
		Version: projectConfigVersion,
		Todos: []Todo{completedTodoWithSnapshotForTest(TodoProjectSnapshot{
			ProjectID:      "project-a",
			Name:           "frontend",
			Path:           "/work/frontend",
			WorktreeBranch: "todo/fix-login",
			BaseBranch:     "main",
		})},
	})
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		return false, errGitWorktreeBranchMissing
	}

	statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "main"),
	})
	if err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	if len(statuses) != 1 || statuses[0].Status != CompletedTodoProjectMergeStatusMerged {
		t.Fatalf("statuses = %#v, want merged status for removed worktree branch", statuses)
	}
	state, err := app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	snapshot := state.Todos[0].ProjectSnapshots[0]
	if snapshot.MergeStatus != CompletedTodoProjectMergeStatusConfirmed ||
		snapshot.MergeStatusReason != CompletedTodoProjectMergeStatusReasonWorktreeBranchRemoved {
		t.Fatalf("snapshot merge status = %q/%q, want confirmed/worktree branch removed", snapshot.MergeStatus, snapshot.MergeStatusReason)
	}
}

func TestAppCompletedTodoProjectMergeStatusPersistsOnlyConfirmedResults(t *testing.T) {
	tests := []struct {
		name        string
		merged      bool
		mergeErr    error
		wantStatus  string
		wantReason  string
		wantPersist bool
	}{
		{
			name:        "merged persists",
			merged:      true,
			wantStatus:  CompletedTodoProjectMergeStatusMerged,
			wantReason:  CompletedTodoProjectMergeStatusReasonMerged,
			wantPersist: true,
		},
		{
			name:       "unmerged does not persist",
			merged:     false,
			wantStatus: CompletedTodoProjectMergeStatusUnmerged,
		},
		{
			name:       "unknown does not persist",
			mergeErr:   errors.New("git failed"),
			wantStatus: CompletedTodoProjectMergeStatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "projects.json")
			app := NewAppWithConfigAndShellStarter(
				configPath,
				newFakeShellStarter().Start,
			)
			writeProjectStateForTest(t, configPath, ProjectState{
				Version: projectConfigVersion,
				Todos: []Todo{completedTodoWithSnapshotForTest(TodoProjectSnapshot{
					ProjectID:      "project-a",
					Name:           "frontend",
					Path:           "/work/frontend",
					WorktreeBranch: "todo/fix-login",
					BaseBranch:     "main",
				})},
			})
			app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
				return tt.merged, tt.mergeErr
			}

			statuses, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
				completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "main"),
			})
			if err != nil {
				t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
			}
			if len(statuses) != 1 || statuses[0].Status != tt.wantStatus {
				t.Fatalf("statuses = %#v, want %s", statuses, tt.wantStatus)
			}

			state, err := app.projects.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			snapshot := state.Todos[0].ProjectSnapshots[0]
			if tt.wantPersist {
				if snapshot.MergeStatus != CompletedTodoProjectMergeStatusConfirmed || snapshot.MergeStatusReason != tt.wantReason {
					t.Fatalf("snapshot merge status = %q/%q, want confirmed/%s", snapshot.MergeStatus, snapshot.MergeStatusReason, tt.wantReason)
				}
			} else if snapshot.MergeStatus != "" || snapshot.MergeStatusReason != "" {
				t.Fatalf("snapshot merge status = %q/%q, want empty", snapshot.MergeStatus, snapshot.MergeStatusReason)
			}
		})
	}
}

func TestAppCompletedTodoProjectMergeStatusSkipsStaleOrNonCompletedWriteback(t *testing.T) {
	tests := []struct {
		name    string
		todo    Todo
		request CompletedTodoProjectMergeStatusRequest
	}{
		{
			name: "fingerprint mismatch",
			todo: completedTodoWithSnapshotForTest(TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend",
				Path:           "/work/frontend",
				WorktreeBranch: "todo/fix-login",
				BaseBranch:     "main",
			}),
			request: completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "release"),
		},
		{
			name: "todo not completed",
			todo: Todo{
				ID:       "todo-a",
				Title:    "修复登录问题",
				Priority: TodoPriorityMedium,
				Status:   TodoStatusInProgress,
				ProjectSnapshots: []TodoProjectSnapshot{{
					ProjectID:      "project-a",
					Name:           "frontend",
					Path:           "/work/frontend",
					WorktreeBranch: "todo/fix-login",
					BaseBranch:     "main",
				}},
				CreatedAt: "2026-06-01T00:00:00Z",
			},
			request: completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "main"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "projects.json")
			app := NewAppWithConfigAndShellStarter(
				configPath,
				newFakeShellStarter().Start,
			)
			writeProjectStateForTest(t, configPath, ProjectState{
				Version: projectConfigVersion,
				Todos:   []Todo{tt.todo},
			})
			app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
				return true, nil
			}

			if _, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{tt.request}); err != nil {
				t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
			}
			state, err := app.projects.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			snapshot := state.Todos[0].ProjectSnapshots[0]
			if snapshot.MergeStatus != "" || snapshot.MergeStatusReason != "" {
				t.Fatalf("snapshot merge status = %q/%q, want empty", snapshot.MergeStatus, snapshot.MergeStatusReason)
			}
		})
	}
}

func TestAppCompletedTodoProjectMergeStatusPersistsMatchingCompletedTodoAfterOtherTodos(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	app := NewAppWithConfigAndShellStarter(
		configPath,
		newFakeShellStarter().Start,
	)
	matching := completedTodoWithSnapshotForTest(TodoProjectSnapshot{
		ProjectID:      "project-a",
		Name:           "frontend",
		Path:           "/work/frontend",
		WorktreeBranch: "todo/fix-login",
		BaseBranch:     "main",
	})
	writeProjectStateForTest(t, configPath, ProjectState{
		Version: projectConfigVersion,
		Todos: []Todo{
			{
				ID:        "todo-other",
				Title:     "其它任务",
				Priority:  TodoPriorityMedium,
				Status:    TodoStatusCompleted,
				CreatedAt: "2026-06-01T00:00:00Z",
			},
			matching,
		},
	})
	app.gitBranchMerged = func(path string, worktreeBranch string, baseBranch string) (bool, error) {
		return true, nil
	}

	if _, err := app.GetCompletedTodoProjectMergeStatuses([]CompletedTodoProjectMergeStatusRequest{
		completedMergeStatusRequestForTest("todo-a", 0, "/work/frontend", "todo/fix-login", "main"),
	}); err != nil {
		t.Fatalf("GetCompletedTodoProjectMergeStatuses() error = %v", err)
	}

	state, err := app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	snapshot := state.Todos[1].ProjectSnapshots[0]
	if snapshot.MergeStatus != CompletedTodoProjectMergeStatusConfirmed ||
		snapshot.MergeStatusReason != CompletedTodoProjectMergeStatusReasonMerged {
		t.Fatalf("snapshot merge status = %q/%q, want confirmed/merged", snapshot.MergeStatus, snapshot.MergeStatusReason)
	}
}

func completedTodoWithSnapshotForTest(snapshot TodoProjectSnapshot) Todo {
	return Todo{
		ID:               "todo-a",
		Title:            "修复登录问题",
		Priority:         TodoPriorityMedium,
		Status:           TodoStatusCompleted,
		ProjectSnapshots: []TodoProjectSnapshot{snapshot},
		CreatedAt:        "2026-06-01T00:00:00Z",
		CompletedAt:      "2026-06-02T00:00:00Z",
		ArchivedAt:       "2026-06-02T00:00:00Z",
	}
}

func completedMergeStatusRequestForTest(todoID string, snapshotIndex int, path string, worktreeBranch string, baseBranch string) CompletedTodoProjectMergeStatusRequest {
	return CompletedTodoProjectMergeStatusRequest{
		ID:             "todo-a::project-a::" + path + "::0",
		TodoID:         todoID,
		SnapshotIndex:  snapshotIndex,
		Path:           path,
		WorktreeBranch: worktreeBranch,
		BaseBranch:     baseBranch,
		Fingerprint:    completedMergeStatusFingerprintForTest(path, worktreeBranch, baseBranch),
	}
}

func completedMergeStatusFingerprintForTest(path string, worktreeBranch string, baseBranch string) string {
	return strings.Join([]string{path, worktreeBranch, baseBranch}, "::")
}

func writeProjectStateForTest(t *testing.T, configPath string, state ProjectState) {
	t.Helper()
	if state.Projects == nil {
		state.Projects = []Project{}
	}
	if state.TodoProjects == nil {
		state.TodoProjects = []TodoProject{}
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestAppInitializesProjectGitRepositoryForAvailableProject(t *testing.T) {
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
	calls := 0
	app.gitInit = func(path string) error {
		calls++
		if path != projectDir {
			t.Fatalf("git init path = %q, want %q", path, projectDir)
		}
		return nil
	}

	if err := app.InitializeProjectGitRepository(projectID); err != nil {
		t.Fatalf("InitializeProjectGitRepository() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("git init calls = %d, want 1", calls)
	}
}

func TestAppInitializeProjectGitRepositoryDoesNotInitializeUnavailableProject(t *testing.T) {
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
	calls := 0
	app.gitInit = func(path string) error {
		calls++
		return nil
	}

	if err := app.InitializeProjectGitRepository(projectID); err == nil {
		t.Fatal("InitializeProjectGitRepository() error = nil, want error")
	}
	if calls != 0 {
		t.Fatalf("git init calls = %d, want 0", calls)
	}
}

func TestAppInitializeProjectGitRepositoryReturnsErrorWhenProjectIsMissing(t *testing.T) {
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	if err := app.InitializeProjectGitRepository("missing-project"); err == nil {
		t.Fatal("InitializeProjectGitRepository() error = nil, want error")
	}
}

func TestAppInitializeProjectGitRepositoryReturnsGitInitError(t *testing.T) {
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
	app.gitInit = func(path string) error {
		return os.ErrPermission
	}

	if err := app.InitializeProjectGitRepository(projectID); err == nil {
		t.Fatal("InitializeProjectGitRepository() error = nil, want error")
	}
}

func TestAppInitializeProjectGitRepositoryRemovesCreatedMetadataWhenInitializationFails(t *testing.T) {
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
	app.gitInit = func(path string) error {
		if err := os.Mkdir(filepath.Join(path, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir .git: %v", err)
		}
		return errors.New("git commit failed: Author identity unknown")
	}

	if err := app.InitializeProjectGitRepository(projectID); err == nil {
		t.Fatal("InitializeProjectGitRepository() error = nil, want commit failure")
	}
	if pathHasGitRepositoryMetadata(projectDir) {
		t.Fatal("pathHasGitRepositoryMetadata() = true, want failed initialization cleanup")
	}
}

func createTodoProjectForApp(t *testing.T, app *App, title string, projectID string) (string, string) {
	t.Helper()

	state, err := app.CreateTodo(CreateTodoRequest{Title: title})
	if err != nil {
		t.Fatalf("CreateTodo(%q) error = %v", title, err)
	}
	todoID := state.Todos[len(state.Todos)-1].ID
	state, err = app.AddProjectToTodo(todoID, projectID)
	if err != nil {
		t.Fatalf("AddProjectToTodo(%q, %q) error = %v", todoID, projectID, err)
	}
	if state.ActiveTodoProjectID == "" {
		t.Fatal("ActiveTodoProjectID = empty, want associated todo project")
	}
	if _, err := app.ChangeTodoStatus(todoID, "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(%q) error = %v", todoID, err)
	}
	return todoID, state.ActiveTodoProjectID
}

func findProjectByPathForApp(t *testing.T, projects []Project, path string) Project {
	t.Helper()
	absolutePath := mustAbs(t, path)
	for _, project := range projects {
		if project.Path == absolutePath {
			return project
		}
	}
	t.Fatalf("project path %q not found in %#v", absolutePath, projects)
	return Project{}
}

func findTodoProjectByIDForApp(t *testing.T, todoProjects []TodoProject, id string) TodoProject {
	t.Helper()
	for _, todoProject := range todoProjects {
		if todoProject.ID == id {
			return todoProject
		}
	}
	t.Fatalf("todo project %q not found in %#v", id, todoProjects)
	return TodoProject{}
}

func findTerminalByID(terminals []ProjectTerminal, terminalID string) ProjectTerminal {
	for _, terminal := range terminals {
		if terminal.ID == terminalID {
			return terminal
		}
	}
	return ProjectTerminal{}
}

type fakeBackgroundCommandRunner struct {
	requests []BackgroundCommandRequest
	err      error
}

func newFakeBackgroundCommandRunner() *fakeBackgroundCommandRunner {
	return &fakeBackgroundCommandRunner{}
}

func (runner *fakeBackgroundCommandRunner) Start(request BackgroundCommandRequest) error {
	runner.requests = append(runner.requests, request)
	return runner.err
}

func receiveLifecycleScriptRequest(t *testing.T, requests <-chan TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunRequest {
	t.Helper()
	select {
	case request := <-requests:
		return request
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for lifecycle script request")
		return TodoLifecycleScriptRunRequest{}
	}
}

func waitForAppLifecycleScriptStatus(t *testing.T, app *App, todoID string, phase string, want string) ProjectState {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		state, err := app.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}
		status := lifecycleScriptStatusByPhase(state.LifecycleScriptStatuses, todoID, phase)
		if status != nil && status.Status == want {
			return state
		}
		time.Sleep(time.Millisecond)
	}
	state, _ := app.ListProjects()
	t.Fatalf("LifecycleScriptStatuses = %#v, want %s for %s/%s", state.LifecycleScriptStatuses, want, todoID, phase)
	return ProjectState{}
}

func waitForAppTodoStatus(t *testing.T, app *App, todoID string, want string) ProjectState {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		state, err := app.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}
		todo := findTodo(state.Todos, todoID)
		if todo != nil && todo.Status == want {
			return state
		}
		time.Sleep(time.Millisecond)
	}
	state, _ := app.ListProjects()
	t.Fatalf("Todos = %#v, want todo %s status %s", state.Todos, todoID, want)
	return ProjectState{}
}

func lifecycleScriptStatusByPhase(statuses []TodoLifecycleScriptStatus, todoID string, phase string) *TodoLifecycleScriptStatus {
	for index := range statuses {
		if statuses[index].TodoID == todoID && statuses[index].Phase == phase {
			return &statuses[index]
		}
	}
	return nil
}
