package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTerminalHistoryStore_Load_MissingFile(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if history.Version != 1 {
		t.Errorf("expected version 1, got %d", history.Version)
	}
	if len(history.Records) != 0 {
		t.Errorf("expected 0 records, got %d", len(history.Records))
	}
}

func TestTerminalHistoryStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	record := TerminalHistoryRecord{
		TerminalID:     "term-1",
		ProjectID:      "proj-1",
		TodoID:         "todo-1",
		TodoProjectID:  "tp-1",
		ShellName:      "bash",
		State:          ShellStateExited,
		CreatedAt:      "2026-06-12T00:00:00Z",
		LastSelectedAt: "2026-06-12T00:00:00Z",
		Output:         "hello world",
	}

	history := TerminalHistoryFile{Version: 1, Records: []TerminalHistoryRecord{record}}
	if err := store.Save(history); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
	r := loaded.Records[0]
	if r.TerminalID != "term-1" || r.Output != "hello world" {
		t.Errorf("unexpected record: %+v", r)
	}

	// Verify the file was written.
	if _, err := os.Stat(filepath.Join(dir, "terminal-history.json")); err != nil {
		t.Errorf("history file not found: %v", err)
	}
}

func TestTerminalHistoryStore_SaveOmitsTransientAgentStatusFields(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history := TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{
				TerminalID:     "term-1",
				ProjectID:      "proj-1",
				ShellName:      "bash",
				State:          ShellStateExited,
				CreatedAt:      "2026-06-12T00:00:00Z",
				LastSelectedAt: "2026-06-12T00:00:00Z",
				Output:         "hello",
			},
		},
	}

	if err := store.Save(history); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "terminal-history.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	raw := string(data)
	for _, field := range []string{"agentStatus", "activityState", "source", "confidence", "reason", "updatedAt", "runtimeTitle"} {
		if strings.Contains(raw, field) {
			t.Fatalf("history contains transient field %q: %s", field, raw)
		}
	}

	var decoded struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if _, ok := decoded.Records[0]["currentCommand"]; ok {
		t.Fatalf("history contains runtime command label: %#v", decoded.Records[0])
	}
}

func TestTerminalHistoryStore_UpsertRecord(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	record1 := TerminalHistoryRecord{TerminalID: "term-1", ProjectID: "proj-1", Output: "first"}
	record2 := TerminalHistoryRecord{TerminalID: "term-2", ProjectID: "proj-2", Output: "second"}
	record1Updated := TerminalHistoryRecord{TerminalID: "term-1", ProjectID: "proj-1", Output: "updated"}

	history := TerminalHistoryFile{Version: 1}
	history, err := store.UpsertRecord(history, record1)
	if err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}
	history, err = store.UpsertRecord(history, record2)
	if err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}
	history, err = store.UpsertRecord(history, record1Updated)
	if err != nil {
		t.Fatalf("UpsertRecord() error = %v", err)
	}

	loaded, _ := store.Load()
	if len(loaded.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(loaded.Records))
	}
	for _, r := range loaded.Records {
		if r.TerminalID == "term-1" && r.Output != "updated" {
			t.Errorf("term-1 output = %q, want %q", r.Output, "updated")
		}
	}
}

func TestTerminalHistoryStore_DeleteRecord(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history := TerminalHistoryFile{Version: 1, Records: []TerminalHistoryRecord{
		{TerminalID: "term-1", ProjectID: "proj-1"},
		{TerminalID: "term-2", ProjectID: "proj-1"},
	}}
	history, err := store.DeleteRecord(history, "term-1")
	if err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}

	loaded, _ := store.Load()
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
	if loaded.Records[0].TerminalID != "term-2" {
		t.Errorf("expected term-2, got %s", loaded.Records[0].TerminalID)
	}
}

func TestTerminalHistoryStore_DeleteRecordsByProject(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history := TerminalHistoryFile{Version: 1, Records: []TerminalHistoryRecord{
		{TerminalID: "term-1", ProjectID: "proj-1"},
		{TerminalID: "term-2", ProjectID: "proj-1"},
		{TerminalID: "term-3", ProjectID: "proj-2"},
	}}
	history, err := store.DeleteRecordsByProject(history, "proj-1")
	if err != nil {
		t.Fatalf("DeleteRecordsByProject() error = %v", err)
	}

	loaded, _ := store.Load()
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
	if loaded.Records[0].TerminalID != "term-3" {
		t.Errorf("expected term-3, got %s", loaded.Records[0].TerminalID)
	}
}

func TestTerminalHistoryStore_DeleteRecordsByTodo(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history := TerminalHistoryFile{Version: 1, Records: []TerminalHistoryRecord{
		{TerminalID: "term-1", ProjectID: "proj-1", TodoID: "todo-1"},
		{TerminalID: "term-2", ProjectID: "proj-1", TodoID: "todo-2"},
	}}
	history, err := store.DeleteRecordsByTodo(history, "todo-1")
	if err != nil {
		t.Fatalf("DeleteRecordsByTodo() error = %v", err)
	}

	loaded, _ := store.Load()
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
	if loaded.Records[0].TerminalID != "term-2" {
		t.Errorf("expected term-2, got %s", loaded.Records[0].TerminalID)
	}
}

func TestTerminalHistoryStore_DeleteRecordsByTodoProject(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)

	history := TerminalHistoryFile{Version: 1, Records: []TerminalHistoryRecord{
		{TerminalID: "term-1", ProjectID: "proj-1", TodoProjectID: "tp-1"},
		{TerminalID: "term-2", ProjectID: "proj-1", TodoProjectID: "tp-2"},
	}}
	history, err := store.DeleteRecordsByTodoProject(history, "tp-1")
	if err != nil {
		t.Fatalf("DeleteRecordsByTodoProject() error = %v", err)
	}

	loaded, _ := store.Load()
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
	if loaded.Records[0].TerminalID != "term-2" {
		t.Errorf("expected term-2, got %s", loaded.Records[0].TerminalID)
	}
}

func TestAppendTerminalOutput_UnderLimit(t *testing.T) {
	result := AppendTerminalOutput("", "hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestAppendTerminalOutput_OverLimit(t *testing.T) {
	// Create output that exceeds the limit.
	existing := make([]byte, MaxTerminalOutputBytes)
	for i := range existing {
		existing[i] = 'a'
	}
	result := AppendTerminalOutput(string(existing), "NEW")
	if len(result) > MaxTerminalOutputBytes {
		t.Errorf("output length %d exceeds limit %d", len(result), MaxTerminalOutputBytes)
	}
	if result[len(result)-3:] != "NEW" {
		t.Errorf("expected trailing 'NEW', got %q", result[len(result)-3:])
	}
}

func TestAppendTerminalOutput_MultipleAppends(t *testing.T) {
	output := ""
	for i := 0; i < 100; i++ {
		output = AppendTerminalOutput(output, "data-chunk-")
	}
	if len(output) > MaxTerminalOutputBytes {
		t.Errorf("output length %d exceeds limit %d", len(output), MaxTerminalOutputBytes)
	}
}

func TestRestoreTerminals_ValidRecords(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)
	store.Save(TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{
				TerminalID:     "term-1",
				ProjectID:      "proj-1",
				TodoID:         "todo-1",
				TodoProjectID:  "tp-1",
				ShellName:      "bash",
				State:          ShellStateExited,
				CreatedAt:      "2026-06-12T00:00:00Z",
				LastSelectedAt: "2026-06-12T00:00:00Z",
				Output:         "some output",
			},
		},
	})

	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithTerminalHistoryStore(store),
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	state := ProjectState{
		Projects:     []Project{{ID: "proj-1", Name: "test", Path: "/tmp/test", Available: true}},
		Todos:        []Todo{{ID: "todo-1", Title: "Test TODO", Status: TodoStatusInProgress}},
		TodoProjects: []TodoProject{{ID: "tp-1", TodoID: "todo-1", ProjectID: "proj-1"}},
	}

	records := manager.RestoreTerminals(state)
	if len(records) != 1 {
		t.Fatalf("expected 1 restored record, got %d", len(records))
	}

	terminals := manager.Terminals()
	if len(terminals) != 1 {
		t.Fatalf("expected 1 terminal, got %d", len(terminals))
	}
	terminal := terminals[0]
	if terminal.State != ShellStateExited {
		t.Errorf("expected state %q, got %q", ShellStateExited, terminal.State)
	}
	if terminal.ID != "term-1" {
		t.Errorf("expected terminal ID term-1, got %s", terminal.ID)
	}
}

func TestRestoreTerminals_UsesTodoProjectCopyWhenGlobalCandidateIsMissing(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)
	store.Save(TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{
				TerminalID:     "term-1",
				ProjectID:      "proj-1",
				TodoID:         "todo-1",
				TodoProjectID:  "tp-1",
				ShellName:      "bash",
				State:          ShellStateExited,
				CreatedAt:      "2026-06-12T00:00:00Z",
				LastSelectedAt: "2026-06-12T00:00:00Z",
				Output:         "restored output",
			},
		},
	})

	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithTerminalHistoryStore(store),
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	state := ProjectState{
		Projects: []Project{},
		Todos:    []Todo{{ID: "todo-1", Title: "Test TODO", Status: TodoStatusInProgress}},
		TodoProjects: []TodoProject{{
			ID:              "tp-1",
			TodoID:          "todo-1",
			ProjectID:       "proj-1",
			SourceProjectID: "proj-1",
			Name:            "frontend-app",
			Path:            "/repo/frontend-app",
			Available:       true,
		}},
	}

	records := manager.RestoreTerminals(state)
	if len(records) != 1 {
		t.Fatalf("expected 1 restored record, got %d", len(records))
	}

	terminals := manager.Terminals()
	if len(terminals) != 1 {
		t.Fatalf("expected 1 terminal, got %d", len(terminals))
	}
	terminal := terminals[0]
	if terminal.ID != "term-1" {
		t.Fatalf("terminal ID = %q, want term-1", terminal.ID)
	}
	if terminal.Output != "restored output" {
		t.Fatalf("terminal output = %q, want restored output", terminal.Output)
	}
}

func TestRestoreTerminals_DropsOrphanedRecords(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)
	store.Save(TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{TerminalID: "term-valid", ProjectID: "proj-1", ShellName: "bash"},
			{TerminalID: "term-orphan", ProjectID: "proj-missing", ShellName: "bash"},
		},
	})

	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithTerminalHistoryStore(store),
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	state := ProjectState{
		Projects: []Project{{ID: "proj-1", Name: "test", Path: "/tmp/test", Available: true}},
	}

	records := manager.RestoreTerminals(state)
	if len(records) != 1 {
		t.Fatalf("expected 1 restored record, got %d", len(records))
	}
	if records[0].TerminalID != "term-valid" {
		t.Errorf("expected term-valid, got %s", records[0].TerminalID)
	}

	// Verify orphan was removed from disk.
	loaded, _ := store.Load()
	if len(loaded.Records) != 1 {
		t.Errorf("expected 1 record on disk, got %d", len(loaded.Records))
	}
}

func TestRestoreTerminals_NormalizesState(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)
	store.Save(TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{
				TerminalID: "term-1", ProjectID: "proj-1", ShellName: "bash",
				State: ShellStateRunning, // persisted as running
			},
		},
	})

	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithTerminalHistoryStore(store),
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	state := ProjectState{
		Projects: []Project{{ID: "proj-1", Name: "test", Path: "/tmp/test", Available: true}},
	}

	manager.RestoreTerminals(state)
	terminals := manager.Terminals()
	if len(terminals) != 1 {
		t.Fatalf("expected 1 terminal, got %d", len(terminals))
	}
	if terminals[0].State != ShellStateExited {
		t.Errorf("expected state %q, got %q", ShellStateExited, terminals[0].State)
	}
}

func TestRestoreTerminals_NoHistoryStore(t *testing.T) {
	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	state := ProjectState{
		Projects: []Project{{ID: "proj-1", Name: "test", Path: "/tmp/test", Available: true}},
	}

	records := manager.RestoreTerminals(state)
	if records != nil {
		t.Errorf("expected nil records when no history store, got %v", records)
	}
}

func TestTerminalHistory_CleanupOnDelete(t *testing.T) {
	dir := t.TempDir()
	store := NewTerminalHistoryStore(dir)
	store.Save(TerminalHistoryFile{
		Version: 1,
		Records: []TerminalHistoryRecord{
			{TerminalID: "term-1", ProjectID: "proj-1", TodoID: "todo-1", TodoProjectID: "tp-1", ShellName: "bash"},
		},
	})

	manager := NewShellSessionManager(nil, ShellSessionCallbacks{},
		WithTerminalHistoryStore(store),
		WithShellPathResolver(func() string { return "/bin/bash" }),
	)

	// Register terminal so it exists in the manager.
	state := ProjectState{
		Projects:     []Project{{ID: "proj-1", Name: "test", Path: "/tmp/test", Available: true}},
		Todos:        []Todo{{ID: "todo-1", Title: "Test", Status: TodoStatusInProgress}},
		TodoProjects: []TodoProject{{ID: "tp-1", TodoID: "todo-1", ProjectID: "proj-1"}},
	}
	manager.RestoreTerminals(state)

	// Delete by TODO project.
	manager.DeleteTodoProjectTerminals("tp-1")

	loaded, _ := store.Load()
	if len(loaded.Records) != 0 {
		t.Errorf("expected 0 records after cleanup, got %d", len(loaded.Records))
	}
}
