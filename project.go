package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const projectConfigVersion = 1

const (
	TodoStatusNotStarted = "not-started"
	TodoStatusInProgress = "in-progress"
	TodoStatusCompleted  = "completed"

	TodoStatusActive   = "active"
	TodoStatusArchived = "archived"

	TodoArchiveReasonCompleted = "completed"
	TodoArchiveReasonDeleted   = "deleted"

	TodoPriorityHigh   = "high"
	TodoPriorityMedium = "medium"
	TodoPriorityLow    = "low"
)

type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	Available      bool   `json:"available"`
	CreatedAt      string `json:"createdAt"`
	LastSelectedAt string `json:"lastSelectedAt"`
}

type Todo struct {
	ID               string                `json:"id"`
	Title            string                `json:"title"`
	Description      string                `json:"description,omitempty"`
	Priority         string                `json:"priority"`
	Status           string                `json:"status"`
	ArchivedReason   string                `json:"archivedReason,omitempty"`
	ProjectSnapshots []TodoProjectSnapshot `json:"projectSnapshots,omitempty"`
	CreatedAt        string                `json:"createdAt"`
	StartedAt        string                `json:"startedAt,omitempty"`
	CompletedAt      string                `json:"completedAt,omitempty"`
	ArchivedAt       string                `json:"archivedAt,omitempty"`
}

type CreateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	ProjectIDs  []string `json:"projectIds,omitempty"`
}

type UpdateTodoRequest struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	ProjectIDs  []string `json:"projectIds,omitempty"`
}

type TodoProject struct {
	ID              string `json:"id"`
	TodoID          string `json:"todoId"`
	ProjectID       string `json:"projectId"`
	SourceProjectID string `json:"sourceProjectId,omitempty"`
	Name            string `json:"name,omitempty"`
	Path            string `json:"path,omitempty"`
	Available       bool   `json:"available"`
	CreatedAt       string `json:"createdAt"`
	LastSelectedAt  string `json:"lastSelectedAt"`
}

type TodoProjectSnapshot struct {
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
}

type ProjectImportSummary struct {
	ParentPath   string    `json:"parentPath"`
	AddedCount   int       `json:"addedCount"`
	SkippedCount int       `json:"skippedCount"`
	Added        []Project `json:"added,omitempty"`
	SkippedPaths []string  `json:"skippedPaths,omitempty"`
}

type persistedGlobalProjectCandidates struct {
	Version  int       `json:"version"`
	Projects []Project `json:"projects"`
}

type ProjectState struct {
	Version             int                   `json:"version"`
	CurrentWorkspace    *Workspace            `json:"currentWorkspace,omitempty"`
	RecentWorkspaces    []Workspace           `json:"recentWorkspaces,omitempty"`
	Projects            []Project             `json:"projects"`
	Todos               []Todo                `json:"todos"`
	TodoProjects        []TodoProject         `json:"todoProjects"`
	ActiveProjectID     string                `json:"activeProjectId"`
	ActiveTodoID        string                `json:"activeTodoId,omitempty"`
	ActiveTodoProjectID string                `json:"activeTodoProjectId,omitempty"`
	Terminals           []ProjectTerminal     `json:"terminals,omitempty"`
	ActiveTerminalID    string                `json:"activeTerminalId,omitempty"`
	ImportSummary       *ProjectImportSummary `json:"importSummary,omitempty"`
}

type ProjectManager struct {
	mu                   sync.Mutex
	configPath           string
	globalCandidatesPath string
	newID                func() string
	now                  func() time.Time
}

type ProjectManagerOption func(*ProjectManager)

func NewProjectManager(configPath string, opts ...ProjectManagerOption) *ProjectManager {
	manager := &ProjectManager{
		configPath: configPath,
		newID:      uuid.NewString,
		now:        time.Now,
	}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func WithProjectIDGenerator(newID func() string) ProjectManagerOption {
	return func(manager *ProjectManager) {
		manager.newID = newID
	}
}

func WithProjectClock(now func() time.Time) ProjectManagerOption {
	return func(manager *ProjectManager) {
		manager.now = now
	}
}

func WithGlobalProjectCandidatesPath(path string) ProjectManagerOption {
	return func(manager *ProjectManager) {
		manager.globalCandidatesPath = path
	}
}

func (manager *ProjectManager) Load() (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.loadLocked()
}

func (manager *ProjectManager) candidateConfigPath() string {
	if manager.globalCandidatesPath != "" {
		return manager.globalCandidatesPath
	}
	return manager.configPath
}

func (manager *ProjectManager) usesSeparateGlobalCandidates() bool {
	return manager.globalCandidatesPath != "" && manager.globalCandidatesPath != manager.configPath
}

func (manager *ProjectManager) AddProjectPath(path string) (Project, bool, error) {
	absolutePath, err := normalizeProjectPath(path)
	if err != nil {
		return Project{}, false, err
	}
	if !directoryAvailable(absolutePath) {
		return Project{}, false, errors.New("project path is not an accessible directory")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return Project{}, false, err
	}

	selectedAt := manager.now().UTC().Format(time.RFC3339)
	for index := range state.Projects {
		if state.Projects[index].Path == absolutePath {
			state.Projects[index].Available = true
			state.Projects[index].LastSelectedAt = selectedAt
			state.ActiveProjectID = state.Projects[index].ID
			if err := manager.saveLocked(state); err != nil {
				return Project{}, false, err
			}
			return state.Projects[index], false, nil
		}
	}

	project := Project{
		ID:             manager.newID(),
		Name:           filepath.Base(absolutePath),
		Path:           absolutePath,
		Available:      true,
		CreatedAt:      selectedAt,
		LastSelectedAt: selectedAt,
	}
	state.Projects = append(state.Projects, project)
	state.ActiveProjectID = project.ID
	if err := manager.saveLocked(state); err != nil {
		return Project{}, false, err
	}

	return project, true, nil
}

func (manager *ProjectManager) ImportProjectsFromParentDirectory(parentPath string) (ProjectState, error) {
	absoluteParentPath, err := normalizeProjectPath(parentPath)
	if err != nil {
		return ProjectState{}, err
	}
	if !directoryAvailable(absoluteParentPath) {
		return ProjectState{}, errors.New("parent path is not an accessible directory")
	}
	entries, err := os.ReadDir(absoluteParentPath)
	if err != nil {
		return ProjectState{}, err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}

	summary := &ProjectImportSummary{ParentPath: absoluteParentPath}
	selectedAt := manager.now().UTC().Format(time.RFC3339)
	for _, entry := range entries {
		childPath := filepath.Join(absoluteParentPath, entry.Name())
		if !entry.IsDir() {
			continue
		}
		absoluteChildPath, err := normalizeProjectPath(childPath)
		if err != nil || !directoryAvailable(absoluteChildPath) {
			summary.SkippedCount++
			summary.SkippedPaths = append(summary.SkippedPaths, childPath)
			continue
		}
		if containsProjectAbsolutePath(state.Projects, absoluteChildPath) {
			summary.SkippedCount++
			summary.SkippedPaths = append(summary.SkippedPaths, absoluteChildPath)
			continue
		}
		project := Project{
			ID:             manager.newID(),
			Name:           filepath.Base(absoluteChildPath),
			Path:           absoluteChildPath,
			Available:      true,
			CreatedAt:      selectedAt,
			LastSelectedAt: selectedAt,
		}
		state.Projects = append(state.Projects, project)
		state.ActiveProjectID = project.ID
		summary.AddedCount++
		summary.Added = append(summary.Added, project)
	}
	state.ImportSummary = summary
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) SelectProject(projectID string) (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	for index := range state.Projects {
		if state.Projects[index].ID == projectID {
			state.ActiveProjectID = projectID
			state.Projects[index].LastSelectedAt = manager.now().UTC().Format(time.RFC3339)
			return state, manager.saveLocked(state)
		}
	}

	return ProjectState{}, errors.New("project not found")
}

func (manager *ProjectManager) CreateTodo(request CreateTodoRequest) (ProjectState, error) {
	normalizedTitle := strings.TrimSpace(request.Title)
	if normalizedTitle == "" {
		return ProjectState{}, errors.New("todo title is required")
	}
	normalizedPriority := normalizeTodoPriority(request.Priority)
	normalizedDescription := strings.TrimSpace(request.Description)
	projectIDs := normalizeProjectIDs(request.ProjectIDs)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	if !containsProjects(state.Projects, projectIDs) {
		return ProjectState{}, errors.New("project not found")
	}
	now := manager.now().UTC().Format(time.RFC3339)
	todo := Todo{
		ID:          manager.newID(),
		Title:       normalizedTitle,
		Description: normalizedDescription,
		Priority:    normalizedPriority,
		Status:      TodoStatusNotStarted,
		CreatedAt:   now,
	}
	state.Todos = append(state.Todos, todo)
	for _, projectID := range projectIDs {
		project, _ := projectByIDFromProjects(state.Projects, projectID)
		state.TodoProjects = append(state.TodoProjects, todoProjectFromProject(manager.newID, todo.ID, project, now))
	}
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) AssociateProjectWithTodo(todoID string, projectID string) (ProjectState, error) {
	return manager.AssociateProjectsWithTodo(todoID, []string{projectID})
}

func (manager *ProjectManager) AssociateProjectsWithTodo(todoID string, projectIDs []string) (ProjectState, error) {
	projectIDs = normalizeProjectIDs(projectIDs)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	if _, ok := openTodoByID(state.Todos, todoID); !ok {
		return ProjectState{}, errors.New("todo not found")
	}
	if len(projectIDs) == 0 {
		return state, nil
	}
	if !containsProjects(state.Projects, projectIDs) {
		return ProjectState{}, errors.New("project not found")
	}
	selectedAt := manager.now().UTC().Format(time.RFC3339)
	activeTodoProjectID := ""
	activeProjectID := ""
	for _, projectID := range projectIDs {
		project, _ := projectByIDFromProjects(state.Projects, projectID)
		found := false
		for index := range state.TodoProjects {
			if state.TodoProjects[index].TodoID == todoID && sameTodoProjectPath(state.TodoProjects[index], project) {
				state.TodoProjects[index].LastSelectedAt = selectedAt
				if activeTodoProjectID == "" {
					activeTodoProjectID = state.TodoProjects[index].ID
					activeProjectID = state.TodoProjects[index].ProjectID
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		todoProject := todoProjectFromProject(manager.newID, todoID, project, selectedAt)
		state.TodoProjects = append(state.TodoProjects, todoProject)
		if activeTodoProjectID == "" {
			activeTodoProjectID = todoProject.ID
			activeProjectID = todoProject.ProjectID
		}
	}
	state.ActiveTodoID = todoID
	state.ActiveTodoProjectID = activeTodoProjectID
	state.ActiveProjectID = activeProjectID
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) UpdateTodo(request UpdateTodoRequest) (ProjectState, []string, error) {
	normalizedTitle := strings.TrimSpace(request.Title)
	if normalizedTitle == "" {
		return ProjectState{}, nil, errors.New("todo title is required")
	}
	normalizedDescription := strings.TrimSpace(request.Description)
	normalizedPriority := normalizeTodoPriority(request.Priority)
	projectIDs, err := normalizeProjectIDsForUpdate(request.ProjectIDs)
	if err != nil {
		return ProjectState{}, nil, err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, nil, err
	}
	todoIndex := -1
	for index := range state.Todos {
		if state.Todos[index].ID == request.ID && isOpenTodoStatus(state.Todos[index].Status) {
			todoIndex = index
			break
		}
	}
	if todoIndex < 0 {
		return ProjectState{}, nil, errors.New("todo not found")
	}
	if !containsProjects(state.Projects, projectIDs) {
		return ProjectState{}, nil, errors.New("project not found")
	}

	state.Todos[todoIndex].Title = normalizedTitle
	state.Todos[todoIndex].Description = normalizedDescription
	state.Todos[todoIndex].Priority = normalizedPriority
	removedTodoProjectIDs := []string{}
	existingByProjectID := map[string]TodoProject{}
	for _, todoProject := range state.TodoProjects {
		if todoProject.TodoID == request.ID {
			existingByProjectID[todoProject.ProjectID] = todoProject
		}
	}
	requestedProjectIDs := map[string]bool{}
	for _, projectID := range projectIDs {
		requestedProjectIDs[projectID] = true
	}
	for _, todoProject := range state.TodoProjects {
		if todoProject.TodoID == request.ID && !requestedProjectIDs[todoProject.ProjectID] {
			removedTodoProjectIDs = append(removedTodoProjectIDs, todoProject.ID)
		}
	}

	now := manager.now().UTC().Format(time.RFC3339)
	updatedTodoProjects := make([]TodoProject, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		project, _ := projectByIDFromProjects(state.Projects, projectID)
		if todoProject, ok := existingTodoProjectByPath(existingByProjectID, project); ok {
			updatedTodoProjects = append(updatedTodoProjects, todoProject)
			continue
		}
		updatedTodoProjects = append(updatedTodoProjects, todoProjectFromProject(manager.newID, request.ID, project, now))
	}
	state.TodoProjects = replaceTodoProjectsForTodo(state.TodoProjects, request.ID, updatedTodoProjects)
	updateActiveContextAfterTodoProjectRemoval(&state, request.ID, removedTodoProjectIDs, updatedTodoProjects)
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, nil, err
	}
	return state, removedTodoProjectIDs, nil
}

func (manager *ProjectManager) RemoveTodoProject(todoProjectID string) (ProjectState, []string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, nil, err
	}
	removedTodoProject := TodoProject{}
	found := false
	nextTodoProjects := make([]TodoProject, 0, len(state.TodoProjects))
	for _, todoProject := range state.TodoProjects {
		if todoProject.ID == todoProjectID {
			removedTodoProject = todoProject
			found = true
			continue
		}
		nextTodoProjects = append(nextTodoProjects, todoProject)
	}
	if !found {
		return ProjectState{}, nil, errors.New("todo project not found")
	}
	if _, ok := openTodoByID(state.Todos, removedTodoProject.TodoID); !ok {
		return ProjectState{}, nil, errors.New("todo not found")
	}
	state.TodoProjects = nextTodoProjects
	removedTodoProjectIDs := []string{todoProjectID}
	updateActiveContextAfterTodoProjectRemoval(&state, removedTodoProject.TodoID, removedTodoProjectIDs, nil)
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, nil, err
	}
	return state, removedTodoProjectIDs, nil
}

func (manager *ProjectManager) SelectTodoProject(todoProjectID string) (ProjectState, TodoProject, Project, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, TodoProject{}, Project{}, err
	}
	for index := range state.TodoProjects {
		if state.TodoProjects[index].ID != todoProjectID {
			continue
		}
		if _, ok := openTodoByID(state.Todos, state.TodoProjects[index].TodoID); !ok {
			return ProjectState{}, TodoProject{}, Project{}, errors.New("todo not found")
		}
		project, ok := projectFromTodoProject(state.TodoProjects[index])
		if !ok {
			project, ok = projectByIDFromProjects(state.Projects, state.TodoProjects[index].ProjectID)
			if !ok {
				return ProjectState{}, TodoProject{}, Project{}, errors.New("project not found")
			}
		}
		state.TodoProjects[index].LastSelectedAt = manager.now().UTC().Format(time.RFC3339)
		state.ActiveTodoID = state.TodoProjects[index].TodoID
		state.ActiveTodoProjectID = todoProjectID
		state.ActiveProjectID = state.TodoProjects[index].ProjectID
		if err := manager.saveLocked(state); err != nil {
			return ProjectState{}, TodoProject{}, Project{}, err
		}
		return state, state.TodoProjects[index], project, nil
	}
	return ProjectState{}, TodoProject{}, Project{}, errors.New("todo project not found")
}

func (manager *ProjectManager) CompleteTodo(todoID string) (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	now := manager.now().UTC().Format(time.RFC3339)
	todoIndex := -1
	for index := range state.Todos {
		if state.Todos[index].ID == todoID && state.Todos[index].Status == TodoStatusInProgress {
			todoIndex = index
			break
		}
	}
	if todoIndex < 0 {
		return ProjectState{}, errors.New("todo not found")
	}
	snapshots := todoProjectSnapshots(state.Projects, state.TodoProjects, todoID)
	state.Todos[todoIndex].Status = TodoStatusCompleted
	state.Todos[todoIndex].ArchivedReason = ""
	state.Todos[todoIndex].ArchivedAt = now
	state.Todos[todoIndex].CompletedAt = now
	state.Todos[todoIndex].ProjectSnapshots = snapshots
	state.TodoProjects = removeTodoProjectsForTodo(state.TodoProjects, todoID)
	clearActiveTodoContext(&state, todoID)
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) DeleteTodo(todoID string) (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	todoIndex := -1
	for index := range state.Todos {
		if state.Todos[index].ID == todoID && isDeletableTodoStatus(state.Todos[index].Status) {
			todoIndex = index
			break
		}
	}
	if todoIndex < 0 {
		return ProjectState{}, errors.New("todo not found")
	}
	state.Todos = append(state.Todos[:todoIndex], state.Todos[todoIndex+1:]...)
	state.TodoProjects = removeTodoProjectsForTodo(state.TodoProjects, todoID)
	clearActiveTodoContext(&state, todoID)
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) DeleteCompletedTodos(todoIDs []string) (ProjectState, error) {
	normalizedTodoIDs := normalizeTodoIDs(todoIDs)
	if len(normalizedTodoIDs) == 0 {
		return ProjectState{}, errors.New("todo not found")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	todosByID := map[string]Todo{}
	for _, todo := range state.Todos {
		todosByID[todo.ID] = todo
	}
	deletedTodoIDs := map[string]bool{}
	for _, todoID := range normalizedTodoIDs {
		todo, ok := todosByID[todoID]
		if !ok {
			return ProjectState{}, errors.New("todo not found")
		}
		if todo.Status != TodoStatusCompleted {
			return ProjectState{}, errors.New("todo is not completed")
		}
		deletedTodoIDs[todoID] = true
	}

	nextTodos := make([]Todo, 0, len(state.Todos)-len(deletedTodoIDs))
	for _, todo := range state.Todos {
		if !deletedTodoIDs[todo.ID] {
			nextTodos = append(nextTodos, todo)
		}
	}
	state.Todos = nextTodos
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) ChangeTodoStatus(todoID string, status string) (ProjectState, error) {
	status = strings.TrimSpace(status)
	if status != TodoStatusInProgress {
		return ProjectState{}, errors.New("invalid todo status")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	for index := range state.Todos {
		if state.Todos[index].ID == todoID {
			if state.Todos[index].Status != TodoStatusNotStarted {
				return ProjectState{}, errors.New("invalid todo status transition")
			}
			state.Todos[index].Status = status
			state.Todos[index].StartedAt = manager.now().UTC().Format(time.RFC3339)
			if err := manager.saveLocked(state); err != nil {
				return ProjectState{}, err
			}
			return state, nil
		}
	}
	return ProjectState{}, errors.New("todo not found")
}

func (manager *ProjectManager) DeleteProject(projectID string) (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}

	nextProjects := make([]Project, 0, len(state.Projects))
	deleted := false
	for _, project := range state.Projects {
		if project.ID == projectID {
			deleted = true
			continue
		}
		nextProjects = append(nextProjects, project)
	}
	if !deleted {
		return ProjectState{}, errors.New("project not found")
	}

	state.Projects = nextProjects
	if state.ActiveProjectID == projectID {
		state.ActiveProjectID = mostRecentlySelectedProjectID(state.Projects)
	}
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) DeleteProjects(projectIDs []string) (ProjectState, error) {
	normalizedProjectIDs := normalizeProjectIDs(projectIDs)
	if len(normalizedProjectIDs) == 0 {
		return ProjectState{}, errors.New("project not found")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return ProjectState{}, err
	}
	if !containsProjects(state.Projects, normalizedProjectIDs) {
		return ProjectState{}, errors.New("project not found")
	}

	deletedProjectIDs := map[string]bool{}
	for _, projectID := range normalizedProjectIDs {
		deletedProjectIDs[projectID] = true
	}

	nextProjects := make([]Project, 0, len(state.Projects)-len(deletedProjectIDs))
	for _, project := range state.Projects {
		if !deletedProjectIDs[project.ID] {
			nextProjects = append(nextProjects, project)
		}
	}

	state.Projects = nextProjects
	if deletedProjectIDs[state.ActiveProjectID] {
		state.ActiveProjectID = mostRecentlySelectedProjectID(state.Projects)
	}
	if err := manager.saveLocked(state); err != nil {
		return ProjectState{}, err
	}
	return state, nil
}

func (manager *ProjectManager) GetProject(projectID string) (Project, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return Project{}, err
	}
	for _, project := range state.Projects {
		if project.ID == projectID {
			return project, nil
		}
	}
	for _, todoProject := range state.TodoProjects {
		if todoProject.ProjectID != projectID && todoProject.SourceProjectID != projectID {
			continue
		}
		if project, ok := projectFromTodoProject(todoProject); ok {
			return project, nil
		}
	}
	return Project{}, errors.New("project not found")
}

func (manager *ProjectManager) TodoProject(todoProjectID string) (TodoProject, Project, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	state, err := manager.loadLocked()
	if err != nil {
		return TodoProject{}, Project{}, err
	}
	for _, todoProject := range state.TodoProjects {
		if todoProject.ID == todoProjectID {
			project, ok := projectFromTodoProject(todoProject)
			if !ok {
				project, ok = projectByIDFromProjects(state.Projects, todoProject.ProjectID)
				if !ok {
					return TodoProject{}, Project{}, errors.New("project not found")
				}
			}
			return todoProject, project, nil
		}
	}
	return TodoProject{}, Project{}, errors.New("todo project not found")
}

func (manager *ProjectManager) loadLocked() (ProjectState, error) {
	state := ProjectState{
		Version:      projectConfigVersion,
		Projects:     []Project{},
		Todos:        []Todo{},
		TodoProjects: []TodoProject{},
	}

	data, err := os.ReadFile(manager.configPath)
	if errors.Is(err, os.ErrNotExist) {
		data = nil
	} else if err != nil {
		return ProjectState{}, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &state); err != nil {
			return ProjectState{}, err
		}
	}
	if state.Version == 0 {
		state.Version = projectConfigVersion
	}
	if state.Projects == nil {
		state.Projects = []Project{}
	}
	if state.Todos == nil {
		state.Todos = []Todo{}
	}
	for index := range state.Todos {
		state.Todos[index].Priority = normalizeTodoPriority(state.Todos[index].Priority)
		state.Todos[index] = normalizeTodoForWorkflow(state.Todos[index])
	}
	state.Todos = filterVisibleTodos(state.Todos)
	if state.TodoProjects == nil {
		state.TodoProjects = []TodoProject{}
	}
	state.TodoProjects = filterTodoProjectsForOpenTodos(state.Todos, state.TodoProjects)
	legacyProjects := append([]Project{}, state.Projects...)
	if manager.usesSeparateGlobalCandidates() {
		globalProjects, err := manager.loadGlobalCandidatesLocked()
		if err != nil {
			return ProjectState{}, err
		}
		mergedProjects, changed := mergeProjectsByPath(globalProjects, legacyProjects)
		state.Projects = mergedProjects
		if changed {
			if err := manager.saveGlobalCandidatesLocked(state.Projects); err != nil {
				return ProjectState{}, err
			}
		}
	}
	for index := range state.Projects {
		state.Projects[index].Available = directoryAvailable(state.Projects[index].Path)
	}
	migratedTodoProjects := false
	for index := range state.TodoProjects {
		normalized := normalizeTodoProjectCopy(state.TodoProjects[index], state.Projects)
		if normalized != state.TodoProjects[index] {
			migratedTodoProjects = true
		}
		state.TodoProjects[index] = normalized
	}
	if manager.usesSeparateGlobalCandidates() && (len(legacyProjects) > 0 || migratedTodoProjects) {
		persistedState := state
		persistedState.ImportSummary = nil
		persistedState.Projects = []Project{}
		if err := manager.saveWorkspaceStateLocked(persistedState); err != nil {
			return ProjectState{}, err
		}
	}
	if state.ActiveProjectID != "" && !containsProject(state.Projects, state.ActiveProjectID) {
		state.ActiveProjectID = ""
	}
	if state.ActiveTodoID != "" && !containsActiveTodo(state.Todos, state.ActiveTodoID) {
		state.ActiveTodoID = ""
		state.ActiveTodoProjectID = ""
		state.ActiveTerminalID = ""
	}
	if state.ActiveTodoProjectID != "" && !containsTodoProject(state.TodoProjects, state.ActiveTodoProjectID) {
		state.ActiveTodoProjectID = ""
	}

	return state, nil
}

func (manager *ProjectManager) saveLocked(state ProjectState) error {
	state.Version = projectConfigVersion
	if manager.usesSeparateGlobalCandidates() {
		if err := manager.saveGlobalCandidatesLocked(state.Projects); err != nil {
			return err
		}
	}
	persistedState := state
	persistedState.ImportSummary = nil
	if manager.usesSeparateGlobalCandidates() {
		persistedState.Projects = []Project{}
		persistedState.ActiveProjectID = activeProjectIDFromTodoProjects(persistedState)
	}

	return manager.saveWorkspaceStateLocked(persistedState)
}

func (manager *ProjectManager) saveWorkspaceStateLocked(state ProjectState) error {
	if err := os.MkdirAll(filepath.Dir(manager.configPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tempPath := manager.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, manager.configPath)
}

func (manager *ProjectManager) loadGlobalCandidatesLocked() ([]Project, error) {
	state := persistedGlobalProjectCandidates{
		Version:  projectConfigVersion,
		Projects: []Project{},
	}
	data, err := os.ReadFile(manager.candidateConfigPath())
	if errors.Is(err, os.ErrNotExist) {
		return state.Projects, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return state.Projects, nil
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	if state.Projects == nil {
		state.Projects = []Project{}
	}
	for index := range state.Projects {
		state.Projects[index].Available = directoryAvailable(state.Projects[index].Path)
	}
	return state.Projects, nil
}

func (manager *ProjectManager) saveGlobalCandidatesLocked(projects []Project) error {
	state := persistedGlobalProjectCandidates{
		Version:  projectConfigVersion,
		Projects: append([]Project{}, projects...),
	}
	if err := os.MkdirAll(filepath.Dir(manager.candidateConfigPath()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tempPath := manager.candidateConfigPath() + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, manager.candidateConfigPath())
}

func mergeProjectsByPath(globalProjects []Project, legacyProjects []Project) ([]Project, bool) {
	merged := append([]Project{}, globalProjects...)
	changed := false
	for _, legacyProject := range legacyProjects {
		if legacyProject.Path == "" {
			continue
		}
		if _, ok := projectByPathFromProjects(merged, legacyProject.Path); ok {
			continue
		}
		legacyProject.Available = directoryAvailable(legacyProject.Path)
		merged = append(merged, legacyProject)
		changed = true
	}
	return merged, changed
}

func activeProjectIDFromTodoProjects(state ProjectState) string {
	if state.ActiveTodoProjectID == "" {
		return state.ActiveProjectID
	}
	for _, todoProject := range state.TodoProjects {
		if todoProject.ID == state.ActiveTodoProjectID {
			return todoProject.ProjectID
		}
	}
	return state.ActiveProjectID
}

func normalizeProjectPath(path string) (string, error) {
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return absolutePath, nil
}

func normalizeTodoPriority(priority string) string {
	switch strings.TrimSpace(priority) {
	case TodoPriorityHigh:
		return TodoPriorityHigh
	case TodoPriorityLow:
		return TodoPriorityLow
	default:
		return TodoPriorityMedium
	}
}

func normalizeTodoWorkflowStatus(status string) string {
	switch strings.TrimSpace(status) {
	case TodoStatusInProgress:
		return TodoStatusInProgress
	case TodoStatusCompleted:
		return TodoStatusCompleted
	case TodoStatusArchived:
		return TodoStatusArchived
	case TodoStatusActive, TodoStatusNotStarted:
		return TodoStatusNotStarted
	default:
		return TodoStatusNotStarted
	}
}

func normalizeTodoForWorkflow(todo Todo) Todo {
	switch todo.Status {
	case TodoStatusActive:
		todo.Status = TodoStatusNotStarted
	case TodoStatusArchived:
		if todo.ArchivedReason == TodoArchiveReasonCompleted {
			todo.Status = TodoStatusCompleted
			todo.ArchivedReason = ""
			if todo.CompletedAt == "" {
				todo.CompletedAt = todo.ArchivedAt
			}
			break
		}
		if todo.ArchivedReason == TodoArchiveReasonDeleted {
			return Todo{ID: todo.ID, Status: TodoArchiveReasonDeleted}
		}
		todo.Status = TodoStatusCompleted
		todo.ArchivedReason = ""
	case TodoStatusInProgress, TodoStatusCompleted, TodoStatusNotStarted:
	default:
		todo.Status = TodoStatusNotStarted
	}
	return todo
}

func filterVisibleTodos(todos []Todo) []Todo {
	filtered := make([]Todo, 0, len(todos))
	for _, todo := range todos {
		if todo.Status == TodoArchiveReasonDeleted {
			continue
		}
		filtered = append(filtered, todo)
	}
	return filtered
}

func filterTodoProjectsForOpenTodos(todos []Todo, todoProjects []TodoProject) []TodoProject {
	openTodoIDs := map[string]bool{}
	for _, todo := range todos {
		if isOpenTodoStatus(todo.Status) {
			openTodoIDs[todo.ID] = true
		}
	}
	filtered := make([]TodoProject, 0, len(todoProjects))
	for _, todoProject := range todoProjects {
		if openTodoIDs[todoProject.TodoID] {
			filtered = append(filtered, todoProject)
		}
	}
	return filtered
}

func isOpenTodoStatus(status string) bool {
	return status == TodoStatusNotStarted || status == TodoStatusInProgress
}

func isDeletableTodoStatus(status string) bool {
	return isOpenTodoStatus(status) || status == TodoStatusCompleted
}

func directoryAvailable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func containsProject(projects []Project, projectID string) bool {
	for _, project := range projects {
		if project.ID == projectID {
			return true
		}
	}
	return false
}

func containsProjects(projects []Project, projectIDs []string) bool {
	for _, projectID := range projectIDs {
		if !containsProject(projects, projectID) {
			return false
		}
	}
	return true
}

func containsProjectAbsolutePath(projects []Project, path string) bool {
	for _, project := range projects {
		if project.Path == path {
			return true
		}
	}
	return false
}

func projectByPathFromProjects(projects []Project, path string) (Project, bool) {
	for _, project := range projects {
		if project.Path == path {
			return project, true
		}
	}
	return Project{}, false
}

func containsActiveTodo(todos []Todo, todoID string) bool {
	_, ok := openTodoByID(todos, todoID)
	return ok
}

func activeTodoByID(todos []Todo, todoID string) (Todo, bool) {
	return openTodoByID(todos, todoID)
}

func openTodoByID(todos []Todo, todoID string) (Todo, bool) {
	for _, todo := range todos {
		if todo.ID == todoID && isOpenTodoStatus(todo.Status) {
			return todo, true
		}
	}
	return Todo{}, false
}

func clearActiveTodoContext(state *ProjectState, todoID string) {
	if state.ActiveTodoID == todoID {
		state.ActiveTodoID = ""
		state.ActiveTodoProjectID = ""
		state.ActiveTerminalID = ""
	}
}

func containsTodoProject(todoProjects []TodoProject, todoProjectID string) bool {
	for _, todoProject := range todoProjects {
		if todoProject.ID == todoProjectID {
			return true
		}
	}
	return false
}

func projectByIDFromProjects(projects []Project, projectID string) (Project, bool) {
	for _, project := range projects {
		if project.ID == projectID {
			return project, true
		}
	}
	return Project{}, false
}

func projectFromTodoProject(todoProject TodoProject) (Project, bool) {
	if todoProject.Path == "" {
		return Project{}, false
	}
	projectID := todoProject.ID
	if todoProject.ProjectID != "" {
		projectID = todoProject.ProjectID
	}
	project := Project{
		ID:             projectID,
		Name:           todoProject.Name,
		Path:           todoProject.Path,
		Available:      directoryAvailable(todoProject.Path),
		CreatedAt:      todoProject.CreatedAt,
		LastSelectedAt: todoProject.LastSelectedAt,
	}
	if project.Name == "" {
		project.Name = filepath.Base(todoProject.Path)
	}
	return project, true
}

func todoProjectFromProject(newID func() string, todoID string, project Project, now string) TodoProject {
	return TodoProject{
		ID:              newID(),
		TodoID:          todoID,
		ProjectID:       project.ID,
		SourceProjectID: project.ID,
		Name:            project.Name,
		Path:            project.Path,
		Available:       directoryAvailable(project.Path),
		CreatedAt:       now,
		LastSelectedAt:  now,
	}
}

func normalizeTodoProjectCopy(todoProject TodoProject, projects []Project) TodoProject {
	project, ok := projectByIDFromProjects(projects, todoProject.ProjectID)
	if todoProject.SourceProjectID == "" {
		todoProject.SourceProjectID = todoProject.ProjectID
	}
	if todoProject.Name == "" && ok {
		todoProject.Name = project.Name
	}
	if todoProject.Path == "" && ok {
		todoProject.Path = project.Path
	}
	if todoProject.Path != "" {
		todoProject.Available = directoryAvailable(todoProject.Path)
	}
	return todoProject
}

func removeTodoProjectsForTodo(todoProjects []TodoProject, todoID string) []TodoProject {
	next := make([]TodoProject, 0, len(todoProjects))
	for _, todoProject := range todoProjects {
		if todoProject.TodoID != todoID {
			next = append(next, todoProject)
		}
	}
	return next
}

func removeTodoProjectsForProject(todoProjects []TodoProject, projectID string) []TodoProject {
	next := make([]TodoProject, 0, len(todoProjects))
	for _, todoProject := range todoProjects {
		if todoProject.ProjectID != projectID {
			next = append(next, todoProject)
		}
	}
	return next
}

func sameTodoProjectPath(todoProject TodoProject, project Project) bool {
	if todoProject.Path != "" && project.Path != "" {
		return todoProject.Path == project.Path
	}
	return todoProject.ProjectID == project.ID
}

func existingTodoProjectByPath(existingByProjectID map[string]TodoProject, project Project) (TodoProject, bool) {
	if todoProject, ok := existingByProjectID[project.ID]; ok {
		return todoProject, true
	}
	for _, todoProject := range existingByProjectID {
		if sameTodoProjectPath(todoProject, project) {
			return todoProject, true
		}
	}
	return TodoProject{}, false
}

func replaceTodoProjectsForTodo(todoProjects []TodoProject, todoID string, replacement []TodoProject) []TodoProject {
	next := make([]TodoProject, 0, len(todoProjects)-countTodoProjectsForTodo(todoProjects, todoID)+len(replacement))
	inserted := false
	for _, todoProject := range todoProjects {
		if todoProject.TodoID != todoID {
			next = append(next, todoProject)
			continue
		}
		if !inserted {
			next = append(next, replacement...)
			inserted = true
		}
	}
	if !inserted {
		next = append(next, replacement...)
	}
	return next
}

func countTodoProjectsForTodo(todoProjects []TodoProject, todoID string) int {
	count := 0
	for _, todoProject := range todoProjects {
		if todoProject.TodoID == todoID {
			count++
		}
	}
	return count
}

func updateActiveContextAfterTodoProjectRemoval(state *ProjectState, todoID string, removedTodoProjectIDs []string, preferred []TodoProject) {
	removed := map[string]bool{}
	for _, todoProjectID := range removedTodoProjectIDs {
		removed[todoProjectID] = true
	}
	if !removed[state.ActiveTodoProjectID] {
		return
	}
	state.ActiveTodoID = todoID
	state.ActiveTodoProjectID = ""
	state.ActiveProjectID = ""
	state.ActiveTerminalID = ""
	if len(preferred) > 0 {
		state.ActiveTodoProjectID = preferred[0].ID
		state.ActiveProjectID = preferred[0].ProjectID
		return
	}
	if replacement, ok := mostRecentlySelectedTodoProjectForTodo(state.TodoProjects, todoID); ok {
		state.ActiveTodoProjectID = replacement.ID
		state.ActiveProjectID = replacement.ProjectID
	}
}

func mostRecentlySelectedTodoProjectForTodo(todoProjects []TodoProject, todoID string) (TodoProject, bool) {
	selected := TodoProject{}
	for _, todoProject := range todoProjects {
		if todoProject.TodoID != todoID {
			continue
		}
		if selected.ID == "" || todoProject.LastSelectedAt > selected.LastSelectedAt ||
			(todoProject.LastSelectedAt == selected.LastSelectedAt && todoProject.ID < selected.ID) {
			selected = todoProject
		}
	}
	return selected, selected.ID != ""
}

func todoProjectSnapshots(projects []Project, todoProjects []TodoProject, todoID string) []TodoProjectSnapshot {
	snapshots := []TodoProjectSnapshot{}
	for _, todoProject := range todoProjects {
		if todoProject.TodoID != todoID {
			continue
		}
		project, ok := projectFromTodoProject(todoProject)
		if !ok {
			project, ok = projectByIDFromProjects(projects, todoProject.ProjectID)
			if !ok {
				continue
			}
		}
		snapshots = append(snapshots, TodoProjectSnapshot{
			ProjectID: todoProject.SourceProjectID,
			Name:      project.Name,
			Path:      project.Path,
		})
	}
	return snapshots
}

func mostRecentlySelectedProjectID(projects []Project) string {
	selectedProjectID := ""
	selectedAt := ""
	for _, project := range projects {
		if selectedProjectID == "" || project.LastSelectedAt > selectedAt {
			selectedProjectID = project.ID
			selectedAt = project.LastSelectedAt
		}
	}
	return selectedProjectID
}

func normalizeProjectIDs(projectIDs []string) []string {
	normalized := []string{}
	seen := map[string]bool{}
	for _, projectID := range projectIDs {
		projectID = strings.TrimSpace(projectID)
		if projectID == "" || seen[projectID] {
			continue
		}
		seen[projectID] = true
		normalized = append(normalized, projectID)
	}
	return normalized
}

func normalizeTodoIDs(todoIDs []string) []string {
	normalized := []string{}
	seen := map[string]bool{}
	for _, todoID := range todoIDs {
		todoID = strings.TrimSpace(todoID)
		if todoID == "" || seen[todoID] {
			continue
		}
		seen[todoID] = true
		normalized = append(normalized, todoID)
	}
	return normalized
}

func normalizeProjectIDsForUpdate(projectIDs []string) ([]string, error) {
	normalized := []string{}
	seen := map[string]bool{}
	for _, projectID := range projectIDs {
		projectID = strings.TrimSpace(projectID)
		if projectID == "" {
			continue
		}
		if seen[projectID] {
			return nil, errors.New("duplicate project")
		}
		seen[projectID] = true
		normalized = append(normalized, projectID)
	}
	return normalized, nil
}
