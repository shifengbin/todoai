## ADDED Requirements

### Requirement: Preserve Todo Clipboard Text On macOS

系统 SHALL 在 macOS 桌面应用中保持 TODO 表单和 TODO 菜单剪贴板文本的 Unicode 保真。创建 TODO 表单的名称和描述输入框 SHALL 支持系统粘贴动作。TODO 菜单复制标题和描述时 SHALL 将原始 Unicode 文本写入系统剪贴板，MUST NOT 产生 mojibake、替换字符或额外空白行。剪贴板写入失败时，系统 SHALL 显示不会改变当前工作区状态的错误信息。

#### Scenario: Paste Unicode text into create todo form on macOS

- **WHEN** 应用运行在 macOS
- **AND** 用户打开创建 TODO 表单
- **AND** 系统剪贴板包含 `修复登录问题`
- **AND** 用户聚焦 TODO 名称输入框并触发系统粘贴动作
- **THEN** TODO 名称输入框内容为 `修复登录问题`
- **WHEN** 系统剪贴板包含 `登录后跳回首页 🔧`
- **AND** 用户聚焦 TODO 描述输入框并触发系统粘贴动作
- **THEN** TODO 描述输入框内容为 `登录后跳回首页 🔧`

#### Scenario: Copy Unicode todo title and description

- **WHEN** TODO `修复登录问题 ✓` 的描述为 `登录后跳回首页 🔧 `
- **AND** 用户在 TODO 行上打开 TODO 菜单
- **AND** 用户选择复制标题和描述
- **THEN** 系统剪贴板内容第一行为 `修复登录问题 ✓`
- **AND** 系统剪贴板内容第二行为 `登录后跳回首页 🔧 `
- **AND** 系统剪贴板内容不包含 mojibake 或替换字符
- **AND** 系统关闭 TODO 菜单

#### Scenario: Todo clipboard copy failure is reported

- **WHEN** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** 用户选择复制标题和描述
- **AND** 系统剪贴板写入失败
- **THEN** 系统显示错误信息
- **AND** 当前 TODO 列表和 TODO 详情数据保持不变
