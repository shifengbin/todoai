## 1. Backend Terminal History Persistence

- [x] 1.1 Define persisted terminal history data structures for terminal metadata and capped output history.
- [x] 1.2 Add a terminal history store next to the existing project config path, including load, save, missing-file handling, and atomic writes.
- [x] 1.3 Add output append and trimming logic with a hard per-terminal history size limit.
- [x] 1.4 Extend shell session creation, selection, output, status, and deletion paths to keep persisted terminal metadata current.

## 2. Backend Restore And Cleanup Semantics

- [x] 2.1 Restore valid persisted terminal records into application state on startup without starting shell processes.
- [x] 2.2 Normalize restored terminals to a non-running state and preserve active terminal selection per TODO project context.
- [x] 2.3 Drop persisted terminal records that reference missing TODOs, TODO project associations, or projects.
- [x] 2.4 Clean persisted terminal history when deleting terminals, completing or deleting TODOs, removing TODO projects, and deleting projects.
- [x] 2.5 Add Go tests for restart restoration, active terminal restoration, capped output history, missing history storage, and cleanup paths.

## 3. Frontend Terminal History Replay

- [x] 3.1 Update Wails models and frontend state handling to accept restored terminal history from backend state.
- [x] 3.2 Update `TerminalSessionManager` so restored history is replayed once into the owning xterm session before live output continues.
- [x] 3.3 Ensure restored terminals do not trigger automatic shell startup and display as non-running/exited terminals.
- [x] 3.4 Add frontend automated tests for restored terminal output replay and avoiding duplicate replay with later live output.

## 4. TODO Menu And Clipboard Interaction

- [x] 4.1 Add a three-dot icon button to each active TODO item that opens the same menu used by right-click.
- [x] 4.2 Reuse existing TODO menu state, placement, closing behavior, and confirmation-popover interactions for both right-click and three-dot entry points.
- [x] 4.3 Rename the menu action from copying description to copying title and description.
- [x] 4.4 Update clipboard formatting so the first line is the TODO title and the remaining lines contain the description when present.
- [x] 4.5 Add frontend automated tests for the three-dot menu, shared menu actions, outside-click closing, and title/description clipboard format.

## 5. Verification And Packaging

- [x] 5.1 Run backend Go tests covering shell sessions, project state, and terminal history persistence.
- [x] 5.2 Run frontend automated tests for `App.vue`, `ProjectSidebar.vue`, and `terminalManager.js`.
- [x] 5.3 Run an automated code review pass for the change and address findings.
- [x] 5.4 Run `openspec status --change persist-terminal-history-and-todo-menu-actions` and resolve any incomplete artifacts.
- [x] 5.5 Run `wails build -tags webkit2_41` to generate the executable.
