package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectManagerLoadsEmptyStateWhenConfigIsMissing(t *testing.T) {
	manager := NewProjectManager(filepath.Join(t.TempDir(), "projects.json"))

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if state.Version != projectConfigVersion {
		t.Fatalf("Version = %d, want %d", state.Version, projectConfigVersion)
	}
	if len(state.Projects) != 0 {
		t.Fatalf("Projects length = %d, want 0", len(state.Projects))
	}
	if state.ActiveProjectID != "" {
		t.Fatalf("ActiveProjectID = %q, want empty", state.ActiveProjectID)
	}
	if len(state.Todos) != 0 {
		t.Fatalf("Todos length = %d, want 0", len(state.Todos))
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0", len(state.TodoProjects))
	}
	if state.ActiveTodoID != "" {
		t.Fatalf("ActiveTodoID = %q, want empty", state.ActiveTodoID)
	}
	if state.ActiveTodoProjectID != "" {
		t.Fatalf("ActiveTodoProjectID = %q, want empty", state.ActiveTodoProjectID)
	}
}

func TestProjectManagerLoadsLegacyProjectStateWithTodoDefaults(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	legacyJSON := `{
  "version": 1,
  "projects": [
    {
      "id": "project-a",
      "name": "alpha",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "activeProjectId": "project-a"
}`
	if err := os.WriteFile(configPath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	manager := NewProjectManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].ID != "project-a" {
		t.Fatalf("Projects = %#v, want legacy project", state.Projects)
	}
	if len(state.Todos) != 0 {
		t.Fatalf("Todos length = %d, want 0", len(state.Todos))
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0", len(state.TodoProjects))
	}
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" {
		t.Fatalf("active TODO context = %q/%q, want empty", state.ActiveTodoID, state.ActiveTodoProjectID)
	}
}

func TestProjectManagerNormalizesLegacyTodoWorkflowStatuses(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	legacyJSON := `{
  "version": 1,
  "projects": [
    {
      "id": "project-a",
      "name": "alpha",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todos": [
    {
      "id": "todo-active",
      "title": "活动任务",
      "priority": "high",
      "status": "active",
      "createdAt": "2026-06-10T09:00:00Z"
    },
    {
      "id": "todo-completed",
      "title": "已完成任务",
      "priority": "medium",
      "status": "archived",
      "archivedReason": "completed",
      "createdAt": "2026-06-10T08:00:00Z",
      "archivedAt": "2026-06-10T10:00:00Z"
    },
    {
      "id": "todo-deleted",
      "title": "已删除任务",
      "priority": "low",
      "status": "archived",
      "archivedReason": "deleted",
      "createdAt": "2026-06-10T07:00:00Z"
    }
  ],
  "todoProjects": [
    {
      "id": "todo-project-active",
      "todoId": "todo-active",
      "projectId": "project-a",
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    },
    {
      "id": "todo-project-deleted",
      "todoId": "todo-deleted",
      "projectId": "project-a",
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "activeProjectId": "project-a",
  "activeTodoId": "todo-active",
  "activeTodoProjectId": "todo-project-active"
}`
	if err := os.WriteFile(configPath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	manager := NewProjectManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(state.Todos) != 2 {
		t.Fatalf("Todos = %#v, want active and completed legacy todos only", state.Todos)
	}
	active := findTodo(state.Todos, "todo-active")
	if active == nil || active.Status != "not-started" {
		t.Fatalf("active legacy todo = %#v, want not-started", active)
	}
	completed := findTodo(state.Todos, "todo-completed")
	if completed == nil || completed.Status != "completed" || completed.ArchivedReason != "" {
		t.Fatalf("completed legacy todo = %#v, want completed without archive reason", completed)
	}
	if findTodo(state.Todos, "todo-deleted") != nil {
		t.Fatalf("deleted legacy todo still visible: %#v", state.Todos)
	}
	if len(state.TodoProjects) != 1 || state.TodoProjects[0].ID != "todo-project-active" {
		t.Fatalf("TodoProjects = %#v, want deleted todo association removed", state.TodoProjects)
	}
}

func TestProjectManagerPersistsProjectCreatedFromDirectory(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)

	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(func() string { return "project-1" }),
		WithProjectClock(func() time.Time { return now }),
	)

	project, added, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if !added {
		t.Fatalf("AddProjectPath() added = false, want true")
	}
	if project.ID != "project-1" {
		t.Fatalf("Project ID = %q, want project-1", project.ID)
	}
	if project.Name != filepath.Base(projectDir) {
		t.Fatalf("Project name = %q, want %q", project.Name, filepath.Base(projectDir))
	}
	if project.Path != projectDir {
		t.Fatalf("Project path = %q, want %q", project.Path, projectDir)
	}
	if !project.Available {
		t.Fatalf("Project Available = false, want true")
	}

	reloaded := NewProjectManager(configPath)
	state, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Projects) != 1 {
		t.Fatalf("Projects length = %d, want 1", len(state.Projects))
	}
	if state.ActiveProjectID != "project-1" {
		t.Fatalf("ActiveProjectID = %q, want project-1", state.ActiveProjectID)
	}
	if state.Projects[0].Name != filepath.Base(projectDir) {
		t.Fatalf("Reloaded project name = %q, want %q", state.Projects[0].Name, filepath.Base(projectDir))
	}
}

func TestProjectManagerSelectsExistingProjectForDuplicatePath(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	nextID := sequenceIDs("project-1", "project-2")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(nextID))

	first, added, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("first AddProjectPath() error = %v", err)
	}
	if !added {
		t.Fatalf("first AddProjectPath() added = false, want true")
	}

	second, added, err := manager.AddProjectPath(filepath.Clean(projectDir))
	if err != nil {
		t.Fatalf("second AddProjectPath() error = %v", err)
	}
	if added {
		t.Fatalf("second AddProjectPath() added = true, want false")
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate project ID = %q, want %q", second.ID, first.ID)
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Projects) != 1 {
		t.Fatalf("Projects length = %d, want 1", len(state.Projects))
	}
	if state.ActiveProjectID != first.ID {
		t.Fatalf("ActiveProjectID = %q, want %q", state.ActiveProjectID, first.ID)
	}
}

func TestProjectManagerCreatesTodoAndAssociatesProjects(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "todo-a", "todo-project-a", "todo-project-b", "todo-b", "todo-project-c")),
		WithProjectClock(func() time.Time { return now }),
	)
	projectA, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(A) error = %v", err)
	}
	projectB, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(B) error = %v", err)
	}

	state, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if state.ActiveTodoID != "" {
		t.Fatalf("ActiveTodoID = %q, want unchanged empty", state.ActiveTodoID)
	}
	if len(state.Todos) != 1 || state.Todos[0].Title != "修复登录问题" || state.Todos[0].Status != TodoStatusNotStarted {
		t.Fatalf("Todos = %#v, want not-started todo", state.Todos)
	}

	state, err = manager.AssociateProjectWithTodo("todo-a", projectA.ID)
	if err != nil {
		t.Fatalf("AssociateProjectWithTodo(A) error = %v", err)
	}
	state, err = manager.AssociateProjectWithTodo("todo-a", projectB.ID)
	if err != nil {
		t.Fatalf("AssociateProjectWithTodo(B) error = %v", err)
	}
	state, err = manager.AssociateProjectWithTodo("todo-a", projectA.ID)
	if err != nil {
		t.Fatalf("duplicate AssociateProjectWithTodo(A) error = %v", err)
	}
	if len(state.TodoProjects) != 2 {
		t.Fatalf("TodoProjects length = %d, want 2 after duplicate association", len(state.TodoProjects))
	}

	state, err = manager.CreateTodo(CreateTodoRequest{Title: "升级依赖"})
	if err != nil {
		t.Fatalf("CreateTodo(second) error = %v", err)
	}
	state, err = manager.AssociateProjectWithTodo("todo-b", projectA.ID)
	if err != nil {
		t.Fatalf("AssociateProjectWithTodo(second todo) error = %v", err)
	}
	if countTodoProjectAssociations(state.TodoProjects, projectA.ID) != 2 {
		t.Fatalf("project A associations = %#v, want project in both todos", state.TodoProjects)
	}

	reloaded := NewProjectManager(configPath)
	persisted, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(persisted.Todos) != 2 {
		t.Fatalf("persisted Todos length = %d, want 2", len(persisted.Todos))
	}
	if len(persisted.TodoProjects) != 3 {
		t.Fatalf("persisted TodoProjects length = %d, want 3", len(persisted.TodoProjects))
	}
}

func TestProjectManagerCreatesNotStartedTodoWithoutChangingActiveContext(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "todo-existing", "todo-project-existing", "todo-new", "todo-project-new")),
		WithProjectClock(func() time.Time { return now }),
	)
	projectA, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(A) error = %v", err)
	}
	projectB, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(B) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "升级依赖"}); err != nil {
		t.Fatalf("CreateTodo(existing) error = %v", err)
	}
	if _, err := manager.AssociateProjectWithTodo("todo-existing", projectA.ID); err != nil {
		t.Fatalf("AssociateProjectWithTodo(existing) error = %v", err)
	}
	if _, _, _, err := manager.SelectTodoProject("todo-project-existing"); err != nil {
		t.Fatalf("SelectTodoProject(existing) error = %v", err)
	}

	state, err := manager.CreateTodo(CreateTodoRequest{
		Title:      "修复登录问题",
		ProjectIDs: []string{projectB.ID},
	})
	if err != nil {
		t.Fatalf("CreateTodo(new) error = %v", err)
	}

	todo := findTodo(state.Todos, "todo-new")
	if todo == nil || todo.Status != "not-started" {
		t.Fatalf("new todo = %#v, want not-started", todo)
	}
	if state.ActiveTodoID != "todo-existing" || state.ActiveTodoProjectID != "todo-project-existing" || state.ActiveProjectID != projectA.ID {
		t.Fatalf("active context = %q/%q/%q, want existing todo context", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID)
	}
	if len(state.TodoProjects) != 2 || state.TodoProjects[1].TodoID != "todo-new" || state.TodoProjects[1].ProjectID != projectB.ID {
		t.Fatalf("TodoProjects = %#v, want new project association without selecting it", state.TodoProjects)
	}
}

func TestProjectManagerChangesTodoWorkflowStatusManually(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 22, 9, 30, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("todo-a")),
		WithProjectClock(func() time.Time { return now }),
	)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题"}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	now = now.Add(2 * time.Hour)
	state, err := manager.ChangeTodoStatus("todo-a", "in-progress")
	if err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}
	if state.Todos[0].Status != "in-progress" {
		t.Fatalf("Status = %q, want in-progress", state.Todos[0].Status)
	}
	if startedAt := todoJSONField(t, state.Todos[0], "startedAt"); startedAt != "2026-06-22T11:30:00Z" {
		t.Fatalf("StartedAt = %q, want start transition timestamp", startedAt)
	}

	now = now.Add(30 * time.Minute)
	if _, err := manager.ChangeTodoStatus("todo-a", "not-started"); err == nil {
		t.Fatal("ChangeTodoStatus(not-started) error = nil, want invalid status transition error")
	}
	state, err = manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Todos[0].Status != "in-progress" {
		t.Fatalf("Status after rejected transition = %q, want in-progress", state.Todos[0].Status)
	}
	if startedAt := todoJSONField(t, state.Todos[0], "startedAt"); startedAt != "2026-06-22T11:30:00Z" {
		t.Fatalf("StartedAt after rejected transition = %q, want original start timestamp", startedAt)
	}

	if _, err := manager.ChangeTodoStatus("todo-a", "completed"); err == nil {
		t.Fatal("ChangeTodoStatus(completed) error = nil, want invalid status error")
	}
}

func TestProjectManagerCreatesTodoWithDetailsAndOptionalProjects(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "todo-a", "todo-project-a", "todo-project-b")),
		WithProjectClock(func() time.Time { return now }),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	otherProject, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(other) error = %v", err)
	}

	state, err := manager.CreateTodo(CreateTodoRequest{
		Title:       "  修复登录问题  ",
		Description: "  登录后跳回首页  ",
		Priority:    TodoPriorityHigh,
		ProjectIDs:  []string{project.ID, otherProject.ID},
		InitializationFiles: []TodoInitializationFileSnapshot{
			{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "请先阅读任务说明"},
			{Name: "Prompt", Description: "可选提示词", FileName: "prompt.md", Content: "生成实现计划"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1", len(state.Todos))
	}
	todo := state.Todos[0]
	if todo.Title != "修复登录问题" {
		t.Fatalf("Title = %q, want trimmed title", todo.Title)
	}
	if todo.Description != "登录后跳回首页" {
		t.Fatalf("Description = %q, want trimmed description", todo.Description)
	}
	if todo.Priority != TodoPriorityHigh {
		t.Fatalf("Priority = %q, want %q", todo.Priority, TodoPriorityHigh)
	}
	assertTodoInitializationFileSnapshots(t, todo.InitializationFiles, []TodoInitializationFileSnapshot{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "请先阅读任务说明"},
		{Name: "Prompt", Description: "可选提示词", FileName: "prompt.md", Content: "生成实现计划"},
	})
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" || state.ActiveProjectID != otherProject.ID {
		t.Fatalf("active context = %q/%q/%q, want unchanged project context %q", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, otherProject.ID)
	}
	if len(state.TodoProjects) != 2 {
		t.Fatalf("TodoProjects length = %d, want 2", len(state.TodoProjects))
	}
	if state.TodoProjects[0].TodoID != "todo-a" || state.TodoProjects[0].ProjectID != project.ID {
		t.Fatalf("first TodoProject = %#v, want first selected project", state.TodoProjects[0])
	}
	if state.TodoProjects[1].TodoID != "todo-a" || state.TodoProjects[1].ProjectID != otherProject.ID {
		t.Fatalf("second TodoProject = %#v, want second selected project", state.TodoProjects[1])
	}

	reloaded := NewProjectManager(configPath)
	persisted, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(persisted.Todos) != 1 || persisted.Todos[0].Description != "登录后跳回首页" || persisted.Todos[0].Priority != TodoPriorityHigh {
		t.Fatalf("persisted Todos = %#v, want description and high priority", persisted.Todos)
	}
	if len(persisted.TodoProjects) != 2 {
		t.Fatalf("persisted TodoProjects length = %d, want 2", len(persisted.TodoProjects))
	}
}

func TestProjectManagerCreatesTodoWithProjectBaseBranches(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "todo-a", "todo-project-a", "todo-project-b")),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	otherProject, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(other) error = %v", err)
	}

	state, err := manager.CreateTodo(CreateTodoRequest{
		Title: "修复登录问题",
		Projects: []TodoProjectSelection{
			{ProjectID: project.ID, BaseBranch: " main "},
			{ProjectID: otherProject.ID, BaseBranch: "release/2026"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	if len(state.TodoProjects) != 2 {
		t.Fatalf("TodoProjects length = %d, want 2", len(state.TodoProjects))
	}
	if state.TodoProjects[0].ProjectID != project.ID || state.TodoProjects[0].BaseBranch != "main" {
		t.Fatalf("first TodoProject = %#v, want project A with base branch main", state.TodoProjects[0])
	}
	if state.TodoProjects[1].ProjectID != otherProject.ID || state.TodoProjects[1].BaseBranch != "release/2026" {
		t.Fatalf("second TodoProject = %#v, want project B with base branch release/2026", state.TodoProjects[1])
	}

	reloaded := NewProjectManager(configPath)
	persisted, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if persisted.TodoProjects[0].BaseBranch != "main" || persisted.TodoProjects[1].BaseBranch != "release/2026" {
		t.Fatalf("persisted TodoProjects = %#v, want base branches preserved", persisted.TodoProjects)
	}
}

func TestProjectManagerCreatesTodoWithoutProjectUsingMediumPriority(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(sequenceIDs("todo-a")))

	state, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题"})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1", len(state.Todos))
	}
	if state.Todos[0].Priority != TodoPriorityMedium {
		t.Fatalf("Priority = %q, want %q", state.Todos[0].Priority, TodoPriorityMedium)
	}
	if state.Todos[0].Description != "" {
		t.Fatalf("Description = %q, want empty", state.Todos[0].Description)
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0", len(state.TodoProjects))
	}
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" || state.ActiveTerminalID != "" {
		t.Fatalf("active context = %q/%q/%q, want unchanged empty", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveTerminalID)
	}
}

func TestProjectManagerRejectsBlankTodoTitleWithoutChangingState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(sequenceIDs("todo-a")))

	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "   ", Priority: TodoPriorityLow}); err == nil {
		t.Fatal("CreateTodo() error = nil, want title error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Todos) != 0 {
		t.Fatalf("Todos = %#v, want unchanged empty state", state.Todos)
	}
}

func TestProjectManagerLoadsLegacyTodoWithoutPriorityAsMedium(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	legacyJSON := `{
  "version": 1,
  "projects": [],
  "todos": [
    {
      "id": "todo-a",
      "title": "修复登录问题",
      "status": "active",
      "createdAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todoProjects": [],
  "activeTodoId": "todo-a"
}`
	if err := os.WriteFile(configPath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	manager := NewProjectManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1", len(state.Todos))
	}
	if state.Todos[0].Priority != TodoPriorityMedium {
		t.Fatalf("Priority = %q, want %q", state.Todos[0].Priority, TodoPriorityMedium)
	}
}

func TestProjectManagerPersistsTodoWorkspaceAndWorktreeMetadata(t *testing.T) {
	projectDir := t.TempDir()
	worktreeDir := filepath.Join(t.TempDir(), "tasks", "abc123", "frontend-app")
	configPath := filepath.Join(t.TempDir(), "projects.json")
	configJSON := `{
  "version": 1,
  "projects": [
    {
      "id": "project-a",
      "name": "frontend-app",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todos": [
    {
      "id": "todo-a",
      "title": "修复登录问题",
      "description": "登录后跳回首页",
      "priority": "high",
      "status": "in-progress",
      "workspaceDirName": "abc123",
      "createdAt": "2026-06-10T09:00:00Z",
      "startedAt": "2026-06-10T10:00:00Z"
    }
  ],
  "todoProjects": [
    {
      "id": "todo-project-a",
      "todoId": "todo-a",
      "projectId": "project-a",
      "sourceProjectId": "project-a",
      "name": "frontend-app",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "baseBranch": "main",
      "worktreeBranch": "feature/login-fix",
      "worktreePath": "` + filepath.ToSlash(worktreeDir) + `",
      "worktreeStatus": "ready",
      "worktreeError": "",
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "activeTodoId": "todo-a",
  "activeTodoProjectId": "todo-project-a"
}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewProjectManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	todo := findTodo(state.Todos, "todo-a")
	if todo == nil {
		t.Fatal("todo-a not found")
	}
	if todo.WorkspaceDirName != "abc123" {
		t.Fatalf("WorkspaceDirName = %q, want abc123", todo.WorkspaceDirName)
	}
	if len(state.TodoProjects) != 1 {
		t.Fatalf("TodoProjects length = %d, want 1", len(state.TodoProjects))
	}
	todoProject := state.TodoProjects[0]
	if todoProject.BaseBranch != "main" {
		t.Fatalf("BaseBranch = %q, want main", todoProject.BaseBranch)
	}
	if todoProject.WorktreeBranch != "feature/login-fix" {
		t.Fatalf("WorktreeBranch = %q, want feature/login-fix", todoProject.WorktreeBranch)
	}
	if todoProject.WorktreePath != worktreeDir {
		t.Fatalf("WorktreePath = %q, want %q", todoProject.WorktreePath, worktreeDir)
	}
	if todoProject.WorktreeStatus != "ready" {
		t.Fatalf("WorktreeStatus = %q, want ready", todoProject.WorktreeStatus)
	}
	if todoProject.WorktreeError != "" {
		t.Fatalf("WorktreeError = %q, want empty", todoProject.WorktreeError)
	}

	if err := manager.saveLocked(state); err != nil {
		t.Fatalf("saveLocked() error = %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("unmarshal persisted config: %v", err)
	}
	todos := persisted["todos"].([]any)
	if got := todos[0].(map[string]any)["workspaceDirName"]; got != "abc123" {
		t.Fatalf("persisted workspaceDirName = %#v, want abc123", got)
	}
	todoProjects := persisted["todoProjects"].([]any)
	persistedTodoProject := todoProjects[0].(map[string]any)
	if got := persistedTodoProject["baseBranch"]; got != "main" {
		t.Fatalf("persisted baseBranch = %#v, want main", got)
	}
	if got := persistedTodoProject["worktreeBranch"]; got != "feature/login-fix" {
		t.Fatalf("persisted worktreeBranch = %#v, want feature/login-fix", got)
	}
	if got := persistedTodoProject["worktreeStatus"]; got != "ready" {
		t.Fatalf("persisted worktreeStatus = %#v, want ready", got)
	}
}

func TestProjectManagerUpdatesTodoDetailsAndProjectAssociations(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	thirdProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "project-c", "todo-a", "todo-project-a", "todo-project-b", "todo-project-c")),
		WithProjectClock(func() time.Time { return now }),
	)
	projectA, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(A) error = %v", err)
	}
	projectB, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(B) error = %v", err)
	}
	projectC, _, err := manager.AddProjectPath(thirdProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(C) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "旧标题", ProjectIDs: []string{projectA.ID, projectB.ID}}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	now = now.Add(time.Minute)

	state, removedTodoProjectIDs, err := manager.UpdateTodo(UpdateTodoRequest{
		ID:          "todo-a",
		Title:       "  新标题  ",
		Description: "  新描述  ",
		Priority:    TodoPriorityLow,
		ProjectIDs:  []string{projectB.ID, projectC.ID},
	})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}

	if len(removedTodoProjectIDs) != 1 || removedTodoProjectIDs[0] != "todo-project-a" {
		t.Fatalf("removedTodoProjectIDs = %#v, want todo-project-a", removedTodoProjectIDs)
	}
	todo := findTodo(state.Todos, "todo-a")
	if todo == nil {
		t.Fatal("updated todo not found")
	}
	if todo.Title != "新标题" || todo.Description != "新描述" || todo.Priority != TodoPriorityLow {
		t.Fatalf("updated todo = %#v, want trimmed title, description, and low priority", todo)
	}
	if len(state.TodoProjects) != 2 {
		t.Fatalf("TodoProjects length = %d, want 2", len(state.TodoProjects))
	}
	if state.TodoProjects[0].ID != "todo-project-b" || state.TodoProjects[0].ProjectID != projectB.ID {
		t.Fatalf("first TodoProject = %#v, want existing project B association preserved", state.TodoProjects[0])
	}
	if state.TodoProjects[1].ID != "todo-project-c" || state.TodoProjects[1].ProjectID != projectC.ID {
		t.Fatalf("second TodoProject = %#v, want new project C association", state.TodoProjects[1])
	}
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" || state.ActiveProjectID != projectC.ID {
		t.Fatalf("active context = %q/%q/%q, want unchanged project context %q", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, projectC.ID)
	}
}

func TestProjectManagerUpdatesTodoProjectBaseBranches(t *testing.T) {
	projectDir := t.TempDir()
	otherProjectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "todo-a", "todo-project-a", "todo-project-b")),
	)
	projectA, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(A) error = %v", err)
	}
	projectB, _, err := manager.AddProjectPath(otherProjectDir)
	if err != nil {
		t.Fatalf("AddProjectPath(B) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{
		Title:    "旧标题",
		Projects: []TodoProjectSelection{{ProjectID: projectA.ID, BaseBranch: "main"}},
	}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	state, _, err := manager.UpdateTodo(UpdateTodoRequest{
		ID:       "todo-a",
		Title:    "新标题",
		Projects: []TodoProjectSelection{{ProjectID: projectA.ID, BaseBranch: "release"}, {ProjectID: projectB.ID, BaseBranch: "develop"}},
	})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}

	if len(state.TodoProjects) != 2 {
		t.Fatalf("TodoProjects length = %d, want 2", len(state.TodoProjects))
	}
	if state.TodoProjects[0].ID != "todo-project-a" || state.TodoProjects[0].BaseBranch != "release" {
		t.Fatalf("existing TodoProject = %#v, want existing association with updated base branch", state.TodoProjects[0])
	}
	if state.TodoProjects[1].ID != "todo-project-b" || state.TodoProjects[1].BaseBranch != "develop" {
		t.Fatalf("new TodoProject = %#v, want new association with base branch", state.TodoProjects[1])
	}
}

func TestProjectManagerRejectsInvalidTodoUpdateWithoutChangingState(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-a", "todo-project-a")),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{project.ID}}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	cases := []struct {
		name    string
		request UpdateTodoRequest
	}{
		{
			name:    "blank title",
			request: UpdateTodoRequest{ID: "todo-a", Title: "   ", ProjectIDs: []string{project.ID}},
		},
		{
			name:    "missing todo",
			request: UpdateTodoRequest{ID: "missing-todo", Title: "新标题", ProjectIDs: []string{project.ID}},
		},
		{
			name:    "missing project",
			request: UpdateTodoRequest{ID: "todo-a", Title: "新标题", ProjectIDs: []string{"missing-project"}},
		},
		{
			name:    "duplicate project",
			request: UpdateTodoRequest{ID: "todo-a", Title: "新标题", ProjectIDs: []string{project.ID, project.ID}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := manager.UpdateTodo(tc.request); err == nil {
				t.Fatal("UpdateTodo() error = nil, want error")
			}
			state, err := manager.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if state.Todos[0].Title != "修复登录问题" {
				t.Fatalf("Title after invalid update = %q, want unchanged", state.Todos[0].Title)
			}
			if len(state.TodoProjects) != 1 || state.TodoProjects[0].ProjectID != project.ID {
				t.Fatalf("TodoProjects after invalid update = %#v, want unchanged association", state.TodoProjects)
			}
		})
	}
}

func TestProjectManagerRemovesTodoProjectWithoutAffectingOtherTodos(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-a", "todo-project-a", "todo-b", "todo-project-b")),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{project.ID}}); err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "升级依赖", ProjectIDs: []string{project.ID}}); err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}

	state, removedTodoProjectIDs, err := manager.RemoveTodoProject("todo-project-a")
	if err != nil {
		t.Fatalf("RemoveTodoProject() error = %v", err)
	}

	if len(removedTodoProjectIDs) != 1 || removedTodoProjectIDs[0] != "todo-project-a" {
		t.Fatalf("removedTodoProjectIDs = %#v, want todo-project-a", removedTodoProjectIDs)
	}
	if len(state.TodoProjects) != 1 || state.TodoProjects[0].ID != "todo-project-b" {
		t.Fatalf("TodoProjects = %#v, want only todo-project-b", state.TodoProjects)
	}
	if state.ActiveTodoID != "" || state.ActiveTodoProjectID != "" || state.ActiveProjectID != project.ID {
		t.Fatalf("active context = %q/%q/%q, want unchanged project context %q", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveProjectID, project.ID)
	}
}

func TestProjectManagerArchivesTodoWithProjectSnapshots(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-a", "todo-project-a")),
		WithProjectClock(func() time.Time { return now }),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题"}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := manager.AssociateProjectWithTodo("todo-a", project.ID); err != nil {
		t.Fatalf("AssociateProjectWithTodo() error = %v", err)
	}

	if _, err := manager.CompleteTodo("todo-a"); err == nil {
		t.Fatal("CompleteTodo(not-started) error = nil, want invalid status error")
	}
	if _, err := manager.ChangeTodoStatus("todo-a", "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}

	now = now.Add(time.Minute)
	state, err := manager.CompleteTodo("todo-a")
	if err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want 0 after completion", len(state.TodoProjects))
	}
	if len(state.Todos) != 1 {
		t.Fatalf("Todos length = %d, want 1 completed todo", len(state.Todos))
	}
	todo := state.Todos[0]
	if todo.Status != "completed" || todo.ArchivedReason != "" {
		t.Fatalf("todo state = %q/%q, want completed without archive reason", todo.Status, todo.ArchivedReason)
	}
	if todo.CompletedAt == "" || todo.ArchivedAt == "" {
		t.Fatalf("CompletedAt/ArchivedAt = %q/%q, want timestamps", todo.CompletedAt, todo.ArchivedAt)
	}
	if startedAt := todoJSONField(t, todo, "startedAt"); startedAt != "2026-06-10T09:00:00Z" {
		t.Fatalf("StartedAt = %q, want start timestamp preserved after completion", startedAt)
	}
	if len(todo.ProjectSnapshots) != 1 {
		t.Fatalf("ProjectSnapshots length = %d, want 1", len(todo.ProjectSnapshots))
	}
	if snapshot := todo.ProjectSnapshots[0]; snapshot.ProjectID != project.ID || snapshot.Name != filepath.Base(projectDir) || snapshot.Path != projectDir {
		t.Fatalf("ProjectSnapshot = %#v, want archived project details", snapshot)
	}

	state, err = manager.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	if len(state.Todos[0].ProjectSnapshots) != 1 || state.Todos[0].ProjectSnapshots[0].Path != projectDir {
		t.Fatalf("archived snapshots after project delete = %#v, want unchanged", state.Todos[0].ProjectSnapshots)
	}
}

func TestProjectManagerArchivesTodoWithSnapshotBranches(t *testing.T) {
	projectDir := t.TempDir()
	runGitForTest(t, projectDir, "init", "-b", "main")
	runGitForTest(t, projectDir, "checkout", "-b", "todo/fix-login")
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-a", "todo-project-a")),
		WithProjectClock(func() time.Time { return now }),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{
		Title:    "修复登录问题",
		Projects: []TodoProjectSelection{{ProjectID: project.ID, BaseBranch: "main"}},
	}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-a", "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(in-progress) error = %v", err)
	}

	now = now.Add(time.Minute)
	state, err := manager.CompleteTodo("todo-a")
	if err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	todo := state.Todos[0]
	if len(todo.ProjectSnapshots) != 1 {
		t.Fatalf("ProjectSnapshots length = %d, want 1", len(todo.ProjectSnapshots))
	}
	snapshot := todo.ProjectSnapshots[0]
	if snapshot.BaseBranch != "main" || snapshot.WorktreeBranch != "" {
		t.Fatalf("ProjectSnapshot = %#v, want base main before branch fill", snapshot)
	}

	state, err = manager.FillCompletedTodoSnapshotBranches("todo-a")
	if err != nil {
		t.Fatalf("FillCompletedTodoSnapshotBranches() error = %v", err)
	}
	snapshot = state.Todos[0].ProjectSnapshots[0]
	if snapshot.BaseBranch != "main" || snapshot.WorktreeBranch != "todo/fix-login" {
		t.Fatalf("ProjectSnapshot = %#v, want base main and worktree todo/fix-login", snapshot)
	}
}

func TestProjectManagerDeletesTodoFromVisibleState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("todo-a")),
	)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "临时任务"}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	state, err := manager.DeleteTodo("todo-a")
	if err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}

	if len(state.Todos) != 0 {
		t.Fatalf("Todos length = %d, want deleted todo removed", len(state.Todos))
	}
}

func TestProjectManagerDeletesCompletedTodoFromVisibleState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("todo-a")),
	)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "完成任务"}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-a", TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus() error = %v", err)
	}
	if _, err := manager.CompleteTodo("todo-a"); err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	state, err := manager.DeleteTodo("todo-a")
	if err != nil {
		t.Fatalf("DeleteTodo(completed) error = %v", err)
	}

	if len(state.Todos) != 0 {
		t.Fatalf("Todos length = %d, want completed todo removed", len(state.Todos))
	}
	if len(state.TodoProjects) != 0 {
		t.Fatalf("TodoProjects length = %d, want no restored associations", len(state.TodoProjects))
	}
}

func TestProjectManagerDeleteCompletedTodosRemovesOnlyCompletedTodos(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("todo-a", "todo-b")),
	)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "完成任务 A"}); err != nil {
		t.Fatalf("CreateTodo(A) error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-a", TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(A) error = %v", err)
	}
	if _, err := manager.CompleteTodo("todo-a"); err != nil {
		t.Fatalf("CompleteTodo(A) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "完成任务 B"}); err != nil {
		t.Fatalf("CreateTodo(B) error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-b", TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(B) error = %v", err)
	}
	if _, err := manager.CompleteTodo("todo-b"); err != nil {
		t.Fatalf("CompleteTodo(B) error = %v", err)
	}

	state, err := manager.DeleteCompletedTodos([]string{"todo-a", "todo-b"})
	if err != nil {
		t.Fatalf("DeleteCompletedTodos() error = %v", err)
	}

	if len(state.Todos) != 0 {
		t.Fatalf("Todos = %#v, want all completed todos removed", state.Todos)
	}
}

func TestProjectManagerDeleteCompletedTodosRejectsInvalidInputWithoutChangingState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("todo-a", "todo-b")),
	)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "完成任务"}); err != nil {
		t.Fatalf("CreateTodo(completed) error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-a", TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(completed) error = %v", err)
	}
	if _, err := manager.CompleteTodo("todo-a"); err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "执行中任务"}); err != nil {
		t.Fatalf("CreateTodo(open) error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-b", TodoStatusInProgress); err != nil {
		t.Fatalf("ChangeTodoStatus(open) error = %v", err)
	}

	if _, err := manager.DeleteCompletedTodos([]string{" "}); err == nil {
		t.Fatal("DeleteCompletedTodos(empty) error = nil, want error")
	}
	if _, err := manager.DeleteCompletedTodos([]string{"todo-a", "missing-todo"}); err == nil {
		t.Fatal("DeleteCompletedTodos(missing) error = nil, want error")
	}
	if _, err := manager.DeleteCompletedTodos([]string{"todo-a", "todo-b"}); err == nil {
		t.Fatal("DeleteCompletedTodos(open) error = nil, want completed-only error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if findTodo(state.Todos, "todo-a") == nil || findTodo(state.Todos, "todo-b") == nil {
		t.Fatalf("Todos after invalid bulk delete = %#v, want unchanged", state.Todos)
	}
}

func TestProjectManagerImportsProjectsFromParentDirectory(t *testing.T) {
	parentDir := t.TempDir()
	existingDir := filepath.Join(parentDir, "existing")
	newDir := filepath.Join(parentDir, "new-app")
	nonGitDir := filepath.Join(parentDir, "docs")
	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatalf("mkdir existing: %v", err)
	}
	if err := os.Mkdir(filepath.Join(existingDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir existing .git: %v", err)
	}
	if err := os.Mkdir(newDir, 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	if err := os.Mkdir(filepath.Join(newDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir new .git: %v", err)
	}
	if err := os.Mkdir(nonGitDir, 0o755); err != nil {
		t.Fatalf("mkdir non-git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(parentDir, "README.md"), []byte("not a project"), 0o600); err != nil {
		t.Fatalf("write file child: %v", err)
	}
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-existing", "project-new")),
	)
	if _, _, err := manager.AddProjectPath(existingDir); err != nil {
		t.Fatalf("AddProjectPath(existing) error = %v", err)
	}

	state, err := manager.ImportProjectsFromParentDirectory(parentDir)
	if err != nil {
		t.Fatalf("ImportProjectsFromParentDirectory() error = %v", err)
	}

	if len(state.Projects) != 2 {
		t.Fatalf("Projects length = %d, want existing + one new", len(state.Projects))
	}
	if state.ImportSummary == nil {
		t.Fatal("ImportSummary = nil, want summary")
	}
	if state.ImportSummary.AddedCount != 1 || state.ImportSummary.SkippedCount != 2 {
		t.Fatalf("ImportSummary = %#v, want 1 added and 2 skipped", state.ImportSummary)
	}
	if !containsProjectPath(state.Projects, newDir) {
		t.Fatalf("Projects = %#v, want imported %q", state.Projects, newDir)
	}
	if containsProjectPath(state.Projects, nonGitDir) {
		t.Fatalf("Projects = %#v, want non-Git directory %q skipped", state.Projects, nonGitDir)
	}
}

func TestProjectManagerDeleteProjectPreservesActiveTodoCopiesAndArchivedSnapshots(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-active", "todo-project-active", "todo-archived", "todo-project-archived")),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "活动任务"}); err != nil {
		t.Fatalf("CreateTodo(active) error = %v", err)
	}
	if _, err := manager.AssociateProjectWithTodo("todo-active", project.ID); err != nil {
		t.Fatalf("AssociateProjectWithTodo(active) error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "归档任务"}); err != nil {
		t.Fatalf("CreateTodo(archived) error = %v", err)
	}
	if _, err := manager.AssociateProjectWithTodo("todo-archived", project.ID); err != nil {
		t.Fatalf("AssociateProjectWithTodo(archived) error = %v", err)
	}
	if _, err := manager.ChangeTodoStatus("todo-archived", "in-progress"); err != nil {
		t.Fatalf("ChangeTodoStatus(archived) error = %v", err)
	}
	if _, err := manager.CompleteTodo("todo-archived"); err != nil {
		t.Fatalf("CompleteTodo() error = %v", err)
	}

	state, err := manager.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	if len(state.TodoProjects) != 1 || state.TodoProjects[0].ID != "todo-project-active" {
		t.Fatalf("TodoProjects = %#v, want active TODO project copy preserved", state.TodoProjects)
	}
	if state.TodoProjects[0].Path != projectDir {
		t.Fatalf("TodoProject path = %q, want %q", state.TodoProjects[0].Path, projectDir)
	}
	archived := findTodo(state.Todos, "todo-archived")
	if archived == nil || len(archived.ProjectSnapshots) != 1 || archived.ProjectSnapshots[0].Path != projectDir {
		t.Fatalf("archived todo = %#v, want preserved snapshot", archived)
	}
}

func TestProjectManagerMarksMissingPathsUnavailable(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")

	manager := NewProjectManager(configPath, WithProjectIDGenerator(func() string { return "project-1" }))
	if _, _, err := manager.AddProjectPath(projectDir); err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	reloaded := NewProjectManager(configPath)
	state, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Projects) != 1 {
		t.Fatalf("Projects length = %d, want 1", len(state.Projects))
	}
	if state.Projects[0].Available {
		t.Fatalf("Available = true, want false")
	}
}

func TestProjectManagerDeletesProjectWithoutRemovingDirectory(t *testing.T) {
	projectDir := t.TempDir()
	otherDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b")),
		WithProjectClock(func() time.Time { return now }),
	)
	if _, _, err := manager.AddProjectPath(projectDir); err != nil {
		t.Fatalf("AddProjectPath(projectDir) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, err := manager.AddProjectPath(otherDir); err != nil {
		t.Fatalf("AddProjectPath(otherDir) error = %v", err)
	}

	state, err := manager.DeleteProject("project-a")
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	if len(state.Projects) != 1 {
		t.Fatalf("Projects length = %d, want 1", len(state.Projects))
	}
	if state.Projects[0].ID != "project-b" {
		t.Fatalf("remaining project ID = %q, want project-b", state.Projects[0].ID)
	}
	if _, err := os.Stat(projectDir); err != nil {
		t.Fatalf("deleted project directory should remain on disk: %v", err)
	}

	reloaded := NewProjectManager(configPath)
	persisted, err := reloaded.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(persisted.Projects) != 1 || persisted.Projects[0].ID != "project-b" {
		t.Fatalf("persisted projects = %#v, want only project-b", persisted.Projects)
	}
}

func TestProjectManagerDeleteProjectSelectsMostRecentlySelectedRemainingProject(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	dirC := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "project-c")),
		WithProjectClock(func() time.Time { return now }),
	)
	if _, _, err := manager.AddProjectPath(dirA); err != nil {
		t.Fatalf("AddProjectPath(dirA) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, err := manager.AddProjectPath(dirB); err != nil {
		t.Fatalf("AddProjectPath(dirB) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, err := manager.AddProjectPath(dirC); err != nil {
		t.Fatalf("AddProjectPath(dirC) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.SelectProject("project-b"); err != nil {
		t.Fatalf("SelectProject(project-b) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.SelectProject("project-c"); err != nil {
		t.Fatalf("SelectProject(project-c) error = %v", err)
	}

	state, err := manager.DeleteProject("project-c")
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	if state.ActiveProjectID != "project-b" {
		t.Fatalf("ActiveProjectID = %q, want project-b", state.ActiveProjectID)
	}

	state, err = manager.DeleteProject("project-b")
	if err != nil {
		t.Fatalf("DeleteProject(project-b) error = %v", err)
	}
	if state.ActiveProjectID != "project-a" {
		t.Fatalf("ActiveProjectID after deleting project-b = %q, want project-a", state.ActiveProjectID)
	}

	state, err = manager.DeleteProject("project-a")
	if err != nil {
		t.Fatalf("DeleteProject(project-a) error = %v", err)
	}
	if state.ActiveProjectID != "" {
		t.Fatalf("ActiveProjectID after deleting last project = %q, want empty", state.ActiveProjectID)
	}
}

func TestProjectManagerDeleteProjectReturnsErrorWhenProjectIsMissing(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(func() string { return "project-a" }))
	if _, _, err := manager.AddProjectPath(projectDir); err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}

	if _, err := manager.DeleteProject("missing-project"); err == nil {
		t.Fatal("DeleteProject() error = nil, want error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Projects) != 1 || state.Projects[0].ID != "project-a" {
		t.Fatalf("projects after missing delete = %#v, want project-a unchanged", state.Projects)
	}
	if state.ActiveProjectID != "project-a" {
		t.Fatalf("ActiveProjectID = %q, want project-a", state.ActiveProjectID)
	}
}

func TestProjectManagerGlobalCandidateRemovalPreservesTodoProjectCopy(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "workspace", "projects.json")
	globalPath := filepath.Join(t.TempDir(), "global-projects.json")
	manager := NewProjectManager(
		configPath,
		WithGlobalProjectCandidatesPath(globalPath),
		WithProjectIDGenerator(sequenceIDs("project-a", "todo-a", "todo-project-a")),
	)
	project, _, err := manager.AddProjectPath(projectDir)
	if err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{project.ID}}); err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}

	state, err := manager.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	if len(state.Projects) != 0 {
		t.Fatalf("global candidates after delete = %#v, want empty", state.Projects)
	}
	if len(state.TodoProjects) != 1 {
		t.Fatalf("TodoProjects after deleting candidate = %#v, want preserved copy", state.TodoProjects)
	}
	todoProject := state.TodoProjects[0]
	if todoProject.SourceProjectID != project.ID || todoProject.Name != filepath.Base(projectDir) || todoProject.Path != projectDir {
		t.Fatalf("TodoProject copy = %#v, want source/name/path preserved", todoProject)
	}
	if _, _, projectCopy, err := manager.SelectTodoProject(todoProject.ID); err != nil {
		t.Fatalf("SelectTodoProject() after candidate delete error = %v", err)
	} else if projectCopy.Path != projectDir || projectCopy.Name != filepath.Base(projectDir) {
		t.Fatalf("selected project copy = %#v, want todo project path %q", projectCopy, projectDir)
	}
}

func TestProjectManagerMigratesLegacyWorkspaceProjectsToGlobalCandidatesAndTodoCopies(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "workspace", "projects.json")
	globalPath := filepath.Join(t.TempDir(), "global-projects.json")
	legacyJSON := `{
  "version": 1,
  "projects": [
    {
      "id": "project-a",
      "name": "frontend-app",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todos": [
    {
      "id": "todo-a",
      "title": "修复登录问题",
      "priority": "medium",
      "status": "not-started",
      "createdAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todoProjects": [
    {
      "id": "todo-project-a",
      "todoId": "todo-a",
      "projectId": "project-a",
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ]
}`
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}
	manager := NewProjectManager(configPath, WithGlobalProjectCandidatesPath(globalPath))

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].Path != projectDir {
		t.Fatalf("global candidates = %#v, want migrated project path %q", state.Projects, projectDir)
	}
	if len(state.TodoProjects) != 1 {
		t.Fatalf("TodoProjects = %#v, want migrated copy", state.TodoProjects)
	}
	todoProject := state.TodoProjects[0]
	if todoProject.SourceProjectID != "project-a" || todoProject.Name != "frontend-app" || todoProject.Path != projectDir {
		t.Fatalf("migrated TodoProject = %#v, want source/name/path populated", todoProject)
	}

	reloaded := NewProjectManager(configPath, WithGlobalProjectCandidatesPath(globalPath))
	persisted, err := reloaded.Load()
	if err != nil {
		t.Fatalf("reloaded Load() error = %v", err)
	}
	if len(persisted.Projects) != 1 || persisted.Projects[0].Path != projectDir {
		t.Fatalf("persisted global candidates = %#v, want migrated project", persisted.Projects)
	}
	if len(persisted.TodoProjects) != 1 || persisted.TodoProjects[0].Path != projectDir {
		t.Fatalf("persisted TodoProjects = %#v, want copied path", persisted.TodoProjects)
	}
}

func TestProjectManagerLoadsLegacyProjectCopiesAndSnapshotsWithoutBranches(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	legacyJSON := `{
  "version": 1,
  "projects": [
    {
      "id": "project-a",
      "name": "frontend-app",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "todos": [
    {
      "id": "todo-a",
      "title": "修复登录问题",
      "priority": "medium",
      "status": "not-started",
      "createdAt": "2026-06-10T09:00:00Z"
    },
    {
      "id": "todo-completed",
      "title": "已完成任务",
      "priority": "medium",
      "status": "completed",
      "projectSnapshots": [
        { "projectId": "project-a", "name": "frontend-app", "path": "` + filepath.ToSlash(projectDir) + `" }
      ],
      "createdAt": "2026-06-10T09:00:00Z",
      "completedAt": "2026-06-10T10:00:00Z"
    }
  ],
  "todoProjects": [
    {
      "id": "todo-project-a",
      "todoId": "todo-a",
      "projectId": "project-a",
      "name": "frontend-app",
      "path": "` + filepath.ToSlash(projectDir) + `",
      "available": true,
      "createdAt": "2026-06-10T09:00:00Z",
      "lastSelectedAt": "2026-06-10T09:00:00Z"
    }
  ],
  "activeProjectId": "project-a"
}`
	if err := os.WriteFile(configPath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	manager := NewProjectManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(state.TodoProjects) != 1 || state.TodoProjects[0].BaseBranch != "" {
		t.Fatalf("legacy TodoProjects = %#v, want preserved with empty base branch", state.TodoProjects)
	}
	completed := findTodo(state.Todos, "todo-completed")
	if completed == nil || len(completed.ProjectSnapshots) != 1 {
		t.Fatalf("completed todo = %#v, want legacy snapshot", completed)
	}
	snapshot := completed.ProjectSnapshots[0]
	if snapshot.BaseBranch != "" || snapshot.WorktreeBranch != "" {
		t.Fatalf("legacy ProjectSnapshot = %#v, want empty branch fields", snapshot)
	}
}

func TestProjectManagerDeleteProjectsRemovesProjectsAndAssociationsSelectsFallback(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	dirC := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	manager := NewProjectManager(
		configPath,
		WithProjectIDGenerator(sequenceIDs("project-a", "project-b", "project-c", "todo-a", "todo-project-a", "todo-b", "todo-project-b")),
		WithProjectClock(func() time.Time { return now }),
	)
	if _, _, err := manager.AddProjectPath(dirA); err != nil {
		t.Fatalf("AddProjectPath(dirA) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, err := manager.AddProjectPath(dirB); err != nil {
		t.Fatalf("AddProjectPath(dirB) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, err := manager.AddProjectPath(dirC); err != nil {
		t.Fatalf("AddProjectPath(dirC) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "修复登录问题", ProjectIDs: []string{"project-a"}}); err != nil {
		t.Fatalf("CreateTodo(project-a) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.CreateTodo(CreateTodoRequest{Title: "升级依赖", ProjectIDs: []string{"project-b"}}); err != nil {
		t.Fatalf("CreateTodo(project-b) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.SelectProject("project-b"); err != nil {
		t.Fatalf("SelectProject(project-b) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, err := manager.SelectProject("project-c"); err != nil {
		t.Fatalf("SelectProject(project-c) error = %v", err)
	}
	now = now.Add(time.Minute)
	if _, _, _, err := manager.SelectTodoProject("todo-project-a"); err != nil {
		t.Fatalf("SelectTodoProject(todo-project-a) error = %v", err)
	}

	state, err := manager.DeleteProjects([]string{"project-a", "project-c"})
	if err != nil {
		t.Fatalf("DeleteProjects() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].ID != "project-b" {
		t.Fatalf("Projects = %#v, want only project-b", state.Projects)
	}
	if state.ActiveProjectID != "project-b" {
		t.Fatalf("ActiveProjectID = %q, want project-b", state.ActiveProjectID)
	}
	if state.ActiveTodoID != "todo-a" || state.ActiveTodoProjectID != "todo-project-a" || state.ActiveTerminalID != "" {
		t.Fatalf("active TODO/terminal context = %q/%q/%q, want active todo project preserved", state.ActiveTodoID, state.ActiveTodoProjectID, state.ActiveTerminalID)
	}
	if countTodoProjectAssociations(state.TodoProjects, "project-a") != 1 || countTodoProjectAssociations(state.TodoProjects, "project-c") != 0 {
		t.Fatalf("TodoProjects = %#v, want project-a TODO copy preserved and no project-c copy", state.TodoProjects)
	}
	if countTodoProjectAssociations(state.TodoProjects, "project-b") != 1 {
		t.Fatalf("TodoProjects = %#v, want project-b association preserved", state.TodoProjects)
	}
	if _, err := os.Stat(dirA); err != nil {
		t.Fatalf("project A directory should remain on disk: %v", err)
	}
	if _, err := os.Stat(dirC); err != nil {
		t.Fatalf("project C directory should remain on disk: %v", err)
	}
}

func TestProjectManagerDeleteProjectsClearsActiveWhenNoProjectsRemain(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(sequenceIDs("project-a", "project-b")))
	if _, _, err := manager.AddProjectPath(dirA); err != nil {
		t.Fatalf("AddProjectPath(dirA) error = %v", err)
	}
	if _, _, err := manager.AddProjectPath(dirB); err != nil {
		t.Fatalf("AddProjectPath(dirB) error = %v", err)
	}

	state, err := manager.DeleteProjects([]string{"project-a", "project-b"})
	if err != nil {
		t.Fatalf("DeleteProjects() error = %v", err)
	}

	if len(state.Projects) != 0 {
		t.Fatalf("Projects = %#v, want empty", state.Projects)
	}
	if state.ActiveProjectID != "" {
		t.Fatalf("ActiveProjectID = %q, want empty", state.ActiveProjectID)
	}
}

func TestProjectManagerDeleteProjectsNormalizesDuplicateIDs(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(sequenceIDs("project-a", "project-b")))
	if _, _, err := manager.AddProjectPath(dirA); err != nil {
		t.Fatalf("AddProjectPath(dirA) error = %v", err)
	}
	if _, _, err := manager.AddProjectPath(dirB); err != nil {
		t.Fatalf("AddProjectPath(dirB) error = %v", err)
	}

	state, err := manager.DeleteProjects([]string{" project-a ", "project-a"})
	if err != nil {
		t.Fatalf("DeleteProjects() error = %v", err)
	}

	if len(state.Projects) != 1 || state.Projects[0].ID != "project-b" {
		t.Fatalf("Projects = %#v, want only project-b", state.Projects)
	}
}

func TestProjectManagerDeleteProjectsRejectsInvalidInputWithoutChangingState(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "projects.json")
	manager := NewProjectManager(configPath, WithProjectIDGenerator(func() string { return "project-a" }))
	if _, _, err := manager.AddProjectPath(projectDir); err != nil {
		t.Fatalf("AddProjectPath() error = %v", err)
	}

	if _, err := manager.DeleteProjects([]string{" "}); err == nil {
		t.Fatal("DeleteProjects(empty) error = nil, want error")
	}
	if _, err := manager.DeleteProjects([]string{"project-a", "missing-project"}); err == nil {
		t.Fatal("DeleteProjects(missing) error = nil, want error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Projects) != 1 || state.Projects[0].ID != "project-a" {
		t.Fatalf("projects after invalid bulk delete = %#v, want project-a unchanged", state.Projects)
	}
	if state.ActiveProjectID != "project-a" {
		t.Fatalf("ActiveProjectID = %q, want project-a", state.ActiveProjectID)
	}
}

func sequenceIDs(ids ...string) func() string {
	index := 0
	return func() string {
		id := ids[index]
		index++
		return id
	}
}

func countTodoProjectAssociations(todoProjects []TodoProject, projectID string) int {
	count := 0
	for _, todoProject := range todoProjects {
		if todoProject.ProjectID == projectID {
			count++
		}
	}
	return count
}

func containsProjectPath(projects []Project, path string) bool {
	for _, project := range projects {
		if project.Path == path {
			return true
		}
	}
	return false
}

func findTodo(todos []Todo, todoID string) *Todo {
	for index := range todos {
		if todos[index].ID == todoID {
			return &todos[index]
		}
	}
	return nil
}

func assertTodoInitializationFileSnapshots(t *testing.T, got []TodoInitializationFileSnapshot, want []TodoInitializationFileSnapshot) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(InitializationFiles) = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("InitializationFiles[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func runGitForTest(t *testing.T, path string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"-C", path}, args...)
	output, err := newBackgroundCommand(context.Background(), "git", cmdArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func todoJSONField(t *testing.T, todo Todo, field string) string {
	t.Helper()
	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("Marshal(todo) error = %v", err)
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatalf("Unmarshal(todo) error = %v", err)
	}
	value, ok := values[field].(string)
	if !ok {
		return ""
	}
	return value
}
