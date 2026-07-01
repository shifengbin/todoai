## Context

当前 completed TODO 的项目快照只保存项目名称、路径、base 分支和 worktree 分支。前端进入 `已完成` 视图时会基于这些快照组装请求，调用后端 `GetCompletedTodoProjectMergeStatuses`，后端再用 Git 判断 worktree 分支是否已经合并到 base 分支。前端只有运行期内存缓存，切换回已完成视图或重新打开应用后仍会重新查询。

新的产品规则把 completed 分类中的 worktree 清理视为终态：如果 worktree 目录已经删除，或 worktree 分支已经删除，系统应直接显示对号，并把该结果记录下来，后续不再查询 Git。

## Goals / Non-Goals

**Goals:**

- 为 completed TODO 项目快照增加可选的持久化合并/确认状态，兼容旧数据。
- 在已合并、worktree 目录已删除或 worktree 分支已删除时写回快照状态。
- 前端优先使用快照持久状态，已确认快照不再触发后端 Git 检查。
- 保留未合并、Git 不可用、非 Git 仓库、超时和历史分支信息缺失时的警告展示。

**Non-Goals:**

- 不在 TODO 完成、删除或关闭 workspace 时自动清理 worktree 目录或 Git 分支。
- 不改变 `todoai list --done` 的输出字段。
- 不把未合并状态持久化为终态；未合并仍允许后续重新检查。

## Decisions

### 1. 在 `TodoProjectSnapshot` 上持久化终态

为 completed 快照增加可选字段，例如 `mergeStatus` 和 `mergeStatusReason`。`mergeStatus` 用于表示快照已经有可直接展示的终态，`mergeStatusReason` 用于区分来源，例如真实合并、worktree 目录已删除或 worktree 分支已删除。

原因：状态属于 completed 快照本身，放在 `projects.json` 可跨应用重启保留，也不会和运行期 UI 缓存混在一起。备选方案是写入 `todo-project-ui-state.json`，但该文件表达的是布局与视图偏好，不适合保存业务状态。

### 2. 查询接口执行幂等写回

`GetCompletedTodoProjectMergeStatuses` 继续作为前端获取状态的入口，但请求需要携带足够定位快照的信息，例如 `todoId`、`snapshotIndex` 和当前快照 fingerprint。后端在确认状态为已合并、worktree 目录已删除或 worktree 分支已删除时，加载并匹配当前 completed 快照，确认快照仍与请求一致后写回持久状态。

原因：前端已经在该接口处集中发起异步状态检查，保持入口不变可以减少 UI 交互面变化。匹配 fingerprint 可以避免异步旧请求把已经变化的快照写错。

### 3. 区分“展示对号”和“真实 Git 合并”

UI 对真实合并和清理后确认都显示对号，但持久化原因保留差异。这样满足已完成分类的展示规则，同时避免未来需要审计时无法区分“代码已合并”和“用户清理后不再追踪”。

备选方案是统一存为 `merged`。实现更简单，但会把清理确认伪装成 Git 合并事实，不利于后续扩展。

### 4. 只持久化正向终态

只有能显示对号并停止追踪的状态会写回。`unmerged` 和 `unknown` 不写入快照，因为这些状态可能随着用户后续合并、恢复仓库或修复 Git 环境而变化。

## Risks / Trade-offs

- [Risk] `GetCompletedTodoProjectMergeStatuses` 从只读查询变为可能写入状态。 → 写入必须幂等，只在匹配 completed 快照且状态为空时落盘；失败时仍返回可展示状态，不阻塞列表。
- [Risk] worktree 目录删除并不证明代码已合并。 → 持久化原因明确记录为 worktree 清理确认，UI 显示对号但数据层不把它混同为真实合并。
- [Risk] 旧前端或旧数据没有新增字段。 → 新字段使用 `omitempty`，旧快照缺失字段时继续走现有异步检查。
- [Risk] 路径存在但已经不是 Git worktree。 → 不按清理确认处理，继续归为 unknown 警告，避免把损坏或错误路径误判为已确认。

## Migration Plan

新增字段为可选字段，不需要一次性迁移历史 `projects.json`。历史 completed 快照首次进入已完成视图时仍按现有信息检查；一旦确认真实合并、worktree 目录删除或 worktree 分支删除，系统写回新增字段。回滚时旧版本会忽略这些未知 JSON 字段。

## Open Questions

无。
