## Context

The app currently supports a persisted project list and runtime terminal sessions under each project. Project persistence is owned by `ProjectManager`, while PTY-backed terminal lifecycle is owned by `ShellSessionManager`. The frontend renders both as a project tree through `ProjectSidebar`.

Users can create projects and additional terminals, but they cannot remove stale projects or accidental terminals. Deletion needs to keep persisted project state, runtime terminal state, active selection, and PTY cleanup consistent.

## Goals / Non-Goals

**Goals:**

- Remove opened projects from the app without deleting their directories or files.
- Close and remove all runtime terminal sessions owned by a deleted project.
- Remove individual runtime terminal sessions and close their PTY process when running.
- Keep active project and active terminal IDs valid after deletion.
- Add discoverable project-tree delete controls with project delete confirmation.
- Cover behavior with backend manager/API tests and frontend component/app tests.

**Non-Goals:**

- Delete project directories or files from disk.
- Persist deleted runtime terminal history, scrollback, or terminal labels.
- Add undo, recycle bin, or soft-delete behavior.
- Add bulk delete, drag/drop, or terminal close confirmation.

## Decisions

### Backend owns deletion state transitions

Add explicit app APIs for project and terminal deletion instead of filtering items only in the frontend. `DeleteProject(projectID)` should coordinate `ProjectManager` and `ShellSessionManager`, then return a complete `ProjectState`. `DeleteTerminal(terminalID)` should delegate terminal removal to `ShellSessionManager` and return `withShellState`.

Alternative considered: frontend-only removal. That would leave persisted projects and running PTYs alive, so refresh or shell events could resurrect deleted items.

### Project deletion removes records, not directories

`ProjectManager.DeleteProject` should remove the project from `projects.json` and should not touch `Project.Path`. If the deleted project was active, the manager should choose the remaining project with the most recent `LastSelectedAt`; if none remain, active project becomes empty.

Alternative considered: mark projects hidden. That adds migration and recovery behavior the app does not currently need.

### Shell deletion closes processes before removing descriptors

`ShellSessionManager.DeleteTerminal` should remove the terminal descriptor, clear active mappings that point at it, and close the running process if one exists. Closing should reuse existing process cleanup paths so integration temp files are released. Deleting an exited terminal should only remove runtime state.

When a deleted terminal was active for its project, the manager should choose another terminal in the same project by most recent `LastSelectedAt`. If none remain, the project has no active terminal and the frontend shows the existing "Select a terminal" empty state.

Alternative considered: automatically create a replacement terminal when deleting the last active terminal. That makes deletion feel reversible but surprising, because the item immediately returns as a new terminal.

### Project deletion removes owned terminals in one backend step

`DeleteProject` should call a shell-manager project cleanup method after the project record is removed. This closes every terminal for that project and removes their descriptors so the returned `ProjectState` cannot include orphan terminals.

Alternative considered: have the frontend call `DeleteTerminal` for each child before deleting the project. That creates partial-failure states and leaks backend lifecycle coordination into UI code.

### Sidebar uses compact icon controls

Add delete icon buttons to project and terminal rows. Project deletion asks for confirmation in `App.vue` before calling the API. Terminal deletion does not ask for confirmation. Delete buttons stop event propagation so they do not also select/toggle rows.

Alternative considered: context menus. They keep rows visually cleaner but make the feature harder to discover and require more UI state than this change needs.

## Risks / Trade-offs

- Closing a PTY can race with the existing `waitForExit` goroutine -> deletion should tolerate late exit/status events for terminals no longer present.
- Deleting the active project can leave no active terminal -> reuse the existing "Select a terminal" or "Select a project" states instead of auto-creating shells.
- Project delete confirmation uses browser/Wails confirm behavior -> keep confirmation text specific and test frontend calls by mocking `window.confirm`.
- More row controls can crowd the sidebar -> use stable-width icon buttons and preserve text truncation.

## Migration Plan

No data migration is required. Existing `projects.json` files remain compatible. Rollback removes the delete APIs and UI controls; project records and runtime terminals created before rollback remain in their normal formats.

## Open Questions

None.
