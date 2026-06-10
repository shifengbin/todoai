## Why

当前终端树只能显示 shell 层面的当前命令，例如 `codex`。当 `codex` 这类交互式 TUI 命令长期运行时，用户无法从多终端列表中判断它内部正在执行、等待用户回答，还是已完成本轮提示词并恢复空闲。

系统终端通常会通过窗口标题变化表达这些运行时状态。嵌入式终端也应捕获这些标题变化，并在终端列表中给出轻量、可扫视的状态提示。

## What Changes

- 监听每个 xterm 会话的 OSC 0 / OSC 2 title change 事件，并将标题变化关联到对应 terminal。
- 为终端维护非持久化的交互式活动状态，例如空闲、执行中、需要用户输入。
- 在左侧项目终端树的终端行中展示活动状态：
  - 执行中显示动态指示图标。
  - 需要用户输入显示注意提示，例如叹号。
  - 本轮完成后停止动态效果，并回到命令名或 shell 名的稳定标题。
- 保持现有 shell 命令标签机制：普通命令仍由 command-start / command-end 控制，交互式 TUI 内部状态由 title change 控制。
- 不新增持久化字段，不改变后端 shell 会话生命周期。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `embedded-terminal-emulation`: 增加对 xterm title change 事件的捕获和 terminal 级路由要求。
- `project-workspace`: 增加在项目终端树中展示交互式终端活动状态的要求。

## Impact

- 前端 xterm 会话创建逻辑需要订阅 `onTitleChange` 并向 `TerminalSessionManager` / `App.vue` 转发 terminal-scoped 事件。
- 前端 runtime terminal descriptor 需要维护 title/activity 相关的临时 UI 状态。
- `ProjectSidebar.vue` 需要基于 terminal activity state 渲染状态图标和动画类。
- `style.css` 需要增加执行中、需要输入、空闲状态的视觉样式。
- 前端测试需要覆盖 title change 路由、状态推导和终端树展示。
