package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/menu"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	applicationDisplayName = "TodoAI"
	applicationID          = "todoai"
	legacyApplicationID    = "tui-helper"
	workspaceStateEvent    = "workspace-state"
	workspaceRecentEvent   = "workspace-recent"
)

var ErrWorkspaceRequired = errors.New("open a project first")

type App struct {
	ctx                           context.Context
	workspace                     *WorkspaceManager
	projects                      *ProjectManager
	shells                        *ShellSessionManager
	settings                      *SettingsManager
	history                       *TerminalHistoryStore
	starter                       ShellStarter
	shellOpts                     []ShellSessionManagerOption
	gitStatus                     func(path string) (GitStatus, error)
	gitInit                       func(path string) error
	claudeStatusDir               string
	claudeStatusWatcher           *ClaudeStatusWatcher
	claudeStatusStop              chan struct{}
	claudeStatusStopOnce          sync.Once
	terminalAgentStatusEmitter    func(TerminalAgentStatusEvent)
	initialWorkspaceClosed        bool
	restoreLastWorkspaceOnStartup bool
}

type AppOption func(*App)

func NewApp() *App {
	return NewAppWithConfig(defaultProjectConfigPath(), WithInitialWorkspaceClosed(), WithRestoreLastWorkspaceOnStartup())
}

func NewAppWithConfig(configPath string, opts ...AppOption) *App {
	typedOpts := make([]any, 0, len(opts))
	for _, opt := range opts {
		typedOpts = append(typedOpts, opt)
	}
	return NewAppWithConfigAndShellStarter(configPath, NewPtyProcess, typedOpts...)
}

func NewAppWithConfigAndShellStarter(configPath string, starter ShellStarter, opts ...any) *App {
	configDir := filepath.Dir(configPath)
	workspaceManager := NewWorkspaceManager(configDir)
	historyStore := NewTerminalHistoryStore(configDir)
	app := &App{
		workspace:       workspaceManager,
		projects:        NewProjectManager(configPath),
		settings:        NewSettingsManager(defaultSettingsConfigPath(configPath)),
		history:         historyStore,
		starter:         starter,
		gitStatus:       queryGitStatus,
		gitInit:         initializeGitRepository,
		claudeStatusDir: defaultClaudeStatusDir,
	}
	var shellOpts []ShellSessionManagerOption
	for _, opt := range opts {
		switch typed := opt.(type) {
		case ShellSessionManagerOption:
			shellOpts = append(shellOpts, typed)
		case AppOption:
			typed(app)
		}
	}
	app.shellOpts = append([]ShellSessionManagerOption{}, shellOpts...)
	if !app.initialWorkspaceClosed {
		workspace := Workspace{
			Name:      filepath.Base(configDir),
			Path:      configDir,
			DataPath:  configDir,
			Available: directoryAvailable(configDir),
		}
		app.workspace.current = &workspace
	}
	app.rebuildShellSessionManager()
	return app
}

func WithInitialWorkspaceClosed() AppOption {
	return func(app *App) {
		app.initialWorkspaceClosed = true
	}
}

func WithRestoreLastWorkspaceOnStartup() AppOption {
	return func(app *App) {
		app.restoreLastWorkspaceOnStartup = true
	}
}

func (a *App) rebuildShellSessionManager() {
	shellOpts := append([]ShellSessionManagerOption{
		WithShellPathResolver(a.settings.ResolveShellPath),
		WithTerminalHistoryStore(a.history),
	}, a.shellOpts...)
	a.shells = NewShellSessionManager(a.starter, ShellSessionCallbacks{
		OnOutput:       a.emitTerminalOutput,
		OnStatus:       a.emitShellStatus,
		OnCommandState: a.emitTerminalCommandState,
	}, shellOpts...)
}

func WithClaudeStatusDir(dir string) AppOption {
	return func(app *App) {
		app.claudeStatusDir = dir
	}
}

func WithTerminalAgentStatusEmitter(emit func(TerminalAgentStatusEvent)) AppOption {
	return func(app *App) {
		app.terminalAgentStatusEmitter = emit
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	state, _ := a.workspace.MigrateLegacyGlobalData()
	if a.restoreLastWorkspaceOnStartup {
		a.restoreLastWorkspace(state)
	}
	a.startClaudeStatusWatcher()
}

func (a *App) restoreLastWorkspace(state WorkspaceState) {
	if a.workspace.CurrentWorkspace() != nil {
		return
	}
	if len(state.RecentWorkspaces) == 0 {
		return
	}
	workspaceState, err := a.workspace.OpenWorkspace(state.RecentWorkspaces[0].Path)
	if err != nil || workspaceState.CurrentWorkspace == nil {
		a.bindNoWorkspace()
		return
	}
	a.bindWorkspace(*workspaceState.CurrentWorkspace)
}

func (a *App) shutdown(ctx context.Context) {
	a.stopClaudeStatusWatcher()
	a.shells.Shutdown()
}

func (a *App) startClaudeStatusWatcher() {
	if a.claudeStatusDir == "" {
		return
	}
	a.claudeStatusStop = make(chan struct{})
	a.claudeStatusStopOnce = sync.Once{}
	a.claudeStatusWatcher = NewClaudeStatusWatcher(a.claudeStatusDir, a.shells.Terminals, a.emitTerminalAgentStatus)
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.pollClaudeStatus()
			case <-a.claudeStatusStop:
				return
			}
		}
	}()
}

func (a *App) stopClaudeStatusWatcher() {
	if a.claudeStatusStop == nil {
		return
	}
	a.claudeStatusStopOnce.Do(func() {
		close(a.claudeStatusStop)
	})
}

func (a *App) pollClaudeStatus() {
	if a.claudeStatusWatcher != nil {
		a.claudeStatusWatcher.Poll()
	}
}

func (a *App) applicationMenu() *menu.Menu {
	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("文件")
	fileMenu.AddText("打开项目", nil, func(_ *menu.CallbackData) {
		_, _ = a.OpenWorkspaceFromDialog()
	})
	fileMenu.AddText("最近打开", nil, func(_ *menu.CallbackData) {
		a.emitWorkspaceRecent()
	})
	fileMenu.AddText("清理最近打开", nil, func(_ *menu.CallbackData) {
		_, _ = a.ClearRecentWorkspaces()
	})
	fileMenu.AddText("关闭", nil, func(_ *menu.CallbackData) {
		_, _ = a.CloseWorkspace()
	})
	return appMenu
}

func (a *App) WorkspaceState() (WorkspaceState, error) {
	return a.workspace.LoadState()
}

func (a *App) OpenWorkspaceFromDialog() (ProjectState, error) {
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "打开项目",
	})
	if err != nil {
		return ProjectState{}, err
	}
	if path == "" {
		return a.currentProjectState()
	}
	return a.OpenWorkspaceFromPath(path)
}

func (a *App) OpenWorkspaceFromPath(path string) (ProjectState, error) {
	previousWorkspace := a.workspace.CurrentWorkspace()
	workspaceState, err := a.workspace.OpenWorkspace(path)
	if err != nil {
		state, _ := a.currentProjectState()
		return state, err
	}
	nextWorkspace := workspaceState.CurrentWorkspace
	if nextWorkspace == nil {
		return a.emptyProjectState(workspaceState), nil
	}
	if previousWorkspace == nil || previousWorkspace.Path != nextWorkspace.Path {
		a.resetRuntimeForWorkspaceChange()
	}
	a.bindWorkspace(*nextWorkspace)
	state, err := a.ListProjects()
	if err == nil {
		a.emitWorkspaceState(state)
	}
	return state, err
}

func (a *App) OpenRecentWorkspace(path string) (ProjectState, error) {
	return a.OpenWorkspaceFromPath(path)
}

func (a *App) ClearRecentWorkspaces() (WorkspaceState, error) {
	state, err := a.workspace.ClearRecentWorkspaces()
	if err == nil {
		a.emitWorkspaceRecent()
	}
	return state, err
}

func (a *App) CloseWorkspace() (ProjectState, error) {
	a.resetRuntimeForWorkspaceChange()
	workspaceState, err := a.workspace.CloseWorkspace()
	if err != nil {
		return ProjectState{}, err
	}
	a.bindNoWorkspace()
	state := a.emptyProjectState(workspaceState)
	a.emitWorkspaceState(state)
	return state, nil
}

func (a *App) ListProjects() (ProjectState, error) {
	if !a.hasWorkspace() {
		workspaceState, err := a.workspace.LoadState()
		if err != nil {
			return ProjectState{}, err
		}
		return a.emptyProjectState(workspaceState), nil
	}
	state, err := a.projects.Load()
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.RestoreTerminals(state)
	return a.withShellState(state), nil
}

func (a *App) CreateProjectFromDialog() (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return ProjectState{}, err
	}
	if path == "" {
		return a.ListProjects()
	}
	_, _, err = a.projects.AddProjectPath(path)
	if err != nil {
		return ProjectState{}, err
	}
	return a.ListProjects()
}

func (a *App) AddProjectFromPath(path string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	_, _, err := a.projects.AddProjectPath(path)
	if err != nil {
		return ProjectState{}, err
	}
	return a.ListProjects()
}

func (a *App) ImportProjectsFromParentDirectory(parentPath string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.ImportProjectsFromParentDirectory(parentPath)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) ImportProjectsFromParentDirectoryDialog() (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Parent Directory",
	})
	if err != nil {
		return ProjectState{}, err
	}
	if path == "" {
		return a.ListProjects()
	}
	return a.ImportProjectsFromParentDirectory(path)
}

func (a *App) SelectProject(projectID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.SelectProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CreateTodo(request CreateTodoRequest) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.CreateTodo(request)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) UpdateTodo(request UpdateTodoRequest) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, removedTodoProjectIDs, err := a.projects.UpdateTodo(request)
	if err != nil {
		return ProjectState{}, err
	}
	for _, todoProjectID := range removedTodoProjectIDs {
		a.shells.DeleteTodoProjectTerminals(todoProjectID)
	}
	return a.withShellState(state), nil
}

func (a *App) AddProjectToTodo(todoID string, projectID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.AssociateProjectWithTodo(todoID, projectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) AddProjectsToTodo(todoID string, projectIDs []string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.AssociateProjectsWithTodo(todoID, projectIDs)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) RemoveTodoProject(todoProjectID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, removedTodoProjectIDs, err := a.projects.RemoveTodoProject(todoProjectID)
	if err != nil {
		return ProjectState{}, err
	}
	for _, removedTodoProjectID := range removedTodoProjectIDs {
		a.shells.DeleteTodoProjectTerminals(removedTodoProjectID)
	}
	return a.withShellState(state), nil
}

func (a *App) SelectTodoProject(todoProjectID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, _, _, err := a.projects.SelectTodoProject(todoProjectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CreateTerminal(projectID string, cols int, rows int) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.SelectProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	project, ok := projectByID(state.Projects, projectID)
	if !ok {
		return ProjectState{}, os.ErrNotExist
	}
	if _, err := a.shells.CreateTerminal(project, normalizeTerminalSize(cols, rows)); err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CreateTodoTerminal(todoProjectID string, cols int, rows int) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	currentState, err := a.projects.Load()
	if err != nil {
		return ProjectState{}, err
	}
	terminalTodoProject, ok := todoProjectByID(currentState.TodoProjects, todoProjectID)
	if !ok {
		return ProjectState{}, os.ErrNotExist
	}
	todo, ok := openTodoByID(currentState.Todos, terminalTodoProject.TodoID)
	if !ok {
		return ProjectState{}, os.ErrNotExist
	}
	if todo.Status != TodoStatusInProgress {
		return ProjectState{}, fmt.Errorf("todo is not in progress")
	}

	state, todoProject, project, err := a.projects.SelectTodoProject(todoProjectID)
	if err != nil {
		return ProjectState{}, err
	}
	if _, err := a.shells.CreateTodoProjectTerminal(todoProject, project, normalizeTerminalSize(cols, rows)); err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) DeleteProject(projectID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.DeleteProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteProjectTerminals(projectID)
	return a.withShellState(state), nil
}

func (a *App) DeleteProjects(projectIDs []string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.DeleteProjects(projectIDs)
	if err != nil {
		return ProjectState{}, err
	}
	for _, projectID := range normalizeProjectIDs(projectIDs) {
		a.shells.DeleteProjectTerminals(projectID)
	}
	return a.withShellState(state), nil
}

func (a *App) SelectTerminal(terminalID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	terminal, err := a.shells.SelectTerminal(terminalID)
	if err != nil {
		return ProjectState{}, err
	}
	if terminal.TodoProjectID != "" {
		state, _, _, err := a.projects.SelectTodoProject(terminal.TodoProjectID)
		if err != nil {
			return ProjectState{}, err
		}
		return a.withShellState(state), nil
	}
	state, err := a.projects.SelectProject(terminal.ProjectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) DeleteTerminal(terminalID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	if err := a.shells.DeleteTerminal(terminalID); err != nil {
		return ProjectState{}, err
	}
	state, err := a.projects.Load()
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CompleteTodo(todoID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.CompleteTodo(todoID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteTodoTerminals(todoID)
	return a.withShellState(state), nil
}

func (a *App) DeleteTodo(todoID string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.DeleteTodo(todoID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteTodoTerminals(todoID)
	return a.withShellState(state), nil
}

func (a *App) DeleteCompletedTodos(todoIDs []string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.DeleteCompletedTodos(todoIDs)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) ChangeTodoStatus(todoID string, status string) (ProjectState, error) {
	if !a.hasWorkspace() {
		return a.currentProjectStateWithError(ErrWorkspaceRequired)
	}
	state, err := a.projects.ChangeTodoStatus(todoID, status)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) StartShell(terminalID string, cols int, rows int) (ShellStatus, error) {
	if !a.hasWorkspace() {
		return ShellStatus{}, ErrWorkspaceRequired
	}
	return a.shells.StartTerminal(terminalID, normalizeTerminalSize(cols, rows))
}

func (a *App) SendTerminalInput(terminalID string, data string) error {
	if !a.hasWorkspace() {
		return ErrWorkspaceRequired
	}
	return a.shells.WriteInput(terminalID, data)
}

func (a *App) ResizeTerminal(terminalID string, cols int, rows int) error {
	if !a.hasWorkspace() {
		return ErrWorkspaceRequired
	}
	return a.shells.Resize(terminalID, normalizeTerminalSize(cols, rows))
}

func (a *App) GetShellStatus(terminalID string) ShellStatus {
	return a.shells.Status(terminalID)
}

func (a *App) GetProjectGitStatus(projectID string) (GitStatus, error) {
	if !a.hasWorkspace() {
		return GitStatus{}, ErrWorkspaceRequired
	}
	project, err := a.projects.GetProject(projectID)
	if err != nil {
		return GitStatus{}, err
	}
	status := GitStatus{ProjectID: projectID}
	if !project.Available {
		status.PathUnavailable = true
		return status, nil
	}

	status, err = a.gitStatus(project.Path)
	if err != nil {
		return GitStatus{}, err
	}
	status.ProjectID = projectID
	return status, nil
}

func (a *App) InitializeProjectGitRepository(projectID string) error {
	if !a.hasWorkspace() {
		return ErrWorkspaceRequired
	}
	project, err := a.projects.GetProject(projectID)
	if err != nil {
		return err
	}
	if !project.Available {
		return fmt.Errorf("project path unavailable")
	}
	return a.gitInit(project.Path)
}

func (a *App) LoadTerminalSettings() (TerminalSettingsState, error) {
	return a.settings.Load()
}

func (a *App) SaveTerminalShell(path string, source string) (TerminalSettingsState, error) {
	return a.settings.SaveShellPath(path, source)
}

func (a *App) SaveTerminalLaunchProfiles(profiles []TerminalLaunchProfileSetting) (TerminalSettingsState, error) {
	return a.settings.SaveLaunchProfiles(profiles)
}

func (a *App) SaveTerminalTheme(theme string) (TerminalSettingsState, error) {
	return a.settings.SaveTheme(theme)
}

func (a *App) DetectTerminalShell() (TerminalShellSetting, error) {
	return a.settings.DetectShell()
}

func (a *App) emitTerminalOutput(event TerminalOutputEvent) {
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "terminal-output", event)
	}
}

func (a *App) emitWorkspaceState(state ProjectState) {
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, workspaceStateEvent, state)
	}
}

func (a *App) emitWorkspaceRecent() {
	if a.ctx == nil {
		return
	}
	state, err := a.workspace.LoadState()
	if err != nil {
		wailsruntime.EventsEmit(a.ctx, workspaceRecentEvent, WorkspaceState{Version: workspaceStateFileVersion})
		return
	}
	wailsruntime.EventsEmit(a.ctx, workspaceRecentEvent, state)
}

func (a *App) emitTerminalCommandState(event TerminalCommandStateEvent) {
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "terminal-command-state", event)
	}
}

func (a *App) emitTerminalAgentStatus(event TerminalAgentStatusEvent) {
	if a.terminalAgentStatusEmitter != nil {
		a.terminalAgentStatusEmitter(event)
		return
	}
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "terminal-agent-status", event)
	}
}

func (a *App) emitShellStatus(status ShellStatus) {
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "terminal-status", status)
	}
}

func defaultProjectConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	appConfigDir := resolveAppConfigDir(filepath.Join(configDir, legacyApplicationID), filepath.Join(configDir, applicationID), copyDir)
	return filepath.Join(appConfigDir, "projects.json")
}

func defaultSettingsConfigPath(projectConfigPath string) string {
	return filepath.Join(filepath.Dir(projectConfigPath), "settings.json")
}

func resolveAppConfigDir(legacyDir string, appConfigDir string, migrate func(string, string) error) string {
	if legacyDir == appConfigDir {
		return appConfigDir
	}
	if _, err := os.Stat(appConfigDir); err == nil {
		return appConfigDir
	}
	if _, err := os.Stat(legacyDir); err != nil {
		return appConfigDir
	}
	if err := migrate(legacyDir, appConfigDir); err != nil {
		_ = os.RemoveAll(appConfigDir)
		return legacyDir
	}
	return appConfigDir
}

func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return nil
	}
	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode().Perm())
}

func normalizeTerminalSize(cols int, rows int) TerminalSize {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return TerminalSize{Cols: cols, Rows: rows}
}

func (a *App) withShellState(state ProjectState) ProjectState {
	workspaceState, err := a.workspace.LoadState()
	if err == nil {
		state.CurrentWorkspace = workspaceState.CurrentWorkspace
		state.RecentWorkspaces = workspaceState.RecentWorkspaces
	}
	state.Terminals = a.shells.Terminals()
	activeContextID := state.ActiveTodoProjectID
	if activeContextID == "" {
		activeContextID = state.ActiveProjectID
	}
	state.ActiveTerminalID = a.shells.ActiveTerminalID(activeContextID)
	return state
}

func (a *App) hasWorkspace() bool {
	return a.workspace != nil && a.workspace.CurrentWorkspace() != nil
}

func (a *App) bindWorkspace(workspace Workspace) {
	a.projects = NewProjectManager(filepath.Join(workspace.DataPath, "projects.json"))
	a.history = NewTerminalHistoryStore(workspace.DataPath)
	a.rebuildShellSessionManager()
}

func (a *App) bindNoWorkspace() {
	a.projects = NewProjectManager(filepath.Join(os.TempDir(), applicationID, "closed-workspace", "projects.json"))
	a.history = NewTerminalHistoryStore(filepath.Join(os.TempDir(), applicationID, "closed-workspace"))
	a.rebuildShellSessionManager()
}

func (a *App) resetRuntimeForWorkspaceChange() {
	if a.shells != nil {
		a.shells.Reset()
	}
}

func (a *App) currentProjectState() (ProjectState, error) {
	if !a.hasWorkspace() {
		workspaceState, err := a.workspace.LoadState()
		if err != nil {
			return ProjectState{}, err
		}
		return a.emptyProjectState(workspaceState), nil
	}
	return a.ListProjects()
}

func (a *App) currentProjectStateWithError(err error) (ProjectState, error) {
	state, stateErr := a.currentProjectState()
	if stateErr != nil {
		return ProjectState{}, stateErr
	}
	return state, err
}

func (a *App) emptyProjectState(workspaceState WorkspaceState) ProjectState {
	return ProjectState{
		Version:          projectConfigVersion,
		CurrentWorkspace: workspaceState.CurrentWorkspace,
		RecentWorkspaces: workspaceState.RecentWorkspaces,
		Projects:         []Project{},
		Todos:            []Todo{},
		TodoProjects:     []TodoProject{},
		Terminals:        []ProjectTerminal{},
	}
}

func projectByID(projects []Project, projectID string) (Project, bool) {
	for _, project := range projects {
		if project.ID == projectID {
			return project, true
		}
	}
	return Project{}, false
}

func todoProjectByID(todoProjects []TodoProject, todoProjectID string) (TodoProject, bool) {
	for _, todoProject := range todoProjects {
		if todoProject.ID == todoProjectID {
			return todoProject, true
		}
	}
	return TodoProject{}, false
}
