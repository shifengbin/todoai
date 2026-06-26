package main

// ensureClaudeStatusHooksForActiveWorkspace installs (or refreshes) the Claude
// status hook for every project loaded in the currently active workspace. It
// runs after startup workspace restore so the hook is present before the user
// opens any terminal. Idempotent, so it is also safe to re-run after a workspace
// switch. Individual project failures are skipped rather than aborting the rest
// — the settings panel exposes ClaudeStatusHookState for manual diagnosis.
func (a *App) ensureClaudeStatusHooksForActiveWorkspace() {
	if !a.hasWorkspace() {
		return
	}
	state, err := a.projects.Load()
	if err != nil {
		return
	}
	for _, project := range state.Projects {
		if project.Path == "" {
			continue
		}
		_ = ensureClaudeStatusHook(project.Path)
	}
	// TODO project terminals run inside their prepared git worktrees, whose
	// .claude/settings.json is independent of the source project checkout (the
	// worktree is a separate working directory). Install the hook there too, or
	// a claude started in a todo terminal never fires the hook and its status is
	// silently lost.
	for _, todoProject := range state.TodoProjects {
		if todoProject.WorktreeStatus != WorktreeStatusReady || todoProject.WorktreePath == "" {
			continue
		}
		_ = ensureClaudeStatusHook(todoProject.WorktreePath)
	}
}

// EnsureClaudeStatusHook installs or refreshes the Claude status hook for a
// project's .claude/settings.json. Exposed to the frontend settings panel.
func (a *App) EnsureClaudeStatusHook(projectPath string) error {
	return ensureClaudeStatusHook(projectPath)
}

// RemoveClaudeStatusHook removes the todoai status-hook entries from a
// project's settings, leaving any user-installed hooks untouched.
func (a *App) RemoveClaudeStatusHook(projectPath string) error {
	return removeClaudeStatusHook(projectPath)
}

// ClaudeStatusHookState reports whether the status hook is installed for a
// project and whether the installed command matches the current executable
// (Stale signals a dev-mode path drift the UI should prompt to reinstall).
func (a *App) ClaudeStatusHookState(projectPath string) (ClaudeHookState, error) {
	return claudeStatusHookState(projectPath)
}
