package main

import (
	"os"
	"path/filepath"
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

func TestAppCreatesAndSelectsTodoProjectTerminals(t *testing.T) {
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

	if err := os.WriteFile(filepath.Join(statusDir, "session-a.status"), []byte(`{"session":"session-a","terminalId":"terminal-1","status":"waiting","event":"Notification","cwd":"`+projectDir+`","ts":1718450010}`), 0644); err != nil {
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

func TestAppImportsProjectsFromParentDirectory(t *testing.T) {
	parentDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(parentDir, "frontend-app"), 0o755); err != nil {
		t.Fatalf("mkdir frontend-app: %v", err)
	}
	if err := os.Mkdir(filepath.Join(parentDir, "api-service"), 0o755); err != nil {
		t.Fatalf("mkdir api-service: %v", err)
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

func TestAppSavesTerminalThemeWithoutChangingProjectShellBehavior(t *testing.T) {
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
	projectBID := state.ActiveProjectID
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

func TestAppDeletesProjectsAndOwnedTerminals(t *testing.T) {
	projectDirA := t.TempDir()
	projectDirB := t.TempDir()
	projectDirC := t.TempDir()
	starter := newFakeShellStarter()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		starter.Start,
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
	projectBID := state.ActiveProjectID
	_, todoProjectBID := createTodoProjectForApp(t, app, "升级依赖", projectBID)
	if _, err := app.CreateTodoTerminal(todoProjectBID, 80, 24); err != nil {
		t.Fatalf("CreateTodoTerminal(B) error = %v", err)
	}

	state, err = app.AddProjectFromPath(projectDirC)
	if err != nil {
		t.Fatalf("AddProjectFromPath(C) error = %v", err)
	}
	projectCID := state.ActiveProjectID
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
	if len(state.Terminals) != 1 || state.Terminals[0].ID != "terminal-b1" {
		t.Fatalf("Terminals = %#v, want only terminal-b1", state.Terminals)
	}
	if !starter.processes[0].closed || !starter.processes[2].closed {
		t.Fatal("deleted project terminal processes were not closed")
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
