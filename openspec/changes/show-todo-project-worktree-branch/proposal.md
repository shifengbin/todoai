## Why

当前左侧 TODO 项目列表只显示项目名称，用户在 worktree 内切换分支后无法直接从 TODO 树确认该项目的真实当前分支。同时没有选择项目时底部状态栏仍显示 `No project` 形式的 Git 状态空信息，增加了无效状态噪音。

## What Changes

- 左侧 TODO 项目列表中的项目名称后追加当前 worktree 真实分支，格式为 `项目名称(分支名称)`。
- 分支名称以 TODO project worktree 路径当前 Git 状态为准，不使用创建 worktree 时保存的静态 `worktreeBranch` 字段作为显示来源。
- 当用户在该 TODO project 的 worktree 终端中切换分支并命令结束后，左侧 TODO 项目列表中的分支名称刷新为最新分支。
- 当没有选择项目或 TODO project 时，底部状态栏不显示 Git 状态 chip，也不显示 `No project` 文案。
- 当选择 TODO 级控制台时，底部状态栏只检查该 TODO 任务文件夹根目录本身的 Git 仓库；若该目录本身不是 Git 仓库，则不显示 Git 状态 chip。
- TODO 初始化文件写入必须在该 TODO 关联的 worktree 都创建完成后执行；不能在 worktree 仍未准备完成时提前执行初始化。

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `todo-workspace`: 左侧 TODO 工作树中的 TODO project 行需要展示当前 worktree 真实分支，并在相关终端命令结束后刷新。
- `project-workspace`: 当前没有激活项目上下文时，底部 Git 状态栏不再显示 Git 状态空信息。
- `project-workspace`: 当前激活的是 TODO 级控制台时，底部 Git 状态栏使用 TODO 任务文件夹根目录的 Git 状态，不沿用上一个项目或 TODO project 的 Git 状态。
- `todo-initialization-files`: 初始化文件写入顺序需要晚于 TODO project worktree 准备完成。

## Impact

- Frontend TODO tree rendering in `frontend/src/components/ProjectSidebar.vue`.
- Frontend Git status state and refresh flow in `frontend/src/App.vue`.
- TODO workspace preparation ordering in `todo_workspace_app.go`.
- Existing Wails API usage for `GetTodoProjectGitStatus`; no new backend API is expected unless implementation finds the current API insufficient.
- Unit tests for TODO project list rendering, Git status empty state, branch refresh after terminal command completion, and initialization ordering after worktree preparation.
