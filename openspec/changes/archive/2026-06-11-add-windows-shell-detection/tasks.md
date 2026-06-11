## 1. Backend Tests

- [x] 1.1 Add `ShellDetector` tests for Windows priority order: `pwsh.exe`, `powershell.exe`, `COMSPEC`/`cmd.exe`, and no Unix-only fallback.
- [x] 1.2 Add `ShellDetector` tests proving non-Windows detection still prefers `$SHELL` and known Unix candidates.
- [x] 1.3 Add shell path availability tests for Windows executable extensions, `PATHEXT`, missing paths, directories, and non-executable extensions.
- [x] 1.4 Add shell path availability tests proving non-Windows validation still requires execute permission.
- [x] 1.5 Add `DefaultShellPath` or resolver tests showing Windows fallback resolves to a Windows shell path and never `/bin/sh`.
- [x] 1.6 Move or guard Unix-only real PTY process tests so `GOOS=windows GOARCH=amd64 go test -c` can compile the test package.

## 2. Backend Implementation

- [x] 2.1 Refactor shell detection to accept injectable platform, environment, lookup, candidates, and path availability checks for deterministic tests.
- [x] 2.2 Implement Windows shell candidate discovery for `pwsh.exe`, `powershell.exe`, `SystemRoot` system paths, `COMSPEC`, and `cmd.exe`.
- [x] 2.3 Implement platform-aware shell path availability while preserving Unix execute-bit behavior.
- [x] 2.4 Update `SettingsManager` shell detection, saved-shell availability, and manual save validation to use platform-aware availability.
- [x] 2.5 Update `DefaultShellPath` to reuse platform-aware detection and avoid Unix-only fallback on Windows.
- [x] 2.6 Normalize Windows shell display names so detected paths show useful names such as `pwsh`, `powershell`, or `cmd`.

## 3. Session Startup Boundary

- [x] 3.1 Add or update session resolver coverage proving unavailable saved shell settings use a Windows detected fallback on Windows.
- [x] 3.2 Ensure Windows PTY unsupported errors remain visible and are not hidden by substituting Unix shell paths.
- [x] 3.3 Preserve existing non-Windows shell integration behavior for `zsh`, `bash`, and unsupported shells.

## 4. Client And Generated Surface

- [x] 4.1 Confirm the existing Wails settings API and `TerminalShellSetting` model remain compatible; regenerate Wails bindings only if backend exported types change.
- [x] 4.2 Run the frontend/client automated test suite to verify terminal settings UI continues to render detected shell names and validation errors correctly.

## 5. Verification And Review

- [x] 5.1 Run `go test ./...`.
- [x] 5.2 Run `GOOS=windows GOARCH=amd64 go build ./...`.
- [x] 5.3 Run `GOOS=windows GOARCH=amd64 go test -c -o /tmp/tui-helper-windows.test.exe .`.
- [x] 5.4 Run an automated code review pass over the implementation and address correctness, portability, and test coverage findings.
