## Why

Users need control over which shell the embedded terminal starts, and the current fallback only checks `$SHELL`, `/bin/bash`, and `/bin/sh`. A persisted terminal setting lets first-time users get a sensible automatically detected shell while allowing later sessions to honor the user's explicit choice.

## What Changes

- Add a settings interface for viewing and changing the embedded terminal shell.
- Automatically detect a usable shell the first time the application opens when no terminal setting exists.
- Persist the selected shell so future application launches use the saved setting instead of re-detecting.
- Provide a way to re-run automatic detection from settings.
- Apply the configured shell to newly created embedded terminals.
- Keep existing terminal sessions stable; changing the setting does not mutate already created running terminals.

## Capabilities

### New Capabilities

- `terminal-settings`: Covers terminal settings UI, persisted shell preference, first-run automatic shell detection, and manual shell selection.

### Modified Capabilities

- `embedded-shell-sessions`: Shell session startup uses the configured or automatically detected shell path instead of the existing hard-coded default resolver alone.

## Impact

- Go backend: application settings model, settings persistence, shell detection, Wails-exposed settings methods, and shell path resolver wiring.
- Vue frontend: settings entry point, settings panel or modal, terminal shell selection controls, state loading/saving, and tests.
- Generated Wails bindings: new settings APIs and models.
- OpenSpec specs: new terminal settings capability and updated embedded shell session startup requirements.
