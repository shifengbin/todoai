## Why

当前 TODO 添加项目存在两条后端路径：旧路径会在执行中 TODO 上准备任务工作区和 Git worktree，而下拉菜单使用的新路径只保存项目关联，导致执行中 TODO 通过下拉添加项目后没有创建对应 worktree。与此同时，未关联任何项目的 TODO 进入执行中时也会创建任务目录和 `README.md`，与“没有添加项目就不需要新建 README 文件”的使用预期不一致。

## What Changes

- 下拉菜单向执行中 TODO 添加项目后，系统 SHALL 为新增 TODO 项目准备 Git worktree，并刷新任务 README 中的项目信息。
- 未关联任何项目的 TODO SHALL NOT 仅因为进入执行中而创建 `README.md`。
- 未关联任何项目且没有其它需要落盘文件的 TODO SHALL NOT 创建任务工作区目录。
- 添加第一个项目到已执行中的 TODO 时，系统 SHALL 创建任务工作区目录、准备该项目 worktree，并生成包含项目信息的 `README.md`。
- 更新现有无项目 TODO 初始化文件相关要求，避免“无项目只创建 README”的旧行为继续作为期望。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 调整通过项目选择控件向 TODO 添加项目后的执行中 TODO 行为。
- `todo-worktree-workspaces`: 调整任务工作区目录和 README 的创建条件，并补充执行中 TODO 后续添加项目时的 worktree 准备要求。
- `todo-initialization-files`: 调整无关联项目 TODO 的初始化文件与 README 创建语义，移除“无项目只创建 README”的要求。

## Impact

- 后端：`App.AddProjectSelectionsToTodo`、任务工作区准备和 README 写入流程。
- 测试：补充下拉入口对应后端 API 的 worktree 准备覆盖；更新无项目 TODO 不创建 README 的期望。
- 前端 Wails API 不需要新增接口，现有下拉菜单调用保持不变。
