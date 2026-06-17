## 1. State Model

- [x] 1.1 Update the TODO UI state data model so `todoView` remains per TODO project and `sidebarWidth` is stored as workspace-level layout state.
- [x] 1.2 Add backward-compatible loading for existing `todo-project-ui-state.json` files that still contain per TODO project `sidebarWidth`.
- [x] 1.3 Update Wails bindings and generated frontend models for the revised UI state shape.

## 2. Frontend Behavior

- [x] 2.1 Split frontend persistence helpers so TODO view saves do not overwrite workspace sidebar width and sidebar width saves do not overwrite TODO project view state.
- [x] 2.2 Change `applyState` so active TODO project changes from business refreshes do not automatically restore TODO view or sidebar width.
- [x] 2.3 Keep explicit restoration for workspace load/reload and user-selected TODO projects, restoring TODO view only for TODO project selection and restoring sidebar width only from workspace layout state.
- [x] 2.4 Ensure adding a terminal preserves the current TODO view and left TODO sidebar width.
- [x] 2.5 Ensure selecting a terminal under a TODO item preserves the current TODO view and left TODO sidebar width.

## 3. Tests

- [x] 3.1 Add or update Go tests for workspace-level sidebar width load/save and legacy per TODO project sidebar width compatibility.
- [x] 3.2 Add or update Vue tests proving TODO view restoration is still per TODO project.
- [x] 3.3 Add Vue tests proving left TODO sidebar width is workspace-level and does not change when selecting different TODO projects.
- [x] 3.4 Add Vue tests proving adding a terminal does not change the current TODO view or left TODO sidebar width.
- [x] 3.5 Add Vue tests proving selecting a terminal does not change the current TODO view or left TODO sidebar width.

## 4. Verification

- [x] 4.1 Run Go tests covering the TODO UI state store and app UI state APIs.
- [x] 4.2 Run frontend automated tests covering TODO workspace UI state behavior.
- [x] 4.3 Run `openspec validate preserve-todo-ui-state-on-terminal-actions --strict`.
- [x] 4.4 Perform automatic review of the implementation for state ownership, migration compatibility, and regression risk.
- [x] 4.5 Run `wails build -tags webkit2_41` to generate the executable.
