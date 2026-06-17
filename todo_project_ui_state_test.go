package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTodoProjectUIStateStoreLoadMissingFile(t *testing.T) {
	store := NewTodoProjectUIStateStore(t.TempDir())

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Version != 1 {
		t.Fatalf("Version = %d, want 1", state.Version)
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0", len(state.TodoProjects))
	}
}

func TestTodoProjectUIStateStoreSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewTodoProjectUIStateStore(dir)

	state := TodoProjectUIStateFile{
		Version: 1,
		TodoProjects: map[string]TodoProjectUIState{
			"todo-project-a": {TodoView: "completed", SidebarWidth: 360},
			"todo-project-b": {TodoView: "in-progress", SidebarWidth: 420},
		},
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.TodoProjects["todo-project-a"].TodoView != "completed" {
		t.Fatalf("todo-project-a TodoView = %q, want completed", loaded.TodoProjects["todo-project-a"].TodoView)
	}
	if loaded.TodoProjects["todo-project-b"].SidebarWidth != 420 {
		t.Fatalf("todo-project-b SidebarWidth = %d, want 420", loaded.TodoProjects["todo-project-b"].SidebarWidth)
	}
	if _, err := os.Stat(filepath.Join(dir, "todo-project-ui-state.json")); err != nil {
		t.Fatalf("ui state file not found: %v", err)
	}
}

func TestTodoProjectUIStateStoreLoadInvalidFileReturnsEmptyState(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "todo-project-ui-state.json"), []byte("{invalid"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store := NewTodoProjectUIStateStore(dir)

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Version != 1 {
		t.Fatalf("Version = %d, want 1", state.Version)
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0", len(state.TodoProjects))
	}
}

func TestTodoProjectUIStateStoreUpsertAndDeleteByTodoProject(t *testing.T) {
	store := NewTodoProjectUIStateStore(t.TempDir())
	state := TodoProjectUIStateFile{Version: 1}

	var err error
	state, err = store.UpsertTodoProject(state, "todo-project-a", TodoProjectUIState{TodoView: "completed", SidebarWidth: 360})
	if err != nil {
		t.Fatalf("UpsertTodoProject(a) error = %v", err)
	}
	state, err = store.UpsertTodoProject(state, "todo-project-b", TodoProjectUIState{TodoView: "in-progress", SidebarWidth: 420})
	if err != nil {
		t.Fatalf("UpsertTodoProject(b) error = %v", err)
	}
	state, err = store.DeleteTodoProjects(state, []string{"todo-project-a"})
	if err != nil {
		t.Fatalf("DeleteTodoProjects() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := loaded.TodoProjects["todo-project-a"]; ok {
		t.Fatalf("todo-project-a still exists: %#v", loaded.TodoProjects)
	}
	if loaded.TodoProjects["todo-project-b"].TodoView != "in-progress" {
		t.Fatalf("todo-project-b TodoView = %q, want in-progress", loaded.TodoProjects["todo-project-b"].TodoView)
	}
}
