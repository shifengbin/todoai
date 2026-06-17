## Why

当前 TODO 下可能同时运行多个后台终端，但后台终端从忙碌恢复为空闲后不会留下明显提示。用户需要一个中间确认态来区分“已经空闲且已查看”和“后台刚结束但还没看过”，避免错过后台任务完成结果。

## What Changes

- 为非当前激活终端增加“后台完成待确认”UI 状态：当终端从 `busy` 切换到 `idle`、`done`、`failed` 或 `exited`，且该终端不是当前激活终端时进入确认态。
- 用户点击或选择对应终端后清除确认态，并按终端当前运行状态显示为空闲、结束或其它状态。
- 终端行使用三角感叹号展示确认态，颜色与忙碌态、等待输入态区分。
- 折叠 TODO 下若存在确认态终端，则 TODO 行展示确认态聚合反馈；聚合优先级为 `needs-input` 高于确认态，高于 `busy`，高于 `idle`。
- 折叠 TODO 的确认态反馈使用急促呼吸效果，颜色与忙碌态区分。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `agent-status`: 终端活动状态需要派生后台完成待确认 UI 状态，并在选择终端时确认。
- `todo-workspace`: 折叠 TODO 的隐藏终端聚合反馈需要包含确认态，并定义与 `needs-input`、`busy`、`idle` 的优先级和视觉表达。

## Impact

- 前端 `App.vue` 需要在终端活动状态变化时记录和清除确认态，并将确认态传递给侧边栏展示。
- 前端 `ProjectSidebar.vue` 需要识别终端确认态、展示三角感叹号，并将确认态纳入折叠 TODO 聚合状态。
- 前端样式需要新增终端确认态颜色和折叠 TODO 确认态急促呼吸动画，同时保留 reduced-motion 降级。
- 前端测试需要覆盖后台 busy-to-idle/done/failed/exited 触发确认态、当前激活终端不触发、点击终端清除、折叠 TODO 聚合优先级和样式规则。
