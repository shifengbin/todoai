## ADDED Requirements

### Requirement: Persist Todo Project UI State

系统 SHALL 以 TODO 工程为单位持久化 TODO 工作区 UI 状态。TODO 工程 UI 状态 SHALL 包含上次选择的 TODO 视图标签和左侧 TODO 栏宽度。TODO 视图标签 SHALL 支持 `not-started`、`in-progress` 和 `completed`，分别对应 `未执行`、`执行中` 和 `已完成`。TODO 工程 UI 状态 SHALL 按当前 workspace 隔离，并 SHALL 在应用重启、前端重新加载或重新打开 workspace 后恢复。

#### Scenario: Todo project UI state is restored after reopening workspace
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 当前 TODO 工程为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 用户选择 `已完成` 视图
- **AND** 用户将左侧 TODO 栏宽度调整为 `360`
- **AND** 用户关闭并重新打开 workspace `/work/customer-a`
- **THEN** 当前 TODO 工程仍为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** TODO 工作区显示 `已完成` 视图
- **AND** 左侧 TODO 栏宽度恢复为 `360`

#### Scenario: Todo project UI state is isolated by todo project
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** TODO 工程 `frontend-app` 的 TODO 视图标签已保存为 `completed`
- **AND** TODO 工程 `frontend-app` 的左侧 TODO 栏宽度已保存为 `360`
- **AND** TODO 工程 `api-service` 的 TODO 视图标签已保存为 `in-progress`
- **AND** TODO 工程 `api-service` 的左侧 TODO 栏宽度已保存为 `420`
- **WHEN** 用户选择 TODO 工程 `frontend-app`
- **THEN** TODO 工作区显示 `已完成` 视图
- **AND** 左侧 TODO 栏宽度为 `360`
- **WHEN** 用户选择 TODO 工程 `api-service`
- **THEN** TODO 工作区显示 `执行中` 视图
- **AND** 左侧 TODO 栏宽度为 `420`

#### Scenario: Todo project without UI state uses defaults
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 当前 TODO 工程没有已保存的 UI 状态
- **THEN** TODO 工作区显示 `未执行` 视图
- **AND** 左侧 TODO 栏宽度使用系统默认值

#### Scenario: Todo view selection is saved for active todo project
- **WHEN** 当前 TODO 工程为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 用户点击 `执行中` 视图标签
- **THEN** 系统将该 TODO 工程的 TODO 视图标签保存为 `in-progress`
- **AND** 用户重新选择该 TODO 工程时恢复 `执行中` 视图

#### Scenario: Sidebar width is saved after divider drag ends
- **WHEN** 当前 TODO 工程为 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 用户拖动分割线将左侧 TODO 栏宽度调整为 `380`
- **AND** 用户结束拖动
- **THEN** 系统将该 TODO 工程的左侧 TODO 栏宽度保存为 `380`
- **AND** 用户重新选择该 TODO 工程时恢复左侧 TODO 栏宽度 `380`

#### Scenario: Removing todo project removes its UI state
- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 已保存 TODO 工程 UI 状态
- **AND** 用户从 TODO `修复登录问题` 移除工程 `frontend-app`
- **THEN** 系统删除该 TODO 工程的 UI 状态
- **AND** 其他 TODO 工程的 UI 状态保持不变
