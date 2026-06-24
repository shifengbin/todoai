package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const gitStatusTimeout = 2 * time.Second

var errGitUnavailable = errors.New("Git is not installed")

type gitCommandChecker func() error
type gitStatusRunner func(context.Context, string) ([]byte, error)
type gitInitRunner func(context.Context, string) ([]byte, error)
type gitBranchesRunner func(context.Context, string) ([]byte, error)

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

func initializeGitRepositoryForPath(path string, checker gitCommandChecker, runner gitInitRunner) error {
	if err := checker(); err != nil {
		return gitUnavailableError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()

	output, err := runner(ctx, path)
	if err != nil {
		if errorsIsDeadlineExceeded(ctx.Err()) {
			return fmt.Errorf("git init timed out")
		}
		if message := strings.TrimSpace(string(output)); message != "" {
			return fmt.Errorf("git init failed: %w: %s", err, message)
		}
		return fmt.Errorf("git init failed: %w", err)
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
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "branch", "--format=%(refname:short)", "--list", "--remotes")
	return cmd.CombinedOutput()
}

func runGitInitCommand(ctx context.Context, path string) ([]byte, error) {
	cmd := newBackgroundCommand(ctx, "git", "-C", path, "init")
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
