## Context

The application already renders one xterm.js terminal per project and routes terminal input through the existing Wails-bound `SendTerminalInput` method. The backend owns PTY lifecycle and output events, while the frontend owns terminal rendering and user interaction.

Copy and paste are frontend interaction concerns. The terminal needs access to selected text, system clipboard reads and writes, and the active project input route. Wails runtime already exposes `ClipboardGetText` and `ClipboardSetText`, so no new backend API or dependency is required.

## Goals / Non-Goals

**Goals:**

- Support `Ctrl+Shift+C` to copy selected terminal text to the system clipboard.
- Support `Ctrl+Shift+V` to paste system clipboard text into the active project shell.
- Provide a terminal context menu with Copy and Paste actions.
- Preserve normal shell behavior for `Ctrl+C`, including interrupting running commands.
- Keep clipboard actions scoped to the terminal session that owns the active project.

**Non-Goals:**

- Add clipboard history, paste confirmation, bracketed paste handling, or paste sanitization.
- Add middle-click paste or platform-specific alternate shortcuts.
- Change backend PTY lifecycle, shell process behavior, or project persistence.
- Solve unrelated terminal display issues such as color output or duplicate key echo.

## Decisions

### Use Wails Clipboard APIs

Clipboard actions will use `ClipboardSetText` for copy and `ClipboardGetText` for paste. This keeps behavior tied to the desktop runtime instead of relying on browser clipboard permissions, focus state, or secure-context rules.

Alternative considered: `navigator.clipboard`. It is simpler in browser-only apps, but Wails/WebKit clipboard permission behavior can vary and is unnecessary because the runtime API already exists.

### Keep Shortcut Handling In The Terminal Session Layer

The xterm session factory will attach a custom key handler for `Ctrl+Shift+C` and `Ctrl+Shift+V`. The handler will consume those two shortcuts and leave `Ctrl+C` untouched so the shell continues receiving interrupt signals.

The session manager will expose focused actions such as copy selection and paste text for a project session. This preserves existing project-specific routing and keeps App-level Vue code from reaching into xterm internals for every action.

Alternative considered: bind global window keydown handlers in `App.vue`. That risks firing while focus is outside the terminal and makes active-project routing less explicit.

### Render A Small Terminal Context Menu In Vue

Right-clicking the terminal area will open a menu at the pointer location with Copy and Paste actions. Copy will be disabled or ignored when the current terminal has no selection. Paste will read clipboard text and route it through the same active project input path used by keyboard paste.

The menu will close after an action, when clicking elsewhere, or when switching projects. The menu belongs to the active terminal surface and does not affect project sidebar interactions.

Alternative considered: use the native browser context menu. It provides less control over command routing and may expose browser actions that do not interact with the xterm session.

## Risks / Trade-offs

- Clipboard read/write can fail or return an empty string → show the existing error bar for failures and ignore empty paste content.
- Right-click might interfere with terminal selection workflows → only open the menu on contextmenu events and keep the menu lightweight.
- Shortcut handling could accidentally consume shell control sequences → match only `ctrlKey && shiftKey` with `C` or `V`, and leave plain `Ctrl+C` and `Ctrl+V` unchanged.
- Hidden project terminals retain session state → context menu actions target only the active project ID.

## Migration Plan

No data migration is required. The change is frontend-only and can be rolled back by removing the shortcut and context-menu wiring while leaving existing shell session behavior intact.

## Open Questions

- None.
