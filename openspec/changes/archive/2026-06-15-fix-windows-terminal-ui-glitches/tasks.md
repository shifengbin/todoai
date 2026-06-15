## 1. Reproduction And Diagnostics

- [ ] 1.1 Capture the Windows TODO delete issue separately for right-click menu, three-dot menu, and delete confirmation popover.
- [ ] 1.2 Capture raw Windows terminal output samples for pure `Terminal` startup and non-empty launch profiles such as `calc`, Claude, and Codex, including any visible base64-like unknown control text and surrounding prefixes.
- [ ] 1.3 Capture Claude xterm title frames for idle startup, left/middle/right dot busy animation, completed response, and needs-input states.

## 2. TODO Delete Menu And Popover Positioning

- [x] 2.1 Add focused ProjectSidebar tests for TODO context menu placement from right-click and three-dot button triggers.
- [x] 2.2 Add or adjust ProjectSidebar tests proving the delete confirmation popover remains owned by the corresponding TODO action context.
- [x] 2.3 Fix TODO menu and delete confirmation placement so Windows WebView does not anchor either UI to the global top-right corner.
- [x] 2.4 Verify existing complete TODO, delete TODO, outside-click, and sidebar popover mutual-close behavior still works.

## 3. Terminal Control Payload Filtering

- [x] 3.1 Add Go filter tests proving unrecognized Windows launch profile base64-like/control-text-shaped output is preserved.
- [x] 3.2 Add regression tests proving ordinary base64-like command output is still rendered and persisted.
- [x] 3.3 Extend command-state/control-payload filtering to consume only recognized private or control-sequence payload shapes.
- [x] 3.4 Verify filtered payloads are excluded from visible terminal output and terminal history while valid command-state events still update labels.

## 4. Claude Activity Classification

- [x] 4.1 Add frontend tests for Claude startup title frames staying idle when only a single dot frame or stable initial title is observed.
- [x] 4.2 Add frontend tests for Claude left/middle/right dot title animation becoming busy only after a sequence is observed.
- [x] 4.3 Add frontend tests for Claude returning to idle baseline, command-end, and shell-exit clearing busy state.
- [x] 4.4 Update terminal title activity classification to use idle baseline and sequence-aware Claude dot animation detection.
- [x] 4.5 Verify explicit busy text, spinner titles, Windows/Unix path titles, and needs-input titles keep their existing intended behavior.

## 5. Windows Manual Verification

- [ ] 5.1 On Windows, verify TODO right-click menu, three-dot menu, and delete confirmation popover appear near the triggering TODO.
- [ ] 5.2 On Windows, launch `Terminal`, `calc`, Claude, and Codex profiles and verify supported application-private command-state payloads remain hidden while unrelated base64-like launch output is not hidden by heuristic.
- [ ] 5.3 On Windows, verify Claude idle startup is not shown as busy, the left/middle/right dot animation shows busy, and idle return clears busy.

## 6. Automated Verification And Review

- [x] 6.1 Run `go test ./...`.
- [x] 6.2 Run `npm test -- --run` from `frontend/`.
- [x] 6.3 Run an automated code review pass for regressions in terminal filtering, title classification, and sidebar popover behavior.
- [x] 6.4 Run `npm run build` from `frontend/`.
- [x] 6.5 Run `wails build -tags webkit2_41`.
