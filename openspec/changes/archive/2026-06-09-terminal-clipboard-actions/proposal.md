## Why

The embedded terminal currently lacks basic copy and paste interactions, which makes it hard to reuse command output or paste prepared commands into an active shell. Terminal users expect clipboard actions to work without taking over `Ctrl+C`, because `Ctrl+C` must remain available for interrupting shell processes.

## What Changes

- Add `Ctrl+Shift+C` support for copying selected terminal text to the system clipboard.
- Add `Ctrl+Shift+V` support for pasting system clipboard text into the active project shell.
- Add a terminal context menu with Copy and Paste actions.
- Keep `Ctrl+C` and `Ctrl+V` behavior available to the shell instead of using them for clipboard shortcuts.
- Do not add multi-shell, split-pane, or clipboard history features.

## Capabilities

### New Capabilities

### Modified Capabilities

- `embedded-shell-sessions`: Adds clipboard copy and paste interactions for the embedded terminal while preserving shell control-key behavior.

## Impact

- Frontend terminal session creation and input routing.
- Vue application state for displaying and closing the terminal context menu.
- Wails runtime clipboard APIs for reading from and writing to the system clipboard.
- Frontend unit tests for terminal shortcut handling, context-menu actions, and active-project input routing.
