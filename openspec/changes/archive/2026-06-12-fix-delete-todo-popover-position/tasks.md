## 1. Sidebar Popover Structure

- [x] 1.1 Move or wrap the TODO delete confirmation popover in `ProjectSidebar.vue` so `.todo-action-popover` is anchored by the current TODO action context.
- [x] 1.2 Preserve the existing right-click/three-dot menu delete flow: selecting Delete TODO closes the menu and opens the confirmation popover without emitting delete.
- [x] 1.3 Keep `todoActionConfirm` as the single source of truth for active TODO action confirmation state.
- [x] 1.4 Ensure opening terminal launch menus, TODO project removal popovers, project delete popovers, bulk delete popovers, or outside clicks closes the active TODO action popover.

## 2. Popover Styling

- [x] 2.1 Update `frontend/src/style.css` only as needed so the TODO delete confirmation popover uses a stable relative positioning container.
- [x] 2.2 Preserve the existing visual style and placement of the complete TODO confirmation popover.
- [x] 2.3 Verify project delete, bulk project delete, and TODO project removal popovers are not visually or structurally changed by the fix.

## 3. Frontend Tests

- [x] 3.1 Add or update `ProjectSidebar` tests so the TODO delete menu action opens a confirmation popover and does not emit `delete-todo` before confirmation.
- [x] 3.2 Add a regression test that the delete TODO confirmation popover is rendered inside a stable TODO action positioning container instead of as an unanchored TODO node child.
- [x] 3.3 Keep or update tests for delete confirmation cancel, confirm, outside-click close, and mutual closing with other sidebar popovers.
- [x] 3.4 Run the affected frontend test suite with `npm test -- --runInBand` or the project-supported equivalent from `frontend/`.

## 4. Verification And Packaging

- [x] 4.1 Run `npm run build` from `frontend/`.
- [x] 4.2 Perform an automated code review pass for the changed frontend files and tests, checking behavior, regressions, and test coverage.
- [x] 4.3 Run `openspec status --change fix-delete-todo-popover-position` and confirm all required artifacts and tasks are ready for implementation.
- [x] 4.4 Run `wails build -tags webkit2_41` to generate the executable.
