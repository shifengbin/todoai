package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) handleTodoIPCCommand(ctx context.Context, request todoIPCCommandRequest) error {
	_ = ctx
	command := strings.TrimSpace(request.Command)
	todoID, err := a.todoIDForTaskWorkingDir(request.WorkingDir)
	if err != nil {
		return err
	}
	var state ProjectState
	switch command {
	case "start":
		state, err = a.ChangeTodoStatus(todoID, TodoStatusInProgress)
	case "done":
		state, err = a.CompleteTodo(todoID)
	default:
		return errors.New("unsupported ipc command")
	}
	if err != nil {
		return todoIPCCommandError(err)
	}
	a.emitWorkspaceState(state)
	return nil
}

func todoIPCCommandError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("todo not found")
	}
	return err
}

func (a *App) todoIDForTaskWorkingDir(workingDir string) (string, error) {
	if !a.hasWorkspace() {
		return "", ErrWorkspaceRequired
	}
	normalizedWorkingDir, err := normalizeWorkspacePath(workingDir)
	if err != nil {
		return "", err
	}
	workspace := a.workspace.CurrentWorkspace()
	if workspace == nil {
		return "", ErrWorkspaceRequired
	}
	state, err := a.projects.Load()
	if err != nil {
		return "", err
	}
	for _, todo := range state.Todos {
		if strings.TrimSpace(todo.WorkspaceDirName) == "" {
			continue
		}
		taskDir := filepath.Join(todoWorkspaceRootPath(workspace.Path), todo.WorkspaceDirName)
		if pathContains(taskDir, normalizedWorkingDir) {
			return todo.ID, nil
		}
	}
	return "", errors.New("unable to locate current todo task")
}

func (a *App) startTodoIPCServer() {
	if a.todoIPCServer != nil {
		return
	}
	server := newTodoIPCServer(filepath.Dir(a.projectConfigPath), a.handleTodoIPCCommand)
	if err := server.Start(context.Background()); err != nil {
		return
	}
	a.todoIPCServer = server
}

func (a *App) stopTodoIPCServer() {
	if a.todoIPCServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = a.todoIPCServer.Stop(ctx)
	a.todoIPCServer = nil
}
