## 1. Settings Data Model

- [x] 1.1 Add a Go terminal launch profile setting model with `name` and `command` fields.
- [x] 1.2 Extend terminal settings state and persisted settings to include `launchProfiles`.
- [x] 1.3 Implement default launch profiles for missing settings fields: `codex` -> `codex`, `claude` -> `claude`.
- [x] 1.4 Preserve explicit empty launch profile lists without restoring defaults.
- [x] 1.5 Add normalization and validation for launch profile names, commands, duplicate names, and the reserved `Terminal` name.
- [x] 1.6 Ensure shell setting saves preserve launch profiles, and launch profile saves preserve the selected shell setting.

## 2. Backend API And Bindings

- [x] 2.1 Add an App method for saving terminal launch profiles from settings.
- [x] 2.2 Return launch profiles from `LoadTerminalSettings`.
- [x] 2.3 Regenerate Wails frontend bindings and models for the updated settings state/API.
- [x] 2.4 Add Go tests for default profiles, saved profiles, explicit empty profiles, invalid profiles, and preserving settings fields across saves.

## 3. Settings UI

- [x] 3.1 Load terminal settings early enough for the project launch menu to use saved profiles.
- [x] 3.2 Add settings dialog controls for listing, adding, editing, reordering, and removing custom launch profiles.
- [x] 3.3 Keep the built-in `Terminal` launch option visible as non-configurable context while excluding it from the editable profile list.
- [x] 3.4 Surface launch profile validation errors without closing the settings dialog.
- [x] 3.5 Add frontend tests for rendering default profiles, saving edits, removing profiles, reordering profiles, and handling validation errors.

## 4. Project Launch Menu

- [x] 4.1 Replace the project add-terminal click action with a launch menu anchored to the project row.
- [x] 4.2 Render `Terminal` first and append configured launch profiles in saved order.
- [x] 4.3 Create a normal terminal when the user selects `Terminal`.
- [x] 4.4 Create and activate a new terminal, then send `command + "\\n"` to that terminal when the user selects a custom launch profile.
- [x] 4.5 Close the launch menu after selection, outside click, project deletion, or unavailable project state changes.
- [x] 4.6 Add frontend tests for launch menu rendering, plain terminal creation, custom profile command submission, and unavailable project behavior.

## 5. Verification

- [x] 5.1 Run Go tests covering settings, app, and shell behavior.
- [x] 5.2 Run frontend unit tests.
- [x] 5.3 Run the project build or Wails build used by this repository.
- [x] 5.4 Validate the OpenSpec change.
