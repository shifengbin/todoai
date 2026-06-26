package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

// claudeHookStatusAction says whether a hook event should write the status file
// or remove it.
type claudeHookStatusAction int

const (
	claudeHookActionWrite claudeHookStatusAction = iota
	claudeHookActionDelete
)

// claudeHookInput is the subset of Claude Code's hook stdin payload that we
// care about. Fields are optional; missing ones decode as empty strings.
type claudeHookInput struct {
	SessionID      string `json:"session_id"`
	HookEventName  string `json:"hook_event_name"`
	CWD            string `json:"cwd"`
	ToolName       string `json:"tool_name"`
	TranscriptPath string `json:"transcript_path"`
}

// mapHookEventToStatus translates a Claude Code hook event into the status token
// expected by parseClaudeStatusFile/claudePhaseAndReason. The mapping follows
// docs/claude-status-monitoring.md section 2. SessionEnd removes the file.
func mapHookEventToStatus(event, toolName string) (string, claudeHookStatusAction) {
	switch event {
	case "UserPromptSubmit":
		return "busy", claudeHookActionWrite
	case "PreToolUse":
		if toolName == "" {
			return "tool", claudeHookActionWrite
		}
		return "tool:" + toolName, claudeHookActionWrite
	case "PostToolUse":
		return "tool-done", claudeHookActionWrite
	case "Notification":
		return "waiting", claudeHookActionWrite
	case "Stop", "SessionStart":
		return "idle", claudeHookActionWrite
	case "SessionEnd":
		return "", claudeHookActionDelete
	default:
		if event == "" {
			return "busy", claudeHookActionWrite
		}
		// Unknown but non-empty event: record it verbatim so the watcher still
		// surfaces activity (claudePhaseAndReason maps unknown tokens to busy).
		return event, claudeHookActionWrite
	}
}

// runClaudeHookCommand implements the `todoai claude-hook` subcommand entry
// point used by main(). It reads Claude Code's hook payload from stdin.
func runClaudeHookCommand() int {
	return runClaudeHookCommandWithInput(os.Stdin)
}

// runClaudeHookCommandWithInput is the testable core. It inherits the terminal
// environment (TODOAI_STATUS_DIR via claudeStatusDirectory, plus
// TUI_HELPER_TERMINAL_ID) from the calling claude process. It must be fast
// (PreToolUse/PostToolUse fire per tool call) and must never fail loudly — a
// hook error blocks or corrupts Claude — so all errors are swallowed and it
// always returns 0.
func runClaudeHookCommandWithInput(stdin io.Reader) int {
	payload, err := io.ReadAll(stdin)
	if err != nil {
		return 0
	}

	var input claudeHookInput
	_ = json.Unmarshal(payload, &input) // tolerate missing/garbled fields

	if input.SessionID == "" {
		return 0
	}

	statusDir := claudeStatusDirectory()
	statusPath := filepath.Join(statusDir, input.SessionID+".status")

	status, action := mapHookEventToStatus(input.HookEventName, input.ToolName)
	if action == claudeHookActionDelete {
		_ = os.Remove(statusPath)
		return 0
	}

	if err := os.MkdirAll(statusDir, 0o755); err != nil {
		return 0
	}

	out := ClaudeStatus{
		Session:    input.SessionID,
		Status:     status,
		Event:      input.HookEventName,
		CWD:        input.CWD,
		Tool:       input.ToolName,
		Transcript: input.TranscriptPath,
		TerminalID: os.Getenv("TUI_HELPER_TERMINAL_ID"),
		TS:         time.Now().Unix(),
	}
	data, err := json.Marshal(out)
	if err != nil {
		return 0
	}
	_ = atomicWriteFile(statusPath, data, 0o644)
	return 0
}

// atomicWriteFile writes data to path atomically via a sibling temp file +
// rename, so ClaudeStatusWatcher never observes a half-written .status. The
// pattern mirrors SettingsManager.saveLocked in settings.go. Shared by the
// claude-hook subcommand and the settings.json merge logic.
func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	removeTmp := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		removeTmp()
		return err
	}
	if err := tmp.Close(); err != nil {
		removeTmp()
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		removeTmp()
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		removeTmp()
		return err
	}
	return nil
}
