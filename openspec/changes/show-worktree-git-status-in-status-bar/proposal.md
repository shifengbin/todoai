## Why

当前 TODO project 已经在执行态使用独立 Git worktree 作为终端工作目录，但底部状态栏仍按原项目路径查询 Git 状态。用户在任务 worktree 中改动文件后，状态栏会显示原仓库分支和改动数量，无法反映当前任务目录的真实状态。

## What Changes

- 当当前上下文是 TODO project 且该 TODO project 已准备好 worktree 时，状态栏 Git 信息改为查询该 worktree 目录。
- 状态栏刷新、focus 去重和命令结束刷新按 TODO project 上下文区分，避免同一个原项目的多个 TODO worktree 互相串状态。
- 没有选中 TODO project、TODO project 未准备 worktree 或 worktree 路径不可用时，继续显示稳定的空状态、不可用状态或既有错误状态。
- 不改变 Git 状态解析格式，不新增提交、拉取、推送等 Git 操作。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `project-workspace`: 当前激活项目的 Git 状态栏在 TODO project 上下文中需要展示对应 worktree 目录的 Git 信息，而不是来源项目目录的 Git 信息。

## Impact

- 后端：新增或调整 Git 状态查询 API，使调用方能够按 TODO project 的 `worktreePath` 查询状态，并保留普通项目路径查询能力。
- 前端：调整 `App.vue` 中状态栏的 active context、请求参数、缓存/去重 key 和刷新触发逻辑。
- 测试：增加 Go 测试覆盖 worktree path 查询，增加前端测试覆盖同一原项目下不同 TODO worktree 的状态栏隔离。
