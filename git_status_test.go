package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
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

func TestGitBranchForPathReturnsTrimmedBranchName(t *testing.T) {
	branch, err := gitBranchForPath("/work/repo", gitAvailable, func(ctx context.Context, path string) ([]byte, error) {
		if path != "/work/repo" {
			t.Fatalf("path = %q, want /work/repo", path)
		}
		return []byte("todo/fix-login\n"), nil
	})

	if err != nil {
		t.Fatalf("gitBranchForPath() error = %v, want nil", err)
	}
	if branch != "todo/fix-login" {
		t.Fatalf("branch = %q, want todo/fix-login", branch)
	}
}

func TestGitBranchMergedForPathReturnsMergedAndUnmerged(t *testing.T) {
	merged, err := gitBranchMergedForPath("/work/repo", "todo/fix-login", "main", gitAvailable, func(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
		if path != "/work/repo" || worktreeBranch != "todo/fix-login" || baseBranch != "main" {
			t.Fatalf("args = %q/%q/%q, want repo/worktree/base", path, worktreeBranch, baseBranch)
		}
		return []byte(""), nil
	})
	if err != nil {
		t.Fatalf("gitBranchMergedForPath(merged) error = %v, want nil", err)
	}
	if !merged {
		t.Fatal("merged = false, want true")
	}

	errExitOne := exec.Command("sh", "-c", "exit 1").Run()
	var exitErr *exec.ExitError
	if !errors.As(errExitOne, &exitErr) {
		t.Fatalf("exit error = %T, want *exec.ExitError", errExitOne)
	}
	merged, err = gitBranchMergedForPath("/work/repo", "todo/fix-login", "main", gitAvailable, func(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
		return []byte(""), exitErr
	})
	if err != nil {
		t.Fatalf("gitBranchMergedForPath(unmerged) error = %v, want nil", err)
	}
	if merged {
		t.Fatal("merged = true, want false")
	}
}

func TestGitBranchMergedForPathRejectsMissingBranches(t *testing.T) {
	_, err := gitBranchMergedForPath("/work/repo", "", "main", gitAvailable, func(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("gitBranchMergedForPath() error = nil, want missing branch error")
	}
}

func TestGitBranchMergedForPathTreatsUnexpectedExitCodeAsUnknown(t *testing.T) {
	errExitTwo := exec.Command("sh", "-c", "exit 2").Run()
	var exitErr *exec.ExitError
	if !errors.As(errExitTwo, &exitErr) {
		t.Fatalf("exit error = %T, want *exec.ExitError", errExitTwo)
	}

	merged, err := gitBranchMergedForPath("/work/repo", "todo/fix-login", "main", gitAvailable, func(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
		return []byte{}, exitErr
	})

	if err == nil {
		t.Fatal("gitBranchMergedForPath() error = nil, want unexpected exit error")
	}
	if merged {
		t.Fatal("merged = true, want false")
	}
	if !strings.Contains(err.Error(), "git merge-base failed") {
		t.Fatalf("error = %q, want git merge-base failed", err.Error())
	}
}

func TestGitBranchMergedForPathTreatsExitOneAsUnmerged(t *testing.T) {
	errExitOne := exec.Command("sh", "-c", "exit 1").Run()
	var exitErr *exec.ExitError
	if !errors.As(errExitOne, &exitErr) {
		t.Fatalf("exit error = %T, want *exec.ExitError", errExitOne)
	}
	if exitErr.ExitCode() != 1 && os.Getenv("GOOS") != "windows" {
		t.Fatalf("exit code = %d, want 1", exitErr.ExitCode())
	}

	merged, err := gitBranchMergedForPath("/work/repo", "todo/fix-login", "main", gitAvailable, func(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
		return []byte{}, exitErr
	})

	if err != nil {
		t.Fatalf("gitBranchMergedForPath() error = %v, want nil", err)
	}
	if merged {
		t.Fatal("merged = true, want false")
	}
}
