## 1. Backend Deletion Semantics

- [x] 1.1 Add failing `ProjectManager` tests for deleting an opened project, preserving disk files, active-project fallback, and not-found errors.
- [x] 1.2 Implement `ProjectManager.DeleteProject` and helper selection logic.
- [x] 1.3 Add failing `ShellSessionManager` tests for deleting running, exited, active, last, and missing terminals.
- [x] 1.4 Implement `ShellSessionManager.DeleteTerminal` with PTY close, descriptor removal, and active-terminal fallback.
- [x] 1.5 Add failing `ShellSessionManager` tests for removing all terminals owned by a deleted project while preserving other projects.
- [x] 1.6 Implement project-scoped terminal cleanup in `ShellSessionManager`.

## 2. App API And Bindings

- [x] 2.1 Add failing `App` API tests for `DeleteProject` and `DeleteTerminal` returning consistent `ProjectState`.
- [x] 2.2 Implement `App.DeleteProject` and `App.DeleteTerminal`.
- [x] 2.3 Regenerate or update Wails frontend bindings for the new delete APIs.

## 3. Frontend Behavior

- [x] 3.1 Add failing `ProjectSidebar` tests for delete-project and delete-terminal events without triggering row selection.
- [x] 3.2 Add delete controls to `ProjectSidebar` rows and emit delete events.
- [x] 3.3 Add failing `App.vue` tests for project delete confirmation/cancel and terminal delete API calls.
- [x] 3.4 Implement App delete handlers, confirmation behavior, state application, and stale menu cleanup.
- [x] 3.5 Style delete controls with stable dimensions and preserve row truncation and active-state visuals.

## 4. Verification

- [x] 4.1 Run Go tests covering project, shell, and app deletion behavior.
- [x] 4.2 Run frontend unit tests covering App and ProjectSidebar.
- [x] 4.3 Run the project build to verify generated bindings and frontend compile.
