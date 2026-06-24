package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const gitWorktreeTimeout = 10 * time.Second

// gitWorktreeCommandRunner runs a git command rooted at repoPath with the
// supplied arguments and returns combined output. It is abstracted so the
// worktree service can be tested without a real git binary.
type gitWorktreeCommandRunner func(ctx context.Context, repoPath string, args ...string) ([]byte, error)

// TodoWorktreePreparer prepares an isolated Git worktree for a TODO project.
// It is implemented by GitWorktreeService and abstracted so the App layer can
// be tested without performing real git operations.
type TodoWorktreePreparer interface {
	PrepareWorktree(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult
}

// GitWorktreeService performs read-only repository inspection and worktree
// creation for TODO projects. All git interaction is funnelled through an
// injectable checker and runner so behaviour is deterministic under test.
type GitWorktreeService struct {
	checker gitCommandChecker
	runner  gitWorktreeCommandRunner
	timeout time.Duration
}

// NewGitWorktreeService returns a worktree service backed by the real git
// binary.
func NewGitWorktreeService() *GitWorktreeService {
	return &GitWorktreeService{
		checker: gitCommandAvailable,
		runner:  runGitWorktreeCommand,
		timeout: gitWorktreeTimeout,
	}
}

// WorktreePrepareResult describes the outcome of preparing a worktree for a
// TODO project. Status is WorktreeStatusReady on success or
// WorktreeStatusFailed when git is unavailable, the path is not a repository,
// the branch does not resolve or a checkout conflict occurs.
type WorktreePrepareResult struct {
	BaseBranch     string
	WorktreeBranch string
	WorktreePath   string
	Status         string
	Error          string
}

func runGitWorktreeCommand(ctx context.Context, repoPath string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	cmd := newBackgroundCommand(ctx, "git", fullArgs...)
	return cmd.CombinedOutput()
}

func (service *GitWorktreeService) run(repoPath string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), service.timeout)
	defer cancel()
	return service.runner(ctx, repoPath, args...)
}

// IsGitRepository reports whether the path is inside a git work tree.
func (service *GitWorktreeService) IsGitRepository(repoPath string) bool {
	if !directoryAvailable(repoPath) {
		return false
	}
	output, err := service.run(repoPath, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// DefaultBranch resolves the repository's default branch. It prefers the
// remote HEAD symbolic ref, then falls back to "main" or "master" if either
// exists locally, and finally to "main".
func (service *GitWorktreeService) DefaultBranch(repoPath string) string {
	if output, err := service.run(repoPath, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed != "" {
			return strings.TrimPrefix(trimmed, "origin/")
		}
	}
	for _, candidate := range []string{"main", "master"} {
		if service.BranchExists(repoPath, candidate) {
			return candidate
		}
	}
	return "main"
}

// BranchExists reports whether a local or remote-tracking branch with the
// given name exists in the repository.
func (service *GitWorktreeService) BranchExists(repoPath string, branch string) bool {
	if strings.TrimSpace(branch) == "" {
		return false
	}
	_, err := service.run(repoPath, "rev-parse", "--verify", "--quiet", branch)
	return err == nil
}

// worktreeBranchName builds a deterministic, unique-enough isolated worktree
// branch name for a TODO project. The name embeds the task workspace
// directory name (an md5 of title+description) so two tasks with the same
// project do not collide, while remaining readable.
func worktreeBranchName(projectName, taskWorkspaceDirName string) string {
	sanitized := sanitizeGitBranchSegment(projectName)
	if sanitized == "" {
		sanitized = "project"
	}
	suffix := taskWorkspaceDirName
	if len(suffix) > 12 {
		suffix = suffix[:12]
	}
	return fmt.Sprintf("todo-workspace/%s/%s", sanitized, suffix)
}

// worktreeDirectoryName returns the directory name used for a project's
// worktree inside the task workspace directory.
func worktreeDirectoryName(projectName string) string {
	sanitized := sanitizeGitBranchSegment(projectName)
	if sanitized == "" {
		sanitized = "project"
	}
	return sanitized
}

// sanitizeGitBranchSegment keeps characters that are safe for both git branch
// names and filesystem directory names: alphanumerics, dashes, underscores and
// dots. Everything else becomes a dash; leading/trailing dashes and dots are
// trimmed.
func sanitizeGitBranchSegment(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-' || r == '_' || r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}
	segment := builder.String()
	segment = strings.Trim(segment, "-.")
	// Collapse runs of dashes for tidier names.
	for strings.Contains(segment, "--") {
		segment = strings.ReplaceAll(segment, "--", "-")
	}
	return segment
}

// PrepareWorktree creates a Git worktree for a TODO project inside the task
// workspace directory. The behaviour follows the selected branch:
//
//   - An empty requestedBranch defaults to the repository default branch,
//     which is then treated as an existing base branch.
//   - An existing branch becomes the base branch; a new isolated worktree
//     branch is created from it and the worktree checks out the isolated
//     branch.
//   - A non-existent branch is created from the default branch and the
//     worktree checks out that new branch directly.
//
// Failures (git unavailable, not a repository, branch conflicts) are recorded
// as a failed result rather than returned as errors so a single TODO can have
// a mix of ready and failed projects.
func (service *GitWorktreeService) PrepareWorktree(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult {
	failed := func(message string) WorktreePrepareResult {
		return WorktreePrepareResult{Status: WorktreeStatusFailed, Error: message}
	}

	if !directoryAvailable(repoPath) {
		return failed("project path is unavailable")
	}
	if err := service.checker(); err != nil {
		return failed("Git is not installed")
	}
	if !service.IsGitRepository(repoPath) {
		return failed("project is not a Git repository")
	}

	defaultBranch := service.DefaultBranch(repoPath)
	branch := strings.TrimSpace(requestedBranch)
	if branch == "" {
		branch = defaultBranch
	}

	worktreePath := filepath.Join(taskWorkspaceDir, worktreeDirectoryName(projectName))

	// Refuse if the target path is already occupied by something that is not
	// an existing worktree directory for this project.
	if info, err := os.Stat(worktreePath); err == nil {
		if !info.IsDir() {
			return failed(fmt.Sprintf("worktree path already exists and is not a directory: %s", worktreePath))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return failed(err.Error())
	}

	if service.BranchExists(repoPath, branch) {
		// Existing branch: create an isolated worktree branch from it.
		worktreeBranch := worktreeBranchName(projectName, filepath.Base(taskWorkspaceDir))
		if !service.BranchExists(repoPath, worktreeBranch) {
			if _, err := service.run(repoPath, "branch", worktreeBranch, branch); err != nil {
				return failed(fmt.Sprintf("create worktree branch failed: %s", gitFirstErrorLine(err)))
			}
		}
		if err := service.addWorktreeForBranch(repoPath, worktreePath, worktreeBranch, branch); err != nil {
			return failed(err.Error())
		}
		return WorktreePrepareResult{
			BaseBranch:     branch,
			WorktreeBranch: worktreeBranch,
			WorktreePath:   worktreePath,
			Status:         WorktreeStatusReady,
		}
	}

	// Non-existent branch: create it from the default branch and check it out
	// directly in the worktree.
	if err := service.addWorktreeForBranch(repoPath, worktreePath, branch, defaultBranch); err != nil {
		return failed(err.Error())
	}
	return WorktreePrepareResult{
		BaseBranch:     defaultBranch,
		WorktreeBranch: branch,
		WorktreePath:   worktreePath,
		Status:         WorktreeStatusReady,
	}
}

// addWorktreeForBranch creates a worktree at worktreePath on branch, creating
// the branch from startPoint when it does not yet exist. If the worktree path
// already holds a worktree on branch (an idempotent retry) it succeeds.
func (service *GitWorktreeService) addWorktreeForBranch(repoPath, worktreePath, branch, startPoint string) error {
	if directoryAvailable(worktreePath) {
		if worktreeOnBranch(service, repoPath, worktreePath, branch) {
			return nil
		}
		return fmt.Errorf("worktree path already exists: %s", worktreePath)
	}
	args := []string{"worktree", "add"}
	if service.BranchExists(repoPath, branch) {
		args = append(args, worktreePath, branch)
	} else {
		args = append(args, "-b", branch, worktreePath, startPoint)
	}
	if _, err := service.run(repoPath, args...); err != nil {
		return fmt.Errorf("create worktree failed: %s", gitFirstErrorLine(err))
	}
	return nil
}

// worktreeOnBranch reports whether the worktree at worktreePath is checked out
// on the given branch.
func worktreeOnBranch(service *GitWorktreeService, repoPath, worktreePath, branch string) bool {
	output, err := service.run(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return false
	}
	worktreeAbs, _ := filepath.Abs(worktreePath)
	currentPath := ""
	for _, line := range strings.Split(string(output), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			currentPath = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		case strings.HasPrefix(line, "branch "):
			branchName := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			branchShort := strings.TrimPrefix(branchName, "refs/heads/")
			if branchShort == branch && samePath(currentPath, worktreeAbs) {
				return true
			}
		}
	}
	return false
}

func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	absA, err := filepath.Abs(a)
	if err != nil {
		absA = a
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		absB = b
	}
	return absA == absB
}

// gitFirstErrorLine extracts the most useful single line from a git command
// error's combined output for display.
func gitFirstErrorLine(err error) string {
	message := err.Error()
	if exitErr, ok := err.(*exec.ExitError); ok {
		output := strings.TrimSpace(string(exitErr.Stderr))
		if output != "" {
			message = output
		}
	}
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return message
}
