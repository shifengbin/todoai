## 1. Frontend Behavior

- [x] 1.1 Add computed state or helpers for active TODO ids and whether bulk controls should be enabled.
- [x] 1.2 Implement bulk collapse by adding all active TODO ids to the existing collapsed TODO id set.
- [x] 1.3 Implement bulk expand by removing all active TODO ids from the existing collapsed TODO id set.
- [x] 1.4 Preserve existing automatic expansion when the active TODO, TODO project, or terminal context changes.

## 2. User Interface

- [x] 2.1 Add compact expand-all and collapse-all controls to the active TODO list toolbar.
- [x] 2.2 Add accessible labels, titles, disabled states, and stable test ids for the new controls.
- [x] 2.3 Adjust sidebar styles so the new controls fit without crowding the TODO view tabs or list content.

## 3. Tests

- [x] 3.1 Add a ProjectSidebar unit test for collapsing all active TODO branches.
- [x] 3.2 Add a ProjectSidebar unit test for expanding all active TODO branches.
- [x] 3.3 Add or extend a unit test proving active context changes still expand only the relevant TODO after bulk collapse.
- [x] 3.4 Run the frontend automated test suite for the affected component.

## 4. Review And Verification

- [x] 4.1 Run project linting or formatting checks if available for the frontend.
- [x] 4.2 Run an automated code review check for the changed files.
- [x] 4.3 Manually verify the TODO active view with multiple TODOs, empty active TODO list, and archived view.
