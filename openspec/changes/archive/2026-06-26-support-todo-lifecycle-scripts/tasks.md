## 1. Backend Data Model And Settings

- [x] 1.1 Add lifecycle script template and snapshot structs to settings/project models.
- [x] 1.2 Persist global lifecycle script templates in terminal settings state with empty defaults for existing settings files.
- [x] 1.3 Add validation for script template name, at-least-one script body, and single default selection.
- [x] 1.4 Add app APIs to load and save lifecycle script templates.
- [x] 1.5 Add tests for settings persistence, validation failures, default loading, and preservation of existing settings fields.

## 2. Todo Creation And Snapshot Storage

- [x] 2.1 Extend create TODO request and TODO model with an optional lifecycle script snapshot.
- [x] 2.2 Normalize and persist selected script pair snapshots when creating TODOs.
- [x] 2.3 Ensure existing TODOs without script snapshots load as no-script TODOs.
- [x] 2.4 Add tests proving global template edits do not change existing TODO snapshots.

## 3. Lifecycle Script Execution Backend

- [x] 3.1 Implement a lifecycle script executor that runs scripts asynchronously in the TODO workspace directory.
- [x] 3.2 Use background command startup so Windows script execution does not show a console window.
- [x] 3.3 Resolve current-platform shell execution for Unix and Windows without translating script content.
- [x] 3.4 Capture exit code and truncated output tail for failed script runs.
- [x] 3.5 Prevent duplicate runs for the same TODO and phase while a script is queued or running.
- [x] 3.6 Add unit tests for command construction, working directory, output truncation, duplicate prevention, and Windows hidden-window behavior.

## 4. Todo Lifecycle Integration

- [x] 4.1 Trigger initialization script execution after TODO successfully enters `in-progress`.
- [x] 4.2 Keep initialization non-gating: TODO remains `in-progress` even when initialization is running or fails.
- [x] 4.3 Route complete requests with completion scripts through asynchronous script execution.
- [x] 4.4 Complete and archive the TODO only after completion script success.
- [x] 4.5 Keep TODO `in-progress` and retryable when completion script fails.
- [x] 4.6 Clear lifecycle script states when TODOs are deleted or successfully completed.
- [x] 4.7 Add app/project tests for start behavior, completion gating, failure states, retry behavior, and no-script fallback.

## 5. Script Status State And Events

- [x] 5.1 Define lifecycle script status payloads keyed by TODO ID and phase.
- [x] 5.2 Expose script statuses through project state or a workspace-scoped status store.
- [x] 5.3 Emit Wails events for queued, running, failed, and cleared script states.
- [x] 5.4 Preserve failed states until retry success, TODO deletion, or successful completion.
- [x] 5.5 Add tests for status persistence, event emission, success clearing, and failed-state retention.

## 6. Frontend Management And Selection UI

- [x] 6.1 Add a global management menu item for script management.
- [x] 6.2 Build script management dialog for adding, editing, reordering, deleting, and marking one default script pair.
- [x] 6.3 Add create TODO script pair dropdown with filtering by name and description.
- [x] 6.4 Show selected script descriptions in the create TODO form.
- [x] 6.5 Send selected script snapshot in `CreateTodo` requests and support no selection.
- [x] 6.6 Add frontend tests for script management validation, default selection, filtering, description display, and create TODO payloads.

## 7. Frontend Runtime Status And Retry UI

- [x] 7.1 Subscribe to lifecycle script status events and merge them into local app state.
- [x] 7.2 Display running states for initialization and completion scripts on affected TODOs.
- [x] 7.3 Hide successful script states after backend clear events.
- [x] 7.4 Display failed states with error summary and retry actions.
- [x] 7.5 Disable duplicate retry actions while the same phase is running.
- [x] 7.6 Add frontend automated tests for running display, success hiding, failed display, retry triggers, and duplicate-action disabling.

## 8. Generated Bindings And Regression Verification

- [x] 8.1 Regenerate Wails frontend bindings after backend API/model changes.
- [x] 8.2 Run Go unit tests for backend model, settings, executor, app lifecycle, and Windows-specific command behavior.
- [x] 8.3 Run frontend automated tests for Vue UI and state handling.
- [x] 8.4 Run OpenSpec validation/status checks for the completed change.
- [x] 8.5 Run automated code review checks for changed backend, frontend, and spec artifacts.
- [x] 8.6 Run `wails build -tags webkit2_41` to generate the executable.
