package main

import (
	"context"
	"os"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	projects *ProjectManager
	shells   *ShellSessionManager
}

func NewApp() *App {
	return NewAppWithConfig(defaultProjectConfigPath())
}

func NewAppWithConfig(configPath string) *App {
	app := &App{
		projects: NewProjectManager(configPath),
	}
	app.shells = NewShellSessionManager(NewPtyProcess, ShellSessionCallbacks{
		OnOutput: app.emitTerminalOutput,
		OnStatus: app.emitShellStatus,
	})
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.shells.Shutdown()
}

func (a *App) ListProjects() (ProjectState, error) {
	return a.projects.Load()
}

func (a *App) CreateProjectFromDialog() (ProjectState, error) {
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return ProjectState{}, err
	}
	if path == "" {
		return a.projects.Load()
	}
	_, _, err = a.projects.AddProjectPath(path)
	if err != nil {
		return ProjectState{}, err
	}
	return a.projects.Load()
}

func (a *App) AddProjectFromPath(path string) (ProjectState, error) {
	_, _, err := a.projects.AddProjectPath(path)
	if err != nil {
		return ProjectState{}, err
	}
	return a.projects.Load()
}

func (a *App) SelectProject(projectID string) (ProjectState, error) {
	return a.projects.SelectProject(projectID)
}

func (a *App) StartShell(projectID string, cols int, rows int) (ShellStatus, error) {
	project, err := a.projects.GetProject(projectID)
	if err != nil {
		return ShellStatus{}, err
	}
	return a.shells.EnsureSession(project, normalizeTerminalSize(cols, rows))
}

func (a *App) SendTerminalInput(projectID string, data string) error {
	return a.shells.WriteInput(projectID, data)
}

func (a *App) ResizeTerminal(projectID string, cols int, rows int) error {
	return a.shells.Resize(projectID, normalizeTerminalSize(cols, rows))
}

func (a *App) GetShellStatus(projectID string) ShellStatus {
	return a.shells.Status(projectID)
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

func normalizeTerminalSize(cols int, rows int) TerminalSize {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return TerminalSize{Cols: cols, Rows: rows}
}
