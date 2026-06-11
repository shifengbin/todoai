package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	settingsConfigVersion      = 1
	ShellSourceDetected        = "detected"
	ShellSourceManual          = "manual"
	AppearanceThemeLight       = "light"
	AppearanceThemeDark        = "dark"
	legacyCodexLaunchCommand   = "codex"
	legacyClaudeLaunchCommand  = "claude"
	defaultCodexLaunchCommand  = "codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox"
	defaultClaudeLaunchCommand = "claude --dangerously-skip-permissions"
)

type TerminalShellSetting struct {
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
	Source      string `json:"source"`
	Available   bool   `json:"available"`
}

type TerminalLaunchProfileSetting struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type TerminalSettingsState struct {
	Version        int                            `json:"version"`
	Selected       TerminalShellSetting           `json:"selected"`
	Detected       *TerminalShellSetting          `json:"detected,omitempty"`
	Fallback       *TerminalShellSetting          `json:"fallback,omitempty"`
	LaunchProfiles []TerminalLaunchProfileSetting `json:"launchProfiles"`
	Theme          string                         `json:"theme"`
}

type SettingsManagerOption func(*SettingsManager)

type SettingsManager struct {
	mu             sync.Mutex
	configPath     string
	detect         func() (TerminalShellSetting, error)
	shellAvailable func(string) bool
}

type persistedSettings struct {
	Version        int                            `json:"version"`
	Selected       TerminalShellSetting           `json:"selected"`
	LaunchProfiles []TerminalLaunchProfileSetting `json:"launchProfiles"`
	Theme          string                         `json:"theme"`
}

func NewSettingsManager(configPath string, opts ...SettingsManagerOption) *SettingsManager {
	detector := NewShellDetector()
	manager := &SettingsManager{
		configPath:     configPath,
		detect:         detector.Detect,
		shellAvailable: shellPathAvailable,
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

func WithSettingsShellPathAvailable(available func(string) bool) SettingsManagerOption {
	return func(manager *SettingsManager) {
		manager.shellAvailable = available
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

func (manager *SettingsManager) SaveLaunchProfiles(profiles []TerminalLaunchProfileSetting) (TerminalSettingsState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	normalizedProfiles, err := normalizeTerminalLaunchProfiles(profiles)
	if err != nil {
		state, _ := manager.loadExistingLocked()
		return state, err
	}
	state, err := manager.loadLocked()
	if err != nil {
		return TerminalSettingsState{}, err
	}
	state.LaunchProfiles = normalizedProfiles
	return state, manager.saveLocked(state)
}

func (manager *SettingsManager) SaveTheme(theme string) (TerminalSettingsState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	theme = strings.TrimSpace(theme)
	if !supportedAppearanceTheme(theme) {
		state, _ := manager.loadExistingLocked()
		return state, fmt.Errorf("unsupported appearance theme: %s", theme)
	}
	state, err := manager.loadLocked()
	if err != nil {
		return TerminalSettingsState{}, err
	}
	state.Theme = theme
	return state, manager.saveLocked(state)
}

func (manager *SettingsManager) SaveShellPath(path string, source string) (TerminalSettingsState, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.shellAvailable(path) {
		state, _ := manager.loadExistingLocked()
		return state, fmt.Errorf("terminal shell path is not executable: %s", path)
	}

	selected := TerminalShellSetting{
		Path:        filepath.Clean(path),
		DisplayName: shellNameFromPath(path),
		Source:      normalizeShellSource(source),
		Available:   true,
	}
	state, _ := manager.loadExistingLocked()
	state.Version = settingsConfigVersion
	state.Selected = selected
	if state.LaunchProfiles == nil {
		state.LaunchProfiles = defaultTerminalLaunchProfiles()
	}
	state.Theme = normalizeAppearanceTheme(state.Theme)
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
			Version:        settingsConfigVersion,
			Selected:       detected,
			Detected:       &detected,
			LaunchProfiles: defaultTerminalLaunchProfiles(),
			Theme:          AppearanceThemeLight,
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
	state.Theme = normalizeAppearanceTheme(state.Theme)
	if launchProfiles, migrated := migrateDefaultTerminalLaunchProfileCommands(state.LaunchProfiles); migrated {
		state.LaunchProfiles = launchProfiles
		if saveErr := manager.saveLocked(state); saveErr != nil {
			return TerminalSettingsState{}, saveErr
		}
	}
	state.Selected.Available = manager.shellAvailable(state.Selected.Path)
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
	launchProfiles := persisted.LaunchProfiles
	if launchProfiles == nil {
		launchProfiles = defaultTerminalLaunchProfiles()
	}
	return TerminalSettingsState{
		Version:        persisted.Version,
		Selected:       persisted.Selected,
		LaunchProfiles: launchProfiles,
		Theme:          normalizeAppearanceTheme(persisted.Theme),
	}, nil
}

func (manager *SettingsManager) saveLocked(state TerminalSettingsState) error {
	if err := os.MkdirAll(filepath.Dir(manager.configPath), 0o755); err != nil {
		return err
	}
	launchProfiles := state.LaunchProfiles
	if launchProfiles == nil {
		launchProfiles = defaultTerminalLaunchProfiles()
	}
	data, err := json.MarshalIndent(persistedSettings{
		Version:        settingsConfigVersion,
		Selected:       normalizeShellSetting(state.Selected),
		LaunchProfiles: append([]TerminalLaunchProfileSetting{}, launchProfiles...),
		Theme:          normalizeAppearanceTheme(state.Theme),
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
	detected.Available = manager.shellAvailable(detected.Path)
	if !detected.Available {
		return TerminalShellSetting{}, fmt.Errorf("detected terminal shell is not executable: %s", detected.Path)
	}
	return detected, nil
}

func normalizeTerminalLaunchProfiles(profiles []TerminalLaunchProfileSetting) ([]TerminalLaunchProfileSetting, error) {
	normalized := make([]TerminalLaunchProfileSetting, 0, len(profiles))
	seen := map[string]struct{}{}
	for _, profile := range profiles {
		name := strings.TrimSpace(profile.Name)
		command := strings.TrimSpace(profile.Command)
		if name == "" {
			return nil, errors.New("terminal launch profile name is required")
		}
		if command == "" {
			return nil, fmt.Errorf("terminal launch profile %q command is required", name)
		}
		key := strings.ToLower(name)
		if key == "terminal" {
			return nil, errors.New("terminal launch profile name Terminal is reserved")
		}
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("terminal launch profile name is duplicated: %s", name)
		}
		seen[key] = struct{}{}
		normalized = append(normalized, TerminalLaunchProfileSetting{Name: name, Command: command})
	}
	return normalized, nil
}

func defaultTerminalLaunchProfiles() []TerminalLaunchProfileSetting {
	return []TerminalLaunchProfileSetting{
		{Name: "codex", Command: defaultCodexLaunchCommand},
		{Name: "claude", Command: defaultClaudeLaunchCommand},
	}
}

func migrateDefaultTerminalLaunchProfileCommands(profiles []TerminalLaunchProfileSetting) ([]TerminalLaunchProfileSetting, bool) {
	migrated := append([]TerminalLaunchProfileSetting{}, profiles...)
	changed := false
	for index := range migrated {
		profile := &migrated[index]
		switch {
		case profile.Name == "codex" && profile.Command == legacyCodexLaunchCommand:
			profile.Command = defaultCodexLaunchCommand
			changed = true
		case profile.Name == "claude" && profile.Command == legacyClaudeLaunchCommand:
			profile.Command = defaultClaudeLaunchCommand
			changed = true
		}
	}
	if !changed {
		return profiles, false
	}
	return migrated, true
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

func normalizeAppearanceTheme(theme string) string {
	theme = strings.TrimSpace(theme)
	if supportedAppearanceTheme(theme) {
		return theme
	}
	return AppearanceThemeLight
}

func supportedAppearanceTheme(theme string) bool {
	return theme == AppearanceThemeLight || theme == AppearanceThemeDark
}

type ShellDetectorOption func(*ShellDetector)

type ShellDetector struct {
	goos           string
	getenv         func(string) string
	lookPath       func(string) (string, error)
	candidates     []string
	candidatesSet  bool
	shellAvailable func(string) bool
}

func NewShellDetector(opts ...ShellDetectorOption) ShellDetector {
	detector := ShellDetector{
		goos:     runtime.GOOS,
		getenv:   os.Getenv,
		lookPath: exec.LookPath,
	}
	detector.shellAvailable = func(path string) bool {
		return shellPathAvailableForOS(detector.goos, path, detector.getenv)
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

func WithShellDetectorPlatform(goos string) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.goos = goos
	}
}

func WithShellDetectorLookup(lookPath func(string) (string, error)) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.lookPath = lookPath
	}
}

func WithShellDetectorPathAvailable(available func(string) bool) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.shellAvailable = available
	}
}

func WithShellDetectorCandidates(candidates []string) ShellDetectorOption {
	return func(detector *ShellDetector) {
		detector.candidates = candidates
		detector.candidatesSet = true
	}
}

func (detector ShellDetector) Detect() (TerminalShellSetting, error) {
	seen := map[string]bool{}
	for _, path := range detector.candidatePaths() {
		path = normalizeShellPath(path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if detector.shellAvailable(path) {
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

func (detector ShellDetector) candidatePaths() []string {
	if detector.goos == "windows" {
		return detector.windowsCandidatePaths()
	}
	candidates := detector.candidates
	if !detector.candidatesSet {
		candidates = defaultUnixShellCandidates()
	}
	return append([]string{detector.getenv("SHELL")}, candidates...)
}

func (detector ShellDetector) windowsCandidatePaths() []string {
	paths := []string{}
	addLookup := func(name string) {
		if path, err := detector.lookPath(name); err == nil {
			paths = append(paths, path)
		}
	}
	addLookup("pwsh.exe")
	addLookup("pwsh")
	addLookup("powershell.exe")
	addLookup("powershell")
	if systemRoot := detector.windowsRoot(); systemRoot != "" {
		paths = append(paths, windowsPathJoin(systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"))
	}
	if comspec := detector.getenv("COMSPEC"); comspec != "" {
		paths = append(paths, comspec)
	}
	if systemRoot := detector.windowsRoot(); systemRoot != "" {
		paths = append(paths, windowsPathJoin(systemRoot, "System32", "cmd.exe"))
	}
	addLookup("cmd.exe")
	addLookup("cmd")
	if shell := detector.getenv("SHELL"); shell != "" {
		paths = append(paths, shell)
	}
	if detector.candidatesSet {
		paths = append(paths, detector.candidates...)
	}
	return paths
}

func (detector ShellDetector) windowsRoot() string {
	if root := detector.getenv("SystemRoot"); root != "" {
		return root
	}
	return detector.getenv("WINDIR")
}

func defaultUnixShellCandidates() []string {
	return []string{
		"/bin/zsh",
		"/usr/bin/zsh",
		"/bin/bash",
		"/usr/bin/bash",
		"/bin/fish",
		"/usr/bin/fish",
		"/bin/sh",
		"/usr/bin/sh",
	}
}

func windowsPathJoin(base string, parts ...string) string {
	path := strings.TrimRight(base, `\/`)
	for _, part := range parts {
		path += `\` + strings.Trim(part, `\/`)
	}
	return path
}

func normalizeShellPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	if path == "." {
		return ""
	}
	return path
}

func executablePath(path string) bool {
	return shellPathAvailable(path)
}

func shellPathAvailable(path string) bool {
	return shellPathAvailableForOS(runtime.GOOS, path, os.Getenv)
}

func shellPathAvailableForOS(goos string, path string, getenv func(string) string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if goos == "windows" {
		return hasWindowsExecutableExtension(path, getenv)
	}
	return info.Mode().Perm()&0o111 != 0
}

func hasWindowsExecutableExtension(path string, getenv func(string) string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}
	for _, candidate := range windowsExecutableExtensions(getenv) {
		if ext == candidate {
			return true
		}
	}
	return false
}

func windowsExecutableExtensions(getenv func(string) string) []string {
	extensions := []string{".exe", ".cmd", ".bat", ".com"}
	pathext := getenv("PATHEXT")
	if pathext == "" {
		return extensions
	}
	seen := map[string]bool{}
	for _, ext := range extensions {
		seen[ext] = true
	}
	for _, ext := range strings.FieldsFunc(pathext, func(r rune) bool { return r == ';' || r == ':' }) {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if !seen[ext] {
			extensions = append(extensions, ext)
			seen[ext] = true
		}
	}
	return extensions
}
