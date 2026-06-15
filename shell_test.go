package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestShellSessionManagerCreatesDefaultTerminalInProjectDirectoryAndReusesIt(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return "/custom/shell" }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)

	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.EnsureProjectTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("EnsureProjectTerminal() error = %v", err)
	}
	if terminal.ID != "terminal-1" {
		t.Fatalf("Terminal ID = %q, want terminal-1", terminal.ID)
	}
	if terminal.ProjectID != "project-a" {
		t.Fatalf("ProjectID = %q, want project-a", terminal.ProjectID)
	}
	if terminal.State != ShellStateRunning {
		t.Fatalf("State = %q, want %q", terminal.State, ShellStateRunning)
	}
	if terminal.ShellName != "shell" {
		t.Fatalf("ShellName = %q, want shell", terminal.ShellName)
	}

	second, err := manager.EnsureProjectTerminal(project, TerminalSize{Cols: 120, Rows: 30})
	if err != nil {
		t.Fatalf("second EnsureProjectTerminal() error = %v", err)
	}
	if second.ID != terminal.ID {
		t.Fatalf("second terminal ID = %q, want %q", second.ID, terminal.ID)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].TerminalID != "terminal-1" {
		t.Fatalf("TerminalID = %q, want terminal-1", starter.requests[0].TerminalID)
	}
	if starter.requests[0].ProjectID != "project-a" {
		t.Fatalf("ProjectID = %q, want project-a", starter.requests[0].ProjectID)
	}
	if starter.requests[0].WorkingDir != project.Path {
		t.Fatalf("WorkingDir = %q, want %q", starter.requests[0].WorkingDir, project.Path)
	}
	if starter.requests[0].ShellPath != "/custom/shell" {
		t.Fatalf("ShellPath = %q, want /custom/shell", starter.requests[0].ShellPath)
	}
	if starter.requests[0].Size != (TerminalSize{Cols: 80, Rows: 24}) {
		t.Fatalf("Size = %#v, want 80x24", starter.requests[0].Size)
	}
}

func TestShellSessionManagerCreatesMultipleTerminalsForSameProject(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return "/bin/zsh" }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminalA, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal(A) error = %v", err)
	}
	terminalB, err := manager.CreateTerminal(project, TerminalSize{Cols: 100, Rows: 32})
	if err != nil {
		t.Fatalf("CreateTerminal(B) error = %v", err)
	}

	if terminalA.ID == terminalB.ID {
		t.Fatalf("terminal IDs should be distinct, both were %q", terminalA.ID)
	}
	if terminalA.ProjectID != "project-a" || terminalB.ProjectID != "project-a" {
		t.Fatalf("ProjectIDs = %q, %q; want both project-a", terminalA.ProjectID, terminalB.ProjectID)
	}
	if len(starter.requests) != 2 {
		t.Fatalf("start count = %d, want 2", len(starter.requests))
	}
	if starter.requests[0].TerminalID != "terminal-a" || starter.requests[1].TerminalID != "terminal-b" {
		t.Fatalf("TerminalIDs = %q, %q; want terminal-a, terminal-b", starter.requests[0].TerminalID, starter.requests[1].TerminalID)
	}
	if starter.requests[0].WorkingDir != project.Path || starter.requests[1].WorkingDir != project.Path {
		t.Fatalf("shells should start in project path %q", project.Path)
	}
}

func TestShellSessionManagerIsolatesTerminalsByTodoProjectContext(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	todoProjectA := TodoProject{ID: "todo-project-a", TodoID: "todo-a", ProjectID: project.ID}
	todoProjectB := TodoProject{ID: "todo-project-b", TodoID: "todo-b", ProjectID: project.ID}

	terminalA, err := manager.CreateTodoProjectTerminal(todoProjectA, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(A) error = %v", err)
	}
	terminalB, err := manager.CreateTodoProjectTerminal(todoProjectB, project, TerminalSize{Cols: 100, Rows: 32})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(B) error = %v", err)
	}

	if terminalA.TodoID != "todo-a" || terminalA.TodoProjectID != "todo-project-a" {
		t.Fatalf("terminal A TODO context = %q/%q, want todo-a/todo-project-a", terminalA.TodoID, terminalA.TodoProjectID)
	}
	if terminalB.TodoID != "todo-b" || terminalB.TodoProjectID != "todo-project-b" {
		t.Fatalf("terminal B TODO context = %q/%q, want todo-b/todo-project-b", terminalB.TodoID, terminalB.TodoProjectID)
	}
	if manager.ActiveTerminalID("todo-project-a") != terminalA.ID {
		t.Fatalf("ActiveTerminalID(todo-project-a) = %q, want %q", manager.ActiveTerminalID("todo-project-a"), terminalA.ID)
	}
	if manager.ActiveTerminalID("todo-project-b") != terminalB.ID {
		t.Fatalf("ActiveTerminalID(todo-project-b) = %q, want %q", manager.ActiveTerminalID("todo-project-b"), terminalB.ID)
	}

	if _, err := manager.SelectTerminal(terminalA.ID); err != nil {
		t.Fatalf("SelectTerminal(A) error = %v", err)
	}
	if manager.ActiveTerminalID("todo-project-b") != terminalB.ID {
		t.Fatalf("selecting terminal A changed todo-project-b active terminal to %q", manager.ActiveTerminalID("todo-project-b"))
	}
	if len(starter.requests) != 2 {
		t.Fatalf("start count = %d, want 2", len(starter.requests))
	}
}

func TestShellSessionManagerRoutesInputAndResizeByTerminalID(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminalA, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal(A) error = %v", err)
	}
	terminalB, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal(B) error = %v", err)
	}

	if err := manager.WriteInput(terminalB.ID, "pwd\n"); err != nil {
		t.Fatalf("WriteInput() error = %v", err)
	}
	if err := manager.Resize(terminalB.ID, TerminalSize{Cols: 100, Rows: 32}); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}

	if got := starter.processes[0].written; got != "" {
		t.Fatalf("project A written input = %q, want empty", got)
	}
	if got := starter.processes[1].written; got != "pwd\n" {
		t.Fatalf("project B written input = %q, want pwd newline", got)
	}
	if got := starter.processes[1].sizes[len(starter.processes[1].sizes)-1]; got != (TerminalSize{Cols: 100, Rows: 32}) {
		t.Fatalf("project B size = %#v, want 100x32", got)
	}
	if terminalA.ID == terminalB.ID {
		t.Fatal("test setup created duplicate terminal IDs")
	}
}

func TestShellSessionManagerStartsShellWithEmbeddedTerminalEnvironment(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("COLORTERM", "")

	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if got := envValue(starter.requests[0].Env, "TERM"); got != "xterm-256color" {
		t.Fatalf("TERM = %q, want xterm-256color", got)
	}
	if got := envValue(starter.requests[0].Env, "COLORTERM"); got != "truecolor" {
		t.Fatalf("COLORTERM = %q, want truecolor", got)
	}
}

func TestShellSessionManagerStartsShellWithTerminalIdentityEnvironment(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	env := starter.requests[0].Env
	if got := envValue(env, "TUI_HELPER_TERMINAL_ID"); got != "terminal-1" {
		t.Fatalf("TUI_HELPER_TERMINAL_ID = %q, want terminal-1", got)
	}
	if got := envValue(env, "TUI_HELPER_PROJECT_ID"); got != "project-a" {
		t.Fatalf("TUI_HELPER_PROJECT_ID = %q, want project-a", got)
	}
}

func TestShellSessionManagerSerializesConcurrentStartsForSameTerminal(t *testing.T) {
	starter := newBlockingShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	terminal, err := manager.RegisterTerminal(project)
	if err != nil {
		t.Fatalf("RegisterTerminal() error = %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 2)
	start := func() {
		defer wg.Done()
		_, err := manager.StartTerminal(terminal.ID, TerminalSize{Cols: 80, Rows: 24})
		errors <- err
	}

	wg.Add(1)
	go start()

	select {
	case <-starter.firstStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first shell start")
	}

	wg.Add(1)
	go start()

	secondStarted := false
	select {
	case <-starter.secondStarted:
		secondStarted = true
	case <-time.After(50 * time.Millisecond):
	}

	close(starter.release)
	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("EnsureSession() error = %v", err)
		}
	}
	if secondStarted {
		t.Fatal("second shell start began before first session was stored")
	}
	if got := starter.startCount(); got != 1 {
		t.Fatalf("start count = %d, want 1", got)
	}
}

func TestShellSessionManagerEmitsOutputWithProjectAndTerminalID(t *testing.T) {
	starter := newFakeShellStarter()
	outputs := make(chan TerminalOutputEvent, 1)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{
			OnOutput: func(event TerminalOutputEvent) {
				outputs <- event
			},
		},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	starter.processes[0].emit("hello")

	select {
	case event := <-outputs:
		if event.ProjectID != "project-a" {
			t.Fatalf("ProjectID = %q, want project-a", event.ProjectID)
		}
		if event.TerminalID != "terminal-1" {
			t.Fatalf("TerminalID = %q, want terminal-1", event.TerminalID)
		}
		if event.Data != "hello" {
			t.Fatalf("Data = %q, want hello", event.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for terminal output event")
	}
}

func TestShellSessionManagerFiltersCommandStatePayloadFromOutputAndHistory(t *testing.T) {
	starter := newFakeShellStarter()
	outputs := make(chan TerminalOutputEvent, 1)
	commandStates := make(chan TerminalCommandStateEvent, 1)
	store := NewTerminalHistoryStore(t.TempDir())
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{
			OnOutput: func(event TerminalOutputEvent) {
				outputs <- event
			},
			OnCommandState: func(event TerminalCommandStateEvent) {
				commandStates <- event
			},
		},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
		WithTerminalHistoryStore(store),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	starter.processes[0].emit("before\x1b]777;todoai;command-start;Y29kZXg=\aafter")

	select {
	case event := <-outputs:
		if event.Data != "beforeafter" {
			t.Fatalf("Data = %q, want beforeafter", event.Data)
		}
		if strings.Contains(event.Data, "todoai") || strings.Contains(event.Data, "Y29kZXg") {
			t.Fatalf("Data contains private payload: %q", event.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for terminal output event")
	}

	select {
	case event := <-commandStates:
		if event.ProjectID != "project-a" || event.TerminalID != "terminal-1" {
			t.Fatalf("Command state identity = %#v, want project-a/terminal-1", event)
		}
		if event.Type != "command-start" || event.Command != "codex" {
			t.Fatalf("Command state = %#v, want command-start codex", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command-state event")
	}

	eventually(t, func() bool {
		history, err := store.Load()
		if err != nil || len(history.Records) != 1 {
			return false
		}
		return history.Records[0].Output == "beforeafter"
	})
}

func TestShellSessionManagerEmitsExitedStatusWithProjectAndTerminalID(t *testing.T) {
	starter := newFakeShellStarter()
	statuses := make(chan ShellStatus, 1)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{
			OnStatus: func(status ShellStatus) {
				statuses <- status
			},
		},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	starter.processes[0].exit(errors.New("shell exited"))

	select {
	case status := <-statuses:
		if status.ProjectID != "project-a" {
			t.Fatalf("ProjectID = %q, want project-a", status.ProjectID)
		}
		if status.TerminalID != "terminal-1" {
			t.Fatalf("TerminalID = %q, want terminal-1", status.TerminalID)
		}
		if status.State != ShellStateExited {
			t.Fatalf("State = %q, want %q", status.State, ShellStateExited)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for shell status event")
	}
}

func TestShellSessionManagerRestartsExitedTerminalWithSameID(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	starter.processes[0].exit(errors.New("shell exited"))
	eventually(t, func() bool {
		return manager.Status(terminal.ID).State == ShellStateExited
	})

	status, err := manager.StartTerminal(terminal.ID, TerminalSize{Cols: 120, Rows: 30})
	if err != nil {
		t.Fatalf("StartTerminal() error = %v", err)
	}

	if status.TerminalID != "terminal-1" {
		t.Fatalf("TerminalID = %q, want terminal-1", status.TerminalID)
	}
	if status.State != ShellStateRunning {
		t.Fatalf("State = %q, want running", status.State)
	}
	if len(starter.requests) != 2 {
		t.Fatalf("start count = %d, want 2", len(starter.requests))
	}
	if starter.requests[1].TerminalID != "terminal-1" {
		t.Fatalf("restart TerminalID = %q, want terminal-1", starter.requests[1].TerminalID)
	}
}

func TestShellSessionManagerStartsSupportedShellsWithCommandLabelIntegration(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return "/bin/zsh" }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	request := starter.requests[0]
	if terminal.ShellName != "zsh" {
		t.Fatalf("ShellName = %q, want zsh", terminal.ShellName)
	}
	if len(request.ShellArgs) == 0 {
		t.Fatal("ShellArgs is empty, want shell integration startup args")
	}
	if envValue(request.Env, "ZDOTDIR") == "" {
		t.Fatal("ZDOTDIR is empty, want zsh integration env")
	}
}

func TestShellSessionManagerStartsPowerShellWithCommandLabelIntegration(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return `C:\Program Files\PowerShell\7\pwsh.exe` }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	request := starter.requests[0]
	if terminal.ShellName != "pwsh" {
		t.Fatalf("ShellName = %q, want pwsh", terminal.ShellName)
	}
	if len(request.ShellArgs) < 4 {
		t.Fatalf("ShellArgs = %#v, want PowerShell integration args", request.ShellArgs)
	}
	if !containsShellArgPair(request.ShellArgs, "-ExecutionPolicy", "Bypass") {
		t.Fatalf("ShellArgs = %#v, want execution policy bypass", request.ShellArgs)
	}
	if containsShellArg(request.ShellArgs, "-NoProfile") {
		t.Fatalf("ShellArgs = %#v, should not skip user profile", request.ShellArgs)
	}
	if request.ShellArgs[0] != "-NoLogo" || request.ShellArgs[1] != "-NoExit" || request.ShellArgs[len(request.ShellArgs)-2] != "-Command" {
		t.Fatalf("ShellArgs = %#v, want -NoLogo -NoExit ... -Command", request.ShellArgs)
	}
	scriptPath := powerShellIntegrationPathFromCommand(t, request.ShellArgs[len(request.ShellArgs)-1])
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", scriptPath, err)
	}
	scriptText := string(script)
	for _, want := range []string{
		"todoai;command-start;",
		"todoai;command-end",
		"Set-PSReadLineOption -AddToHistoryHandler",
		"function global:prompt",
	} {
		if !strings.Contains(scriptText, want) {
			t.Fatalf("PowerShell integration script missing %q:\n%s", want, scriptText)
		}
	}

	if err := manager.DeleteTerminal(terminal.ID); err != nil {
		t.Fatalf("DeleteTerminal() error = %v", err)
	}
	if _, err := os.Stat(scriptPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("PowerShell integration script still exists after cleanup: %v", err)
	}
}

func TestPowerShellIntegrationSupportsWindowsPowerShellName(t *testing.T) {
	launch, err := IntegratedShellLaunch(`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, []string{})
	if err != nil {
		t.Fatalf("IntegratedShellLaunch() error = %v", err)
	}
	defer launch.Cleanup()

	if launch.ShellName != "powershell" {
		t.Fatalf("ShellName = %q, want powershell", launch.ShellName)
	}
	if !containsShellArgPair(launch.Args, "-ExecutionPolicy", "Bypass") || launch.Args[len(launch.Args)-2] != "-Command" {
		t.Fatalf("Args = %#v, want PowerShell integration args", launch.Args)
	}
}

func containsShellArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func containsShellArgPair(args []string, left string, right string) bool {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == left && args[index+1] == right {
			return true
		}
	}
	return false
}

func powerShellIntegrationPathFromCommand(t *testing.T, command string) string {
	t.Helper()
	const prefix = ". '"
	const suffix = "'"
	if !strings.HasPrefix(command, prefix) || !strings.HasSuffix(command, suffix) {
		t.Fatalf("PowerShell command = %q, want dot-sourced script path", command)
	}
	return strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(command, prefix), suffix), "''", "'")
}

func TestBashIntegrationSkipsPromptCommandWhileIdle(t *testing.T) {
	script := bashIntegrationScript()

	if !strings.Contains(script, "__tui_helper_in_prompt") {
		t.Fatal("bash integration should guard DEBUG trap while PROMPT_COMMAND runs")
	}
}

func TestShellSessionManagerFallsBackToShellNameWithoutIntegrationForUnsupportedShells(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return "/opt/custom/fish" }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-1")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if terminal.ShellName != "fish" {
		t.Fatalf("ShellName = %q, want fish", terminal.ShellName)
	}
	if len(starter.requests[0].ShellArgs) != 0 {
		t.Fatalf("ShellArgs = %#v, want no integration args", starter.requests[0].ShellArgs)
	}
}

func TestShellSessionManagerSurfacesStartupErrorForWindowsShellPath(t *testing.T) {
	cmdPath := `C:\Windows\System32\cmd.exe`
	wantErr := errors.New("pty unsupported")
	starter := newFailingShellStarter(wantErr)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return cmdPath }),
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); !errors.Is(err, wantErr) {
		t.Fatalf("CreateTerminal() error = %v, want %v", err, wantErr)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("start count = %d, want 1", len(starter.requests))
	}
	if starter.requests[0].ShellPath != cmdPath {
		t.Fatalf("ShellPath = %q, want %q", starter.requests[0].ShellPath, cmdPath)
	}
	if starter.requests[0].ShellPath == "/bin/sh" || starter.requests[0].ShellPath == "/bin/bash" {
		t.Fatalf("ShellPath used Unix fallback %q", starter.requests[0].ShellPath)
	}
}

func TestShellSessionManagerKeepsUnsupportedEmbeddedTerminal(t *testing.T) {
	starter := newFailingShellStarter(ErrEmbeddedShellUnsupported)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v, want nil", err)
	}
	if terminal.State != ShellStateUnsupported {
		t.Fatalf("State = %q, want %q", terminal.State, ShellStateUnsupported)
	}
	status := manager.Status(terminal.ID)
	if status.State != ShellStateUnsupported {
		t.Fatalf("Status.State = %q, want %q", status.State, ShellStateUnsupported)
	}
	if len(starter.requests) != 1 {
		t.Fatalf("start count = %d, want 1", len(starter.requests))
	}
}

func TestShellSessionManagerDeletesRunningTerminalAndClosesProcess(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if err := manager.DeleteTerminal(terminal.ID); err != nil {
		t.Fatalf("DeleteTerminal() error = %v", err)
	}

	if len(manager.Terminals()) != 0 {
		t.Fatalf("Terminals length = %d, want 0", len(manager.Terminals()))
	}
	if got := manager.ActiveTerminalID(project.ID); got != "" {
		t.Fatalf("ActiveTerminalID = %q, want empty", got)
	}
	if !starter.processes[0].closed {
		t.Fatal("deleted running terminal process was not closed")
	}
}

func TestShellSessionManagerShutdownClosesRunningProcess(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}

	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	manager.Shutdown()

	if !starter.processes[0].closed {
		t.Fatal("running terminal process was not closed")
	}
}

func TestShellSessionManagerDeletesExitedTerminalWithoutStartingReplacement(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	terminal, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}
	starter.processes[0].exit(errors.New("shell exited"))
	eventually(t, func() bool {
		return manager.Status(terminal.ID).State == ShellStateExited
	})

	if err := manager.DeleteTerminal(terminal.ID); err != nil {
		t.Fatalf("DeleteTerminal() error = %v", err)
	}

	if len(manager.Terminals()) != 0 {
		t.Fatalf("Terminals length = %d, want 0", len(manager.Terminals()))
	}
	if len(starter.requests) != 1 {
		t.Fatalf("start count = %d, want 1", len(starter.requests))
	}
}

func TestShellSessionManagerDeleteTerminalSelectsRemainingProjectTerminal(t *testing.T) {
	starter := newFakeShellStarter()
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
		WithShellClock(func() time.Time { return now }),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	terminalA, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal(A) error = %v", err)
	}
	now = now.Add(time.Minute)
	terminalB, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTerminal(B) error = %v", err)
	}
	if got := manager.ActiveTerminalID(project.ID); got != terminalB.ID {
		t.Fatalf("ActiveTerminalID setup = %q, want %q", got, terminalB.ID)
	}

	if err := manager.DeleteTerminal(terminalB.ID); err != nil {
		t.Fatalf("DeleteTerminal() error = %v", err)
	}

	if got := manager.ActiveTerminalID(project.ID); got != terminalA.ID {
		t.Fatalf("ActiveTerminalID = %q, want %q", got, terminalA.ID)
	}
}

func TestShellSessionManagerDeleteTerminalSelectsRemainingTodoProjectTerminal(t *testing.T) {
	starter := newFakeShellStarter()
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
		WithShellClock(func() time.Time { return now }),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	todoProject := TodoProject{ID: "todo-project-a", TodoID: "todo-a", ProjectID: project.ID}

	terminalA, err := manager.CreateTodoProjectTerminal(todoProject, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(A) error = %v", err)
	}
	now = now.Add(time.Minute)
	terminalB, err := manager.CreateTodoProjectTerminal(todoProject, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(B) error = %v", err)
	}
	if manager.ActiveTerminalID(todoProject.ID) != terminalB.ID {
		t.Fatalf("ActiveTerminalID setup = %q, want %q", manager.ActiveTerminalID(todoProject.ID), terminalB.ID)
	}

	if err := manager.DeleteTerminal(terminalB.ID); err != nil {
		t.Fatalf("DeleteTerminal(B) error = %v", err)
	}

	if manager.ActiveTerminalID(todoProject.ID) != terminalA.ID {
		t.Fatalf("ActiveTerminalID after delete = %q, want %q", manager.ActiveTerminalID(todoProject.ID), terminalA.ID)
	}
}

func TestShellSessionManagerDeleteTerminalReturnsErrorWhenMissing(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	if _, err := manager.CreateTerminal(project, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal() error = %v", err)
	}

	if err := manager.DeleteTerminal("missing-terminal"); err == nil {
		t.Fatal("DeleteTerminal() error = nil, want error")
	}

	if len(manager.Terminals()) != 1 {
		t.Fatalf("Terminals length = %d, want 1", len(manager.Terminals()))
	}
}

func TestShellSessionManagerDeletesProjectTerminalsAndPreservesOtherProjects(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b", "terminal-c")),
	)
	projectA := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	projectB := Project{ID: "project-b", Path: t.TempDir(), Available: true}
	if _, err := manager.CreateTerminal(projectA, TerminalSize{Cols: 80, Rows: 24}); err != nil {
		t.Fatalf("CreateTerminal(projectA first) error = %v", err)
	}
	if _, err := manager.CreateTerminal(projectA, TerminalSize{Cols: 100, Rows: 32}); err != nil {
		t.Fatalf("CreateTerminal(projectA second) error = %v", err)
	}
	terminalB, err := manager.CreateTerminal(projectB, TerminalSize{Cols: 120, Rows: 40})
	if err != nil {
		t.Fatalf("CreateTerminal(projectB) error = %v", err)
	}

	manager.DeleteProjectTerminals(projectA.ID)

	terminals := manager.Terminals()
	if len(terminals) != 1 {
		t.Fatalf("Terminals length = %d, want 1", len(terminals))
	}
	if terminals[0].ID != terminalB.ID {
		t.Fatalf("remaining terminal ID = %q, want %q", terminals[0].ID, terminalB.ID)
	}
	if got := manager.ActiveTerminalID(projectA.ID); got != "" {
		t.Fatalf("project A ActiveTerminalID = %q, want empty", got)
	}
	if got := manager.ActiveTerminalID(projectB.ID); got != terminalB.ID {
		t.Fatalf("project B ActiveTerminalID = %q, want %q", got, terminalB.ID)
	}
	if !starter.processes[0].closed || !starter.processes[1].closed {
		t.Fatal("project A terminal processes were not closed")
	}
	if starter.processes[2].closed {
		t.Fatal("project B terminal process was closed")
	}
}

func TestShellSessionManagerDeletesTodoTerminalsAndPreservesOtherTodos(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	todoProjectA := TodoProject{ID: "todo-project-a", TodoID: "todo-a", ProjectID: project.ID}
	todoProjectB := TodoProject{ID: "todo-project-b", TodoID: "todo-b", ProjectID: project.ID}
	terminalA, err := manager.CreateTodoProjectTerminal(todoProjectA, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(A) error = %v", err)
	}
	terminalB, err := manager.CreateTodoProjectTerminal(todoProjectB, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(B) error = %v", err)
	}

	manager.DeleteTodoTerminals("todo-a")

	if _, err := manager.Terminal(terminalA.ID); err == nil {
		t.Fatal("Terminal(A) error = nil, want deleted terminal")
	}
	if remaining, err := manager.Terminal(terminalB.ID); err != nil || remaining.ID != terminalB.ID {
		t.Fatalf("Terminal(B) = %#v, %v; want preserved terminal B", remaining, err)
	}
	if !starter.processes[0].closed {
		t.Fatal("deleted todo terminal process was not closed")
	}
	if starter.processes[1].closed {
		t.Fatal("other todo terminal process was closed")
	}
}

func TestShellSessionManagerDeletesTodoProjectTerminalsAndPreservesSameProjectInOtherTodos(t *testing.T) {
	starter := newFakeShellStarter()
	manager := NewShellSessionManager(
		starter.Start,
		ShellSessionCallbacks{},
		WithShellTerminalIDGenerator(sequenceIDs("terminal-a", "terminal-b", "terminal-c")),
	)
	project := Project{ID: "project-a", Path: t.TempDir(), Available: true}
	todoProjectA := TodoProject{ID: "todo-project-a", TodoID: "todo-a", ProjectID: project.ID}
	todoProjectB := TodoProject{ID: "todo-project-b", TodoID: "todo-b", ProjectID: project.ID}
	terminalA1, err := manager.CreateTodoProjectTerminal(todoProjectA, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(A1) error = %v", err)
	}
	terminalA2, err := manager.CreateTodoProjectTerminal(todoProjectA, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(A2) error = %v", err)
	}
	terminalB, err := manager.CreateTodoProjectTerminal(todoProjectB, project, TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatalf("CreateTodoProjectTerminal(B) error = %v", err)
	}

	manager.DeleteTodoProjectTerminals("todo-project-a")

	if _, err := manager.Terminal(terminalA1.ID); err == nil {
		t.Fatal("Terminal(A1) error = nil, want deleted terminal")
	}
	if _, err := manager.Terminal(terminalA2.ID); err == nil {
		t.Fatal("Terminal(A2) error = nil, want deleted terminal")
	}
	if remaining, err := manager.Terminal(terminalB.ID); err != nil || remaining.ID != terminalB.ID {
		t.Fatalf("Terminal(B) = %#v, %v; want preserved terminal B", remaining, err)
	}
	if got := manager.ActiveTerminalID("todo-project-a"); got != "" {
		t.Fatalf("ActiveTerminalID(todo-project-a) = %q, want empty", got)
	}
	if got := manager.ActiveTerminalID("todo-project-b"); got != terminalB.ID {
		t.Fatalf("ActiveTerminalID(todo-project-b) = %q, want %q", got, terminalB.ID)
	}
	if !starter.processes[0].closed || !starter.processes[1].closed {
		t.Fatal("todo-project A terminal processes were not closed")
	}
	if starter.processes[2].closed {
		t.Fatal("todo-project B terminal process was closed")
	}
}

type fakeShellStarter struct {
	requests  []ShellStartRequest
	processes []*fakePtyProcess
}

func newFakeShellStarter() *fakeShellStarter {
	return &fakeShellStarter{}
}

func (starter *fakeShellStarter) Start(request ShellStartRequest) (PtyProcess, error) {
	process := newFakePtyProcess(request.Size)
	starter.requests = append(starter.requests, request)
	starter.processes = append(starter.processes, process)
	return process, nil
}

type failingShellStarter struct {
	requests []ShellStartRequest
	err      error
}

func newFailingShellStarter(err error) *failingShellStarter {
	return &failingShellStarter{err: err}
}

func (starter *failingShellStarter) Start(request ShellStartRequest) (PtyProcess, error) {
	starter.requests = append(starter.requests, request)
	return nil, starter.err
}

type fakePtyProcess struct {
	output  chan string
	wait    chan error
	written string
	sizes   []TerminalSize
	closed  bool
}

func newFakePtyProcess(size TerminalSize) *fakePtyProcess {
	return &fakePtyProcess{
		output: make(chan string, 8),
		wait:   make(chan error, 1),
		sizes:  []TerminalSize{size},
	}
}

func (process *fakePtyProcess) Read(data []byte) (int, error) {
	output, ok := <-process.output
	if !ok {
		return 0, io.EOF
	}
	return copy(data, output), nil
}

func (process *fakePtyProcess) Write(data []byte) (int, error) {
	process.written += string(data)
	return len(data), nil
}

func (process *fakePtyProcess) Resize(size TerminalSize) error {
	process.sizes = append(process.sizes, size)
	return nil
}

func (process *fakePtyProcess) Wait() error {
	return <-process.wait
}

func (process *fakePtyProcess) Close() error {
	process.closed = true
	close(process.output)
	process.wait <- nil
	return nil
}

func (process *fakePtyProcess) emit(output string) {
	process.output <- output
}

func (process *fakePtyProcess) exit(err error) {
	process.wait <- err
	close(process.output)
}

type blockingShellStarter struct {
	mu            sync.Mutex
	requests      []ShellStartRequest
	firstStarted  chan struct{}
	secondStarted chan struct{}
	release       chan struct{}
}

func newBlockingShellStarter() *blockingShellStarter {
	return &blockingShellStarter{
		firstStarted:  make(chan struct{}),
		secondStarted: make(chan struct{}),
		release:       make(chan struct{}),
	}
}

func (starter *blockingShellStarter) Start(request ShellStartRequest) (PtyProcess, error) {
	starter.mu.Lock()
	starter.requests = append(starter.requests, request)
	count := len(starter.requests)
	if count == 1 {
		close(starter.firstStarted)
	}
	if count == 2 {
		close(starter.secondStarted)
	}
	starter.mu.Unlock()

	<-starter.release
	return newFakePtyProcess(request.Size), nil
}

func (starter *blockingShellStarter) startCount() int {
	starter.mu.Lock()
	defer starter.mu.Unlock()
	return len(starter.requests)
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, entry := range env {
		if len(entry) >= len(prefix) && entry[:len(prefix)] == prefix {
			return entry[len(prefix):]
		}
	}
	return ""
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
