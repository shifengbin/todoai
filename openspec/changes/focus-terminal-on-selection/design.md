## Context

用户在左侧 TODO 工作树中点击终端条目时，前端会通过 `select-terminal` 事件调用 `App.vue` 的 `selectTerminal(terminalId)`。该流程已负责调用后端 `SelectTerminal`、应用新的 active terminal 状态、挂载或切换右侧 xterm pane，并在需要时恢复退出终端。

`TerminalSessionManager` 已提供 `focus(terminalId)`，当前只在终端右键菜单粘贴后使用。缺失的是从左侧终端条目显式选择终端后，将焦点交给右侧 xterm 的调用点。

## Goals / Non-Goals

**Goals:**

- 用户点击左侧 TODO 树中的终端条目后，可以直接在右侧终端输入。
- 自动聚焦限定在用户明确选择终端的交互路径，避免后台状态更新或页面初始化抢焦点。
- 复用现有 xterm session 聚焦能力，不改变后端 API 或终端会话模型。

**Non-Goals:**

- 不改变终端创建、删除、自动重启、历史回放、右键菜单或复制粘贴语义。
- 不为所有 active terminal 变化统一抢焦点。
- 不新增全局焦点管理系统或外部依赖。

## Decisions

### 在 `selectTerminal` 流程中显式聚焦

实现应在 `App.vue` 的 `selectTerminal(terminalId)` 完成状态应用和 `activateActiveTerminal()` 后，调用现有 `terminalManager.focus(terminalId)`。该调用点只覆盖用户点击左侧终端条目的路径，符合需求边界。

备选方案是在 `activateActiveTerminal()` 增加 `focus` 参数，由选择终端时传入。这可以复用激活流程，但会扩大共享函数的语义，需要确认所有调用点是否允许抢焦点。

另一个备选方案是在 `TerminalSessionManager.activate()` 内部总是聚焦。该方案代码最少，但会让页面初始化、自动恢复、删除后切换、TODO project 切换等路径也抢焦点，因此不采用。

### 保持聚焦为前端行为

聚焦是浏览器/xterm DOM 行为，不需要后端参与。后端 `SelectTerminal` 继续只负责 active terminal 状态和最近选择时间。

### 用现有选择终端测试覆盖

在现有前端测试 “selects a terminal from the project tree” 中增加 xterm `focus()` 断言，确保点击左侧终端条目后右侧对应终端获得焦点。

## Risks / Trade-offs

- 用户选择退出终端时也会聚焦右侧终端 pane → 该行为符合“点击终端后右侧终端获取焦点”，且现有自动重启流程保持不变。
- 如果 xterm session 尚未创建，直接聚焦可能无效 → 先执行 `activateActiveTerminal()` 可确保 session 已创建并挂载。
- 如果后端 `SelectTerminal` 失败，不应聚焦旧终端 → 聚焦调用应保留在 try 成功路径内。
