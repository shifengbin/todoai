package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const projectConfigVersion = 1

type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	Available      bool   `json:"available"`
	CreatedAt      string `json:"createdAt"`
	LastSelectedAt string `json:"lastSelectedAt"`
}

type ProjectState struct {
	Version         int       `json:"version"`
	Projects        []Project `json:"projects"`
	ActiveProjectID string    `json:"activeProjectId"`
}

type ProjectManager struct {
	mu         sync.Mutex
	configPath string
	newID      func() string
	now        func() time.Time
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

func (manager *ProjectManager) Load() (ProjectState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.loadLocked()
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
	return Project{}, errors.New("project not found")
}

func (manager *ProjectManager) loadLocked() (ProjectState, error) {
	state := ProjectState{
		Version: projectConfigVersion,
	}

	data, err := os.ReadFile(manager.configPath)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
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
	for index := range state.Projects {
		state.Projects[index].Available = directoryAvailable(state.Projects[index].Path)
	}
	if state.ActiveProjectID != "" && !containsProject(state.Projects, state.ActiveProjectID) {
		state.ActiveProjectID = ""
	}

	return state, nil
}

func (manager *ProjectManager) saveLocked(state ProjectState) error {
	state.Version = projectConfigVersion

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

func normalizeProjectPath(path string) (string, error) {
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return absolutePath, nil
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
