package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	settingsConfigVersion = 1
	ShellSourceDetected   = "detected"
	ShellSourceManual     = "manual"
)

type TerminalShellSetting struct {
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
	Source      string `json:"source"`
	Available   bool   `json:"available"`
}

type TerminalSettingsState struct {
	Version  int                   `json:"version"`
	Selected TerminalShellSetting  `json:"selected"`
	Detected *TerminalShellSetting `json:"detected,omitempty"`
	Fallback *TerminalShellSetting `json:"fallback,omitempty"`
}

type SettingsManagerOption func(*SettingsManager)

type SettingsManager struct {
	mu         sync.Mutex
	configPath string
	detect     func() (TerminalShellSetting, error)
}

type persistedSettings struct {
	Version  int                  `json:"version"`
	Selected TerminalShellSetting `json:"selected"`
}

func NewSettingsManager(configPath string, opts ...SettingsManagerOption) *SettingsManager {
	detector := NewShellDetector()
	manager := &SettingsManager{
		configPath: configPath,
		detect:     detector.Detect,
	}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func WithSettingsShellDetector(detect func() (TerminalShellSetting, error)) SettingsManagerOption {
	return func(manager *SettingsManager) {
		manager.detect = detect
	}
}

func (manager *SettingsManager) Load() (TerminalSettingsState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.loadLocked()
}

func (manager *SettingsManager) DetectShell() (TerminalShellSetting, error) {
	return manager.detectShell()
}

func (manager *SettingsManager) SaveShellPath(path string, source string) (TerminalSettingsState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !executablePath(path) {
		state, _ := manager.loadExistingLocked()
		return state, fmt.Errorf("terminal shell path is not executable: %s", path)
	}

	selected := TerminalShellSetting{
		Path:        filepath.Clean(path),
		DisplayName: shellNameFromPath(path),
		Source:      normalizeShellSource(source),
		Available:   true,
	}
	state := TerminalSettingsState{
		Version:  settingsConfigVersion,
		Selected: selected,
	}
	return state, manager.saveLocked(state)
}

func (manager *SettingsManager) ResolveShellPath() string {
	state, err := manager.Load()
	if err != nil {
		return DefaultShellPath()
	}
	if state.Selected.Available && state.Selected.Path != "" {
		return state.Selected.Path
	}
	if state.Fallback != nil && state.Fallback.Available && state.Fallback.Path != "" {
		return state.Fallback.Path
	}
	return DefaultShellPath()
}

func (manager *SettingsManager) loadLocked() (TerminalSettingsState, error) {
	state, err := manager.loadExistingLocked()
	if errors.Is(err, os.ErrNotExist) || state.Selected.Path == "" {
		detected, detectErr := manager.detectShell()
		if detectErr != nil {
			return TerminalSettingsState{}, detectErr
		}
		state = TerminalSettingsState{
			Version:  settingsConfigVersion,
			Selected: detected,
			Detected: &detected,
		}
		if saveErr := manager.saveLocked(state); saveErr != nil {
			return TerminalSettingsState{}, saveErr
		}
		return state, nil
	}
	if err != nil {
		return TerminalSettingsState{}, err
	}

	state.Selected = normalizeShellSetting(state.Selected)
	state.Selected.Available = executablePath(state.Selected.Path)
	if !state.Selected.Available {
		fallback, fallbackErr := manager.detectShell()
		if fallbackErr == nil {
			state.Fallback = &fallback
		}
	}
	return state, nil
}

func (manager *SettingsManager) loadExistingLocked() (TerminalSettingsState, error) {
	state := TerminalSettingsState{
		Version: settingsConfigVersion,
	}

	data, err := os.ReadFile(manager.configPath)
	if err != nil {
		return state, err
	}
	if len(data) == 0 {
		return state, nil
	}

	var persisted persistedSettings
	if err := json.Unmarshal(data, &persisted); err != nil {
		return TerminalSettingsState{}, err
	}
	if persisted.Version == 0 {
		persisted.Version = settingsConfigVersion
	}
	return TerminalSettingsState{
		Version:  persisted.Version,
		Selected: persisted.Selected,
	}, nil
}

func (manager *SettingsManager) saveLocked(state TerminalSettingsState) error {
	if err := os.MkdirAll(filepath.Dir(manager.configPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(persistedSettings{
		Version:  settingsConfigVersion,
		Selected: normalizeShellSetting(state.Selected),
	}, "", "  ")
	if err != nil {
		return err
	}
	tempPath := manager.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, manager.configPath)
}

func (manager *SettingsManager) detectShell() (TerminalShellSetting, error) {
	detected, err := manager.detect()
	if err != nil {
		return TerminalShellSetting{}, err
	}
	detected = normalizeShellSetting(detected)
	detected.Source = ShellSourceDetected
	detected.Available = executablePath(detected.Path)
	if !detected.Available {
		return TerminalShellSetting{}, fmt.Errorf("detected terminal shell is not executable: %s", detected.Path)
	}
	return detected, nil
}

func normalizeShellSetting(setting TerminalShellSetting) TerminalShellSetting {
	setting.Path = filepath.Clean(setting.Path)
	if setting.Path == "." {
		setting.Path = ""
	}
	if setting.DisplayName == "" {
		setting.DisplayName = shellNameFromPath(setting.Path)
	}
	if setting.Source == "" {
		setting.Source = ShellSourceManual
	}
	return setting
}

func normalizeShellSource(source string) string {
	if source == ShellSourceDetected {
		return ShellSourceDetected
	}
	return ShellSourceManual
}

type ShellDetectorOption func(*ShellDetector)

type ShellDetector struct {
	getenv     func(string) string
	candidates []string
}

func NewShellDetector(opts ...ShellDetectorOption) ShellDetector {
	detector := ShellDetector{
		getenv: os.Getenv,
		candidates: []string{
			"/bin/zsh",
			"/usr/bin/zsh",
			"/bin/bash",
			"/usr/bin/bash",
			"/bin/fish",
			"/usr/bin/fish",
			"/bin/sh",
			"/usr/bin/sh",
		},
	}
	for _, opt := range opts {
		opt(&detector)
	}
	return detector
}

func WithShellDetectorEnv(getenv func(string) string) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.getenv = getenv
	}
}

func WithShellDetectorCandidates(candidates []string) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.candidates = candidates
	}
}

func (detector ShellDetector) Detect() (TerminalShellSetting, error) {
	seen := map[string]bool{}
	for _, path := range append([]string{detector.getenv("SHELL")}, detector.candidates...) {
		path = filepath.Clean(path)
		if path == "." || path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if executablePath(path) {
			return TerminalShellSetting{
				Path:        path,
				DisplayName: shellNameFromPath(path),
				Source:      ShellSourceDetected,
				Available:   true,
			}, nil
		}
	}
	return TerminalShellSetting{}, errors.New("no executable terminal shell found")
}

func executablePath(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
}
