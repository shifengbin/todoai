## 1. Backend Settings Model

- [x] 1.1 Add an application settings model with terminal shell path, display name, source, availability, and fallback shell fields.
- [x] 1.2 Add a settings manager that loads and saves `settings.json` atomically under the existing `tui-helper` config directory.
- [x] 1.3 Implement first-load behavior that detects a shell and persists it when no terminal shell setting exists.
- [x] 1.4 Validate manual shell paths by checking that the path exists and is executable before saving.
- [x] 1.5 Add tests for first-load persistence, saved setting restoration, invalid manual path rejection, and unavailable saved shell reporting.

## 2. Shell Detection And Resolver Integration

- [x] 2.1 Implement shell detection with a deterministic priority order based on `$SHELL` and known shell paths.
- [x] 2.2 Return detected shell metadata suitable for the settings UI, including path and display name.
- [x] 2.3 Wire `ShellSessionManager` shell path resolution through the settings-backed resolver.
- [x] 2.4 Preserve per-terminal shell path capture so existing terminals keep their original shell after settings change.
- [x] 2.5 Add tests that new terminals use the saved shell and fall back to detected shell when the saved path is unavailable.

## 3. Wails API And Generated Bindings

- [x] 3.1 Add Wails-exposed methods to load terminal settings, save a selected shell path, and re-run shell detection.
- [x] 3.2 Ensure settings API errors are user-facing and do not corrupt the previous saved setting.
- [x] 3.3 Regenerate Wails frontend bindings for the new methods and settings models.
- [x] 3.4 Add or update Go app tests for settings APIs and shell resolver wiring.

## 4. Frontend Settings Interface

- [x] 4.1 Add a settings entry point to the application shell using the existing visual style and icon button patterns.
- [x] 4.2 Build a settings panel or modal focused on terminal shell selection.
- [x] 4.3 Show the current saved shell, detected candidate, unavailable saved-shell state, and validation errors.
- [x] 4.4 Allow users to select a detected shell, enter a manual executable path, re-run detection, save, and cancel.
- [x] 4.5 Apply saved settings to subsequent terminal creation without disrupting active terminal panes.

## 5. Tests And Verification

- [x] 5.1 Add Vue tests for opening settings, rendering loaded shell state, saving detected shell, saving manual path, and showing invalid path errors.
- [x] 5.2 Add Vue tests that changing the setting affects newly created terminals only through the backend state refresh path.
- [x] 5.3 Update existing frontend tests for any new app-shell controls or Wails mocks.
- [x] 5.4 Run Go tests for backend settings and shell session behavior.
- [x] 5.5 Run frontend tests for settings UI and terminal workflow behavior.
- [x] 5.6 Run `openspec validate add-terminal-settings --strict`.
