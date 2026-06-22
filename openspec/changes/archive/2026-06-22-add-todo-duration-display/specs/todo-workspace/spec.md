## ADDED Requirements

### Requirement: Display Todo Completion Duration

系统 SHALL 在用户将 TODO 从 `not-started` 标记为 `in-progress` 时记录该 TODO 的开始执行时间。系统 SHALL 在 TODO 完成后根据开始执行时间和完成时间计算总持续时长，并 SHALL 在 `已完成` 视图中展示该持续时长。系统 MUST NOT 使用 TODO 创建时间推断开始执行时间。若 completed TODO 缺少有效开始执行时间、缺少有效完成时间或完成时间早于开始执行时间，系统 SHALL 不展示持续时长或 SHALL 使用明确的缺失占位。

#### Scenario: Start time is recorded when todo enters in-progress

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **THEN** TODO `修复登录问题` 的状态保存为 `in-progress`
- **AND** 系统记录 TODO `修复登录问题` 的开始执行时间

#### Scenario: Completed todo shows duration from start to completion

- **WHEN** TODO `修复登录问题` 的开始执行时间为 `2026-06-22T01:00:00Z`
- **AND** TODO `修复登录问题` 的完成时间为 `2026-06-22T02:15:30Z`
- **AND** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示 TODO `修复登录问题` 的总持续时长为 1 小时 15 分 30 秒

#### Scenario: Historical completed todo without start time does not show inferred duration

- **WHEN** TODO `历史任务` 的状态为 `completed`
- **AND** TODO `历史任务` 包含有效完成时间
- **AND** TODO `历史任务` 不包含有效开始执行时间
- **AND** 用户打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `历史任务`
- **AND** 系统不使用 TODO `历史任务` 的创建时间推断持续时长
- **AND** `已完成` 视图不展示 TODO `历史任务` 的持续时长或显示明确的缺失占位

#### Scenario: Invalid duration is hidden

- **WHEN** TODO `异常任务` 的状态为 `completed`
- **AND** TODO `异常任务` 的开始执行时间晚于完成时间
- **AND** 用户打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `异常任务`
- **AND** `已完成` 视图不展示负数持续时长
