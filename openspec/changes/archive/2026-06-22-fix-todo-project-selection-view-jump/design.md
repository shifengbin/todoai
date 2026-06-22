## Context

TODO 工作区把 `currentTodoView` 作为当前视图标签，同时按 `activeTodoProjectId` 保存每个 TODO 工程上次选择的视图标签。后端在选择终端时会同步选择终端所属的 TODO 工程，因此前端 active TODO 工程可能因为终端操作而改变。

当前前端在用户点击 TODO 项目行时调用 `selectTodoProject()`，并要求 `applyState()` 恢复该 TODO 工程保存过的视图标签。这个规则在“进入 workspace 或重新加载时恢复上次上下文”是合理的，但在用户已经手动切到 `未执行` 后点击当前列表里的项目 item 时，会把项目选择误变成视图切换。

## Goals / Non-Goals

**Goals:**

- 用户点击 TODO 项目 item 时保持当前 TODO 视图标签。
- 保留 workspace 打开、重新打开和前端重新加载时的 TODO 工程视图恢复能力。
- 保持选择终端后 active TODO 工程、active 项目、active terminal 和终端聚焦的现有行为。
- 用前端回归测试覆盖“执行中选择终端 -> 切到未执行 -> 点击项目 item 不回跳执行中”。

**Non-Goals:**

- 不修改后端 `SelectTerminal` 或 `SelectTodoProject` 的业务语义。
- 不改变 TODO 视图标签的持久化格式。
- 不改变左侧栏宽度的 workspace 级持久化规则。
- 不改变 TODO 工作流状态和排序规则。

## Decisions

### 普通项目点击不触发视图恢复

选择：前端 `selectTodoProject(todoProjectId)` 仍调用后端选择 TODO 工程，但不再以“恢复 TODO 工程 UI 状态”的方式应用返回状态。当前 `currentTodoView` 由用户显式点击 tab、workspace 加载/重载恢复路径或无状态默认值控制。

原因：点击项目 item 的用户意图是选择工作上下文，不是切换 TODO 状态视图。当前视图里展示的是用户刚选择的状态列表，点击其中 item 不应让列表切走。

备选方案：继续恢复项目保存视图，但在终端选择后特殊清理保存状态。该方案会把问题转移到状态持久化层，且会破坏“每个 TODO 工程记住上次视图”的历史数据。

### 恢复路径保持显式

选择：保留打开 workspace、重新打开 workspace、前端重新加载等路径中的 `restoreTodoProjectUIState` 逻辑，只收窄项目 item 点击路径。

原因：上次视图恢复是启动/加载上下文的一部分；普通点击项目 item 是当前会话内的导航动作。二者用户意图不同，前端已经知道触发来源，不需要扩展后端 API。

备选方案：让后端返回状态变更来源，由前端根据来源决定是否恢复。该方案扩大协议和 Wails 绑定变更范围，对当前 bug 没有必要。

## Risks / Trade-offs

- [Risk] 现有测试断言“点击 TODO 工程恢复保存视图”会失效。-> Mitigation: 将该测试改为加载/重开恢复场景，并新增点击项目保持当前视图的回归测试。
- [Risk] 用户过去依赖点击项目快速跳到该项目保存视图。-> Mitigation: 保留 workspace 加载恢复；当前会话内切 tab 仍会保存到 active TODO 工程，后续加载仍可恢复。
- [Risk] active TODO 工程来自终端选择，切 tab 保存到终端所属工程的行为仍存在。-> Mitigation: 本变更只阻止普通项目点击覆盖当前视图，避免用户可见回跳；不扩大到后端 active context 语义调整。
