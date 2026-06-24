package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGitWorktreeServicePrepareExistingBranchCreatesIsolatedWorktreeBranch(t *testing.T) {
	repoDir := t.TempDir()
	taskDir := t.TempDir()
	runner := newFakeGitWorktreeRunner()
	runner.branches["develop"] = true
	service := &GitWorktreeService{
		checker: func() error { return nil },
		runner:  runner.run,
		timeout: gitWorktreeTimeout,
	}

	result := service.PrepareWorktree(repoDir, "develop", "frontend-app", taskDir)

	wantBranch := worktreeBranchName("frontend-app", filepath.Base(taskDir))
	if result.BaseBranch != "develop" || result.WorktreeBranch != wantBranch || result.Status != WorktreeStatusReady {
		t.Fatalf("result = %#v, want develop -> %s ready", result, wantBranch)
	}
	if result.WorktreePath != filepath.Join(taskDir, "frontend-app") {
		t.Fatalf("WorktreePath = %q, want task project path", result.WorktreePath)
	}
	if !runner.branches[wantBranch] {
		t.Fatalf("isolated branch %q was not created", wantBranch)
	}
	if !runner.called("branch", wantBranch, "develop") {
		t.Fatalf("commands = %#v, want branch creation", runner.commands)
	}
	if !runner.called("worktree", "add", result.WorktreePath, wantBranch) {
		t.Fatalf("commands = %#v, want worktree add for isolated branch", runner.commands)
	}
}

func TestGitWorktreeServicePrepareNewBranchCreatesItFromDefaultBranch(t *testing.T) {
	repoDir := t.TempDir()
	taskDir := t.TempDir()
	runner := newFakeGitWorktreeRunner()
	runner.branches["main"] = true
	service := &GitWorktreeService{
		checker: func() error { return nil },
		runner:  runner.run,
		timeout: gitWorktreeTimeout,
	}

	result := service.PrepareWorktree(repoDir, "feature/login-fix", "frontend-app", taskDir)

	if result.BaseBranch != "main" || result.WorktreeBranch != "feature/login-fix" || result.Status != WorktreeStatusReady {
		t.Fatalf("result = %#v, want feature branch from main ready", result)
	}
	if !runner.called("worktree", "add", "-b", "feature/login-fix", result.WorktreePath, "main") {
		t.Fatalf("commands = %#v, want new branch worktree add from main", runner.commands)
	}
}

func TestGitWorktreeServiceRecordsFailuresForMissingGitAndConflictingPath(t *testing.T) {
	repoDir := t.TempDir()
	taskDir := t.TempDir()
	service := &GitWorktreeService{
		checker: func() error { return errors.New("git missing") },
		runner:  newFakeGitWorktreeRunner().run,
		timeout: gitWorktreeTimeout,
	}

	result := service.PrepareWorktree(repoDir, "main", "frontend-app", taskDir)
	if result.Status != WorktreeStatusFailed || !strings.Contains(result.Error, "Git is not installed") {
		t.Fatalf("missing git result = %#v, want failed Git message", result)
	}

	runner := newFakeGitWorktreeRunner()
	runner.branches["main"] = true
	service = &GitWorktreeService{
		checker: func() error { return nil },
		runner:  runner.run,
		timeout: gitWorktreeTimeout,
	}
	conflictPath := filepath.Join(taskDir, "frontend-app")
	if err := os.MkdirAll(conflictPath, 0o755); err != nil {
		t.Fatalf("mkdir conflict path: %v", err)
	}

	result = service.PrepareWorktree(repoDir, "main", "frontend-app", taskDir)
	if result.Status != WorktreeStatusFailed || !strings.Contains(result.Error, "worktree path already exists") {
		t.Fatalf("conflict result = %#v, want failed conflict message", result)
	}
}

type fakeGitWorktreeRunner struct {
	branches map[string]bool
	commands [][]string
}

func newFakeGitWorktreeRunner() *fakeGitWorktreeRunner {
	return &fakeGitWorktreeRunner{
		branches: map[string]bool{"main": true},
	}
}

func (runner *fakeGitWorktreeRunner) run(ctx context.Context, repoPath string, args ...string) ([]byte, error) {
	_ = ctx
	_ = repoPath
	command := append([]string(nil), args...)
	runner.commands = append(runner.commands, command)

	if reflect.DeepEqual(command, []string{"rev-parse", "--is-inside-work-tree"}) {
		return []byte("true\n"), nil
	}
	if reflect.DeepEqual(command, []string{"symbolic-ref", "--short", "refs/remotes/origin/HEAD"}) {
		return nil, errors.New("no origin head")
	}
	if len(command) == 4 && reflect.DeepEqual(command[:3], []string{"rev-parse", "--verify", "--quiet"}) {
		if runner.branches[command[3]] {
			return []byte(command[3] + "\n"), nil
		}
		return nil, errors.New("missing branch")
	}
	if len(command) == 3 && command[0] == "branch" {
		runner.branches[command[1]] = true
		return []byte{}, nil
	}
	if len(command) >= 3 && reflect.DeepEqual(command[:2], []string{"worktree", "add"}) {
		return []byte{}, nil
	}
	if reflect.DeepEqual(command, []string{"worktree", "list", "--porcelain"}) {
		return []byte{}, nil
	}
	return nil, errors.New("unexpected git command: " + strings.Join(command, " "))
}

func (runner *fakeGitWorktreeRunner) called(args ...string) bool {
	for _, command := range runner.commands {
		if reflect.DeepEqual(command, args) {
			return true
		}
	}
	return false
}
