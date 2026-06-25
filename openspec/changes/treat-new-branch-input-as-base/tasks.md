## 1. Backend Worktree Semantics

- [x] 1.1 Update `git_worktree_test.go` so the nonexistent branch scenario expects the user input branch to become `BaseBranch` and the generated `todo-workspace/...` branch to become `WorktreeBranch`.
- [x] 1.2 Add or adjust a test assertion for Git command ordering: create the missing base branch from the default branch before creating the isolated worktree branch.
- [x] 1.3 Change `GitWorktreeService.PrepareWorktree` so missing requested branches are first created from `DefaultBranch`, then handled by the existing isolated worktree branch path.
- [x] 1.4 Ensure missing base branch creation failures return `WorktreeStatusFailed` with a clear error and do not create the worktree.

## 2. State And Documentation Integration

- [x] 2.1 Verify `RecordTodoWorkspace` continues to persist `BaseBranch` and `WorktreeBranch` from `WorktreePrepareResult` without model changes.
- [x] 2.2 Verify generated TODO workspace `README.md` project lines show the new base branch and isolated worktree branch values after worktree preparation.
- [x] 2.3 Confirm front-end TODO creation, TODO detail, and add-project payloads still submit user input through `projects[].baseBranch` without API changes.

## 3. Automated Tests And Review

- [x] 3.1 Run the focused Go tests covering git worktree behavior.
- [x] 3.2 Run the broader Go test suite for the repository.
- [x] 3.3 Run the affected client automated tests for TODO branch input payload behavior.
- [x] 3.4 Run OpenSpec validation for `treat-new-branch-input-as-base`.
- [x] 3.5 Run an automated code review pass over the implementation and address correctness, regression, or test coverage findings.

## 4. Packaging

- [x] 4.1 Run `wails build -tags webkit2_41` to generate the executable.
