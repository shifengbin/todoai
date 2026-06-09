## 1. Project Setup

- [x] 1.1 Confirm Wails major version, app name, package name, icon placeholder, and `.deb` metadata values.
- [x] 1.2 Scaffold the Wails application with Vue frontend and Go backend in the workspace.
- [x] 1.3 Add frontend terminal dependencies, including xterm.js and a fit/resize addon.
- [x] 1.4 Add Go PTY dependency and confirm it builds on Linux.
- [x] 1.5 Add baseline frontend and backend test/build commands.

## 2. Project Workspace

- [x] 2.1 Define the project model with stable ID, display name, absolute path, availability status, and timestamps.
- [x] 2.2 Implement local JSON config storage under the application config directory.
- [x] 2.3 Implement startup loading for the persisted project list, including missing-file handling.
- [x] 2.4 Implement native directory selection through the Wails backend.
- [x] 2.5 Implement project creation from selected directory with basename default name.
- [x] 2.6 Prevent duplicate project entries for the same absolute path and select the existing entry instead.
- [x] 2.7 Detect missing or inaccessible persisted project paths and mark them unavailable.
- [x] 2.8 Build the left-side Vue project list with new-project and selection behavior.

## 3. Shell Session Backend

- [x] 3.1 Implement a Go shell session manager keyed by project ID.
- [x] 3.2 Start shell processes through a PTY using the selected project path as the working directory.
- [x] 3.3 Use the user's configured shell when available, with Linux shell fallbacks.
- [x] 3.4 Keep one live session per project while the app is running and reuse it on project reselection.
- [x] 3.5 Emit shell output events with project ID metadata.
- [x] 3.6 Accept terminal input requests with project ID metadata and write to the matching PTY.
- [x] 3.7 Accept terminal resize requests and resize the matching PTY.
- [x] 3.8 Detect shell process exit and update session status without closing the app.
- [x] 3.9 Shut down all live PTY sessions cleanly on application exit.

## 4. Terminal Frontend

- [x] 4.1 Build the main layout with fixed left project sidebar and central terminal area.
- [x] 4.2 Create xterm.js terminal instances per project and preserve them while the app runs.
- [x] 4.3 Route backend output events to the matching project terminal, including inactive projects.
- [x] 4.4 Route user input from the active terminal to the active project session.
- [x] 4.5 Fit and resize the active terminal when the window or container size changes.
- [x] 4.6 Show unavailable-project and exited-shell states in the terminal area.
- [x] 4.7 Ensure project switching restores the previous runtime terminal state.

## 5. Linux Debian Packaging

- [x] 5.1 Configure Wails/Linux packaging metadata for `.deb` output.
- [x] 5.2 Add or document the Linux packaging command that produces a `.deb` artifact.
- [x] 5.3 Verify the `.deb` artifact is generated in the expected build output directory.
- [x] 5.4 Verify the installed package provides a desktop-launchable application.

## 6. Verification

- [x] 6.1 Add backend tests for project config load/save, duplicate path handling, and missing path detection.
- [x] 6.2 Add backend tests or integration coverage for shell session reuse, input routing, resize routing, and exit status.
- [x] 6.3 Add frontend tests for project selection, terminal output routing, and runtime terminal restoration.
- [x] 6.4 Run backend tests, frontend tests, frontend build, and Wails build.
- [x] 6.5 Run OpenSpec validation for `desktop-project-shell`.
