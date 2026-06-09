## Why

Developers need a small desktop tool that keeps a list of local project directories and opens an embedded shell in the selected project. This removes repeated terminal setup while keeping each project shell available during the app session.

## What Changes

- Add a Wails desktop application using Vue for the UI and Go for backend services.
- Add a left-side project list with a create action that lets the user choose a local directory.
- Use the selected directory name as the default project name.
- Persist the opened project list across application restarts.
- Add an embedded shell area that starts in the selected project's directory.
- Keep one shell session per opened project while the application is running, so switching projects restores that project's active shell session.
- Do not restore shell processes or command output after the application exits.
- Add Linux packaging support that produces a `.deb` package.

## Capabilities

### New Capabilities

- `project-workspace`: Manage opened project directories, default names, selection state, and persisted project list.
- `embedded-shell-sessions`: Provide embedded per-project shell sessions that remain alive while the app is running and restore on project switch.
- `linux-deb-packaging`: Build the desktop application into a Linux `.deb` package.

### Modified Capabilities

None.

## Impact

- Introduces a Wails desktop app structure with Vue frontend and Go backend.
- Adds frontend dependencies for terminal rendering, expected to use `xterm.js`.
- Adds Go-side shell/PTY management dependencies, expected to use a PTY library such as `creack/pty`.
- Adds local configuration storage for opened projects.
- Adds build and packaging configuration for Linux `.deb` output.
