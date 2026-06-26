//go:build !windows

package main

import (
	"context"
	"reflect"
	"testing"
)

func TestNewBackgroundCommandKeepsDefaultProcessAttributesOnNonWindows(t *testing.T) {
	cmd := newBackgroundCommand(context.Background(), "git", "status")

	if len(cmd.Args) == 0 || cmd.Args[0] != "git" {
		t.Fatalf("Args = %#v, want first arg git", cmd.Args)
	}
	if cmd.SysProcAttr != nil {
		t.Fatalf("SysProcAttr = %#v, want nil on non-Windows", cmd.SysProcAttr)
	}
}

func TestBackgroundShellLaunchUsesOneShotShellArguments(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
		wantArgs  []string
	}{
		{name: "zsh", shellPath: "/bin/zsh", wantArgs: []string{"-lc", "npm run sync"}},
		{name: "bash", shellPath: "/usr/bin/bash", wantArgs: []string{"-lc", "npm run sync"}},
		{name: "sh", shellPath: "/bin/sh", wantArgs: []string{"-lc", "npm run sync"}},
		{name: "pwsh", shellPath: "C:\\Program Files\\PowerShell\\7\\pwsh.exe", wantArgs: []string{"-NoLogo", "-ExecutionPolicy", "Bypass", "-Command", "npm run sync"}},
		{name: "cmd", shellPath: "C:\\Windows\\System32\\cmd.exe", wantArgs: []string{"/C", "npm run sync"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := BackgroundShellLaunch(tt.shellPath, "npm run sync", "/work/project", []string{"PATH=/bin"})
			if err != nil {
				t.Fatalf("BackgroundShellLaunch() error = %v", err)
			}
			if request.ShellPath != tt.shellPath {
				t.Fatalf("ShellPath = %q, want %q", request.ShellPath, tt.shellPath)
			}
			if !reflect.DeepEqual(request.ShellArgs, tt.wantArgs) {
				t.Fatalf("ShellArgs = %#v, want %#v", request.ShellArgs, tt.wantArgs)
			}
			if request.WorkingDir != "/work/project" {
				t.Fatalf("WorkingDir = %q, want /work/project", request.WorkingDir)
			}
			if request.Command != "npm run sync" {
				t.Fatalf("Command = %q, want npm run sync", request.Command)
			}
		})
	}
}
