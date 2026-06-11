## 1. Backend Todo State And API

- [x] 1.1 Add an `UpdateTodoRequest` model with TODO id, title, description, priority, and ordered project IDs.
- [x] 1.2 Implement TODO update validation for required title, normalized priority, existing active TODO, existing project IDs, and duplicate project IDs.
- [x] 1.3 Implement backend TODO-project diffing so update saves new associations, removes missing associations, and preserves unchanged associations.
- [x] 1.4 Update active TODO, active project, active TODO-project, and active terminal state when an edited TODO loses the active TODO-project association.
- [x] 1.5 Add a backend API for removing a single TODO-project association directly from the TODO tree.
- [x] 1.6 Regenerate Wails bindings after API and model changes.

## 2. Terminal Cleanup Semantics

- [x] 2.1 Add shell manager cleanup by `todoProjectID` that closes running PTY processes and removes runtime terminal state for that TODO-project only.
- [x] 2.2 Wire TODO update removals to return or apply the removed `todoProjectID` set and clean their terminals after state save.
- [x] 2.3 Wire direct TODO-project removal to the same terminal cleanup path.
- [x] 2.4 Preserve existing TODO-level and project-level terminal cleanup behavior for complete/delete TODO and delete project flows.

## 3. Frontend Todo Details And Project Removal

- [x] 3.1 Add an eye icon action to each active TODO item for opening the TODO detail editor.
- [x] 3.2 Implement TODO detail editor state in `App.vue` for title, description, priority, searchable multi-project selection, removable tags, save, cancel, and loading/error states.
- [x] 3.3 Show save-time confirmation when TODO detail edits remove one or more projects that own terminals, and keep the editor open when the user cancels.
- [x] 3.4 Submit detail edits through the new structured update API and apply returned state.
- [x] 3.5 Add a delete action to TODO project rows and show an inline confirmation popover next to the delete button.
- [x] 3.6 Close the project removal popover on cancel, outside click, target switch, or successful removal.
- [x] 3.7 Call the direct remove TODO-project API from the popover confirm action and apply returned state.

## 4. Sidebar Layout And Visual Refinements

- [x] 4.1 Move TODO priority background styling to the full TODO item header so it covers the expand control, content, and action area.
- [x] 4.2 Remove priority text badges from TODO items while preserving priority color styling.
- [x] 4.3 Ensure TODO title, description, project counts, and action buttons remain readable in light and dark themes.
- [x] 4.4 Make the TODO workspace list scroll inside the sidebar when content exceeds available height.
- [x] 4.5 Add a draggable divider between sidebar and terminal workspace with min/max sidebar width constraints.
- [x] 4.6 Refit the active terminal during and after sidebar resize so PTY dimensions stay in sync.

## 5. Backend Tests

- [x] 5.1 Add Go tests for updating TODO title, description, priority, and project associations.
- [x] 5.2 Add Go tests for update validation failures and unchanged state on invalid requests.
- [x] 5.3 Add Go tests for removing one TODO-project association while preserving the same project under another TODO.
- [x] 5.4 Add Go tests for `todoProjectID` terminal cleanup, including running terminal close and active terminal correction.
- [x] 5.5 Run `go test ./...`.

## 6. Frontend Tests And Build

- [x] 6.1 Add or update ProjectSidebar tests for full-width priority background classes and absence of item priority text badges.
- [x] 6.2 Add ProjectSidebar tests for the eye icon action and TODO project deletion popover confirm/cancel behavior.
- [x] 6.3 Add App tests for opening the detail editor, editing TODO metadata, adding/removing projects, save-time terminal-close confirmation, and canceled saves.
- [x] 6.4 Add App tests for sidebar drag width changes and active terminal fit behavior.
- [x] 6.5 Run frontend unit tests.
- [x] 6.6 Run frontend build.

## 7. Review And OpenSpec Verification

- [x] 7.1 Run automated review for the completed change and address findings.
- [x] 7.2 Run `openspec status --change refine-todo-sidebar-interactions`.
- [x] 7.3 Confirm all tasks are complete and the change is ready for implementation handoff or archive.
