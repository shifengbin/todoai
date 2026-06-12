## Why

TODO 右键菜单中选择删除后，删除确认气泡会出现在侧栏右上角，而不是出现在当前 TODO 操作附近，容易让用户误判当前要删除的对象。这个问题影响删除前确认的可用性，需要在不改变删除语义的前提下修正气泡定位。

## What Changes

- 修正 TODO 删除确认气泡的渲染锚点，使其显示在触发删除动作的 TODO 操作上下文附近。
- 保持现有删除确认流程：打开确认气泡时不立即删除，确认后删除，取消或点击外部后关闭。
- 保持完成 TODO 确认气泡、TODO 项目移除气泡、项目删除气泡和批量项目删除气泡的既有行为。
- 补充前端测试覆盖 TODO 删除确认气泡的定位容器/样式约束，防止回归。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: 明确 TODO 删除确认气泡应锚定在对应 TODO 操作上下文附近，而不是脱离当前 TODO 行显示。

## Impact

- 影响前端侧边栏组件：`frontend/src/components/ProjectSidebar.vue`。
- 影响侧边栏气泡相关样式：`frontend/src/style.css`。
- 影响前端组件/应用测试：`frontend/src/components/ProjectSidebar.test.js` 和必要时的 `frontend/src/App.test.js`。
- 不涉及 Go 后端、Wails API、持久化数据结构或依赖变更。
