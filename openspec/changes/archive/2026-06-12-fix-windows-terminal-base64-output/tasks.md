## 1. Regression Coverage

- [x] 1.1 Add Go unit tests for the command-state output filter covering raw OSC BEL, raw OSC ST, Windows textual fallback, split chunks, invalid base64, and non-application output preservation.
- [x] 1.2 Add shell session manager tests proving cleaned terminal output is emitted, command-state events are produced, and terminal history excludes application-private payloads.
- [x] 1.3 Add frontend tests proving backend command-state events update terminal labels and cleaned terminal output is written without visible `tui-helper` or base64 payload text.
- [x] 1.4 Preserve existing xterm OSC 777 compatibility tests for non-filtered or legacy paths.

## 2. Backend Implementation

- [x] 2.1 Define a command-state event shape for `command-start` and `command-end`, including terminal identity and decoded command text when present.
- [x] 2.2 Implement a per-session streaming parser/filter for application-private `OSC 777;tui-helper` payloads with bounded pending buffering.
- [x] 2.3 Wire the filter into `ShellSessionManager.readOutput` so cleaned output is emitted to the frontend and appended to terminal history.
- [x] 2.4 Emit extracted command-state events from the backend through the Wails runtime without duplicating cleaned terminal output.
- [x] 2.5 Keep unknown OSC sequences and ordinary command output unchanged unless they exactly match the application-private command-state protocol.

## 3. Frontend Integration

- [x] 3.1 Listen for backend command-state events and route them through the existing terminal command-state handler.
- [x] 3.2 Ensure terminal-output event handling writes only cleaned output to xterm sessions.
- [x] 3.3 Keep the existing xterm OSC 777 handler as a defensive compatibility path without causing duplicate command-state updates.
- [x] 3.4 Regenerate or update Wails frontend bindings if the backend event/model surface requires it.

## 4. Verification

- [x] 4.1 Run Go tests covering shell sessions, terminal history, Windows adapter compilation boundaries, and command-state filtering.
- [x] 4.2 Run frontend automated tests with `npm test` from `frontend/`.
- [x] 4.3 Run `GOOS=windows GOARCH=amd64 go build ./...` to verify Windows-specific files still compile.
- [x] 4.4 Run `openspec validate fix-windows-terminal-base64-output --strict`.
- [x] 4.5 Perform an automatic review focused on filter correctness, false-positive deletion risk, event duplication risk, and test coverage.
- [x] 4.6 Run `wails build -tags webkit2_41` to generate the executable package.
