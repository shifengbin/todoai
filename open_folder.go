package main

import (
	"errors"
	"os/exec"
	"runtime"
)

// openFolderInFileManager reveals the directory at path in the host operating
// system's file manager (Finder/Explorer/xdg-open). The directory must already
// exist on disk; callers validate that before invoking.
func openFolderInFileManager(path string) error {
	if !directoryAvailable(path) {
		return errors.New("directory does not exist")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
