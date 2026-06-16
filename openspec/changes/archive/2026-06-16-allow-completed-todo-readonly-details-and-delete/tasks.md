## 1. Backend Behavior

- [x] 1.1 Extend single TODO deletion so `completed` TODO records can be deleted while existing `not-started` and `in-progress` deletion behavior remains unchanged.
- [x] 1.2 Add completed-only bulk TODO deletion behavior that rejects empty input, unknown TODO IDs, and any non-`completed` TODO.
- [x] 1.3 Ensure completed TODO deletion does not recreate terminals, start shell processes, or restore TODO project associations.
- [x] 1.4 Add or update Go tests for completed single deletion, completed bulk deletion, and rejection of non-completed bulk deletion.

## 2. Wails API

- [x] 2.1 Expose the completed-only bulk delete operation through the app API.
- [x] 2.2 Regenerate Wails frontend bindings after API changes.

## 3. Frontend Details And Deletion

- [x] 3.1 Reuse the existing TODO detail dialog for completed TODOs with a read-only mode.
- [x] 3.2 In completed read-only mode, show title, description, priority, and completed project snapshots with project name and path.
- [x] 3.3 In completed read-only mode, hide the save button and disable editing of metadata and projects.
- [x] 3.4 Add completed TODO menu support for opening read-only details and confirming single deletion.
- [x] 3.5 Add completed-view-only selection state and bulk delete confirmation for selected completed TODOs.
- [x] 3.6 Ensure TODO bulk delete controls are not shown in `未执行` or `执行中` views.

## 4. Frontend Tests

- [x] 4.1 Add component tests for opening completed TODO details in read-only mode.
- [x] 4.2 Add component tests verifying the save button is hidden and completed TODO metadata/projects cannot be edited.
- [x] 4.3 Add component tests for completed single deletion from the completed view.
- [x] 4.4 Add component tests for completed bulk deletion and cancellation.
- [x] 4.5 Add component tests verifying open TODO views do not expose bulk delete.

## 5. Verification

- [x] 5.1 Run Go test suite covering TODO workspace behavior.
- [x] 5.2 Run frontend automated tests.
- [x] 5.3 Run automatic code review for the completed change and address findings.
- [x] 5.4 Run `wails build -tags webkit2_41` to generate the executable.
