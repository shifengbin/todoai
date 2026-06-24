## MODIFIED Requirements

### Requirement: Store Workspace Data Under Data Directory

系统 SHALL 将当前 workspace 相关元数据保存在该 workspace 目录下的 `.data` 子目录中。workspace 相关元数据 SHALL 包含导入工程列表、TODO、TODO 与工程关联、任务工作区目录引用、TODO 项目 worktree 元数据、选中上下文、终端历史和 TODO 工程 UI 状态。TODO 工程 UI 状态 SHALL 包含按 TODO 工程保存的 TODO 视图标签和左侧 TODO 栏宽度。任务工作区目录和 Git worktree 内容 SHALL 保存在 workspace 根目录下的任务工作区根目录中，SHALL NOT 保存在 `.data` 目录中。终端 shell 设置、终端启动配置和外观设置 SHALL 作为应用全局 settings 保存，不属于 workspace `.data`。

#### Scenario: New workspace creates data directory

- **WHEN** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **AND** `/home/user/work/customer-a/.data` 不存在
- **THEN** 系统创建 `/home/user/work/customer-a/.data`
- **AND** 后续 workspace 元数据写入该 `.data` 目录

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

#### Scenario: Todo workspace directory is outside data directory

- **WHEN** 用户打开 `/home/user/work/customer-a` 作为 workspace
- **AND** TODO `修复登录问题` 创建任务工作区目录
- **THEN** 任务工作区目录位于 `/home/user/work/customer-a` 下的任务工作区根目录中
- **AND** 任务工作区目录不位于 `/home/user/work/customer-a/.data` 下
- **AND** `.data/projects.json` 保存该 TODO 的任务工作区目录引用

### Requirement: Close Current Workspace

系统 SHALL 允许用户通过文件菜单关闭当前 workspace。关闭 workspace SHALL 清空当前 workspace 上下文和运行时终端，但 SHALL NOT 删除 workspace 目录、`.data` 目录、任务工作区目录、Git worktree 或最近打开记录。

#### Scenario: User closes current workspace

- **WHEN** 当前 workspace 为 `/home/user/work/customer-a`
- **AND** 用户选择文件菜单中的 `关闭`
- **THEN** 当前 workspace 为空
- **AND** 项目库、TODO、当前项目、当前 TODO 项目和当前终端上下文为空
- **AND** `/home/user/work/customer-a/.data` 仍保留在磁盘上
- **AND** `/home/user/work/customer-a` 下的任务工作区目录仍保留在磁盘上

#### Scenario: Closing workspace stops runtime terminals

- **WHEN** 当前 workspace 中存在运行时终端
- **AND** 用户关闭当前 workspace
- **THEN** 系统关闭这些运行时终端进程
- **AND** 前端不再显示这些终端
- **AND** 关闭后的终端不再接收输入或输出
- **AND** 系统不删除任何任务工作区目录或 Git worktree

#### Scenario: Close is unavailable without workspace

- **WHEN** 当前没有打开 workspace
- **THEN** 文件菜单中的 `关闭` 不执行 workspace 关闭流程
- **AND** 系统不改变最近打开列表

### Requirement: Isolate Workspaces

系统 SHALL 保证不同 workspace 的项目库、TODO、任务工作区元数据和终端历史互相隔离。打开或切换 workspace SHALL 只展示目标 workspace 的数据。应用全局 settings SHALL NOT 随 workspace 切换而变化。切换 workspace SHALL 停止源 workspace 的运行时终端，但 SHALL NOT 删除源 workspace 的任务工作区目录或 Git worktree。

#### Scenario: Projects and todos do not leak across workspaces

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录`
- **AND** 用户导入工程 `frontend-a`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** `/work/customer-b` 中不显示 TODO `修复登录`
- **AND** `/work/customer-b` 中不显示工程 `frontend-a`

#### Scenario: Todo workspace metadata does not leak across workspaces

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** TODO `修复登录` 创建任务工作区目录
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** `/work/customer-b` 中不显示 TODO `修复登录` 的任务工作区目录引用

#### Scenario: Switching workspace clears runtime terminal context

- **WHEN** workspace `/work/customer-a` 中存在运行时终端
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** `/work/customer-a` 的运行时终端被关闭并从前端移除
- **AND** 当前项目、当前 TODO 项目和当前终端上下文来自 `/work/customer-b` 或为空
- **AND** `/work/customer-a` 的任务工作区目录和 Git worktree 保留在磁盘上
