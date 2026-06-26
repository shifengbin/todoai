## Why

当前 TODO 工作区侧边栏中，`未执行`、`执行中`、`已完成` 三个工作流 tab 位于可滚动内容内部。TODO 列表较长时，用户向下滚动后无法直接切换状态视图，需要先滚回顶部，影响长列表操作效率。

## What Changes

- TODO 工作区侧边栏 SHALL 在列表滚动时保持 `未执行`、`执行中`、`已完成` 三个工作流 tab 可见且可点击。
- 三个工作流 tab SHALL 位于 TODO 列表滚动区域之外，不随 TODO 内容滚动。
- TODO 列表内容 SHALL 在工作流 tab 下方的独立滚动区域内滚动。
- 侧边栏 header 已位于滚动区域之外，本变更不改变 header 行行为。
- 本变更不引入后端 API 或数据模型变化。

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `todo-workspace`: 收紧 TODO 工作区滚动要求，明确工作流状态 tab 在 TODO 列表滚动时保持固定可见。

## Impact

- Frontend component layout in `frontend/src/components/ProjectSidebar.vue`.
- Frontend CSS in `frontend/src/style.css`.
- Frontend tests in `frontend/src/components/ProjectSidebar.test.js`.
- No backend changes.
