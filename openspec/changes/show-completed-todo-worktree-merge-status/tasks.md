## 1. Data Model And API

- [x] 1.1 Extend Go request/model types so TODO project selections can carry `projectId` and `baseBranch`, while preserving compatibility with existing `projectIds` callers.
- [x] 1.2 Persist `baseBranch` on `TodoProject` for create TODO, update TODO, and add project(s) to TODO flows.
- [x] 1.3 Extend `TodoProjectSnapshot` with `worktreeBranch` and `baseBranch`, and copy `baseBranch` from `TodoProject` when completing a TODO.
- [x] 1.4 Read the worktree branch from the project path during TODO completion and store it in the completed snapshot without blocking terminal cleanup semantics.
- [x] 1.5 Update legacy normalization so old TODO projects and old completed snapshots without branch fields remain loadable and degrade to unknown merge status.

## 2. Git Merge Status Backend

- [x] 2.1 Add Git helpers for reading the current branch and checking whether a worktree branch is contained by a base branch using bounded-time Git commands.
- [x] 2.2 Add a Wails-accessible backend method for completed snapshot merge checks that returns merged, unmerged, or unknown status with a reason.
- [x] 2.3 Handle path unavailable, Git unavailable, non-Git repository, missing branch, deleted branch, and timeout cases as unknown status instead of hard UI failures.
- [x] 2.4 Add Go unit tests for branch parsing, merge-status command outcomes, timeout/error handling, and completion snapshot branch persistence.

## 3. Frontend Selection Flow

- [x] 3.1 Update generated Wails bindings after backend API/model changes.
- [x] 3.2 Update create TODO, edit TODO, and add-project picker state to store selected project/base-branch pairs rather than only project IDs.
- [x] 3.3 Add or wire a base-branch selector for each selected project, defaulting to the selected project branch when available.
- [x] 3.4 Ensure existing project search, selected project tags, removal, read-only completed details, and open TODO project management continue to work.

## 4. Completed View UI

- [x] 4.1 Add frontend state in `App.vue` for asynchronous completed snapshot merge checks, including request generation, stale-response protection, and result caching.
- [x] 4.2 Pass completed snapshot merge status into `ProjectSidebar.vue` and render worktree/base branch text instead of project path in the completed list.
- [x] 4.3 Render a check icon for confirmed merged snapshots and a yellow triangle warning icon for unmerged or unknown snapshots.
- [x] 4.4 Update completed TODO read-only details to show project name and `worktreeBranch -> baseBranch`, with unknown fallback for legacy snapshots.
- [x] 4.5 Keep completed list initial rendering, view switching, sorting, duration display, delete menu, and bulk delete interactions non-blocking while merge checks run.

## 5. Tests And Verification

- [x] 5.1 Add frontend component tests for completed list branch display, merged check icon, unmerged warning icon, unknown legacy warning, and non-blocking loading state.
- [x] 5.2 Add frontend App tests for asynchronous merge-status request lifecycle, stale response handling, and completed TODO data changes.
- [x] 5.3 Add tests for create/edit/add-project flows preserving selected base branches in requests and restored state.
- [x] 5.4 Run Go tests with `go test ./...`.
- [x] 5.5 Run frontend automated tests with the repository's npm test command.
- [x] 5.6 Run OpenSpec validation/status for `show-completed-todo-worktree-merge-status`.
- [x] 5.7 Run automated review for the completed change and address findings.
- [x] 5.8 Run `wails build -tags webkit2_41` to generate the executable.
