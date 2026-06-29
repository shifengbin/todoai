## 1. CLI Entry Point

- [x] 1.1 Add `todoai list --done` dispatch in `main.go` before `NewApp()` and `wails.Run`.
- [x] 1.2 Keep existing `todoai claude-hook` behavior unchanged.
- [x] 1.3 Add a testable CLI command runner that accepts arguments, working directory, stdout, stderr, and config path dependencies.
- [x] 1.4 Define stable exit codes and messages for success, unknown project directory, and unsupported command arguments.

## 2. Project Resolution And Completed Todo Data

- [x] 2.1 Add a helper that loads TodoAI recent workspaces from the app config directory and reads each available workspace `.data/projects.json`.
- [x] 2.2 Resolve the current working directory to a TodoAI project by matching project root paths and child paths.
- [x] 2.3 Include completed TODO project snapshot paths in matching so worktree snapshot directories can resolve back to the current project context.
- [x] 2.4 Select the most recent matching workspace when multiple workspace records match the same current directory.
- [x] 2.5 Collect only `completed` TODO rows for the matched project from `Todo.ProjectSnapshots`.
- [x] 2.6 Format output as a stable JSON array with task name, worktree branch, and base branch, using `-` for missing branch fields.
- [x] 2.7 Return a successful empty JSON array when the matched project has no completed TODO rows.

## 3. Automated Tests

- [x] 3.1 Add Go tests proving `todoai list --done` runs through the CLI path without starting Wails GUI.
- [x] 3.2 Add Go tests for project-root and project-child-directory resolution.
- [x] 3.3 Add Go tests proving `not-started` and `in-progress` TODOs are excluded while `completed` TODOs are listed.
- [x] 3.4 Add Go tests for JSON branch output from completed snapshots and `-` placeholders for missing branch fields.
- [x] 3.5 Add Go tests for unknown project directory errors and no-completed-todos empty state.
- [x] 3.6 Add Go tests for resolving a Git linked worktree child directory back to the source project.
- [x] 3.7 Run backend automated tests with `go test ./...`.
- [x] 3.8 Run frontend automated tests with `cd frontend && npm test`.

## 4. Review And Build

- [x] 4.1 Run OpenSpec validation for `add-cli-list-done`.
- [x] 4.2 Perform an automated code review pass focused on CLI routing, workspace matching, persisted data compatibility, and test coverage.
- [x] 4.3 Address review findings or document accepted residual risk in implementation notes.
- [x] 4.4 Run `wails build -tags webkit2_41` to generate the executable file.
