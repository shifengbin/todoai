package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type BackgroundCommandRequest struct {
	Command    string
	WorkingDir string
	ShellPath  string
	ShellArgs  []string
	Env        []string
}

type BackgroundCommandRunner func(BackgroundCommandRequest) error

func runBackgroundCommand(request BackgroundCommandRequest) error {
	cmd := newBackgroundCommand(context.Background(), request.ShellPath, request.ShellArgs...)
	cmd.Dir = request.WorkingDir
	cmd.Env = request.Env
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

func BackgroundShellLaunch(shellPath string, command string, workingDir string, baseEnv []string) (BackgroundCommandRequest, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return BackgroundCommandRequest{}, errors.New("background command is required")
	}
	shellPath = strings.TrimSpace(shellPath)
	if shellPath == "" {
		return BackgroundCommandRequest{}, errors.New("background shell path is required")
	}

	args, err := backgroundShellArgs(shellNameFromPath(shellPath), command)
	if err != nil {
		return BackgroundCommandRequest{}, err
	}
	return BackgroundCommandRequest{
		Command:    command,
		WorkingDir: workingDir,
		ShellPath:  shellPath,
		ShellArgs:  args,
		Env:        EmbeddedTerminalEnv(baseEnv),
	}, nil
}

func backgroundShellArgs(shellName string, command string) ([]string, error) {
	switch strings.ToLower(shellName) {
	case "bash", "sh", "zsh", "fish":
		return []string{"-lc", command}, nil
	case "pwsh", "powershell":
		return []string{"-NoLogo", "-ExecutionPolicy", "Bypass", "-Command", command}, nil
	case "cmd":
		return []string{"/C", command}, nil
	default:
		return nil, fmt.Errorf("background command shell is unsupported: %s", shellName)
	}
}
