## MODIFIED Requirements

### Requirement: Persist Todos

系统 SHALL 在当前 workspace 数据目录中持久化 TODO、TODO 描述、TODO 优先级、TODO 工作流状态、TODO 工程副本、TODO 选中状态和已完成状态，并 SHALL 在该 workspace 重新打开后恢复。不同 workspace 的 TODO 数据 SHALL NOT 全局共享。TODO 工程副本 SHALL 保存添加时的项目名称、路径、来源候选 ID 和选择工程时选择的 base 分支，且 SHALL NOT 依赖全局候选继续存在。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。旧数据中状态为 `active` 的 TODO SHALL 按 `not-started` 处理。旧数据中状态为 `archived` 且归档原因为 `completed` 的 TODO SHALL 按 `completed` 处理。旧数据中状态为 `archived` 且归档原因为 `deleted` 的 TODO SHALL 不在 TODO 工作区列表中展示。旧 TODO 工程副本缺少 base 分支时，系统 SHALL 保留该工程副本并在需要分支信息的视图中按无法确认处理。

#### Scenario: Todo workspace is restored after reopening workspace

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户选择 base 分支 `main`
- **AND** 用户关闭并重新打开 workspace `/work/customer-a`
- **THEN** `执行中` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** TODO `修复登录问题` 的状态仍为 `in-progress`
- **AND** `frontend-app` 仍保存为该 TODO 下的工程副本
- **AND** 该工程副本仍包含添加时保存的路径
- **AND** 该工程副本仍包含选择工程时选择的 base 分支 `main`

#### Scenario: Todo workspace is isolated by workspace

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** TODO 工作区不显示 `修复登录问题`

#### Scenario: Legacy todo project copies are populated

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 包含旧 `todoProject` 引用 `project-a`
- **AND** 旧项目库中 `project-a` 的名称为 `frontend-app`
- **AND** 旧项目库中 `project-a` 的路径为 `/repo/frontend-app`
- **AND** 用户打开该 workspace
- **THEN** TODO `修复登录问题` 下的工程副本名称为 `frontend-app`
- **AND** 该工程副本路径为 `/repo/frontend-app`
- **AND** 该工程副本来源候选 ID 指向迁移后的全局候选或旧项目 ID
- **AND** 该工程副本缺少 base 分支时仍被保留

#### Scenario: Legacy todo without priority uses medium

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开该 workspace
- **THEN** TODO 工作区显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

#### Scenario: Legacy active todo becomes not-started

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `active`
- **AND** 用户打开该 workspace
- **THEN** `未执行` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `not-started` 处理

#### Scenario: Legacy completed archived todo remains completed

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `archived`
- **AND** TODO `修复登录问题` 的归档原因为 `completed`
- **AND** 用户打开该 workspace
- **THEN** `已完成` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `completed` 处理

#### Scenario: Legacy deleted archived todo is hidden

- **WHEN** 当前 workspace 持久化数据中 TODO `废弃任务` 的状态为 `archived`
- **AND** TODO `废弃任务` 的归档原因为 `deleted`
- **AND** 用户打开该 workspace
- **THEN** TODO 工作区不显示 `废弃任务`
- **AND** TODO 工作区不在 `已完成` 视图显示 `废弃任务`

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供已完成查看功能。已完成列表 SHALL 只显示状态为 `completed` 的 TODO，并 SHALL 按完成时间倒序展示，最近完成的 TODO 排在前面。完成时间 SHALL 优先使用 `completedAt`，当 `completedAt` 缺失时 SHALL 使用 `archivedAt` 作为兼容旧数据的兜底。缺失有效完成时间的已完成 TODO SHALL 排在有完成时间的 TODO 之后。已完成列表 SHALL 展示完成时保存的项目快照。项目快照 SHALL 优先展示完成时保存的 worktree 分支和 base 分支，格式为 `worktree 分支 -> base 分支`，并 SHALL 使用保存的项目路径异步检查 worktree 分支是否已合并到 base 分支。异步合并检查 SHALL NOT 阻塞已完成列表渲染、视图切换或其它 TODO 操作。检查结果确认已合并时系统 SHALL 显示对号；检查结果确认未合并时系统 SHALL 显示黄色三角感叹号；路径不可用、Git 不可用、非 Git 仓库、分支缺失、检查超时或历史快照缺少分支信息时系统 SHALL 显示黄色三角感叹号并按无法确认处理。已删除 TODO SHALL NOT 显示在已完成列表中。

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 已完成
- **AND** TODO `修复登录问题` 的完成时项目快照包含项目 `frontend-app`
- **AND** 该快照的 worktree 分支为 `todo/fix-login`
- **AND** 该快照的 base 分支为 `main`
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示该 TODO 的完成时间
- **AND** `已完成` 视图显示项目快照 `frontend-app`
- **AND** `已完成` 视图显示分支关系 `todo/fix-login -> main`
- **AND** 系统异步检查 `todo/fix-login` 是否已合并到 `main`

#### Scenario: Completed snapshot shows merged status

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照 base 分支为 `main`
- **AND** 异步 Git 检查确认 `todo/fix-login` 已合并到 `main`
- **THEN** `已完成` 视图在该项目快照旁显示对号

#### Scenario: Completed snapshot shows unmerged warning

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **AND** completed TODO `修复登录问题` 的项目快照 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照 base 分支为 `main`
- **AND** 异步 Git 检查确认 `todo/fix-login` 未合并到 `main`
- **THEN** `已完成` 视图在该项目快照旁显示黄色三角感叹号

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

### Requirement: View Completed Todo Details

系统 SHALL 允许用户从 `已完成` 视图打开 completed TODO 的详情。completed TODO 详情 SHALL 复用 TODO 详情弹窗，并 SHALL 以只读模式显示 TODO 标题、描述、优先级和完成时项目快照。完成时项目快照 SHALL 显示项目名称、worktree 分支和 base 分支；当历史快照缺少分支信息时，系统 SHALL 保留项目名称并按无法确认展示。只读详情 SHALL 隐藏保存按钮，SHALL 禁止编辑 TODO 字段或项目，且 SHALL NOT 恢复终端、启动 shell 进程或重新建立项目关联。

#### Scenario: User opens completed todo details

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级为 `高`
- **AND** TODO `修复登录问题` 的完成时项目快照包含名称为 `frontend-app` 的项目
- **AND** 该完成时项目快照的 worktree 分支为 `todo/fix-login`
- **AND** 该完成时项目快照的 base 分支为 `main`
- **AND** 用户在 `已完成` 视图中打开 TODO `修复登录问题` 的详情
- **THEN** 系统打开 TODO 详情弹窗
- **AND** 详情弹窗显示标题 `修复登录问题`
- **AND** 详情弹窗显示描述 `登录后跳回首页`
- **AND** 详情弹窗显示优先级 `高`
- **AND** 详情弹窗显示项目快照 `frontend-app`
- **AND** 详情弹窗显示分支关系 `todo/fix-login -> main`

#### Scenario: Completed todo details are read-only

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统隐藏详情弹窗的保存按钮
- **AND** 系统不允许编辑 TODO 标题
- **AND** 系统不允许编辑 TODO 描述
- **AND** 系统不允许修改 TODO 优先级
- **AND** 系统不允许新增或移除项目

#### Scenario: Completed todo details do not restore runtime context

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程
- **AND** 系统不把完成时项目快照恢复为活动 TODO 项目关联
