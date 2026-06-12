## 1. Frontend Tooltip Behavior

- [x] 1.1 Add local tooltip hover state and timer cleanup to `ProjectSidebar.vue`.
- [x] 1.2 Show the full description tooltip only for TODOs with non-empty descriptions after the configured hover delay.
- [x] 1.3 Hide the tooltip immediately on mouse leave, target switch, sidebar popover/menu opening, and component unmount.
- [x] 1.4 Keep the existing inline one-line TODO description summary visible in the row.

## 2. Tooltip Styling

- [x] 2.1 Add CSS for the TODO description tooltip with readable foreground/background, multi-line wrapping, width limits, and proper stacking.
- [x] 2.2 Verify tooltip styling remains readable in existing light and dark themes.
- [x] 2.3 Ensure the tooltip does not resize TODO rows or disrupt TODO expand/collapse, action buttons, or context menus.

## 3. Frontend Tests

- [x] 3.1 Update existing TODO description rendering tests to expect the row summary and hidden default tooltip state.
- [x] 3.2 Add ProjectSidebar tests using controlled timers for delayed tooltip display.
- [x] 3.3 Add ProjectSidebar tests for mouse leave hiding and empty-description TODOs not showing a tooltip.
- [x] 3.4 Run frontend unit tests.

## 4. Verification And Packaging

- [x] 4.1 Run automated review for the completed change and address findings.
- [x] 4.2 Run `openspec status --change add-todo-description-tooltip`.
- [x] 4.3 Run frontend build.
- [x] 4.4 Run `wails build -tags webkit2_41` to generate the executable.
