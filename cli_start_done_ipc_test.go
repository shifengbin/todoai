package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIStartDispatchesIPCCommandWithoutStartingGUI(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	result, calls := runCLILifecycleCommandForTest(t, fixture, []string{"start"}, nil)

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want %d; stderr=%q", result.exitCode, cliExitSuccess, result.stderr)
	}
	if !result.handled {
		t.Fatal("handled = false, want true")
	}
	if len(calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(calls))
	}
	if calls[0].command != "start" {
		t.Fatalf("command = %q, want start", calls[0].command)
	}
	if calls[0].appConfigDir != fixture.appConfig {
		t.Fatalf("appConfigDir = %q, want %q", calls[0].appConfigDir, fixture.appConfig)
	}
	if calls[0].workingDir != fixture.cwd {
		t.Fatalf("workingDir = %q, want %q", calls[0].workingDir, fixture.cwd)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
}

func TestCLIDoneDispatchesNormalizedWorkingDirectory(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	childDir := filepath.Join(fixture.cwd, "nested")
	mustMkdirAllForCLITest(t, childDir)
	fixture.cwd = filepath.Join(childDir, "..", "nested")
	wantWorkingDir, err := normalizeWorkspacePath(fixture.cwd)
	if err != nil {
		t.Fatalf("normalizeWorkspacePath() error = %v", err)
	}

	result, calls := runCLILifecycleCommandForTest(t, fixture, []string{"done"}, nil)

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want %d; stderr=%q", result.exitCode, cliExitSuccess, result.stderr)
	}
	if len(calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(calls))
	}
	if calls[0].command != "done" {
		t.Fatalf("command = %q, want done", calls[0].command)
	}
	if calls[0].workingDir != wantWorkingDir {
		t.Fatalf("workingDir = %q, want %q", calls[0].workingDir, wantWorkingDir)
	}
}

func TestCLILifecycleCommandReturnsIPCError(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	result, calls := runCLILifecycleCommandForTest(t, fixture, []string{"start"}, errors.New("todoai gui is not running or unreachable"))

	if result.exitCode != cliExitError {
		t.Fatalf("exitCode = %d, want %d", result.exitCode, cliExitError)
	}
	if len(calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(calls))
	}
	assertCLIOutputContains(t, result.stderr, "todoai gui is not running or unreachable")
}

type cliLifecycleCommandCall struct {
	appConfigDir string
	command      string
	workingDir   string
}

func runCLILifecycleCommandForTest(t *testing.T, fixture *cliListDoneFixture, args []string, senderErr error) (cliListDoneResult, []cliLifecycleCommandCall) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	calls := []cliLifecycleCommandCall{}
	handled, exitCode := runCLICommand(cliCommandOptions{
		args:         args,
		workingDir:   fixture.cwd,
		appConfigDir: fixture.appConfig,
		stdout:       &stdout,
		stderr:       &stderr,
		ipcCommandSender: func(ctx context.Context, appConfigDir string, command string, workingDir string) error {
			calls = append(calls, cliLifecycleCommandCall{
				appConfigDir: appConfigDir,
				command:      command,
				workingDir:   workingDir,
			})
			return senderErr
		},
	})
	return cliListDoneResult{
		handled:  handled,
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}, calls
}

func mustMkdirAllForCLITest(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}
