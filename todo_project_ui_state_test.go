package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestTodoProjectUIStateStoreLoadMigratesLegacyStateAndDefaultsTodoOrdering(t *testing.T) {
	tests := []struct {
		name         string
		legacyJSON   string
		sidebarWidth int
	}{
		{
			name: "v0 project sidebar width",
			legacyJSON: `{
  "todoProjects": {
    "todo-project-1": {
      "todoView": "completed",
      "sidebarWidth": 372
    }
  }
}`,
			sidebarWidth: 372,
		},
		{
			name: "v1 top-level sidebar width",
			legacyJSON: `{
  "version": 1,
  "sidebarWidth": 384,
  "todoProjects": {
    "todo-project-1": {
      "todoView": "completed"
    }
  }
}`,
			sidebarWidth: 384,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configDir := t.TempDir()
			statePath := filepath.Join(configDir, "todo-project-ui-state.json")
			if err := os.WriteFile(statePath, []byte(test.legacyJSON), 0o644); err != nil {
				t.Fatalf("write legacy todo project UI state: %v", err)
			}

			state, err := NewTodoProjectUIStateStore(configDir).Load()
			if err != nil {
				t.Fatalf("load todo project UI state: %v", err)
			}
			if state.Version != todoProjectUIStateVersion {
				t.Fatalf("expected version %d, got %d", todoProjectUIStateVersion, state.Version)
			}
			if state.SidebarWidth != test.sidebarWidth {
				t.Fatalf("expected sidebar width %d, got %d", test.sidebarWidth, state.SidebarWidth)
			}
			if state.TodoSortMode != TodoSortModePriority {
				t.Fatalf("expected default sort mode %q, got %q", TodoSortModePriority, state.TodoSortMode)
			}
			if len(state.TodoOrders.NotStarted) != 0 || len(state.TodoOrders.InProgress) != 0 {
				t.Fatalf("expected empty manual orders, got %#v", state.TodoOrders)
			}
			if state.TodoOrdersInitialized {
				t.Fatal("expected legacy manual orders to remain uninitialized")
			}
			if got := state.TodoProjects["todo-project-1"].TodoView; got != "completed" {
				t.Fatalf("expected todo view completed, got %q", got)
			}
		})
	}
}

func TestTodoProjectUIStateStoreDefaultsMissingAndCorruptFiles(t *testing.T) {
	for _, test := range []struct {
		name    string
		content []byte
	}{
		{name: "missing"},
		{name: "corrupt", content: []byte("{not-json")},
	} {
		t.Run(test.name, func(t *testing.T) {
			configDir := t.TempDir()
			if test.content != nil {
				if err := os.WriteFile(filepath.Join(configDir, "todo-project-ui-state.json"), test.content, 0o644); err != nil {
					t.Fatalf("write corrupt state: %v", err)
				}
			}
			state, err := NewTodoProjectUIStateStore(configDir).Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if state.Version != todoProjectUIStateVersion || state.TodoSortMode != TodoSortModePriority || state.TodoProjects == nil {
				t.Fatalf("default state = %#v", state)
			}
		})
	}
}

func TestNormalizeTodoListUIStateFiltersInvalidIDsAndAppendsMissingTodos(t *testing.T) {
	state := TodoProjectUIStateFile{
		Version:      todoProjectUIStateVersion,
		TodoSortMode: TodoSortModeManual,
		TodoOrders: TodoManualOrders{
			NotStarted: []string{"todo-late", "deleted", "todo-late", "todo-progress"},
			InProgress: []string{"todo-progress", "todo-early"},
		},
		TodoProjects: map[string]TodoProjectUIState{},
	}
	todos := []Todo{
		{ID: "todo-late", Status: TodoStatusNotStarted, CreatedAt: "2026-07-02T00:00:00Z"},
		{ID: "todo-early", Status: TodoStatusNotStarted, CreatedAt: "2026-07-01T00:00:00Z"},
		{ID: "todo-progress", Status: TodoStatusInProgress, CreatedAt: "2026-07-03T00:00:00Z"},
		{ID: "todo-progress-missing", Status: TodoStatusInProgress, CreatedAt: "2026-07-01T00:00:00Z"},
		{ID: "todo-completed", Status: TodoStatusCompleted, CreatedAt: "2026-07-04T00:00:00Z"},
	}

	normalized := normalizeTodoListUIState(state, todos)

	wantNotStarted := []string{"todo-late", "todo-early"}
	if !reflect.DeepEqual(normalized.TodoOrders.NotStarted, wantNotStarted) {
		t.Fatalf("expected not-started order %#v, got %#v", wantNotStarted, normalized.TodoOrders.NotStarted)
	}
	wantInProgress := []string{"todo-progress", "todo-progress-missing"}
	if !reflect.DeepEqual(normalized.TodoOrders.InProgress, wantInProgress) {
		t.Fatalf("expected in-progress order %#v, got %#v", wantInProgress, normalized.TodoOrders.InProgress)
	}
}

func TestTodoProjectUIStateStoreKeepsNormalizedAutomaticOrdersUninitialized(t *testing.T) {
	store := NewTodoProjectUIStateStore(t.TempDir())
	state := normalizeTodoListUIState(TodoProjectUIStateFile{
		TodoSortMode: TodoSortModePriority,
	}, []Todo{
		{ID: "todo-late", Status: TodoStatusNotStarted, CreatedAt: "2026-07-02T00:00:00Z"},
		{ID: "todo-early", Status: TodoStatusNotStarted, CreatedAt: "2026-07-01T00:00:00Z"},
	})
	if state.TodoOrdersInitialized {
		t.Fatal("normalized automatic orders initialized before save")
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.TodoOrdersInitialized {
		t.Fatal("normalized automatic orders initialized after save and load")
	}
	if !reflect.DeepEqual(loaded.TodoOrders.NotStarted, []string{"todo-early", "todo-late"}) {
		t.Fatalf("NotStarted order = %#v", loaded.TodoOrders.NotStarted)
	}
}

func TestTodoProjectUIStateStoresKeepWorkspaceOrderingIndependent(t *testing.T) {
	workspaceA := NewTodoProjectUIStateStore(t.TempDir())
	workspaceB := NewTodoProjectUIStateStore(t.TempDir())
	stateA := TodoProjectUIStateFile{
		TodoSortMode: TodoSortModeManual,
		TodoOrders: TodoManualOrders{
			NotStarted: []string{"todo-a"},
		},
	}
	stateB := TodoProjectUIStateFile{
		TodoSortMode: TodoSortModeTime,
		TodoOrders: TodoManualOrders{
			NotStarted: []string{"todo-b"},
		},
	}
	if err := workspaceA.Save(stateA); err != nil {
		t.Fatalf("save workspace A UI state: %v", err)
	}
	if err := workspaceB.Save(stateB); err != nil {
		t.Fatalf("save workspace B UI state: %v", err)
	}

	loadedA, err := workspaceA.Load()
	if err != nil {
		t.Fatalf("load workspace A UI state: %v", err)
	}
	loadedB, err := workspaceB.Load()
	if err != nil {
		t.Fatalf("load workspace B UI state: %v", err)
	}

	if loadedA.TodoSortMode != TodoSortModeManual || !reflect.DeepEqual(loadedA.TodoOrders.NotStarted, []string{"todo-a"}) {
		t.Fatalf("unexpected workspace A state: %#v", loadedA)
	}
	if loadedB.TodoSortMode != TodoSortModeTime || !reflect.DeepEqual(loadedB.TodoOrders.NotStarted, []string{"todo-b"}) {
		t.Fatalf("unexpected workspace B state: %#v", loadedB)
	}
}

func TestTodoProjectUIStateStoreSerializesConcurrentFieldUpdates(t *testing.T) {
	store := NewTodoProjectUIStateStore(t.TempDir())
	initial := emptyTodoProjectUIStateFile()
	if err := store.Save(initial); err != nil {
		t.Fatalf("Save(initial) error = %v", err)
	}

	start := make(chan struct{})
	errors := make(chan error, 3)
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	go func() {
		defer waitGroup.Done()
		<-start
		_, err := store.UpsertSidebarWidth(360)
		errors <- err
	}()
	go func() {
		defer waitGroup.Done()
		<-start
		_, err := store.UpsertTodoProject("todo-project-a", TodoProjectUIState{TodoView: "completed"})
		errors <- err
	}()
	go func() {
		defer waitGroup.Done()
		<-start
		_, err := store.UpsertTodoListUIState(TodoListUIState{
			TodoSortMode: TodoSortModeManual,
			TodoOrders:   TodoManualOrders{NotStarted: []string{"todo-a"}},
		}, []Todo{{ID: "todo-a", Status: TodoStatusNotStarted}})
		errors <- err
	}()
	close(start)
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent update error = %v", err)
		}
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.SidebarWidth != 360 {
		t.Fatalf("SidebarWidth = %d, want 360", loaded.SidebarWidth)
	}
	if loaded.TodoProjects["todo-project-a"].TodoView != "completed" {
		t.Fatalf("TodoProjects = %#v", loaded.TodoProjects)
	}
	if loaded.TodoSortMode != TodoSortModeManual || !reflect.DeepEqual(loaded.TodoOrders.NotStarted, []string{"todo-a"}) {
		t.Fatalf("todo list UI state = %#v", loaded)
	}
}
