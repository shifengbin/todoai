package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	workspaceDataDirName      = ".data"
	recentWorkspacesFileName  = "recent-workspaces.json"
	legacyWorkspaceDirName    = "legacy-workspace"
	workspaceStateFileVersion = 1
	legacyGlobalProjectsFile  = "projects.json"
	legacyGlobalHistoryFile   = "terminal-history.json"
)

type Workspace struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	DataPath     string `json:"dataPath"`
	Available    bool   `json:"available"`
	LastOpenedAt string `json:"lastOpenedAt"`
}

type WorkspaceState struct {
	Version          int         `json:"version"`
	CurrentWorkspace *Workspace  `json:"currentWorkspace,omitempty"`
	RecentWorkspaces []Workspace `json:"recentWorkspaces"`
}

type WorkspaceManager struct {
	mu           sync.Mutex
	appConfigDir string
	now          func() time.Time
	current      *Workspace
}

type WorkspaceManagerOption func(*WorkspaceManager)

type persistedWorkspaceState struct {
	Version                  int         `json:"version"`
	RecentWorkspaces         []Workspace `json:"recentWorkspaces"`
	LegacyGlobalDataMigrated bool        `json:"legacyGlobalDataMigrated,omitempty"`
}

func NewWorkspaceManager(appConfigDir string, opts ...WorkspaceManagerOption) *WorkspaceManager {
	manager := &WorkspaceManager{
		appConfigDir: appConfigDir,
		now:          time.Now,
	}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func WithWorkspaceClock(now func() time.Time) WorkspaceManagerOption {
	return func(manager *WorkspaceManager) {
		manager.now = now
	}
}

func (manager *WorkspaceManager) LoadState() (WorkspaceState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	persisted, err := manager.loadPersistedLocked()
	if err != nil {
		return WorkspaceState{}, err
	}
	return manager.stateFromPersistedLocked(persisted), nil
}

func (manager *WorkspaceManager) OpenWorkspace(path string) (WorkspaceState, error) {
	absolutePath, err := normalizeWorkspacePath(path)
	if err != nil {
		state, _ := manager.LoadState()
		return state, err
	}
	if !directoryAvailable(absolutePath) {
		state, _ := manager.LoadState()
		return state, errors.New("workspace path is not an accessible directory")
	}

	dataPath := workspaceDataPath(absolutePath)
	if err := os.MkdirAll(dataPath, 0o755); err != nil {
		state, _ := manager.LoadState()
		return state, err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	persisted, err := manager.loadPersistedLocked()
	if err != nil {
		return WorkspaceState{}, err
	}
	openedAt := manager.now().UTC().Format(time.RFC3339Nano)
	previousCurrent := manager.current
	workspace := Workspace{
		Name:         filepath.Base(absolutePath),
		Path:         absolutePath,
		DataPath:     dataPath,
		Available:    true,
		LastOpenedAt: openedAt,
	}
	manager.current = copyWorkspace(workspace)
	persisted.RecentWorkspaces = upsertRecentWorkspace(persisted.RecentWorkspaces, workspace)
	if err := manager.savePersistedLocked(persisted); err != nil {
		manager.current = previousCurrent
		return manager.stateFromPersistedLocked(persisted), err
	}
	return manager.stateFromPersistedLocked(persisted), nil
}

func (manager *WorkspaceManager) ClearRecentWorkspaces() (WorkspaceState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	persisted, err := manager.loadPersistedLocked()
	if err != nil {
		return WorkspaceState{}, err
	}
	persisted.RecentWorkspaces = []Workspace{}
	if err := manager.savePersistedLocked(persisted); err != nil {
		return WorkspaceState{}, err
	}
	return manager.stateFromPersistedLocked(persisted), nil
}

func (manager *WorkspaceManager) CloseWorkspace() (WorkspaceState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	persisted, err := manager.loadPersistedLocked()
	if err != nil {
		return WorkspaceState{}, err
	}
	manager.current = nil
	return manager.stateFromPersistedLocked(persisted), nil
}

func (manager *WorkspaceManager) MigrateLegacyGlobalData() (WorkspaceState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	persisted, err := manager.loadPersistedLocked()
	if err != nil {
		return WorkspaceState{}, err
	}
	if persisted.LegacyGlobalDataMigrated || !manager.hasLegacyGlobalDataLocked() {
		return manager.stateFromPersistedLocked(persisted), nil
	}

	legacyWorkspacePath := filepath.Join(manager.appConfigDir, legacyWorkspaceDirName)
	legacyDataPath := workspaceDataPath(legacyWorkspacePath)
	if err := os.MkdirAll(legacyDataPath, 0o755); err != nil {
		return WorkspaceState{}, err
	}
	for _, name := range []string{legacyGlobalProjectsFile, legacyGlobalHistoryFile} {
		if err := copyFileIfMissing(filepath.Join(manager.appConfigDir, name), filepath.Join(legacyDataPath, name)); err != nil {
			return WorkspaceState{}, err
		}
	}

	openedAt := manager.now().UTC().Format(time.RFC3339Nano)
	workspace := Workspace{
		Name:         filepath.Base(legacyWorkspacePath),
		Path:         legacyWorkspacePath,
		DataPath:     legacyDataPath,
		Available:    true,
		LastOpenedAt: openedAt,
	}
	persisted.RecentWorkspaces = upsertRecentWorkspace(persisted.RecentWorkspaces, workspace)
	persisted.LegacyGlobalDataMigrated = true
	if err := manager.savePersistedLocked(persisted); err != nil {
		return WorkspaceState{}, err
	}
	return manager.stateFromPersistedLocked(persisted), nil
}

func (manager *WorkspaceManager) CurrentWorkspace() *Workspace {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.current == nil {
		return nil
	}
	return copyWorkspace(*manager.current)
}

func (manager *WorkspaceManager) loadPersistedLocked() (persistedWorkspaceState, error) {
	state := persistedWorkspaceState{
		Version:          workspaceStateFileVersion,
		RecentWorkspaces: []Workspace{},
	}
	data, err := os.ReadFile(manager.recentPath())
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return persistedWorkspaceState{}, err
	}
	if len(data) == 0 {
		return state, nil
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return persistedWorkspaceState{}, err
	}
	if state.Version == 0 {
		state.Version = workspaceStateFileVersion
	}
	if state.RecentWorkspaces == nil {
		state.RecentWorkspaces = []Workspace{}
	}
	state.RecentWorkspaces = normalizeRecentWorkspaces(state.RecentWorkspaces)
	return state, nil
}

func (manager *WorkspaceManager) savePersistedLocked(state persistedWorkspaceState) error {
	state.Version = workspaceStateFileVersion
	state.RecentWorkspaces = normalizeRecentWorkspaces(state.RecentWorkspaces)
	if err := os.MkdirAll(manager.appConfigDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tempPath := manager.recentPath() + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, manager.recentPath())
}

func (manager *WorkspaceManager) stateFromPersistedLocked(persisted persistedWorkspaceState) WorkspaceState {
	state := WorkspaceState{
		Version:          workspaceStateFileVersion,
		RecentWorkspaces: normalizeRecentWorkspaces(persisted.RecentWorkspaces),
	}
	if manager.current != nil {
		state.CurrentWorkspace = copyWorkspace(*manager.current)
	}
	return state
}

func (manager *WorkspaceManager) recentPath() string {
	return filepath.Join(manager.appConfigDir, recentWorkspacesFileName)
}

func (manager *WorkspaceManager) hasLegacyGlobalDataLocked() bool {
	for _, name := range []string{legacyGlobalProjectsFile, legacyGlobalHistoryFile} {
		if _, err := os.Stat(filepath.Join(manager.appConfigDir, name)); err == nil {
			return true
		}
	}
	return false
}

func normalizeWorkspacePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("workspace path is required")
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolutePath), nil
}

func workspaceDataPath(workspacePath string) string {
	return filepath.Join(workspacePath, workspaceDataDirName)
}

func upsertRecentWorkspace(recent []Workspace, workspace Workspace) []Workspace {
	next := []Workspace{workspace}
	for _, existing := range recent {
		if existing.Path == workspace.Path {
			continue
		}
		next = append(next, existing)
	}
	return normalizeRecentWorkspaces(next)
}

func normalizeRecentWorkspaces(recent []Workspace) []Workspace {
	byPath := map[string]Workspace{}
	for _, workspace := range recent {
		normalizedPath, err := normalizeWorkspacePath(workspace.Path)
		if err != nil {
			continue
		}
		workspace.Path = normalizedPath
		workspace.DataPath = workspaceDataPath(normalizedPath)
		if workspace.Name == "" {
			workspace.Name = filepath.Base(normalizedPath)
		}
		workspace.Available = directoryAvailable(normalizedPath)
		if existing, ok := byPath[normalizedPath]; !ok || workspaceOpenedAfter(workspace, existing) {
			byPath[normalizedPath] = workspace
		}
	}
	normalized := make([]Workspace, 0, len(byPath))
	for _, workspace := range byPath {
		normalized = append(normalized, workspace)
	}
	sort.Slice(normalized, func(left, right int) bool {
		leftOpenedAt, leftOK := parseWorkspaceOpenedAt(normalized[left].LastOpenedAt)
		rightOpenedAt, rightOK := parseWorkspaceOpenedAt(normalized[right].LastOpenedAt)
		if leftOK && rightOK {
			if leftOpenedAt.Equal(rightOpenedAt) {
				return normalized[left].Path < normalized[right].Path
			}
			return leftOpenedAt.After(rightOpenedAt)
		}
		if normalized[left].LastOpenedAt == normalized[right].LastOpenedAt {
			return normalized[left].Path < normalized[right].Path
		}
		return normalized[left].LastOpenedAt > normalized[right].LastOpenedAt
	})
	return normalized
}

func workspaceOpenedAfter(left Workspace, right Workspace) bool {
	leftOpenedAt, leftOK := parseWorkspaceOpenedAt(left.LastOpenedAt)
	rightOpenedAt, rightOK := parseWorkspaceOpenedAt(right.LastOpenedAt)
	if leftOK && rightOK {
		return leftOpenedAt.After(rightOpenedAt)
	}
	return left.LastOpenedAt > right.LastOpenedAt
}

func parseWorkspaceOpenedAt(value string) (time.Time, bool) {
	openedAt, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return openedAt, true
	}
	openedAt, err = time.Parse(time.RFC3339, value)
	return openedAt, err == nil
}

func copyWorkspace(workspace Workspace) *Workspace {
	copied := workspace
	return &copied
}

func copyFileIfMissing(src string, dst string) error {
	data, err := os.ReadFile(src)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := os.Stat(dst); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
