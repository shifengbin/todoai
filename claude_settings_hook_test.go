package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func setHookExe(t *testing.T, exe string) {
	t.Helper()
	t.Setenv("TODOAI_HOOK_EXE", exe)
}

func writeProjectSettings(t *testing.T, project string, settings map[string]any) {
	t.Helper()
	path := filepath.Join(project, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeRawProjectSettings(t *testing.T, project, content string) {
	t.Helper()
	path := filepath.Join(project, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readProjectSettings(t *testing.T, project string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(project, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func hookEntry(command string) map[string]any {
	return map[string]any{"type": "command", "command": command}
}

func hookGroups(entries ...map[string]any) []any {
	return []any{map[string]any{"hooks": entries}}
}

func collectCommands(m map[string]any, event string) []string {
	hooks, _ := m["hooks"].(map[string]any)
	groups, _ := hooks[event].([]any)
	var cmds []string
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		items, _ := gm["hooks"].([]any)
		for _, item := range items {
			entry, _ := item.(map[string]any)
			if c, _ := entry["command"].(string); c != "" {
				cmds = append(cmds, c)
			}
		}
	}
	return cmds
}

func TestEnsureClaudeStatusHook_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	setHookExe(t, filepath.Join(dir, "bin", "todoai.exe"))

	if err := ensureClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	m := readProjectSettings(t, dir)
	for _, event := range claudeStatusHookEvents {
		cmds := collectCommands(m, event)
		ours := 0
		for _, c := range cmds {
			if commandLooksLikeTodoaiHook(c) {
				ours++
			}
		}
		if ours != 1 {
			t.Fatalf("event %s has %d todoai entries, want 1", event, ours)
		}
	}
}

func TestEnsureClaudeStatusHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	setHookExe(t, filepath.Join(dir, "todoai.exe"))

	if err := ensureClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("ensure not idempotent:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestEnsureClaudeStatusHook_PreservesUserHooks(t *testing.T) {
	dir := t.TempDir()
	writeProjectSettings(t, dir, map[string]any{
		"permissions": map[string]any{"defaultMode": "acceptEdits"},
		"hooks": map[string]any{
			"Stop": hookGroups(hookEntry("echo user-stop")),
		},
	})
	setHookExe(t, filepath.Join(dir, "todoai.exe"))

	if err := ensureClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	m := readProjectSettings(t, dir)
	perms, _ := m["permissions"].(map[string]any)
	if perms == nil || perms["defaultMode"] != "acceptEdits" {
		t.Fatalf("permissions not preserved: %#v", m["permissions"])
	}
	cmds := collectCommands(m, "Stop")
	hasUser, hasOurs := false, false
	for _, c := range cmds {
		if c == "echo user-stop" {
			hasUser = true
		}
		if commandLooksLikeTodoaiHook(c) {
			hasOurs = true
		}
	}
	if !hasUser || !hasOurs {
		t.Fatalf("expected user+ours preserved, got %v", cmds)
	}
}

func TestEnsureClaudeStatusHook_InvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeRawProjectSettings(t, dir, `{not valid json`)
	setHookExe(t, filepath.Join(dir, "todoai.exe"))

	before, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	err := ensureClaudeStatusHook(dir)
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
	after, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if !reflect.DeepEqual(before, after) {
		t.Fatal("invalid JSON file was modified")
	}
}

func TestEnsureClaudeStatusHook_UpdatesStaleCommand(t *testing.T) {
	oldExe := filepath.Join(t.TempDir(), "old-todoai.exe")
	dir := t.TempDir()
	writeProjectSettings(t, dir, map[string]any{
		"hooks": map[string]any{
			"Stop": hookGroups(hookEntry("\"" + oldExe + "\" claude-hook")),
		},
	})
	setHookExe(t, filepath.Join(dir, "todoai.exe"))

	if err := ensureClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	state, err := claudeStatusHookState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Installed {
		t.Fatalf("expected installed, got %#v", state)
	}
	if state.Stale {
		t.Fatalf("expected not stale after update, got %#v", state)
	}
}

func TestRemoveClaudeStatusHook_RemovesOnlyOurs(t *testing.T) {
	dir := t.TempDir()
	ours := "\"" + filepath.Join(t.TempDir(), "todoai.exe") + "\" claude-hook"
	writeProjectSettings(t, dir, map[string]any{
		"hooks": map[string]any{
			"Stop": hookGroups(hookEntry("echo user-stop"), hookEntry(ours)),
		},
	})

	if err := removeClaudeStatusHook(dir); err != nil {
		t.Fatal(err)
	}
	m := readProjectSettings(t, dir)
	cmds := collectCommands(m, "Stop")
	for _, c := range cmds {
		if commandLooksLikeTodoaiHook(c) {
			t.Fatalf("our hook still present after remove: %s", c)
		}
	}
	hasUser := false
	for _, c := range cmds {
		if c == "echo user-stop" {
			hasUser = true
		}
	}
	if !hasUser {
		t.Fatalf("user hook removed, got %v", cmds)
	}
}

func TestRemoveClaudeStatusHook_NoFileIsNoop(t *testing.T) {
	dir := t.TempDir()
	if err := removeClaudeStatusHook(dir); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no file created, err=%v", err)
	}
}

func TestClaudeStatusHookState_NotInstalled(t *testing.T) {
	dir := t.TempDir()
	writeProjectSettings(t, dir, map[string]any{"permissions": map[string]any{}})
	setHookExe(t, filepath.Join(dir, "todoai.exe"))
	state, err := claudeStatusHookState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if state.Installed || state.EventsCovered != 0 {
		t.Fatalf("expected not installed, got %#v", state)
	}
}

func TestClaudeStatusHookState_StaleDetection(t *testing.T) {
	setHookExe(t, filepath.Join(t.TempDir(), "current-todoai.exe"))
	dir := t.TempDir()
	writeProjectSettings(t, dir, map[string]any{
		"hooks": map[string]any{
			"Stop": hookGroups(hookEntry("\"" + filepath.Join(t.TempDir(), "old-todoai.exe") + "\" claude-hook")),
		},
	})
	state, err := claudeStatusHookState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Installed {
		t.Fatalf("expected installed, got %#v", state)
	}
	if !state.Stale {
		t.Fatalf("expected stale, got %#v", state)
	}
}
