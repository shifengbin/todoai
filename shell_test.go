package main

import (
	"errors"
	"io"
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

type fakePtyProcess struct {
	output  chan string
	wait    chan error
	written string
	sizes   []TerminalSize
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
