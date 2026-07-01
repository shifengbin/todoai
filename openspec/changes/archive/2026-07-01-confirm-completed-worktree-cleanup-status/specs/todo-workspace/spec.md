## MODIFIED Requirements

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供已完成查看功能。已完成列表 SHALL 只显示状态为 `completed` 的 TODO，并 SHALL 按完成时间倒序展示，最近完成的 TODO 排在前面。完成时间 SHALL 优先使用 `completedAt`，当 `completedAt` 缺失时 SHALL 使用 `archivedAt` 作为兼容旧数据的兜底。缺失有效完成时间的已完成 TODO SHALL 排在有完成时间的 TODO 之后。已完成列表 SHALL 展示完成时保存的项目快照。项目快照 SHALL 优先展示完成时保存的 worktree 分支和 base 分支，格式为 `worktree 分支 -> base 分支`。当项目快照包含持久化的已确认状态时，系统 SHALL 直接显示对号并 SHALL NOT 对该快照发起异步 Git 检查。当项目快照没有持久化的已确认状态且保存的项目路径、worktree 分支和 base 分支均可用于检查时，系统 SHALL 使用保存的项目路径异步检查 worktree 分支是否已合并到 base 分支。异步合并检查 SHALL NOT 阻塞已完成列表渲染、视图切换或其它 TODO 操作。检查结果确认已合并时系统 SHALL 显示对号并持久记录该快照为已确认；检查发现保存的 worktree 路径不存在时系统 SHALL 显示对号并持久记录该快照为已确认；检查发现 worktree 分支不存在时系统 SHALL 显示对号并持久记录该快照为已确认。检查结果确认未合并时系统 SHALL 显示黄色三角感叹号；Git 不可用、非 Git 仓库、检查超时或历史快照缺少分支信息时系统 SHALL 显示黄色三角感叹号并按无法确认处理，且 SHALL NOT 持久记录为已确认。已删除 TODO SHALL NOT 显示在已完成列表中。

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 已完成
- **AND** TODO `修复登录问题` 的完成时项目快照包含项目 `frontend-app`
- **AND** 该快照的 worktree 分支为 `todo/fix-login`
- **AND** 该快照的 base 分支为 `main`
- **AND** 该快照没有持久化的已确认状态
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示该 TODO 的完成时间
- **AND** `已完成` 视图显示项目快照 `frontend-app`
- **AND** `已完成` 视图显示分支关系 `todo/fix-login -> main`
- **AND** 系统异步检查 `todo/fix-login` 是否已合并到 `main`

#### Scenario: Completed snapshot shows persisted confirmed status

- **WHEN** completed TODO `修复登录问题` 的项目快照包含持久化的已确认状态
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图在该项目快照旁显示对号
- **AND** 系统不对该项目快照发起异步 Git 检查

#### Scenario: Completed snapshot shows merged status

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照 base 分支为 `main`
- **AND** 该项目快照没有持久化的已确认状态
- **AND** 异步 Git 检查确认 `todo/fix-login` 已合并到 `main`
- **THEN** `已完成` 视图在该项目快照旁显示对号
- **AND** 系统持久记录该项目快照为已确认

#### Scenario: Completed snapshot with removed worktree directory shows confirmed status

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照保存了 worktree 路径
- **AND** 该 worktree 路径在磁盘上不存在
- **THEN** `已完成` 视图在该项目快照旁显示对号
- **AND** 系统持久记录该项目快照为已确认
- **AND** 用户之后再次打开 `已完成` 视图时系统不再检查该项目快照的 Git 状态

#### Scenario: Completed snapshot with removed worktree branch shows confirmed status

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照 worktree 路径仍是 Git 仓库
- **AND** 该项目快照 worktree 分支 `todo/fix-login` 已不存在
- **THEN** `已完成` 视图在该项目快照旁显示对号
- **AND** 系统持久记录该项目快照为已确认
- **AND** 用户之后再次打开 `已完成` 视图时系统不再检查该项目快照的 Git 状态

#### Scenario: Completed snapshot shows unmerged warning

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照 base 分支为 `main`
- **AND** 异步 Git 检查确认 `todo/fix-login` 未合并到 `main`
- **THEN** `已完成` 视图在该项目快照旁显示黄色三角感叹号
- **AND** 系统不把该项目快照持久记录为已确认

#### Scenario: Completed snapshot merge check is asynchronous

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 需要检查 worktree 分支是否已合并到 base 分支
- **THEN** 系统立即渲染 `已完成` 列表
- **AND** 系统在 Git 检查完成前不阻塞视图切换
- **AND** 系统在 Git 检查完成后更新该项目快照的合并状态图标

#### Scenario: Completed snapshot with unknown merge status shows warning

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `旧任务` 的项目快照缺少 worktree 分支或 base 分支
- **THEN** `已完成` 视图保留该 completed TODO
- **AND** `已完成` 视图在该项目快照旁显示黄色三角感叹号
- **AND** 系统不把该项目快照显示为已合并
- **AND** 系统不把该项目快照持久记录为已确认

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
