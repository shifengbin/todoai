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

The system SHALL allow the user to select an opened project from the project library. Selecting a project from the project library SHALL update the selected project for project management and status display, but SHALL NOT create, select, or reveal a terminal session. Terminal activation SHALL occur only through a TODO project context.

#### Scenario: User selects a project from the project library

- **WHEN** the user clicks a project in the project library
- **THEN** that project becomes the selected project for project management
- **AND** no shell session is created
- **AND** no terminal becomes active from that project selection

#### Scenario: User selects a project through a TODO context

- **WHEN** the user clicks project `demo-app` under TODO `fix-login`
- **THEN** project `demo-app` becomes the active project for the shell area
- **AND** the shell area is associated only with terminals under that TODO project context

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

The system SHALL allow the user to remove an opened project from the application without deleting the project's directory or files from disk. Removing a project SHALL remove that project from active TODO project associations and SHALL close runtime terminal sessions for that project across all TODO contexts. Archived TODO project snapshots SHALL remain unchanged.

#### Scenario: User confirms project removal

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app`
- **AND** confirms the deletion
- **THEN** the project list no longer contains `/home/user/work/demo-app`
- **AND** the persisted opened project list no longer contains `/home/user/work/demo-app`
- **AND** active TODOs no longer contain associations to `/home/user/work/demo-app`
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal

- **WHEN** the user requests to delete an opened project
- **AND** cancels the confirmation
- **THEN** the project list remains unchanged
- **AND** TODO project associations remain unchanged

#### Scenario: Active project is removed

- **WHEN** the active project is removed
- **THEN** the system selects the remaining opened project with the most recent selection time as the selected project for project management
- **AND** if no opened projects remain, the selected project is empty
- **AND** any active terminal owned by the removed project is cleared

#### Scenario: Removed project is not found

- **WHEN** the user requests to delete a project that is not in the opened project list
- **THEN** the system reports an error and leaves the opened project list unchanged
- **AND** TODO project associations remain unchanged

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

### Requirement: Display Project Library Tab

系统 SHALL 在左侧工作区的 `项目` tab 中展示全局项目库。项目库 SHALL 用于导入、查看和删除项目，但 SHALL 不直接提供终端创建、终端选择或终端树操作。

#### Scenario: Project tab shows imported projects

- **WHEN** 用户打开 `项目` tab
- **AND** 项目库包含 `frontend-app` 和 `api-service`
- **THEN** `项目` tab 显示 `frontend-app`
- **AND** `项目` tab 显示 `api-service`
- **AND** 项目行不显示终端子行

#### Scenario: Project tab does not expose terminal actions

- **WHEN** 用户打开 `项目` tab
- **AND** 项目 `frontend-app` 可用
- **THEN** 项目行不显示新增终端启动菜单
- **AND** 点击项目行不会创建 shell 终端

### Requirement: Import Projects From Parent Directory

系统 SHALL 允许用户选择一个父目录，并将该父目录下第一层可访问子目录批量导入为项目。系统 SHALL 跳过普通文件、不可访问目录和已存在的项目路径。

#### Scenario: User imports child directories from a parent directory

- **WHEN** 用户选择父目录 `/home/user/work`
- **AND** 该目录包含子目录 `/home/user/work/frontend-app`
- **AND** 该目录包含子目录 `/home/user/work/api-service`
- **THEN** 项目库包含项目 `frontend-app`，路径为 `/home/user/work/frontend-app`
- **AND** 项目库包含项目 `api-service`，路径为 `/home/user/work/api-service`

#### Scenario: Import skips duplicate project paths

- **WHEN** 项目库已包含路径 `/home/user/work/frontend-app`
- **AND** 用户从父目录 `/home/user/work` 批量导入
- **THEN** 系统不会创建第二个 `/home/user/work/frontend-app` 项目
- **AND** 导入结果显示该路径被跳过

#### Scenario: Import ignores non-directory children

- **WHEN** 用户选择父目录 `/home/user/work`
- **AND** 该目录包含文件 `/home/user/work/readme.md`
- **THEN** 系统不会为 `readme.md` 创建项目

#### Scenario: User cancels parent directory import

- **WHEN** 用户打开父目录导入选择器并取消
- **THEN** 项目库保持不变

### Requirement: Report Parent Directory Import Summary

系统 SHALL 在父目录批量导入完成后向用户展示导入摘要。摘要 SHALL 至少包含新增项目数量和跳过项目数量。

#### Scenario: Import summary is shown

- **WHEN** 用户从父目录批量导入 2 个新项目
- **AND** 系统跳过 1 个已存在路径
- **THEN** 系统显示 2 个项目已导入
- **AND** 系统显示 1 个项目已跳过

