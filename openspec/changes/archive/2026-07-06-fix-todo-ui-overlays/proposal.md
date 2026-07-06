## Why

当前 Todo 相关界面存在两个影响日常操作的 UI 问题：手动输入分支后分支候选下拉框不会在点击其他位置时关闭，任务列表出现滚动条后会遮盖右侧操作按钮。它们会让创建/维护 Todo 的操作反馈不清晰，并降低任务列表按钮的可点击性。

## What Changes

- 分支候选下拉框在用户完成手动输入并离开输入区域后应关闭，同时保留已输入的自定义分支值。
- 分支候选下拉框的关闭行为应覆盖创建 Todo、编辑 Todo、给 Todo 添加项目这三处共用分支选择交互。
- 任务列表滚动区域应为滚动条预留稳定空间，避免遮盖 Todo 行右侧操作按钮。
- 增加前端测试覆盖分支下拉关闭行为和任务列表滚动条避让样式。

## Capabilities

### New Capabilities
- `todo-ui-overlays`: 约束 Todo 相关浮层、分支选择器和可滚动任务列表操作区的交互与布局行为。

### Modified Capabilities
- 无。

## Impact

- 影响前端 Vue 组件与样式：
  - `tui-helper/frontend/src/App.vue`
  - `tui-helper/frontend/src/style.css`
- 影响前端测试：
  - `tui-helper/frontend/src/App.test.js`
  - `tui-helper/frontend/src/components/ProjectSidebar.test.js`
- 不涉及后端 API、数据结构或依赖变更。
