# project-workspace Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.
## Requirements
### Requirement: Create Project From Directory

The system SHALL allow the user to create a project by selecting a local directory through a native directory picker. The created project's default display name SHALL be the basename of the selected directory.

#### Scenario: User creates a project from a directory

- **WHEN** the user clicks the new project action and selects `/home/user/work/demo-app`
- **THEN** the project list contains a project named `demo-app` with path `/home/user/work/demo-app`

#### Scenario: User cancels directory selection

- **WHEN** the user opens the directory picker and cancels it
- **THEN** the project list remains unchanged

### Requirement: Persist Opened Projects

The system SHALL persist the opened project list locally and reload it when the application starts.

#### Scenario: Project list is restored after restart

- **WHEN** the user creates projects and then closes and reopens the application
- **THEN** the previously opened projects appear in the left-side project list

### Requirement: Select Active Project

The system SHALL allow the user to select an opened project from the left-side project tree and SHALL expose the selected project as the active project for the shell area. If the selected project has an existing terminal session, the system SHALL make that project's most recently active terminal active. If the selected project has no terminal session and the project path is available, the system SHALL create and select a default terminal for that project.

#### Scenario: User selects a project

- **WHEN** the user clicks a project in the left-side project tree
- **THEN** that project becomes active
- **AND** the shell area is associated with an active terminal in that project's directory

### Requirement: Display Project Terminal Tree

The system SHALL display opened projects as top-level rows in the left sidebar and SHALL display each project's terminal sessions as child rows under the owning project.

#### Scenario: Project has multiple terminals

- **WHEN** a project has terminal sessions named `zsh`, `npm run dev`, and `go test ./...`
- **THEN** the left sidebar shows the project row with those terminal rows nested beneath it

### Requirement: Select Active Terminal From Project Tree

The system SHALL allow the user to select a terminal row under a project and SHALL expose that terminal as the active terminal for the shell area.

#### Scenario: User selects a terminal under a project

- **WHEN** the user clicks terminal `go test ./...` under project `demo-app`
- **THEN** project `demo-app` becomes the active project
- **AND** terminal `go test ./...` becomes the active terminal shown in the shell area

### Requirement: Collapse Project Terminal Branches

The system SHALL allow the user to expand and collapse each project's terminal child rows independently in the left-side project terminal tree.

#### Scenario: User collapses a project branch

- **WHEN** a project branch is expanded and has terminal child rows
- **AND** the user activates that project's collapse control
- **THEN** the terminal child rows for that project are hidden
- **AND** the project row remains visible

#### Scenario: User expands a project branch

- **WHEN** a project branch is collapsed and has terminal child rows
- **AND** the user activates that project's expand control
- **THEN** the terminal child rows for that project are shown beneath the project row

#### Scenario: Collapsing one project does not affect another project

- **WHEN** project A and project B both have terminal child rows
- **AND** the user collapses project A
- **THEN** project A's terminal child rows are hidden
- **AND** project B's expanded or collapsed state is unchanged

### Requirement: Reveal Active Project Terminal Branch

The system SHALL expand the branch for the project that becomes active through project selection, terminal selection, or terminal creation.

#### Scenario: User selects a collapsed project

- **WHEN** a project branch is collapsed
- **AND** the user selects that project
- **THEN** that project's branch is expanded

#### Scenario: User selects a terminal under a project

- **WHEN** a terminal becomes active under a project
- **THEN** the owning project's branch is expanded

#### Scenario: User creates a terminal under a project

- **WHEN** the user creates a terminal under an available project
- **THEN** the owning project's branch is expanded
- **AND** the new terminal row is visible under that project

### Requirement: Show Project Terminal Hierarchy

The system SHALL visually distinguish project parent rows from terminal child rows and SHALL communicate terminal ownership through tree indentation or branch guides.

#### Scenario: Project has terminal children

- **WHEN** a project has visible terminal child rows
- **THEN** the terminal rows appear visually nested under the project row
- **AND** the sidebar communicates that the terminal rows belong to that project

#### Scenario: Project branch is collapsed

- **WHEN** a project with terminal child rows is collapsed
- **THEN** the project row indicates that hidden terminal children can be expanded

### Requirement: Handle Duplicate Project Paths

The system SHALL avoid creating duplicate project entries for the same absolute path.

#### Scenario: User selects an already opened directory

- **WHEN** the user creates a project from a directory that is already in the project list
- **THEN** the existing project entry is selected instead of adding a duplicate entry

### Requirement: Handle Missing Project Paths

The system SHALL detect when a persisted project path no longer exists or is inaccessible and SHALL prevent shell startup for that project until the path is valid again.

#### Scenario: Persisted project path is missing

- **WHEN** the application starts and a persisted project path no longer exists
- **THEN** the project remains visible as unavailable and selecting it does not launch a shell

### Requirement: Remove Opened Project

The system SHALL allow the user to remove an opened project from the application without deleting the project's directory or files from disk.

#### Scenario: User confirms project removal

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app`
- **AND** confirms the deletion
- **THEN** the project list no longer contains `/home/user/work/demo-app`
- **AND** the persisted opened project list no longer contains `/home/user/work/demo-app`
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal

- **WHEN** the user requests to delete an opened project
- **AND** cancels the confirmation
- **THEN** the project list remains unchanged

#### Scenario: Active project is removed

- **WHEN** the active project is removed
- **THEN** the system selects the remaining opened project with the most recent selection time as the active project
- **AND** if no opened projects remain, the active project is empty

#### Scenario: Removed project is not found

- **WHEN** the user requests to delete a project that is not in the opened project list
- **THEN** the system reports an error and leaves the opened project list unchanged

### Requirement: Display Active Project Git Status Bar

系统 SHALL 在主工作区底部显示固定高度状态栏，用于展示当前激活项目的 Git 仓库信息。状态栏 SHALL 使用可区分的圆角信息块展示 Git 信息，并 SHALL 至少包含当前分支名称和当前项目的已改动文件数量。

#### Scenario: Active Git project shows branch and change count

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 当前分支是 `main`
- **AND** 仓库有 3 个已改动文件
- **THEN** 状态栏显示分支 `main`
- **AND** 状态栏显示 3 个已改动文件
- **AND** 分支和改动数量显示在独立圆角信息块中

#### Scenario: Active Git project shows detailed change counts

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 仓库有暂存、未暂存或未跟踪文件
- **THEN** 状态栏分别显示 staged、unstaged 和 untracked 数量
- **AND** 每类数量显示在可区分的圆角信息块中

#### Scenario: Active Git project shows ahead and behind counts

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 当前分支相对上游存在 ahead 或 behind 数量
- **THEN** 状态栏显示 ahead 或 behind 数量
- **AND** ahead 或 behind 数量显示在独立圆角信息块中

#### Scenario: No active project keeps status bar stable

- **WHEN** 没有项目被选中
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏显示没有项目的空状态

### Requirement: Refresh Active Project Git Status

系统 SHALL 在当前项目可能变化或仓库状态可能变化时刷新状态栏中的 Git 信息。刷新 SHALL 不改变终端区域的布局高度。

#### Scenario: Project selection refreshes git status

- **WHEN** 用户从项目树选择另一个可用项目
- **THEN** 系统查询新激活项目的 Git 状态
- **AND** 状态栏显示新激活项目的 Git 信息

#### Scenario: Terminal command completion refreshes git status

- **WHEN** 激活项目的终端命令结束
- **THEN** 系统刷新当前激活项目的 Git 状态
- **AND** 状态栏反映命令结束后的改动文件数量

#### Scenario: Window focus refreshes git status

- **WHEN** 应用窗口重新获得焦点
- **AND** 当前激活项目可用
- **THEN** 系统刷新当前激活项目的 Git 状态

### Requirement: Handle Git Status Unavailable

系统 SHALL 在当前项目不是 Git 仓库、项目路径不可用或 Git 状态查询失败时保持状态栏可用，并显示不会阻断项目或终端操作的状态信息。

#### Scenario: Active project is not a git repository

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径不是 Git 仓库
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 终端区域保持可用

#### Scenario: Active project path is unavailable

- **WHEN** 当前激活项目路径不可访问
- **THEN** 状态栏显示项目路径不可用的状态
- **AND** 系统不执行 Git 状态查询

#### Scenario: Git executable or command fails

- **WHEN** 当前激活项目可用
- **AND** Git 状态查询失败
- **THEN** 状态栏显示 Git 状态不可用
- **AND** 项目选择和终端会话功能保持可用

### Requirement: Highlight Active Terminal Branch Guide

The system SHALL visually highlight the visible project-terminal branch guide for the project that owns the active terminal.

#### Scenario: Active terminal branch guide is highlighted

- **WHEN** a project has visible terminal child rows
- **AND** one of those terminal rows is the active terminal
- **THEN** the visible vertical branch guide for that project's terminal list uses the same active color as the active terminal row's horizontal branch guide

#### Scenario: Inactive terminal branch guides remain neutral

- **WHEN** a project has visible terminal child rows
- **AND** none of those terminal rows is the active terminal
- **THEN** the visible vertical branch guide for that project's terminal list uses the default neutral branch guide color

### Requirement: Display Interactive Terminal Activity In Project Tree

系统 SHALL 在左侧项目终端树中展示交互式终端程序的运行时活动状态。该状态 SHALL 基于终端标题变化推导，并 SHALL 不替代现有 shell 命令标签。

#### Scenario: Interactive terminal records an idle launch title

- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 刚启动并收到静态运行时标题 `codex - alpha`
- **THEN** terminal A 的项目树终端行不显示执行中的动态指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Interactive terminal is busy after launch

- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 已记录空闲运行时标题 `codex - alpha`
- **AND** terminal A 后续收到不同于空闲标题的执行中运行时标题
- **THEN** terminal A 的项目树终端行显示执行中的动态指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Interactive terminal needs user input

- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 收到表示需要注意的运行时标题，例如包含 `!`
- **THEN** terminal A 的项目树终端行显示需要用户输入的注意指示
- **AND** terminal A 不显示执行中的动态指示

#### Scenario: Interactive terminal returns to idle title

- **WHEN** terminal A 的项目树终端行正在显示执行中或需要输入状态
- **AND** terminal A 的运行时标题恢复为当前命令标签 `codex` 或已记录的空闲标题
- **THEN** terminal A 的项目树终端行停止活动指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Shell command state still controls command label

- **WHEN** terminal A 的 shell command-start 事件报告当前命令为 `codex`
- **AND** terminal A 后续收到运行时标题变化
- **THEN** terminal A 的主标签由 shell 命令状态保持为 `codex`
- **AND** 运行时标题只影响活动指示状态

#### Scenario: Shell command exits

- **WHEN** terminal A 正在显示交互式活动状态
- **AND** terminal A 的 shell command-end 事件到达
- **THEN** terminal A 清除交互式活动指示
- **AND** terminal A 的主标签恢复为 shell 名称

### Requirement: Show Project Terminal Launch Menu

The system SHALL show a terminal launch menu when the user activates the add-terminal control for an available project.

#### Scenario: Launch menu contains terminal and configured profiles

- **WHEN** settings contains launch profiles named `codex` and `claude`
- **AND** the user activates the add-terminal control for an available project
- **THEN** the launch menu shows `Terminal` as the first option
- **AND** the launch menu shows `codex` and `claude` after `Terminal` in the configured order

#### Scenario: Unavailable project has no launch menu

- **WHEN** a project path is unavailable
- **THEN** the project row does not expose an add-terminal launch menu action

### Requirement: Create Terminal From Launch Menu

The system SHALL create a new terminal for the selected project using the launch option chosen from the project terminal launch menu.

#### Scenario: User chooses terminal launch option

- **WHEN** the user opens the launch menu for project `demo-app`
- **AND** chooses `Terminal`
- **THEN** the system creates a new terminal under project `demo-app`
- **AND** the new terminal starts as a normal shell session without an automatic startup command

#### Scenario: User chooses configured launch profile

- **WHEN** the user opens the launch menu for project `demo-app`
- **AND** chooses the `codex` launch profile
- **THEN** the system creates a new terminal under project `demo-app`
- **AND** the new terminal is selected as the active terminal
- **AND** the owning project branch is expanded so the new terminal row is visible

### Requirement: Initialize Active Project Git Repository From Status Bar

系统 SHALL 在当前激活项目可用且不是 Git 仓库时，在底部状态栏提供初始化 Git 仓库的操作。该操作 SHALL 直接在当前项目路径执行 Git 初始化，并在成功后刷新状态栏中的 Git 信息。

#### Scenario: Non-Git project shows initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径不是 Git 仓库
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 状态栏显示 `Initialize Git Repository` 操作

#### Scenario: User initializes Git repository from status bar

- **WHEN** 当前激活项目可用
- **AND** 该项目路径不是 Git 仓库
- **AND** 用户点击 `Initialize Git Repository`
- **THEN** 系统在当前项目路径执行 `git init`
- **AND** 状态栏刷新当前项目的 Git 信息
- **AND** 状态栏不再显示非 Git 仓库状态

#### Scenario: Git initialization is in progress

- **WHEN** 用户点击 `Initialize Git Repository`
- **AND** Git 初始化请求尚未完成
- **THEN** 初始化操作不可重复触发
- **AND** 状态栏显示初始化进行中的状态

#### Scenario: Git initialization fails

- **WHEN** 当前激活项目可用
- **AND** 用户点击 `Initialize Git Repository`
- **AND** Git 初始化失败
- **THEN** 状态栏仍显示该项目不是 Git 仓库
- **AND** 系统显示不会阻断终端操作的错误信息

