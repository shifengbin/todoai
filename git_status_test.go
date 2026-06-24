package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

func gitAvailable() error {
	return nil
}

func TestGitStatusForPathReturnsNonRepoStatus(t *testing.T) {
	status, err := gitStatusForPath("/work/not-repo", gitAvailable, func(ctx context.Context, path string) ([]byte, error) {
		return []byte("fatal: not a git repository (or any of the parent directories): .git"), errors.New("exit status 128")
	})

	if err != nil {
		t.Fatalf("gitStatusForPath() error = %v, want nil", err)
	}
	if status.IsRepo {
		t.Fatal("IsRepo = true, want false")
	}
}

func TestGitStatusForPathReturnsGitUnavailableWithoutRunningStatus(t *testing.T) {
	called := false
	status, err := gitStatusForPath("/work/repo", func() error {
		return exec.ErrNotFound
	}, func(ctx context.Context, path string) ([]byte, error) {
		called = true
		return []byte("should not run"), nil
	})

	if err != nil {
		t.Fatalf("gitStatusForPath() error = %v, want nil", err)
	}
	if called {
		t.Fatal("git status runner was called, want skipped")
	}
	if !status.GitUnavailable {
		t.Fatal("GitUnavailable = false, want true")
	}
}

func TestGitStatusForPathReturnsCommandFailure(t *testing.T) {
	_, err := gitStatusForPath("/work/repo", gitAvailable, func(ctx context.Context, path string) ([]byte, error) {
		return []byte("git: command not found"), errors.New("exec: git not found")
	})

	if err == nil {
		t.Fatal("gitStatusForPath() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git status failed") {
		t.Fatalf("error = %q, want git status failed", err.Error())
	}
}

func TestPathHasGitRepositoryMetadataDetectsGitDirectory(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	if !pathHasGitRepositoryMetadata(repoDir) {
		t.Fatal("pathHasGitRepositoryMetadata() = false, want true for .git directory")
	}
}

func TestPathHasGitRepositoryMetadataDetectsGitFile(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoDir, ".git"), []byte("gitdir: /tmp/worktree.git"), 0o600); err != nil {
		t.Fatalf("write .git file: %v", err)
	}

	if !pathHasGitRepositoryMetadata(repoDir) {
		t.Fatal("pathHasGitRepositoryMetadata() = false, want true for .git file")
	}
}

func TestPathHasGitRepositoryMetadataRejectsNonGitDirectory(t *testing.T) {
	if pathHasGitRepositoryMetadata(t.TempDir()) {
		t.Fatal("pathHasGitRepositoryMetadata() = true, want false for non-Git directory")
	}
}

func TestInitializeGitRepositoryForPathRunsGitInit(t *testing.T) {
	called := false
	gotPath := ""
	err := initializeGitRepositoryForPath("/work/new-repo", gitAvailable, func(ctx context.Context, path string) ([]byte, error) {
		called = true
		gotPath = path
		return []byte("Initialized empty Git repository"), nil
	})

	if err != nil {
		t.Fatalf("initializeGitRepositoryForPath() error = %v, want nil", err)
	}
	if !called {
		t.Fatal("git init runner was not called")
	}
	if gotPath != "/work/new-repo" {
		t.Fatalf("git init path = %q, want /work/new-repo", gotPath)
	}
}

func TestInitializeGitRepositoryForPathReturnsGitUnavailableWithoutRunningInit(t *testing.T) {
	called := false
	err := initializeGitRepositoryForPath("/work/new-repo", func() error {
		return exec.ErrNotFound
	}, func(ctx context.Context, path string) ([]byte, error) {
		called = true
		return []byte("should not run"), nil
	})

	if err == nil {
		t.Fatal("initializeGitRepositoryForPath() error = nil, want error")
	}
	if called {
		t.Fatal("git init runner was called, want skipped")
	}
	if !strings.Contains(err.Error(), "Git is not installed") {
		t.Fatalf("error = %q, want Git is not installed", err.Error())
	}
}

func TestInitializeGitRepositoryForPathReturnsCommandFailure(t *testing.T) {
	err := initializeGitRepositoryForPath("/work/new-repo", gitAvailable, func(ctx context.Context, path string) ([]byte, error) {
		return []byte("fatal: could not create work tree dir"), errors.New("exit status 128")
	})

	if err == nil {
		t.Fatal("initializeGitRepositoryForPath() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "git init failed") {
		t.Fatalf("error = %q, want git init failed", err.Error())
	}
}
