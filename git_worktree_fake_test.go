package main

import (
	"os"
	"path/filepath"
)

// readyWorktreePreparer is a test double for TodoWorktreePreparer that simulates
// successful worktree preparation without invoking git. It materialises the
// worktree directory on disk so directory checks pass.
type readyWorktreePreparer struct{}

func newReadyWorktreePreparer() *readyWorktreePreparer {
	return &readyWorktreePreparer{}
}

func (preparer *readyWorktreePreparer) PrepareWorktree(repoPath, requestedBranch, projectName, taskWorkspaceDir string) WorktreePrepareResult {
	branch := requestedBranch
	if branch == "" {
		branch = "main"
	}
	worktreePath := filepath.Join(taskWorkspaceDir, worktreeDirectoryName(projectName))
	_ = os.MkdirAll(worktreePath, 0o755)
	return WorktreePrepareResult{
		BaseBranch:     branch,
		WorktreeBranch: worktreeBranchName(projectName, filepath.Base(taskWorkspaceDir)),
		WorktreePath:   worktreePath,
		Status:         WorktreeStatusReady,
	}
}
