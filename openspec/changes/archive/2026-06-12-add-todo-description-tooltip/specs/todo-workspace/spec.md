## ADDED Requirements

### Requirement: Preview Todo Description On Hover

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行中保留描述摘要，并 SHALL 允许用户通过鼠标悬浮查看完整 TODO 描述。仅当 TODO 存在非空描述时，系统 SHALL 在鼠标悬浮于 TODO 行一段时间后显示完整描述 tooltip；鼠标移开后 tooltip SHALL 消失。tooltip SHALL 不要求用户打开 TODO 详情，且 SHALL 不改变 TODO 行的展开收起、菜单和状态操作行为。

#### Scenario: Todo row shows description summary

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页，需要保留原始跳转地址`
- **THEN** TODO 行显示该描述的摘要

#### Scenario: User previews full todo description after hover delay

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页，需要保留原始跳转地址`
- **AND** 用户将鼠标悬浮在 TODO `修复登录问题` 行上但尚未达到显示延迟
- **THEN** 系统不显示完整描述 tooltip
- **WHEN** 悬浮时间达到显示延迟
- **THEN** 系统显示包含完整描述 `登录后跳回首页，需要保留原始跳转地址` 的 tooltip

#### Scenario: Todo description tooltip hides on mouse leave

- **WHEN** TODO `修复登录问题` 的完整描述 tooltip 已显示
- **AND** 用户将鼠标移出 TODO `修复登录问题` 行
- **THEN** 系统隐藏该 tooltip

#### Scenario: Todo without description has no tooltip

- **WHEN** 活动 TODO 列表显示 TODO `整理文档`
- **AND** TODO `整理文档` 的描述为空
- **AND** 用户将鼠标悬浮在 TODO `整理文档` 行上
- **THEN** 系统不显示描述 tooltip
