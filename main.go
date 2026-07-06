package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// `todoai claude-hook` is invoked by Claude Code on every lifecycle event
	// (configured in .claude/settings.json). It must run without starting the
	// GUI, so branch off before NewApp()/wails.Run.
	if len(os.Args) > 1 && os.Args[1] == "claude-hook" {
		os.Exit(runClaudeHookCommand())
	}
	if handled, exitCode := runCLICommand(cliCommandOptions{args: os.Args[1:]}); handled {
		os.Exit(exitCode)
	}

	ensureProcessUTF8Locale()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  applicationDisplayName,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Menu:             app.applicationMenu(),
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
