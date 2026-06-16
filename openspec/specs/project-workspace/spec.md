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
The system SHALL persist the opened project list inside the current workspace data directory and SHALL reload it when that workspace is opened again. Opened project lists SHALL NOT be shared globally across workspaces.

#### Scenario: Project list is restored after reopening workspace
- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates projects in that workspace
- **AND** the user closes and reopens workspace `/home/user/work/customer-a`
- **THEN** the previously opened projects appear in the left-side project list

#### Scenario: Project list is isolated by workspace
- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates project `/home/user/repos/frontend-a`
- **AND** the user opens workspace `/home/user/work/customer-b`
- **THEN** project `/home/user/repos/frontend-a` does not appear in the project list for `/home/user/work/customer-b`

#### Scenario: No workspace has no opened project list
- **WHEN** no workspace is open
- **THEN** the project list is empty
- **AND** creating or importing opened projects is unavailable until a workspace is opened

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

The system SHALL allow the user to remove an opened project from the application without deleting the project's directory or files from disk. Removing a project SHALL require confirmation in a contextual popover anchored to the project row delete button. Removing a project SHALL remove that project from active TODO project associations and SHALL close runtime terminal sessions for that project across all TODO contexts. Archived TODO project snapshots SHALL remain unchanged.

#### Scenario: User opens project removal confirmation

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app` from the project library row
- **THEN** the system shows a confirmation popover next to that project row delete button
- **AND** the system does not use the browser native confirmation dialog
- **AND** the project list remains unchanged until the user confirms the popover

#### Scenario: User confirms project removal

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app`
- **AND** the system shows the project removal confirmation popover
- **AND** the user confirms the deletion
- **THEN** the project list no longer contains `/home/user/work/demo-app`
- **AND** the persisted opened project list no longer contains `/home/user/work/demo-app`
- **AND** active TODOs no longer contain associations to `/home/user/work/demo-app`
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal

- **WHEN** the user requests to delete an opened project
- **AND** the system shows the project removal confirmation popover
- **AND** the user cancels the confirmation
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

系统 SHALL 在当前项目可能变化或仓库状态可能变化时刷新状态栏中的 Git 信息。刷新 SHALL 不改变终端区域的布局高度。系统 SHALL 避免在项目导入完成后立即触发 Git 状态查询，并 SHALL 将导入后的 Git 状态查询延迟到用户展开 TODO、选择 TODO 项目、显式选择项目或其他明确刷新时机。系统在 Windows 上执行后台 Git 状态刷新和 Git 初始化命令时 SHALL 不显示临时控制台窗口，并且 SHALL 防止窗口 focus 事件在短时间内重复触发同一项目的 Git 状态刷新。

#### Scenario: Project selection refreshes git status

- **WHEN** 用户从项目树选择另一个可用项目
- **THEN** 系统查询新激活项目的 Git 状态
- **AND** 状态栏显示新激活项目的 Git 信息

#### Scenario: Importing projects defers git status refresh

- **WHEN** 用户从父目录导入一个或多个项目
- **THEN** 系统更新项目列表和导入摘要
- **AND** 系统不立即查询任何导入项目的 Git 状态
- **AND** 状态栏保持导入前的 Git 状态或空状态

#### Scenario: Expanding todo refreshes active todo project git status

- **WHEN** 用户展开一个包含当前激活 TODO project 的 TODO
- **AND** 当前激活项目路径可用
- **THEN** 系统查询当前激活项目的 Git 状态
- **AND** 状态栏显示当前激活项目的 Git 信息

#### Scenario: Selecting todo project refreshes git status

- **WHEN** 用户选择 TODO 下的一个项目
- **THEN** 系统查询该项目的 Git 状态
- **AND** 状态栏显示该项目的 Git 信息

#### Scenario: Terminal command completion refreshes git status

- **WHEN** 激活项目的终端命令结束
- **THEN** 系统刷新当前激活项目的 Git 状态
- **AND** 状态栏反映命令结束后的改动文件数量

#### Scenario: Window focus refreshes git status

- **WHEN** 应用窗口重新获得焦点
- **AND** 当前激活项目可用
- **THEN** 系统刷新当前激活项目的 Git 状态

#### Scenario: Windows Git refresh does not flash console windows

- **WHEN** Windows 用户打开应用或切换到一个可用 Git 项目
- **THEN** 系统在后台查询该项目的 Git 状态
- **AND** 查询过程不显示系统控制台窗口
- **AND** 应用窗口保持可用

#### Scenario: Focus jitter does not start repeated Git refreshes

- **WHEN** 应用窗口在短时间内重复获得焦点
- **AND** 当前激活项目没有变化
- **THEN** 系统最多启动一次当前项目的 focus Git 状态刷新
- **AND** 后续重复 focus 事件不会启动新的后台 Git 命令，直到去重窗口结束或已有请求完成

#### Scenario: Windows Git initialization does not flash console windows

- **WHEN** Windows 用户从状态栏初始化当前项目的 Git 仓库
- **THEN** 系统在后台执行 Git 初始化
- **AND** 初始化过程不显示系统控制台窗口
- **AND** 初始化完成后系统刷新当前项目的 Git 状态

### Requirement: Handle Git Status Unavailable

系统 SHALL 在当前项目不是 Git 仓库、项目路径不可用、Git 命令未安装或 Git 状态查询失败时保持状态栏可用，并显示不会阻断项目或终端操作的状态信息。系统在执行 Git 状态查询前 SHALL 校验 `git` 命令是否存在；若 `git` 命令不存在或不可执行，系统 SHALL 不执行 Git 状态查询命令。

#### Scenario: Active project is not a git repository

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径不是 Git 仓库
- **AND** 系统可以执行 `git` 命令
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 终端区域保持可用

#### Scenario: Active project path is unavailable

- **WHEN** 当前激活项目路径不可访问
- **THEN** 状态栏显示项目路径不可用的状态
- **AND** 系统不执行 Git 状态查询

#### Scenario: Git executable is not installed

- **WHEN** 当前激活项目可用
- **AND** 系统无法找到可执行的 `git` 命令
- **THEN** 系统不执行 Git 状态查询命令
- **AND** 状态栏显示 `未安装 Git`
- **AND** 项目选择和终端会话功能保持可用

#### Scenario: Git status command fails

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
- **AND** Git 状态查询失败
- **THEN** 状态栏显示 Git 状态不可用
- **AND** 项目选择和终端会话功能保持可用

### Requirement: Initialize Active Project Git Repository From Status Bar

系统 SHALL 在当前激活项目可用、系统可以执行 `git` 命令且该项目不是 Git 仓库时，在底部状态栏提供初始化 Git 仓库的操作。该操作 SHALL 直接在当前项目路径执行 Git 初始化，并在成功后刷新状态栏中的 Git 信息。系统在执行 Git 初始化前 SHALL 校验 `git` 命令是否存在；若 `git` 命令不存在或不可执行，系统 SHALL 不执行 Git 初始化命令，并 SHALL 返回明确的未安装 Git 错误。

#### Scenario: Non-Git project shows initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 系统可以执行 `git` 命令
- **AND** 该项目路径不是 Git 仓库
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 状态栏显示 `Initialize Git Repository` 操作

#### Scenario: Missing Git hides initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 系统无法找到可执行的 `git` 命令
- **THEN** 状态栏显示 `未安装 Git`
- **AND** 状态栏不显示 `Initialize Git Repository` 操作

#### Scenario: User initializes Git repository from status bar

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
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

#### Scenario: Git initialization is blocked when Git is missing

- **WHEN** 当前激活项目可用
- **AND** 系统无法找到可执行的 `git` 命令
- **AND** 用户或客户端请求初始化当前项目的 Git 仓库
- **THEN** 系统不执行 `git init`
- **AND** 系统返回未安装 Git 的错误

#### Scenario: Git initialization fails

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
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

### Requirement: Bulk Remove Opened Projects
The system SHALL allow the user to select one or more opened projects in the project library and remove the selected projects in one confirmed bulk action. Bulk removal SHALL remove only application records, active TODO project associations, and runtime terminal sessions for the selected projects; it SHALL NOT delete directories or files from disk. Bulk removal SHALL be all-or-nothing when any requested project ID is invalid.

#### Scenario: User selects projects for bulk removal
- **WHEN** the user opens the `项目` tab
- **AND** the project library contains `frontend-app` and `api-service`
- **AND** the user checks `frontend-app`
- **AND** the user checks `api-service`
- **THEN** the project library shows both projects as selected
- **AND** the bulk delete action is enabled
- **AND** the bulk delete action indicates that 2 projects are selected

#### Scenario: Bulk delete is unavailable without selection
- **WHEN** the user opens the `项目` tab
- **AND** no project is selected
- **THEN** the bulk delete action is disabled
- **AND** activating the bulk delete action does not remove any project

#### Scenario: User confirms bulk project removal
- **WHEN** the user selects projects `frontend-app` and `api-service` in the project library
- **AND** requests bulk deletion
- **THEN** the system shows a confirmation popover for deleting the 2 selected projects
- **WHEN** the user confirms the bulk deletion
- **THEN** the project list no longer contains `frontend-app`
- **AND** the project list no longer contains `api-service`
- **AND** the persisted opened project list no longer contains `frontend-app`
- **AND** the persisted opened project list no longer contains `api-service`
- **AND** active TODOs no longer contain associations to `frontend-app` or `api-service`
- **AND** runtime terminal sessions for `frontend-app` and `api-service` are closed
- **AND** the directories for `frontend-app` and `api-service` remain on disk
- **AND** the project selection is cleared

#### Scenario: User cancels bulk project removal
- **WHEN** the user selects projects `frontend-app` and `api-service` in the project library
- **AND** requests bulk deletion
- **AND** the system shows the bulk deletion confirmation popover
- **WHEN** the user cancels the confirmation
- **THEN** the project list still contains `frontend-app`
- **AND** the project list still contains `api-service`
- **AND** TODO project associations remain unchanged
- **AND** runtime terminal sessions remain unchanged

#### Scenario: Active project is removed by bulk deletion
- **WHEN** the active project is `frontend-app`
- **AND** the user bulk deletes `frontend-app` and `api-service`
- **THEN** the system selects the remaining opened project with the most recent selection time as the selected project for project management
- **AND** if no opened projects remain, the selected project is empty
- **AND** any active terminal owned by removed projects is cleared

#### Scenario: Bulk removal request includes a missing project
- **WHEN** the system receives a bulk removal request for `frontend-app` and missing project `missing-project`
- **THEN** the system reports an error
- **AND** the opened project list remains unchanged
- **AND** TODO project associations remain unchanged
- **AND** runtime terminal sessions remain unchanged

