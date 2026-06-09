## Why

The embedded shell currently inherits the desktop process terminal environment, which can expose `TERM=dumb` to zsh even though the UI renders with xterm.js. This causes incorrect colors, broken `clear` behavior, leftover input text, and unreliable editing feedback such as Backspace appearing not to work.

## What Changes

- Start embedded PTY shells with an xterm-compatible terminal environment instead of blindly trusting inherited terminal capability variables.
- Align xterm.js output handling with real PTY behavior so line endings and cursor positioning remain stable.
- Prevent duplicate shell startup for the same project when frontend activation paths overlap.
- Add regression coverage for terminal environment normalization, PTY line-ending behavior, and single live shell startup per project.

## Capabilities

### New Capabilities

- `embedded-terminal-emulation`: Covers terminal capability negotiation, shell editing behavior, clear-screen behavior, color support, and one-live-session guarantees for embedded PTY-backed terminals.

### Modified Capabilities

## Impact

- Backend shell startup in `shell.go`, especially PTY environment construction and session concurrency.
- Frontend terminal creation in `frontend/src/xtermFactory.js`, especially xterm.js options for PTY-backed output.
- Backend and frontend tests for shell session lifecycle, terminal options, and input/rendering regressions.
- No new external runtime dependencies are expected.
