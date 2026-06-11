## Why

当前 TODO 工作流允许 `not-started` 和 `in-progress` 双向切换，并且未执行 TODO 也可以创建终端。这会让“未执行”和“执行中”的语义变弱：用户可以在未开始的任务下启动工作终端，也可以把已经执行的任务退回未开始。

这个变更将 TODO 工作流收紧为单向流程，使未执行任务只能开始或删除，执行中任务才能添加终端并完成或删除。

## What Changes

- 将 TODO 状态操作收紧为单向流转：`not-started` 只能开始为 `in-progress`，`in-progress` 只能完成为 `completed`。
- 移除执行中 TODO 退回未执行的用户入口和后端能力。
- 未执行 TODO 不再显示完成入口，也不显示添加终端入口。
- 执行中 TODO 保留完成和删除入口，并允许添加终端。
- 后端 API 对状态切换、完成 TODO、创建 TODO 终端执行同样的状态约束，避免绕过前端造成非法状态变化。
- 保留查看/编辑 TODO、添加项目和删除 TODO 等非状态执行动作的现有能力。

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `todo-workspace`: 收紧 TODO 工作流状态动作和完成入口的可用状态。
- `embedded-shell-sessions`: 收紧 TODO 项目上下文下创建终端的前置条件，只允许执行中的 TODO 创建终端。

## Impact

- Go 后端：`ProjectManager.ChangeTodoStatus`、`ProjectManager.CompleteTodo`、`App.CreateTodoTerminal` 及相关测试。
- Wails API 行为：现有方法签名不变，但非法状态下调用会返回错误。
- Vue 前端：TODO 行动作按钮、添加终端按钮显示条件、对应组件测试。
- OpenSpec：更新 `todo-workspace` 与 `embedded-shell-sessions` 的行为要求。
