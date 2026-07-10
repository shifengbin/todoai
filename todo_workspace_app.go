package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// todoProjectWorktreeName returns the display name used for a TODO project's
// worktree directory and branch naming.
func todoProjectWorktreeName(todoProject TodoProject) string {
	name := strings.TrimSpace(todoProject.Name)
	if name == "" && strings.TrimSpace(todoProject.Path) != "" {
		name = filepath.Base(todoProject.Path)
	}
	if name == "" || name == "." || name == string(os.PathSeparator) {
		name = todoProject.ProjectID
	}
	return name
}

// todoProjectsForTodoID returns the todo projects linked to a TODO, preserving
// their stored order.
func todoProjectsForTodoID(todoProjects []TodoProject, todoID string) []TodoProject {
	var matches []TodoProject
	for _, todoProject := range todoProjects {
		if todoProject.TodoID == todoID {
			matches = append(matches, todoProject)
		}
	}
	return matches
}

func todoWorkspaceInitializationReady(todoProjects []TodoProject, todoID string) bool {
	matches := todoProjectsForTodoID(todoProjects, todoID)
	if len(matches) == 0 {
		return true
	}
	for _, todoProject := range matches {
		if todoProject.WorktreeStatus != WorktreeStatusReady {
			return false
		}
	}
	return true
}

func writeTodoWorkspaceInitializationFilesWhenReady(todo Todo, todoProjects []TodoProject, workspacePath string) {
	if todoWorkspaceInitializationReady(todoProjects, todo.ID) {
		_ = writeTodoWorkspaceInitializationFiles(todo, workspacePath)
	}
}

func todoNeedsTaskWorkspace(todo Todo, todoProjects []TodoProject) bool {
	if len(todoProjects) > 0 || len(todo.InitializationFiles) > 0 {
		return true
	}
	if todo.LifecycleScript == nil {
		return false
	}
	return strings.TrimSpace(todo.LifecycleScript.InitScript) != "" || strings.TrimSpace(todo.LifecycleScript.CompleteScript) != ""
}

// workspacePathOrEmpty returns the current workspace root path or an empty
// string when no workspace is bound.
func (a *App) workspacePathOrEmpty() string {
	workspace := a.workspace.CurrentWorkspace()
	if workspace == nil {
		return ""
	}
	return workspace.Path
}

// loadTodoForWorkspace resolves the current state, workspace root path and the
// TODO for a workspace-scoped action. It returns ok=false when the workspace is
// unavailable or the TODO does not exist.
func (a *App) loadTodoForWorkspace(todoID string) (string, ProjectState, Todo, bool) {
	workspacePath := a.workspacePathOrEmpty()
	if workspacePath == "" {
		return "", ProjectState{}, Todo{}, false
	}
	state, err := a.projects.Load()
	if err != nil {
		return "", ProjectState{}, Todo{}, false
	}
	todo, ok := openTodoByID(state.Todos, todoID)
	if !ok {
		return "", ProjectState{}, Todo{}, false
	}
	return workspacePath, state, todo, true
}

// ensureTaskWorkspaceDir resolves the on-disk task workspace directory for a
// TODO. When the TODO has a persisted directory name the path is reused (and
// recreated on disk if missing). When the name is unset it is only computed and
// persisted for an in-progress TODO; a non-in-progress TODO without a persisted
// directory yields an error so we never create task folders for tasks that have
// not entered progress.
func (a *App) ensureTaskWorkspaceDir(todo Todo, workspacePath string) (string, error) {
	if workspacePath == "" {
		return "", errors.New("workspace is not available")
	}
	if todo.WorkspaceDirName == "" {
		if todo.Status != TodoStatusInProgress {
			return "", errors.New("task workspace has not been created")
		}
		state, err := a.projects.Load()
		if err != nil {
			return "", err
		}
		for _, candidate := range state.Todos {
			if candidate.ID == todo.ID {
				todo = candidate
				break
			}
		}
		if !todoNeedsTaskWorkspace(todo, todoProjectsForTodoID(state.TodoProjects, todo.ID)) {
			return "", errors.New("task workspace has not been created")
		}
		dirName := todoWorkspaceDirName(todo.Title, todo.Description)
		state, err = a.projects.RecordTodoWorkspace(todo.ID, dirName, nil)
		if err != nil {
			return "", err
		}
		for _, candidate := range state.Todos {
			if candidate.ID == todo.ID {
				todo = candidate
				break
			}
		}
	}
	taskDir, _ := todoWorkspacePath(todo, workspacePath)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return "", err
	}
	return taskDir, nil
}

// prepareTodoWorkspace ensures the task workspace directory exists for an
// in-progress TODO and prepares Git worktrees for any linked projects that
// have not been prepared yet (empty worktree status). The directory name is
// persisted on the TODO and each project's worktree result is recorded, then
// the README is regenerated. The operation is idempotent: already-ready
// projects are skipped and an existing task directory is reused.
func (a *App) prepareTodoWorkspace(todoID string) {
	workspacePath := a.workspacePathOrEmpty()
	if workspacePath == "" {
		return
	}
	state, err := a.projects.Load()
	if err != nil {
		return
	}
	todo, ok := openTodoByID(state.Todos, todoID)
	if !ok || todo.Status != TodoStatusInProgress {
		return
	}
	todoProjects := todoProjectsForTodoID(state.TodoProjects, todoID)
	if !todoNeedsTaskWorkspace(todo, todoProjects) {
		return
	}
	if len(todoProjects) > 0 && a.worktreePreparer == nil {
		return
	}

	dirName := todo.WorkspaceDirName
	if dirName == "" {
		dirName = todoWorkspaceDirName(todo.Title, todo.Description)
	}
	taskDir := filepath.Join(todoWorkspaceRootPath(workspacePath), dirName)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return
	}

	updates := []TodoProjectWorktreeUpdate{}
	for _, todoProject := range todoProjects {
		if todoProject.WorktreeStatus == WorktreeStatusReady {
			continue
		}
		result := a.worktreePreparer.PrepareWorktree(
			todoProject.Path,
			todoProject.BaseBranch,
			todoProjectWorktreeName(todoProject),
			taskDir,
		)
		updates = append(updates, TodoProjectWorktreeUpdate{
			TodoProjectID:  todoProject.ID,
			BaseBranch:     result.BaseBranch,
			WorktreeBranch: result.WorktreeBranch,
			WorktreePath:   result.WorktreePath,
			WorktreeStatus: result.Status,
			WorktreeError:  result.Error,
		})
	}

	persistedState, err := a.projects.RecordTodoWorkspace(todoID, dirName, updates)
	if err != nil {
		return
	}
	for _, candidate := range persistedState.Todos {
		if candidate.ID == todoID {
			writeTodoWorkspaceInitializationFilesWhenReady(candidate, persistedState.TodoProjects, workspacePath)
			break
		}
	}
	a.writeTodoReadmeFromState(persistedState, todoID, workspacePath)
}

// refreshTodoReadme regenerates the README for a TODO's task workspace from the
// current persisted state. It is a no-op when the TODO has no task workspace
// directory yet.
func (a *App) refreshTodoReadme(todoID string) {
	workspacePath := a.workspacePathOrEmpty()
	if workspacePath == "" {
		return
	}
	state, err := a.projects.Load()
	if err != nil {
		return
	}
	a.writeTodoReadmeFromState(state, todoID, workspacePath)
}

func (a *App) refreshTodoReadmeAfterProjectRemoval(todoID string) {
	workspacePath := a.workspacePathOrEmpty()
	if workspacePath == "" {
		return
	}
	state, err := a.projects.Load()
	if err != nil {
		return
	}
	if len(todoProjectsForTodoID(state.TodoProjects, todoID)) == 0 {
		a.removeTodoReadmeFromState(state, todoID, workspacePath)
		return
	}
	a.writeTodoReadmeFromState(state, todoID, workspacePath)
}

// writeTodoReadmeFromState regenerates the README for a TODO using the supplied
// state snapshot. Projects are listed in their stored order.
func (a *App) writeTodoReadmeFromState(state ProjectState, todoID string, workspacePath string) {
	var todo Todo
	found := false
	for _, candidate := range state.Todos {
		if candidate.ID == todoID {
			todo = candidate
			found = true
			break
		}
	}
	if !found || todo.WorkspaceDirName == "" {
		return
	}
	writeTodoWorkspaceInitializationFilesWhenReady(todo, state.TodoProjects, workspacePath)
	todoProjects := todoProjectsForTodoID(state.TodoProjects, todoID)
	if len(todoProjects) == 0 {
		return
	}
	_ = writeTodoWorkspaceReadme(todo, todoProjects, workspacePath)
}

func (a *App) removeTodoReadmeFromState(state ProjectState, todoID string, workspacePath string) {
	for _, todo := range state.Todos {
		if todo.ID == todoID {
			removeTodoWorkspaceReadme(todo, workspacePath)
			return
		}
	}
}
