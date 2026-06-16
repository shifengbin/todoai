## ADDED Requirements

### Requirement: View Completed Todo Details

系统 SHALL 允许用户从 `已完成` 视图打开 completed TODO 的详情。completed TODO 详情 SHALL 复用 TODO 详情弹窗，并 SHALL 以只读模式显示 TODO 标题、描述、优先级和完成时项目快照。只读详情 SHALL 隐藏保存按钮，SHALL 禁止编辑 TODO 字段或项目，且 SHALL NOT 恢复终端、启动 shell 进程或重新建立项目关联。

#### Scenario: User opens completed todo details

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级为 `高`
- **AND** TODO `修复登录问题` 的完成时项目快照包含名称为 `frontend-app`、路径为 `/work/frontend-app` 的项目
- **AND** 用户在 `已完成` 视图中打开 TODO `修复登录问题` 的详情
- **THEN** 系统打开 TODO 详情弹窗
- **AND** 详情弹窗显示标题 `修复登录问题`
- **AND** 详情弹窗显示描述 `登录后跳回首页`
- **AND** 详情弹窗显示优先级 `高`
- **AND** 详情弹窗显示项目快照 `frontend-app`
- **AND** 详情弹窗显示项目快照路径 `/work/frontend-app`

#### Scenario: Completed todo details are read-only

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统隐藏详情弹窗的保存按钮
- **AND** 系统不允许编辑 TODO 标题
- **AND** 系统不允许编辑 TODO 描述
- **AND** 系统不允许修改 TODO 优先级
- **AND** 系统不允许新增或移除项目

#### Scenario: Completed todo details do not restore runtime context

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程
- **AND** 系统不把完成时项目快照恢复为活动 TODO 项目关联

### Requirement: Bulk Delete Completed Todos

系统 SHALL 在 `已完成` 视图中提供 completed TODO 的多选批量删除能力。批量删除 SHALL 仅适用于 `已完成` 视图中的 completed TODO，SHALL 在删除前要求用户确认，确认后 SHALL 将选中的 completed TODO 从可见 TODO 工作区列表中移除。系统 SHALL NOT 在 `未执行` 或 `执行中` 视图中提供 TODO 批量删除能力。

#### Scenario: User bulk deletes completed todos

- **WHEN** `已完成` 视图显示 completed TODO `修复登录问题`
- **AND** `已完成` 视图显示 completed TODO `整理文档`
- **AND** 用户选择 TODO `修复登录问题`
- **AND** 用户选择 TODO `整理文档`
- **AND** 用户触发批量删除
- **AND** 系统显示批量删除确认
- **AND** 用户确认批量删除
- **THEN** `已完成` 视图不再显示 TODO `修复登录问题`
- **AND** `已完成` 视图不再显示 TODO `整理文档`

#### Scenario: User cancels completed todo bulk deletion

- **WHEN** `已完成` 视图显示 completed TODO `修复登录问题`
- **AND** 用户选择 TODO `修复登录问题`
- **AND** 用户触发批量删除
- **AND** 系统显示批量删除确认
- **AND** 用户取消批量删除
- **THEN** `已完成` 视图仍显示 TODO `修复登录问题`

#### Scenario: Open todo views do not expose bulk delete

- **WHEN** 用户在 TODO tab 中打开 `未执行` 视图
- **THEN** 系统不显示 TODO 批量删除入口
- **WHEN** 用户在 TODO tab 中打开 `执行中` 视图
- **THEN** 系统不显示 TODO 批量删除入口

#### Scenario: Bulk delete rejects non-completed todos

- **WHEN** 用户或客户端请求批量删除 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **THEN** 系统拒绝批量删除请求
- **AND** TODO `修复登录问题` 仍显示在 `执行中` 视图

## MODIFIED Requirements

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO 右键菜单中的删除动作和确认气泡删除 `not-started`、`in-progress` 或 `completed` TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 从可见 TODO 工作区列表中移除。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a not-started todo

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User deletes an in-progress todo

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User deletes a completed todo

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户在 `已完成` 视图中打开 TODO `修复登录问题` 的操作菜单
- **AND** 用户选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 不再显示在 `已完成` 视图
- **AND** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程

#### Scenario: User cancels todo deletion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变
