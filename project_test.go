package main

import (
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

func sequenceIDs(ids ...string) func() string {
	index := 0
	return func() string {
		id := ids[index]
		index++
		return id
	}
}
