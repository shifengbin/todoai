## Why

当前 TODO 项目行在 worktree 分支或目录被外部清理后只会丢失分支后缀或创建终端失败，用户无法从任务列表判断该 TODO project 的 worktree 已被清除。清理后的项目终端仍需要能打开，以便用户回到原项目目录继续查看或处理后续工作。

## What Changes

- 左侧 TODO 工作树中的 TODO project 行在 ready worktree 分支或目录不可用时，在原分支显示位置展示 `worktree已清除`。
- 新增/明确 TODO project worktree 的“已清除”语义：区别于 worktree 准备失败，表示曾经准备成功但后续 worktree 路径或分支已不存在。
- 对已清除 worktree 的 TODO project，项目终端创建不再报 worktree 不可用，而是使用该 TODO project 保存的原项目目录作为 shell 工作目录。
- 已清除状态不改变失败状态行为：worktree 准备失败的 TODO project 仍显示失败信息，并继续阻止创建项目终端。
- 不改变未执行 TODO 的终端限制：`not-started` TODO project 仍不允许创建项目终端。

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `todo-workspace`: TODO project 行需要在分支显示位置标记 worktree 已清除，并允许执行中 TODO 的已清除 project 继续暴露终端启动入口。
- `todo-worktree-workspaces`: TODO project worktree 状态需要表达“已清除”，并把 ready worktree 路径或分支缺失归类为清理后的状态而不是准备失败。
- `embedded-shell-sessions`: TODO project shell 启动目录需要在 worktree 已清除时回退到保存的原项目目录。

## Impact

- Go 状态模型和 worktree 校验：`project.go`、`app.go`、`todo_workspace_app.go`。
- Shell 创建和工作目录选择：`shell.go`、项目终端和背景启动命令入口。
- 前端 TODO 工作树展示与终端启动按钮状态：`frontend/src/components/ProjectSidebar.vue`。
- 前端分支刷新和 Git 状态处理：`frontend/src/App.vue`。
- Wails 前端模型生成文件：`frontend/wailsjs/go/models.ts`、`frontend/wailsjs/go/main/App.*`（如新增或调整导出模型字段）。
- 单元测试：Go 侧终端 cwd fallback、worktree 清除判定；Vue 侧项目行标记、分支刷新和启动按钮状态。
