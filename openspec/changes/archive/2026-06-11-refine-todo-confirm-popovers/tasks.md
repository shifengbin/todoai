## 1. Todo Action Confirmation Popovers

- [x] 1.1 Add `ProjectSidebar.vue` state and helpers for TODO action confirmations keyed by TODO id and action type.
- [x] 1.2 Change the TODO complete and delete buttons to open inline confirmation popovers instead of emitting immediately.
- [x] 1.3 Add confirm and cancel controls for complete/delete popovers with stable test ids and accessible labels.
- [x] 1.4 Wire TODO action popovers into the existing floating-menu close flow so terminal launch menus, project removal popovers, outside clicks, cancel, and successful confirm close the active popover.

## 2. App Todo Action Flow

- [x] 2.1 Remove `window.confirm` usage from `App.vue` complete TODO and delete TODO handlers.
- [x] 2.2 Keep existing TODO completion/deletion backend calls, state application, terminal activation, and error handling unchanged after sidebar confirmation.

## 3. Todo Project Row Styling

- [x] 3.1 Move TODO project row hover and active background styling to the full `.todo-project-header-row` area.
- [x] 3.2 Keep the inner TODO project `.project-row` background transparent so the full-row background covers the project info, create terminal button, and delete button columns.
- [x] 3.3 Preserve readable create terminal and delete button hover/focus states on top of the full-row background.
- [x] 3.4 Ensure project library rows keep their existing row background behavior and are not affected by TODO project row styling.

## 4. Frontend Tests

- [x] 4.1 Update `ProjectSidebar` tests so complete/delete buttons open confirmation popovers and do not emit before confirmation.
- [x] 4.2 Add `ProjectSidebar` tests for confirming and canceling complete/delete popovers.
- [x] 4.3 Add `ProjectSidebar` tests for outside-click closing and mutual closing with terminal launch or TODO project removal popovers.
- [x] 4.4 Add style assertions that TODO project active/hover backgrounds are defined on the full project header and still use a three-column row.
- [x] 4.5 Update `App` tests so TODO completion/deletion no longer depend on `window.confirm` and only run after sidebar popover confirmation.

## 5. Verification And Review

- [x] 5.1 Run frontend unit tests.
- [x] 5.2 Run frontend build.
- [x] 5.3 Run automated review for the completed change and address findings.
- [x] 5.4 Run `openspec status --change refine-todo-confirm-popovers` and confirm all required artifacts and tasks are ready for implementation.
