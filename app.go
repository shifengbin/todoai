package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	projects  *ProjectManager
	shells    *ShellSessionManager
	settings  *SettingsManager
	gitStatus func(path string) (GitStatus, error)
	gitInit   func(path string) error
}

func NewApp() *App {
	return NewAppWithConfig(defaultProjectConfigPath())
}

func NewAppWithConfig(configPath string) *App {
	return NewAppWithConfigAndShellStarter(configPath, NewPtyProcess)
}

func NewAppWithConfigAndShellStarter(configPath string, starter ShellStarter, shellOpts ...ShellSessionManagerOption) *App {
	app := &App{
		projects:  NewProjectManager(configPath),
		settings:  NewSettingsManager(defaultSettingsConfigPath(configPath)),
		gitStatus: queryGitStatus,
		gitInit:   initializeGitRepository,
	}
	shellOpts = append([]ShellSessionManagerOption{
		WithShellPathResolver(app.settings.ResolveShellPath),
	}, shellOpts...)
	app.shells = NewShellSessionManager(starter, ShellSessionCallbacks{
		OnOutput: app.emitTerminalOutput,
		OnStatus: app.emitShellStatus,
	}, shellOpts...)
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.shells.Shutdown()
}

func (a *App) ListProjects() (ProjectState, error) {
	state, err := a.projects.Load()
	if err != nil {
		return ProjectState{}, err
	}
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

func (a *App) SelectProject(projectID string) (ProjectState, error) {
	state, err := a.projects.SelectProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	project, ok := projectByID(state.Projects, projectID)
	if ok && project.Available {
		if _, err := a.shells.EnsureProjectTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
			return ProjectState{}, err
		}
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

func (a *App) DeleteProject(projectID string) (ProjectState, error) {
	state, err := a.projects.DeleteProject(projectID)
	if err != nil {
		return ProjectState{}, err
	}
	a.shells.DeleteProjectTerminals(projectID)
	return a.withShellState(state), nil
}

func (a *App) SelectTerminal(terminalID string) (ProjectState, error) {
	terminal, err := a.shells.SelectTerminal(terminalID)
	if err != nil {
		return ProjectState{}, err
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
	return filepath.Join(configDir, "tui-helper", "projects.json")
}

func defaultSettingsConfigPath(projectConfigPath string) string {
	return filepath.Join(filepath.Dir(projectConfigPath), "settings.json")
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
	state.ActiveTerminalID = a.shells.ActiveTerminalID(state.ActiveProjectID)
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
