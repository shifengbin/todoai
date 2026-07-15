package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const todoProjectUIStateVersion = 2

const (
	TodoSortModePriority = "priority"
	TodoSortModeTime     = "time"
	TodoSortModeManual   = "manual"
)

type TodoProjectUIState struct {
	TodoView string `json:"todoView"`
}

type TodoManualOrders struct {
	NotStarted []string `json:"notStarted,omitempty"`
	InProgress []string `json:"inProgress,omitempty"`
}

type TodoListUIState struct {
	TodoSortMode string           `json:"todoSortMode"`
	TodoOrders   TodoManualOrders `json:"todoOrders"`
}

type TodoProjectUIStateFile struct {
	Version               int                           `json:"version"`
	SidebarWidth          int                           `json:"sidebarWidth,omitempty"`
	TodoSortMode          string                        `json:"todoSortMode,omitempty"`
	TodoOrdersInitialized bool                          `json:"todoOrdersInitialized,omitempty"`
	TodoOrders            TodoManualOrders              `json:"todoOrders,omitempty"`
	TodoProjects          map[string]TodoProjectUIState `json:"todoProjects"`
}

type legacyTodoProjectUIState struct {
	TodoView     string `json:"todoView"`
	SidebarWidth int    `json:"sidebarWidth"`
}

type legacyTodoProjectUIStateFile struct {
	Version      int                                 `json:"version"`
	SidebarWidth int                                 `json:"sidebarWidth"`
	TodoProjects map[string]legacyTodoProjectUIState `json:"todoProjects"`
}

type TodoProjectUIStateStore struct {
	path string
	mu   sync.Mutex
}

func NewTodoProjectUIStateStore(configDir string) *TodoProjectUIStateStore {
	return &TodoProjectUIStateStore{
		path: filepath.Join(configDir, "todo-project-ui-state.json"),
	}
}

func (store *TodoProjectUIStateStore) Load() (TodoProjectUIStateFile, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.loadLocked()
}

func (store *TodoProjectUIStateStore) loadLocked() (TodoProjectUIStateFile, error) {
	data, err := os.ReadFile(store.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return emptyTodoProjectUIStateFile(), nil
		}
		return TodoProjectUIStateFile{}, err
	}

	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return emptyTodoProjectUIStateFile(), nil
	}
	if header.Version >= todoProjectUIStateVersion {
		var state TodoProjectUIStateFile
		if err := json.Unmarshal(data, &state); err != nil {
			return emptyTodoProjectUIStateFile(), nil
		}
		return normalizeTodoProjectUIStateFile(state), nil
	}

	var legacyState legacyTodoProjectUIStateFile
	if err := json.Unmarshal(data, &legacyState); err != nil {
		return emptyTodoProjectUIStateFile(), nil
	}
	return normalizeTodoProjectUIStateFile(todoProjectUIStateFileFromLegacy(legacyState)), nil
}

func (store *TodoProjectUIStateStore) Save(state TodoProjectUIStateFile) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.saveLocked(state)
}

func (store *TodoProjectUIStateStore) saveLocked(state TodoProjectUIStateFile) error {
	state = normalizeTodoProjectUIStateFile(state)
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := store.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, store.path)
}

func (store *TodoProjectUIStateStore) update(mutate func(*TodoProjectUIStateFile) error) (TodoProjectUIStateFile, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	state, err := store.loadLocked()
	if err != nil {
		return TodoProjectUIStateFile{}, err
	}
	if err := mutate(&state); err != nil {
		return state, err
	}
	state = normalizeTodoProjectUIStateFile(state)
	if err := store.saveLocked(state); err != nil {
		return state, err
	}
	return state, nil
}

func (store *TodoProjectUIStateStore) UpsertTodoProject(todoProjectID string, todoProjectState TodoProjectUIState) (TodoProjectUIStateFile, error) {
	return store.update(func(state *TodoProjectUIStateFile) error {
		if todoProjectID != "" {
			state.TodoProjects[todoProjectID] = todoProjectState
		}
		return nil
	})
}

func (store *TodoProjectUIStateStore) UpsertSidebarWidth(sidebarWidth int) (TodoProjectUIStateFile, error) {
	return store.update(func(state *TodoProjectUIStateFile) error {
		state.SidebarWidth = sidebarWidth
		return nil
	})
}

func (store *TodoProjectUIStateStore) UpsertTodoListUIState(listState TodoListUIState, todos []Todo) (TodoProjectUIStateFile, error) {
	return store.update(func(state *TodoProjectUIStateFile) error {
		if !supportedTodoSortMode(listState.TodoSortMode) {
			return fmt.Errorf("unsupported todo sort mode: %s", listState.TodoSortMode)
		}
		state.TodoSortMode = listState.TodoSortMode
		state.TodoOrders = TodoManualOrders{
			NotStarted: append([]string{}, listState.TodoOrders.NotStarted...),
			InProgress: append([]string{}, listState.TodoOrders.InProgress...),
		}
		if listState.TodoSortMode == TodoSortModeManual {
			state.TodoOrdersInitialized = true
		}
		*state = normalizeTodoListUIState(*state, todos)
		return nil
	})
}

func (store *TodoProjectUIStateStore) DeleteTodoProjects(todoProjectIDs []string) (TodoProjectUIStateFile, error) {
	return store.update(func(state *TodoProjectUIStateFile) error {
		for _, todoProjectID := range todoProjectIDs {
			delete(state.TodoProjects, todoProjectID)
		}
		return nil
	})
}

func emptyTodoProjectUIStateFile() TodoProjectUIStateFile {
	return TodoProjectUIStateFile{
		Version:      todoProjectUIStateVersion,
		TodoSortMode: TodoSortModePriority,
		TodoProjects: map[string]TodoProjectUIState{},
	}
}

func normalizeTodoProjectUIStateFile(state TodoProjectUIStateFile) TodoProjectUIStateFile {
	state.Version = todoProjectUIStateVersion
	state.TodoSortMode = normalizeTodoSortMode(state.TodoSortMode)
	if state.TodoSortMode == TodoSortModeManual {
		state.TodoOrdersInitialized = true
	}
	if state.TodoProjects == nil {
		state.TodoProjects = map[string]TodoProjectUIState{}
	}
	return state
}

func normalizeTodoSortMode(mode string) string {
	if supportedTodoSortMode(mode) {
		return mode
	}
	return TodoSortModePriority
}

func supportedTodoSortMode(mode string) bool {
	switch mode {
	case TodoSortModePriority, TodoSortModeTime, TodoSortModeManual:
		return true
	default:
		return false
	}
}

func normalizeTodoListUIState(state TodoProjectUIStateFile, todos []Todo) TodoProjectUIStateFile {
	state = normalizeTodoProjectUIStateFile(state)
	state.TodoOrders.NotStarted = normalizeTodoOrder(state.TodoOrders.NotStarted, todos, TodoStatusNotStarted)
	state.TodoOrders.InProgress = normalizeTodoOrder(state.TodoOrders.InProgress, todos, TodoStatusInProgress)
	return state
}

type todoOrderCandidate struct {
	todo  Todo
	index int
}

func normalizeTodoOrder(order []string, todos []Todo, status string) []string {
	valid := map[string]todoOrderCandidate{}
	for index, todo := range todos {
		if normalizedTodoOrderStatus(todo.Status) == status && todo.ID != "" {
			valid[todo.ID] = todoOrderCandidate{todo: todo, index: index}
		}
	}

	normalized := make([]string, 0, len(valid))
	seen := map[string]bool{}
	for _, todoID := range order {
		if _, ok := valid[todoID]; !ok || seen[todoID] {
			continue
		}
		seen[todoID] = true
		normalized = append(normalized, todoID)
	}

	missing := make([]todoOrderCandidate, 0, len(valid)-len(normalized))
	for todoID, candidate := range valid {
		if !seen[todoID] {
			missing = append(missing, candidate)
		}
	}
	sort.SliceStable(missing, func(leftIndex int, rightIndex int) bool {
		return todoOrderCandidateLess(missing[leftIndex], missing[rightIndex])
	})
	for _, candidate := range missing {
		normalized = append(normalized, candidate.todo.ID)
	}
	return normalized
}

func normalizedTodoOrderStatus(status string) string {
	if status == TodoStatusActive || status == "" {
		return TodoStatusNotStarted
	}
	return status
}

func todoOrderCandidateLess(left todoOrderCandidate, right todoOrderCandidate) bool {
	leftTime, leftErr := time.Parse(time.RFC3339, left.todo.CreatedAt)
	rightTime, rightErr := time.Parse(time.RFC3339, right.todo.CreatedAt)
	if leftErr == nil && rightErr == nil && !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}
	if leftErr == nil && rightErr != nil {
		return true
	}
	if leftErr != nil && rightErr == nil {
		return false
	}
	return left.index < right.index
}

func todoProjectUIStateFileFromLegacy(legacyState legacyTodoProjectUIStateFile) TodoProjectUIStateFile {
	state := TodoProjectUIStateFile{
		Version:      legacyState.Version,
		SidebarWidth: legacyState.SidebarWidth,
		TodoProjects: map[string]TodoProjectUIState{},
	}
	if len(legacyState.TodoProjects) == 0 {
		return state
	}
	ids := make([]string, 0, len(legacyState.TodoProjects))
	for todoProjectID := range legacyState.TodoProjects {
		ids = append(ids, todoProjectID)
	}
	sort.Strings(ids)
	for _, todoProjectID := range ids {
		legacyTodoProjectState := legacyState.TodoProjects[todoProjectID]
		state.TodoProjects[todoProjectID] = TodoProjectUIState{TodoView: legacyTodoProjectState.TodoView}
		if state.SidebarWidth == 0 && legacyTodoProjectState.SidebarWidth > 0 {
			state.SidebarWidth = legacyTodoProjectState.SidebarWidth
		}
	}
	return state
}
