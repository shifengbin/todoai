## MODIFIED Requirements

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供已完成查看功能。已完成列表 SHALL 只显示状态为 `completed` 的 TODO，并 SHALL 按完成时间倒序展示，最近完成的 TODO 排在前面。完成时间 SHALL 优先使用 `completedAt`，当 `completedAt` 缺失时 SHALL 使用 `archivedAt` 作为兼容旧数据的兜底。缺失有效完成时间的已完成 TODO SHALL 排在有完成时间的 TODO 之后。已完成列表 SHALL 展示完成时保存的项目快照。已删除 TODO SHALL NOT 显示在已完成列表中。

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 已完成
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示该 TODO 的完成时间
- **AND** `已完成` 视图显示该 TODO 完成时关联项目的名称和路径快照

#### Scenario: Completed todos are ordered by newest completion time

- **WHEN** `已完成` 视图包含 TODO `整理文档`
- **AND** TODO `整理文档` 的 `completedAt` 为 `2026-06-14T09:00:00Z`
- **AND** `已完成` 视图包含 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `修复登录问题` 排在 TODO `整理文档` 前面

#### Scenario: Completed todo order falls back to archivedAt

- **WHEN** `已完成` 视图包含 TODO `旧任务`
- **AND** TODO `旧任务` 不包含 `completedAt`
- **AND** TODO `旧任务` 的 `archivedAt` 为 `2026-06-15T10:00:00Z`
- **AND** `已完成` 视图包含 TODO `较早任务`
- **AND** TODO `较早任务` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `旧任务` 排在 TODO `较早任务` 前面

#### Scenario: Completed todo without completion time is ordered last

- **WHEN** `已完成` 视图包含 TODO `缺失时间任务`
- **AND** TODO `缺失时间任务` 不包含有效的 `completedAt` 或 `archivedAt`
- **AND** `已完成` 视图包含 TODO `有时间任务`
- **AND** TODO `有时间任务` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `有时间任务` 排在 TODO `缺失时间任务` 前面

#### Scenario: Deleted todo is not shown as completed

- **WHEN** TODO `废弃任务` 已被删除
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图不显示 TODO `废弃任务`

#### Scenario: Completed todo does not restore terminals

- **WHEN** 用户打开 `已完成` 视图
- **AND** 用户查看 TODO `修复登录问题`
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程
