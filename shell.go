package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
)

const (
	ShellStateRunning = "running"
	ShellStateExited  = "exited"
)

type TerminalSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type ShellStartRequest struct {
	ProjectID  string
	WorkingDir string
	ShellPath  string
	Size       TerminalSize
	Env        []string
}

type ShellStatus struct {
	ProjectID string `json:"projectId"`
	State     string `json:"state"`
}

type TerminalOutputEvent struct {
	ProjectID string `json:"projectId"`
	Data      string `json:"data"`
}

type PtyProcess interface {
	io.Reader
	io.Writer
	Resize(size TerminalSize) error
	Wait() error
	Close() error
}

type ShellStarter func(request ShellStartRequest) (PtyProcess, error)

type ShellSessionCallbacks struct {
	OnOutput func(event TerminalOutputEvent)
	OnStatus func(status ShellStatus)
}

type ShellSessionManagerOption func(*ShellSessionManager)

type ShellSessionManager struct {
	mu                sync.Mutex
	starter           ShellStarter
	callbacks         ShellSessionCallbacks
	shellPathResolver func() string
	sessions          map[string]*ShellSession
}

type ShellSession struct {
	projectID string
	process   PtyProcess
	size      TerminalSize
	state     string
}

func NewShellSessionManager(starter ShellStarter, callbacks ShellSessionCallbacks, opts ...ShellSessionManagerOption) *ShellSessionManager {
	manager := &ShellSessionManager{
		starter:           starter,
		callbacks:         callbacks,
		shellPathResolver: DefaultShellPath,
		sessions:          map[string]*ShellSession{},
	}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func WithShellPathResolver(resolve func() string) ShellSessionManagerOption {
	return func(manager *ShellSessionManager) {
		manager.shellPathResolver = resolve
	}
}

func (manager *ShellSessionManager) EnsureSession(project Project, size TerminalSize) (ShellStatus, error) {
	if !project.Available {
		return ShellStatus{}, errors.New("project path is unavailable")
	}

	manager.mu.Lock()
	if session, ok := manager.sessions[project.ID]; ok && session.state == ShellStateRunning {
		status := ShellStatus{ProjectID: project.ID, State: session.state}
		manager.mu.Unlock()
		return status, nil
	}

	request := ShellStartRequest{
		ProjectID:  project.ID,
		WorkingDir: project.Path,
		ShellPath:  manager.shellPathResolver(),
		Size:       size,
		Env:        EmbeddedTerminalEnv(os.Environ()),
	}
	process, err := manager.starter(request)
	if err != nil {
		manager.mu.Unlock()
		return ShellStatus{}, err
	}

	session := &ShellSession{
		projectID: project.ID,
		process:   process,
		size:      size,
		state:     ShellStateRunning,
	}

	manager.sessions[project.ID] = session
	manager.mu.Unlock()

	go manager.readOutput(session)
	go manager.waitForExit(session)

	return ShellStatus{ProjectID: project.ID, State: ShellStateRunning}, nil
}

func (manager *ShellSessionManager) WriteInput(projectID string, data string) error {
	session, err := manager.runningSession(projectID)
	if err != nil {
		return err
	}
	_, err = session.process.Write([]byte(data))
	return err
}

func (manager *ShellSessionManager) Resize(projectID string, size TerminalSize) error {
	session, err := manager.runningSession(projectID)
	if err != nil {
		return err
	}
	if err := session.process.Resize(size); err != nil {
		return err
	}

	manager.mu.Lock()
	session.size = size
	manager.mu.Unlock()
	return nil
}

func (manager *ShellSessionManager) Status(projectID string) ShellStatus {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if session, ok := manager.sessions[projectID]; ok {
		return ShellStatus{ProjectID: projectID, State: session.state}
	}
	return ShellStatus{ProjectID: projectID, State: ShellStateExited}
}

func (manager *ShellSessionManager) Shutdown() {
	manager.mu.Lock()
	sessions := make([]*ShellSession, 0, len(manager.sessions))
	for _, session := range manager.sessions {
		sessions = append(sessions, session)
	}
	manager.mu.Unlock()

	for _, session := range sessions {
		_ = session.process.Close()
	}
}

func (manager *ShellSessionManager) runningSession(projectID string) (*ShellSession, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	session, ok := manager.sessions[projectID]
	if !ok || session.state != ShellStateRunning {
		return nil, errors.New("shell session is not running")
	}
	return session, nil
}

func (manager *ShellSessionManager) readOutput(session *ShellSession) {
	buffer := make([]byte, 4096)
	for {
		n, err := session.process.Read(buffer)
		if n > 0 && manager.callbacks.OnOutput != nil {
			manager.callbacks.OnOutput(TerminalOutputEvent{
				ProjectID: session.projectID,
				Data:      string(buffer[:n]),
			})
		}
		if err != nil {
			return
		}
	}
}

func (manager *ShellSessionManager) waitForExit(session *ShellSession) {
	_ = session.process.Wait()

	manager.mu.Lock()
	if current, ok := manager.sessions[session.projectID]; ok && current == session {
		current.state = ShellStateExited
	}
	manager.mu.Unlock()

	if manager.callbacks.OnStatus != nil {
		manager.callbacks.OnStatus(ShellStatus{ProjectID: session.projectID, State: ShellStateExited})
	}
}

func DefaultShellPath() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	if fileExists("/bin/bash") {
		return "/bin/bash"
	}
	return "/bin/sh"
}

func NewPtyProcess(request ShellStartRequest) (PtyProcess, error) {
	cmd := exec.Command(request.ShellPath)
	cmd.Dir = request.WorkingDir
	cmd.Env = request.Env
	if len(cmd.Env) == 0 {
		cmd.Env = EmbeddedTerminalEnv(os.Environ())
	}

	file, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(request.Size.Cols),
		Rows: uint16(request.Size.Rows),
	})
	if err != nil {
		return nil, err
	}
	return &realPtyProcess{file: file, cmd: cmd}, nil
}

type realPtyProcess struct {
	file *os.File
	cmd  *exec.Cmd
}

func (process *realPtyProcess) Read(data []byte) (int, error) {
	return process.file.Read(data)
}

func (process *realPtyProcess) Write(data []byte) (int, error) {
	return process.file.Write(data)
}

func (process *realPtyProcess) Resize(size TerminalSize) error {
	return pty.Setsize(process.file, &pty.Winsize{
		Cols: uint16(size.Cols),
		Rows: uint16(size.Rows),
	})
}

func (process *realPtyProcess) Wait() error {
	return process.cmd.Wait()
}

func (process *realPtyProcess) Close() error {
	return process.file.Close()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func EmbeddedTerminalEnv(base []string) []string {
	return envWithOverrides(base, map[string]string{
		"TERM":      "xterm-256color",
		"COLORTERM": "truecolor",
	})
}

func envWithOverrides(base []string, overrides map[string]string) []string {
	result := make([]string, 0, len(base)+len(overrides))
	seen := map[string]bool{}
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			result = append(result, entry)
			continue
		}
		if value, ok := overrides[key]; ok {
			result = append(result, key+"="+value)
			seen[key] = true
			continue
		}
		result = append(result, entry)
	}

	for _, key := range []string{"TERM", "COLORTERM"} {
		if !seen[key] {
			result = append(result, key+"="+overrides[key])
		}
	}
	return result
}
