## MODIFIED Requirements

### Requirement: Manage Todo Action Confirmation Popovers

系统 SHALL 在 TODO item 的完成按钮和右键菜单删除操作上使用确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭 TODO 右键菜单、终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。删除确认气泡 SHALL 锚定在对应 TODO item 的操作上下文附近，且 SHALL NOT 脱离当前 TODO item 显示到侧边栏右上角或应用全局右上角。

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
- **AND** 删除确认气泡不显示在应用窗口全局右上角

#### Scenario: Windows delete confirmation popover remains near trigger context

- **WHEN** 应用运行在 Windows
- **AND** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户通过右键菜单或三点菜单选择删除 TODO
- **THEN** 系统在 TODO `修复登录问题` 的操作上下文附近显示删除确认气泡
- **AND** 删除确认气泡不锚定到侧边栏或应用窗口的全局角落

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

### Requirement: Use Todo Context Menu

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行上提供右键菜单，并 SHALL 在 TODO item 上提供一个三点图标按钮用于打开同一组菜单动作。TODO 菜单 SHALL 包含查看详情、添加项目、复制标题和描述、删除 TODO 动作。菜单 SHALL 显示在触发它的指针或三点按钮附近，且 SHALL NOT 显示在应用窗口全局右上角。菜单动作完成、用户点击菜单外部或打开其他侧边栏浮层后，系统 SHALL 关闭 TODO 菜单。

#### Scenario: User opens todo context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **THEN** 系统在指针附近显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: User opens todo menu from item action button

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击 TODO `修复登录问题` item 上的三点图标按钮
- **THEN** 系统在三点图标按钮附近显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: Windows todo menu does not appear in global corner

- **WHEN** 应用运行在 Windows
- **AND** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户通过右键或三点图标按钮打开 TODO 菜单
- **THEN** TODO 菜单显示在触发位置附近
- **AND** TODO 菜单不显示在应用窗口全局右上角

#### Scenario: User opens todo detail from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择查看详情
- **THEN** 系统打开 TODO `修复登录问题` 的详情编辑界面
- **AND** 系统关闭 TODO 菜单

#### Scenario: User opens add project dialog from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择添加项目
- **THEN** 系统打开 TODO `修复登录问题` 的添加项目弹窗
- **AND** 系统关闭 TODO 菜单

#### Scenario: User copies todo title and description from context menu

- **WHEN** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择复制标题和描述
- **THEN** 系统将 `修复登录问题` 和 `登录后跳回首页` 写入系统剪贴板
- **AND** 剪贴板内容第一行为 `修复登录问题`
- **AND** 剪贴板内容第二行开始为 `登录后跳回首页`
- **AND** 系统关闭 TODO 菜单

#### Scenario: Empty todo description copies title only

- **WHEN** TODO `修复登录问题` 的描述为空
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择复制标题和描述
- **THEN** 系统将 `修复登录问题` 写入系统剪贴板
- **AND** 剪贴板内容不包含额外空白描述行

#### Scenario: Outside click closes todo context menu

- **WHEN** TODO `修复登录问题` 的 TODO 菜单已显示
- **AND** 用户点击 TODO 菜单外部
- **THEN** 系统关闭 TODO 菜单
- **AND** 系统不执行菜单动作
