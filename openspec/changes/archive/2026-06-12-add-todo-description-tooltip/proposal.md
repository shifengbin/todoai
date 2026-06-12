## Why

TODO 列表已经显示描述摘要，但当描述较长时用户仍需要打开详情才能读完整内容。为减少查看成本，TODO 行应在不离开列表上下文的情况下提供完整描述预览。

## What Changes

- TODO 行继续保留现有的一行描述摘要，方便快速扫视列表。
- 当 TODO 存在描述时，鼠标悬浮在 TODO 行上一段时间后，系统 SHALL 显示完整描述 tooltip。
- 鼠标移开 TODO 行后，描述 tooltip SHALL 消失。
- 无描述的 TODO 不显示描述 tooltip。
- tooltip SHALL 支持较长描述的多行阅读，且不改变 TODO 标题、项目数量、操作按钮和上下文菜单的既有行为。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 修改 TODO 工作树中 TODO 描述的浏览要求，增加悬浮延迟显示完整描述的行为。

## Impact

- `frontend/src/components/ProjectSidebar.vue` 需要增加 TODO 描述 tooltip 的悬浮状态、延迟显示和移开隐藏逻辑。
- `frontend/src/style.css` 需要增加 tooltip 样式，并确保在浅色和深色主题下可读。
- `frontend/src/components/ProjectSidebar.test.js` 需要更新描述展示测试，覆盖默认隐藏 tooltip、延迟显示和移开隐藏。
- 不影响 Go 后端、Wails API、数据结构或持久化格式。
