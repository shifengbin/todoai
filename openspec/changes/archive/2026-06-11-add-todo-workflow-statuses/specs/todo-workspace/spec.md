## ADDED Requirements

### Requirement: Display Todo Workflow Status Views

系统 SHALL 在 TODO tab 中提供 `未执行`、`执行中`、`已完成` 三个状态视图。`未执行` 视图 SHALL 只显示状态为 `not-started` 的 TODO，`执行中` 视图 SHALL 只显示状态为 `in-progress` 的 TODO，`已完成` 视图 SHALL 只显示状态为 `completed` 的 TODO。

#### Scenario: User views not-started todos

- **WHEN** TODO `整理文档` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户打开 `未执行` 视图
- **THEN** TODO tab 显示 `整理文档`
- **AND** TODO tab 不显示 `修复登录问题`

#### Scenario: User views in-progress todos

- **WHEN** TODO `整理文档` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户打开 `执行中` 视图
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO tab 不显示 `整理文档`

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户打开 `已完成` 视图
- **THEN** TODO tab 显示 `修复登录问题`

#### Scenario: Deleted todos are hidden from workflow views

- **WHEN** 用户删除 TODO `修复登录问题`
- **THEN** TODO `修复登录问题` 不显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

### Requirement: Manually Change Todo Workflow Status

系统 SHALL 允许用户手动在 `未执行` 和 `执行中` 状态之间切换 TODO。系统 SHALL NOT 根据终端创建、终端删除或终端活动状态自动改变 TODO 的工作流状态。

#### Scenario: User marks a not-started todo as in-progress

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **THEN** TODO `修复登录问题` 的状态保存为 `in-progress`
- **AND** TODO `修复登录问题` 显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `未执行` 视图

#### Scenario: User moves an in-progress todo back to not-started

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户将 TODO `修复登录问题` 标记为未执行
- **THEN** TODO `修复登录问题` 的状态保存为 `not-started`
- **AND** TODO `修复登录问题` 显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图

#### Scenario: Terminal activity does not change todo workflow status

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 TODO `修复登录问题` 下创建终端
- **THEN** TODO `修复登录问题` 的状态仍为 `not-started`

## MODIFIED Requirements

### Requirement: Create Todo

系统 SHALL 允许用户通过应用内表单创建 TODO。创建表单 SHALL 包含必填 TODO 名称、选填 TODO 描述、任务优先级和选填工程。工程 SHALL 支持多选。任务优先级 SHALL 支持 `高`、`中`、`低` 三档，并 SHALL 默认选择 `中`。创建后的 TODO SHALL 默认保存为 `not-started` 状态，出现在 `未执行` 视图中，并 SHALL 可作为后续项目关联和终端创建的上下文。创建后的 TODO 分支 SHALL 默认收起，且创建操作 SHALL 不改变当前 TODO、当前项目、当前 TODO 项目或当前终端上下文。

#### Scenario: User creates a todo without a project

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为空、优先级为 `中` 且未选择工程的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 该 TODO 的优先级为 `中`
- **AND** 该 TODO 初始不包含关联项目
- **AND** TODO `修复登录问题` 的分支默认收起

#### Scenario: User creates a todo with description and priority

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为 `登录后跳回首页`、优先级为 `高` 的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** 该 TODO 保存描述 `登录后跳回首页`
- **AND** 该 TODO 的优先级为 `高`
- **AND** 该 TODO 的状态为 `not-started`

#### Scenario: User creates a todo with optional projects

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 当前 TODO 项目上下文为 TODO `升级依赖` 下的 `docs-site`
- **AND** 用户在创建 TODO 表单中输入名称 `修复登录问题`
- **AND** 用户选择优先级 `高`
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 下保存项目 `frontend-app`
- **AND** TODO `修复登录问题` 下保存项目 `api-service`
- **AND** 当前 TODO 项目上下文仍为 TODO `升级依赖` 下的 `docs-site`
- **AND** TODO `修复登录问题` 的分支默认收起

#### Scenario: Todo title is required

- **WHEN** 用户尝试创建名称为空的 TODO
- **THEN** 系统不创建 TODO
- **AND** 系统显示不会改变当前工作区状态的错误信息

#### Scenario: Project selection can be searched while creating todo

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户打开创建 TODO 表单
- **AND** 用户在工程筛选框输入 `front`
- **THEN** 工程选择列表显示 `frontend-app`
- **AND** 工程选择列表不显示 `api-service`

#### Scenario: Selected projects can be removed while creating todo

- **WHEN** 用户在创建 TODO 表单中选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** 创建 TODO 表单在工程筛选框上方以 tag 展示 `frontend-app`
- **AND** 创建 TODO 表单在工程筛选框上方以 tag 展示 `api-service`
- **WHEN** 用户删除 `frontend-app` tag
- **THEN** 创建 TODO 表单不再选中 `frontend-app`
- **AND** 创建 TODO 表单仍选中 `api-service`

### Requirement: Persist Todos

系统 SHALL 持久化 TODO、TODO 描述、TODO 优先级、TODO 工作流状态、TODO 与项目的关联、TODO 选中状态和已完成状态，并 SHALL 在应用重启后恢复。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。旧数据中状态为 `active` 的 TODO SHALL 按 `not-started` 处理。旧数据中状态为 `archived` 且归档原因为 `completed` 的 TODO SHALL 按 `completed` 处理。旧数据中状态为 `archived` 且归档原因为 `deleted` 的 TODO SHALL 不在 TODO 工作区列表中展示。

#### Scenario: Todo workspace is restored after restart

- **WHEN** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户关闭并重新打开应用
- **THEN** `执行中` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** TODO `修复登录问题` 的状态仍为 `in-progress`
- **AND** `frontend-app` 仍保存为该 TODO 下的关联项目

#### Scenario: Legacy todo without priority uses medium

- **WHEN** 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开应用
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

#### Scenario: Legacy active todo becomes not-started

- **WHEN** 持久化数据中 TODO `修复登录问题` 的状态为 `active`
- **AND** 用户打开应用
- **THEN** `未执行` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `not-started` 处理

#### Scenario: Legacy completed archived todo remains completed

- **WHEN** 持久化数据中 TODO `修复登录问题` 的状态为 `archived`
- **AND** TODO `修复登录问题` 的归档原因为 `completed`
- **AND** 用户打开应用
- **THEN** `已完成` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `completed` 处理

#### Scenario: Legacy deleted archived todo is hidden

- **WHEN** 持久化数据中 TODO `废弃任务` 的状态为 `archived`
- **AND** TODO `废弃任务` 的归档原因为 `deleted`
- **AND** 用户打开应用
- **THEN** TODO tab 不在 `未执行` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `执行中` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `已完成` 视图显示 `废弃任务`

### Requirement: Collapse Todo Branches

系统 SHALL 允许用户独立展开和收起 `未执行` 与 `执行中` 视图中的 TODO 分支。收起 TODO SHALL 隐藏其项目和终端子项，但 SHALL 保留 TODO 行可见。若收起的 TODO 下存在终端，TODO 行 SHALL 反映被隐藏子终端的聚合活动状态；聚合优先级 SHALL 为 `needs-input` 高于 `busy` 高于 `idle`。

#### Scenario: User collapses a todo

- **WHEN** TODO `修复登录问题` 已展开并显示项目子项
- **AND** 用户激活该 TODO 的收起控件
- **THEN** 该 TODO 下的项目和终端子项被隐藏
- **AND** TODO `修复登录问题` 仍显示在当前状态视图中

#### Scenario: Collapsed todo shows hidden terminal needing input

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的终端活动提示

#### Scenario: Collapsed todo prioritizes needs input over busy

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 下还存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的终端活动提示
- **AND** TODO `修复登录问题` 行不以运行中状态作为最高优先级提示

#### Scenario: Expanded todo relies on terminal rows for activity state

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已展开
- **THEN** 该终端行显示运行中的活动提示
- **AND** TODO `修复登录问题` 行不重复显示收起态聚合活动提示

#### Scenario: User expands a todo

- **WHEN** TODO `修复登录问题` 已收起
- **AND** 用户激活该 TODO 的展开控件
- **THEN** 系统显示该 TODO 下的项目子项

### Requirement: Complete Todo

系统 SHALL 允许用户通过 TODO item 完成按钮旁的确认气泡完成 `未执行` 或 `执行中` TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 状态保存为 `completed`。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User completes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户在确认气泡中确认完成
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 显示在 `已完成` 视图
- **AND** TODO `修复登录问题` 的状态为 `completed`

#### Scenario: User cancels todo completion

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在原状态视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO item 删除按钮旁的确认气泡删除 `未执行` 或 `执行中` TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 从可见 TODO 工作区列表中移除。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User cancels todo deletion

- **WHEN** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在原状态视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供已完成查看功能。已完成列表 SHALL 只显示状态为 `completed` 的 TODO，并 SHALL 展示完成时保存的项目快照。已删除 TODO SHALL NOT 显示在已完成列表中。

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 已完成
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示该 TODO 的完成时间
- **AND** `已完成` 视图显示该 TODO 完成时关联项目的名称和路径快照

#### Scenario: Deleted todo is not shown as completed

- **WHEN** TODO `废弃任务` 已被删除
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图不显示 TODO `废弃任务`

#### Scenario: Completed todo does not restore terminals

- **WHEN** 用户打开 `已完成` 视图
- **AND** 用户查看 TODO `修复登录问题`
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程

### Requirement: Sort Active Todos

系统 SHALL 在 TODO tab 的 `未执行` 和 `执行中` 列表中提供排序切换控件，并 SHALL 支持按任务优先级排序和按创建时间排序。系统 SHALL 默认选择优先级排序。优先级排序 SHALL 为 `高`、`中`、`低`，相同优先级的 TODO SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。时间排序 SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。`已完成` 列表 SHALL 不受 `未执行` 和 `执行中` 的排序规则影响。

#### Scenario: Open todo sort control defaults to priority

- **WHEN** 用户打开 TODO tab 的 `未执行` 或 `执行中` 列表
- **THEN** 系统显示 TODO 排序切换控件
- **AND** 排序切换控件默认选中优先级排序

#### Scenario: Not-started todos are ordered by priority

- **WHEN** `未执行` 列表包含优先级为 `低` 的 TODO `整理文档`
- **AND** `未执行` 列表包含优先级为 `高` 的 TODO `修复登录问题`
- **AND** `未执行` 列表包含优先级为 `中` 的 TODO `升级依赖`
- **THEN** TODO tab 的 `未执行` 列表依次显示 `修复登录问题`、`升级依赖`、`整理文档`

#### Scenario: In-progress todos with same priority are ordered by creation time

- **WHEN** `执行中` 列表包含优先级同为 `高` 的 TODO `修复登录问题` 和 `排查线上报警`
- **AND** TODO `修复登录问题` 的 `createdAt` 早于 TODO `排查线上报警`
- **THEN** TODO tab 的 `执行中` 列表中 `修复登录问题` 排在 `排查线上报警` 前面

#### Scenario: User switches open todos to creation time order

- **WHEN** 当前状态视图包含创建时间更晚且优先级为 `高` 的 TODO `修复登录问题`
- **AND** 当前状态视图包含创建时间更早且优先级为 `低` 的 TODO `整理文档`
- **AND** 用户选择时间排序
- **THEN** 当前状态视图中 `整理文档` 排在 `修复登录问题` 前面

#### Scenario: Completed todo order is unaffected

- **WHEN** 用户查看 `已完成` 列表
- **THEN** 系统不按 `未执行` 或 `执行中` 的优先级排序规则重排 `已完成` 列表

### Requirement: Expand Or Collapse All Todo Branches

系统 SHALL 允许用户在 `未执行` 和 `执行中` 列表中一键展开或收起当前状态视图里的所有 TODO 分支。批量收起 SHALL 隐藏当前状态视图中所有 TODO 下的项目和终端子项，但 SHALL 保留 TODO 行可见。批量展开 SHALL 显示当前状态视图中所有 TODO 下的项目子项。

#### Scenario: User collapses all visible todo branches

- **WHEN** 当前 `未执行` 或 `执行中` 列表中存在多个已展开 TODO
- **AND** 用户激活全部收起控件
- **THEN** 当前状态视图中的所有 TODO 下的项目和终端子项被隐藏
- **AND** 当前状态视图中的所有 TODO 行仍显示

#### Scenario: User expands all visible todo branches

- **WHEN** 当前 `未执行` 或 `执行中` 列表中存在多个已收起 TODO
- **AND** 用户激活全部展开控件
- **THEN** 当前状态视图中的所有 TODO 下的项目子项被显示

#### Scenario: Active context expands after bulk collapse

- **WHEN** 用户已批量收起当前状态视图中的所有 TODO 分支
- **AND** 当前 TODO、TODO 项目或终端上下文切换到某个已收起 TODO 下
- **THEN** 系统自动展开该 TODO 分支
- **AND** 其他已收起 TODO 分支保持收起
