## 1. Backend Project Deletion API

- [x] 1.1 Add `ProjectManager.DeleteProjects` tests for successful multi-project removal, TODO association cleanup, active project fallback, and empty active project fallback.
- [x] 1.2 Add `ProjectManager.DeleteProjects` tests for duplicate IDs, empty input, and missing project all-or-nothing failure.
- [x] 1.3 Implement `ProjectManager.DeleteProjects` with ID normalization, validation before mutation, single save, and existing delete semantics.
- [x] 1.4 Add `App.DeleteProjects` tests verifying selected project terminals are closed and unrelated project terminals remain.
- [x] 1.5 Implement `App.DeleteProjects` and call shell project cleanup for each deleted project after successful project state persistence.

## 2. Wails Bindings

- [x] 2.1 Regenerate or update Wails JS/TS bindings for `DeleteProjects`.
- [x] 2.2 Verify frontend imports and type declarations expose `DeleteProjects(projectIDs: string[])`.

## 3. Sidebar Project Library UX

- [x] 3.1 Add `ProjectSidebar` tests that single project delete opens a row popover, cancel keeps the project, and confirm emits `delete-project`.
- [x] 3.2 Implement project row delete confirmation popover in `ProjectSidebar` and remove direct single-delete emit from the delete icon.
- [x] 3.3 Add `ProjectSidebar` tests for project checkbox selection, no row selection on checkbox click, disabled empty bulk delete, selected count, bulk cancel, and bulk confirm emit.
- [x] 3.4 Implement project selection state, selection cleanup when projects change, and top toolbar bulk delete confirmation popover.
- [x] 3.5 Add or adjust CSS for fixed-width checkbox controls, delete popovers, and compact top toolbar actions without breaking project name/path truncation.

## 4. App Integration

- [x] 4.1 Update `App.vue` single project deletion to trust sidebar confirmation and remove `window.confirm` from `deleteProject`.
- [x] 4.2 Add `deleteProjects` handler in `App.vue` that calls `DeleteProjects`, applies returned state, and activates the active terminal.
- [x] 4.3 Wire `ProjectSidebar` `delete-projects` event to the new App handler.
- [x] 4.4 Update `App` frontend tests so single deletion no longer calls `window.confirm` and bulk deletion calls `DeleteProjects` only after popover confirmation.

## 5. Verification

- [x] 5.1 Run Go tests covering project manager, app API, and shell cleanup behavior.
- [x] 5.2 Run client automated tests for `ProjectSidebar` and `App`.
- [x] 5.3 Run frontend lint/type checks or the closest existing validation command.
- [x] 5.4 Run automated review/static review for the completed change and address findings.
- [x] 5.5 Run `openspec validate add-project-bulk-delete --strict` and resolve any artifact issues.
