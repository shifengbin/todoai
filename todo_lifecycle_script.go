package main

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	TodoLifecycleScriptPhaseInit     = "init"
	TodoLifecycleScriptPhaseComplete = "complete"

	TodoLifecycleScriptStatusQueued  = "queued"
	TodoLifecycleScriptStatusRunning = "running"
	TodoLifecycleScriptStatusFailed  = "failed"

	todoLifecycleScriptOutputTailLimit = 4096
)

type TodoLifecycleScriptRunRequest struct {
	TodoID          string
	Phase           string
	ScriptName      string
	Script          string
	WorkingDir      string
	ShellPath       string
	GOOS            string
	Parameters      []TodoLifecycleScriptParameter
	ParameterValues map[string]string
}

type TodoLifecycleScriptRunResult struct {
	Output   string
	ExitCode int
	Err      error
}

type TodoLifecycleScriptStatus struct {
	TodoID     string `json:"todoId"`
	Phase      string `json:"phase"`
	Status     string `json:"status"`
	ScriptName string `json:"scriptName,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
	ExitCode   int    `json:"exitCode,omitempty"`
	OutputTail string `json:"outputTail,omitempty"`
	Message    string `json:"message,omitempty"`
}

type todoLifecycleScriptRunner func(context.Context, TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult

type TodoLifecycleScriptExecutor struct {
	mu       sync.Mutex
	statuses map[string]TodoLifecycleScriptStatus
	runner   todoLifecycleScriptRunner
	now      func() time.Time
	onStatus func(TodoLifecycleScriptStatus)
}

type TodoLifecycleScriptExecutorOption func(*TodoLifecycleScriptExecutor)

func NewTodoLifecycleScriptExecutor(runner todoLifecycleScriptRunner, opts ...TodoLifecycleScriptExecutorOption) *TodoLifecycleScriptExecutor {
	if runner == nil {
		runner = runTodoLifecycleScriptCommand
	}
	executor := &TodoLifecycleScriptExecutor{
		statuses: map[string]TodoLifecycleScriptStatus{},
		runner:   runner,
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(executor)
	}
	return executor
}

func WithTodoLifecycleScriptClock(now func() time.Time) TodoLifecycleScriptExecutorOption {
	return func(executor *TodoLifecycleScriptExecutor) {
		if now != nil {
			executor.now = now
		}
	}
}

func WithTodoLifecycleScriptStatusHandler(handler func(TodoLifecycleScriptStatus)) TodoLifecycleScriptExecutorOption {
	return func(executor *TodoLifecycleScriptExecutor) {
		executor.onStatus = handler
	}
}

func (executor *TodoLifecycleScriptExecutor) Start(ctx context.Context, request TodoLifecycleScriptRunRequest) (TodoLifecycleScriptStatus, bool, error) {
	normalized, err := normalizeTodoLifecycleScriptRunRequest(request)
	if err != nil {
		return TodoLifecycleScriptStatus{}, false, err
	}
	key := todoLifecycleScriptStatusKey(normalized.TodoID, normalized.Phase)
	now := executor.now().UTC().Format(time.RFC3339)
	queued := TodoLifecycleScriptStatus{
		TodoID:     normalized.TodoID,
		Phase:      normalized.Phase,
		Status:     TodoLifecycleScriptStatusQueued,
		ScriptName: normalized.ScriptName,
		StartedAt:  now,
	}
	running := queued
	running.Status = TodoLifecycleScriptStatusRunning

	executor.mu.Lock()
	if existing, ok := executor.statuses[key]; ok && (existing.Status == TodoLifecycleScriptStatusQueued || existing.Status == TodoLifecycleScriptStatusRunning) {
		executor.mu.Unlock()
		return existing, false, nil
	}
	executor.statuses[key] = queued
	executor.mu.Unlock()
	executor.emitStatus(queued)

	executor.mu.Lock()
	executor.statuses[key] = running
	executor.mu.Unlock()
	executor.emitStatus(running)

	go executor.run(ctx, key, normalized, running)
	return running, true, nil
}

func (executor *TodoLifecycleScriptExecutor) Status(todoID string, phase string) (TodoLifecycleScriptStatus, bool) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	status, ok := executor.statuses[todoLifecycleScriptStatusKey(todoID, phase)]
	return status, ok
}

func (executor *TodoLifecycleScriptExecutor) Statuses() []TodoLifecycleScriptStatus {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	statuses := make([]TodoLifecycleScriptStatus, 0, len(executor.statuses))
	for _, status := range executor.statuses {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool {
		if statuses[i].TodoID == statuses[j].TodoID {
			return statuses[i].Phase < statuses[j].Phase
		}
		return statuses[i].TodoID < statuses[j].TodoID
	})
	return statuses
}

func (executor *TodoLifecycleScriptExecutor) Clear(todoID string, phase string) {
	key := todoLifecycleScriptStatusKey(todoID, phase)
	executor.mu.Lock()
	status, ok := executor.statuses[key]
	if ok {
		delete(executor.statuses, key)
	}
	executor.mu.Unlock()
	if ok {
		status.Status = ""
		executor.emitStatus(status)
	}
}

func (executor *TodoLifecycleScriptExecutor) ClearTodo(todoID string) {
	executor.mu.Lock()
	for key, status := range executor.statuses {
		if status.TodoID == todoID {
			delete(executor.statuses, key)
		}
	}
	executor.mu.Unlock()
}

func (executor *TodoLifecycleScriptExecutor) run(ctx context.Context, key string, request TodoLifecycleScriptRunRequest, runningStatus TodoLifecycleScriptStatus) {
	result := executor.runner(ctx, request)
	if result.Err == nil {
		executor.mu.Lock()
		delete(executor.statuses, key)
		executor.mu.Unlock()
		cleared := runningStatus
		cleared.Status = ""
		cleared.FinishedAt = executor.now().UTC().Format(time.RFC3339)
		executor.emitStatus(cleared)
		return
	}

	failed := runningStatus
	failed.Status = TodoLifecycleScriptStatusFailed
	failed.FinishedAt = executor.now().UTC().Format(time.RFC3339)
	failed.ExitCode = result.ExitCode
	failed.OutputTail = lifecycleScriptOutputTail(result.Output)
	failed.Message = result.Err.Error()
	executor.mu.Lock()
	executor.statuses[key] = failed
	executor.mu.Unlock()
	executor.emitStatus(failed)
}

func (executor *TodoLifecycleScriptExecutor) emitStatus(status TodoLifecycleScriptStatus) {
	if executor.onStatus != nil {
		executor.onStatus(status)
	}
}

func normalizeTodoLifecycleScriptRunRequest(request TodoLifecycleScriptRunRequest) (TodoLifecycleScriptRunRequest, error) {
	request.TodoID = strings.TrimSpace(request.TodoID)
	request.Phase = strings.TrimSpace(request.Phase)
	request.ScriptName = strings.TrimSpace(request.ScriptName)
	request.Script = strings.TrimSpace(request.Script)
	request.WorkingDir = strings.TrimSpace(request.WorkingDir)
	request.ShellPath = strings.TrimSpace(request.ShellPath)
	request.GOOS = strings.TrimSpace(request.GOOS)
	parameters, err := normalizeTodoLifecycleScriptParameters(request.Parameters)
	if err != nil {
		return TodoLifecycleScriptRunRequest{}, err
	}
	request.Parameters = parameters
	parameterValues, err := normalizeTodoLifecycleScriptParameterValues(parameters, request.ParameterValues)
	if err != nil {
		return TodoLifecycleScriptRunRequest{}, err
	}
	request.ParameterValues = parameterValues
	if request.TodoID == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("todo id is required")
	}
	if request.Phase != TodoLifecycleScriptPhaseInit && request.Phase != TodoLifecycleScriptPhaseComplete {
		return TodoLifecycleScriptRunRequest{}, errors.New("lifecycle script phase is invalid")
	}
	if request.Script == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("lifecycle script is required")
	}
	if request.WorkingDir == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("todo workspace directory is required")
	}
	if request.ShellPath == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("lifecycle script shell path is required")
	}
	if request.GOOS == "" {
		request.GOOS = runtime.GOOS
	}
	return request, nil
}

func runTodoLifecycleScriptCommand(ctx context.Context, request TodoLifecycleScriptRunRequest) TodoLifecycleScriptRunResult {
	cmd, err := newTodoLifecycleScriptCommand(ctx, request)
	if err != nil {
		return TodoLifecycleScriptRunResult{ExitCode: -1, Err: err}
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		return TodoLifecycleScriptRunResult{Output: string(output)}
	}
	return TodoLifecycleScriptRunResult{
		Output:   string(output),
		ExitCode: exitCodeFromCommandError(err),
		Err:      err,
	}
}

func newTodoLifecycleScriptCommand(ctx context.Context, request TodoLifecycleScriptRunRequest) (*exec.Cmd, error) {
	normalized, err := normalizeTodoLifecycleScriptCommandRequest(request)
	if err != nil {
		return nil, err
	}
	cmd := newBackgroundCommand(ctx, normalized.ShellPath, todoLifecycleScriptShellArgs(normalized.GOOS, normalized.ShellPath, normalized.Script)...)
	cmd.Dir = normalized.WorkingDir
	return cmd, nil
}

func normalizeTodoLifecycleScriptCommandRequest(request TodoLifecycleScriptRunRequest) (TodoLifecycleScriptRunRequest, error) {
	request.Script = strings.TrimSpace(request.Script)
	request.WorkingDir = strings.TrimSpace(request.WorkingDir)
	request.ShellPath = strings.TrimSpace(request.ShellPath)
	request.GOOS = strings.TrimSpace(request.GOOS)
	parameters, err := normalizeTodoLifecycleScriptParameters(request.Parameters)
	if err != nil {
		return TodoLifecycleScriptRunRequest{}, err
	}
	request.Parameters = parameters
	parameterValues, err := normalizeTodoLifecycleScriptParameterValues(parameters, request.ParameterValues)
	if err != nil {
		return TodoLifecycleScriptRunRequest{}, err
	}
	request.ParameterValues = parameterValues
	if request.Script == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("lifecycle script is required")
	}
	if request.WorkingDir == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("todo workspace directory is required")
	}
	if request.ShellPath == "" {
		return TodoLifecycleScriptRunRequest{}, errors.New("lifecycle script shell path is required")
	}
	if request.GOOS == "" {
		request.GOOS = runtime.GOOS
	}
	request.Script = renderTodoLifecycleScriptParameters(request.Script, request.Parameters, request.ParameterValues, request.GOOS, request.ShellPath)
	return request, nil
}

func renderTodoLifecycleScriptParameters(script string, parameters []TodoLifecycleScriptParameter, values map[string]string, goos string, shellPath string) string {
	if len(parameters) == 0 {
		return script
	}
	for _, parameter := range parameters {
		value := values[parameter.Name]
		script = strings.ReplaceAll(script, "{{"+parameter.Name+"}}", todoLifecycleScriptShellStringLiteral(value, goos, shellPath))
	}
	return script
}

func todoLifecycleScriptShellStringLiteral(value string, goos string, shellPath string) string {
	switch shellNameFromPath(shellPath) {
	case "pwsh", "powershell":
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	}
	if goos == "windows" {
		switch shellNameFromPath(shellPath) {
		case "cmd":
			return todoLifecycleScriptCmdStringLiteral(value)
		}
		return todoLifecycleScriptCmdStringLiteral(value)
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func todoLifecycleScriptCmdStringLiteral(value string) string {
	replacer := strings.NewReplacer(
		"^", "^^",
		`"`, `^"`,
		"%", "^%",
		"!", "^!",
		"&", "^&",
		"|", "^|",
		"<", "^<",
		">", "^>",
	)
	return `"` + replacer.Replace(value) + `"`
}

func todoLifecycleScriptShellArgs(goos string, shellPath string, script string) []string {
	switch shellNameFromPath(shellPath) {
	case "pwsh", "powershell":
		return []string{"-NoLogo", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script}
	}
	if goos == "windows" {
		switch shellNameFromPath(shellPath) {
		case "cmd":
			return []string{"/d", "/s", "/c", script}
		}
	}
	switch shellNameFromPath(shellPath) {
	case "zsh", "bash", "fish":
		return []string{"-i", "-c", script}
	}
	return []string{"-c", script}
}

func lifecycleScriptOutputTail(output string) string {
	if len(output) <= todoLifecycleScriptOutputTailLimit {
		return output
	}
	return output[len(output)-todoLifecycleScriptOutputTailLimit:]
}

func todoLifecycleScriptStatusKey(todoID string, phase string) string {
	return todoID + ":" + phase
}

func exitCodeFromCommandError(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
