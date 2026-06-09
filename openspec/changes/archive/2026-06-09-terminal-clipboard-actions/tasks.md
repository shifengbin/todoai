## 1. Clipboard Routing Tests

- [x] 1.1 Add terminal session manager tests for copying selected terminal text through an injected clipboard writer.
- [x] 1.2 Add terminal session manager tests for pasting non-empty clipboard text into the owning project shell.
- [x] 1.3 Add tests proving plain `Ctrl+C` remains routed to the shell instead of the clipboard handler.

## 2. Keyboard Shortcut Implementation

- [x] 2.1 Extend terminal session creation to intercept only `Ctrl+Shift+C` and `Ctrl+Shift+V`.
- [x] 2.2 Add session manager copy and paste actions that use Wails clipboard functions and existing project input routing.
- [x] 2.3 Surface clipboard failures through the existing application error display.

## 3. Context Menu Implementation

- [x] 3.1 Add Vue state for terminal context menu position, target project, and visibility.
- [x] 3.2 Render Copy and Paste menu actions in the terminal area and close the menu after actions or outside clicks.
- [x] 3.3 Disable or ignore Copy when the active terminal has no selected text.
- [x] 3.4 Route context-menu Paste through the same paste path as `Ctrl+Shift+V`.

## 4. Verification

- [x] 4.1 Run frontend tests for terminal manager and application interaction behavior.
- [x] 4.2 Run backend tests to confirm no shell session regressions.
- [ ] 4.3 Manually verify `Ctrl+Shift+C`, `Ctrl+Shift+V`, right-click Copy, right-click Paste, and plain `Ctrl+C` in the running app.
