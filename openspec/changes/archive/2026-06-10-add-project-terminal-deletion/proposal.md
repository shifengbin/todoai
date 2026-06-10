## Why

Users can add projects and create multiple terminal sessions, but they cannot remove items they no longer need. This leaves stale projects and ended or accidental terminal sessions visible for the rest of the app session.

## What Changes

- Add a project delete action in the project tree.
- Add a terminal delete action for terminal rows in the project tree.
- Deleting a project removes only the app's saved project record; it does not delete the directory or any files on disk.
- Deleting a project closes and removes all runtime terminal sessions owned by that project.
- Deleting a terminal closes and removes that runtime terminal session.
- Require confirmation before deleting a project. Terminal deletion does not require confirmation.
- Return an updated project/terminal state after each delete so the UI can update active project and active terminal selection.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `project-workspace`: Allow users to remove opened projects from the app without deleting files from disk.
- `embedded-shell-sessions`: Allow users to remove runtime terminal sessions and close associated PTY processes.

## Impact

- Backend project persistence: add project deletion and active-project fallback behavior.
- Backend shell session management: add terminal deletion and project-terminal cleanup behavior.
- Wails API bindings: expose delete methods for projects and terminals.
- Frontend app state: call delete APIs, apply returned state, clear stale terminal menu/session state.
- Project sidebar UI: add delete controls for project and terminal rows.
- Tests: cover Go manager/API behavior plus Vue sidebar and app event flows.
