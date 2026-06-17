# workspace-lifecycle Specification

## Purpose
TBD - created by archiving change workspace-scoped-project-management. Update Purpose after archive.
## Requirements
### Requirement: Open Workspace From Directory
系统 SHALL 允许用户通过文件菜单选择一个本地目录作为当前 workspace。打开 workspace 时，系统 SHALL 在所选目录下使用 `.data` 目录保存和加载该 workspace 的应用数据，并 SHALL 将该 workspace 记录到最近打开列表。

#### Scenario: User opens workspace from file menu
- **WHEN** 用户选择文件菜单中的 `打开项目`
- **AND** 用户在目录选择器中选择 `/home/user/work/customer-a`
- **THEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 系统使用 `/home/user/work/customer-a/.data` 作为该 workspace 的数据目录
- **AND** 最近打开列表包含 `/home/user/work/customer-a`

#### Scenario: User cancels opening workspace
- **WHEN** 用户选择文件菜单中的 `打开项目`
- **AND** 用户取消目录选择
- **THEN** 当前 workspace 保持不变
- **AND** 最近打开列表保持不变

#### Scenario: Opening unavailable workspace fails safely
- **WHEN** 用户选择文件菜单中的 `打开项目`
- **AND** 用户选择的目录不可访问
- **THEN** 系统报告 workspace 无法打开
- **AND** 当前 workspace 保持不变
- **AND** 系统不创建或切换项目库、TODO 或终端状态

### Requirement: Store Workspace Data Under Data Directory
系统 SHALL 将当前 workspace 相关数据保存在该 workspace 目录下的 `.data` 子目录中。workspace 相关数据 SHALL 包含导入工程列表、TODO、TODO 与工程关联、选中上下文、终端历史和 TODO 工程 UI 状态。TODO 工程 UI 状态 SHALL 包含按 TODO 工程保存的 TODO 视图标签和左侧 TODO 栏宽度。终端 shell 设置、终端启动配置和外观设置 SHALL 作为应用全局 settings 保存，不属于 workspace `.data`。

#### Scenario: New workspace creates data directory
- **WHEN** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **AND** `/home/user/work/customer-a/.data` 不存在
- **THEN** 系统创建 `/home/user/work/customer-a/.data`
- **AND** 后续 workspace 数据写入该 `.data` 目录

#### Scenario: Workspace data is loaded from data directory
- **WHEN** `/home/user/work/customer-a/.data/projects.json` 包含项目库和 TODO 数据
- **AND** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **THEN** 系统从 `/home/user/work/customer-a/.data/projects.json` 加载项目库和 TODO 数据

#### Scenario: Todo project UI state is loaded from data directory
- **WHEN** `/home/user/work/customer-a/.data/todo-project-ui-state.json` 包含 TODO 工程 UI 状态
- **AND** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **THEN** 系统从 `/home/user/work/customer-a/.data/todo-project-ui-state.json` 加载 TODO 工程 UI 状态

#### Scenario: Missing todo project UI state file uses empty state
- **WHEN** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **AND** `/home/user/work/customer-a/.data/todo-project-ui-state.json` 不存在
- **THEN** 系统按空 TODO 工程 UI 状态处理
- **AND** workspace 打开流程不失败

### Requirement: Display Workspace File Menu
系统 SHALL 在桌面原生菜单栏的 `文件` 菜单下提供 `打开项目`、`最近打开`、`清理最近打开` 和 `关闭` 四个选项。

#### Scenario: File menu contains workspace actions
- **WHEN** 应用启动
- **THEN** 原生菜单栏包含 `文件` 菜单
- **AND** `文件` 菜单包含 `打开项目`
- **AND** `文件` 菜单包含 `最近打开`
- **AND** `文件` 菜单包含 `清理最近打开`
- **AND** `文件` 菜单包含 `关闭`

#### Scenario: Recent menu opens recent workspace picker
- **WHEN** 用户选择文件菜单中的 `最近打开`
- **THEN** 系统显示最近打开 workspace 列表
- **AND** 用户可以从该列表选择一个 workspace 打开

### Requirement: Manage Recent Workspaces
系统 SHALL 在应用全局配置中保存最近打开 workspace 记录。最近打开记录 SHALL 使用规范化后的绝对路径去重，并 SHALL 按最近打开时间倒序展示。清理最近打开 SHALL 删除全部最近打开记录，但 SHALL NOT 删除任何 workspace 目录或 `.data` 数据。

#### Scenario: Opening workspace updates recent list
- **WHEN** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **AND** 用户随后打开 `/home/user/work/customer-b` 作为 workspace
- **THEN** 最近打开列表包含 `/home/user/work/customer-b`
- **AND** 最近打开列表包含 `/home/user/work/customer-a`
- **AND** `/home/user/work/customer-b` 排在 `/home/user/work/customer-a` 之前

#### Scenario: Reopening existing recent workspace deduplicates path
- **WHEN** 最近打开列表已包含 `/home/user/work/customer-a`
- **AND** 用户再次打开 `/home/user/work/customer-a`
- **THEN** 最近打开列表只包含一个 `/home/user/work/customer-a` 记录
- **AND** 该记录的最近打开时间被更新

#### Scenario: Selecting recent workspace opens it
- **WHEN** 最近打开列表包含 `/home/user/work/customer-a`
- **AND** 用户从最近打开列表选择 `/home/user/work/customer-a`
- **THEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** 系统加载 `/home/user/work/customer-a/.data` 中的数据

#### Scenario: Clear recent workspaces
- **WHEN** 最近打开列表包含一个或多个 workspace
- **AND** 用户选择文件菜单中的 `清理最近打开`
- **THEN** 最近打开列表为空
- **AND** 已存在的 workspace 目录和 `.data` 数据不被删除

### Requirement: Close Current Workspace
系统 SHALL 允许用户通过文件菜单关闭当前 workspace。关闭 workspace SHALL 清空当前 workspace 上下文和运行时终端，但 SHALL NOT 删除 workspace 目录、`.data` 目录或最近打开记录。

#### Scenario: User closes current workspace
- **WHEN** 当前 workspace 为 `/home/user/work/customer-a`
- **AND** 用户选择文件菜单中的 `关闭`
- **THEN** 当前 workspace 为空
- **AND** 项目库、TODO、当前项目、当前 TODO 项目和当前终端上下文为空
- **AND** `/home/user/work/customer-a/.data` 仍保留在磁盘上

#### Scenario: Closing workspace stops runtime terminals
- **WHEN** 当前 workspace 中存在运行时终端
- **AND** 用户关闭当前 workspace
- **THEN** 系统关闭这些运行时终端进程
- **AND** 前端不再显示这些终端
- **AND** 关闭后的终端不再接收输入或输出

#### Scenario: Close is unavailable without workspace
- **WHEN** 当前没有打开 workspace
- **THEN** 文件菜单中的 `关闭` 不执行 workspace 关闭流程
- **AND** 系统不改变最近打开列表

### Requirement: Show No Workspace State
系统 SHALL 在没有当前 workspace 时显示无 workspace 空态，并 SHALL 禁用或拒绝依赖 workspace 的项目库、TODO、终端和 Git 操作。

#### Scenario: Startup without current workspace shows empty state
- **WHEN** 应用启动
- **AND** 没有当前打开 workspace
- **AND** 最近打开列表为空或最近一次打开的 workspace 不可访问
- **THEN** TODO 工作区显示无 workspace 空态
- **AND** 项目库显示无 workspace 空态
- **AND** 终端区域不显示任何 workspace 终端

#### Scenario: Startup restores most recent workspace
- **WHEN** 应用启动
- **AND** 最近打开列表中的第一项为 `/work/customer-b`
- **AND** `/work/customer-b` 可访问
- **THEN** 当前 workspace 路径为 `/work/customer-b`
- **AND** 系统加载 `/work/customer-b/.data` 中的数据

#### Scenario: Startup does not fallback when most recent workspace is unavailable
- **WHEN** 应用启动
- **AND** 最近打开列表中的第一项为 `/work/customer-b`
- **AND** `/work/customer-b` 不可访问
- **AND** 最近打开列表中还包含可访问的 `/work/customer-a`
- **THEN** 当前 workspace 为空
- **AND** 系统不自动打开 `/work/customer-a`
- **AND** 最近打开列表仍保留 `/work/customer-b`

#### Scenario: Workspace dependent action is rejected without workspace
- **WHEN** 当前没有打开 workspace
- **AND** 用户尝试创建 TODO、导入工程或创建终端
- **THEN** 系统不执行该操作
- **AND** 系统提示需要先打开项目

### Requirement: Isolate Workspaces
系统 SHALL 保证不同 workspace 的项目库、TODO 和终端历史互相隔离。打开或切换 workspace SHALL 只展示目标 workspace 的数据。应用全局 settings SHALL NOT 随 workspace 切换而变化。

#### Scenario: Projects and todos do not leak across workspaces
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录`
- **AND** 用户导入工程 `frontend-a`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** `/work/customer-b` 中不显示 TODO `修复登录`
- **AND** `/work/customer-b` 中不显示工程 `frontend-a`

#### Scenario: Switching workspace clears runtime terminal context
- **WHEN** workspace `/work/customer-a` 中存在运行时终端
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** `/work/customer-a` 的运行时终端被关闭并从前端移除
- **AND** 当前项目、当前 TODO 项目和当前终端上下文来自 `/work/customer-b` 或为空

