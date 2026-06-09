## Context

The app renders embedded project shells with xterm.js while the backend owns zsh/PTY startup through `creack/pty`. Today the shell process inherits the desktop process environment verbatim. In this workspace that exposes `TERM=dumb` and no `COLORTERM`, so zsh and terminal utilities do not emit xterm-compatible control sequences even though the frontend can render them.

The frontend also enables xterm.js `convertEol`, which is intended for non-PTY data sources. For a real PTY, termios already owns newline conversion. Keeping a second conversion layer can make cursor positioning and line redraw behavior harder to reason about.

There is also a lifecycle race: the Vue mounted path and terminal ref callback can both activate the selected project. The backend checks for an existing running session, unlocks, then starts the PTY, which can allow duplicate starts for the same project under overlapping calls.

## Goals / Non-Goals

**Goals:**

- Ensure embedded shells see an xterm-compatible terminal environment regardless of the desktop process `TERM`.
- Restore expected zsh editing, color, and clear-screen behavior in the embedded terminal.
- Use xterm.js as a PTY renderer without redundant newline conversion.
- Maintain at most one live shell process per project, including concurrent activation attempts.
- Add focused regression coverage for the environment, terminal options, and lifecycle race.

**Non-Goals:**

- Add terminal themes, profile configuration, or user-selectable shell types.
- Change clipboard shortcuts or context-menu behavior.
- Add bracketed paste, shell command sanitization, or terminal multiplexing.
- Persist terminal output or shell process state after application restart.

## Decisions

### Normalize PTY terminal capability variables in the backend

Backend shell startup will build the process environment from `os.Environ()` plus explicit terminal capability overrides. The embedded PTY should set `TERM=xterm-256color` and `COLORTERM=truecolor`, replacing inherited values such as `TERM=dumb`.

Alternative considered: require users to launch the desktop app from a terminal with the right environment. That makes behavior depend on launch method and fails for desktop launcher usage.

Alternative considered: configure the user's `.zshrc` to override `TERM`. That hides an app integration bug in user shell configuration and can affect unrelated terminals.

### Treat xterm.js output as real PTY output

The xterm session factory will stop enabling `convertEol` for PTY-backed shell output. Termios and the PTY stream should remain the source of truth for line endings, cursor moves, redraws, and clear-screen control sequences.

Alternative considered: leave `convertEol` enabled and only fix `TERM`. That may improve colors and `clear`, but keeps a non-PTY rendering option in the data path and risks future redraw artifacts.

### Serialize shell startup per project

`ShellSessionManager.EnsureSession` will make the "check existing session or start one" path atomic for a project so concurrent activation calls cannot start duplicate PTYs. The simplest acceptable implementation is to keep the session manager lock through the fast PTY start path, then store the session before output and wait goroutines begin.

Alternative considered: add a separate `starting` state with waiters. That is more flexible for long startup paths, but the current app only starts local shells and does not need the extra coordination complexity.

## Risks / Trade-offs

- Some user shell configuration may branch on the inherited `TERM` value → Use standard `xterm-256color`, which matches xterm.js and common terminal tooling expectations.
- Holding the session manager lock during PTY start can briefly block other shell operations → Shell startup is local and rare; if this becomes slow later, introduce a per-project starting state.
- Removing `convertEol` could expose commands that only print bare `\n` without PTY termios conversion → The embedded shell is PTY-backed, so termios should handle normal shell output.
- Manual symptoms such as Backspace are hard to fully assert in unit tests → Cover the data-path contracts in tests and leave an explicit manual verification task for the running app.

## Migration Plan

No data migration is required. Existing project configuration remains valid. The change affects new shell processes; already-running sessions in a live app will need to be restarted or the app relaunched to receive the normalized terminal environment.

Rollback is limited to restoring inherited environment behavior and the previous xterm option, though doing so would reintroduce the observed terminal display problems.

## Open Questions

- None.
