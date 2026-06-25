## 1. Backend Settings Model and API

- [x] 1.1 Add Go structs for TODO initialization file templates with `name`, `description`, `fileName`, `content`, and `defaultSelected` fields.
- [x] 1.2 Extend settings persistence to load and save initialization file templates while preserving existing shell, launch profile, and theme settings.
- [x] 1.3 Implement backend validation for template name, unique root-level file names, path traversal, absolute paths, and directory separators.
- [x] 1.4 Add Wails-facing methods to load and save initialization file templates.
- [x] 1.5 Regenerate Wails frontend bindings for the new or changed backend methods and models.

## 2. Todo Snapshot and Workspace File Creation

- [x] 2.1 Extend TODO data and create request models to store selected initialization file snapshots.
- [x] 2.2 Update TODO creation to persist selected template snapshots with file name and content exactly as submitted.
- [x] 2.3 Implement task workspace initialization file writing that creates missing selected files in the TODO task folder.
- [x] 2.4 Integrate initialization file writing into all task workspace preparation paths without changing README regeneration behavior.
- [x] 2.5 Ensure existing task workspace files are not overwritten when preparation runs repeatedly.

## 3. Frontend Global File Management UI

- [x] 3.1 Add a menu bar “Global management” entry with a “File management” item.
- [x] 3.2 Move initialization file template loading and saving into an independent file management dialog.
- [x] 3.3 Add file management UI controls to create, edit, reorder or remove initialization file templates.
- [x] 3.4 Add fields for template name, description, file name, content, and default selection.
- [x] 3.5 Surface backend validation errors in the file management dialog without losing unsaved user input.
- [x] 3.6 Keep TODO initialization file management out of the terminal Settings dialog.

## 4. Frontend Todo Creation UI

- [x] 4.1 Load initialization file templates before opening the create TODO dialog.
- [x] 4.2 Display each template's name, description, and file name in the create TODO dialog.
- [x] 4.3 Preselect templates whose `defaultSelected` value is true.
- [x] 4.4 Allow users to manually select or unselect templates before submitting the TODO.
- [x] 4.5 Include selected initialization file snapshots in the `CreateTodo` request.

## 5. Tests

- [x] 5.1 Add Go unit tests for settings persistence, migration defaults, and template validation.
- [x] 5.2 Add Go unit tests for TODO creation snapshot persistence and snapshot stability after global template changes.
- [x] 5.3 Add Go unit tests for task workspace initialization file writing, including non-overwrite behavior.
- [x] 5.4 Add frontend automated tests for global file management editing, Settings exclusion, and validation error display.
- [x] 5.5 Add frontend automated tests for create TODO default selection, manual selection, and request payload snapshots.

## 6. Review and Verification

- [x] 6.1 Run Go tests for backend settings, project manager, app, and todo workspace behavior.
- [x] 6.2 Run frontend automated tests for the settings and create TODO flows.
- [x] 6.3 Run OpenSpec validation for `add-todo-initialization-files`.
- [x] 6.4 Perform an automated code review pass and address findings before completion.
- [x] 6.5 Run `wails build -tags webkit2_41` to generate the executable file.
