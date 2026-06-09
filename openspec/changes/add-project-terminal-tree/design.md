## Context

The current application has a flat project list and a single terminal session per project. Backend shell sessions are stored in `ShellSessionManager.sessions` keyed by `projectID`; frontend xterm sessions are also keyed by `projectId`; Wails methods and runtime events carry only `projectId`. That model preserves one shell while switching projects, but it cannot represent multiple independent shells inside the same project.

The new sidebar shape is a project tree:

```text
Projects
└─ demo-app
   ├─ npm run dev
   ├─ go test ./...
   └─ zsh
```

Project identity and terminal identity need to be separated so selecting a project and selecting a terminal do not collapse into the same state.

## Goals / Non-Goals

**Goals:**

- Support multiple independent PTY-backed terminals per project during an application session.
- Display terminal children under their owning project in the left sidebar.
- Route input, output, resize, status, clipboard, and restart actions to the active terminal by `terminalId`.
- Keep terminal labels synchronized with shell command state: current command while a command runs, shell name when idle.
- Preserve existing behavior where selecting a project gives the user a usable shell without extra setup.

**Non-Goals:**

- Persist live PTY processes, terminal scrollback, or command history across app restart.
- Add explicit terminal close/removal behavior.
- Add split panes, terminal layout management, or cross-project terminal moves.
- Implement perfect command-state hooks for every possible shell. Unknown shells can fall back to the shell name label.

## Decisions

### Terminal identity is first-class

Introduce a terminal descriptor separate from the persisted project record:

```go
type ProjectTerminal struct {
    ID             string `json:"id"`
    ProjectID      string `json:"projectId"`
    ShellName      string `json:"shellName"`
    CurrentCommand string `json:"currentCommand"`
    State          string `json:"state"`
    CreatedAt      string `json:"createdAt"`
    LastSelectedAt string `json:"lastSelectedAt"`
}
```

Backend shell sessions should be keyed by `terminalId`, not `projectId`. Runtime events should include both IDs:

```text
terminal-output { projectId, terminalId, data }
terminal-status { projectId, terminalId, state }
```

The frontend should track both `activeProjectId` and `activeTerminalId`. `activeProjectId` follows the selected terminal's project; project rows remain context and grouping, while terminal rows are the primary shell target.

Alternative considered: keep `projectId` as the shell key and add an index. That makes event routing fragile because terminal order can change when terminals are closed or recreated. Stable `terminalId` is clearer and safer.

### Terminal records are runtime state

Project records continue to be persisted locally. Terminal descriptors are runtime records owned by the shell/session layer. On startup or when selecting a project with no terminal, the app creates a default terminal for that project so the existing "select project and get a shell" workflow remains intact.

Creating a terminal starts a new PTY in the owning project's directory, adds a terminal child under that project, and activates it. Exited terminals remain visible with an exited state and can be restarted using the same `terminalId`.

Alternative considered: persist terminal descriptors in `projects.json`. That would preserve terminal names across restarts, but the PTYs and scrollback would still be gone, making the restored tree misleading. Runtime-only terminals are less surprising for this change.

### Sidebar becomes a project tree

`ProjectSidebar` should render project rows with nested terminal rows. It should emit separate actions for:

- creating a project
- creating a terminal under a project
- selecting a project
- selecting a terminal

The workspace header should continue to show the active project name and path. The terminal surface should render one pane per terminal descriptor, keyed by `terminalId`; only the active terminal pane is visible.

### Frontend terminal manager keys by terminalId

`TerminalSessionManager` should rename its project-scoped methods to terminal-scoped behavior internally:

- `ensure(terminalId, metadata)`
- `activate(terminalId, container)`
- `write(terminalId, data)`
- `fit(terminalId, reportResize)`
- clipboard actions by terminalId

The App layer remains responsible for mapping a terminal to its project when calling backend APIs.

### Command labels come from shell integration OSC events

Use a lightweight shell integration hook for supported shells. The hook emits an app-specific OSC sequence before a command starts and after the shell returns to the prompt. xterm.js exposes `terminal.parser.registerOscHandler`, so the frontend can consume those app-specific messages without confusing them with normal terminal titles.

Proposed event shapes:

```text
OSC 777 ; tui-helper ; command-start ; <encoded command> BEL
OSC 777 ; tui-helper ; command-end BEL
```

When `command-start` is received, the frontend sets `currentCommand` for that terminal. When `command-end` is received, the frontend clears `currentCommand`, causing the display label to fall back to `shellName`.

The displayed label should be derived, not separately edited:

```text
displayName = currentCommand != "" ? currentCommand : shellName
```

Alternative considered: infer commands from frontend keyboard input. That is simpler, but it breaks with history navigation, shell editing, pasted multi-line commands, completion, and interactive TUI programs. Shell integration is more accurate and matches the requested "current executing command" behavior.

## Risks / Trade-offs

- Shell hook support is shell-specific -> Start with `zsh` and `bash`, fall back to shell basename for unsupported shells, and keep the terminal usable without command labels.
- Commands may emit unusual control sequences -> Use an app-specific OSC identifier and sanitize/truncate labels before rendering them in the sidebar.
- Introducing `terminalId` changes Wails method signatures and generated frontend bindings -> Update Go tests, generated bindings, and Vue tests together so routing regressions are caught.
- Runtime-only terminal descriptors disappear on app restart -> Preserve current project persistence while documenting that live terminals are session-scoped.

## Migration Plan

No persisted project config migration is required if terminal descriptors remain runtime-only. Existing `projects.json` files continue to load as before. The implementation should tolerate missing terminal arrays in frontend test fixtures by creating a default terminal when an available project is activated.

Rollback is straightforward at the data level because persistent project records are unchanged, but code rollback must also restore project-scoped Wails shell APIs and events.
