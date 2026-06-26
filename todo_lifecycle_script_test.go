package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewTodoLifecycleScriptCommandUsesUnixShellAndWorkingDir(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
		Script:     "echo ok",
		WorkingDir: "/work/tasks/task-a",
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{"/bin/sh", "-c", "echo ok"})
	if cmd.Dir != "/work/tasks/task-a" {
		t.Fatalf("Dir = %q, want todo workspace", cmd.Dir)
	}
}

func TestNewTodoLifecycleScriptCommandUsesInteractiveUserShellForZsh(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  "/bin/zsh",
		GOOS:       "linux",
		Script:     "openspec status\ncodegraph --help",
		WorkingDir: "/work/tasks/task-a",
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{"/bin/zsh", "-i", "-c", "openspec status\ncodegraph --help"})
}

func TestNewTodoLifecycleScriptCommandUsesWindowsPowerShellAndWorkingDir(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		GOOS:       "windows",
		Script:     "Write-Output ok",
		WorkingDir: `C:\work\tasks\task-a`,
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		"-NoLogo",
		"-NoProfile",
		"-ExecutionPolicy",
		"Bypass",
		"-Command",
		"Write-Output ok",
	})
	if cmd.Dir != `C:\work\tasks\task-a` {
		t.Fatalf("Dir = %q, want todo workspace", cmd.Dir)
	}
}

func TestTodoLifecycleScriptExecutorRecordsFailedOutputTail(t *testing.T) {
	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	output := strings.Repeat("x", todoLifecycleScriptOutputTailLimit+20) + "tail"
	executor := NewTodoLifecycleScriptExecutor(
		func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			return TodoLifecycleScriptRunResult{Output: output, ExitCode: 7, Err: errors.New("exit status 7")}
		},
		WithTodoLifecycleScriptClock(func() time.Time { return now }),
	)

	_, started, err := executor.Start(context.Background(), TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseInit,
		ScriptName: "Node setup",
		Script:     "npm install",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !started {
		t.Fatal("Start() started = false, want true")
	}

	status := waitForLifecycleScriptStatus(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, TodoLifecycleScriptStatusFailed)
	if status.ScriptName != "Node setup" {
		t.Fatalf("ScriptName = %q, want Node setup", status.ScriptName)
	}
	if status.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", status.ExitCode)
	}
	if len(status.OutputTail) > todoLifecycleScriptOutputTailLimit {
		t.Fatalf("len(OutputTail) = %d, want <= %d", len(status.OutputTail), todoLifecycleScriptOutputTailLimit)
	}
	if !strings.HasSuffix(status.OutputTail, "tail") {
		t.Fatalf("OutputTail suffix = %q, want tail", status.OutputTail)
	}
}

func TestTodoLifecycleScriptExecutorPreventsDuplicateRuns(t *testing.T) {
	release := make(chan struct{})
	startedRun := make(chan struct{})
	executor := NewTodoLifecycleScriptExecutor(func(ctx context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
		close(startedRun)
		select {
		case <-release:
			return TodoLifecycleScriptRunResult{}
		case <-ctx.Done():
			return TodoLifecycleScriptRunResult{Err: ctx.Err(), ExitCode: -1}
		}
	})
	defer close(release)

	_, started, err := executor.Start(context.Background(), TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseInit,
		ScriptName: "Node setup",
		Script:     "npm install",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	})
	if err != nil {
		t.Fatalf("Start(first) error = %v", err)
	}
	if !started {
		t.Fatal("Start(first) started = false, want true")
	}
	<-startedRun

	status, started, err := executor.Start(context.Background(), TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseInit,
		ScriptName: "Node setup",
		Script:     "npm install",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	})
	if err != nil {
		t.Fatalf("Start(second) error = %v", err)
	}
	if started {
		t.Fatal("Start(second) started = true, want duplicate prevention")
	}
	if status.Status != TodoLifecycleScriptStatusRunning {
		t.Fatalf("duplicate status = %#v, want running", status)
	}
}

func TestTodoLifecycleScriptExecutorEmitsQueuedRunningFailedAndClearedStatuses(t *testing.T) {
	statuses := make(chan TodoLifecycleScriptStatus, 8)
	attempts := 0
	executor := NewTodoLifecycleScriptExecutor(
		func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			attempts++
			if attempts == 1 {
				return TodoLifecycleScriptRunResult{Output: "setup failed", ExitCode: 1, Err: errors.New("exit status 1")}
			}
			return TodoLifecycleScriptRunResult{}
		},
		WithTodoLifecycleScriptStatusHandler(func(status TodoLifecycleScriptStatus) {
			statuses <- status
		}),
	)

	request := TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseInit,
		ScriptName: "Node setup",
		Script:     "npm install",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	}
	if _, started, err := executor.Start(context.Background(), request); err != nil || !started {
		t.Fatalf("Start(first) status = started %v error %v, want started", started, err)
	}
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusQueued)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusRunning)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusFailed)

	if _, started, err := executor.Start(context.Background(), request); err != nil || !started {
		t.Fatalf("Start(second) status = started %v error %v, want started", started, err)
	}
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusQueued)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusRunning)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), "")
}

func waitForLifecycleScriptStatus(t *testing.T, executor *TodoLifecycleScriptExecutor, todoID string, phase string, want string) TodoLifecycleScriptStatus {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status, ok := executor.Status(todoID, phase)
		if ok && status.Status == want {
			return status
		}
		time.Sleep(time.Millisecond)
	}
	status, _ := executor.Status(todoID, phase)
	t.Fatalf("status = %#v, want %s", status, want)
	return TodoLifecycleScriptStatus{}
}

func receiveLifecycleExecutorStatus(t *testing.T, statuses <-chan TodoLifecycleScriptStatus) TodoLifecycleScriptStatus {
	t.Helper()
	select {
	case status := <-statuses:
		return status
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for lifecycle script status event")
		return TodoLifecycleScriptStatus{}
	}
}

func assertLifecycleScriptEventStatus(t *testing.T, status TodoLifecycleScriptStatus, want string) {
	t.Helper()
	if status.Status != want {
		t.Fatalf("status event = %#v, want %q", status, want)
	}
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(slice) = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("slice[%d] = %q, want %q; full=%#v", index, got[index], want[index], got)
		}
	}
}
