## Context

当前 TODO 数据模型只保存 `createdAt`、`completedAt` 和 `archivedAt`。用户点击开始时，后端 `ChangeTodoStatus` 只将状态从 `not-started` 改为 `in-progress`，没有记录开始执行时间。用户点击完成时，后端 `CompleteTodo` 会写入完成时间并将 TODO 移入 `已完成` 视图。

因为 `createdAt` 表示 TODO 创建时间，不等于用户开始执行时间，所以不能用它计算执行耗时。要准确展示“从点击开始到完成”的总时长，需要在进入 `in-progress` 时持久化开始时间。

## Goals / Non-Goals

**Goals:**

- 记录用户将 TODO 标记为 `in-progress` 的时间。
- 完成 TODO 后，根据开始时间和完成时间计算总持续时长。
- 在 `已完成` 视图中展示该持续时长。
- 对历史 TODO 或异常数据降级处理，不展示推断出的错误时长。

**Non-Goals:**

- 不统计终端运行时长、Agent 忙碌时长或暂停时间。
- 不提供手动修改开始时间或完成时间的能力。
- 不改变 TODO 状态机，不新增从 `in-progress` 回退到 `not-started` 的能力。
- 不重新计算或迁移历史 completed TODO 的开始时间。

## Decisions

### 在后端 TODO 模型中新增 `startedAt`

在 Go `Todo` 结构体中新增可选字段 `StartedAt string json:"startedAt,omitempty"`，并让 Wails 前端模型同步包含 `startedAt`。

理由：开始时间是业务事实，必须随 workspace TODO 数据持久化。前端本地计算或临时保存会在重启、刷新或跨窗口状态同步时丢失。

备选方案是复用 `createdAt`。该方案实现更少，但语义错误：TODO 可能创建很久后才开始执行，会显著夸大耗时。

### 在 `ChangeTodoStatus` 进入 `in-progress` 时写入开始时间

当 TODO 从 `not-started` 成功切换为 `in-progress` 时，后端写入当前 UTC RFC3339 时间到 `startedAt`。现有状态转换只允许 `not-started -> in-progress`，因此不需要处理重复开始覆盖旧值。

理由：点击开始是用户定义的计时起点，写入点应与状态转换保持原子性。

备选方案是在前端点击按钮时先生成时间再传给后端。该方案会让客户端时间成为数据来源，且增加 API 参数和时钟一致性问题。

### 完成耗时由展示层按持久化时间点计算

后端持久化 `startedAt` 与 `completedAt`，前端在已完成列表中解析两个时间点并格式化为简短时长。格式建议按最大有效单位展示，例如 `45s`、`12m 05s`、`2h 03m`、`1d 04h`。

理由：时长是派生值，不需要额外持久化，避免开始时间或完成时间变更后出现不一致。

备选方案是完成时保存 `durationMs`。该方案读取简单，但会增加冗余字段和迁移复杂度；当前需求没有需要固定快照值的场景。

### 历史数据不推断耗时

如果 completed TODO 缺少有效 `startedAt` 或有效完成时间，已完成列表不展示耗时。系统不得用 `createdAt`、`archivedAt` 或当前时间推断开始时间；`archivedAt` 仅继续作为既有完成时间排序和完成时间展示的兼容兜底。

理由：错误耗时比缺失耗时更误导用户。

## Risks / Trade-offs

- 历史 completed TODO 没有持续时长 → 在 UI 中降级为空或占位，避免展示错误信息。
- 系统时钟被修改导致 `completedAt < startedAt` → 前端不展示负数时长，并可在测试中覆盖。
- Wails 模型未重新生成会导致前端类型缺字段 → 实现任务中明确包含重新生成或同步 `frontend/wailsjs/go/models.ts`。
- 前端布局空间有限 → 将持续时长放在已完成 TODO 的 meta 行中，使用短标签，避免挤占标题和菜单。

## Migration Plan

新增字段使用 `omitempty`，旧 workspace JSON 可直接读取。旧 TODO 不包含 `startedAt` 时保持原样；只有新开始的 TODO 会写入该字段。回滚时旧版本会忽略未知 JSON 字段，不需要数据迁移。

## Open Questions

无。当前需求按“点击开始到完成”的自然语义实现，不包含暂停、恢复或终端活动统计。
