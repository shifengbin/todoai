## Why

当前 TODO 工作区会在添加终端或切换终端时把当前视图切回已保存的 `未执行` 状态，并可能改变左侧 TODO 列表宽度。这会把终端操作误当作用户主动切换 TODO 工程，打断当前工作上下文。

左侧 TODO 列表宽度是 workspace 级布局偏好，不属于某个 TODO item、TODO 工程或终端。需要明确 UI 状态归属，并限制只有明确的恢复场景才应用持久化状态。

## What Changes

- 将左侧 TODO 列表宽度从 TODO 工程 UI 状态中拆出，定义为当前 workspace 的 TODO 工作区布局状态。
- 保留 TODO 视图标签按 TODO 工程保存和恢复的语义。
- 添加终端、切换终端和终端状态刷新 SHALL NOT 改变当前 TODO 视图标签。
- 添加终端、切换终端和终端状态刷新 SHALL NOT 改变左侧 TODO 列表宽度。
- 用户主动选择 TODO 工程时可以恢复该 TODO 工程的 TODO 视图标签，但不恢复左侧 TODO 列表宽度。
- 打开 workspace、重新打开 workspace 或前端重新加载时恢复当前 workspace 的左侧 TODO 列表宽度。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 修正 TODO 工作区 UI 状态归属和恢复触发条件，区分 TODO 工程视图状态与 workspace 级左侧栏宽度状态。

## Impact

- 前端状态管理：`frontend/src/App.vue` 中 `applyState`、TODO 工程 UI 状态恢复、侧栏宽度保存逻辑。
- 前端测试：TODO 视图持久化、侧栏宽度持久化、添加终端、切换终端相关用例。
- 后端 UI 状态模型和 Wails API：可能需要迁移或兼容现有 `TodoProjectUIState.sidebarWidth` 存储。
- OpenSpec：更新 `todo-workspace` 规范中“Persist Todo Project UI State”相关需求和场景。
