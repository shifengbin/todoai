## 1. Title Event Pipeline

- [x] 1.1 Extend `createXtermSession` to accept an optional title-change callback and subscribe to xterm `onTitleChange`.
- [x] 1.2 Extend `TerminalSessionManager` to receive title-change callbacks from sessions and forward them with the owning terminalId.
- [x] 1.3 Add unit coverage for terminal-scoped title-change routing, including inactive terminal title updates.

## 2. Activity State Model

- [x] 2.1 Add front-end runtime fields for each terminal's latest `runtimeTitle` and `activityState`.
- [x] 2.2 Preserve runtime title/activity fields across `applyState` merges for the same terminalId.
- [x] 2.3 Implement a title activity classifier that maps title changes to `idle`, `busy`, or `needs-input` while keeping `currentCommand` as the stable label source.
- [x] 2.4 Clear runtime title/activity state when a new shell command starts, when a command ends, and when the shell exits.
- [x] 2.5 Add component tests for busy, needs-input, idle restoration, and command-end cleanup behavior.

## 3. Project Tree Presentation

- [x] 3.1 Update `ProjectSidebar.vue` terminal rows to include a fixed-width activity indicator slot.
- [x] 3.2 Add terminal row classes or attributes for `busy` and `needs-input` states without changing the main label text.
- [x] 3.3 Add CSS for a running animation and an attention indicator while keeping row dimensions stable.
- [x] 3.4 Add accessible labeling or title text so activity state is not conveyed only by animation or color.
- [x] 3.5 Add sidebar tests covering displayed indicators and stable terminal labels.

## 4. Verification

- [x] 4.1 Run frontend unit tests for terminal manager, xterm factory, app, and project sidebar.
- [x] 4.2 Run full project test suite.
- [x] 4.3 Manually verify with an interactive title-changing command or `codex`: busy title changes animate, `!` shows attention, and returning to the command title stops the animation.
