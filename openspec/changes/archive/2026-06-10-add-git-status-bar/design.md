## Context

应用是 Wails 桌面应用，后端 Go 负责项目列表、终端会话和本地文件系统访问，前端 Vue 在 `App.vue` 中渲染主工作区。当前 `Project` 模型只保存打开目录的持久化信息，`ProjectState` 在返回给前端时叠加运行中的终端状态。

Git 状态是易变的运行时信息，不适合写入项目配置。状态栏属于当前项目工作区体验，应跟随 `activeProject` 变化，并且不能让终端区域在刷新时反复改变高度。

## Goals / Non-Goals

**Goals:**

- 在工作区底部提供固定高度状态栏，展示当前项目的分支名和改动文件数量。
- 对非 Git 仓库、不可用项目、未选择项目和 Git 查询失败提供稳定且可理解的显示状态。
- 通过后端读取项目目录的 Git 状态，避免前端直接访问文件系统或执行命令。
- 在用户自然需要新状态的时机刷新：启动、切换项目、终端命令结束、窗口重新获得焦点。
- 保持终端区域布局稳定，避免状态栏出现或隐藏导致 xterm 尺寸抖动。

**Non-Goals:**

- 不实现提交、暂存、切换分支等 Git 操作。
- 不展示完整文件列表、diff、远端详情或复杂同步状态。
- 不持久化 Git 状态到项目配置。
- 不引入文件系统 watcher 或持续高频轮询。

## Decisions

### 后端提供运行时 Git 查询 API

新增 `GitStatus` 结构和 `App.GetProjectGitStatus(projectID string)` Wails 方法。方法通过 `ProjectManager.GetProject` 找到项目路径，校验项目可用后执行 Git 查询。

替代方案是把 Git 信息合并到 `ProjectState`。这会让每次项目状态查询都触发 Git 命令，并把易变信息混入持久化项目模型，后续也更难控制刷新频率。因此选择单独 API。

### 使用 `git status --porcelain=v2 --branch`

后端使用 `git -C <path> status --porcelain=v2 --branch` 获取机器可解析输出。解析 `# branch.head` 得到当前分支，解析普通变更行统计 changed/staged/unstaged/untracked 数量，并从 branch 行保留 ahead/behind 信息。

替代方案是分别运行 `git branch --show-current` 和 `git status --short`。多命令实现更简单但开销更高，也更容易出现两次读取状态不一致。单命令更适合状态栏。

### 非 Git 仓库不是错误

当 Git 返回“not a git repository”一类结果时，API 返回 `isRepo: false`，前端展示 `Not a git repository`。只有项目不存在、路径不可用或不可预期命令失败才作为错误状态处理。

### 刷新由前端事件驱动

前端维护 `gitStatus`、`gitStatusLoading` 和 `gitStatusError`。在 `applyState` 后如果 active project 变化则刷新；应用启动和 `focus` 事件刷新；`handleTerminalCommandState` 收到 command-end 时刷新当前项目。

不做定时轮询。大仓库的 `git status` 可能较慢，事件驱动能覆盖主要交互场景，同时减少不必要的命令执行。

### 状态栏固定高度

`workspace` 的第三行改为固定状态栏高度。错误信息可以在状态栏右侧呈现，Git 信息缺失时也保留栏位。这样状态栏内容变化不会改变终端 surface 高度，xterm 只在窗口尺寸或布局实际变化时重新 fit。

## Risks / Trade-offs

- Git 查询在大仓库中可能较慢 → 后端使用短超时并让前端显示 loading/failed 状态，不阻塞已有终端会话。
- `git` 可执行文件不存在 → API 返回错误状态，前端展示 Git unavailable，不影响项目和终端功能。
- 命令结束刷新可能漏掉外部编辑器造成的改动 → 窗口 focus 刷新覆盖常见外部修改场景，暂不引入 watcher。
- porcelain v2 解析需要覆盖多种状态行 → 用独立解析函数和表驱动测试约束行为。

## Migration Plan

这是纯新增运行时能力，不需要数据迁移。回滚时移除 Wails 方法、前端状态栏渲染和相关测试即可；已有项目配置保持兼容。

## Open Questions

无。
