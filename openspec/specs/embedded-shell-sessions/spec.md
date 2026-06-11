# embedded-shell-sessions Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.
## Requirements
### Requirement: Start Shell In Project Directory

The system SHALL start an embedded shell session in the owning project's directory when a terminal session is created for an available TODO project context.

#### Scenario: Shell starts with project working directory

- **WHEN** the user creates a terminal for TODO `fix-login` under project with path `/home/user/work/demo-app`
- **THEN** the shell session for that terminal starts with working directory `/home/user/work/demo-app`
- **AND** the terminal is owned by the TODO project context for `fix-login` and `/home/user/work/demo-app`

### Requirement: Create Additional Project Terminal

The system SHALL allow the user to create additional terminal sessions only for an available project within an `in-progress` TODO project context. Each created terminal session SHALL start an independent shell process in the owning project's directory and SHALL belong only to that TODO project context. The system SHALL reject terminal creation for `not-started` TODO project contexts.

#### Scenario: User creates another terminal for an in-progress todo project

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user creates a new terminal under TODO `fix-login` and project `/home/user/work/demo-app`
- **THEN** the system starts a new shell process with working directory `/home/user/work/demo-app`
- **AND** the new terminal is independent from existing terminal sessions in that TODO project context
- **AND** the new terminal is not shown under other TODOs that reference the same project

#### Scenario: Not-started todo project cannot create terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a new terminal under TODO `fix-login` and project `/home/user/work/demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no runtime terminal session is added to that TODO project context

### Requirement: Maintain Multiple Runtime Shell Sessions Per Project

The system SHALL support multiple live shell sessions per TODO project context while the application is running. Shell sessions for the same project in different TODOs SHALL remain independent.

#### Scenario: Same todo project runs commands in separate terminals

- **WHEN** the user creates terminal A and terminal B under TODO `fix-login` and project `demo-app`
- **AND** starts a long-running command in terminal A
- **THEN** terminal B remains available for independent shell input
- **AND** terminal A's command continues running in the background

#### Scenario: Switching terminals keeps previous terminal alive

- **WHEN** the user switches from terminal A to terminal B and then back to terminal A within the same TODO project context
- **THEN** the shell area shows terminal A's existing session instead of creating a new shell

#### Scenario: Same project in another todo has separate sessions

- **WHEN** project `demo-app` is associated with TODO `fix-login` and TODO `upgrade-deps`
- **AND** terminal A is running under TODO `fix-login`
- **THEN** selecting TODO `upgrade-deps` does not show terminal A
- **AND** creating a terminal under TODO `upgrade-deps` starts a separate shell process

### Requirement: Route Terminal Input To Active Session

The system SHALL send user terminal input only to the currently active terminal's shell session.

#### Scenario: User types after switching terminals

- **WHEN** the user switches from terminal A to terminal B and types in the terminal
- **THEN** the input is sent to terminal B's shell session and not terminal A's shell session

### Requirement: Route Terminal Output By Project

The system SHALL associate shell output with the terminal session that produced it so output is displayed in the correct terminal state under the owning TODO project context.

#### Scenario: Background terminal produces output

- **WHEN** terminal A under TODO `fix-login` is running a command in the background while terminal B is active
- **THEN** terminal A's output is retained for terminal A
- **AND** terminal A's output is not shown in terminal B's terminal

#### Scenario: Same project output does not cross todo contexts

- **WHEN** terminal A under TODO `fix-login` and terminal B under TODO `upgrade-deps` both use project `demo-app`
- **AND** terminal A produces output
- **THEN** terminal A's output is retained under TODO `fix-login`
- **AND** terminal A's output is not shown under TODO `upgrade-deps`

### Requirement: Resize Active Shell PTY

The system SHALL resize the active terminal's PTY when the terminal viewport dimensions change, including application window resize and workspace sidebar width changes.

#### Scenario: Terminal viewport changes size

- **WHEN** the user resizes the application window and the terminal rows or columns change
- **THEN** the active terminal's PTY receives the updated terminal size

#### Scenario: Sidebar resize changes terminal viewport

- **WHEN** the user drags the workspace sidebar divider and the active terminal rows or columns change
- **THEN** the active terminal's PTY receives the updated terminal size

### Requirement: Handle Shell Exit

The system SHALL detect when a terminal shell exits and show that terminal as exited without closing the application or other terminal sessions.

#### Scenario: Shell process exits

- **WHEN** the active terminal's shell process exits
- **THEN** the application marks that terminal session as exited
- **AND** the application remains usable

### Requirement: Label Terminal By Command State

The system SHALL display each terminal's shell name when that terminal is idle and SHALL display the currently executing command while that terminal is running a command. When the command finishes, the terminal label SHALL return to the shell name.

#### Scenario: Terminal starts a command

- **WHEN** terminal A is idle with label `zsh`
- **AND** the user starts command `npm run dev`
- **THEN** terminal A's label becomes `npm run dev`

#### Scenario: Terminal command finishes

- **WHEN** terminal A is labeled `npm run dev` because that command is running
- **AND** the command finishes and the shell returns to the prompt
- **THEN** terminal A's label becomes `zsh`

#### Scenario: Shell command state is unavailable

- **WHEN** a terminal's shell does not report command start or command end state
- **THEN** the terminal label remains the shell name

### Requirement: Use Configured Terminal Shell

The system SHALL start newly created embedded shell sessions with the configured terminal shell path when a usable setting exists.

#### Scenario: New terminal uses saved shell setting

- **WHEN** the terminal shell setting is saved as `/usr/bin/zsh`
- **AND** the user creates a new embedded terminal under TODO `fix-login` for project `/home/user/work/demo-app`
- **THEN** the shell process starts with shell path `/usr/bin/zsh`
- **AND** the shell process working directory is `/home/user/work/demo-app`

#### Scenario: Existing terminal keeps original shell after setting changes

- **WHEN** terminal A was created with shell path `/usr/bin/bash`
- **AND** the user changes the terminal shell setting to `/usr/bin/zsh`
- **THEN** terminal A keeps using `/usr/bin/bash`
- **AND** a terminal created after the setting change uses `/usr/bin/zsh`

#### Scenario: New terminal uses fallback when saved shell is unavailable

- **WHEN** the saved terminal shell setting is unavailable
- **AND** automatic detection selects `/bin/sh` as the fallback shell
- **AND** the user creates a new embedded terminal under a TODO project context
- **THEN** the shell process starts with shell path `/bin/sh`

### Requirement: Copy Terminal Selection To Clipboard

The system SHALL allow users to copy selected text from the active embedded terminal to the system clipboard without using plain `Ctrl+C`.

#### Scenario: Copy selected terminal text with shortcut

- **WHEN** the user has selected text in the active terminal and presses `Ctrl+Shift+C`
- **THEN** the selected text is written to the system clipboard

#### Scenario: Preserve shell interrupt shortcut

- **WHEN** the user presses plain `Ctrl+C` in the active terminal
- **THEN** the input is sent to the active shell instead of being handled as a clipboard copy action

### Requirement: Paste Clipboard Text Into Active Shell

The system SHALL allow users to paste system clipboard text into the active terminal's shell.

#### Scenario: Paste clipboard text with shortcut

- **WHEN** the user presses `Ctrl+Shift+V` in the active terminal and the system clipboard contains text
- **THEN** the clipboard text is sent to the active terminal's shell input

#### Scenario: Ignore empty clipboard paste

- **WHEN** the user triggers paste and the system clipboard has no text
- **THEN** no terminal input is sent to the active terminal's shell

### Requirement: Provide Terminal Clipboard Context Menu

The system SHALL provide a context menu in the active terminal area with Copy and Paste actions.

#### Scenario: Open terminal context menu

- **WHEN** the user right-clicks the active terminal area
- **THEN** the system shows a terminal context menu with Copy and Paste actions at the pointer location

#### Scenario: Copy from context menu

- **WHEN** the user chooses Copy from the terminal context menu while text is selected in the active terminal
- **THEN** the selected text is written to the system clipboard and the menu closes

#### Scenario: Paste from context menu

- **WHEN** the user chooses Paste from the terminal context menu and the system clipboard contains text
- **THEN** the clipboard text is sent to the active terminal's shell input and the menu closes

### Requirement: Remove Runtime Terminal Session

The system SHALL allow the user to remove a runtime terminal session from a TODO project context. If the terminal session has a running PTY process, the system SHALL close that process before removing the terminal session from runtime state.

#### Scenario: User removes a running terminal

- **WHEN** the user deletes a terminal session with a running shell process under TODO `fix-login`
- **THEN** the system closes that shell process
- **AND** the terminal no longer appears under its TODO project context

#### Scenario: User removes an exited terminal

- **WHEN** the user deletes a terminal session whose shell process has exited
- **THEN** the terminal no longer appears under its TODO project context
- **AND** no new shell process is started automatically

#### Scenario: Active terminal is removed

- **WHEN** the active terminal is removed from a TODO project context that still has other terminals
- **THEN** the system selects that TODO project context's most recently selected remaining terminal as the active terminal

#### Scenario: Last terminal is removed

- **WHEN** the last terminal under the active TODO project context is removed
- **THEN** the active terminal is empty
- **AND** no replacement terminal is created automatically

#### Scenario: Removed terminal is not found

- **WHEN** the user requests to delete a terminal that is not in runtime terminal state
- **THEN** the system reports an error and leaves runtime terminal state unchanged

### Requirement: Remove Project Terminal Sessions

The system SHALL close and remove all runtime terminal sessions owned by a project when that project is removed from the application, across every TODO project context that references the project.

#### Scenario: Project with running terminals is removed

- **WHEN** the user confirms deletion of a project that owns running terminal sessions across one or more TODOs
- **THEN** the system closes every running shell process owned by that project
- **AND** removes those terminal sessions from runtime state

#### Scenario: Project terminal cleanup preserves other projects

- **WHEN** the user deletes project A while project B has terminal sessions
- **THEN** project A's terminal sessions are removed from every TODO context
- **AND** project B's terminal sessions remain available

#### Scenario: Project cleanup preserves archived snapshots

- **WHEN** project A is deleted
- **AND** archived TODOs contain snapshots for project A
- **THEN** archived TODO snapshots remain readable
- **AND** no shell process is started for archived TODOs

### Requirement: Run Terminal Launch Profile Command

The system SHALL execute the selected terminal launch profile startup parameters inside the newly created shell session for the selected `in-progress` TODO project context. The system SHALL NOT execute launch profile commands for `not-started` TODO project contexts because terminal creation is not allowed there.

#### Scenario: Launch profile submits command to new shell

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system submits `codex` followed by Enter to that new shell session
- **AND** the new terminal belongs to TODO `fix-login`

#### Scenario: Launch profile supports startup parameters

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex --model gpt-5`
- **THEN** the system submits `codex --model gpt-5` followed by Enter to the new shell session as a single command

#### Scenario: Plain terminal launch does not submit command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses the built-in `Terminal` launch option under a TODO project context
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system does not submit any automatic command to that shell session

#### Scenario: Not-started todo launch profile is not executed

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** the system does not submit `codex` to any shell session

### Requirement: Keep Launch Profile Commands In Configured Shell

The system SHALL run launch profile startup parameters inside the configured terminal shell instead of replacing the shell process with the startup command.

#### Scenario: Launch profile command exits

- **WHEN** a terminal launch profile command exits after running in a new terminal
- **THEN** the terminal remains associated with its configured shell session unless the shell itself exits

### Requirement: Isolate Terminal Sessions By Todo Project Context

The system SHALL isolate runtime terminal sessions by TODO project context. If the same project is associated with multiple TODOs, each TODO project context SHALL have its own terminal collection and active terminal selection.

#### Scenario: Same project has isolated terminals across todos

- **WHEN** project `frontend-app` is associated with TODO `fix-login`
- **AND** project `frontend-app` is associated with TODO `upgrade-deps`
- **AND** the user creates terminal A under `fix-login` and `frontend-app`
- **AND** the user creates terminal B under `upgrade-deps` and `frontend-app`
- **THEN** terminal A appears only under TODO `fix-login`
- **AND** terminal B appears only under TODO `upgrade-deps`
- **AND** selecting terminal A does not change the active terminal for TODO `upgrade-deps`

### Requirement: Remove Todo Terminal Sessions

The system SHALL close and remove all runtime terminal sessions owned by a TODO when that TODO is completed or deleted.

#### Scenario: Completed todo closes owned terminals

- **WHEN** TODO `fix-login` owns running terminal sessions
- **AND** the user confirms completing TODO `fix-login`
- **THEN** the system closes every running shell process owned by TODO `fix-login`
- **AND** removes those terminal sessions from runtime state

#### Scenario: Deleted todo cleanup preserves other todos

- **WHEN** TODO `fix-login` and TODO `upgrade-deps` both have terminal sessions for project `frontend-app`
- **AND** the user confirms deleting TODO `fix-login`
- **THEN** terminal sessions under TODO `fix-login` are closed and removed
- **AND** terminal sessions under TODO `upgrade-deps` remain available

### Requirement: Remove Todo Project Terminal Sessions

The system SHALL close and remove all runtime terminal sessions owned by a single TODO project context when that TODO-project association is removed. Removing one TODO project context SHALL NOT close terminal sessions for the same project under other TODOs.

#### Scenario: Removed todo project closes owned terminals

- **WHEN** TODO `fix-login` has project `frontend-app` with running terminal sessions
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** the system closes every running shell process owned by that TODO project context
- **AND** removes those terminal sessions from runtime state
- **AND** project `frontend-app` no longer appears under TODO `fix-login`

#### Scenario: Todo project cleanup preserves same project in other todos

- **WHEN** project `frontend-app` is associated with TODO `fix-login`
- **AND** project `frontend-app` is associated with TODO `upgrade-deps`
- **AND** both TODO project contexts have terminal sessions
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** terminal sessions under TODO `fix-login` and project `frontend-app` are closed and removed
- **AND** terminal sessions under TODO `upgrade-deps` and project `frontend-app` remain available

#### Scenario: Active todo project is removed

- **WHEN** the active terminal belongs to TODO `fix-login` and project `frontend-app`
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** the active terminal is cleared or moved to a remaining terminal in a valid TODO project context
- **AND** the removed terminal no longer receives input or output

### Requirement: Resolve Platform Terminal Shell For Session Startup
系统 SHALL 在创建新的嵌入式 shell session 前使用平台感知的终端 shell 路径解析，并 SHALL 避免在 Windows 上选择 Unix-only fallback 路径。

#### Scenario: Windows new terminal uses detected shell fallback
- **WHEN** 应用运行在 Windows 上
- **AND** 已保存的终端 shell 设置不可用
- **AND** 自动探测选择 `cmd.exe` 作为 fallback shell
- **AND** 用户为可用项目创建新的嵌入式终端
- **THEN** 新终端的 shell path 解析为 `cmd.exe`
- **AND** shell path 不解析为 `/bin/sh`、`/bin/bash` 或其他 Unix-only fallback

#### Scenario: Windows unsupported PTY startup surfaces startup error
- **WHEN** 应用运行在 Windows 上
- **AND** shell path 已解析为可用的 Windows shell
- **AND** 当前 PTY backend 不支持 Windows session startup
- **AND** 用户创建新的嵌入式终端
- **THEN** 系统报告 shell startup error
- **AND** 系统不通过改用 Unix-only shell path 隐藏该错误

#### Scenario: Non-Windows terminal startup remains unchanged
- **WHEN** 应用运行在非 Windows 系统上
- **AND** 已保存的终端 shell 设置可用
- **AND** 用户创建新的嵌入式终端
- **THEN** 新终端继续使用已保存的 shell path 启动
- **AND** shell process working directory 仍是所属项目路径

### Requirement: Handle Unsupported Windows Embedded Shell Backend

The system SHALL detect when the embedded shell backend is unsupported on Windows and SHALL present a stable unavailable state without repeatedly attempting to start a shell process.

#### Scenario: Windows embedded shell backend is unsupported

- **WHEN** a Windows user creates or restarts an embedded terminal
- **AND** the configured PTY backend does not support Windows shell sessions
- **THEN** the system marks the terminal shell as exited or unavailable
- **AND** the application remains usable
- **AND** the terminal area displays a clear unsupported-platform message

#### Scenario: Unsupported backend does not retry automatically

- **WHEN** a Windows terminal start fails because the embedded shell backend is unsupported
- **THEN** the system does not automatically start another shell process for the same terminal
- **AND** switching focus back to the application does not retry the failed terminal start

#### Scenario: Unsupported backend does not show system console windows

- **WHEN** a Windows terminal start fails because the embedded shell backend is unsupported
- **THEN** the failure is reported inside the application
- **AND** no system console window is displayed for the failed embedded terminal start

### Requirement: Start Windows Embedded Shell With ConPTY

The system SHALL start embedded shell sessions through a Windows ConPTY backend when the application runs on a Windows version that supports ConPTY.

#### Scenario: Windows ConPTY shell starts in project directory

- **WHEN** the application runs on Windows 10 1809 or later
- **AND** the configured terminal shell path resolves to an available Windows shell such as `pwsh.exe`, `powershell.exe`, or `cmd.exe`
- **AND** the user creates an embedded terminal for an available TODO project context
- **THEN** the system starts the shell through the Windows ConPTY backend
- **AND** the shell process working directory is the owning project's path
- **AND** the terminal state becomes `running`
- **AND** shell output is emitted to the owning terminal session

#### Scenario: Windows ConPTY terminal receives input

- **WHEN** a Windows ConPTY-backed terminal is running
- **AND** the user types in the active embedded terminal
- **THEN** the input is written to that ConPTY shell session
- **AND** the input is not written to other terminal sessions

#### Scenario: Windows ConPTY terminal resizes

- **WHEN** a Windows ConPTY-backed terminal is running
- **AND** the active terminal viewport rows or columns change
- **THEN** the Windows ConPTY backend receives the updated terminal size

#### Scenario: Windows ConPTY terminal closes on removal

- **WHEN** the user removes a running Windows ConPTY-backed terminal
- **THEN** the system closes the ConPTY process
- **AND** removes the terminal session from runtime state
- **AND** the removed terminal no longer receives input or output

### Requirement: Preserve Unsupported State For Windows Without ConPTY

The system SHALL keep the existing unsupported embedded terminal behavior when Windows ConPTY is unavailable.

#### Scenario: Windows version does not support ConPTY

- **WHEN** the application runs on a Windows version where ConPTY is unavailable
- **AND** the user creates or restarts an embedded terminal
- **THEN** the system marks the terminal shell as `unsupported`
- **AND** the application remains usable
- **AND** the terminal area displays the unsupported-platform message

#### Scenario: Windows ConPTY initialization reports unsupported

- **WHEN** the application runs on Windows
- **AND** the configured terminal shell path resolves to an available Windows shell
- **AND** ConPTY initialization fails because the backend is unsupported in the current environment
- **THEN** the system marks the terminal shell as `unsupported`
- **AND** the system does not automatically start another shell process for the same terminal

#### Scenario: Windows shell configuration errors remain startup errors

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured shell path is invalid or cannot be started
- **AND** the failure is not an unsupported ConPTY backend failure
- **THEN** the system reports a shell startup error
- **AND** the system does not mark the failure as unsupported
