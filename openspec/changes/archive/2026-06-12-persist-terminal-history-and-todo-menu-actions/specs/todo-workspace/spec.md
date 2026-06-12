## MODIFIED Requirements

### Requirement: Use Todo Context Menu

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行上提供右键菜单，并 SHALL 在 TODO item 上提供一个三点图标按钮用于打开同一组菜单动作。TODO 菜单 SHALL 包含查看详情、添加项目、复制标题和描述、删除 TODO 动作。菜单动作完成、用户点击菜单外部或打开其他侧边栏浮层后，系统 SHALL 关闭 TODO 菜单。

#### Scenario: User opens todo context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **THEN** 系统显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: User opens todo menu from item action button

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击 TODO `修复登录问题` item 上的三点图标按钮
- **THEN** 系统显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

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
