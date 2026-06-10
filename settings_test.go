package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsManagerDetectsAndPersistsShellOnFirstLoad(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	detectedPath := executableFile(t, "zsh")
	detectCount := 0
	manager := NewSettingsManager(
		configPath,
		WithSettingsShellDetector(func() (TerminalShellSetting, error) {
			detectCount++
			return TerminalShellSetting{
				Path:        detectedPath,
				DisplayName: "zsh",
				Source:      ShellSourceDetected,
				Available:   true,
			}, nil
		}),
	)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Selected.Path != detectedPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, detectedPath)
	}
	if state.Selected.DisplayName != "zsh" {
		t.Fatalf("Selected.DisplayName = %q, want zsh", state.Selected.DisplayName)
	}
	if state.Selected.Source != ShellSourceDetected {
		t.Fatalf("Selected.Source = %q, want %q", state.Selected.Source, ShellSourceDetected)
	}
	if !state.Selected.Available {
		t.Fatal("Selected.Available = false, want true")
	}
	if detectCount != 1 {
		t.Fatalf("detect count = %d, want 1", detectCount)
	}

	reloaded := NewSettingsManager(
		configPath,
		WithSettingsShellDetector(func() (TerminalShellSetting, error) {
			detectCount++
			return TerminalShellSetting{Path: executableFile(t, "bash"), DisplayName: "bash", Source: ShellSourceDetected, Available: true}, nil
		}),
	)
	state, err = reloaded.Load()
	if err != nil {
		t.Fatalf("reloaded Load() error = %v", err)
	}
	if state.Selected.Path != detectedPath {
		t.Fatalf("reloaded Selected.Path = %q, want persisted %q", state.Selected.Path, detectedPath)
	}
	if detectCount != 1 {
		t.Fatalf("detect count after reload = %d, want still 1", detectCount)
	}
}

func TestSettingsManagerAddsDefaultLaunchProfilesOnFirstLoad(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	detectedPath := executableFile(t, "zsh")
	manager := NewSettingsManager(
		configPath,
		WithSettingsShellDetector(func() (TerminalShellSetting, error) {
			return TerminalShellSetting{
				Path:        detectedPath,
				DisplayName: "zsh",
				Source:      ShellSourceDetected,
				Available:   true,
			}, nil
		}),
	)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "codex", Command: "codex"},
		{Name: "claude", Command: "claude"},
	})
}

func TestSettingsManagerRejectsInvalidManualShellWithoutChangingSavedSetting(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	validPath := executableFile(t, "bash")
	manager := NewSettingsManager(configPath)

	if _, err := manager.SaveShellPath(validPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath(valid) error = %v", err)
	}

	if _, err := manager.SaveShellPath(filepath.Join(t.TempDir(), "missing-shell"), ShellSourceManual); err == nil {
		t.Fatal("SaveShellPath(invalid) error = nil, want error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Selected.Path != validPath {
		t.Fatalf("Selected.Path after invalid save = %q, want %q", state.Selected.Path, validPath)
	}
}

func TestSettingsManagerPreservesExplicitEmptyLaunchProfilesWhenSavingShellPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	oldShellPath := executableFile(t, "zsh")
	newShellPath := executableFile(t, "bash")
	writeSettingsFileWithLaunchProfiles(t, configPath, oldShellPath, "manual", "[]")

	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(newShellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Selected.Path != newShellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, newShellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{})
}

func TestSettingsManagerSavesLaunchProfilesAndPreservesShellSetting(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}

	state, err := manager.SaveLaunchProfiles([]TerminalLaunchProfileSetting{
		{Name: " Codex GPT-5 ", Command: " codex --model gpt-5 "},
		{Name: "Claude Plan", Command: "claude --dangerously-skip-permissions"},
	})
	if err != nil {
		t.Fatalf("SaveLaunchProfiles() error = %v", err)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "Codex GPT-5", Command: "codex --model gpt-5"},
		{Name: "Claude Plan", Command: "claude --dangerously-skip-permissions"},
	})

	reloaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if reloaded.Selected.Path != shellPath {
		t.Fatalf("reloaded Selected.Path = %q, want %q", reloaded.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, reloaded.LaunchProfiles, state.LaunchProfiles)
}

func TestSettingsManagerRejectsInvalidLaunchProfilesWithoutChangingSavedProfiles(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	want := []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex"}}
	if _, err := manager.SaveLaunchProfiles(want); err != nil {
		t.Fatalf("SaveLaunchProfiles(valid) error = %v", err)
	}

	cases := []struct {
		name     string
		profiles []TerminalLaunchProfileSetting
	}{
		{name: "empty name", profiles: []TerminalLaunchProfileSetting{{Name: " ", Command: "codex"}}},
		{name: "empty command", profiles: []TerminalLaunchProfileSetting{{Name: "Codex", Command: " "}}},
		{name: "reserved terminal", profiles: []TerminalLaunchProfileSetting{{Name: "terminal", Command: "bash"}}},
		{name: "duplicate names", profiles: []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex"}, {Name: "codex", Command: "codex --model gpt-5"}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := manager.SaveLaunchProfiles(tt.profiles); err == nil {
				t.Fatal("SaveLaunchProfiles(invalid) error = nil, want error")
			}
			state, err := manager.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			assertLaunchProfiles(t, state.LaunchProfiles, want)
		})
	}
}

func TestSettingsManagerReportsUnavailableSavedShellWithFallback(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	missingPath := filepath.Join(t.TempDir(), "old-zsh")
	fallbackPath := executableFile(t, "sh")
	manager := NewSettingsManager(
		configPath,
		WithSettingsShellDetector(func() (TerminalShellSetting, error) {
			return TerminalShellSetting{
				Path:        fallbackPath,
				DisplayName: "sh",
				Source:      ShellSourceDetected,
				Available:   true,
			}, nil
		}),
	)
	writeSettingsFile(t, configPath, missingPath, "manual")

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Selected.Path != missingPath {
		t.Fatalf("Selected.Path = %q, want saved %q", state.Selected.Path, missingPath)
	}
	if state.Selected.Available {
		t.Fatal("Selected.Available = true, want false")
	}
	if state.Fallback == nil {
		t.Fatal("Fallback = nil, want detected fallback")
	}
	if state.Fallback.Path != fallbackPath {
		t.Fatalf("Fallback.Path = %q, want %q", state.Fallback.Path, fallbackPath)
	}

	reloaded, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(settings) error = %v", err)
	}
	if !strings.Contains(string(reloaded), missingPath) {
		t.Fatalf("settings file no longer contains saved missing path %q: %s", missingPath, string(reloaded))
	}
}

func TestShellDetectorPrefersExecutableEnvShellThenKnownCandidates(t *testing.T) {
	envShell := executableFile(t, "fish")
	candidateShell := executableFile(t, "zsh")
	detector := NewShellDetector(
		WithShellDetectorEnv(func(key string) string {
			if key == "SHELL" {
				return envShell
			}
			return ""
		}),
		WithShellDetectorCandidates([]string{candidateShell}),
	)

	shell, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if shell.Path != envShell {
		t.Fatalf("Path = %q, want env shell %q", shell.Path, envShell)
	}
	if shell.DisplayName != "fish" {
		t.Fatalf("DisplayName = %q, want fish", shell.DisplayName)
	}
	if shell.Source != ShellSourceDetected {
		t.Fatalf("Source = %q, want %q", shell.Source, ShellSourceDetected)
	}
	if !shell.Available {
		t.Fatal("Available = false, want true")
	}
}

func TestShellDetectorSkipsUnavailableEnvShell(t *testing.T) {
	candidateShell := executableFile(t, "bash")
	detector := NewShellDetector(
		WithShellDetectorEnv(func(key string) string {
			if key == "SHELL" {
				return filepath.Join(t.TempDir(), "missing-zsh")
			}
			return ""
		}),
		WithShellDetectorCandidates([]string{candidateShell}),
	)

	shell, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if shell.Path != candidateShell {
		t.Fatalf("Path = %q, want candidate %q", shell.Path, candidateShell)
	}
	if shell.DisplayName != "bash" {
		t.Fatalf("DisplayName = %q, want bash", shell.DisplayName)
	}
}

func assertLaunchProfiles(t *testing.T, got []TerminalLaunchProfileSetting, want []TerminalLaunchProfileSetting) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(LaunchProfiles) = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("LaunchProfiles[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func executableFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	return path
}

func writeSettingsFileWithLaunchProfiles(t *testing.T, configPath string, shellPath string, source string, launchProfilesJSON string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(settings dir) error = %v", err)
	}
	data := `{
  "version": 1,
  "selected": {
    "path": "` + shellPath + `",
    "displayName": "zsh",
    "source": "` + source + `",
    "available": true
  },
  "launchProfiles": ` + launchProfilesJSON + `
}`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(settings) error = %v", err)
	}
}

func writeSettingsFile(t *testing.T, configPath string, shellPath string, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(settings dir) error = %v", err)
	}
	data := `{
  "version": 1,
  "selected": {
    "path": "` + shellPath + `",
    "displayName": "zsh",
    "source": "` + source + `",
    "available": true
  }
}`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(settings) error = %v", err)
	}
}
