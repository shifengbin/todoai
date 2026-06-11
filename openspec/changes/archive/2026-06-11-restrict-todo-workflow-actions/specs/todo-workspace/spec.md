## ADDED Requirements

### Requirement: Restrict Todo Workflow Actions

系统 SHALL 按 TODO 当前工作流状态限制用户可执行的状态动作。`not-started` TODO SHALL 允许开始和删除，SHALL NOT 暴露完成入口或退回入口。`in-progress` TODO SHALL 允许完成和删除，SHALL NOT 暴露退回 `not-started` 的入口。查看/编辑 TODO 和添加项目 SHALL 在 `not-started` 与 `in-progress` 状态下保持可用。

#### Scenario: Not-started todo exposes start and delete actions

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 `未执行` 视图查看该 TODO
- **THEN** 系统显示开始 TODO 的入口
- **AND** 系统显示删除 TODO 的入口
- **AND** 系统不显示完成 TODO 的入口
- **AND** 系统不显示退回未执行的入口

#### Scenario: In-progress todo exposes complete and delete actions

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 `执行中` 视图查看该 TODO
- **THEN** 系统显示完成 TODO 的入口
- **AND** 系统显示删除 TODO 的入口
- **AND** 系统不显示退回未执行的入口

#### Scenario: User starts a not-started todo

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **THEN** TODO `修复登录问题` 的状态保存为 `in-progress`
- **AND** TODO `修复登录问题` 显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `未执行` 视图

#### Scenario: In-progress todo cannot move back to not-started

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户或客户端请求将 TODO `修复登录问题` 标记为未执行
- **THEN** 系统拒绝该状态变更
- **AND** TODO `修复登录问题` 的状态保持为 `in-progress`

## MODIFIED Requirements

### Requirement: Show Todo Project Terminal Launch Menu

系统 SHALL 仅在 `in-progress` TODO 的可用项目行上提供终端启动菜单。启动菜单 SHALL 包含内置 `Terminal` 选项和已配置的终端启动配置。`not-started` TODO 的项目行 SHALL NOT 暴露新增终端启动菜单。

#### Scenario: In-progress todo project launch menu contains configured profiles

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 设置中包含启动配置 `codex` 和 `claude`
- **AND** 用户激活 TODO `修复登录问题` 下项目 `frontend-app` 的新增终端控件
- **THEN** 启动菜单显示 `Terminal` 作为第一项
- **AND** 启动菜单按配置顺序显示 `codex` 和 `claude`

#### Scenario: Not-started todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

#### Scenario: Unavailable todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径不可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

### Requirement: Complete Todo

系统 SHALL 允许用户通过 TODO item 完成按钮旁的确认气泡完成 `in-progress` TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 状态保存为 `completed`。用户取消确认 SHALL 不改变 TODO 或其终端状态。系统 SHALL NOT 允许 `not-started` TODO 直接完成。

#### Scenario: User completes an in-progress todo

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户在确认气泡中确认完成
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 显示在 `已完成` 视图
- **AND** TODO `修复登录问题` 的状态为 `completed`

#### Scenario: Not-started todo cannot be completed

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户或客户端请求完成 TODO `修复登录问题`
- **THEN** 系统拒绝完成请求
- **AND** TODO `修复登录问题` 的状态保持为 `not-started`

#### Scenario: User cancels todo completion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO item 删除按钮旁的确认气泡删除 `not-started` 或 `in-progress` TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 从可见 TODO 工作区列表中移除。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a not-started todo

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User deletes an in-progress todo

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User cancels todo deletion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变
