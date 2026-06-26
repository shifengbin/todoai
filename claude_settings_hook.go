package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// claudeStatusHookEvents are the Claude Code lifecycle events the status hook
// subscribes to (docs/claude-status-monitoring.md section 5).
var claudeStatusHookEvents = []string{
	"SessionStart",
	"UserPromptSubmit",
	"PreToolUse",
	"PostToolUse",
	"Notification",
	"Stop",
	"SessionEnd",
}

// ClaudeHookState describes whether the status hook is installed for a project,
// surfaced to the settings panel UI.
type ClaudeHookState struct {
	Installed       bool   `json:"installed"`
	Command         string `json:"command"`
	ExpectedCommand string `json:"expectedCommand"`
	EventsCovered   int    `json:"eventsCovered"`
	Stale           bool   `json:"stale"`
}

// claudeHookCommand returns the command string to install in settings.json. It
// points at this executable's claude-hook subcommand. TODOAI_HOOK_EXE overrides
// os.Executable (handy for tests). The path is double-quoted so cmd.exe on
// Windows handles paths with spaces; JSON marshalling escapes backslashes.
func claudeHookCommand() (string, error) {
	exe := os.Getenv("TODOAI_HOOK_EXE")
	if exe == "" {
		var err error
		exe, err = os.Executable()
		if err != nil {
			return "", err
		}
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
	}
	return "\"" + exe + "\" claude-hook", nil
}

// commandLooksLikeTodoaiHook identifies a hook command we installed: it ends
// with the `claude-hook` subcommand and the executable path's basename contains
// "todoai". Robust to quoted paths containing spaces.
func commandLooksLikeTodoaiHook(command string) bool {
	command = strings.TrimSpace(command)
	if !strings.HasSuffix(command, "claude-hook") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimSuffix(command, "claude-hook"))
	rest = strings.Trim(rest, "\"")
	if rest == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(rest))
	return strings.Contains(base, "todoai")
}

func isTodoaiStatusHookEntry(entry map[string]any) bool {
	if entry == nil || entry["type"] != "command" {
		return false
	}
	command, _ := entry["command"].(string)
	return commandLooksLikeTodoaiHook(command)
}

// readClaudeSettings reads <project>/.claude/settings.json. Returns the parsed
// object (empty map if the file does not exist), whether it existed, and an
// error if it exists but is invalid JSON — callers must NOT overwrite on error.
func readClaudeSettings(path string) (map[string]any, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, false, nil
		}
		return nil, false, err
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, true, fmt.Errorf("parse %s: %w", path, err)
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return settings, true, nil
}

func writeClaudeSettings(path string, settings map[string]any) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o644)
}

// normalizeHookGroups coerces the hooks[event] value (a list of matcher groups,
// each an object with a "hooks" array) into []map[string]any, dropping malformed
// entries.
func normalizeHookGroups(existing any) []map[string]any {
	list, ok := existing.([]any)
	if !ok {
		return nil
	}
	groups := make([]map[string]any, 0, len(list))
	for _, item := range list {
		group, ok := item.(map[string]any)
		if !ok {
			continue
		}
		groups = append(groups, group)
	}
	return groups
}

// ensureTodoaiHookEntry returns the hooks[event] value with exactly one todoai
// status-hook command in the first matcher group, preserving user entries and
// replacing any stale todoai entry. Idempotent.
func ensureTodoaiHookEntry(existing any, command string) any {
	groups := normalizeHookGroups(existing)
	if len(groups) == 0 {
		groups = []map[string]any{{}}
	}
	first := groups[0]
	hookList, _ := first["hooks"].([]any)
	kept := make([]any, 0, len(hookList)+1)
	for _, item := range hookList {
		entry, ok := item.(map[string]any)
		if ok && isTodoaiStatusHookEntry(entry) {
			continue
		}
		kept = append(kept, item)
	}
	kept = append(kept, map[string]any{"type": "command", "command": command})
	first["hooks"] = kept
	groups[0] = first
	return groups
}

func ensureClaudeStatusHook(projectPath string) error {
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")
	settings, existed, err := readClaudeSettings(settingsPath)
	if err != nil {
		return err
	}
	command, err := claudeHookCommand()
	if err != nil {
		return err
	}
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for _, event := range claudeStatusHookEvents {
		hooks[event] = ensureTodoaiHookEntry(hooks[event], command)
	}
	settings["hooks"] = hooks
	if !existed {
		if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
			return err
		}
	}
	return writeClaudeSettings(settingsPath, settings)
}

func removeClaudeStatusHook(projectPath string) error {
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")
	settings, existed, err := readClaudeSettings(settingsPath)
	if err != nil {
		return err
	}
	if !existed {
		return nil
	}
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		return nil
	}
	for _, event := range claudeStatusHookEvents {
		groups := normalizeHookGroups(hooks[event])
		nonEmpty := make([]map[string]any, 0, len(groups))
		for _, group := range groups {
			hookList, _ := group["hooks"].([]any)
			kept := make([]any, 0, len(hookList))
			for _, item := range hookList {
				entry, ok := item.(map[string]any)
				if ok && isTodoaiStatusHookEntry(entry) {
					continue
				}
				kept = append(kept, item)
			}
			if len(kept) == 0 {
				delete(group, "hooks")
			} else {
				group["hooks"] = kept
			}
			if len(group) > 0 {
				nonEmpty = append(nonEmpty, group)
			}
		}
		if len(nonEmpty) == 0 {
			delete(hooks, event)
		} else {
			hooks[event] = nonEmpty
		}
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}
	return writeClaudeSettings(settingsPath, settings)
}

func claudeStatusHookState(projectPath string) (ClaudeHookState, error) {
	expected, err := claudeHookCommand()
	if err != nil {
		return ClaudeHookState{}, err
	}
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")
	settings, existed, err := readClaudeSettings(settingsPath)
	if err != nil {
		return ClaudeHookState{}, err
	}
	state := ClaudeHookState{ExpectedCommand: expected}
	if !existed {
		return state, nil
	}
	hooks, _ := settings["hooks"].(map[string]any)
	installedCommand := ""
	for _, event := range claudeStatusHookEvents {
		for _, group := range normalizeHookGroups(hooks[event]) {
			hookList, _ := group["hooks"].([]any)
			for _, item := range hookList {
				entry, ok := item.(map[string]any)
				if !ok || !isTodoaiStatusHookEntry(entry) {
					continue
				}
				state.EventsCovered++
				if cmd, _ := entry["command"].(string); cmd != "" {
					installedCommand = cmd
				}
			}
		}
	}
	if state.EventsCovered > 0 {
		state.Installed = true
		state.Command = installedCommand
		state.Stale = installedCommand != expected
	}
	return state, nil
}
