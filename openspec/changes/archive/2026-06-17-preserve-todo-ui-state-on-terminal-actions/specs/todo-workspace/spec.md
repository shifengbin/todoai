## MODIFIED Requirements

### Requirement: Persist Todo Project UI State

系统 SHALL 以 TODO 工程为单位持久化 TODO 视图标签状态，并以 workspace 为单位持久化左侧 TODO 栏宽度。TODO 工程 UI 状态 SHALL 包含上次选择的 TODO 视图标签。TODO 视图标签 SHALL 支持 `not-started`、`in-progress` 和 `completed`，分别对应 `未执行`、`执行中` 和 `已完成`。左侧 TODO 栏宽度 SHALL 属于当前 workspace 的 TODO 工作区布局状态，SHALL NOT 属于某个 TODO item、TODO 工程或终端。

TODO 工程 UI 状态和 workspace 级左侧栏宽度 SHALL 按当前 workspace 隔离，并 SHALL 在应用重启、前端重新加载或重新打开 workspace 后恢复。用户主动选择 TODO 工程时，系统 SHALL 恢复该 TODO 工程的 TODO 视图标签，但 SHALL NOT 因该操作改变左侧 TODO 栏宽度。除用户拖动侧栏、用户切换 TODO 视图、打开 workspace、前端重新加载、重新打开 workspace 和用户主动选择 TODO 工程外，其他业务状态刷新 SHALL NOT 改变当前左侧 TODO 栏宽度或当前 TODO 视图。添加终端、切换终端和终端状态刷新 MUST NOT 改变当前左侧 TODO 栏宽度或当前 TODO 视图。

#### Scenario: Todo project UI state and workspace sidebar width are restored after reopening workspace
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 当前 TODO 工程为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 用户选择 `已完成` 视图
- **AND** 用户将左侧 TODO 栏宽度调整为 `360`
- **AND** 用户关闭并重新打开 workspace `/work/customer-a`
- **THEN** 当前 TODO 工程仍为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** TODO 工作区显示 `已完成` 视图
- **AND** 左侧 TODO 栏宽度恢复为 `360`

#### Scenario: Todo view state is isolated by todo project
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** TODO 工程 `frontend-app` 的 TODO 视图标签已保存为 `completed`
- **AND** TODO 工程 `api-service` 的 TODO 视图标签已保存为 `in-progress`
- **AND** TODO 工程 `frontend-app` 和 `api-service` 可以属于同一个 TODO
- **WHEN** 用户选择 TODO 工程 `frontend-app`
- **THEN** TODO 工作区显示 `已完成` 视图
- **WHEN** 用户选择 TODO 工程 `api-service`
- **THEN** TODO 工作区显示 `执行中` 视图

#### Scenario: Sidebar width is shared across todo projects in the workspace
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 左侧 TODO 栏宽度已保存为 `380`
- **AND** TODO 工程 `frontend-app` 的 TODO 视图标签已保存为 `completed`
- **AND** TODO 工程 `api-service` 的 TODO 视图标签已保存为 `in-progress`
- **WHEN** 用户选择 TODO 工程 `frontend-app`
- **THEN** 左侧 TODO 栏宽度为 `380`
- **WHEN** 用户选择 TODO 工程 `api-service`
- **THEN** 左侧 TODO 栏宽度仍为 `380`

#### Scenario: Todo project without UI state uses default todo view
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 当前 TODO 工程没有已保存的 TODO 视图标签
- **THEN** TODO 工作区显示 `未执行` 视图

#### Scenario: Workspace without sidebar width state uses default width
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 当前 workspace 没有已保存的左侧 TODO 栏宽度
- **THEN** 左侧 TODO 栏宽度使用系统默认值

#### Scenario: Todo view selection is saved for active todo project
- **WHEN** 当前 TODO 工程为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 用户点击 `执行中` 视图标签
- **THEN** 系统将该 TODO 工程的 TODO 视图标签保存为 `in-progress`
- **AND** 用户重新选择该 TODO 工程时恢复 `执行中` 视图

#### Scenario: Sidebar width is saved after divider drag ends
- **WHEN** 用户拖动分割线将左侧 TODO 栏宽度调整为 `380`
- **AND** 用户结束拖动
- **THEN** 系统将当前 workspace 的左侧 TODO 栏宽度保存为 `380`
- **AND** 用户重新打开该 workspace 时恢复左侧 TODO 栏宽度 `380`

#### Scenario: Business state refresh preserves current sidebar width
- **WHEN** 当前 workspace 的左侧 TODO 栏宽度当前为 `380`
- **AND** 用户创建 TODO `整理文档`
- **THEN** 左侧 TODO 栏宽度仍为 `380`

#### Scenario: Business state refresh preserves current todo view
- **WHEN** TODO 工作区当前视图为 `执行中`
- **AND** 用户创建 TODO `整理文档`
- **THEN** TODO 工作区当前视图仍为 `执行中`

#### Scenario: Adding terminal preserves current todo view and sidebar width
- **WHEN** TODO 工作区当前视图为 `执行中`
- **AND** 左侧 TODO 栏宽度当前为 `380`
- **AND** 用户为 TODO 工程 `frontend-app` 添加终端
- **THEN** TODO 工作区当前视图仍为 `执行中`
- **AND** 左侧 TODO 栏宽度仍为 `380`

#### Scenario: Selecting terminal preserves current todo view and sidebar width
- **WHEN** TODO 工作区当前视图为 `执行中`
- **AND** 左侧 TODO 栏宽度当前为 `380`
- **AND** 用户选择 TODO item 下的终端 `terminal-a`
- **THEN** TODO 工作区当前视图仍为 `执行中`
- **AND** 左侧 TODO 栏宽度仍为 `380`

#### Scenario: Terminal status refresh preserves current todo view and sidebar width
- **WHEN** TODO 工作区当前视图为 `执行中`
- **AND** 左侧 TODO 栏宽度当前为 `380`
- **AND** 终端 `terminal-a` 的运行状态刷新
- **THEN** TODO 工作区当前视图仍为 `执行中`
- **AND** 左侧 TODO 栏宽度仍为 `380`

#### Scenario: Removing todo project removes its UI state
- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 已保存 TODO 工程 UI 状态
- **AND** 用户从 TODO `修复登录问题` 移除工程 `frontend-app`
- **THEN** 系统删除该 TODO 工程的 UI 状态
- **AND** 其他 TODO 工程的 UI 状态保持不变
- **AND** 当前 workspace 的左侧 TODO 栏宽度状态保持不变
