## Why

创建 TODO 或切换 TODO 工作流状态会触发前端重新应用后端 ProjectState，但当前实现会顺带恢复 TODO 工程 UI 状态，导致用户刚调整的左侧宽度被重置、点击开始后又回到 `未执行` 视图。需要明确 UI 状态恢复只应发生在用户主动调整或打开/切换上下文时，普通业务刷新不应覆盖当前界面选择。

## What Changes

- 限制 TODO 工程 UI 状态恢复的触发场景：仅在打开 workspace、前端重新加载或用户主动选择 TODO 工程时恢复已保存的 TODO 视图和侧栏宽度。
- 普通业务状态刷新 SHALL NOT 重置左侧 TODO 栏宽度，包括创建 TODO、编辑 TODO、关联项目、删除 TODO、完成 TODO、终端状态刷新等。
- 用户将 `not-started` TODO 标记为 `in-progress` 后，TODO 工作区 SHALL 自动停留在 `执行中` 视图，并显示该 TODO。
- 保留现有持久化能力：用户手动切换 TODO 视图或拖动侧栏宽度后仍按 TODO 工程保存，后续打开/选择该 TODO 工程时恢复。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: 明确 TODO 工程 UI 状态恢复的触发边界，并要求开始 TODO 后自动展示 `执行中` 视图。

## Impact

- 前端状态同步：`frontend/src/App.vue` 中 `applyState` 与 TODO 工程 UI 状态应用逻辑。
- TODO 侧栏交互：`frontend/src/components/ProjectSidebar.vue` 的状态视图切换和开始 TODO 行为。
- 测试覆盖：`frontend/src/App.test.js` 需要覆盖创建 TODO 后侧栏宽度保持、开始 TODO 后停留在 `执行中` 视图、打开/切换 TODO 工程仍恢复 UI 状态。
- 后端 ProjectState 和持久化数据结构不需要变更。
