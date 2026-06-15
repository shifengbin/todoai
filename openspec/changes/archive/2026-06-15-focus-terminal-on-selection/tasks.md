## 1. Test Coverage

- [x] 1.1 Update the App frontend terminal-selection test to assert that clicking a terminal row focuses the selected xterm session after activation.
- [x] 1.2 Add or adjust a failure-path test if needed to ensure focus is not moved when `SelectTerminal` rejects.

## 2. Frontend Implementation

- [x] 2.1 Update `frontend/src/App.vue` so `selectTerminal(terminalId)` calls the existing terminal manager focus behavior only after `SelectTerminal` succeeds and the active terminal has been activated.
- [x] 2.2 Keep `activateActiveTerminal()` and `TerminalSessionManager.activate()` from focusing automatically on unrelated activation paths.

## 3. Verification

- [x] 3.1 Run client automated tests with `cd frontend && npm test`.
- [x] 3.2 Run automatic review of the completed diff to check scope, focus-stealing regressions, and test coverage.

## 4. Packaging

- [x] 4.1 Run `wails build -tags webkit2_41` to generate the executable.
