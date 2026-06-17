## Context

当前终端运行状态由前端 `agentStatus` 与 `activityState` 派生，`busy` 和 `needs-input` 会展示在终端行，并在 TODO 折叠后聚合成整行呼吸反馈。`done`、`failed`、`exited` 等 phase 目前不会继续作为活动态展示，因此后台终端完成后会直接回到空闲视觉状态。

本次变化需要表达“后台终端刚结束但用户还没有查看”的 UI 已读语义。这个语义依赖当前激活终端和用户选择行为，不是 agent 自身状态，因此不适合扩展后端 agent phase。

## Goals / Non-Goals

**Goals:**

- 非当前激活终端从忙碌切换为空闲、完成、失败或退出时，进入待确认 UI 状态。
- 用户点击或选择对应终端后清除待确认状态。
- 终端行和折叠 TODO 聚合反馈都能展示待确认态。
- 待确认态的视觉表达与 `busy`、`needs-input` 清晰区分。
- 保持现有后端事件、agent phase 优先级和终端历史结构稳定。

**Non-Goals:**

- 不新增后端 `TerminalAgentStatusEvent.Phase`。
- 不持久化待确认态到 workspace 或终端历史。
- 不改变 Claude、Codex、shell、title fallback 的状态来源优先级。
- 不改变展开 TODO 时父 TODO 不重复展示子终端活动态的现有规则。

## Decisions

### 待确认态作为前端 UI 派生态

在 `App.vue` 中维护按 `terminalId` 记录的待确认集合，或在传给侧边栏的终端对象上派生 `attentionState: 'needs-ack'`。每次应用终端 agent/shell/title/command 事件时，比较事件前后的可视活动状态：如果前一状态是 `busy`，后一状态不再是 `busy` 或 `needs-input`，且终端不是当前激活终端，则标记该终端待确认。

选择原因：

- 待确认态依赖“是否为后台终端”和“是否被用户选择过”，属于 UI 已读语义。
- 保持 `agentStatus.phase` 只描述 agent/runtime 状态，避免后端和前端状态机混合。
- 不需要迁移持久化数据或更新 Wails Go 绑定模型。

备选方案是新增 `agentStatus.phase = 'needs-ack'`。该方案会让后端事件模型表达 UI 已读状态，并影响 phase 优先级、历史记录和跨来源状态替换规则，因此不采用。

### 确认态清除点使用终端选择

用户触发 `selectTerminal(terminalId)` 时清除该终端待确认态。清除可以发生在调用后端 `SelectTerminal` 前，保证即使后端状态刷新带回当前终端，UI 也不会短暂残留确认态。

选择原因：

- 用户点击对应终端即代表“已查看”。
- 与现有终端选择路径一致，不需要新增后端 API。

### TODO 聚合优先级扩展为 needs-input > needs-ack > busy > idle

`ProjectSidebar.vue` 的终端可视状态计算需要识别待确认态。折叠 TODO 聚合时，`needs-input` 仍最高优先级，其次是确认态，再其次是忙碌态，最后是空闲态。

选择原因：

- `needs-input` 代表当前需要用户操作，必须压过完成提醒。
- 确认态代表后台结果待查看，应压过仍在运行的普通提示，以免完成提醒被忙碌态淹没。
- 保留现有 `busy` 和 `idle` 语义。

### 视觉表达沿用现有组件和样式层

终端行使用 lucide `TriangleAlert` 图标展示待确认态，并增加独立的终端行状态 class。折叠 TODO 使用新的 `todo-activity-needs-ack` class 和独立 keyframes，动画周期比 busy 更短，颜色与 busy 的绿色、needs-input 的橙色区分。`prefers-reduced-motion: reduce` 下取消动画并保留静态状态色。

选择原因：

- 与现有 `LoaderCircle`、`CircleAlert` 图标模式一致。
- 避免在 TODO 折叠行复用终端行图标，符合现有 spec 对折叠 TODO 的约束。

## Risks / Trade-offs

- 待确认态只在前端内存中维护，刷新或重新打开 workspace 后会丢失。→ 这是有意取舍；该状态表示当前会话内的查看提醒，不作为持久任务状态。
- 通过前后 `activityState` 比较触发确认态时，必须避免 `needs-input -> idle` 也触发。→ 只接受前一可视状态为 `busy` 的转换。
- `done` 或 `failed` phase 目前会映射为 `idle` 可视态。→ 确认态通过前一状态捕捉完成瞬间，不改变现有 idle 映射。
- title fallback 会在 1 秒无标题变化后从 busy 回 idle，可能触发确认态。→ 这是符合需求的后台忙碌结束提示；当前激活终端不触发。
