## ADDED Requirements

### Requirement: List Completed Todos From Project Directory

系统 SHALL 提供 `todoai list --done` 命令，用于从命令行列出当前项目相关的已完成 TODO。该命令 MUST 可在已登记项目根目录或其任意子目录执行。系统 SHALL 根据当前工作目录定位 TodoAI 已知项目，并 SHALL 只返回与该项目匹配且状态为 `completed` 的 TODO。成功时，stdout SHALL 输出 JSON 数组；每个数组元素 SHALL 包含 `taskName`、`worktreeBranch`、`baseBranch`。若历史项目快照缺少 worktree 分支或 base 分支，系统 SHALL 保留该 TODO 并用 `-` 显示缺失字段。该命令 SHALL 在执行时不启动 Wails GUI。

#### Scenario: List completed todos from project root

- **WHEN** 当前工作目录为已登记项目 `frontend-app` 的根目录
- **AND** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的完成时项目快照匹配项目 `frontend-app`
- **AND** 该项目快照的 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照的 base 分支为 `main`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout 为 JSON 数组
- **AND** JSON 数组包含 `taskName` 为 `修复登录问题` 的元素
- **AND** 该元素的 `worktreeBranch` 为 `todo/fix-login`
- **AND** 该元素的 `baseBranch` 为 `main`
- **AND** 系统不启动 Wails GUI

#### Scenario: List completed todos from project child directory

- **WHEN** 当前工作目录为已登记项目 `frontend-app` 下的子目录 `src/components`
- **AND** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的完成时项目快照匹配项目 `frontend-app`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `修复登录问题` 的元素

#### Scenario: List completed todos from git worktree child directory

- **WHEN** 当前工作目录为项目 `frontend-app` 的 Git linked worktree 子目录 `build/bin`
- **AND** TodoAI 已知项目 `frontend-app` 记录的是该 linked worktree 的源仓库路径
- **AND** TODO `worktree 子目录任务` 的状态为 `completed`
- **AND** TODO `worktree 子目录任务` 的完成时项目快照匹配项目 `frontend-app`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `worktree 子目录任务` 的元素

#### Scenario: Open todos are excluded

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** TODO `待执行任务` 的状态为 `not-started`
- **AND** TODO `执行中任务` 的状态为 `in-progress`
- **AND** TODO `已完成任务` 的状态为 `completed`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `已完成任务` 的元素
- **AND** stdout JSON 数组不包含 `taskName` 为 `待执行任务` 的元素
- **AND** stdout JSON 数组不包含 `taskName` 为 `执行中任务` 的元素

#### Scenario: Missing branch fields use placeholders

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** TODO `旧任务` 的状态为 `completed`
- **AND** TODO `旧任务` 的完成时项目快照匹配项目 `frontend-app`
- **AND** 该项目快照缺少 worktree 分支或 base 分支
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `旧任务` 的元素
- **AND** 该元素缺失的分支字段值为 `-`

#### Scenario: Unknown project directory returns error

- **WHEN** 当前工作目录不属于任何 TodoAI 已知项目或其子目录
- **THEN** 执行 `todoai list --done` 返回失败
- **AND** 输出说明无法定位 TodoAI 项目

#### Scenario: No completed todos returns empty state

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** 该项目没有匹配的 `completed` TODO
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout 为 JSON 空数组
