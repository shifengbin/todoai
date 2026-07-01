# embedded-shell-sessions Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.
## Requirements
### Requirement: Start Shell In Project Directory

系统 SHALL 在为可用 TODO project context 创建终端 session 时使用该 context 解析出的工作目录启动 embedded shell。若 TODO project 的 worktree 状态为 ready 且 worktree 目录可用，工作目录 SHALL 为该 worktree 目录。若 TODO project 的 worktree 状态为 cleared 且原项目目录可用，工作目录 SHALL 回退为该 TODO project 保存的原项目目录。若 TODO project 尚未准备 worktree、worktree 准备失败或无法解析可用工作目录，系统 SHALL 拒绝创建终端并 SHALL NOT 启动 shell 进程。

#### Scenario: Shell starts with todo project worktree directory

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 ready
- **AND** 该 TODO project worktree 路径为 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **THEN** 该终端 shell session 使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动
- **AND** 该终端归属于 TODO `fix-login` 和项目 `demo-app` 的 TODO project context

#### Scenario: Shell starts with source project directory for cleared worktree

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 原项目目录可用
- **THEN** 该终端 shell session 使用工作目录 `/home/user/work/demo-app` 启动
- **AND** 该终端仍归属于 TODO `fix-login` 和项目 `demo-app` 的 TODO project context

#### Scenario: Todo project without prepared or cleared worktree cannot start shell

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 没有已准备的 worktree 路径
- **AND** 该 TODO project 未被标记为 worktree 已清除
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程

#### Scenario: Cleared worktree with unavailable source project cannot start shell

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录不可用
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程

### Requirement: Create Additional Project Terminal

系统 SHALL 允许用户仅为 `in-progress` TODO 中可用的 TODO project context 创建额外终端 session。若该 TODO project worktree 为 ready，新的 shell 进程 SHALL 在该 worktree 目录启动；若该 TODO project worktree 为 cleared 且原项目目录可用，新的 shell 进程 SHALL 在原项目目录启动。每个新终端 session SHALL 归属于该 TODO project context，并 SHALL 与该 context 下已有终端相互独立。系统 SHALL 拒绝为 `not-started` TODO project context、未准备且未清除 worktree 的 TODO project context、worktree 准备失败的 TODO project context 或项目路径不可用的 TODO project context 创建终端。

#### Scenario: User creates another terminal for an in-progress todo project

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 有已准备 worktree 路径 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** 用户在 TODO `fix-login` 下项目 `demo-app` 创建新终端
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动新 shell 进程
- **AND** 新终端独立于该 TODO project context 中已有终端
- **AND** 新终端不显示在引用同一项目的其它 TODO 下

#### Scenario: User creates another terminal for a cleared worktree todo project

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 用户在 TODO `fix-login` 下项目 `demo-app` 创建新终端
- **THEN** 系统使用工作目录 `/home/user/work/demo-app` 启动新 shell 进程
- **AND** 新终端仍归属于该 TODO project context
- **AND** 新终端不显示在引用同一项目的其它 TODO 下

#### Scenario: Not-started todo project cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `not-started`
- **AND** 用户或客户端请求在 TODO `fix-login` 下项目 `/home/user/work/demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project with failed worktree cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 `failed`
- **AND** 用户请求为项目 `demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project without prepared or cleared worktree cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 未准备完成
- **AND** 该 TODO project 未被标记为 worktree 已清除
- **AND** 用户请求为项目 `demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

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
- **AND** focus returns to that active terminal

#### Scenario: Paste empty clipboard from context menu

- **WHEN** the user chooses Paste from the terminal context menu and the system clipboard has no text
- **THEN** no terminal input is sent to the active terminal's shell
- **AND** the menu closes
- **AND** focus returns to that active terminal

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

The system SHALL execute the selected terminal launch profile startup parameters inside the newly created shell session for the selected `in-progress` TODO project context, and SHALL submit the command using the platform-correct interactive Enter sequence. The system SHALL NOT execute launch profile commands for `not-started` TODO project contexts because terminal creation is not allowed there.

#### Scenario: Launch profile submits command to new shell

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system submits `codex` followed by Enter to that new shell session
- **AND** on Windows ConPTY-backed shells the command starts without requiring the user to press Enter again
- **AND** the new terminal belongs to TODO `fix-login`

#### Scenario: Launch profile supports startup parameters

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex --model gpt-5`
- **THEN** the system submits `codex --model gpt-5` followed by Enter to the new shell session as a single command
- **AND** on Windows ConPTY-backed shells the command starts without requiring the user to press Enter again

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

### Requirement: Start Background Launch Profile Command

系统 SHALL 在不注册 embedded terminal session 的情况下启动后台终端启动配置命令。后台启动命令 SHALL 使用配置的终端 shell 的一次性命令模式，SHALL 运行在同一 selected context 的可见终端会使用的工作目录中，并 SHALL 等待进程退出以释放资源。后台命令输出和退出状态 SHALL NOT 显示在终端 UI 中。

#### Scenario: Background project command runs in todo project worktree

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 有已准备 worktree 路径 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** 用户从该 TODO project 启动菜单选择启动参数为 `npm run sync` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动后台命令
- **AND** 不向该 TODO project context 添加 embedded terminal session

#### Scenario: Background project command runs in source project directory for cleared worktree

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 用户从该 TODO project 启动菜单选择启动参数为 `npm run sync` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/demo-app` 启动后台命令
- **AND** 不向该 TODO project context 添加 embedded terminal session

#### Scenario: Background task command runs in todo workspace directory

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 有任务工作区目录 `/home/user/work/customer-a/tasks/abc123`
- **AND** 用户从任务级启动菜单选择启动参数为 `npm run prepare` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123` 启动后台命令
- **AND** 不向该 TODO context 添加任务终端 session

#### Scenario: Background command does not affect terminal UI state

- **WHEN** 终端 `terminal-a` 是当前活动终端
- **AND** 用户从终端启动菜单选择后台启动配置
- **THEN** `terminal-a` 仍是当前活动终端
- **AND** 终端列表保持不变
- **AND** 终端历史保持不变
- **AND** 系统不为该后台命令发出终端输出、终端状态或终端 agent 状态事件

#### Scenario: Background command exits without UI change

- **WHEN** 后台启动配置命令正在运行
- **AND** 后台进程退出
- **THEN** 系统释放进程资源
- **AND** 没有终端被标记为 exited
- **AND** 不显示新终端

#### Scenario: Not-started todo project cannot start background command

- **WHEN** TODO `fix-login` 的状态为 `not-started`
- **AND** 用户或客户端请求在 TODO `fix-login` 下项目 `demo-app` 启动后台命令
- **THEN** 系统拒绝后台命令请求
- **AND** 不启动后台进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project with failed worktree cannot start background command

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 `failed`
- **AND** 用户请求为该 TODO project 启动后台命令
- **THEN** 系统拒绝后台命令请求
- **AND** 不启动后台进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Background command start failure is reported without terminal creation

- **WHEN** 用户选择无法启动命令的后台启动配置
- **THEN** 系统通过现有应用错误显示报告启动错误
- **AND** 不添加 embedded terminal session
- **AND** 当前活动终端保持不变

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

### Requirement: Persist Terminal History Across Application Restart
The system SHALL persist terminal records and recent terminal output for active TODO project contexts in the current workspace data directory so that they are restored after that workspace is reopened. Restored terminals SHALL NOT restore shell processes, PTY state, or command execution, and SHALL be reported as non-running terminals. Terminal history SHALL NOT be shared globally across workspaces.

#### Scenario: Workspace reopen restores terminal records and output
- **WHEN** workspace `/work/customer-a` has TODO `fix-login` with status `in-progress`
- **AND** project `frontend-app` is associated with TODO `fix-login`
- **AND** the user creates terminal A under TODO `fix-login` and project `frontend-app`
- **AND** terminal A outputs `npm test`
- **AND** the user closes and reopens workspace `/work/customer-a`
- **THEN** terminal A appears under TODO `fix-login` and project `frontend-app`
- **AND** selecting terminal A shows output containing `npm test`
- **AND** terminal A is not reported as running
- **AND** no shell process is started automatically for terminal A

#### Scenario: Workspace reopen restores active terminal selection
- **WHEN** workspace `/work/customer-a` has TODO `fix-login` and project `frontend-app` with terminal A and terminal B
- **AND** terminal B is the active terminal in that TODO project context
- **AND** the user closes and reopens workspace `/work/customer-a`
- **THEN** terminal B remains the active terminal for TODO `fix-login` and project `frontend-app`
- **AND** terminal B's persisted output is shown when the workspace is restored

#### Scenario: Restored terminal history is capped
- **WHEN** terminal A produces output larger than the configured terminal history limit
- **AND** the user closes and reopens the owning workspace
- **THEN** terminal A is restored with only its most recent output up to the configured limit
- **AND** terminal A remains selectable under its TODO project context

#### Scenario: Missing terminal history file is treated as empty
- **WHEN** persisted TODO and project data exists in the current workspace
- **AND** terminal history storage is missing from that workspace data directory
- **AND** the user opens the workspace
- **THEN** the application loads the TODO workspace without error
- **AND** no shell process is started solely to recreate missing terminal history

#### Scenario: Terminal history is isolated by workspace
- **WHEN** terminal A has persisted output in workspace `/work/customer-a`
- **AND** the user opens workspace `/work/customer-b`
- **THEN** terminal A does not appear in `/work/customer-b`
- **AND** terminal A's output is not restored in `/work/customer-b`

### Requirement: Clean Persisted Terminal History

The system SHALL remove persisted terminal records and output history when their owning terminal, TODO, TODO project context, or project is removed from the active workspace.

#### Scenario: Removed terminal clears persisted history

- **WHEN** terminal A has persisted output history
- **AND** the user removes terminal A
- **AND** the user closes and reopens the application
- **THEN** terminal A does not appear in the TODO project terminal tree
- **AND** terminal A's output history is not restored

#### Scenario: Completed todo clears owned terminal history

- **WHEN** TODO `fix-login` owns terminal A with persisted output history
- **AND** the user completes TODO `fix-login`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored under active TODOs
- **AND** terminal A's output history is not restored

#### Scenario: Removed todo project clears owned terminal history

- **WHEN** TODO `fix-login` has project `frontend-app` with terminal A and persisted output history
- **AND** the user removes project `frontend-app` from TODO `fix-login`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored
- **AND** terminal A's output history is not restored

#### Scenario: Deleted project clears owned terminal history

- **WHEN** project `frontend-app` owns terminal A across one or more TODO contexts
- **AND** terminal A has persisted output history
- **AND** the user deletes project `frontend-app`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored
- **AND** terminal A's output history is not restored

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

### Requirement: Reflect Launch Profile Command Label

The system SHALL update the newly created terminal's command label when it submits a launch profile command, even before a shell-specific command-start event is received. The launch profile command label by itself MUST NOT mark the terminal agent activity phase as `busy`; agent activity SHALL be derived by the unified agent status system from shell lifecycle, command-state, structured Claude/Codex events, and title-change fallback according to source priority. Submitting a non-empty launch profile command MUST NOT render supported application-private command-state payloads in the terminal, and MUST NOT hide unrelated base64-like launch output by heuristic.

#### Scenario: Windows launch profile displays command label immediately

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `codex`
- **AND** the terminal label does not remain `pwsh`, `powershell`, or `cmd`
- **AND** the terminal agent activity phase remains `idle` until a shell lifecycle event, structured agent event, command-state event, or terminal title change updates activity

#### Scenario: Windows Claude launch profile displays command label without forcing busy

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `claude` with startup parameters `claude --dangerously-skip-permissions`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `claude --dangerously-skip-permissions`
- **AND** the terminal activity state remains `idle` until a shell lifecycle event, structured agent event, command-state event, or terminal title change updates activity

#### Scenario: Windows arbitrary launch profile preserves unrelated startup text

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `Calculator` with startup parameters `calc`
- **THEN** the system submits the profile command to the new shell session
- **AND** the terminal displays normal shell or program output
- **AND** the terminal does not display supported application-private command-state payloads created by launch profile submission
- **AND** the terminal preserves unrelated base64-like launch output instead of hiding it by heuristic

#### Scenario: Launch profile with parameters displays submitted command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `Codex GPT-5` with startup parameters `codex --model gpt-5`
- **THEN** the TODO terminal tree displays the new terminal label as `codex --model gpt-5`
- **AND** the displayed label is sanitized using the normal terminal command label rules
- **AND** the submitted command label is not by itself treated as an agent busy signal

#### Scenario: Shell command end clears command label when available

- **WHEN** a terminal has command label `codex`
- **AND** the shell integration emits a command-end event for that terminal
- **THEN** the system clears the terminal command label
- **AND** the terminal falls back to its shell display name while still running
- **AND** the unified agent status for that terminal is reset to `idle` unless a newer structured agent event keeps a higher-priority phase

### Requirement: Emit Command State For Windows PowerShell Sessions

The system SHALL provide command state events for Windows `pwsh` and `powershell` embedded shell sessions using an application-private command-state protocol that is consumed before terminal rendering and history persistence. The system MUST NOT expose the command-state protocol payload as visible terminal output. When a valid command-state event cannot be recovered, the system SHALL safely ignore the event and preserve existing launch profile command-label fallback behavior. Valid command-state events SHALL update the command label and SHALL feed the unified agent status system as shell-command lifecycle signals.

#### Scenario: PowerShell command start updates terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `pwsh.exe` or `powershell.exe`
- **AND** the user runs `npm test` in the embedded terminal
- **THEN** the shell integration emits a command-start event for `npm test`
- **AND** the corresponding terminal command label becomes `npm test`
- **AND** the command-state payload is not displayed in the terminal
- **AND** the command-start signal is available to the unified agent status reducer

#### Scenario: PowerShell command completion clears terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal is running command `npm test`
- **AND** the command completes and the shell returns to its prompt
- **THEN** the shell integration emits a command-end event for that terminal
- **AND** the terminal command label is cleared
- **AND** the command-state payload is not displayed in the terminal
- **AND** the command-end signal is available to reset agent activity when no newer structured agent status applies

#### Scenario: PowerShell command-state payload is hidden during launch profile startup

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the user launches any non-empty launch profile in a PowerShell-backed embedded terminal
- **AND** PowerShell emits an application-private command-state payload for the submitted command
- **THEN** the command-state payload is not displayed in the terminal
- **AND** the command-state payload is not persisted in terminal history
- **AND** the launch profile command label remains visible unless a valid command-state event replaces or clears it

#### Scenario: Invalid PowerShell command-state payload is safely ignored

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal produces an application-private command-start payload with invalid base64 command text
- **THEN** the system does not update the terminal command label from that invalid payload
- **AND** the invalid command-state payload is not displayed in the terminal
- **AND** the invalid command-state payload is not persisted in terminal history
- **AND** the invalid command-state payload does not change the unified agent activity phase

#### Scenario: Cmd fallback remains usable without lifecycle hook

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `cmd.exe`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the launch profile command to the cmd-backed shell session
- **AND** the system keeps the application-provided launch profile command label when no shell lifecycle hook event is available
- **AND** the shell session remains associated with `cmd.exe`
- **AND** agent activity falls back to structured Claude/Codex events or terminal title-change fallback when shell command-state is unavailable

### Requirement: Start Shell In Workspace Directory

系统 SHALL 支持为 workspace 全局终端启动 embedded shell。全局终端 shell 进程 SHALL 使用当前 workspace 根目录作为工作目录，并 SHALL 与 TODO project 终端使用相同的 shell 设置、输入、输出、resize 和 clipboard 能力。

#### Scenario: Workspace global shell starts in workspace root

- **WHEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 用户创建全局终端
- **THEN** 系统启动 embedded shell
- **AND** shell 工作目录为 `/home/user/work/customer-a`
- **AND** shell 使用当前配置的终端 shell

#### Scenario: Workspace global shell uses launch fallback

- **WHEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 保存的终端 shell 不可用
- **AND** 自动检测选择 `/bin/sh` 作为 fallback
- **AND** 用户创建全局终端
- **THEN** 系统使用 `/bin/sh` 启动全局终端 shell
- **AND** shell 工作目录为 `/home/user/work/customer-a`

### Requirement: Isolate Workspace Global Terminal Sessions

系统 SHALL 将 workspace 全局终端与 TODO project 终端隔离。全局终端 SHALL 不属于任何 TODO project context，TODO 删除、TODO project 移除和项目候选删除 SHALL NOT 删除全局终端。关闭 workspace SHALL 关闭运行中的全局终端进程。

#### Scenario: Todo deletion preserves global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** TODO `修复登录问题` 下的终端 B 正在运行
- **AND** 用户删除 TODO `修复登录问题`
- **THEN** 终端 B 被关闭并移除
- **AND** 全局终端 A 继续运行并保持显示

#### Scenario: Project candidate deletion preserves global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** 用户删除全局项目候选 `frontend-app`
- **THEN** 全局终端 A 继续运行并保持显示
- **AND** 全局终端 A 的工作目录仍为当前 workspace 根目录

#### Scenario: Closing workspace closes global terminals

- **WHEN** 全局终端 A 正在运行
- **AND** 用户关闭当前 workspace
- **THEN** 系统关闭全局终端 A 的 shell 进程
- **AND** 运行时终端状态被清空

### Requirement: Create Todo Task Terminal

The system SHALL allow the user to create task-level terminal sessions for an `in-progress` TODO whose task workspace directory exists. Each task-level terminal SHALL start an independent shell process in the TODO task workspace directory. Task-level terminals SHALL belong to the TODO but SHALL NOT belong to any TODO project.

#### Scenario: User creates task terminal for in-progress todo

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** TODO `fix-login` has task workspace directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the user creates a task terminal under TODO `fix-login`
- **THEN** the system starts a new shell process with working directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the terminal records TODO `fix-login` as owner
- **AND** the terminal does not record a TODO project owner

#### Scenario: Not-started todo cannot create task terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user requests a task terminal under TODO `fix-login`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no task terminal is added to TODO `fix-login`

#### Scenario: Task terminal does not change todo project context

- **WHEN** current TODO project context is TODO `fix-login` under project `frontend-app`
- **AND** the user selects a task terminal under TODO `fix-login`
- **THEN** the task terminal becomes the active terminal
- **AND** current TODO project context remains TODO `fix-login` under project `frontend-app`

