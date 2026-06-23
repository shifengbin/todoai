## Context

嵌入式终端由 Go 后端启动 PTY shell，前端通过 xterm.js 渲染输出并接收 OSC 777 命令状态事件。`command-start` 会设置终端的 `currentCommand`，`command-end` 会清空它；侧边栏终端名称显示规则是 `currentCommand || shellName`。

当前 bash 和 PowerShell 集成都用内部状态记录“是否已经开始过命令”，因此只有真实命令结束时才发 `command-end`。zsh 集成直接在每次 `precmd` 发 `command-end`，这会覆盖两类正常场景：

```
zsh 启动
  └─ 初始 prompt
       └─ precmd -> command-end  // 没有对应 command-start

launch profile
  ├─ 前端设置 currentCommand = "codex"
  ├─ 延迟到达的初始 command-end 清空 currentCommand
  └─ UI 回退显示 shellName = "zsh"
```

问题的本质是 shell 集成协议发出了无配对的结束事件，而不是 UI 显示规则错误。

## Goals / Non-Goals

**Goals:**

- zsh 集成只在有对应 `command-start` 的情况下发 `command-end`。
- launch profile 设置的命令标签不会被 zsh 初始/空闲 prompt 的结束事件清空。
- 保持 zsh 真实命令执行时的 `command-start`/`command-end` 行为。
- 通过单元测试覆盖 zsh 集成脚本和前端 launch profile 标签竞态。

**Non-Goals:**

- 不重做 OSC 777 协议。
- 不改变 terminal row 的显示优先级。
- 不改变 bash、PowerShell、cmd 的集成策略。
- 不把 launch profile 改为后端直接执行命令。

## Decisions

### 在 zsh 集成源头 gate `command-end`

`zshIntegrationScript()` 增加 `__tui_helper_command_started=0` 状态。`preexec` 发出 `command-start` 后设置为 `1`；`precmd` 只有在状态为 `1` 时才发 `command-end`，随后重置为 `0`。

这样 zsh 与 bash/PowerShell 的语义一致：

```
preexec(command)
  ├─ emit command-start(command)
  └─ command_started = 1

precmd()
  ├─ if command_started == 1:
  │    ├─ emit command-end
  │    └─ command_started = 0
  └─ otherwise no-op
```

备选方案是在前端忽略“当前没有命令标签”的 `command-end`。这能缓解症状，但会让前端承担 shell 协议纠错，并且不能阻止后端历史/其他消费者看到错误事件。因此首选在 zsh hook 源头修复。

### 前端保持真实 `command-end` 清空标签

真实命令结束后 UI 应恢复 shell 名称，这是现有测试和交互预期。实现不应让所有 `command-end` 都失效，只应消除 zsh 产生的无配对事件。

如果实现阶段发现仍存在非 zsh 来源的无配对事件，可增加一个小范围前端防御：只忽略 launch profile 设置后、尚未收到 command-start 的首个空闲 `command-end`。该防御不作为主路径，避免改变普通命令结束语义。

### 测试以协议和用户可见行为双层覆盖

后端测试验证 zsh 脚本包含 command-start gating，并与 bash/PowerShell 保护策略保持一致。前端测试验证 launch profile 设置标签后收到一次无配对 `command-end` 时不会立刻退回 shell 名称；收到真实 `command-start` 后的 `command-end` 仍会清空标签。

## Risks / Trade-offs

- zsh hook 状态变量与用户环境变量同名冲突 -> 使用带项目名前缀的私有变量，降低冲突概率。
- 某些 zsh 配置可能自定义 hook 顺序 -> 使用 `add-zsh-hook` 继续与现有集成方式一致，只调整自身 hook 内部行为。
- 前端测试如果只模拟 `command-end` 可能与真实 shell 事件差异较大 -> 同时覆盖 launch profile 用户路径和命令状态事件路径。
- 修复 zsh 后仍可能被 Codex 自己的标题变化影响活动状态 -> 本变更只处理名称回退；标题 activity 已由现有 `embedded-terminal-emulation` 逻辑处理。
