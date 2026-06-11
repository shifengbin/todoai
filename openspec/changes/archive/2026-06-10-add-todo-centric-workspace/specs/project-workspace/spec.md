## ADDED Requirements

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

## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Display Project Terminal Tree

**Reason**: Terminal hierarchy is now organized by TODO context instead of project library context.
**Migration**: Use `todo-workspace` requirement `Display Todo Project Terminal Tree`.

### Requirement: Select Active Terminal From Project Tree

**Reason**: Terminal selection from the project library would bypass TODO context and make terminal purpose ambiguous.
**Migration**: Select terminals from the TODO tree under a TODO project context.

### Requirement: Collapse Project Terminal Branches

**Reason**: Project library rows no longer own terminal child rows.
**Migration**: Collapse TODO branches and TODO project terminal branches in the TODO tab.

### Requirement: Reveal Active Project Terminal Branch

**Reason**: Active terminal reveal behavior now belongs to the TODO tree.
**Migration**: Reveal the active TODO and TODO project branch when selecting or creating a terminal.

### Requirement: Show Project Terminal Hierarchy

**Reason**: The visible terminal hierarchy moved from project rows to TODO project rows.
**Migration**: Use the TODO tab hierarchy `TODO -> 项目 -> 终端`.

### Requirement: Highlight Active Terminal Branch Guide

**Reason**: Branch guide highlighting is now tied to the active TODO project terminal branch.
**Migration**: Apply active branch guide styling in the TODO tree.

### Requirement: Display Interactive Terminal Activity In Project Tree

**Reason**: Interactive activity indicators must appear where terminals are displayed, which is now the TODO tree.
**Migration**: Display terminal activity indicators on TODO tree terminal rows.

### Requirement: Show Project Terminal Launch Menu

**Reason**: Terminal launch is no longer available from the project library.
**Migration**: Use the launch menu on a project row inside the TODO tab.

### Requirement: Create Terminal From Launch Menu

**Reason**: Creating a terminal must bind the terminal to a TODO project context.
**Migration**: Create terminals from the TODO project launch menu.
