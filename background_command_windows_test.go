//go:build windows

package main

import (
	"context"
	"testing"
)

func TestNewBackgroundCommandHidesConsoleWindowOnWindows(t *testing.T) {
	cmd := newBackgroundCommand(context.Background(), "git", "status")

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr = nil, want Windows attributes")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow = false, want true")
	}
}
