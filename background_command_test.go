//go:build !windows

package main

import (
	"context"
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
