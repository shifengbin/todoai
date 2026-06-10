## 1. Component State

- [x] 1.1 Add a `projectHasActiveTerminal(projectId)` helper in `ProjectSidebar.vue` that checks whether the project owns the current `activeTerminalId`.
- [x] 1.2 Add a stable state class, such as `has-active-terminal`, to the matching `.project-node` without changing existing active project, collapse, unavailable, or terminal row classes.

## 2. Branch Guide Styling

- [x] 2.1 Update `frontend/src/style.css` so `.project-node.has-active-terminal .terminal-list::before` uses the same color as `.terminal-row.active::before`.
- [x] 2.2 Confirm projects without an active child terminal keep the default neutral terminal-list branch guide color.

## 3. Tests

- [x] 3.1 Add a `ProjectSidebar` component test that renders two projects with terminals and asserts only the project owning `activeTerminalId` receives the active-branch state class.
- [x] 3.2 Keep existing project tree interaction tests passing, including select, create, collapse, expand, and delete actions.

## 4. Verification

- [x] 4.1 Run the frontend test suite for the sidebar change.
- [x] 4.2 Run OpenSpec validation for `highlight-active-terminal-branch`.
