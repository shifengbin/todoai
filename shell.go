package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	ShellStateRunning     = "running"
	ShellStateExited      = "exited"
	ShellStateUnsupported = "unsupported"

	WorkspaceTerminalContextID = "__workspace__"
)

var ErrEmbeddedShellUnsupported = errors.New("embedded terminal is not supported on this platform")

type TerminalSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type ShellStartRequest struct {
	TerminalID        string
	ProjectID         string
	TodoID            string
	TodoProjectID     string
	WorkspaceTerminal bool
	WorkingDir        string
	ShellPath         string
	ShellArgs         []string
	ShellName         string
	Size              TerminalSize
	Env               []string
}

type ShellStatus struct {
	ProjectID         string `json:"projectId"`
	TodoID            string `json:"todoId,omitempty"`
	TodoProjectID     string `json:"todoProjectId,omitempty"`
	WorkspaceTerminal bool   `json:"workspaceTerminal,omitempty"`
	TerminalID        string `json:"terminalId"`
	State             string `json:"state"`
}

type TerminalOutputEvent struct {
	ProjectID         string `json:"projectId"`
	TodoID            string `json:"todoId,omitempty"`
	TodoProjectID     string `json:"todoProjectId,omitempty"`
	WorkspaceTerminal bool   `json:"workspaceTerminal,omitempty"`
	TerminalID        string `json:"terminalId"`
	Data              string `json:"data"`
}

type TerminalAgentStatusEvent struct {
	ProjectID         string `json:"projectId"`
	TodoID            string `json:"todoId,omitempty"`
	TodoProjectID     string `json:"todoProjectId,omitempty"`
	WorkspaceTerminal bool   `json:"workspaceTerminal,omitempty"`
	TerminalID        string `json:"terminalId"`
	Phase             string `json:"phase"`
	Source            string `json:"source"`
	Confidence        string `json:"confidence"`
	Reason            string `json:"reason"`
	Label             string `json:"label,omitempty"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type ProjectTerminal struct {
	ID                string `json:"id"`
	ProjectID         string `json:"projectId"`
	TodoID            string `json:"todoId,omitempty"`
	TodoProjectID     string `json:"todoProjectId,omitempty"`
	WorkspaceTerminal bool   `json:"workspaceTerminal,omitempty"`
	ShellName         string `json:"shellName"`
	CurrentCommand    string `json:"currentCommand"`
	State             string `json:"state"`
	CreatedAt         string `json:"createdAt"`
	LastSelectedAt    string `json:"lastSelectedAt"`
	Output            string `json:"output,omitempty"`

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
	OnOutput       func(event TerminalOutputEvent)
	OnStatus       func(status ShellStatus)
	OnCommandState func(event TerminalCommandStateEvent)
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
	activeByContext   map[string]string
	history           *TerminalHistoryStore
	claudeStatusDir   string
}

type ShellSession struct {
	terminalID        string
	projectID         string
	todoID            string
	todoProjectID     string
	workspaceTerminal bool
	process           PtyProcess
	size              TerminalSize
	state             string
	cleanup           func()
	cleanupOnce       sync.Once
	outputFilter      *commandStateOutputFilter
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
		activeByContext:   map[string]string{},
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

// WithShellClaudeStatusDir sets the directory injected into every spawned
// terminal as TODOAI_STATUS_DIR, so the `todoai claude-hook` subcommand (run by
// Claude Code inside that terminal) writes .status files exactly where
// ClaudeStatusWatcher reads them. Empty disables injection.
func WithShellClaudeStatusDir(dir string) ShellSessionManagerOption {
	return func(manager *ShellSessionManager) {
		manager.claudeStatusDir = dir
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

// WithTerminalHistoryStore sets the terminal history store for persisting
// terminal output and metadata across application restarts.
func WithTerminalHistoryStore(history *TerminalHistoryStore) ShellSessionManagerOption {
	return func(manager *ShellSessionManager) {
		manager.history = history
	}
}

func (manager *ShellSessionManager) EnsureSession(project Project, size TerminalSize) (ShellStatus, error) {
	terminal, err := manager.EnsureProjectTerminal(project, size)
	if err != nil {
		return ShellStatus{}, err
	}
	return shellStatusFromTerminal(terminal), nil
}

func (manager *ShellSessionManager) RegisterTerminal(project Project) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.registerTerminalLocked(TodoProject{}, project), nil
}

func (manager *ShellSessionManager) RegisterTodoProjectTerminal(todoProject TodoProject, project Project) (ProjectTerminal, error) {
	if err := validateTodoProjectTerminalContext(todoProject, project); err != nil {
		return ProjectTerminal{}, err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.registerTerminalLocked(todoProject, project), nil
}

func (manager *ShellSessionManager) EnsureProjectTerminal(project Project, size TerminalSize) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	return manager.ensureTerminal(TodoProject{}, project, size)
}

func (manager *ShellSessionManager) EnsureTodoProjectTerminal(todoProject TodoProject, project Project, size TerminalSize) (ProjectTerminal, error) {
	if err := validateTodoProjectTerminalContext(todoProject, project); err != nil {
		return ProjectTerminal{}, err
	}
	return manager.ensureTerminal(todoProject, project, size)
}

func (manager *ShellSessionManager) ensureTerminal(todoProject TodoProject, project Project, size TerminalSize) (ProjectTerminal, error) {
	manager.mu.Lock()
	contextKey := terminalContextKey(todoProject.ID, project.ID)
	if terminalID := manager.activeByContext[contextKey]; terminalID != "" {
		if terminal, ok := manager.terminals[terminalID]; ok {
			manager.touchTerminalLocked(terminal)
			result := *terminal
			manager.mu.Unlock()
			return result, nil
		}
	}
	for _, terminal := range manager.terminals {
		if terminal.ProjectID == project.ID && terminalContextKey(terminal.TodoProjectID, terminal.ProjectID) == contextKey {
			manager.touchTerminalLocked(terminal)
			result := *terminal
			manager.mu.Unlock()
			return result, nil
		}
	}
	terminal := manager.registerTerminalLocked(todoProject, project)
	manager.mu.Unlock()

	// Persist the terminal metadata BEFORE starting the shell so that
	// readOutput's appendOutputToHistory has a record to update.
	manager.saveTerminalToHistory(terminal)

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		return ProjectTerminal{}, err
	}
	result, err := manager.Terminal(terminal.ID)
	if err != nil {
		return ProjectTerminal{}, err
	}
	manager.saveTerminalToHistory(result)
	return result, nil
}

func (manager *ShellSessionManager) CreateTerminal(project Project, size TerminalSize) (ProjectTerminal, error) {
	if !project.Available {
		return ProjectTerminal{}, errors.New("project path is unavailable")
	}

	return manager.createTerminal(TodoProject{}, project, size)
}

func (manager *ShellSessionManager) CreateTodoProjectTerminal(todoProject TodoProject, project Project, size TerminalSize) (ProjectTerminal, error) {
	if err := validateTodoProjectTerminalContext(todoProject, project); err != nil {
		return ProjectTerminal{}, err
	}
	return manager.createTerminal(todoProject, project, size)
}

// CreateTaskTerminal starts a task-level terminal rooted at a TODO's task
// workspace directory. The directory must exist; callers resolve it (and gate it
// on an in-progress TODO) before invoking this.
func (manager *ShellSessionManager) CreateTaskTerminal(todoID, workingDir string, size TerminalSize) (ProjectTerminal, error) {
	if strings.TrimSpace(todoID) == "" {
		return ProjectTerminal{}, errors.New("todo id is required")
	}
	absoluteWorkingDir, err := normalizeProjectPath(workingDir)
	if err != nil {
		return ProjectTerminal{}, err
	}
	if !directoryAvailable(absoluteWorkingDir) {
		return ProjectTerminal{}, errors.New("task workspace directory is unavailable")
	}

	manager.mu.Lock()
	terminal := manager.registerTaskTerminalLocked(todoID, absoluteWorkingDir)
	manager.mu.Unlock()

	manager.saveTerminalToHistory(terminal)

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		manager.mu.Lock()
		delete(manager.terminals, terminal.ID)
		delete(manager.activeByContext, terminalContextKeyForTerminal(terminal))
		manager.mu.Unlock()
		manager.deleteTerminalFromHistory(terminal.ID)
		return ProjectTerminal{}, err
	}
	result, err := manager.Terminal(terminal.ID)
	if err != nil {
		return ProjectTerminal{}, err
	}
	manager.saveTerminalToHistory(result)
	return result, nil
}

func (manager *ShellSessionManager) CreateWorkspaceTerminal(workspacePath string, size TerminalSize) (ProjectTerminal, error) {
	absoluteWorkspacePath, err := normalizeProjectPath(workspacePath)
	if err != nil {
		return ProjectTerminal{}, err
	}
	if !directoryAvailable(absoluteWorkspacePath) {
		return ProjectTerminal{}, errors.New("workspace path is unavailable")
	}

	manager.mu.Lock()
	terminal := manager.registerWorkspaceTerminalLocked(absoluteWorkspacePath)
	manager.mu.Unlock()

	manager.saveTerminalToHistory(terminal)

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		manager.mu.Lock()
		delete(manager.terminals, terminal.ID)
		delete(manager.activeByContext, WorkspaceTerminalContextID)
		manager.mu.Unlock()
		manager.deleteTerminalFromHistory(terminal.ID)
		return ProjectTerminal{}, err
	}
	result, err := manager.Terminal(terminal.ID)
	if err != nil {
		return ProjectTerminal{}, err
	}
	manager.saveTerminalToHistory(result)
	return result, nil
}

func (manager *ShellSessionManager) createTerminal(todoProject TodoProject, project Project, size TerminalSize) (ProjectTerminal, error) {
	manager.mu.Lock()
	terminal := manager.registerTerminalLocked(todoProject, project)
	manager.mu.Unlock()

	// Persist the terminal metadata BEFORE starting the shell so that
	// readOutput's appendOutputToHistory has a record to update.
	manager.saveTerminalToHistory(terminal)

	if _, err := manager.StartTerminal(terminal.ID, size); err != nil {
		manager.mu.Lock()
		delete(manager.terminals, terminal.ID)
		delete(manager.activeByContext, terminalContextKeyForTerminal(terminal))
		manager.mu.Unlock()
		manager.deleteTerminalFromHistory(terminal.ID)
		return ProjectTerminal{}, err
	}
	result, err := manager.Terminal(terminal.ID)
	if err != nil {
		return ProjectTerminal{}, err
	}
	manager.saveTerminalToHistory(result)
	return result, nil
}

func (manager *ShellSessionManager) StartTerminal(terminalID string, size TerminalSize) (ShellStatus, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminal, ok := manager.terminals[terminalID]
	if !ok {
		return ShellStatus{}, errors.New("terminal not found")
	}
	if session, ok := manager.sessions[terminalID]; ok && session.state == ShellStateRunning {
		status := ShellStatus{
			ProjectID:         terminal.ProjectID,
			TodoID:            terminal.TodoID,
			TodoProjectID:     terminal.TodoProjectID,
			WorkspaceTerminal: terminal.WorkspaceTerminal,
			TerminalID:        terminal.ID,
			State:             session.state,
		}
		terminal.State = session.state
		manager.touchTerminalLocked(terminal)
		return status, nil
	}

	launch, err := IntegratedShellLaunch(terminal.shellPath, os.Environ())
	if err != nil {
		terminal.State = ShellStateExited
		return ShellStatus{}, err
	}

	terminalEnv := terminalIdentityEnv(launch.Env, *terminal)
	terminalEnv = envWithoutKeys(terminalEnv, "TODOAI_STATUS_DIR")
	if manager.claudeStatusDir != "" {
		terminalEnv = envWithOverrides(terminalEnv, map[string]string{"TODOAI_STATUS_DIR": manager.claudeStatusDir})
	}
	request := ShellStartRequest{
		TerminalID:        terminal.ID,
		ProjectID:         terminal.ProjectID,
		TodoID:            terminal.TodoID,
		TodoProjectID:     terminal.TodoProjectID,
		WorkspaceTerminal: terminal.WorkspaceTerminal,
		WorkingDir:        terminal.projectPath,
		ShellPath:         launch.Path,
		ShellArgs:         launch.Args,
		ShellName:         launch.ShellName,
		Size:              size,
		Env:               terminalEnv,
	}
	process, err := manager.starter(request)
	if err != nil {
		launch.Cleanup()
		if errors.Is(err, ErrEmbeddedShellUnsupported) {
			terminal.State = ShellStateUnsupported
			manager.touchTerminalLocked(terminal)
			return shellStatusFromTerminal(*terminal), nil
		}
		terminal.State = ShellStateExited
		return ShellStatus{}, err
	}

	session := &ShellSession{
		terminalID:        terminal.ID,
		projectID:         terminal.ProjectID,
		todoID:            terminal.TodoID,
		todoProjectID:     terminal.TodoProjectID,
		workspaceTerminal: terminal.WorkspaceTerminal,
		process:           process,
		size:              size,
		state:             ShellStateRunning,
		cleanup:           launch.Cleanup,
		outputFilter:      newCommandStateOutputFilter(),
	}

	manager.sessions[terminal.ID] = session
	terminal.State = ShellStateRunning
	manager.touchTerminalLocked(terminal)
	go manager.readOutput(session)
	go manager.waitForExit(session)

	return ShellStatus{
		ProjectID:         terminal.ProjectID,
		TodoID:            terminal.TodoID,
		TodoProjectID:     terminal.TodoProjectID,
		WorkspaceTerminal: terminal.WorkspaceTerminal,
		TerminalID:        terminal.ID,
		State:             ShellStateRunning,
	}, nil
}

func (manager *ShellSessionManager) SelectTerminal(terminalID string) (ProjectTerminal, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	terminal, ok := manager.terminals[terminalID]
	if !ok {
		return ProjectTerminal{}, errors.New("terminal not found")
	}
	manager.touchTerminalLocked(terminal)
	result := *terminal
	manager.saveTerminalToHistory(result)
	return result, nil
}

func (manager *ShellSessionManager) DeleteTerminal(terminalID string) error {
	manager.mu.Lock()
	terminal, ok := manager.terminals[terminalID]
	if !ok {
		manager.mu.Unlock()
		return errors.New("terminal not found")
	}

	contextKey := terminalContextKeyForTerminal(*terminal)
	session, hasSession := manager.sessions[terminalID]
	shouldClose := hasSession && session.state == ShellStateRunning
	delete(manager.terminals, terminalID)
	delete(manager.sessions, terminalID)
	if manager.activeByContext[contextKey] == terminalID {
		if nextTerminalID := manager.mostRecentlySelectedTerminalIDLocked(contextKey); nextTerminalID != "" {
			manager.activeByContext[contextKey] = nextTerminalID
		} else {
			delete(manager.activeByContext, contextKey)
		}
	}
	manager.mu.Unlock()

	manager.deleteTerminalFromHistory(terminalID)

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
		delete(manager.activeByContext, terminalContextKeyForTerminal(*terminal))
	}
	manager.mu.Unlock()

	manager.deleteProjectFromHistory(projectID)

	for _, item := range sessions {
		if item.shouldClose {
			_ = item.session.process.Close()
		}
		item.session.cleanupSession()
	}
}

func (manager *ShellSessionManager) DeleteTodoTerminals(todoID string) {
	manager.mu.Lock()
	sessions := []struct {
		session     *ShellSession
		shouldClose bool
	}{}
	for terminalID, terminal := range manager.terminals {
		if terminal.TodoID != todoID {
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
		delete(manager.activeByContext, terminalContextKeyForTerminal(*terminal))
	}
	manager.mu.Unlock()

	manager.deleteTodoFromHistory(todoID)

	for _, item := range sessions {
		if item.shouldClose {
			_ = item.session.process.Close()
		}
		item.session.cleanupSession()
	}
}

func (manager *ShellSessionManager) DeleteTodoProjectTerminals(todoProjectID string) {
	manager.mu.Lock()
	sessions := []struct {
		session     *ShellSession
		shouldClose bool
	}{}
	for terminalID, terminal := range manager.terminals {
		if terminal.TodoProjectID != todoProjectID {
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
		delete(manager.activeByContext, terminalContextKeyForTerminal(*terminal))
	}
	manager.mu.Unlock()

	manager.deleteTodoProjectFromHistory(todoProjectID)

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
		t := *terminal
		// Populate output from history for restored (non-running) terminals.
		if t.State != ShellStateRunning && t.Output == "" && manager.history != nil {
			history, err := manager.history.Load()
			if err == nil {
				for _, record := range history.Records {
					if record.TerminalID == t.ID {
						t.Output = record.Output
						break
					}
				}
			}
		}
		terminals = append(terminals, t)
	}
	sort.Slice(terminals, func(left, right int) bool {
		if terminals[left].CreatedAt == terminals[right].CreatedAt {
			return terminals[left].ID < terminals[right].ID
		}
		return terminals[left].CreatedAt < terminals[right].CreatedAt
	})
	return terminals
}

func (manager *ShellSessionManager) ActiveTerminalID(contextID string) string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.activeByContext[contextID]
}

// RestoreTerminals loads persisted terminal records from the history store
// and registers them as non-running terminals in the manager. Valid records
// are those whose project, TODO, and TODO project references still exist
// in the provided state. Orphaned records are dropped from the history store.
// No shell processes are started.
func (manager *ShellSessionManager) RestoreTerminals(state ProjectState) []TerminalHistoryRecord {
	if manager.history == nil {
		return nil
	}

	history, err := manager.history.Load()
	if err != nil {
		return nil
	}

	// Build lookup sets for validation.
	projectIDs := map[string]bool{}
	for _, project := range state.Projects {
		projectIDs[project.ID] = true
	}
	todoIDs := map[string]bool{}
	for _, todo := range state.Todos {
		todoIDs[todo.ID] = true
	}
	todoProjectIDs := map[string]bool{}
	todoProjectProjectIDs := map[string]string{}
	todoProjectSourceProjectIDs := map[string]string{}
	for _, todoProject := range state.TodoProjects {
		todoProjectIDs[todoProject.ID] = true
		todoProjectProjectIDs[todoProject.ID] = todoProject.ProjectID
		todoProjectSourceProjectIDs[todoProject.ID] = todoProject.SourceProjectID
	}

	valid := make([]TerminalHistoryRecord, 0, len(history.Records))
	orphaned := false

	for _, record := range history.Records {
		isTaskTerminal := !record.WorkspaceTerminal && record.TodoProjectID == "" && record.ProjectID == "" && record.TodoID != ""
		if !record.WorkspaceTerminal && !isTaskTerminal && !restorableTerminalProject(record, projectIDs, todoProjectProjectIDs, todoProjectSourceProjectIDs) {
			orphaned = true
			continue
		}
		if !record.WorkspaceTerminal && record.TodoID != "" && !todoIDs[record.TodoID] {
			orphaned = true
			continue
		}
		if !record.WorkspaceTerminal && record.TodoProjectID != "" && !todoProjectIDs[record.TodoProjectID] {
			orphaned = true
			continue
		}

		// Register as a non-running terminal.
		terminal := &ProjectTerminal{
			ID:                record.TerminalID,
			ProjectID:         record.ProjectID,
			TodoID:            record.TodoID,
			TodoProjectID:     record.TodoProjectID,
			WorkspaceTerminal: record.WorkspaceTerminal,
			ShellName:         record.ShellName,
			State:             ShellStateExited,
			CreatedAt:         record.CreatedAt,
			LastSelectedAt:    record.LastSelectedAt,
			projectPath:       "",
			shellPath:         manager.shellPathResolver(),
		}
		manager.mu.Lock()
		manager.terminals[terminal.ID] = terminal
		contextKey := terminalContextKeyForTerminal(*terminal)
		if _, exists := manager.activeByContext[contextKey]; !exists {
			manager.activeByContext[contextKey] = terminal.ID
		}
		manager.mu.Unlock()

		valid = append(valid, record)
	}

	// Save cleaned history if orphaned records were removed.
	if orphaned {
		history.Records = valid
		_ = manager.history.Save(history)
	}

	return valid
}

func restorableTerminalProject(record TerminalHistoryRecord, projectIDs map[string]bool, todoProjectProjectIDs map[string]string, todoProjectSourceProjectIDs map[string]string) bool {
	if projectIDs[record.ProjectID] {
		return true
	}
	if record.TodoProjectID == "" {
		return false
	}
	if todoProjectProjectIDs[record.TodoProjectID] == record.ProjectID {
		return true
	}
	return todoProjectSourceProjectIDs[record.TodoProjectID] == record.ProjectID
}

func (manager *ShellSessionManager) registerTerminalLocked(todoProject TodoProject, project Project) ProjectTerminal {
	shellPath := manager.shellPathResolver()
	now := manager.now().UTC().Format(time.RFC3339)
	// A TODO project terminal runs inside its prepared worktree so its file
	// changes are isolated from the original project checkout.
	workingDir := project.Path
	if todoProject.WorktreePath != "" {
		workingDir = todoProject.WorktreePath
	}
	terminal := &ProjectTerminal{
		ID:             manager.newID(),
		ProjectID:      project.ID,
		TodoID:         todoProject.TodoID,
		TodoProjectID:  todoProject.ID,
		ShellName:      shellNameFromPath(shellPath),
		State:          ShellStateExited,
		CreatedAt:      now,
		LastSelectedAt: now,
		projectPath:    workingDir,
		shellPath:      shellPath,
	}
	manager.terminals[terminal.ID] = terminal
	manager.activeByContext[terminalContextKeyForTerminal(*terminal)] = terminal.ID
	return *terminal
}

// registerTaskTerminalLocked registers a task-level terminal whose working
// directory is a TODO's task workspace directory. Task terminals carry only a
// TodoID (no project or TODO project reference) and share the TODO's context.
func (manager *ShellSessionManager) registerTaskTerminalLocked(todoID, workingDir string) ProjectTerminal {
	shellPath := manager.shellPathResolver()
	now := manager.now().UTC().Format(time.RFC3339)
	terminal := &ProjectTerminal{
		ID:             manager.newID(),
		TodoID:         todoID,
		ShellName:      shellNameFromPath(shellPath),
		State:          ShellStateExited,
		CreatedAt:      now,
		LastSelectedAt: now,
		projectPath:    workingDir,
		shellPath:      shellPath,
	}
	manager.terminals[terminal.ID] = terminal
	manager.activeByContext[terminalContextKeyForTerminal(*terminal)] = terminal.ID
	return *terminal
}

func (manager *ShellSessionManager) registerWorkspaceTerminalLocked(workspacePath string) ProjectTerminal {
	shellPath := manager.shellPathResolver()
	now := manager.now().UTC().Format(time.RFC3339)
	terminal := &ProjectTerminal{
		ID:                manager.newID(),
		WorkspaceTerminal: true,
		ShellName:         shellNameFromPath(shellPath),
		State:             ShellStateExited,
		CreatedAt:         now,
		LastSelectedAt:    now,
		projectPath:       workspacePath,
		shellPath:         shellPath,
	}
	manager.terminals[terminal.ID] = terminal
	manager.activeByContext[WorkspaceTerminalContextID] = terminal.ID
	return *terminal
}

func (manager *ShellSessionManager) touchTerminalLocked(terminal *ProjectTerminal) {
	terminal.LastSelectedAt = manager.now().UTC().Format(time.RFC3339)
	manager.activeByContext[terminalContextKeyForTerminal(*terminal)] = terminal.ID
}

func (manager *ShellSessionManager) mostRecentlySelectedTerminalIDLocked(contextKey string) string {
	selectedTerminalID := ""
	selectedAt := ""
	for _, terminal := range manager.terminals {
		if terminalContextKeyForTerminal(*terminal) != contextKey {
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

// taskTerminalContextIDPrefix namespaces the active-terminal context key for
// task-level terminals so they never collide with project or workspace
// terminals.
const taskTerminalContextIDPrefix = "__task__"

func taskTerminalContextID(todoID string) string {
	return taskTerminalContextIDPrefix + todoID
}

func terminalContextKey(todoProjectID string, projectID string) string {
	if todoProjectID != "" {
		return todoProjectID
	}
	return projectID
}

func terminalContextKeyForTerminal(terminal ProjectTerminal) string {
	if terminal.WorkspaceTerminal {
		return WorkspaceTerminalContextID
	}
	if terminal.TodoProjectID != "" {
		return terminal.TodoProjectID
	}
	if terminal.TodoID != "" {
		return taskTerminalContextID(terminal.TodoID)
	}
	return terminal.ProjectID
}

func validateTodoProjectTerminalContext(todoProject TodoProject, project Project) error {
	if !project.Available {
		return errors.New("project path is unavailable")
	}
	if todoProject.ID == "" || todoProject.TodoID == "" {
		return errors.New("todo project context is required")
	}
	if todoProject.ProjectID != project.ID {
		return errors.New("todo project does not reference project")
	}
	return nil
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
		return ShellStatus{
			ProjectID:         session.projectID,
			TodoID:            session.todoID,
			TodoProjectID:     session.todoProjectID,
			WorkspaceTerminal: session.workspaceTerminal,
			TerminalID:        terminalID,
			State:             session.state,
		}
	}
	if terminal, ok := manager.terminals[terminalID]; ok {
		return shellStatusFromTerminal(*terminal)
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

func (manager *ShellSessionManager) Reset() {
	manager.mu.Lock()
	sessions := make([]*ShellSession, 0, len(manager.sessions))
	for _, session := range manager.sessions {
		sessions = append(sessions, session)
	}
	manager.sessions = map[string]*ShellSession{}
	manager.terminals = map[string]*ProjectTerminal{}
	manager.activeByContext = map[string]string{}
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
		if n > 0 {
			result := session.outputFilter.Filter(string(buffer[:n]))
			for _, event := range result.Events {
				manager.emitCommandState(session, event)
			}
			if result.Data != "" && manager.callbacks.OnOutput != nil {
				manager.callbacks.OnOutput(TerminalOutputEvent{
					ProjectID:         session.projectID,
					TodoID:            session.todoID,
					TodoProjectID:     session.todoProjectID,
					WorkspaceTerminal: session.workspaceTerminal,
					TerminalID:        session.terminalID,
					Data:              result.Data,
				})
			}
			if result.Data != "" {
				manager.appendOutputToHistory(session.terminalID, result.Data)
			}
		}
		if err != nil {
			return
		}
	}
}

func (manager *ShellSessionManager) emitCommandState(session *ShellSession, event TerminalCommandStateEvent) {
	if manager.callbacks.OnCommandState == nil {
		return
	}
	event.ProjectID = session.projectID
	event.TodoID = session.todoID
	event.TodoProjectID = session.todoProjectID
	event.WorkspaceTerminal = session.workspaceTerminal
	event.TerminalID = session.terminalID
	manager.callbacks.OnCommandState(event)
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
		manager.callbacks.OnStatus(ShellStatus{
			ProjectID:         session.projectID,
			TodoID:            session.todoID,
			TodoProjectID:     session.todoProjectID,
			WorkspaceTerminal: session.workspaceTerminal,
			TerminalID:        session.terminalID,
			State:             ShellStateExited,
		})
	}
}

func shellStatusFromTerminal(terminal ProjectTerminal) ShellStatus {
	return ShellStatus{
		ProjectID:         terminal.ProjectID,
		TodoID:            terminal.TodoID,
		TodoProjectID:     terminal.TodoProjectID,
		WorkspaceTerminal: terminal.WorkspaceTerminal,
		TerminalID:        terminal.ID,
		State:             terminal.State,
	}
}

func (session *ShellSession) cleanupSession() {
	if session.cleanup != nil {
		session.cleanupOnce.Do(session.cleanup)
	}
}

// appendOutputToHistory appends terminal output to the persisted history
// store. This is called from the PTY read loop.
func (manager *ShellSessionManager) appendOutputToHistory(terminalID string, data string) {
	if manager.history == nil {
		return
	}

	history, err := manager.history.Load()
	if err != nil {
		return
	}

	for i, record := range history.Records {
		if record.TerminalID == terminalID {
			history.Records[i].Output = AppendTerminalOutput(record.Output, data)
			_ = manager.history.Save(history)
			return
		}
	}
}

// saveTerminalToHistory persists a terminal record to the history store.
func (manager *ShellSessionManager) saveTerminalToHistory(terminal ProjectTerminal) {
	if manager.history == nil {
		return
	}

	history, err := manager.history.Load()
	if err != nil {
		return
	}

	record := TerminalHistoryRecord{
		TerminalID:        terminal.ID,
		ProjectID:         terminal.ProjectID,
		TodoID:            terminal.TodoID,
		TodoProjectID:     terminal.TodoProjectID,
		WorkspaceTerminal: terminal.WorkspaceTerminal,
		ShellName:         terminal.ShellName,
		State:             terminal.State,
		CreatedAt:         terminal.CreatedAt,
		LastSelectedAt:    terminal.LastSelectedAt,
	}

	// Preserve existing output when updating metadata.
	for _, existing := range history.Records {
		if existing.TerminalID == terminal.ID {
			record.Output = existing.Output
			break
		}
	}

	_, _ = manager.history.UpsertRecord(history, record)
}

// deleteTerminalFromHistory removes a terminal's persisted history record.
func (manager *ShellSessionManager) deleteTerminalFromHistory(terminalID string) {
	if manager.history == nil {
		return
	}
	history, err := manager.history.Load()
	if err != nil {
		return
	}
	_, _ = manager.history.DeleteRecord(history, terminalID)
}

// deleteProjectFromHistory removes all terminal history records for a project.
func (manager *ShellSessionManager) deleteProjectFromHistory(projectID string) {
	if manager.history == nil {
		return
	}
	history, err := manager.history.Load()
	if err != nil {
		return
	}
	_, _ = manager.history.DeleteRecordsByProject(history, projectID)
}

// deleteTodoFromHistory removes all terminal history records for a TODO.
func (manager *ShellSessionManager) deleteTodoFromHistory(todoID string) {
	if manager.history == nil {
		return
	}
	history, err := manager.history.Load()
	if err != nil {
		return
	}
	_, _ = manager.history.DeleteRecordsByTodo(history, todoID)
}

// deleteTodoProjectFromHistory removes all terminal history records for a
// TODO project.
func (manager *ShellSessionManager) deleteTodoProjectFromHistory(todoProjectID string) {
	if manager.history == nil {
		return
	}
	history, err := manager.history.Load()
	if err != nil {
		return
	}
	_, _ = manager.history.DeleteRecordsByTodoProject(history, todoProjectID)
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
	case "pwsh", "powershell":
		return powerShellIntegratedLaunch(launch)
	default:
		return launch, nil
	}
}

func terminalIdentityEnv(env []string, terminal ProjectTerminal) []string {
	return envWithOverrides(env, map[string]string{
		"TUI_HELPER_TERMINAL_ID":        terminal.ID,
		"TUI_HELPER_PROJECT_ID":         terminal.ProjectID,
		"TUI_HELPER_TODO_ID":            terminal.TodoID,
		"TUI_HELPER_TODO_PROJECT_ID":    terminal.TodoProjectID,
		"TUI_HELPER_WORKSPACE_TERMINAL": workspaceTerminalEnvValue(terminal.WorkspaceTerminal),
		"TUI_HELPER_TERMINAL_WORKDIR":   terminal.projectPath,
	})
}

func workspaceTerminalEnvValue(workspaceTerminal bool) string {
	if workspaceTerminal {
		return "true"
	}
	return ""
}

func zshIntegratedLaunch(launch ShellLaunch) (ShellLaunch, error) {
	dir, err := os.MkdirTemp("", "todoai-zsh-*")
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
	file, err := os.CreateTemp("", "todoai-bash-*.bashrc")
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

func powerShellIntegratedLaunch(launch ShellLaunch) (ShellLaunch, error) {
	file, err := os.CreateTemp("", "todoai-powershell-*.ps1")
	if err != nil {
		return ShellLaunch{}, err
	}
	path := file.Name()
	if _, err := file.WriteString(powerShellIntegrationScript()); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return ShellLaunch{}, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return ShellLaunch{}, err
	}
	launch.Args = []string{"-NoLogo", "-NoExit", "-ExecutionPolicy", "Bypass", "-Command", ". " + powerShellSingleQuoted(path)}
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
  printf '\033]777;todoai;command-start;%s\a' "$(printf '%s' "$1" | base64 | tr -d '\n')"
}
__tui_helper_emit_command_end() {
  printf '\033]777;todoai;command-end\a'
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
  printf '\033]777;todoai;command-start;%s\a' "$(printf '%s' "$1" | base64 | tr -d '\n')"
}
__tui_helper_emit_command_end() {
  printf '\033]777;todoai;command-end\a'
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

func powerShellIntegrationScript() string {
	return `
$script:__tui_helper_command_started = $false
$script:__tui_helper_original_prompt = $null
if (Test-Path Function:\prompt) {
  $script:__tui_helper_original_prompt = (Get-Command prompt -CommandType Function).ScriptBlock
}

function __tui_helper_write_osc {
  param([string]$Payload)
  [Console]::Out.Write("$([char]27)]$Payload$([char]7)")
}

function __tui_helper_emit_command_start {
  param([string]$Command)
  if ([string]::IsNullOrWhiteSpace($Command)) {
    return
  }
  $encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Command))
  __tui_helper_write_osc "777;todoai;command-start;$encoded"
  $script:__tui_helper_command_started = $true
}

function __tui_helper_emit_command_end {
  __tui_helper_write_osc "777;todoai;command-end"
}

if (Get-Command Set-PSReadLineOption -ErrorAction SilentlyContinue) {
  Set-PSReadLineOption -AddToHistoryHandler {
    param([string]$commandLine)
    __tui_helper_emit_command_start $commandLine
    return $true
  }
}

function global:prompt {
  if ($script:__tui_helper_command_started) {
    __tui_helper_emit_command_end
    $script:__tui_helper_command_started = $false
  }
  if ($script:__tui_helper_original_prompt) {
    & $script:__tui_helper_original_prompt
  } else {
    "PS $($executionContext.SessionState.Path.CurrentLocation)> "
  }
}
`
}

func powerShellSingleQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func shellNameFromPath(shellPath string) string {
	name := strings.TrimRight(shellPath, `\/`)
	if index := strings.LastIndexAny(name, `\/`); index >= 0 {
		name = name[index+1:]
	} else {
		name = filepath.Base(name)
	}
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "shell"
	}
	if isWindowsExecutableExtension(filepath.Ext(name)) {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return name
}

func DefaultShellPath() string {
	return defaultShellPath(NewShellDetector())
}

func defaultShellPath(detector ShellDetector) string {
	if shell, err := detector.Detect(); err == nil && shell.Path != "" {
		return shell.Path
	}
	if detector.goos == "windows" {
		if comspec := detector.getenv("COMSPEC"); comspec != "" {
			return comspec
		}
		if systemRoot := detector.windowsRoot(); systemRoot != "" {
			return windowsPathJoin(systemRoot, "System32", "cmd.exe")
		}
		return "cmd.exe"
	}
	if fileExists("/bin/bash") {
		return "/bin/bash"
	}
	return "/bin/sh"
}

func isWindowsExecutableExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".exe", ".cmd", ".bat", ".com":
		return true
	default:
		return false
	}
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

func envWithoutKeys(base []string, keys ...string) []string {
	if len(keys) == 0 {
		return append([]string{}, base...)
	}
	omit := map[string]bool{}
	for _, key := range keys {
		omit[key] = true
	}
	result := make([]string, 0, len(base))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if ok && omit[key] {
			continue
		}
		result = append(result, entry)
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
