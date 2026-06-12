## MODIFIED Requirements

### Requirement: Manage Todo Action Confirmation Popovers

系统 SHALL 在 TODO item 的完成按钮和右键菜单删除操作上使用确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭 TODO 右键菜单、终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。删除确认气泡 SHALL 锚定在对应 TODO item 的操作上下文附近，且 SHALL NOT 脱离当前 TODO item 显示到侧边栏右上角。

#### Scenario: Complete confirmation popover opens beside complete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的完成按钮
- **THEN** 系统在完成按钮旁显示完成确认气泡
- **AND** 系统不立即完成 TODO `修复登录问题`

#### Scenario: Delete confirmation popover opens from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择删除 TODO
- **THEN** 系统关闭 TODO 右键菜单
- **AND** 系统在 TODO `修复登录问题` 的操作上下文附近显示删除确认气泡
- **AND** 系统不立即删除 TODO `修复登录问题`

#### Scenario: Delete confirmation popover stays anchored to the todo item

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择删除 TODO
- **THEN** 删除确认气泡显示在 TODO `修复登录问题` 所属操作区域附近
- **AND** 删除确认气泡不显示在侧边栏右上角

#### Scenario: Opening another sidebar popover closes todo action popover

- **WHEN** TODO `修复登录问题` 的完成确认气泡已显示
- **AND** 用户打开 TODO 下项目 `frontend-app` 的终端启动菜单
- **THEN** 系统关闭 TODO `修复登录问题` 的完成确认气泡
- **AND** 系统显示 `frontend-app` 的终端启动菜单

#### Scenario: Outside click closes todo action popover without action

- **WHEN** TODO `修复登录问题` 的删除确认气泡已显示
- **AND** 用户点击确认气泡外部
- **THEN** 系统关闭删除确认气泡
- **AND** TODO `修复登录问题` 保持在活动任务列表中
