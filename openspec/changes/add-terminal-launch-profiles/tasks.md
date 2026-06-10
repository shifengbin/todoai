## 1. Settings Data Model

- [ ] 1.1 Add a Go terminal launch profile setting model with `name` and `command` fields.
- [ ] 1.2 Extend terminal settings state and persisted settings to include `launchProfiles`.
- [ ] 1.3 Implement default launch profiles for missing settings fields: `codex` -> `codex`, `claude` -> `claude`.
- [ ] 1.4 Preserve explicit empty launch profile lists without restoring defaults.
- [ ] 1.5 Add normalization and validation for launch profile names, commands, duplicate names, and the reserved `Terminal` name.
- [ ] 1.6 Ensure shell setting saves preserve launch profiles, and launch profile saves preserve the selected shell setting.

## 2. Backend API And Bindings

- [ ] 2.1 Add an App method for saving terminal launch profiles from settings.
- [ ] 2.2 Return launch profiles from `LoadTerminalSettings`.
- [ ] 2.3 Regenerate Wails frontend bindings and models for the updated settings state/API.
- [ ] 2.4 Add Go tests for default profiles, saved profiles, explicit empty profiles, invalid profiles, and preserving settings fields across saves.

## 3. Settings UI

- [ ] 3.1 Load terminal settings early enough for the project launch menu to use saved profiles.
- [ ] 3.2 Add settings dialog controls for listing, adding, editing, reordering, and removing custom launch profiles.
- [ ] 3.3 Keep the built-in `Terminal` launch option visible as non-configurable context while excluding it from the editable profile list.
- [ ] 3.4 Surface launch profile validation errors without closing the settings dialog.
- [ ] 3.5 Add frontend tests for rendering default profiles, saving edits, removing profiles, reordering profiles, and handling validation errors.

## 4. Project Launch Menu

- [ ] 4.1 Replace the project add-terminal click action with a launch menu anchored to the project row.
- [ ] 4.2 Render `Terminal` first and append configured launch profiles in saved order.
- [ ] 4.3 Create a normal terminal when the user selects `Terminal`.
- [ ] 4.4 Create and activate a new terminal, then send `command + "\\n"` to that terminal when the user selects a custom launch profile.
- [ ] 4.5 Close the launch menu after selection, outside click, project deletion, or unavailable project state changes.
- [ ] 4.6 Add frontend tests for launch menu rendering, plain terminal creation, custom profile command submission, and unavailable project behavior.

## 5. Verification

- [ ] 5.1 Run Go tests covering settings, app, and shell behavior.
- [ ] 5.2 Run frontend unit tests.
- [ ] 5.3 Run the project build or Wails build used by this repository.
- [ ] 5.4 Validate the OpenSpec change.
