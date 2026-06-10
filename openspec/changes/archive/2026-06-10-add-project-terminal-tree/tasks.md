## 1. Backend Terminal Model

- [x] 1.1 Add a runtime terminal descriptor model with stable `terminalId`, owning `projectId`, shell name, command label state, shell state, and selection timestamps.
- [x] 1.2 Refactor `ShellSessionManager` to key sessions by `terminalId` while retaining owning project metadata for working directory and events.
- [x] 1.3 Add backend operations for creating a terminal, selecting or ensuring a project's default terminal, restarting an exited terminal, writing input, resizing, and reporting status by `terminalId`.
- [x] 1.4 Update terminal output and status runtime events to include both `projectId` and `terminalId`.
- [x] 1.5 Add shell integration startup support for `zsh` and `bash` that emits app-specific OSC command-start and command-end messages, with fallback shell-name behavior for unsupported shells.
- [x] 1.6 Regenerate Wails frontend bindings after backend method and model changes.

## 2. Frontend Tree And Routing

- [x] 2.1 Update app state to track `activeProjectId`, `activeTerminalId`, terminal descriptors, and terminal statuses separately.
- [x] 2.2 Refactor `TerminalSessionManager` to key xterm sessions, writes, resize, copy, paste, and shortcut handling by `terminalId`.
- [x] 2.3 Update `App.vue` to create, activate, render, resize, restart, and route events for terminal panes by `terminalId`.
- [x] 2.4 Update `ProjectSidebar` to render projects as tree parents and terminals as selectable child rows, including an add-terminal action per available project.
- [x] 2.5 Register an xterm custom OSC handler for app command-state messages and update the owning terminal descriptor's current command label.
- [x] 2.6 Ensure terminal labels derive from current command when present and shell name when idle, exited, or command state is unavailable.

## 3. Tests And Verification

- [x] 3.1 Update Go project and shell session tests for multiple terminals under one project, terminal-scoped input/output/resize/status, default terminal creation, and unavailable project behavior.
- [x] 3.2 Add Go tests for shell integration command-state hook generation and fallback shell naming.
- [x] 3.3 Update frontend unit tests for tree rendering, terminal selection, add-terminal behavior, and active project derivation.
- [x] 3.4 Update terminal manager tests for terminal-scoped routing, inactive terminal output retention, clipboard actions, resize reporting, and command label OSC handling.
- [x] 3.5 Run `go test ./...` and `npm test` in `frontend/`.
