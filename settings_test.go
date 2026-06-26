package main

import (
	"encoding/json"
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
		{Name: "codex", Command: "codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox", Enabled: true},
		{Name: "claude", Command: "claude --dangerously-skip-permissions", Enabled: true},
	})
}

func TestSettingsManagerMigratesLegacyDefaultLaunchProfileCommands(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	writeSettingsFileWithLaunchProfiles(t, configPath, shellPath, "manual", `[
    {"name": "codex", "command": "codex"},
    {"name": "claude", "command": "claude"},
    {"name": "Codex Custom", "command": "codex"}
  ]`)
	manager := NewSettingsManager(configPath)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantProfiles := []TerminalLaunchProfileSetting{
		{Name: "codex", Command: "codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox", Enabled: true},
		{Name: "claude", Command: "claude --dangerously-skip-permissions", Enabled: true},
		{Name: "Codex Custom", Command: "codex", Enabled: true},
	}
	assertLaunchProfiles(t, state.LaunchProfiles, wantProfiles)

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(settings) error = %v", err)
	}
	var saved persistedSettings
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal(settings) error = %v", err)
	}
	assertLaunchProfiles(t, saved.LaunchProfiles, wantProfiles)
}

func TestSettingsManagerPreservesSavedLaunchProfileEnabledState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	writeSettingsFileWithLaunchProfiles(t, configPath, shellPath, "manual", `[
    {"name": "codex", "command": "codex", "enabled": true},
    {"name": "claude", "command": "claude", "enabled": false}
  ]`)
	manager := NewSettingsManager(configPath)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "codex", Command: "codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox", Enabled: true},
		{Name: "claude", Command: "claude --dangerously-skip-permissions", Enabled: false},
	})
}

func TestSettingsManagerTreatsLegacyLaunchProfilesWithoutEnabledStateAsEnabled(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	writeSettingsFileWithLaunchProfiles(t, configPath, shellPath, "manual", `[
    {"name": "Codex GPT-5", "command": "codex --model gpt-5"},
    {"name": "Claude Plan", "command": "claude"}
  ]`)
	manager := NewSettingsManager(configPath)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "Codex GPT-5", Command: "codex --model gpt-5", Enabled: true},
		{Name: "Claude Plan", Command: "claude", Enabled: true},
	})
}

func TestSettingsManagerUsesLightThemeWhenSavedThemeIsMissing(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	writeSettingsFile(t, configPath, shellPath, "manual")

	manager := NewSettingsManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Theme != AppearanceThemeLight {
		t.Fatalf("Theme = %q, want %q", state.Theme, AppearanceThemeLight)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
}

func TestSettingsManagerRestoresSavedDarkTheme(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	writeSettingsFileWithTheme(t, configPath, executableFile(t, "zsh"), "manual", AppearanceThemeDark)

	manager := NewSettingsManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Theme != AppearanceThemeDark {
		t.Fatalf("Theme = %q, want %q", state.Theme, AppearanceThemeDark)
	}
}

func TestSettingsManagerNormalizesInvalidSavedThemeToLight(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	writeSettingsFileWithTheme(t, configPath, executableFile(t, "zsh"), "manual", "midnight")

	manager := NewSettingsManager(configPath)
	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Theme != AppearanceThemeLight {
		t.Fatalf("Theme = %q, want %q", state.Theme, AppearanceThemeLight)
	}
}

func TestSettingsManagerSavesThemeAndPreservesOtherSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	wantProfiles := []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex", Enabled: true}}
	if _, err := manager.SaveLaunchProfiles(wantProfiles); err != nil {
		t.Fatalf("SaveLaunchProfiles() error = %v", err)
	}

	state, err := manager.SaveTheme(AppearanceThemeDark)
	if err != nil {
		t.Fatalf("SaveTheme() error = %v", err)
	}
	if state.Theme != AppearanceThemeDark {
		t.Fatalf("Theme = %q, want %q", state.Theme, AppearanceThemeDark)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, wantProfiles)

	reloaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if reloaded.Theme != AppearanceThemeDark {
		t.Fatalf("reloaded Theme = %q, want %q", reloaded.Theme, AppearanceThemeDark)
	}
	assertLaunchProfiles(t, reloaded.LaunchProfiles, wantProfiles)
}

func TestSettingsManagerSavesTodoInitializationFilesAndPreservesOtherSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	wantProfiles := []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex", Enabled: true}}
	if _, err := manager.SaveLaunchProfiles(wantProfiles); err != nil {
		t.Fatalf("SaveLaunchProfiles() error = %v", err)
	}

	files := []TodoInitializationFileTemplate{
		{Name: "Agent Rules", Description: "任务执行约束", FileName: "AGENTS.md", Content: "请先阅读任务说明", DefaultSelected: true},
		{Name: "Agent Rules", Description: "可选提示词", FileName: "prompt.md", Content: "生成实现计划"},
	}
	state, err := manager.SaveTodoInitializationFiles(files)
	if err != nil {
		t.Fatalf("SaveTodoInitializationFiles() error = %v", err)
	}

	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, wantProfiles)
	assertTodoInitializationFiles(t, state.TodoInitializationFiles, files)

	reloaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	assertTodoInitializationFiles(t, reloaded.TodoInitializationFiles, files)
	assertLaunchProfiles(t, reloaded.LaunchProfiles, wantProfiles)
}

func TestSettingsManagerSavesTodoLifecycleScriptsAndPreservesOtherSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	shellPath := executableFile(t, "zsh")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(shellPath, ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	wantProfiles := []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex", Enabled: true}}
	if _, err := manager.SaveLaunchProfiles(wantProfiles); err != nil {
		t.Fatalf("SaveLaunchProfiles() error = %v", err)
	}
	files := []TodoInitializationFileTemplate{{Name: "Agent Rules", FileName: "AGENTS.md", Content: "rules"}}
	if _, err := manager.SaveTodoInitializationFiles(files); err != nil {
		t.Fatalf("SaveTodoInitializationFiles() error = %v", err)
	}

	scripts := []TodoLifecycleScriptTemplate{
		{Name: "Node setup", Description: "安装依赖", InitScript: "npm install", CompleteScript: "npm test", DefaultSelected: true},
		{Name: "Cleanup", Description: "清理缓存", CompleteScript: "rm -rf tmp"},
	}
	state, err := manager.SaveTodoLifecycleScripts(scripts)
	if err != nil {
		t.Fatalf("SaveTodoLifecycleScripts() error = %v", err)
	}

	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, wantProfiles)
	assertTodoInitializationFiles(t, state.TodoInitializationFiles, files)
	assertTodoLifecycleScripts(t, state.TodoLifecycleScripts, scripts)

	reloaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	assertTodoLifecycleScripts(t, reloaded.TodoLifecycleScripts, scripts)
	assertTodoInitializationFiles(t, reloaded.TodoInitializationFiles, files)
	assertLaunchProfiles(t, reloaded.LaunchProfiles, wantProfiles)
}

func TestSettingsManagerDefaultsTodoInitializationFilesToEmpty(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	writeSettingsFile(t, configPath, executableFile(t, "zsh"), "manual")
	manager := NewSettingsManager(configPath)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.TodoInitializationFiles) != 0 {
		t.Fatalf("TodoInitializationFiles = %#v, want empty", state.TodoInitializationFiles)
	}
}

func TestSettingsManagerDefaultsTodoLifecycleScriptsToEmpty(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	writeSettingsFile(t, configPath, executableFile(t, "zsh"), "manual")
	manager := NewSettingsManager(configPath)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.TodoLifecycleScripts) != 0 {
		t.Fatalf("TodoLifecycleScripts = %#v, want empty", state.TodoLifecycleScripts)
	}
}

func TestSettingsManagerRejectsInvalidTodoInitializationFiles(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(executableFile(t, "zsh"), ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}

	tests := []struct {
		name  string
		files []TodoInitializationFileTemplate
		want  string
	}{
		{
			name:  "missing name",
			files: []TodoInitializationFileTemplate{{FileName: "AGENTS.md", Content: "rules"}},
			want:  "initialization file name is required",
		},
		{
			name:  "missing file name",
			files: []TodoInitializationFileTemplate{{Name: "Rules", Content: "rules"}},
			want:  "initialization file filename is required",
		},
		{
			name: "duplicate file name",
			files: []TodoInitializationFileTemplate{
				{Name: "Rules", FileName: "AGENTS.md"},
				{Name: "Other", FileName: "agents.md"},
			},
			want: "initialization file filename is duplicated",
		},
		{
			name:  "absolute file name",
			files: []TodoInitializationFileTemplate{{Name: "Rules", FileName: "/tmp/AGENTS.md"}},
			want:  "initialization file filename must be a root-level file name",
		},
		{
			name:  "path traversal",
			files: []TodoInitializationFileTemplate{{Name: "Rules", FileName: "../AGENTS.md"}},
			want:  "initialization file filename must be a root-level file name",
		},
		{
			name:  "directory separator",
			files: []TodoInitializationFileTemplate{{Name: "Rules", FileName: "docs/AGENTS.md"}},
			want:  "initialization file filename must be a root-level file name",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := manager.SaveTodoInitializationFiles(tc.files); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("SaveTodoInitializationFiles() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestSettingsManagerRejectsInvalidTodoLifecycleScripts(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(executableFile(t, "zsh"), ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	if _, err := manager.SaveTodoLifecycleScripts([]TodoLifecycleScriptTemplate{{Name: "Valid", InitScript: "echo ok"}}); err != nil {
		t.Fatalf("SaveTodoLifecycleScripts(valid) error = %v", err)
	}

	tests := []struct {
		name    string
		scripts []TodoLifecycleScriptTemplate
		want    string
	}{
		{
			name:    "missing name",
			scripts: []TodoLifecycleScriptTemplate{{InitScript: "echo ok"}},
			want:    "lifecycle script name is required",
		},
		{
			name:    "empty scripts",
			scripts: []TodoLifecycleScriptTemplate{{Name: "Empty"}},
			want:    "lifecycle script requires an initialization or completion script",
		},
		{
			name: "multiple defaults",
			scripts: []TodoLifecycleScriptTemplate{
				{Name: "One", InitScript: "echo one", DefaultSelected: true},
				{Name: "Two", CompleteScript: "echo two", DefaultSelected: true},
			},
			want: "only one lifecycle script can be default selected",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := manager.SaveTodoLifecycleScripts(tc.scripts); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("SaveTodoLifecycleScripts() error = %v, want containing %q", err, tc.want)
			}
			state, err := manager.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			assertTodoLifecycleScripts(t, state.TodoLifecycleScripts, []TodoLifecycleScriptTemplate{{Name: "Valid", InitScript: "echo ok"}})
		})
	}
}

func TestSettingsManagerRejectsInvalidThemeWithoutChangingSavedTheme(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	manager := NewSettingsManager(configPath)
	if _, err := manager.SaveShellPath(executableFile(t, "zsh"), ShellSourceManual); err != nil {
		t.Fatalf("SaveShellPath() error = %v", err)
	}
	if _, err := manager.SaveTheme(AppearanceThemeDark); err != nil {
		t.Fatalf("SaveTheme(valid) error = %v", err)
	}

	if _, err := manager.SaveTheme("system"); err == nil {
		t.Fatal("SaveTheme(invalid) error = nil, want error")
	}

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Theme != AppearanceThemeDark {
		t.Fatalf("Theme after invalid save = %q, want %q", state.Theme, AppearanceThemeDark)
	}
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
		{Name: " Codex GPT-5 ", Command: " codex --model gpt-5 ", Enabled: true},
		{Name: "Claude Plan", Command: "claude --dangerously-skip-permissions", Enabled: false},
	})
	if err != nil {
		t.Fatalf("SaveLaunchProfiles() error = %v", err)
	}
	if state.Selected.Path != shellPath {
		t.Fatalf("Selected.Path = %q, want %q", state.Selected.Path, shellPath)
	}
	assertLaunchProfiles(t, state.LaunchProfiles, []TerminalLaunchProfileSetting{
		{Name: "Codex GPT-5", Command: "codex --model gpt-5", Enabled: true},
		{Name: "Claude Plan", Command: "claude --dangerously-skip-permissions", Enabled: false},
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
	want := []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex", Enabled: false}}
	if _, err := manager.SaveLaunchProfiles(want); err != nil {
		t.Fatalf("SaveLaunchProfiles(valid) error = %v", err)
	}

	cases := []struct {
		name     string
		profiles []TerminalLaunchProfileSetting
	}{
		{name: "empty name", profiles: []TerminalLaunchProfileSetting{{Name: " ", Command: "codex", Enabled: false}}},
		{name: "empty command", profiles: []TerminalLaunchProfileSetting{{Name: "Codex", Command: " ", Enabled: false}}},
		{name: "reserved terminal", profiles: []TerminalLaunchProfileSetting{{Name: "terminal", Command: "bash", Enabled: false}}},
		{name: "duplicate names", profiles: []TerminalLaunchProfileSetting{{Name: "Codex", Command: "codex", Enabled: true}, {Name: "codex", Command: "codex --model gpt-5", Enabled: false}}},
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

func TestShellDetectorWindowsPrefersNativeShellsInOrder(t *testing.T) {
	cases := []struct {
		name      string
		lookPath  map[string]string
		env       map[string]string
		available map[string]bool
		wantPath  string
		wantName  string
	}{
		{
			name: "prefers PowerShell 7",
			lookPath: map[string]string{
				"pwsh.exe":       `C:\Program Files\PowerShell\7\pwsh.exe`,
				"powershell.exe": `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
			},
			env: map[string]string{"COMSPEC": `C:\Windows\System32\cmd.exe`},
			available: map[string]bool{
				`C:\Program Files\PowerShell\7\pwsh.exe`:                    true,
				`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`: true,
				`C:\Windows\System32\cmd.exe`:                               true,
			},
			wantPath: `C:\Program Files\PowerShell\7\pwsh.exe`,
			wantName: "pwsh",
		},
		{
			name:     "falls back to Windows PowerShell",
			lookPath: map[string]string{"powershell.exe": `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`},
			env:      map[string]string{"COMSPEC": `C:\Windows\System32\cmd.exe`},
			available: map[string]bool{
				`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`: true,
				`C:\Windows\System32\cmd.exe`:                               true,
			},
			wantPath: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
			wantName: "powershell",
		},
		{
			name: "falls back to COMSPEC Cmd",
			env:  map[string]string{"COMSPEC": `C:\Windows\System32\cmd.exe`},
			available: map[string]bool{
				`C:\Windows\System32\cmd.exe`: true,
			},
			wantPath: `C:\Windows\System32\cmd.exe`,
			wantName: "cmd",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewShellDetector(
				WithShellDetectorPlatform("windows"),
				WithShellDetectorEnv(mapGetenv(tt.env)),
				WithShellDetectorLookup(mapLookPath(tt.lookPath)),
				WithShellDetectorPathAvailable(mapPathAvailable(tt.available)),
			)

			shell, err := detector.Detect()
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if shell.Path != tt.wantPath {
				t.Fatalf("Path = %q, want %q", shell.Path, tt.wantPath)
			}
			if shell.DisplayName != tt.wantName {
				t.Fatalf("DisplayName = %q, want %q", shell.DisplayName, tt.wantName)
			}
			if shell.Source != ShellSourceDetected {
				t.Fatalf("Source = %q, want %q", shell.Source, ShellSourceDetected)
			}
			if !shell.Available {
				t.Fatal("Available = false, want true")
			}
		})
	}
}

func TestShellDetectorWindowsDoesNotFallBackToUnixOnlyShell(t *testing.T) {
	detector := NewShellDetector(
		WithShellDetectorPlatform("windows"),
		WithShellDetectorEnv(mapGetenv(nil)),
		WithShellDetectorLookup(mapLookPath(nil)),
		WithShellDetectorPathAvailable(mapPathAvailable(map[string]bool{
			"/bin/sh": true,
		})),
	)

	if shell, err := detector.Detect(); err == nil {
		t.Fatalf("Detect() = %#v, nil error; want no shell", shell)
	}
}

func TestShellDetectorUnixStillPrefersEnvShellAndCandidates(t *testing.T) {
	envShell := executableFile(t, "fish")
	candidateShell := executableFile(t, "zsh")
	detector := NewShellDetector(
		WithShellDetectorPlatform("linux"),
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
}

func TestWindowsShellPathAvailabilityUsesExecutableExtensions(t *testing.T) {
	tempDir := t.TempDir()
	exePath := writeFileMode(t, filepath.Join(tempDir, "pwsh.exe"), 0o600)
	cmdPath := writeFileMode(t, filepath.Join(tempDir, "launch.cmd"), 0o600)
	ps1Path := writeFileMode(t, filepath.Join(tempDir, "profile.ps1"), 0o600)
	txtPath := writeFileMode(t, filepath.Join(tempDir, "notes.txt"), 0o600)
	dirPath := filepath.Join(tempDir, "folder.exe")
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatalf("Mkdir(%s) error = %v", dirPath, err)
	}

	getenv := mapGetenv(map[string]string{"PATHEXT": ".PS1;.EXE"})
	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "exe without Unix execute bit", path: exePath, want: true},
		{name: "cmd default extension", path: cmdPath, want: true},
		{name: "PATHEXT extension", path: ps1Path, want: true},
		{name: "missing path", path: filepath.Join(tempDir, "missing.exe"), want: false},
		{name: "directory", path: dirPath, want: false},
		{name: "non executable extension", path: txtPath, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := shellPathAvailableForOS("windows", tt.path, getenv); got != tt.want {
				t.Fatalf("shellPathAvailableForOS(windows, %q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestUnixShellPathAvailabilityRequiresExecutePermission(t *testing.T) {
	tempDir := t.TempDir()
	plainPath := writeFileMode(t, filepath.Join(tempDir, "plain-sh"), 0o600)
	executablePath := writeFileMode(t, filepath.Join(tempDir, "run-sh"), 0o755)

	if shellPathAvailableForOS("linux", plainPath, os.Getenv) {
		t.Fatalf("shellPathAvailableForOS(linux, %q) = true, want false", plainPath)
	}
	if !shellPathAvailableForOS("linux", executablePath, os.Getenv) {
		t.Fatalf("shellPathAvailableForOS(linux, %q) = false, want true", executablePath)
	}
}

func TestDefaultShellPathUsesWindowsFallback(t *testing.T) {
	cmdPath := `C:\Windows\System32\cmd.exe`
	detector := NewShellDetector(
		WithShellDetectorPlatform("windows"),
		WithShellDetectorEnv(mapGetenv(map[string]string{"COMSPEC": cmdPath})),
		WithShellDetectorLookup(mapLookPath(nil)),
		WithShellDetectorPathAvailable(mapPathAvailable(map[string]bool{cmdPath: true})),
	)

	if got := defaultShellPath(detector); got != cmdPath {
		t.Fatalf("defaultShellPath(windows) = %q, want %q", got, cmdPath)
	}
}

func TestSettingsManagerResolveShellPathUsesDetectedWindowsFallback(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "settings.json")
	savedPath := `C:/OldShell/missing.exe`
	fallbackPath := `C:\Windows\System32\cmd.exe`
	writeSettingsFile(t, configPath, savedPath, ShellSourceManual)
	manager := NewSettingsManager(
		configPath,
		WithSettingsShellDetector(func() (TerminalShellSetting, error) {
			return TerminalShellSetting{
				Path:        fallbackPath,
				DisplayName: "cmd",
				Source:      ShellSourceDetected,
				Available:   true,
			}, nil
		}),
		WithSettingsShellPathAvailable(mapPathAvailable(map[string]bool{fallbackPath: true})),
	)

	state, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if state.Fallback == nil || state.Fallback.Path != fallbackPath {
		t.Fatalf("Fallback = %#v, want path %q", state.Fallback, fallbackPath)
	}
	if got := manager.ResolveShellPath(); got != fallbackPath {
		t.Fatalf("ResolveShellPath() = %q, want fallback %q", got, fallbackPath)
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

func assertTodoInitializationFiles(t *testing.T, got []TodoInitializationFileTemplate, want []TodoInitializationFileTemplate) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(TodoInitializationFiles) = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("TodoInitializationFiles[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func assertTodoLifecycleScripts(t *testing.T, got []TodoLifecycleScriptTemplate, want []TodoLifecycleScriptTemplate) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(TodoLifecycleScripts) = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("TodoLifecycleScripts[%d] = %#v, want %#v", index, got[index], want[index])
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

func writeFileMode(t *testing.T, path string, mode os.FileMode) string {
	t.Helper()
	if err := os.WriteFile(path, []byte("test"), mode); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	return path
}

func mapGetenv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func mapLookPath(values map[string]string) func(string) (string, error) {
	return func(file string) (string, error) {
		if path, ok := values[file]; ok {
			return path, nil
		}
		return "", os.ErrNotExist
	}
}

func mapPathAvailable(values map[string]bool) func(string) bool {
	return func(path string) bool {
		return values[path]
	}
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

func writeSettingsFileWithTheme(t *testing.T, configPath string, shellPath string, source string, theme string) {
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
  "theme": "` + theme + `"
}`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(settings) error = %v", err)
	}
}
