package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMapHookEventToStatus(t *testing.T) {
	cases := []struct {
		event  string
		tool   string
		status string
		delete bool
	}{
		{"UserPromptSubmit", "", "busy", false},
		{"PreToolUse", "Bash", "tool:Bash", false},
		{"PreToolUse", "", "tool", false},
		{"PostToolUse", "Edit", "tool-done", false},
		{"Notification", "", "waiting", false},
		{"Stop", "", "idle", false},
		{"SessionStart", "", "idle", false},
		{"SessionEnd", "", "", true},
		{"SubagentStop", "", "SubagentStop", false}, // unknown event recorded verbatim
		{"", "", "busy", false},                      // empty event defaults to busy
	}
	for _, tc := range cases {
		status, action := mapHookEventToStatus(tc.event, tc.tool)
		if status != tc.status {
			t.Errorf("mapHookEventToStatus(%q,%q) status=%q want %q", tc.event, tc.tool, status, tc.status)
		}
		if gotDelete := action == claudeHookActionDelete; gotDelete != tc.delete {
			t.Errorf("mapHookEventToStatus(%q,%q) delete=%v want %v", tc.event, tc.tool, gotDelete, tc.delete)
		}
	}
}

func TestRunClaudeHookCommandWithInput_WritesStatusFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TODOAI_STATUS_DIR", dir)
	t.Setenv("TUI_HELPER_TERMINAL_ID", "terminal-xyz")

	payload, _ := json.Marshal(map[string]string{
		"session_id":      "sess-1",
		"hook_event_name": "PreToolUse",
		"cwd":             "/some/cwd",
		"tool_name":       "Bash",
		"transcript_path": "/t.jsonl",
	})

	if exit := runClaudeHookCommandWithInput(strings.NewReader(string(payload))); exit != 0 {
		t.Fatalf("exit=%d want 0", exit)
	}

	data, err := os.ReadFile(filepath.Join(dir, "sess-1.status"))
	if err != nil {
		t.Fatalf("read status file: %v", err)
	}
	var got ClaudeStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Session != "sess-1" || got.Status != "tool:Bash" || got.Tool != "Bash" ||
		got.CWD != "/some/cwd" || got.Transcript != "/t.jsonl" || got.TerminalID != "terminal-xyz" {
		t.Fatalf("status = %#v", got)
	}
	if got.TS <= 0 {
		t.Fatalf("TS not set: %d", got.TS)
	}
}

func TestRunClaudeHookCommandWithInput_SessionEndRemovesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TODOAI_STATUS_DIR", dir)

	path := filepath.Join(dir, "sess-2.status")
	if err := os.WriteFile(path, []byte(`{"session":"sess-2","status":"busy"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]string{
		"session_id":      "sess-2",
		"hook_event_name": "SessionEnd",
	})
	if exit := runClaudeHookCommandWithInput(strings.NewReader(string(payload))); exit != 0 {
		t.Fatalf("exit=%d want 0", exit)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected status file removed, got err=%v", err)
	}
}

func TestRunClaudeHookCommandWithInput_MissingSessionIdNoOp(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TODOAI_STATUS_DIR", dir)

	payload, _ := json.Marshal(map[string]string{"hook_event_name": "Stop"})
	if exit := runClaudeHookCommandWithInput(strings.NewReader(string(payload))); exit != 0 {
		t.Fatalf("exit=%d want 0", exit)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no status files, got %d", len(entries))
	}
}

func TestRunClaudeHookCommandWithInput_InvalidJSONNoCrash(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TODOAI_STATUS_DIR", dir)

	if exit := runClaudeHookCommandWithInput(strings.NewReader("not json {{{")); exit != 0 {
		t.Fatalf("exit=%d want 0", exit)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no status files on invalid JSON, got %d", len(entries))
	}
}

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "out.status")
	if err := atomicWriteFile(path, []byte(`{"hello":"world"}`), 0o600); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != `{"hello":"world"}` {
		t.Fatalf("content=%q", got)
	}
}
