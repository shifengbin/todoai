## 1. Application Identity

- [x] 1.1 Update Wails project name, output filename, native window title, frontend document title, and README product metadata from `TUI Helper` / `tui-helper` to `TodoAI` / `todoai`.
- [x] 1.2 Add or update tests that verify the default app identity values used by Go/Wails-facing code where practical.

## 2. Local Data Migration

- [x] 2.1 Change the default configuration directory from `tui-helper` to `todoai`.
- [x] 2.2 Implement legacy directory migration so existing `projects.json`, `settings.json`, and terminal history remain available when only the old `tui-helper` directory exists.
- [x] 2.3 Preserve an existing `todoai` directory without overwriting it when both old and new directories exist.
- [x] 2.4 Add Go tests for new-install path resolution, legacy migration, and no-overwrite behavior.

## 3. Command-State Protocol Compatibility

- [x] 3.1 Update shell integration payload generation to emit the `todoai` command-state identifier.
- [x] 3.2 Update backend command-state output filtering to consume both new `todoai` and legacy `tui-helper` raw/textual payloads.
- [x] 3.3 Update frontend xterm OSC command-state parsing to accept both new `todoai` and legacy `tui-helper` identifiers.
- [x] 3.4 Add backend tests for raw OSC, Windows textual fallback, invalid payload, split payload, legacy payload compatibility, and non-application output preservation.
- [x] 3.5 Add frontend automated tests for new and legacy OSC identifier handling.

## 4. Packaging And Icon Assets

- [x] 4.1 Save the generated TodoAI icon into the project as `build/appicon.png`.
- [x] 4.2 Generate and save a matching multi-size Windows icon at `build/windows/icon.ico`.
- [x] 4.3 Update Debian packaging to build and package `todoai`, install `/usr/bin/todoai`, install the `todoai` icon, and generate a `TodoAI` desktop launcher.
- [x] 4.4 Update Debian packaging tests and fixtures for the new binary name, package path, control metadata, desktop launcher metadata, and icon path.

## 5. Verification And Review

- [x] 5.1 Run Go tests with `go test ./...`.
- [x] 5.2 Run frontend automated tests with `npm run test` from `frontend`.
- [x] 5.3 Run frontend production build with `npm run build` from `frontend`.
- [x] 5.4 Run Debian packaging script tests with `scripts/package-deb.test.sh`.
- [x] 5.5 Run an automated review pass focused on branding consistency, data migration safety, command-state compatibility, and packaging regressions.
- [x] 5.6 Run `wails build -tags webkit2_41` to generate the executable.
