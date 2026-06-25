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

func TestGitWorktreeServicePrepareNewBranchCreatesBaseAndIsolatedWorktreeBranches(t *testing.T) {
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

	wantBranch := worktreeBranchName("frontend-app", filepath.Base(taskDir))
	if result.BaseBranch != "feature/login-fix" || result.WorktreeBranch != wantBranch || result.Status != WorktreeStatusReady {
		t.Fatalf("result = %#v, want feature/login-fix -> %s ready", result, wantBranch)
	}
	if !runner.branches["feature/login-fix"] {
		t.Fatalf("base branch %q was not created", "feature/login-fix")
	}
	if !runner.branches[wantBranch] {
		t.Fatalf("isolated branch %q was not created", wantBranch)
	}
	if !runner.called("branch", "feature/login-fix", "main") {
		t.Fatalf("commands = %#v, want base branch creation from main", runner.commands)
	}
	if !runner.called("branch", wantBranch, "feature/login-fix") {
		t.Fatalf("commands = %#v, want isolated branch creation from base branch", runner.commands)
	}
	if !runner.called("worktree", "add", result.WorktreePath, wantBranch) {
		t.Fatalf("commands = %#v, want worktree add for isolated branch", runner.commands)
	}
	if !runner.calledBefore(
		[]string{"branch", "feature/login-fix", "main"},
		[]string{"branch", wantBranch, "feature/login-fix"},
	) {
		t.Fatalf("commands = %#v, want base branch before isolated branch", runner.commands)
	}
}

func TestGitWorktreeServicePrepareNewBranchStopsWhenBaseBranchCreationFails(t *testing.T) {
	repoDir := t.TempDir()
	taskDir := t.TempDir()
	runner := newFakeGitWorktreeRunner()
	runner.branchCreateErrors["feature/login-fix"] = errors.New("invalid branch name")
	service := &GitWorktreeService{
		checker: func() error { return nil },
		runner:  runner.run,
		timeout: gitWorktreeTimeout,
	}

	result := service.PrepareWorktree(repoDir, "feature/login-fix", "frontend-app", taskDir)

	if result.Status != WorktreeStatusFailed || !strings.Contains(result.Error, "create base branch failed") {
		t.Fatalf("result = %#v, want failed base branch creation message", result)
	}
	wantBranch := worktreeBranchName("frontend-app", filepath.Base(taskDir))
	if runner.branches["feature/login-fix"] || runner.branches[wantBranch] {
		t.Fatalf("branches = %#v, want no base or isolated branch created", runner.branches)
	}
	if runner.called("worktree", "add", filepath.Join(taskDir, "frontend-app"), wantBranch) {
		t.Fatalf("commands = %#v, want no worktree creation after base branch failure", runner.commands)
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
	branches           map[string]bool
	branchCreateErrors map[string]error
	commands           [][]string
}

func newFakeGitWorktreeRunner() *fakeGitWorktreeRunner {
	return &fakeGitWorktreeRunner{
		branches:           map[string]bool{"main": true},
		branchCreateErrors: map[string]error{},
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
		if err := runner.branchCreateErrors[command[1]]; err != nil {
			return nil, err
		}
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

func (runner *fakeGitWorktreeRunner) calledBefore(first []string, second []string) bool {
	firstIndex := -1
	secondIndex := -1
	for index, command := range runner.commands {
		if firstIndex < 0 && reflect.DeepEqual(command, first) {
			firstIndex = index
		}
		if secondIndex < 0 && reflect.DeepEqual(command, second) {
			secondIndex = index
		}
	}
	return firstIndex >= 0 && secondIndex >= 0 && firstIndex < secondIndex
}
