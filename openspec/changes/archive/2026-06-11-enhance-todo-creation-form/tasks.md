## 1. Backend State And API

- [x] 1.1 Add TODO description and priority fields to the Go `Todo` model and generated frontend model shape.
- [x] 1.2 Add priority constants and normalization so missing or empty priority defaults to `medium`.
- [x] 1.3 Replace the title-only TODO creation flow with a structured create request containing title, description, priority, and optional project ID.
- [x] 1.4 Make TODO creation with an optional project association save atomically and select the created TODO project context when a project is chosen.
- [x] 1.5 Preserve existing `AddProjectToTodo` behavior for existing TODOs while keeping duplicate association handling unchanged.
- [x] 1.6 Regenerate Wails bindings after Go API/model changes.

## 2. Frontend Interaction And Rendering

- [x] 2.1 Replace the create TODO `window.prompt` flow with an application overlay form for name, optional description, priority, and optional project.
- [x] 2.2 Implement searchable project selection for create TODO, filtering by project name and path and allowing no project selection.
- [x] 2.3 Replace the add-project-to-TODO prompt with the same searchable project selection pattern, excluding projects already linked to that TODO.
- [x] 2.4 Display TODO descriptions when present without expanding rows unnecessarily when descriptions are empty.
- [x] 2.5 Display `高`、`中`、`低` priority labels and apply red, orange, and green same-family TODO row backgrounds.
- [x] 2.6 Add light and dark theme CSS tokens for priority backgrounds, borders, and labels with readable active states.

## 3. Tests And Verification

- [x] 3.1 Add Go tests for creating TODOs with description, priority, no project, optional project, invalid title, and legacy missing-priority data.
- [x] 3.2 Update app-level tests for the structured create TODO API and selected TODO project context when a project is chosen.
- [x] 3.3 Add or update frontend tests for opening the create form, validating required name, filtering projects, optional project submission, and priority rendering.
- [x] 3.4 Add or update frontend tests for adding a project to an existing TODO through searchable selection.
- [x] 3.5 Run `go test ./...`.
- [x] 3.6 Run frontend unit tests.
- [x] 3.7 Run frontend build.
- [x] 3.8 Validate the OpenSpec change status before implementation handoff.

## 4. Multi-Project Selection Follow-Up

- [x] 4.1 Update backend create TODO request from single `projectId` to ordered `projectIds`.
- [x] 4.2 Create all selected TODO-project associations atomically and select the first chosen project as the active context.
- [x] 4.3 Update frontend create TODO project picker to support searching and selecting multiple projects.
- [x] 4.4 Update existing TODO add-project picker to support selecting and submitting multiple projects.
- [x] 4.5 Update generated Wails bindings for the multi-project request shape.
- [x] 4.6 Add backend and frontend tests for multi-project create/add behavior.
- [x] 4.7 Run Go tests, frontend tests, frontend build, and OpenSpec apply status.

## 5. Removable Project Tags Follow-Up

- [x] 5.1 Show selected projects as individual removable tags in the create TODO form.
- [x] 5.2 Show selected projects as individual removable tags in the add-project-to-TODO picker.
- [x] 5.3 Add frontend tests for removing selected project tags before submission.
- [x] 5.4 Run frontend tests, frontend build, and OpenSpec apply status.
