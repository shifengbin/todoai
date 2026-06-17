package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type TodoProjectUIState struct {
	TodoView string `json:"todoView"`
}

type TodoProjectUIStateFile struct {
	Version      int                           `json:"version"`
	SidebarWidth int                           `json:"sidebarWidth,omitempty"`
	TodoProjects map[string]TodoProjectUIState `json:"todoProjects"`
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
}

func NewTodoProjectUIStateStore(configDir string) *TodoProjectUIStateStore {
	return &TodoProjectUIStateStore{
		path: filepath.Join(configDir, "todo-project-ui-state.json"),
	}
}

func (store *TodoProjectUIStateStore) Load() (TodoProjectUIStateFile, error) {
	data, err := os.ReadFile(store.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return emptyTodoProjectUIStateFile(), nil
		}
		return TodoProjectUIStateFile{}, err
	}

	var legacyState legacyTodoProjectUIStateFile
	if err := json.Unmarshal(data, &legacyState); err != nil {
		return emptyTodoProjectUIStateFile(), nil
	}
	return normalizeTodoProjectUIStateFile(todoProjectUIStateFileFromLegacy(legacyState)), nil
}

func (store *TodoProjectUIStateStore) Save(state TodoProjectUIStateFile) error {
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

func (store *TodoProjectUIStateStore) UpsertTodoProject(state TodoProjectUIStateFile, todoProjectID string, todoProjectState TodoProjectUIState) (TodoProjectUIStateFile, error) {
	state = normalizeTodoProjectUIStateFile(state)
	if todoProjectID == "" {
		return state, store.Save(state)
	}
	state.TodoProjects[todoProjectID] = todoProjectState
	return state, store.Save(state)
}

func (store *TodoProjectUIStateStore) UpsertSidebarWidth(state TodoProjectUIStateFile, sidebarWidth int) (TodoProjectUIStateFile, error) {
	state = normalizeTodoProjectUIStateFile(state)
	state.SidebarWidth = sidebarWidth
	return state, store.Save(state)
}

func (store *TodoProjectUIStateStore) DeleteTodoProjects(state TodoProjectUIStateFile, todoProjectIDs []string) (TodoProjectUIStateFile, error) {
	state = normalizeTodoProjectUIStateFile(state)
	for _, todoProjectID := range todoProjectIDs {
		delete(state.TodoProjects, todoProjectID)
	}
	return state, store.Save(state)
}

func emptyTodoProjectUIStateFile() TodoProjectUIStateFile {
	return TodoProjectUIStateFile{
		Version:      1,
		TodoProjects: map[string]TodoProjectUIState{},
	}
}

func normalizeTodoProjectUIStateFile(state TodoProjectUIStateFile) TodoProjectUIStateFile {
	if state.Version == 0 {
		state.Version = 1
	}
	if state.TodoProjects == nil {
		state.TodoProjects = map[string]TodoProjectUIState{}
	}
	return state
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
