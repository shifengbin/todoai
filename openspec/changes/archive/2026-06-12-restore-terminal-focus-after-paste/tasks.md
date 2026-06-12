## 1. Terminal Focus Behavior

- [x] 1.1 Add a safe `focus(terminalId)` method to `frontend/src/terminalManager.js` that calls the owning xterm instance's `focus` method when available.
- [x] 1.2 Update the terminal context-menu Paste flow in `frontend/src/App.vue` so it closes the menu, waits for Vue DOM updates, and restores focus to the pasted terminal.
- [x] 1.3 Ensure empty clipboard Paste from the context menu closes the menu and restores terminal focus without sending shell input.

## 2. Automated Tests

- [x] 2.1 Extend `frontend/src/terminalManager.test.js` fake terminal support and test that `TerminalSessionManager.focus()` targets the correct terminal session.
- [x] 2.2 Extend `frontend/src/App.test.js` to verify context-menu Paste sends clipboard text, closes the menu, and restores terminal focus.
- [x] 2.3 Add or update a component test for empty clipboard context-menu Paste so no input is sent and terminal focus is restored.
- [x] 2.4 Run the frontend automated test suite for the affected tests.

## 3. Quality Review And Packaging

- [x] 3.1 Perform an automated code review pass for the changed frontend files and tests, checking behavior, regressions, and test coverage.
- [x] 3.2 Run `wails build -tags webkit2_41` to generate the executable package.
