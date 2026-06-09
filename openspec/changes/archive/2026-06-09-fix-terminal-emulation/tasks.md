## 1. Regression Tests

- [x] 1.1 Add a backend test proving shell startup overrides inherited `TERM=dumb` with `TERM=xterm-256color` and sets `COLORTERM=truecolor`.
- [x] 1.2 Add a frontend xterm factory test proving PTY-backed terminal sessions are created without `convertEol`.
- [x] 1.3 Add a backend concurrency test proving overlapping `EnsureSession` calls for the same project start only one PTY process.

## 2. Backend Terminal Environment

- [x] 2.1 Add a helper that merges inherited environment variables with embedded-terminal overrides for `TERM` and `COLORTERM`.
- [x] 2.2 Use the normalized environment when starting PTY shell processes.
- [x] 2.3 Keep existing shell path selection, working directory, output routing, resize routing, and exit status behavior unchanged.

## 3. Frontend PTY Rendering

- [x] 3.1 Remove `convertEol: true` from xterm.js terminal creation for embedded shell sessions.
- [x] 3.2 Confirm clipboard shortcut handling still intercepts only `Ctrl+Shift+C` and `Ctrl+Shift+V`.

## 4. Shell Startup Concurrency

- [x] 4.1 Make `ShellSessionManager.EnsureSession` atomic for a project so concurrent calls cannot create duplicate live sessions.
- [x] 4.2 Ensure output and wait goroutines are still started exactly once for the stored session.

## 5. Verification

- [x] 5.1 Run backend tests with `go test ./...`.
- [x] 5.2 Run frontend tests with `npm test` from `frontend`.
- [x] 5.3 Run OpenSpec validation for `fix-terminal-emulation`.
- [ ] 5.4 Manually verify in the running app with zsh that colors render correctly, Backspace edits input, and `clear` removes prior command text without residual characters.
