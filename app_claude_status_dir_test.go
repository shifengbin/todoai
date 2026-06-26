package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeStatusDirectory_EnvOverride(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "custom-status")
	t.Setenv("TODOAI_STATUS_DIR", custom)

	if got := claudeStatusDirectory(); got != custom {
		t.Fatalf("claudeStatusDirectory() = %q, want %q", got, custom)
	}
}

func TestClaudeStatusDirectory_DefaultLivesUnderAppConfig(t *testing.T) {
	t.Setenv("TODOAI_STATUS_DIR", "")

	got := claudeStatusDirectory()

	wantSuffix := filepath.Join(applicationID, "claude-status")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("claudeStatusDirectory() = %q, want suffix %q", got, wantSuffix)
	}
	// The override must be fully cleared — an empty value should not be returned
	// verbatim (that would point at the filesystem root).
	if got == "" {
		t.Fatalf("claudeStatusDirectory() returned empty string")
	}
}
