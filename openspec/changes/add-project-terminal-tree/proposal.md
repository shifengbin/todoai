## Why

Users need to keep multiple independent terminals open for the same project so they can run concurrent commands without losing shell state when switching context. The current model allows only one live terminal per project, which makes common workflows like running tests, a dev server, and ad-hoc commands awkward.

## What Changes

- Allow each opened project to own multiple independent terminal sessions.
- Change the left sidebar from a flat project list into a project tree where terminal sessions appear under their project.
- Let users create, select, and switch between terminal sessions within the same project.
- Route terminal input, output, resize, status, clipboard, and restart actions by terminal session instead of only by project.
- Display each terminal's label as the currently executing command while a command is running.
- Restore the terminal label to the shell name, such as `zsh`, `bash`, or `sh`, after the command finishes.
- **BREAKING**: Backend shell APIs and runtime terminal events will need to include a terminal identifier in addition to the project identifier.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `project-workspace`: The workspace selector changes from selecting a flat project row to selecting terminals inside a project tree while keeping project context visible.
- `embedded-shell-sessions`: Shell sessions change from at most one live session per project to multiple independent terminal sessions per project, each with terminal-scoped routing and command-aware labels.

## Impact

- Go backend: project state models, shell session manager keys, Wails-exposed shell methods, runtime terminal events, and tests.
- Vue frontend: sidebar tree component, active project and terminal state, terminal container mapping, xterm session manager keys, context menu behavior, and tests.
- OpenSpec specs: project workspace selection requirements and embedded shell session lifecycle/routing requirements.
