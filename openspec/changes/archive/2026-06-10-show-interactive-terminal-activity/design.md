## Context

当前应用已经有两层终端状态：

- 后端 `ShellSessionManager` 维护 shell 进程级状态，例如 `running` / `exited`。
- 前端通过自定义 OSC 777 shell integration 维护 `currentCommand`，普通命令开始时终端树显示命令名，命令结束后回到 shell 名。

这个机制无法表达交互式 TUI 内部状态。以 `codex` 为例，shell 层只知道 `codex` 这个命令仍在运行；但 `codex` 内部会在执行、等待用户回答、本轮完成时更新系统终端标题。xterm.js 已提供 `terminal.onTitleChange`，可监听 OSC 0 / OSC 2 标题变化。

## Goals / Non-Goals

**Goals:**

- 捕获每个嵌入式终端会话的 title change，并保持 terminal 级路由。
- 基于 title change 为终端维护临时交互式活动状态：`idle`、`busy`、`needs-input`。
- 在项目终端树中用轻量图标和动画展示活动状态，让用户切换多终端时可快速判断哪一个正在执行或等待输入。
- 保持现有 shell command label 行为，`currentCommand` 仍代表 shell 正在运行的命令名。
- 在 title change 不可用或程序不设置标题时保持现有行为。

**Non-Goals:**

- 不解析终端屏幕内容来推断状态。
- 不新增后端 Wails API、后端事件或持久化字段。
- 不限定只支持 `codex`，但第一版只做通用 title 模式和明确的 `!` 注意状态。
- 不把交互式活动状态用于控制 shell 生命周期、自动关闭终端或自动切换终端。

## Decisions

### 使用 xterm `onTitleChange` 作为活动信号

在 `createXtermSession` 中订阅 `terminal.onTitleChange`，通过 `TerminalSessionManager` 将事件转发为 `(terminalId, title)`。这样 title 事件和现有 output、command state 一样都保留 terminal 级关联，不需要后端参与。

替代方案是从 PTY 输出流中手动解析 OSC 0 / OSC 2。这个方案会重复 xterm 已经实现的终端解析逻辑，也容易误处理 ST / BEL 终止符、分片输出和控制序列边界。

### 将 title 文本和活动状态作为前端临时字段

每个 runtime terminal descriptor 增加前端临时字段：

```text
runtimeTitle: string
activityState: "idle" | "busy" | "needs-input"
```

这些字段只存在于前端运行态，不写入项目配置，也不扩展 Go `ProjectTerminal` 模型。`applyState` 合并后端状态时应保留同一 terminalId 的临时字段，类似当前保留 `currentCommand`。

### title 只驱动状态，终端显示名仍优先使用命令名

终端行的主要文字继续使用：

```text
currentCommand || shellName || "shell"
```

title change 不直接替换主标题，因为交互式程序可能把 spinner、感叹号或短状态写入窗口标题。主标题保持 `codex` 这类稳定命令名，activity indicator 单独展示 busy / needs-input。

### 使用保守状态推导

状态推导应基于当前 title、当前命令名、shell 名和已观察到的空闲 title 基线：

```text
normalize(title) == ""                  -> idle
attention title, e.g. contains "!"      -> needs-input
first non-attention launch title        -> idle baseline
title equals current command/name       -> idle
title equals idle baseline              -> idle
title has explicit busy signal          -> busy
later title differs from idle baseline  -> busy
```

当 `codex` / `claude` 刚启动并设置静态窗口标题时，该标题应作为空闲基线，不应立即显示 busy。后续如果 title 出现 spinner、working/thinking 等明确执行信号，或从已记录的空闲基线变化到其他状态，终端行显示 `busy`。当 title 回到空闲基线或 `codex` 这类稳定命令名时恢复 `idle`。如果 title 出现明确叹号状态，则显示 `needs-input` 并停止 busy 动画。

这比硬编码完整 `codex` 标题格式更稳妥：如果未来 title 文案变化，只要仍满足“执行时变化、需要输入时包含注意符号、完成时回到命令名”的模式，UI 行为仍然成立。

### shell command state 和 title activity 分层处理

`command-start` 仍设置 `currentCommand`，并在命令开始时清空旧 title activity，避免上一个交互式程序的状态残留。`command-end` 仍清空 `currentCommand` 和 title activity，因为 shell 已回到 prompt。

交互式程序运行期间不会触发 `command-end`，因此 title activity 可以在 `codex` 进程仍运行时独立变化。

### 在终端树中展示状态

`ProjectSidebar.vue` 在 terminal row 中增加一个固定宽度 activity slot：

- `busy`: 动态图标或旋转指示器。
- `needs-input`: 注意图标，例如叹号。
- `idle`: 保持占位但不显示强调图标，避免行宽跳动。

样式应保证终端名不被挤出容器，图标变化不改变行高或行宽。状态也应体现在 `title` 或 `aria-label` 中，避免完全依赖颜色和动画。

## Risks / Trade-offs

- 标题格式不是稳定协议 -> 使用保守通用推导，并通过测试覆盖已知 `busy` / `needs-input` / `idle` 输入。
- 普通程序也可能设置 title -> 只作为 UI 提示，不影响命令生命周期或终端控制。
- 某些程序不设置 title -> 保持现有命令名展示，不显示活动状态。
- title spinner 可能高频变化 -> 只更新小型前端状态和 CSS 类，避免触发后端调用；必要时可在前端做相同状态去重。
- 只用感叹号识别需要输入可能不覆盖所有工具 -> 第一版先满足当前 `codex` 观察到的模式，后续可把解析规则扩展为小型 pattern 列表。
