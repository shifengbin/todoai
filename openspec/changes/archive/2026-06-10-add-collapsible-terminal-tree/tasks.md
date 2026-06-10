## 1. Sidebar Collapse Interaction

- [x] 1.1 Add local collapsed-project state to `ProjectSidebar` keyed by project ID.
- [x] 1.2 Add a per-project expand/collapse icon button that toggles terminal child visibility without selecting the project.
- [x] 1.3 Auto-expand the active project's branch when `activeProjectId` changes.
- [x] 1.4 Auto-expand a terminal's owning project when `activeTerminalId` changes or a new active terminal appears.
- [x] 1.5 Preserve existing project selection, terminal selection, new-project, and add-terminal emits.

## 2. Tree Visual Treatment

- [x] 2.1 Update `ProjectSidebar` markup with stable classes and accessibility labels for expanded and collapsed states.
- [x] 2.2 Restyle project rows, toggle buttons, and add-terminal buttons so controls fit without layout shift.
- [x] 2.3 Restyle terminal rows with nested indentation, branch guides, and connector lines that clearly show parent/child ownership.
- [x] 2.4 Verify long project paths and long command labels remain truncated without overlapping controls.
- [x] 2.5 Keep unavailable project styling and active terminal/project styling visually distinct.

## 3. Tests And Verification

- [x] 3.1 Add `ProjectSidebar` tests for collapsing and expanding a project branch.
- [x] 3.2 Add `ProjectSidebar` tests that collapsing one project does not change another project's branch state.
- [x] 3.3 Add `ProjectSidebar` tests that active project or active terminal changes expand the owning project.
- [x] 3.4 Update existing app/sidebar tests for any changed selectors, classes, or accessible labels.
- [x] 3.5 Run `npm test` in `frontend/`.
