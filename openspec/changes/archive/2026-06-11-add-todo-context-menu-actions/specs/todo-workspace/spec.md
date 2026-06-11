## ADDED Requirements

### Requirement: Manage Todo Dialog Dismissal

系统 SHALL 防止 TODO 创建、TODO 详情编辑和 TODO 添加项目弹窗因点击空白遮罩而关闭。这些弹窗 SHALL 仅通过弹窗关闭按钮、取消按钮或成功提交后的流程关闭。

#### Scenario: Create todo dialog ignores backdrop click

- **WHEN** 用户打开创建 TODO 弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** 创建 TODO 弹窗保持打开
- **AND** 用户已输入的 TODO 表单内容保持不变

#### Scenario: Todo detail dialog ignores backdrop click

- **WHEN** 用户打开 TODO `修复登录问题` 的详情编辑弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** TODO 详情编辑弹窗保持打开
- **AND** 用户已输入的详情表单内容保持不变

#### Scenario: Add project dialog ignores backdrop click

- **WHEN** 用户为 TODO `修复登录问题` 打开添加项目弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** 添加项目弹窗保持打开
- **AND** 用户已选择的项目保持选中

#### Scenario: Todo dialog closes from explicit controls

- **WHEN** 用户打开创建 TODO、TODO 详情编辑或添加项目弹窗
- **AND** 用户点击弹窗关闭按钮或取消按钮
- **THEN** 系统关闭该弹窗

### Requirement: Use Todo Context Menu

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行上提供右键菜单。TODO 右键菜单 SHALL 包含查看详情、添加项目、复制描述和删除 TODO 动作。菜单动作完成、用户点击菜单外部或打开其他侧边栏浮层后，系统 SHALL 关闭 TODO 右键菜单。

#### Scenario: User opens todo context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **THEN** 系统显示 TODO 右键菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: User opens todo detail from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **AND** 用户选择查看详情
- **THEN** 系统打开 TODO `修复登录问题` 的详情编辑界面
- **AND** 系统关闭 TODO 右键菜单

#### Scenario: User opens add project dialog from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **AND** 用户选择添加项目
- **THEN** 系统打开 TODO `修复登录问题` 的添加项目弹窗
- **AND** 系统关闭 TODO 右键菜单

#### Scenario: User copies todo description from context menu

- **WHEN** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **AND** 用户选择复制描述
- **THEN** 系统将 `登录后跳回首页` 写入系统剪贴板
- **AND** 系统关闭 TODO 右键菜单

#### Scenario: Empty todo description copies empty text

- **WHEN** TODO `修复登录问题` 的描述为空
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **AND** 用户选择复制描述
- **THEN** 系统将空字符串写入系统剪贴板
- **AND** 系统不把 TODO 标题写入剪贴板

#### Scenario: Outside click closes todo context menu

- **WHEN** TODO `修复登录问题` 的右键菜单已显示
- **AND** 用户点击右键菜单外部
- **THEN** 系统关闭 TODO 右键菜单
- **AND** 系统不执行菜单动作

## MODIFIED Requirements

### Requirement: Restrict Todo Workflow Actions

系统 SHALL 按 TODO 当前工作流状态限制用户可执行的状态动作。`not-started` TODO SHALL 在行外仅暴露开始状态按钮，SHALL NOT 暴露完成入口或退回入口。`in-progress` TODO SHALL 在行外仅暴露完成状态按钮，SHALL NOT 暴露开始入口或退回 `not-started` 的入口。查看/编辑 TODO、添加项目和删除 TODO SHALL 在 `not-started` 与 `in-progress` 状态下通过 TODO 右键菜单保持可用。

#### Scenario: Not-started todo exposes only start as row action

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 `未执行` 视图查看该 TODO
- **THEN** 系统在 TODO 行外显示开始 TODO 的入口
- **AND** 系统不在 TODO 行外显示完成 TODO 的入口
- **AND** 系统不在 TODO 行外显示退回未执行的入口
- **AND** 系统不在 TODO 行外显示查看详情、添加项目或删除 TODO 入口
- **AND** 系统通过 TODO 右键菜单提供查看详情、添加项目和删除 TODO 入口

#### Scenario: In-progress todo exposes only complete as row action

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 `执行中` 视图查看该 TODO
- **THEN** 系统在 TODO 行外显示完成 TODO 的入口
- **AND** 系统不在 TODO 行外显示开始 TODO 的入口
- **AND** 系统不在 TODO 行外显示退回未执行的入口
- **AND** 系统不在 TODO 行外显示查看详情、添加项目或删除 TODO 入口
- **AND** 系统通过 TODO 右键菜单提供查看详情、添加项目和删除 TODO 入口

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

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO 右键菜单中的删除动作和确认气泡删除 `not-started` 或 `in-progress` TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 从可见 TODO 工作区列表中移除。用户取消确认 SHALL 不改变 TODO 或其终端状态。

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

#### Scenario: User cancels todo deletion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Edit Todo Details

系统 SHALL 允许用户从活动 TODO item 的右键菜单进入 TODO 详情，并 SHALL 在详情中查看和编辑 TODO 名称、描述、优先级和关联工程。保存编辑 SHALL 持久化更新后的 TODO 字段和工程关联。

#### Scenario: User opens todo details from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择查看详情
- **THEN** 系统打开 TODO 详情编辑界面
- **AND** 详情界面显示 TODO 名称、描述、优先级和已关联工程

#### Scenario: User edits todo metadata

- **WHEN** 用户打开 TODO `修复登录问题` 的详情
- **AND** 用户将名称改为 `修复登录跳转`
- **AND** 用户将描述改为 `登录后跳回首页`
- **AND** 用户将优先级改为 `高`
- **AND** 用户保存详情
- **THEN** 活动 TODO 列表显示 `修复登录跳转`
- **AND** TODO 描述保存为 `登录后跳回首页`
- **AND** TODO 优先级保存为 `高`

#### Scenario: User edits todo projects without removed terminals

- **WHEN** TODO `修复登录问题` 已关联工程 `frontend-app`
- **AND** 项目库还包含工程 `api-service`
- **AND** 用户在详情中新增工程 `api-service`
- **AND** 用户保存详情
- **THEN** TODO `修复登录问题` 下显示工程 `frontend-app`
- **AND** TODO `修复登录问题` 下显示工程 `api-service`
- **AND** 系统不自动为 `api-service` 创建终端

#### Scenario: Saving detail edit confirms removed project terminals

- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 存在终端
- **AND** 用户在详情中移除工程 `frontend-app`
- **AND** 用户点击保存
- **THEN** 系统在保存前提示移除该工程会关闭其在当前 TODO 下的终端
- **WHEN** 用户确认保存
- **THEN** 系统移除 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 系统关闭并移除 `frontend-app` 在该 TODO 下的终端

#### Scenario: User cancels terminal-closing detail save

- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 存在终端
- **AND** 用户在详情中移除工程 `frontend-app`
- **AND** 用户点击保存
- **AND** 用户取消关闭终端确认
- **THEN** 详情编辑界面保持打开
- **AND** TODO `修复登录问题` 下的工程 `frontend-app` 仍保持关联
- **AND** `frontend-app` 在该 TODO 下的终端保持运行时状态不变

### Requirement: Manage Todo Action Confirmation Popovers

系统 SHALL 在 TODO item 的完成按钮和右键菜单删除操作上使用确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭 TODO 右键菜单、终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。

#### Scenario: Complete confirmation popover opens beside complete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的完成按钮
- **THEN** 系统在完成按钮旁显示完成确认气泡
- **AND** 系统不立即完成 TODO `修复登录问题`

#### Scenario: Delete confirmation popover opens from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择删除 TODO
- **THEN** 系统关闭 TODO 右键菜单
- **AND** 系统显示删除确认气泡
- **AND** 系统不立即删除 TODO `修复登录问题`

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
