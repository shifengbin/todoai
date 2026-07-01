package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const gitStatusTimeout = 2 * time.Second

var (
	errGitUnavailable           = errors.New("Git is not installed")
	errGitWorktreePathMissing   = errors.New("git worktree path is missing")
	errGitWorktreeBranchMissing = errors.New("git worktree branch is missing")
)

type gitCommandChecker func() error
type gitStatusRunner func(context.Context, string) ([]byte, error)
type gitInitRunner func(context.Context, string, ...string) ([]byte, error)
type gitBranchesRunner func(context.Context, string) ([]byte, error)
type gitBranchRunner func(context.Context, string) ([]byte, error)
type gitBranchMergedRunner func(context.Context, string, string, string) ([]byte, error)

type GitStatus struct {
	ProjectID       string `json:"projectId,omitempty"`
	IsRepo          bool   `json:"isRepo"`
	Branch          string `json:"branch"`
	ChangedCount    int    `json:"changedCount"`
	StagedCount     int    `json:"stagedCount"`
	UnstagedCount   int    `json:"unstagedCount"`
	UntrackedCount  int    `json:"untrackedCount"`
	Ahead           int    `json:"ahead"`
	Behind          int    `json:"behind"`
	PathUnavailable bool   `json:"pathUnavailable,omitempty"`
	GitUnavailable  bool   `json:"gitUnavailable,omitempty"`
	WorktreeCleared bool   `json:"worktreeCleared,omitempty"`
}

func parseGitStatusPorcelainV2(output string) GitStatus {
	status := GitStatus{IsRepo: true}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# branch.head ") {
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
			continue
		}
		if strings.HasPrefix(line, "# branch.ab ") {
			parseBranchAheadBehind(&status, strings.TrimPrefix(line, "# branch.ab "))
			continue
		}
		if strings.HasPrefix(line, "? ") {
			status.ChangedCount++
			status.UntrackedCount++
			continue
		}
		if strings.HasPrefix(line, "! ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "1", "2", "u":
			countTrackedGitStatus(&status, fields[1])
		}
	}
	return status
}

func queryGitStatus(path string) (GitStatus, error) {
	return gitStatusForPath(path, gitCommandAvailable, runGitStatusCommand)
}

func queryGitBranches(path string) ([]string, error) {
	return gitBranchesForPath(path, gitCommandAvailable, runGitBranchesCommand)
}

func initializeGitRepository(path string) error {
	return initializeGitRepositoryForPath(path, gitCommandAvailable, runGitInitCommand)
}

func pathHasGitRepositoryMetadata(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular()
}

func queryGitBranch(path string) (string, error) {
	return gitBranchForPath(path, gitCommandAvailable, runGitBranchCommand)
}

func queryGitBranchMerged(path string, worktreeBranch string, baseBranch string) (bool, error) {
	return gitBranchMergedForPath(path, worktreeBranch, baseBranch, gitCommandAvailable, runGitBranchMergedCommand)
}

func gitStatusForPath(path string, checker gitCommandChecker, runner gitStatusRunner) (GitStatus, error) {
	if err := checker(); err != nil {
		return GitStatus{GitUnavailable: true}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	output, err := runner(ctx, path)
	if err != nil {
		if isNotGitRepositoryOutput(string(output)) {
			return GitStatus{IsRepo: false}, nil
		}
		if errorsIsDeadlineExceeded(ctx.Err()) {
			return GitStatus{}, fmt.Errorf("git status timed out")
		}
		return GitStatus{}, fmt.Errorf("git status failed: %w", err)
	}
	return parseGitStatusPorcelainV2(string(output)), nil
}

func gitBranchesForPath(path string, checker gitCommandChecker, runner gitBranchesRunner) ([]string, error) {
	if err := checker(); err != nil {
		return nil, gitUnavailableError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	output, err := runner(ctx, path)
	if err != nil {
		if errorsIsDeadlineExceeded(ctx.Err()) {
			return nil, fmt.Errorf("git branch list timed out")
		}
		return nil, fmt.Errorf("git branch list failed: %w", err)
	}
	return parseGitBranchList(string(output)), nil
}

func gitBranchForPath(path string, checker gitCommandChecker, runner gitBranchRunner) (string, error) {
	if err := checker(); err != nil {
		return "", gitUnavailableError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	output, err := runner(ctx, path)
	if err != nil {
		if errorsIsDeadlineExceeded(ctx.Err()) {
			return "", fmt.Errorf("git branch timed out")
		}
		return "", fmt.Errorf("git branch failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func gitBranchMergedForPath(path string, worktreeBranch string, baseBranch string, checker gitCommandChecker, runner gitBranchMergedRunner) (bool, error) {
	worktreeBranch = strings.TrimSpace(worktreeBranch)
	baseBranch = strings.TrimSpace(baseBranch)
	if worktreeBranch == "" || baseBranch == "" {
		return false, errors.New("missing branch")
	}
	if err := checker(); err != nil {
		return false, gitUnavailableError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	output, err := runner(ctx, path, worktreeBranch, baseBranch)
	if err == nil {
		return true, nil
	}
	if errorsIsDeadlineExceeded(ctx.Err()) {
		return false, fmt.Errorf("git merge-base timed out")
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && len(output) == 0 {
		return false, nil
	}
	if message := strings.TrimSpace(string(output)); message != "" {
		if specialErr := classifyGitMergeBaseError(message, worktreeBranch); specialErr != nil {
			return false, specialErr
		}
		return false, fmt.Errorf("git merge-base failed: %w: %s", err, message)
	}
	return false, fmt.Errorf("git merge-base failed: %w", err)
}

func classifyGitMergeBaseError(message string, worktreeBranch string) error {
	normalized := strings.ToLower(message)
	switch {
	case strings.Contains(normalized, "cannot change to") &&
		(strings.Contains(normalized, "no such file") || strings.Contains(normalized, "not a directory")):
		return errGitWorktreePathMissing
	case strings.Contains(normalized, "not a valid object name") &&
		strings.Contains(normalized, strings.ToLower(worktreeBranch)):
		return errGitWorktreeBranchMissing
	default:
		return nil
	}
}

func initializeGitRepositoryForPath(path string, checker gitCommandChecker, runner gitInitRunner) error {
	if err := checker(); err != nil {
		return gitUnavailableError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	steps := []struct {
		label string
		args  []string
	}{
		{label: "git init", args: []string{"init"}},
		{label: "git add", args: []string{"add", "-A"}},
		{label: "git commit", args: []string{"commit", "-m", "chore: initial commit"}},
	}
	for _, step := range steps {
		output, err := runner(ctx, path, step.args...)
		if err != nil {
			if errorsIsDeadlineExceeded(ctx.Err()) {
				return fmt.Errorf("%s timed out", step.label)
			}
			if message := strings.TrimSpace(string(output)); message != "" {
				return fmt.Errorf("%s failed: %w: %s", step.label, err, message)
			}
			return fmt.Errorf("%s failed: %w", step.label, err)
		}
	}
	return nil
}

func gitCommandAvailable() error {
	_, err := exec.LookPath("git")
	return err
}

func gitUnavailableError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", errGitUnavailable, err)
}

func runGitStatusCommand(ctx context.Context, path string) ([]byte, error) {
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "status", "--porcelain=v2", "--branch")
	return cmd.CombinedOutput()
}

func runGitBranchesCommand(ctx context.Context, path string) ([]byte, error) {
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "for-each-ref", "--format=%(refname:short)", "refs/heads", "refs/remotes")
	return cmd.CombinedOutput()
}

func runGitInitCommand(ctx context.Context, path string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"-C", path}, args...)
	cmd := newBackgroundCommand(ctx, "git", fullArgs...)
	return cmd.CombinedOutput()
}

func runGitBranchCommand(ctx context.Context, path string) ([]byte, error) {
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "branch", "--show-current")
	return cmd.CombinedOutput()
}

func runGitBranchMergedCommand(ctx context.Context, path string, worktreeBranch string, baseBranch string) ([]byte, error) {
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "merge-base", "--is-ancestor", worktreeBranch, baseBranch)
	return cmd.CombinedOutput()
}

func isNotGitRepositoryOutput(output string) bool {
	return strings.Contains(strings.ToLower(output), "not a git repository")
}

func errorsIsDeadlineExceeded(err error) bool {
	return err == context.DeadlineExceeded
}

func parseBranchAheadBehind(status *GitStatus, value string) {
	for _, field := range strings.Fields(value) {
		if strings.HasPrefix(field, "+") {
			status.Ahead, _ = strconv.Atoi(strings.TrimPrefix(field, "+"))
		}
		if strings.HasPrefix(field, "-") {
			status.Behind, _ = strconv.Atoi(strings.TrimPrefix(field, "-"))
		}
	}
}

func parseGitBranchList(output string) []string {
	seen := map[string]bool{}
	branches := []string{}
	for _, line := range strings.Split(output, "\n") {
		branch := strings.TrimSpace(strings.TrimPrefix(line, "remotes/"))
		if branch == "" || strings.HasSuffix(branch, "/HEAD") || seen[branch] {
			continue
		}
		seen[branch] = true
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}

func countTrackedGitStatus(status *GitStatus, xy string) {
	if xy == "" || xy == ".." {
		return
	}
	status.ChangedCount++
	if xy[0] != '.' {
		status.StagedCount++
	}
	if len(xy) > 1 && xy[1] != '.' {
		status.UnstagedCount++
	}
}
