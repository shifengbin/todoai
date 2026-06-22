## Why

用户在完成 TODO 后只能看到完成时间，无法判断从点击开始执行到完成一共耗时多久。展示持续时长可以帮助用户回顾任务投入、比较任务规模，并让已完成列表的信息更完整。

## What Changes

- 为 TODO 工作流记录开始执行时间：当用户将 TODO 从 `not-started` 标记为 `in-progress` 时保存开始时间。
- 在 TODO 完成后保留开始时间与完成时间，并计算二者之间的总持续时长。
- 在 `已完成` 视图的 TODO 条目中展示总持续时长。
- 对缺少开始时间的历史已完成 TODO 使用降级展示，避免展示不准确的时长。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: TODO 工作流需要记录执行开始时间，并在已完成 TODO 中展示从开始到完成的总持续时长。

## Impact

- 后端 TODO 数据模型和持久化 JSON 需要新增可选开始时间字段。
- TODO 状态切换逻辑需要在进入 `in-progress` 时写入开始时间。
- Wails 生成的前端模型需要包含新增字段。
- 前端 TODO 侧边栏需要格式化并展示完成耗时。
- 测试需要覆盖开始时间记录、历史数据降级展示和已完成列表渲染。
