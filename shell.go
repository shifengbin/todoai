package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/google/uuid"
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
	TerminalID string
	ProjectID  string
	WorkingDir string
	ShellPath  string
	ShellArgs  []string
	ShellName  string
	Size       TerminalSize
	Env        []string
}

type ShellStatus struct {
	ProjectID  string `json:"projectId"`
	TerminalID string `json:"terminalId"`
	State      string `json:"state"`
}

type TerminalOutputEvent struct {
	ProjectID  string `json:"projectId"`
	TerminalID string `json:"terminalId"`
	Data       string `json:"data"`
}

type ProjectTerminal struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	ShellName      string `json:"shellName"`
	CurrentCommand string `json:"currentCommand"`
	State          string `json:"state"`
	CreatedAt      string `json:"createdAt"`
	LastSelectedAt string `json:"lastSelectedAt"`

	projectPath string
	shellPath   string
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
	newID             func() string
	now               func() time.Time
	sessions          map[string]*ShellSession
	terminals         map[string]*ProjectTerminal
	activeByProject   map[string]string
}

type ShellSession struct {
	terminalID  string
	projectID   string
	process     PtyProcess
	size        TerminalSize
	state       string
	cleanup     func()
	cleanupOnce sync.Once
}

func NewShellSessionManager(starter ShellStarter, callbacks ShellSessionCallbacks, opts ...ShellSessionManagerOption) *ShellSessionManager {
	manager := &ShellSessionManager{
		starter:           starter,
		callbacks:         callbacks,
		shellPathResolver: DefaultShellPath,
		newID:             uuid.NewString,
		now:               time.Now,
		sessions:          map[string]*ShellSession{},
		terminals:         map[string]*ProjectTerminal{},
		activeByProject:   map[string]string{},
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

func WithShellTerminalIDGenerator(newID func() string) ShellSessionManagerOption {
	return func(manager *ShellSessionManager) {
		manager.newID = newID
	}
}

func WithShellClock(now func() time.Time) ShellSessionManagerOption {
	return func(manager *ShellSessionManager) {
		manager.now = now
	}
}

func (manager *ShellSessionManager) EnsureSession(project Project, size TerminalSize) (ShellStatus, error) {
	terminal, err := manager.EnsureProjectTerminal(project, size)
	if err != nil {
		return ShellStatus{}, err
	}
	return ShellStatus{ProjectID: terminal.ProjectID, TerminalID: terminal.ID, State: terminal.State}, nil
}

func (manager *ShellSessionManager) RegisterTerminal(project Project) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.registerTerminalLocked(project), nil
}

func (manager *ShellSessionManager) EnsureProjectTerminal(project Project, size TerminalSize) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	manager.mu.Lock()
	if terminalID := manager.activeByProject[project.ID]; terminalID != "" {
		if terminal, ok := manager.terminals[terminalID]; ok {
			manager.touchTerminalLocked(terminal)
			result := *terminal
			manager.mu.Unlock()
			return result, nil
		}
	}
	for _, terminal := range manager.terminals {
		if terminal.ProjectID == project.ID {
			manager.touchTerminalLocked(terminal)
			result := *terminal
			manager.mu.Unlock()
			return result, nil
		}
	}
	terminal := manager.registerTerminalLocked(project)
	manager.mu.Unlock()

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		return ProjectTerminal{}, err
	}
	return manager.Terminal(terminal.ID)
}

func (manager *ShellSessionManager) CreateTerminal(project Project, size TerminalSize) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	manager.mu.Lock()
	terminal := manager.registerTerminalLocked(project)
	manager.mu.Unlock()

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		manager.mu.Lock()
		delete(manager.terminals, terminal.ID)
		delete(manager.activeByProject, project.ID)
		manager.mu.Unlock()
		return ProjectTerminal{}, err
	}
	return manager.Terminal(terminal.ID)
}

func (manager *ShellSessionManager) StartTerminal(terminalID string, size TerminalSize) (ShellStatus, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminal, ok := manager.terminals[terminalID]
	if !ok {
		return ShellStatus{}, errors.New("terminal not found")
	}
	if session, ok := manager.sessions[terminalID]; ok && session.state == ShellStateRunning {
		status := ShellStatus{ProjectID: terminal.ProjectID, TerminalID: terminal.ID, State: session.state}
		terminal.State = session.state
		manager.touchTerminalLocked(terminal)
		return status, nil
	}

	launch, err := IntegratedShellLaunch(terminal.shellPath, os.Environ())
	if err != nil {
		terminal.State = ShellStateExited
		return ShellStatus{}, err
	}

	request := ShellStartRequest{
		TerminalID: terminal.ID,
		ProjectID:  terminal.ProjectID,
		WorkingDir: terminal.projectPath,
		ShellPath:  launch.Path,
		ShellArgs:  launch.Args,
		ShellName:  launch.ShellName,
		Size:       size,
		Env:        launch.Env,
	}
	process, err := manager.starter(request)
	if err != nil {
		launch.Cleanup()
		terminal.State = ShellStateExited
		return ShellStatus{}, err
	}

	session := &ShellSession{
		terminalID: terminal.ID,
		projectID:  terminal.ProjectID,
		process:    process,
		size:       size,
		state:      ShellStateRunning,
		cleanup:    launch.Cleanup,
	}

	manager.sessions[terminal.ID] = session
	terminal.State = ShellStateRunning
	manager.touchTerminalLocked(terminal)
	go manager.readOutput(session)
	go manager.waitForExit(session)

	return ShellStatus{ProjectID: terminal.ProjectID, TerminalID: terminal.ID, State: ShellStateRunning}, nil
}

func (manager *ShellSessionManager) SelectTerminal(terminalID string) (ProjectTerminal, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminal, ok := manager.terminals[terminalID]
	if !ok {
		return ProjectTerminal{}, errors.New("terminal not found")
	}
	manager.touchTerminalLocked(terminal)
	return *terminal, nil
}

func (manager *ShellSessionManager) DeleteTerminal(terminalID string) error {
	manager.mu.Lock()
	terminal, ok := manager.terminals[terminalID]
	if !ok {
		manager.mu.Unlock()
		return errors.New("terminal not found")
	}

	projectID := terminal.ProjectID
	session, hasSession := manager.sessions[terminalID]
	shouldClose := hasSession && session.state == ShellStateRunning
	delete(manager.terminals, terminalID)
	delete(manager.sessions, terminalID)
	if manager.activeByProject[projectID] == terminalID {
		if nextTerminalID := manager.mostRecentlySelectedTerminalIDLocked(projectID); nextTerminalID != "" {
			manager.activeByProject[projectID] = nextTerminalID
		} else {
			delete(manager.activeByProject, projectID)
		}
	}
	manager.mu.Unlock()

	if hasSession {
		if shouldClose {
			_ = session.process.Close()
		}
		session.cleanupSession()
	}
	return nil
}

func (manager *ShellSessionManager) DeleteProjectTerminals(projectID string) {
	manager.mu.Lock()
	sessions := []struct {
		session     *ShellSession
		shouldClose bool
	}{}
	for terminalID, terminal := range manager.terminals {
		if terminal.ProjectID != projectID {
			continue
		}
		if session, ok := manager.sessions[terminalID]; ok {
			sessions = append(sessions, struct {
				session     *ShellSession
				shouldClose bool
			}{
				session:     session,
				shouldClose: session.state == ShellStateRunning,
			})
			delete(manager.sessions, terminalID)
		}
		delete(manager.terminals, terminalID)
	}
	delete(manager.activeByProject, projectID)
	manager.mu.Unlock()

	for _, item := range sessions {
		if item.shouldClose {
			_ = item.session.process.Close()
		}
		item.session.cleanupSession()
	}
}

func (manager *ShellSessionManager) Terminal(terminalID string) (ProjectTerminal, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminal, ok := manager.terminals[terminalID]
	if !ok {
		return ProjectTerminal{}, errors.New("terminal not found")
	}
	return *terminal, nil
}

func (manager *ShellSessionManager) Terminals() []ProjectTerminal {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminals := make([]ProjectTerminal, 0, len(manager.terminals))
	for _, terminal := range manager.terminals {
		terminals = append(terminals, *terminal)
	}
	sort.Slice(terminals, func(left, right int) bool {
		if terminals[left].CreatedAt == terminals[right].CreatedAt {
			return terminals[left].ID < terminals[right].ID
		}
		return terminals[left].CreatedAt < terminals[right].CreatedAt
	})
	return terminals
}

func (manager *ShellSessionManager) ActiveTerminalID(projectID string) string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.activeByProject[projectID]
}

func (manager *ShellSessionManager) registerTerminalLocked(project Project) ProjectTerminal {
	shellPath := manager.shellPathResolver()
	now := manager.now().UTC().Format(time.RFC3339)
	terminal := &ProjectTerminal{
		ID:             manager.newID(),
		ProjectID:      project.ID,
		ShellName:      shellNameFromPath(shellPath),
		State:          ShellStateExited,
		CreatedAt:      now,
		LastSelectedAt: now,
		projectPath:    project.Path,
		shellPath:      shellPath,
	}
	manager.terminals[terminal.ID] = terminal
	manager.activeByProject[project.ID] = terminal.ID
	return *terminal
}

func (manager *ShellSessionManager) touchTerminalLocked(terminal *ProjectTerminal) {
	terminal.LastSelectedAt = manager.now().UTC().Format(time.RFC3339)
	manager.activeByProject[terminal.ProjectID] = terminal.ID
}

func (manager *ShellSessionManager) mostRecentlySelectedTerminalIDLocked(projectID string) string {
	selectedTerminalID := ""
	selectedAt := ""
	for _, terminal := range manager.terminals {
		if terminal.ProjectID != projectID {
			continue
		}
		if selectedTerminalID == "" || terminal.LastSelectedAt > selectedAt ||
			(terminal.LastSelectedAt == selectedAt && terminal.ID < selectedTerminalID) {
			selectedTerminalID = terminal.ID
			selectedAt = terminal.LastSelectedAt
		}
	}
	return selectedTerminalID
}

func (manager *ShellSessionManager) WriteInput(terminalID string, data string) error {
	session, err := manager.runningSession(terminalID)
	if err != nil {
		return err
	}
	_, err = session.process.Write([]byte(data))
	return err
}

func (manager *ShellSessionManager) Resize(terminalID string, size TerminalSize) error {
	session, err := manager.runningSession(terminalID)
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

func (manager *ShellSessionManager) Status(terminalID string) ShellStatus {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if session, ok := manager.sessions[terminalID]; ok {
		return ShellStatus{ProjectID: session.projectID, TerminalID: terminalID, State: session.state}
	}
	if terminal, ok := manager.terminals[terminalID]; ok {
		return ShellStatus{ProjectID: terminal.ProjectID, TerminalID: terminalID, State: terminal.State}
	}
	return ShellStatus{TerminalID: terminalID, State: ShellStateExited}
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
		session.cleanupSession()
	}
}

func (manager *ShellSessionManager) runningSession(terminalID string) (*ShellSession, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	session, ok := manager.sessions[terminalID]
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
				ProjectID:  session.projectID,
				TerminalID: session.terminalID,
				Data:       string(buffer[:n]),
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
	if current, ok := manager.sessions[session.terminalID]; ok && current == session {
		current.state = ShellStateExited
		if terminal, ok := manager.terminals[session.terminalID]; ok {
			terminal.State = ShellStateExited
		}
	}
	manager.mu.Unlock()
	session.cleanupSession()

	if manager.callbacks.OnStatus != nil {
		manager.callbacks.OnStatus(ShellStatus{ProjectID: session.projectID, TerminalID: session.terminalID, State: ShellStateExited})
	}
}

func (session *ShellSession) cleanupSession() {
	if session.cleanup != nil {
		session.cleanupOnce.Do(session.cleanup)
	}
}

type ShellLaunch struct {
	Path      string
	Args      []string
	Env       []string
	ShellName string
	Cleanup   func()
}

func IntegratedShellLaunch(shellPath string, baseEnv []string) (ShellLaunch, error) {
	shellName := shellNameFromPath(shellPath)
	env := EmbeddedTerminalEnv(baseEnv)
	launch := ShellLaunch{
		Path:      shellPath,
		Env:       env,
		ShellName: shellName,
		Cleanup:   func() {},
	}

	switch shellName {
	case "zsh":
		return zshIntegratedLaunch(launch)
	case "bash":
		return bashIntegratedLaunch(launch)
	default:
		return launch, nil
	}
}

func zshIntegratedLaunch(launch ShellLaunch) (ShellLaunch, error) {
	dir, err := os.MkdirTemp("", "tui-helper-zsh-*")
	if err != nil {
		return ShellLaunch{}, err
	}
	originalZDOTDIR := envValueFromList(launch.Env, "ZDOTDIR")
	if originalZDOTDIR == "" {
		originalZDOTDIR = envValueFromList(launch.Env, "HOME")
	}
	if err := os.WriteFile(filepath.Join(dir, ".zshrc"), []byte(zshIntegrationScript()), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return ShellLaunch{}, err
	}
	launch.Args = []string{"-i"}
	launch.Env = envWithOverrides(launch.Env, map[string]string{
		"ZDOTDIR":                     dir,
		"TUI_HELPER_ORIGINAL_ZDOTDIR": originalZDOTDIR,
	})
	launch.Cleanup = func() {
		_ = os.RemoveAll(dir)
	}
	return launch, nil
}

func bashIntegratedLaunch(launch ShellLaunch) (ShellLaunch, error) {
	file, err := os.CreateTemp("", "tui-helper-bash-*.bashrc")
	if err != nil {
		return ShellLaunch{}, err
	}
	path := file.Name()
	if _, err := file.WriteString(bashIntegrationScript()); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return ShellLaunch{}, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return ShellLaunch{}, err
	}
	launch.Args = []string{"--rcfile", path, "-i"}
	launch.Cleanup = func() {
		_ = os.Remove(path)
	}
	return launch, nil
}

func zshIntegrationScript() string {
	return `
if [ -n "$TUI_HELPER_ORIGINAL_ZDOTDIR" ] && [ -f "$TUI_HELPER_ORIGINAL_ZDOTDIR/.zshrc" ]; then
  source "$TUI_HELPER_ORIGINAL_ZDOTDIR/.zshrc"
fi

autoload -Uz add-zsh-hook
__tui_helper_emit_command_start() {
  printf '\033]777;tui-helper;command-start;%s\a' "$(printf '%s' "$1" | base64 | tr -d '\n')"
}
__tui_helper_emit_command_end() {
  printf '\033]777;tui-helper;command-end\a'
}
__tui_helper_preexec() {
  __tui_helper_emit_command_start "$1"
}
__tui_helper_precmd() {
  __tui_helper_emit_command_end
}
add-zsh-hook preexec __tui_helper_preexec
add-zsh-hook precmd __tui_helper_precmd
`
}

func bashIntegrationScript() string {
	return `
if [ -f "$HOME/.bashrc" ]; then
  . "$HOME/.bashrc"
fi

__tui_helper_command_started=0
__tui_helper_in_prompt=0
__tui_helper_original_prompt_command="$PROMPT_COMMAND"
__tui_helper_emit_command_start() {
  printf '\033]777;tui-helper;command-start;%s\a' "$(printf '%s' "$1" | base64 | tr -d '\n')"
}
__tui_helper_emit_command_end() {
  printf '\033]777;tui-helper;command-end\a'
}
__tui_helper_debug_trap() {
  if [ "$__tui_helper_in_prompt" = "1" ]; then
    return
  fi
  local command="$BASH_COMMAND"
  case "$command" in
    __tui_helper_*|trap\ *|PROMPT_COMMAND=*) return ;;
  esac
  __tui_helper_emit_command_start "$command"
  __tui_helper_command_started=1
}
__tui_helper_prompt_command() {
  __tui_helper_in_prompt=1
  if [ "$__tui_helper_command_started" = "1" ]; then
    __tui_helper_emit_command_end
    __tui_helper_command_started=0
  fi
  if [ -n "$__tui_helper_original_prompt_command" ]; then
    eval "$__tui_helper_original_prompt_command"
  fi
  __tui_helper_in_prompt=0
}
trap '__tui_helper_debug_trap' DEBUG
PROMPT_COMMAND="__tui_helper_prompt_command"
`
}

func shellNameFromPath(shellPath string) string {
	name := filepath.Base(shellPath)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "shell"
	}
	return name
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
	cmd := exec.Command(request.ShellPath, request.ShellArgs...)
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

	for key, value := range overrides {
		if !seen[key] {
			result = append(result, key+"="+value)
		}
	}
	return result
}

func envValueFromList(env []string, key string) string {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}
