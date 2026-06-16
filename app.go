package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	applicationDisplayName = "TodoAI"
	applicationID          = "todoai"
	legacyApplicationID    = "tui-helper"
)

type App struct {
	ctx                        context.Context
	projects                   *ProjectManager
	shells                     *ShellSessionManager
	settings                   *SettingsManager
	history                    *TerminalHistoryStore
	gitStatus                  func(path string) (GitStatus, error)
	gitInit                    func(path string) error
	claudeStatusDir            string
	claudeStatusWatcher        *ClaudeStatusWatcher
	claudeStatusStop           chan struct{}
	claudeStatusStopOnce       sync.Once
	terminalAgentStatusEmitter func(TerminalAgentStatusEvent)
}

type AppOption func(*App)

func NewApp() *App {
	return NewAppWithConfig(defaultProjectConfigPath())
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
	historyStore := NewTerminalHistoryStore(configDir)
	app := &App{
		projects:        NewProjectManager(configPath),
		settings:        NewSettingsManager(defaultSettingsConfigPath(configPath)),
		history:         historyStore,
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
	shellOpts = append([]ShellSessionManagerOption{
		WithShellPathResolver(app.settings.ResolveShellPath),
		WithTerminalHistoryStore(historyStore),
	}, shellOpts...)
	app.shells = NewShellSessionManager(starter, ShellSessionCallbacks{
		OnOutput:       app.emitTerminalOutput,
		OnStatus:       app.emitShellStatus,
		OnCommandState: app.emitTerminalCommandState,
	}, shellOpts...)
	return app
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
	a.startClaudeStatusWatcher()
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

func (a *App) ListProjects() (ProjectState, error) {
	state, err := a.projects.Load()
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.RestoreTerminals(state)
	return a.withShellState(state), nil
}

func (a *App) CreateProjectFromDialog() (ProjectState, error) {
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
	_, _, err := a.projects.AddProjectPath(path)
	if err != nil {
		return ProjectState{}, err
	}
	return a.ListProjects()
}

func (a *App) ImportProjectsFromParentDirectory(parentPath string) (ProjectState, error) {
	state, err := a.projects.ImportProjectsFromParentDirectory(parentPath)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) ImportProjectsFromParentDirectoryDialog() (ProjectState, error) {
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
	state, err := a.projects.SelectProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CreateTodo(request CreateTodoRequest) (ProjectState, error) {
	state, err := a.projects.CreateTodo(request)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) UpdateTodo(request UpdateTodoRequest) (ProjectState, error) {
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
	state, err := a.projects.AssociateProjectWithTodo(todoID, projectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) AddProjectsToTodo(todoID string, projectIDs []string) (ProjectState, error) {
	state, err := a.projects.AssociateProjectsWithTodo(todoID, projectIDs)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) RemoveTodoProject(todoProjectID string) (ProjectState, error) {
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
	state, _, _, err := a.projects.SelectTodoProject(todoProjectID)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) CreateTerminal(projectID string, cols int, rows int) (ProjectState, error) {
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
	state, err := a.projects.DeleteProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteProjectTerminals(projectID)
	return a.withShellState(state), nil
}

func (a *App) DeleteProjects(projectIDs []string) (ProjectState, error) {
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
	state, err := a.projects.CompleteTodo(todoID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteTodoTerminals(todoID)
	return a.withShellState(state), nil
}

func (a *App) DeleteTodo(todoID string) (ProjectState, error) {
	state, err := a.projects.DeleteTodo(todoID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteTodoTerminals(todoID)
	return a.withShellState(state), nil
}

func (a *App) DeleteCompletedTodos(todoIDs []string) (ProjectState, error) {
	state, err := a.projects.DeleteCompletedTodos(todoIDs)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) ChangeTodoStatus(todoID string, status string) (ProjectState, error) {
	state, err := a.projects.ChangeTodoStatus(todoID, status)
	if err != nil {
		return ProjectState{}, err
	}
	return a.withShellState(state), nil
}

func (a *App) StartShell(terminalID string, cols int, rows int) (ShellStatus, error) {
	return a.shells.StartTerminal(terminalID, normalizeTerminalSize(cols, rows))
}

func (a *App) SendTerminalInput(terminalID string, data string) error {
	return a.shells.WriteInput(terminalID, data)
}

func (a *App) ResizeTerminal(terminalID string, cols int, rows int) error {
	return a.shells.Resize(terminalID, normalizeTerminalSize(cols, rows))
}

func (a *App) GetShellStatus(terminalID string) ShellStatus {
	return a.shells.Status(terminalID)
}

func (a *App) GetProjectGitStatus(projectID string) (GitStatus, error) {
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
	state.Terminals = a.shells.Terminals()
	activeContextID := state.ActiveTodoProjectID
	if activeContextID == "" {
		activeContextID = state.ActiveProjectID
	}
	state.ActiveTerminalID = a.shells.ActiveTerminalID(activeContextID)
	return state
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
