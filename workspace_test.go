package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceManagerOpensWorkspaceCreatesDataDirAndRecordsRecent(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	now := fixedTime("2026-06-16T10:00:00Z")
	manager := NewWorkspaceManager(appConfigDir, WithWorkspaceClock(func() time.Time { return now }))

	state, err := manager.OpenWorkspace(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspace() error = %v", err)
	}

	absoluteWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		t.Fatalf("Abs(workspaceDir) error = %v", err)
	}
	wantDataPath := filepath.Join(absoluteWorkspace, workspaceDataDirName)
	if state.CurrentWorkspace == nil {
		t.Fatal("CurrentWorkspace = nil, want opened workspace")
	}
	if state.CurrentWorkspace.Path != absoluteWorkspace {
		t.Fatalf("CurrentWorkspace.Path = %q, want %q", state.CurrentWorkspace.Path, absoluteWorkspace)
	}
	if state.CurrentWorkspace.Name != filepath.Base(absoluteWorkspace) {
		t.Fatalf("CurrentWorkspace.Name = %q, want basename", state.CurrentWorkspace.Name)
	}
	if state.CurrentWorkspace.DataPath != wantDataPath {
		t.Fatalf("CurrentWorkspace.DataPath = %q, want %q", state.CurrentWorkspace.DataPath, wantDataPath)
	}
	if state.CurrentWorkspace.LastOpenedAt != now.UTC().Format(time.RFC3339) {
		t.Fatalf("CurrentWorkspace.LastOpenedAt = %q, want fixed time", state.CurrentWorkspace.LastOpenedAt)
	}
	if _, err := os.Stat(wantDataPath); err != nil {
		t.Fatalf("workspace data dir was not created: %v", err)
	}
	if len(state.RecentWorkspaces) != 1 || state.RecentWorkspaces[0].Path != absoluteWorkspace {
		t.Fatalf("RecentWorkspaces = %#v, want opened workspace", state.RecentWorkspaces)
	}

	reloaded, err := NewWorkspaceManager(appConfigDir).LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if reloaded.CurrentWorkspace != nil {
		t.Fatalf("reloaded CurrentWorkspace = %#v, want runtime-only current workspace", reloaded.CurrentWorkspace)
	}
	if len(reloaded.RecentWorkspaces) != 1 || reloaded.RecentWorkspaces[0].Path != absoluteWorkspace {
		t.Fatalf("reloaded RecentWorkspaces = %#v, want persisted recent workspace", reloaded.RecentWorkspaces)
	}
}

func TestWorkspaceManagerRecentWorkspacesAreDeduplicatedAndSorted(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	times := []time.Time{
		fixedTime("2026-06-16T10:00:00Z"),
		fixedTime("2026-06-16T10:05:00Z"),
		fixedTime("2026-06-16T10:10:00Z"),
	}
	index := 0
	manager := NewWorkspaceManager(appConfigDir, WithWorkspaceClock(func() time.Time {
		now := times[index]
		index++
		return now
	}))

	if _, err := manager.OpenWorkspace(workspaceA); err != nil {
		t.Fatalf("OpenWorkspace(A) error = %v", err)
	}
	if _, err := manager.OpenWorkspace(workspaceB); err != nil {
		t.Fatalf("OpenWorkspace(B) error = %v", err)
	}
	state, err := manager.OpenWorkspace(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspace(A again) error = %v", err)
	}

	absoluteA, _ := filepath.Abs(workspaceA)
	absoluteB, _ := filepath.Abs(workspaceB)
	if len(state.RecentWorkspaces) != 2 {
		t.Fatalf("RecentWorkspaces length = %d, want 2: %#v", len(state.RecentWorkspaces), state.RecentWorkspaces)
	}
	if state.RecentWorkspaces[0].Path != absoluteA || state.RecentWorkspaces[1].Path != absoluteB {
		t.Fatalf("RecentWorkspaces order = %#v, want A then B", state.RecentWorkspaces)
	}
	if state.RecentWorkspaces[0].LastOpenedAt != times[2].UTC().Format(time.RFC3339) {
		t.Fatalf("deduped LastOpenedAt = %q, want latest reopen time", state.RecentWorkspaces[0].LastOpenedAt)
	}
}

func TestWorkspaceManagerRecentWorkspacesSortParsedTimestamps(t *testing.T) {
	older := t.TempDir()
	newer := t.TempDir()

	recent := normalizeRecentWorkspaces([]Workspace{
		{Path: older, LastOpenedAt: "2026-06-16T10:00:00Z"},
		{Path: newer, LastOpenedAt: "2026-06-16T10:00:00.000000001Z"},
	})

	if len(recent) != 2 {
		t.Fatalf("RecentWorkspaces length = %d, want 2: %#v", len(recent), recent)
	}
	if recent[0].Path != mustAbs(t, newer) {
		t.Fatalf("RecentWorkspaces order = %#v, want parsed newer timestamp first", recent)
	}
}

func TestWorkspaceManagerClearRecentDoesNotDeleteWorkspaceData(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	manager := NewWorkspaceManager(appConfigDir)
	state, err := manager.OpenWorkspace(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspace() error = %v", err)
	}
	dataPath := state.CurrentWorkspace.DataPath
	writeTestFile(t, filepath.Join(dataPath, "projects.json"), `{"version":1}`)

	state, err = manager.ClearRecentWorkspaces()
	if err != nil {
		t.Fatalf("ClearRecentWorkspaces() error = %v", err)
	}

	if len(state.RecentWorkspaces) != 0 {
		t.Fatalf("RecentWorkspaces = %#v, want empty", state.RecentWorkspaces)
	}
	assertFileContent(t, filepath.Join(dataPath, "projects.json"), `{"version":1}`)
	if state.CurrentWorkspace == nil || state.CurrentWorkspace.DataPath != dataPath {
		t.Fatalf("CurrentWorkspace = %#v, want current workspace preserved", state.CurrentWorkspace)
	}
}

func TestWorkspaceManagerOpenUnavailableWorkspaceKeepsCurrentWorkspace(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceDir := t.TempDir()
	missingDir := filepath.Join(t.TempDir(), "missing")
	manager := NewWorkspaceManager(appConfigDir)
	opened, err := manager.OpenWorkspace(workspaceDir)
	if err != nil {
		t.Fatalf("OpenWorkspace(initial) error = %v", err)
	}

	state, err := manager.OpenWorkspace(missingDir)
	if err == nil {
		t.Fatal("OpenWorkspace(missing) error = nil, want error")
	}

	if state.CurrentWorkspace == nil || state.CurrentWorkspace.Path != opened.CurrentWorkspace.Path {
		t.Fatalf("CurrentWorkspace after failed open = %#v, want previous %#v", state.CurrentWorkspace, opened.CurrentWorkspace)
	}
	if len(state.RecentWorkspaces) != 1 || state.RecentWorkspaces[0].Path != opened.CurrentWorkspace.Path {
		t.Fatalf("RecentWorkspaces after failed open = %#v, want previous only", state.RecentWorkspaces)
	}
}

func TestWorkspaceManagerOpenWorkspaceSaveFailureKeepsCurrentWorkspace(t *testing.T) {
	appConfigDir := t.TempDir()
	workspaceA := t.TempDir()
	workspaceB := t.TempDir()
	manager := NewWorkspaceManager(appConfigDir)
	opened, err := manager.OpenWorkspace(workspaceA)
	if err != nil {
		t.Fatalf("OpenWorkspace(A) error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(appConfigDir, recentWorkspacesFileName+".tmp"), 0o755); err != nil {
		t.Fatalf("Mkdir(recent tmp blocker) error = %v", err)
	}

	state, err := manager.OpenWorkspace(workspaceB)
	if err == nil {
		t.Fatal("OpenWorkspace(B) error = nil, want save error")
	}

	if state.CurrentWorkspace == nil || state.CurrentWorkspace.Path != opened.CurrentWorkspace.Path {
		t.Fatalf("CurrentWorkspace after save failure = %#v, want previous %#v", state.CurrentWorkspace, opened.CurrentWorkspace)
	}
}

func TestWorkspaceManagerMigratesLegacyGlobalDataToLegacyWorkspaceOnce(t *testing.T) {
	appConfigDir := t.TempDir()
	writeTestFile(t, filepath.Join(appConfigDir, "projects.json"), `{"version":1,"source":"projects"}`)
	writeTestFile(t, filepath.Join(appConfigDir, "settings.json"), `{"version":1,"source":"settings"}`)
	writeTestFile(t, filepath.Join(appConfigDir, "terminal-history.json"), `{"version":1,"records":[]}`)
	manager := NewWorkspaceManager(appConfigDir, WithWorkspaceClock(func() time.Time {
		return fixedTime("2026-06-16T10:00:00Z")
	}))

	state, err := manager.MigrateLegacyGlobalData()
	if err != nil {
		t.Fatalf("MigrateLegacyGlobalData() error = %v", err)
	}

	legacyWorkspacePath := filepath.Join(appConfigDir, legacyWorkspaceDirName)
	legacyDataPath := filepath.Join(legacyWorkspacePath, workspaceDataDirName)
	assertFileContent(t, filepath.Join(legacyDataPath, "projects.json"), `{"version":1,"source":"projects"}`)
	assertFileContent(t, filepath.Join(legacyDataPath, "terminal-history.json"), `{"version":1,"records":[]}`)
	assertFileContent(t, filepath.Join(appConfigDir, "projects.json"), `{"version":1,"source":"projects"}`)
	assertFileContent(t, filepath.Join(appConfigDir, "settings.json"), `{"version":1,"source":"settings"}`)
	if _, err := os.Stat(filepath.Join(legacyDataPath, "settings.json")); !os.IsNotExist(err) {
		t.Fatalf("legacy workspace settings.json exists or stat error = %v, want not exist", err)
	}
	if len(state.RecentWorkspaces) != 1 || state.RecentWorkspaces[0].Path != legacyWorkspacePath {
		t.Fatalf("RecentWorkspaces = %#v, want app-managed legacy workspace", state.RecentWorkspaces)
	}

	writeTestFile(t, filepath.Join(legacyDataPath, "projects.json"), `{"version":1,"source":"user-edited"}`)
	state, err = manager.MigrateLegacyGlobalData()
	if err != nil {
		t.Fatalf("MigrateLegacyGlobalData(second) error = %v", err)
	}
	assertFileContent(t, filepath.Join(legacyDataPath, "projects.json"), `{"version":1,"source":"user-edited"}`)
	if len(state.RecentWorkspaces) != 1 || state.RecentWorkspaces[0].Path != legacyWorkspacePath {
		t.Fatalf("RecentWorkspaces after second migration = %#v, want unchanged legacy workspace", state.RecentWorkspaces)
	}
}

func TestWorkspaceManagerDoesNotMigrateSettingsOnlyLegacyData(t *testing.T) {
	appConfigDir := t.TempDir()
	writeTestFile(t, filepath.Join(appConfigDir, "settings.json"), `{"version":1,"source":"settings"}`)
	manager := NewWorkspaceManager(appConfigDir)

	state, err := manager.MigrateLegacyGlobalData()
	if err != nil {
		t.Fatalf("MigrateLegacyGlobalData() error = %v", err)
	}

	if len(state.RecentWorkspaces) != 0 {
		t.Fatalf("RecentWorkspaces = %#v, want empty for global-only settings", state.RecentWorkspaces)
	}
	if _, err := os.Stat(filepath.Join(appConfigDir, legacyWorkspaceDirName)); !os.IsNotExist(err) {
		t.Fatalf("legacy workspace exists or stat error = %v, want not exist", err)
	}
	assertFileContent(t, filepath.Join(appConfigDir, "settings.json"), `{"version":1,"source":"settings"}`)
}

func fixedTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
