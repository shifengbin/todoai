## Context

This is a new desktop application with no existing application code in the workspace. The target stack is Vue for the frontend, Go for backend services, and Wails for desktop integration. The application needs a persistent list of opened local projects and an embedded shell whose working directory follows the selected project.

The important behavioral boundary is that shell sessions are runtime state only. Each opened project keeps its own live shell session while the application is running, and switching projects restores that session. After the application exits, only the project list is restored; shell processes, running commands, and output history are not restored.

## Goals / Non-Goals

**Goals:**

- Provide a left-side project list with a create action that opens a native directory chooser.
- Store opened projects locally and reload them on the next application start.
- Start a shell in the selected project's directory.
- Keep one shell session per project alive while the app runs.
- Restore the visible terminal state when switching back to a project during the same app session.
- Produce a Linux `.deb` package for installation.

**Non-Goals:**

- Restore shell processes or command output after an application restart.
- Support multiple shells per project, split panes, tabs, or remote hosts.
- Provide project indexing, Git status, or file browsing.
- Use a database for the initial project list.

## Decisions

### Use Wails as the application boundary

The frontend will render the application UI with Vue, while Go will own native desktop operations such as directory selection, local configuration storage, PTY creation, shell process lifecycle, and packaging integration.

Alternative considered: have the frontend launch or emulate shell behavior directly. That does not fit browser security boundaries and would still require Go-side PTY control in Wails.

### Store projects as local JSON configuration

The backend will persist opened projects in a JSON file under the application config directory. Each project record will include a stable ID, display name, absolute path, and timestamps such as creation or last selection time. The display name defaults to the selected directory basename.

Alternative considered: SQLite. It is unnecessary for the initial scope because the data model is small, single-user, and append/update oriented.

### Manage shell sessions in Go with a PTY per project

The backend will maintain an in-memory map of `projectId -> shell session`. A session owns the PTY, shell process, output stream, current size, and status. Selecting a project starts a session if none exists; otherwise it re-attaches to the existing session.

The shell command should use the user's `$SHELL` when available and fall back to `/bin/bash` or `/bin/sh`. The process working directory must be the selected project path.

Alternative considered: close and recreate the shell on every project switch. That loses running commands and violates the requested restore behavior.

### Route terminal IO by project ID

PTY output will be emitted to the frontend with the associated project ID. The Vue terminal manager will keep terminal state per project and append output to the correct terminal even when that project is not visible. When the selected project changes, the UI swaps to the matching terminal instance.

Input, resize, and restart requests will include the target project ID. This avoids accidental writes to the wrong shell during fast project switching.

### Keep output memory bounded

Terminal state can grow quickly if commands produce large output. The frontend terminal should use xterm.js scrollback limits, and the backend should keep only a small recent output buffer if replay is needed after a frontend component remount.

### Package for Linux as a deb artifact

The implementation will use the Wails Linux packaging flow available for the chosen Wails version and must produce an installable `.deb`. Packaging metadata should include app name, version, architecture, maintainer, description, desktop launcher metadata, and required runtime dependencies.

The exact Wails CLI version should be confirmed during implementation. If Wails v3 is used, prefer its Linux packaging tasks. If Wails v2 is used, keep the same acceptance behavior and adapt the packaging command accordingly.

## Risks / Trade-offs

- PTY process leaks on project removal or app exit -> centralize session shutdown and test lifecycle cleanup.
- Background projects can produce large output -> cap terminal scrollback and backend replay buffers.
- A persisted project path may be deleted or become inaccessible -> mark the project unavailable and avoid launching a shell until the path is valid.
- Shell exits are normal process behavior -> show an exited status and provide a way to restart the session.
- `.deb` packaging can vary by Wails version and Linux environment -> verify the actual build command in the target environment before treating packaging as complete.

## Migration Plan

This is a new application, so no existing user data needs migration. Initial implementation should create the config file on first write and tolerate a missing config file on startup. Future config format changes should add a version field and migration path.

## Open Questions

- Which exact Wails major version should be used for the initial scaffold?
- What package name, app display name, icon, and maintainer metadata should be used for the `.deb`?
