## MODIFIED Requirements

### Requirement: Persist Todos
系统 SHALL 在当前 workspace 数据目录中持久化 TODO、TODO 描述、TODO 优先级、TODO 工作流状态、TODO 与项目的关联、TODO 选中状态和已完成状态，并 SHALL 在该 workspace 重新打开后恢复。不同 workspace 的 TODO 数据 SHALL NOT 全局共享。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。旧数据中状态为 `active` 的 TODO SHALL 按 `not-started` 处理。旧数据中状态为 `archived` 且归档原因为 `completed` 的 TODO SHALL 按 `completed` 处理。旧数据中状态为 `archived` 且归档原因为 `deleted` 的 TODO SHALL 不在 TODO 工作区列表中展示。

#### Scenario: Todo workspace is restored after reopening workspace
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户关闭并重新打开 workspace `/work/customer-a`
- **THEN** `执行中` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** TODO `修复登录问题` 的状态仍为 `in-progress`
- **AND** `frontend-app` 仍保存为该 TODO 下的关联项目

#### Scenario: Todo workspace is isolated by workspace
- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** TODO tab 不显示 `修复登录问题`

#### Scenario: Legacy todo without priority uses medium
- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开该 workspace
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

#### Scenario: Legacy active todo becomes not-started
- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `active`
- **AND** 用户打开该 workspace
- **THEN** `未执行` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `not-started` 处理

#### Scenario: Legacy completed archived todo remains completed
- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `archived`
- **AND** TODO `修复登录问题` 的归档原因为 `completed`
- **AND** 用户打开该 workspace
- **THEN** `已完成` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `completed` 处理

#### Scenario: Legacy deleted archived todo is hidden
- **WHEN** 当前 workspace 持久化数据中 TODO `废弃任务` 的状态为 `archived`
- **AND** TODO `废弃任务` 的归档原因为 `deleted`
- **AND** 用户打开该 workspace
- **THEN** TODO tab 不在 `未执行` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `执行中` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `已完成` 视图显示 `废弃任务`

