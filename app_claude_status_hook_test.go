package main

import (
	"path/filepath"
	"testing"
)

func TestApp_EnsureClaudeStatusHook_InstallThenRemove(t *testing.T) {
	setHookExe(t, filepath.Join(t.TempDir(), "todoai.exe"))
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
	)

	if err := app.EnsureClaudeStatusHook(projectDir); err != nil {
		t.Fatalf("EnsureClaudeStatusHook: %v", err)
	}
	state, err := app.ClaudeStatusHookState(projectDir)
	if err != nil {
		t.Fatalf("ClaudeStatusHookState: %v", err)
	}
	if !state.Installed {
		t.Fatalf("expected installed after EnsureClaudeStatusHook, got %#v", state)
	}

	if err := app.RemoveClaudeStatusHook(projectDir); err != nil {
		t.Fatalf("RemoveClaudeStatusHook: %v", err)
	}
	state, err = app.ClaudeStatusHookState(projectDir)
	if err != nil {
		t.Fatalf("ClaudeStatusHookState after remove: %v", err)
	}
	if state.Installed {
		t.Fatalf("expected not installed after RemoveClaudeStatusHook, got %#v", state)
	}
}

func TestApp_EnsureClaudeStatusHooksForActiveWorkspace_InstallsAllProjects(t *testing.T) {
	setHookExe(t, filepath.Join(t.TempDir(), "todoai.exe"))
	workspaceDir := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithShellPathResolver(func() string { return "/bin/zsh" }),
	)

	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath: %v", err)
	}
	if _, err := app.AddProjectFromPath(projectA); err != nil {
		t.Fatalf("AddProjectFromPath(A): %v", err)
	}
	if _, err := app.AddProjectFromPath(projectB); err != nil {
		t.Fatalf("AddProjectFromPath(B): %v", err)
	}

	app.ensureClaudeStatusHooksForActiveWorkspace()

	for _, dir := range []string{projectA, projectB} {
		state, err := app.ClaudeStatusHookState(dir)
		if err != nil {
			t.Fatalf("ClaudeStatusHookState(%s): %v", dir, err)
		}
		if !state.Installed {
			t.Fatalf("project %s hook not installed after workspace fill, got %#v", dir, state)
		}
	}
}

func TestApp_EnsureClaudeStatusHooksForActiveWorkspace_NoopWithoutWorkspace(t *testing.T) {
	setHookExe(t, filepath.Join(t.TempDir(), "todoai.exe"))
	projectDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithInitialWorkspaceClosed(),
	)

	app.ensureClaudeStatusHooksForActiveWorkspace()

	// No active workspace project → EnsureClaudeStatusHook never ran, so the
	// project still reports not-installed.
	state, err := app.ClaudeStatusHookState(projectDir)
	if err != nil {
		t.Fatalf("ClaudeStatusHookState: %v", err)
	}
	if state.Installed {
		t.Fatalf("expected hook not installed without workspace fill, got %#v", state)
	}
}

// TestApp_EnsureClaudeStatusHooksForActiveWorkspace_InstallsTodoWorktree is the
// regression test for the multi-claude todo-terminal gap: a TODO project runs in
// its own git worktree, whose .claude/settings.json is separate from the source
// project. The workspace fill must install the hook there too, or a claude
// started in a todo terminal never fires the hook.
func TestApp_EnsureClaudeStatusHooksForActiveWorkspace_InstallsTodoWorktree(t *testing.T) {
	setHookExe(t, filepath.Join(t.TempDir(), "todoai.exe"))
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	worktreeDir := t.TempDir()
	app := NewAppWithConfigAndShellStarter(
		filepath.Join(t.TempDir(), "projects.json"),
		newFakeShellStarter().Start,
		WithShellPathResolver(func() string { return "/bin/zsh" }),
	)
	if _, err := app.OpenWorkspaceFromPath(workspaceDir); err != nil {
		t.Fatalf("OpenWorkspaceFromPath: %v", err)
	}
	state, err := app.AddProjectFromPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectFromPath: %v", err)
	}
	_, todoProjectID := createTodoProjectForApp(t, app, "修复登录问题", state.Projects[0].ID)

	// Simulate a prepared (ready) worktree for the todo project.
	state, err = app.projects.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for i := range state.TodoProjects {
		if state.TodoProjects[i].ID == todoProjectID {
			state.TodoProjects[i].WorktreeStatus = WorktreeStatusReady
			state.TodoProjects[i].WorktreePath = worktreeDir
		}
	}
	if err := app.projects.saveLocked(state); err != nil {
		t.Fatalf("saveLocked: %v", err)
	}

	app.ensureClaudeStatusHooksForActiveWorkspace()

	hookState, err := app.ClaudeStatusHookState(worktreeDir)
	if err != nil {
		t.Fatalf("ClaudeStatusHookState(worktree): %v", err)
	}
	if !hookState.Installed {
		t.Fatalf("todo worktree hook not installed, got %#v", hookState)
	}
}
