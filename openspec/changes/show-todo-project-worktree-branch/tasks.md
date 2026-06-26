## 1. Backend Workspace Initialization Ordering

- [x] 1.1 Add a backend helper that determines whether a TODO has no associated projects or all associated TODO project worktrees are `ready`.
- [x] 1.2 Remove early initialization-file writes from task workspace directory creation paths that can run before worktree preparation is complete.
- [x] 1.3 Ensure selected initialization files are written only after worktree preparation records all associated TODO project worktrees as `ready`.
- [x] 1.4 Preserve existing behavior for TODOs without associated projects by writing selected initialization files after the task workspace directory is created.
- [x] 1.5 Add or update Go tests covering initialization after all worktrees are ready, delayed initialization while any worktree is not ready, no-project TODO initialization, and existing-file preservation.

## 2. Frontend Todo Project Branch Display

- [x] 2.1 Add `App.vue` state for TODO project Git status keyed by `todoProjectId`, including request dedupe and stale-response protection.
- [x] 2.2 Refresh TODO project branch state when relevant TODO project rows become visible, when a TODO project is selected, when the window regains focus, and when a terminal command ends for that TODO project.
- [x] 2.3 Pass current TODO project branch labels from `App.vue` into `ProjectSidebar.vue` without making `ProjectSidebar` call Wails APIs directly.
- [x] 2.4 Render left-sidebar TODO project names as `项目名称(分支名称)` only when a live ready-worktree branch is available, and keep the top workspace heading unchanged.
- [x] 2.5 Ensure the rendered branch comes from live worktree Git status rather than the persisted `worktreeBranch` value.

## 3. Status Bar Empty State

- [x] 3.1 Change the no-active-project Git status behavior so the status bar keeps its layout height but renders no Git status chip.
- [x] 3.2 Remove or update frontend assertions that expect `No project`, replacing them with assertions that no Git status chip is shown when no project or TODO project is selected.
- [x] 3.3 Add TODO task-terminal Git status context so the status bar reads the TODO workspace root instead of the previously selected project.
- [x] 3.4 Hide Git status chips for TODO task terminals when the TODO workspace root itself is not a Git repository.

## 4. Automated Tests

- [x] 4.1 Add `ProjectSidebar` or `App.vue` tests for branch suffix display, branch suffix omission on unavailable status, and top-heading non-regression.
- [x] 4.2 Add `App.vue` tests proving terminal command completion refreshes the relevant TODO project branch label after a worktree branch switch.
- [x] 4.3 Add frontend tests proving the left list uses live Git status branch instead of stored `worktreeBranch`.
- [x] 4.4 Run frontend automated tests with `cd frontend && npm test`.
- [x] 4.5 Run backend automated tests with `go test ./...`.
- [x] 4.6 Run OpenSpec validation for `show-todo-project-worktree-branch`.
- [x] 4.7 Add frontend and backend regression tests for TODO task-terminal Git status, including nested repository non-detection.

## 5. Review And Build

- [x] 5.1 Perform an automated code review pass focused on data-flow correctness, request race handling, initialization ordering, and test coverage.
- [x] 5.2 Address review findings or document any accepted residual risk in the implementation notes.
- [x] 5.3 Run `wails build -tags webkit2_41` to generate the executable file.
