## Why

已完成 TODO 目前按原始列表顺序展示，用户很难快速回看最近刚完成的任务。把已完成列表按完成时间倒序排列，可以让最新完成的工作自然出现在最上方。

## What Changes

- `已完成` 视图中的 TODO SHALL 按完成时间倒序展示。
- 完成时间优先使用 `completedAt`，缺失时使用兼容旧数据的 `archivedAt`。
- 缺失有效完成时间的已完成 TODO SHALL 排在有完成时间的 TODO 之后。
- `未执行` 和 `执行中` 视图的现有排序规则不变。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 已完成 TODO 列表的展示顺序从持久化/输入顺序变为按完成时间倒序。

## Impact

- 前端：`frontend/src/components/ProjectSidebar.vue` 的已完成 TODO 列表计算逻辑。
- 测试：`frontend/src/components/ProjectSidebar.test.js` 增加已完成列表排序覆盖。
- 数据：不改变现有持久化结构，不引入新字段或迁移；继续使用已有 `completedAt` / `archivedAt`。
- API：不改变 Wails Go API 或前后端模型契约。
