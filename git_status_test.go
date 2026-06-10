package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseGitStatusPorcelainV2CountsBranchAndChanges(t *testing.T) {
	status := parseGitStatusPorcelainV2(`# branch.oid 1234567890abcdef
# branch.head main
# branch.upstream origin/main
# branch.ab +2 -1
1 .M N... 100644 100644 100644 abcdef abcdef modified.go
1 M. N... 100644 100644 100644 abcdef abcdef staged.go
1 MM N... 100644 100644 100644 abcdef abcdef both.go
2 R. N... 100644 100644 100644 abcdef abcdef R100 old.go	new.go
? untracked.go
`)

	if !status.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if status.Branch != "main" {
		t.Fatalf("Branch = %q, want main", status.Branch)
	}
	if status.ChangedCount != 5 {
		t.Fatalf("ChangedCount = %d, want 5", status.ChangedCount)
	}
	if status.StagedCount != 3 {
		t.Fatalf("StagedCount = %d, want 3", status.StagedCount)
	}
	if status.UnstagedCount != 2 {
		t.Fatalf("UnstagedCount = %d, want 2", status.UnstagedCount)
	}
	if status.UntrackedCount != 1 {
		t.Fatalf("UntrackedCount = %d, want 1", status.UntrackedCount)
	}
	if status.Ahead != 2 {
		t.Fatalf("Ahead = %d, want 2", status.Ahead)
	}
	if status.Behind != 1 {
		t.Fatalf("Behind = %d, want 1", status.Behind)
	}
}

func TestParseGitStatusPorcelainV2HandlesCleanDetachedRepository(t *testing.T) {
	status := parseGitStatusPorcelainV2(`# branch.oid 1234567890abcdef
# branch.head (detached)
`)

	if !status.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if status.Branch != "(detached)" {
		t.Fatalf("Branch = %q, want (detached)", status.Branch)
	}
	if status.ChangedCount != 0 {
		t.Fatalf("ChangedCount = %d, want 0", status.ChangedCount)
	}
}

func TestGitStatusForPathReturnsNonRepoStatus(t *testing.T) {
	status, err := gitStatusForPath("/work/not-repo", func(ctx context.Context, path string) ([]byte, error) {
		return []byte("fatal: not a git repository (or any of the parent directories): .git"), errors.New("exit status 128")
	})

	if err != nil {
		t.Fatalf("gitStatusForPath() error = %v, want nil", err)
	}
	if status.IsRepo {
		t.Fatal("IsRepo = true, want false")
	}
}

func TestGitStatusForPathReturnsCommandFailure(t *testing.T) {
	_, err := gitStatusForPath("/work/repo", func(ctx context.Context, path string) ([]byte, error) {
		return []byte("git: command not found"), errors.New("exec: git not found")
	})

	if err == nil {
		t.Fatal("gitStatusForPath() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git status failed") {
		t.Fatalf("error = %q, want git status failed", err.Error())
	}
}
