package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppResolvesTodoFromTaskFolderAndChildDirectory(t *testing.T) {
	app, workspaceDir, todoID, taskDir := newAppWithTodoTaskDirForIPCTest(t, TodoStatusNotStarted)

	gotTodoID, err := app.todoIDForTaskWorkingDir(taskDir)
	if err != nil {
		t.Fatalf("todoIDForTaskWorkingDir(taskDir) error = %v", err)
	}
	if gotTodoID != todoID {
		t.Fatalf("todoID = %q, want %q", gotTodoID, todoID)
	}

	childDir := filepath.Join(taskDir, "src", "nested")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(childDir) error = %v", err)
	}
	gotTodoID, err = app.todoIDForTaskWorkingDir(childDir)
	if err != nil {
		t.Fatalf("todoIDForTaskWorkingDir(childDir) error = %v", err)
	}
	if gotTodoID != todoID {
		t.Fatalf("todoID = %q, want %q", gotTodoID, todoID)
	}

	otherWorkspaceTaskDir := filepath.Join(t.TempDir(), todoWorkspaceRootDirName, filepath.Base(taskDir))
	if _, err := app.todoIDForTaskWorkingDir(otherWorkspaceTaskDir); err == nil {
		t.Fatal("todoIDForTaskWorkingDir(other workspace) error = nil, want error")
	}
	if _, err := app.todoIDForTaskWorkingDir(filepath.Join(workspaceDir, "not-a-task")); err == nil {
		t.Fatal("todoIDForTaskWorkingDir(unknown) error = nil, want error")
	}
}

func TestAppTodoIPCStartUsesButtonLogicAndEmitsWorkspaceState(t *testing.T) {
	emitted := make(chan ProjectState, 1)
	app, _, todoID, taskDir := newAppWithTodoTaskDirForIPCTest(
		t,
		TodoStatusNotStarted,
		WithWorkspaceStateEmitter(func(state ProjectState) {
			emitted <- state
		}),
	)
	drainWorkspaceStatesForIPCTest(emitted)

	if err := app.handleTodoIPCCommand(context.Background(), todoIPCCommandRequest{Command: "start", WorkingDir: taskDir}); err != nil {
		t.Fatalf("handleTodoIPCCommand(start) error = %v", err)
	}

	state, err := app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if todo := findTodo(state.Todos, todoID); todo == nil || todo.Status != TodoStatusInProgress {
		t.Fatalf("todo after start = %#v, want in-progress", todo)
	}
	emittedState := receiveWorkspaceStateForIPCTest(t, emitted)
	if todo := findTodo(emittedState.Todos, todoID); todo == nil || todo.Status != TodoStatusInProgress {
		t.Fatalf("emitted todo after start = %#v, want in-progress", todo)
	}
}

func TestAppTodoIPCDoneUsesButtonLogicAndEmitsWorkspaceState(t *testing.T) {
	emitted := make(chan ProjectState, 1)
	app, _, todoID, taskDir := newAppWithTodoTaskDirForIPCTest(
		t,
		TodoStatusInProgress,
		WithWorkspaceStateEmitter(func(state ProjectState) {
			emitted <- state
		}),
	)
	drainWorkspaceStatesForIPCTest(emitted)

	if err := app.handleTodoIPCCommand(context.Background(), todoIPCCommandRequest{Command: "done", WorkingDir: taskDir}); err != nil {
		t.Fatalf("handleTodoIPCCommand(done) error = %v", err)
	}

	state, err := app.projects.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if todo := findTodo(state.Todos, todoID); todo == nil || todo.Status != TodoStatusCompleted {
		t.Fatalf("todo after done = %#v, want completed", todo)
	}
	emittedState := receiveWorkspaceStateForIPCTest(t, emitted)
	if todo := findTodo(emittedState.Todos, todoID); todo == nil || todo.Status != TodoStatusCompleted {
		t.Fatalf("emitted todo after done = %#v, want completed", todo)
	}
}

func TestAppTodoIPCRejectsInvalidStateWithoutMutatingTodo(t *testing.T) {
	app, _, todoID, taskDir := newAppWithTodoTaskDirForIPCTest(t, TodoStatusNotStarted)

	err := app.handleTodoIPCCommand(context.Background(), todoIPCCommandRequest{Command: "done", WorkingDir: taskDir})
	if err == nil {
		t.Fatal("handleTodoIPCCommand(done not-started) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "todo not found") {
		t.Fatalf("error = %q, want todo not found", err.Error())
	}
	state, loadErr := app.projects.Load()
	if loadErr != nil {
		t.Fatalf("Load() error = %v", loadErr)
	}
	if todo := findTodo(state.Todos, todoID); todo == nil || todo.Status != TodoStatusNotStarted {
		t.Fatalf("todo after rejected done = %#v, want not-started", todo)
	}
}

func TestAppStartupStartsTodoIPCServerAndShutdownRemovesRuntimeFile(t *testing.T) {
	appConfigDir := t.TempDir()
	emitted := make(chan ProjectState, 2)
	workspaceDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(appConfigDir, "projects.json"),
		newFakeShellStarter().Start,
		WithClaudeStatusDir(""),
		WithWorkspaceStateEmitter(func(state ProjectState) {
			emitted <- state
		}),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.CreateTodo(CreateTodoRequest{Title: "实现登录"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	if _, err := app.projects.RecordTodoWorkspace(todoID, "task-"+todoID, nil); err != nil {
		t.Fatalf("RecordTodoWorkspace() error = %v", err)
	}
	taskDir := filepath.Join(workspaceDir, todoWorkspaceRootDirName, "task-"+todoID)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(taskDir) error = %v", err)
	}

	app.startup(nil)
	drainWorkspaceStatesForIPCTest(emitted)

	if _, err := readTodoIPCRuntimeFile(todoIPCRuntimePath(appConfigDir)); err != nil {
		t.Fatalf("runtime file after startup error = %v", err)
	}
	if err := sendTodoIPCCommand(context.Background(), appConfigDir, "start", taskDir); err != nil {
		t.Fatalf("sendTodoIPCCommand(start) error = %v", err)
	}
	emittedState := receiveWorkspaceStateForIPCTest(t, emitted)
	if todo := findTodo(emittedState.Todos, todoID); todo == nil || todo.Status != TodoStatusInProgress {
		t.Fatalf("emitted todo after IPC start = %#v, want in-progress", todo)
	}

	app.shutdown(nil)
	if _, err := os.Stat(todoIPCRuntimePath(appConfigDir)); !os.IsNotExist(err) {
		t.Fatalf("runtime file after shutdown stat error = %v, want not exist", err)
	}
}

func newAppWithTodoTaskDirForIPCTest(t *testing.T, status string, opts ...AppOption) (*App, string, string, string) {
	t.Helper()
	workspaceDir := t.TempDir()
	appOpts := []any{
		WithWorktreePreparer(newReadyWorktreePreparer()),
	}
	for _, opt := range opts {
		appOpts = append(appOpts, opt)
	}
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		appOpts...,
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath() error = %v", err)
	}
	state, err := app.CreateTodo(CreateTodoRequest{Title: "实现登录"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	todoID := state.Todos[0].ID
	if status == TodoStatusInProgress {
		state, err = app.ChangeTodoStatus(todoID, TodoStatusInProgress)
		if err != nil {
			t.Fatalf("ChangeTodoStatus() error = %v", err)
		}
	} else if status != TodoStatusNotStarted {
		t.Fatalf("unsupported setup status %q", status)
	}
	state, err = app.projects.RecordTodoWorkspace(todoID, "task-"+todoID, nil)
	if err != nil {
		t.Fatalf("RecordTodoWorkspace() error = %v", err)
	}
	taskDir := filepath.Join(workspaceDir, todoWorkspaceRootDirName, "task-"+todoID)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(taskDir) error = %v", err)
	}
	return app, workspaceDir, todoID, taskDir
}

func receiveWorkspaceStateForIPCTest(t *testing.T, states <-chan ProjectState) ProjectState {
	t.Helper()
	select {
	case state := <-states:
		return state
	default:
		t.Fatal("workspace state was not emitted")
		return ProjectState{}
	}
}

func drainWorkspaceStatesForIPCTest(states <-chan ProjectState) {
	for {
		select {
		case <-states:
		default:
			return
		}
	}
}
