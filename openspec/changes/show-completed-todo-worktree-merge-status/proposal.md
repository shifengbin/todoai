## Why

已完成 TODO 当前只展示完成时关联工程的名称和路径，用户无法直接判断该 TODO 对应的 worktree 分支是否已经合并回选择工程时的 base 分支。完成后仍需人工切换到 Git 检查，容易遗漏未合并的任务分支。

## What Changes

- 在选择工程加入 TODO 时记录当时选择的 base 分支，并随 TODO 工程副本持久化。
- 完成 TODO 时将 worktree 分支名、base 分支名和用于检查的工程路径保存到完成快照。
- `已完成` 视图中的工程信息由项目路径展示改为展示 `worktree 分支 -> base 分支`。
- `已完成` 视图异步检查每个完成快照的 worktree 分支是否已合并到 base 分支，避免阻塞界面。
- 合并状态展示为：已合并显示对号，未合并显示黄色三角感叹号；无法确认的 Git 状态也使用警告提示。
- 不改变 TODO 完成、删除、排序、批量删除和只读详情的现有交互语义。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: TODO 工程副本需要保存选择工程时的 base 分支；已完成 TODO 视图需要展示 worktree/base 分支信息并异步显示合并状态。

## Impact

- 后端模型：`CreateTodoRequest`、`UpdateTodoRequest`、`TodoProject`、`TodoProjectSnapshot` 需要扩展分支字段或项目选择结构。
- 后端 Git 能力：新增 worktree 分支读取和分支合并状态检查，复用现有 Git 命令超时与不可用处理思路。
- 前端：创建/编辑/添加工程选择 UI 需要携带 base 分支；已完成列表需要异步加载并渲染合并状态。
- Wails 绑定：新增或调整接口后需要重新生成前端绑定。
- 测试：需要覆盖数据持久化、完成快照、Git 合并检查、异步 UI 状态和已完成列表展示。
