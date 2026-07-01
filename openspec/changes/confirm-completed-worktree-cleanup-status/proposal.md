## Why

已完成列表当前会反复异步检查 completed TODO 项目快照的 worktree 分支是否已合并到 base 分支；当用户已经清理 worktree 目录或删除 worktree 分支后，系统只能把它当作无法确认并显示警告。对已完成分类来说，worktree 已被清理代表该完成记录不再需要继续追踪合并状态，继续提示警告和重复检查会制造噪音。

## What Changes

- 已完成项目快照的 worktree 目录不存在时，系统将该快照视为已确认，不再显示无法确认警告。
- 已完成项目快照的 worktree 分支不存在时，系统将该快照视为已确认，不再显示无法确认警告。
- 系统在确认快照已合并、worktree 目录已清理或 worktree 分支已清理后，持久记录该快照状态；后续打开已完成视图不再重复执行 Git 检查。
- 未合并、Git 不可用、非 Git 仓库、检查超时、历史快照缺少分支信息等仍按无法确认或未合并状态展示，不自动持久化为已确认。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 调整已完成列表中项目快照合并状态的展示与持久化规则，新增 worktree 目录或分支清理后的已确认状态。

## Impact

- 后端 completed TODO 项目快照数据结构与持久化迁移兼容。
- 后端已完成项目快照合并状态查询 API 需要能识别 worktree 路径缺失和 worktree 分支缺失，并在可确认时写回状态。
- 前端已完成列表需要优先使用持久化状态，跳过已确认快照的后续查询。
- 测试覆盖 Go 后端状态判断、快照持久化、Vue 已完成视图查询跳过和图标展示。
