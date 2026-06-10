## Context

The application currently starts embedded PTY sessions through `ShellSessionManager`. New terminal descriptors capture a `shellPath` at creation time by calling `DefaultShellPath()`, which checks `$SHELL`, then `/bin/bash`, then `/bin/sh`. There is no application-level settings model, no settings UI, and no persisted user preference for the embedded terminal shell.

This change crosses backend settings persistence, shell path resolution, Wails bindings, and Vue UI. It should remain compatible with the current runtime terminal model: existing terminal descriptors keep their shell path after creation, while future terminal creation uses the current resolver.

## Goals / Non-Goals

**Goals:**

- Provide a settings interface for the embedded terminal shell.
- Detect a sensible shell automatically when no saved setting exists.
- Persist the detected or selected shell so subsequent launches use a stable value.
- Let users re-run detection or manually choose a shell path.
- Apply the selected shell to newly created embedded terminals.
- Keep shell integration support for `zsh` and `bash`, with usable fallback behavior for other shells.

**Non-Goals:**

- Launch an external terminal emulator such as Warp, GNOME Terminal, Kitty, or Alacritty.
- Change already running terminal sessions when the setting is saved.
- Persist live PTY processes, scrollback, or command history.
- Add per-project or per-terminal shell preferences.
- Add shell-specific command integration for every supported shell.

## Decisions

### Add a dedicated application settings model

Introduce a persisted settings record separate from `projects.json`, for example `settings.json` under the same `tui-helper` config directory. The settings record should include the selected terminal shell path and enough metadata for the UI to explain whether it was detected or manually selected.

Alternative considered: extend `projects.json`. Keeping project state and application preferences separate avoids coupling unrelated migrations and keeps the existing project persistence contract smaller.

### Persist the first automatic detection result

On first load, when no terminal shell setting exists, the backend should detect the shell and save it immediately. Later application launches should use the saved value rather than re-running detection every time.

Alternative considered: detect on every startup until the user manually saves a value. That can make behavior drift with environment changes and makes "what shell will I get?" harder to reason about.

### Prefer explicit saved settings over automatic detection

If a terminal shell path has been saved, the shell resolver should use that path for new terminal descriptors. If the saved path is missing or not executable, the backend should surface that state to the settings UI and fall back to automatic detection for shell startup rather than failing all new terminals.

Alternative considered: fail terminal startup when the saved shell is unavailable. That makes configuration errors obvious, but it can leave the app with no usable terminal after a shell package is removed or a path changes.

### Settings changes affect future terminals only

Each `ProjectTerminal` already stores its `shellPath` at creation time. Saving a new setting should not mutate existing running terminals or restart PTYs. New terminal descriptors created after the save should use the new path.

Alternative considered: restart or rewrite all terminal descriptors immediately. That would make the setting visibly global, but it risks killing user work and introduces ambiguous behavior for active shells.

### Keep shell detection backend-owned

Detection should live in Go near shell launching, not in the Vue layer. The backend can validate executable paths, inspect environment variables, and share the same resolver used by terminal creation.

Alternative considered: detect shells in the frontend. That would require exposing filesystem probes through additional APIs anyway, while still duplicating backend validation logic.

## Risks / Trade-offs

- Saved shell path becomes unavailable -> Report unavailable status in settings and use a detected fallback for new terminal startup.
- Manual path points to a non-shell executable -> Validate that the path exists and is executable, but do not try to prove interactive shell behavior before saving.
- Settings UI can become too broad -> Scope the page to terminal shell selection for this change.
- Changing settings does not alter existing terminals -> Make the UI behavior clear through state and tests; new terminal creation is the observable boundary.
- Concurrent first-run detection could race with settings writes -> Guard settings load/save with a mutex and write atomically through a temporary file.

## Migration Plan

No existing config migration is required. On first settings load, create a default settings file with the automatically detected shell. Existing `projects.json` files continue to load unchanged.

Rollback can remove the settings UI and settings APIs. If `settings.json` remains on disk after rollback, older versions ignore it.

## Open Questions

- None.
