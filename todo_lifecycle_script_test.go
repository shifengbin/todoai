package main

import (
	"context"
	"encoding/json"
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

func TestNewTodoLifecycleScriptCommandRendersParametersForUnixShell(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
		Script:     "echo {{message}} {{unknown_value}}",
		WorkingDir: "/work/tasks/task-a",
		Parameters: []TodoLifecycleScriptParameter{
			{Name: "message"},
		},
		ParameterValues: map[string]string{"message": "say 'hi'; rm -rf /"},
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{"/bin/sh", "-c", "echo 'say '\\''hi'\\''; rm -rf /' {{unknown_value}}"})
}

func TestNewTodoLifecycleScriptCommandRendersEmptyParameterValue(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
		Script:     "echo {{optional_value}}",
		WorkingDir: "/work/tasks/task-a",
		Parameters: []TodoLifecycleScriptParameter{
			{Name: "optional_value"},
		},
		ParameterValues: map[string]string{"optional_value": ""},
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{"/bin/sh", "-c", "echo ''"})
}

func TestNewTodoLifecycleScriptCommandRendersParametersForPowerShell(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  `C:\Program Files\PowerShell\7\pwsh.exe`,
		GOOS:       "windows",
		Script:     "Write-Output {{message}}",
		WorkingDir: `C:\work\tasks\task-a`,
		Parameters: []TodoLifecycleScriptParameter{
			{Name: "message"},
		},
		ParameterValues: map[string]string{"message": "say 'hi'; Remove-Item"},
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{
		`C:\Program Files\PowerShell\7\pwsh.exe`,
		"-NoLogo",
		"-NoProfile",
		"-ExecutionPolicy",
		"Bypass",
		"-Command",
		"Write-Output 'say ''hi''; Remove-Item'",
	})
}

func TestNewTodoLifecycleScriptCommandUsesPowerShellCoreOnUnix(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  "/usr/local/bin/pwsh",
		GOOS:       "linux",
		Script:     "Write-Output {{message}}",
		WorkingDir: "/work/tasks/task-a",
		Parameters: []TodoLifecycleScriptParameter{
			{Name: "message"},
		},
		ParameterValues: map[string]string{"message": "don't split"},
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{
		"/usr/local/bin/pwsh",
		"-NoLogo",
		"-NoProfile",
		"-ExecutionPolicy",
		"Bypass",
		"-Command",
		"Write-Output 'don''t split'",
	})
}

func TestNewTodoLifecycleScriptCommandRendersParametersForCmd(t *testing.T) {
	cmd, err := newTodoLifecycleScriptCommand(context.Background(), TodoLifecycleScriptRunRequest{
		ShellPath:  `C:\Windows\System32\cmd.exe`,
		GOOS:       "windows",
		Script:     "echo {{message}}",
		WorkingDir: `C:\work\tasks\task-a`,
		Parameters: []TodoLifecycleScriptParameter{
			{Name: "message"},
		},
		ParameterValues: map[string]string{"message": `hello "quoted" %USERPROFILE% & dir`},
	})
	if err != nil {
		t.Fatalf("newTodoLifecycleScriptCommand() error = %v", err)
	}

	assertStringSlice(t, cmd.Args, []string{
		`C:\Windows\System32\cmd.exe`,
		"/d",
		"/s",
		"/c",
		`echo "hello ^"quoted^" ^%USERPROFILE^% ^& dir"`,
	})
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
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, output)
}

func TestTodoLifecycleScriptExecutorRetryInvalidatesAndReplacesFailedOutput(t *testing.T) {
	retryStarted := make(chan struct{})
	releaseRetry := make(chan struct{}, 1)
	defer close(releaseRetry)
	attempts := 0
	executor := NewTodoLifecycleScriptExecutor(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
		attempts++
		if attempts == 1 {
			return TodoLifecycleScriptRunResult{Output: "first failure output", ExitCode: 1, Err: errors.New("first failure")}
		}
		close(retryStarted)
		<-releaseRetry
		return TodoLifecycleScriptRunResult{Output: "second failure output", ExitCode: 2, Err: errors.New("second failure")}
	})
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
		t.Fatalf("Start(first) started = %v, error = %v; want started", started, err)
	}
	waitForLifecycleScriptStatus(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, TodoLifecycleScriptStatusFailed)
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, "first failure output")

	if _, started, err := executor.Start(context.Background(), request); err != nil || !started {
		t.Fatalf("Start(retry) started = %v, error = %v; want started", started, err)
	}
	<-retryStarted
	assertNoLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit)

	releaseRetry <- struct{}{}
	waitForLifecycleScriptStatus(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, TodoLifecycleScriptStatusFailed)
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, "second failure output")
}

func TestTodoLifecycleScriptExecutorSuccessfulRetryClearsFailedOutput(t *testing.T) {
	statuses := make(chan TodoLifecycleScriptStatus, 8)
	attempts := 0
	executor := NewTodoLifecycleScriptExecutor(
		func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			attempts++
			if attempts == 1 {
				return TodoLifecycleScriptRunResult{Output: "failure output", ExitCode: 1, Err: errors.New("failure")}
			}
			return TodoLifecycleScriptRunResult{}
		},
		WithTodoLifecycleScriptStatusHandler(func(status TodoLifecycleScriptStatus) {
			statuses <- status
		}),
	)
	request := TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseComplete,
		ScriptName: "Finish checks",
		Script:     "npm test",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	}

	if _, started, err := executor.Start(context.Background(), request); err != nil || !started {
		t.Fatalf("Start(first) started = %v, error = %v; want started", started, err)
	}
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusQueued)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusRunning)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusFailed)
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseComplete, "failure output")

	if _, started, err := executor.Start(context.Background(), request); err != nil || !started {
		t.Fatalf("Start(retry) started = %v, error = %v; want started", started, err)
	}
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusQueued)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), TodoLifecycleScriptStatusRunning)
	assertLifecycleScriptEventStatus(t, receiveLifecycleExecutorStatus(t, statuses), "")
	assertNoLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseComplete)
}

func TestTodoLifecycleScriptExecutorCleanupRemovesFailedOutput(t *testing.T) {
	executor := NewTodoLifecycleScriptExecutor(func(_ context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
		return TodoLifecycleScriptRunResult{
			Output:   request.TodoID + ":" + request.Phase,
			ExitCode: 1,
			Err:      errors.New("failure"),
		}
	})
	startFailure := func(todoID string, phase string) {
		t.Helper()
		_, started, err := executor.Start(context.Background(), TodoLifecycleScriptRunRequest{
			TodoID:     todoID,
			Phase:      phase,
			ScriptName: "Lifecycle script",
			Script:     "exit 1",
			WorkingDir: t.TempDir(),
			ShellPath:  "/bin/sh",
			GOOS:       "linux",
		})
		if err != nil || !started {
			t.Fatalf("Start(%s, %s) started = %v, error = %v; want started", todoID, phase, started, err)
		}
		waitForLifecycleScriptStatus(t, executor, todoID, phase, TodoLifecycleScriptStatusFailed)
	}

	startFailure("todo-a", TodoLifecycleScriptPhaseInit)
	startFailure("todo-a", TodoLifecycleScriptPhaseComplete)
	startFailure("todo-b", TodoLifecycleScriptPhaseInit)

	executor.Clear("todo-a", TodoLifecycleScriptPhaseInit)
	assertNoLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit)
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseComplete, "todo-a:complete")

	executor.ClearTodo("todo-a")
	assertNoLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseComplete)
	assertLifecycleScriptFailureOutput(t, executor, "todo-b", TodoLifecycleScriptPhaseInit, "todo-b:init")

	executor.ClearAll()
	assertNoLifecycleScriptFailureOutput(t, executor, "todo-b", TodoLifecycleScriptPhaseInit)
	if statuses := executor.Statuses(); len(statuses) != 0 {
		t.Fatalf("Statuses() after ClearAll = %#v, want empty", statuses)
	}
}

func TestTodoLifecycleScriptExecutorClearAllInvalidatesInFlightRun(t *testing.T) {
	oldStarted := make(chan struct{})
	newStarted := make(chan struct{})
	oldReturned := make(chan struct{})
	releaseOld := make(chan struct{}, 1)
	releaseNew := make(chan struct{}, 1)
	defer close(releaseOld)
	defer close(releaseNew)
	failedStatuses := make(chan TodoLifecycleScriptStatus, 2)
	executor := NewTodoLifecycleScriptExecutor(
		func(_ context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			switch request.Script {
			case "old-script":
				close(oldStarted)
				<-releaseOld
				close(oldReturned)
				return TodoLifecycleScriptRunResult{Output: "old failure output", ExitCode: 1, Err: errors.New("old failure")}
			case "new-script":
				close(newStarted)
				<-releaseNew
				return TodoLifecycleScriptRunResult{Output: "new failure output", ExitCode: 2, Err: errors.New("new failure")}
			default:
				return TodoLifecycleScriptRunResult{ExitCode: -1, Err: errors.New("unexpected script: " + request.Script)}
			}
		},
		WithTodoLifecycleScriptStatusHandler(func(status TodoLifecycleScriptStatus) {
			if status.Status == TodoLifecycleScriptStatusFailed {
				failedStatuses <- status
			}
		}),
	)
	oldRequest := TodoLifecycleScriptRunRequest{
		TodoID:     "todo-a",
		Phase:      TodoLifecycleScriptPhaseInit,
		ScriptName: "Old workspace script",
		Script:     "old-script",
		WorkingDir: t.TempDir(),
		ShellPath:  "/bin/sh",
		GOOS:       "linux",
	}

	if _, started, err := executor.Start(context.Background(), oldRequest); err != nil || !started {
		t.Fatalf("Start(old) started = %v, error = %v; want started", started, err)
	}
	<-oldStarted
	executor.ClearAll()
	if statuses := executor.Statuses(); len(statuses) != 0 {
		t.Fatalf("Statuses() after ClearAll = %#v, want empty", statuses)
	}

	newRequest := oldRequest
	newRequest.ScriptName = "New workspace script"
	newRequest.Script = "new-script"
	newRequest.WorkingDir = t.TempDir()
	if _, started, err := executor.Start(context.Background(), newRequest); err != nil || !started {
		t.Fatalf("Start(new) started = %v, error = %v; want started", started, err)
	}
	<-newStarted
	releaseNew <- struct{}{}
	newFailed := receiveLifecycleExecutorStatus(t, failedStatuses)
	if newFailed.ScriptName != "New workspace script" {
		t.Fatalf("first failed event = %#v, want new run", newFailed)
	}
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, "new failure output")

	releaseOld <- struct{}{}
	<-oldReturned
	select {
	case stale := <-failedStatuses:
		current, _ := executor.Status("todo-a", TodoLifecycleScriptPhaseInit)
		output, _ := executor.FailureOutput("todo-a", TodoLifecycleScriptPhaseInit)
		t.Fatalf("stale run emitted after ClearAll: event=%#v current=%#v output=%q", stale, current, output)
	case <-time.After(100 * time.Millisecond):
	}

	status, ok := executor.Status("todo-a", TodoLifecycleScriptPhaseInit)
	if !ok || status.Status != TodoLifecycleScriptStatusFailed || status.ScriptName != "New workspace script" {
		t.Fatalf("Status() after old run returned = %#v, found = %v; want new failed run", status, ok)
	}
	assertLifecycleScriptFailureOutput(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, "new failure output")
}

func TestTodoLifecycleScriptExecutorClearAllFromQueuedHandlerPreventsRunningWriteback(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	var executor *TodoLifecycleScriptExecutor
	executor = NewTodoLifecycleScriptExecutor(
		func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			<-release
			return TodoLifecycleScriptRunResult{}
		},
		WithTodoLifecycleScriptStatusHandler(func(status TodoLifecycleScriptStatus) {
			if status.Status == TodoLifecycleScriptStatusQueued {
				executor.ClearAll()
			}
		}),
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
	if err != nil || !started {
		t.Fatalf("Start() started = %v, error = %v; want started", started, err)
	}
	if status, ok := executor.Status("todo-a", TodoLifecycleScriptPhaseInit); ok {
		t.Fatalf("Status() after queued handler ClearAll = %#v, want cleared", status)
	}
}

func TestTodoLifecycleScriptExecutorResetScopeRejectsStartFromPreviousScope(t *testing.T) {
	runnerCalled := make(chan struct{}, 1)
	executor := NewTodoLifecycleScriptExecutor(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
		runnerCalled <- struct{}{}
		return TodoLifecycleScriptRunResult{}
	})
	executor.ResetScope("workspace-a")
	request := TodoLifecycleScriptRunRequest{
		TodoID:         "todo-a",
		Phase:          TodoLifecycleScriptPhaseInit,
		ScriptName:     "Old workspace script",
		Script:         "npm install",
		WorkingDir:     t.TempDir(),
		ShellPath:      "/bin/sh",
		GOOS:           "linux",
		WorkspaceScope: "workspace-a",
	}

	executor.ResetScope("workspace-b")
	_, started, err := executor.Start(context.Background(), request)
	if err == nil {
		t.Fatal("Start(previous scope) error = nil, want stale scope error")
	}
	if started {
		t.Fatal("Start(previous scope) started = true, want false")
	}
	if statuses := executor.Statuses(); len(statuses) != 0 {
		t.Fatalf("Statuses() after stale Start = %#v, want empty", statuses)
	}
	select {
	case <-runnerCalled:
		t.Fatal("runner called for stale workspace scope")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestTodoLifecycleScriptExecutorTagsLateFailedEventWithPreviousScopeEpoch(t *testing.T) {
	oldFailedEntered := make(chan TodoLifecycleScriptStatus, 1)
	releaseOldDelivery := make(chan struct{}, 1)
	oldFailedDelivered := make(chan TodoLifecycleScriptStatus, 1)
	releaseNewRun := make(chan struct{})
	defer close(releaseNewRun)
	executor := NewTodoLifecycleScriptExecutor(
		func(_ context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
			if request.WorkspaceScope == "workspace-a" {
				return TodoLifecycleScriptRunResult{Output: "old failure", ExitCode: 1, Err: errors.New("old failure")}
			}
			<-releaseNewRun
			return TodoLifecycleScriptRunResult{}
		},
		WithTodoLifecycleScriptStatusHandler(func(status TodoLifecycleScriptStatus) {
			if status.Status != TodoLifecycleScriptStatusFailed || status.ScriptName != "Old workspace script" {
				return
			}
			oldFailedEntered <- status
			<-releaseOldDelivery
			oldFailedDelivered <- status
		}),
	)
	executor.ResetScope("workspace-a")
	oldRequest := TodoLifecycleScriptRunRequest{
		TodoID:         "todo-a",
		Phase:          TodoLifecycleScriptPhaseInit,
		ScriptName:     "Old workspace script",
		Script:         "old-script",
		WorkingDir:     t.TempDir(),
		ShellPath:      "/bin/sh",
		GOOS:           "linux",
		WorkspaceScope: "workspace-a",
	}
	if _, started, err := executor.Start(context.Background(), oldRequest); err != nil || !started {
		t.Fatalf("Start(old scope) started = %v, error = %v; want started", started, err)
	}
	oldFailed := receiveLifecycleExecutorStatus(t, oldFailedEntered)

	executor.ResetScope("workspace-b")
	newRequest := oldRequest
	newRequest.ScriptName = "New workspace script"
	newRequest.Script = "new-script"
	newRequest.WorkspaceScope = "workspace-b"
	newStatus, started, err := executor.Start(context.Background(), newRequest)
	if err != nil || !started {
		t.Fatalf("Start(new scope) started = %v, error = %v; want started", started, err)
	}
	if oldFailed.ScopeEpoch >= newStatus.ScopeEpoch {
		t.Fatalf("old ScopeEpoch = %d, new ScopeEpoch = %d; want old < new", oldFailed.ScopeEpoch, newStatus.ScopeEpoch)
	}
	if oldFailed.RunID == 0 || newStatus.RunID == 0 || oldFailed.RunID == newStatus.RunID {
		t.Fatalf("old RunID = %d, new RunID = %d; want distinct non-zero IDs", oldFailed.RunID, newStatus.RunID)
	}

	payload, err := json.Marshal(oldFailed)
	if err != nil {
		t.Fatalf("Marshal(old failed status) error = %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("Unmarshal(old failed status) error = %v", err)
	}
	if _, ok := fields["runId"]; !ok {
		t.Fatalf("status JSON = %s, want runId", payload)
	}
	if _, ok := fields["scopeEpoch"]; !ok {
		t.Fatalf("status JSON = %s, want scopeEpoch", payload)
	}

	releaseOldDelivery <- struct{}{}
	delivered := receiveLifecycleExecutorStatus(t, oldFailedDelivered)
	if delivered.ScopeEpoch >= newStatus.ScopeEpoch {
		t.Fatalf("late event ScopeEpoch = %d, current ScopeEpoch = %d; want late < current", delivered.ScopeEpoch, newStatus.ScopeEpoch)
	}
}

func TestTodoLifecycleScriptExecutorSnapshotKeepsStatusesAndScopeEpochConsistent(t *testing.T) {
	executor := NewTodoLifecycleScriptExecutor(func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
		return TodoLifecycleScriptRunResult{Output: "failure output", ExitCode: 1, Err: errors.New("failure")}
	})
	executor.ResetScope("workspace-a")
	_, started, err := executor.Start(context.Background(), TodoLifecycleScriptRunRequest{
		TodoID:         "todo-a",
		Phase:          TodoLifecycleScriptPhaseInit,
		ScriptName:     "Old workspace script",
		Script:         "exit 1",
		WorkingDir:     t.TempDir(),
		WorkspaceScope: "workspace-a",
		ShellPath:      "/bin/sh",
		GOOS:           "linux",
	})
	if err != nil || !started {
		t.Fatalf("Start() started = %v, error = %v; want started", started, err)
	}
	waitForLifecycleScriptStatus(t, executor, "todo-a", TodoLifecycleScriptPhaseInit, TodoLifecycleScriptStatusFailed)

	statuses, epoch := executor.Snapshot()
	if len(statuses) != 1 {
		t.Fatalf("Snapshot() statuses = %#v, want one failed status", statuses)
	}
	if statuses[0].ScopeEpoch != epoch {
		t.Fatalf("Snapshot() status epoch = %d, snapshot epoch = %d, want equal", statuses[0].ScopeEpoch, epoch)
	}

	executor.ResetScope("workspace-b")
	nextStatuses, nextEpoch := executor.Snapshot()
	if len(nextStatuses) != 0 {
		t.Fatalf("Snapshot() after ResetScope statuses = %#v, want empty", nextStatuses)
	}
	if nextEpoch <= epoch {
		t.Fatalf("Snapshot() next epoch = %d, want greater than %d", nextEpoch, epoch)
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

func assertLifecycleScriptFailureOutput(t *testing.T, executor *TodoLifecycleScriptExecutor, todoID string, phase string, want string) {
	t.Helper()
	output, ok := executor.FailureOutput(todoID, phase)
	if !ok {
		t.Fatalf("FailureOutput(%q, %q) found = false, want true", todoID, phase)
	}
	if output != want {
		t.Fatalf("FailureOutput(%q, %q) = %q, want %q", todoID, phase, output, want)
	}
}

func assertNoLifecycleScriptFailureOutput(t *testing.T, executor *TodoLifecycleScriptExecutor, todoID string, phase string) {
	t.Helper()
	if output, ok := executor.FailureOutput(todoID, phase); ok {
		t.Fatalf("FailureOutput(%q, %q) = %q, want no output", todoID, phase, output)
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
