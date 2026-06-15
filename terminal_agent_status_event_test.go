package main

import (
	"encoding/json"
	"testing"
)

func TestTerminalAgentStatusEventJSONShape(t *testing.T) {
	event := TerminalAgentStatusEvent{
		ProjectID:     "project-a",
		TodoID:        "todo-a",
		TodoProjectID: "todo-project-a",
		TerminalID:    "terminal-a",
		Phase:         "needs-input",
		Source:        "claude-hook",
		Confidence:    "structured",
		Reason:        "permission-prompt",
		Label:         "Claude permission",
		UpdatedAt:     1770000000000,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"projectId":     "project-a",
		"todoId":        "todo-a",
		"todoProjectId": "todo-project-a",
		"terminalId":    "terminal-a",
		"phase":         "needs-input",
		"source":        "claude-hook",
		"confidence":    "structured",
		"reason":        "permission-prompt",
		"label":         "Claude permission",
		"updatedAt":     float64(1770000000000),
	}
	for key, value := range want {
		if payload[key] != value {
			t.Fatalf("%s = %#v, want %#v in %s", key, payload[key], value, data)
		}
	}
}

func TestTerminalAgentStatusEventOmitsEmptyOptionalFields(t *testing.T) {
	event := TerminalAgentStatusEvent{
		ProjectID:  "project-a",
		TerminalID: "terminal-a",
		Phase:      "busy",
		Source:     "codex-jsonl",
		Confidence: "authoritative",
		Reason:     "turn-started",
		UpdatedAt:  1770000000000,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	for _, key := range []string{"todoId", "todoProjectId", "label"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("%s should be omitted from %s", key, data)
		}
	}
}
