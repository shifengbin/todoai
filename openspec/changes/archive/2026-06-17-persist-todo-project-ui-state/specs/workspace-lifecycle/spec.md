## MODIFIED Requirements

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
