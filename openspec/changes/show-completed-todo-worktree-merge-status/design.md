## Context

当前 TODO 工程副本只保存添加时的项目名称、路径和来源候选 ID。完成 TODO 时，后端将这些工程副本转换为 `TodoProjectSnapshot`，保存到 completed TODO，并删除运行时 `TodoProject` 关联。`已完成` 视图和 completed TODO 只读详情都读取这个快照展示项目名称和路径。

用户希望在完成任务后看到 worktree 分支是否已经合并回 base 分支。base 分支不是运行时推断值，而是用户选择工程时选择的分支；因此该分支必须随 TODO 工程副本持久化，并在完成时复制到完成快照。

## Goals / Non-Goals

**Goals:**

- 在创建 TODO、编辑 TODO 和向 TODO 添加工程时记录每个工程的 base 分支。
- 完成 TODO 时保存每个工程快照的 worktree 分支名、base 分支名和用于 Git 检查的路径。
- 在 `已完成` 视图中用 `worktree 分支 -> base 分支` 替代路径作为主要项目信息。
- 异步检查 worktree 分支是否已合并到 base 分支，不能阻塞界面渲染或切换。
- 在 completed TODO 只读详情中展示与已完成列表一致的分支快照信息。

**Non-Goals:**

- 不自动执行 merge、rebase、push 或删除 worktree。
- 不改变 TODO 完成、删除、批量删除、排序和终端清理语义。
- 不为历史 completed TODO 推断不存在的 base 分支；旧数据只能降级展示。

## Decisions

1. **把 base 分支作为选择工程的一部分保存，而不是完成时推断。**

   `CreateTodoRequest`、`UpdateTodoRequest` 和添加工程接口应支持传入带 `projectId`、`baseBranch` 的选择对象。`TodoProject` 持久化 `BaseBranch`。旧的 `projectIds` 可在迁移期保留兼容，缺少 base 分支的旧数据显示为无法确认。这样可以忠实表达“选择工程时选择的分支”，避免用户后来切换工程分支导致完成记录失真。

2. **完成快照保存分支信息，合并状态不持久化。**

   `TodoProjectSnapshot` 保存 `Path`、`WorktreeBranch`、`BaseBranch`。`Path` 用于后续 Git 检查和兼容旧详情；列表主要展示分支信息。合并状态由前端打开 `已完成` 视图后异步查询，因为 merge 状态会随用户后续合并操作变化，持久化会变旧。

3. **后端提供专用合并状态查询接口。**

   新接口接收完成快照标识或必要参数，返回每个 snapshot 的状态：`checking` 由前端管理，后端返回 `merged`、`unmerged`、`unknown` 及原因。Git 命令沿用现有 2 秒超时和 git 不可用处理思路。合并判断使用 Git ancestry 语义：worktree 分支提交已被 base 分支包含时视为已合并。

4. **前端异步批量刷新并缓存结果。**

   `App.vue` 管理完成快照合并状态 map，进入 `已完成` 视图或 completed TODO 数据变化时触发检查。请求使用 request id 或 generation 丢弃过期响应，避免用户快速切换 workspace、TODO 视图或重新加载时出现旧结果覆盖。`ProjectSidebar.vue` 接收状态并渲染图标，不直接执行 Wails 调用。

5. **历史和异常数据使用警告降级。**

   缺少路径、路径不可用、非 Git 仓库、缺少 worktree/base 分支、Git 未安装、检查超时或分支不存在时，显示黄色三角感叹号并标记为无法确认。只有明确确认 worktree 分支已被 base 分支包含时才显示对号。

## Risks / Trade-offs

- [请求结构变化影响现有前端和 Wails 绑定] -> 保留兼容转换层，测试覆盖旧 `projectIds` 请求和新选择对象请求。
- [大量 completed TODO 同时触发 Git 命令] -> 前端去重缓存，后端单次接口支持批量查询或前端限制并发，避免界面和系统资源抖动。
- [Git 分支引用本地不存在] -> 返回 `unknown`，UI 用黄色警告表达需要用户处理本地分支状态。
- [历史 completed TODO 缺少分支字段] -> 继续展示项目名称，分支信息显示降级状态，不尝试从路径或分支命名规则猜测。
- [合并状态随时间变化] -> 不持久化状态，用户重新打开或刷新完成视图时重新异步检查。

## Migration Plan

1. 读取旧数据时为缺少 `baseBranch` 的 `TodoProject` 和缺少分支字段的 `TodoProjectSnapshot` 保持空值，不进行破坏性迁移。
2. 新创建或新编辑的 TODO 工程关联开始写入 base 分支。
3. 新完成的 TODO 快照写入 worktree/base 分支；旧 completed TODO 使用降级展示。
4. 若需要回滚，旧客户端会忽略新增 JSON 字段；新接口未被调用时不会影响已有 TODO 数据。

## Open Questions

- 工程选择 UI 中 base 分支列表是否只来自本地分支，还是同时包含远端分支。实现时应优先复用已有 Git 命令能力；如果没有现成分支选择能力，需要在本 change 内补齐最小可用分支选择。
