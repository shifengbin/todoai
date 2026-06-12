## Why

Windows 上通过内置终端打开 `codex` 或 `claude` 时，终端会显示大量类似 base64 的字符，影响交互可读性，也可能把内部控制协议写入终端历史。当前最可能的来源是 PowerShell 命令状态集成在 ConPTY 路径下输出的私有 OSC 777 payload 未被前端消费。

## What Changes

- 确保应用内部命令状态控制 payload 不会作为普通终端文本显示给用户。
- 确保泄漏的命令状态 payload 不会写入持久化终端历史，避免重新打开终端时重复出现。
- 调整 Windows `pwsh` / `powershell` 命令状态集成，使其在 ConPTY 下要么被可靠消费，要么安全降级为不发出会泄漏的 payload。
- 保留 zsh/bash 现有命令状态行为，以及 Windows launch profile 已有的应用侧命令标签 fallback。
- 增加覆盖 Windows PowerShell command-state payload 泄漏、终端渲染、历史存储和 fallback 标签的回归测试。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `embedded-terminal-emulation`: 内部命令状态控制 payload 不得显示为终端正文，也不得污染终端历史。
- `embedded-shell-sessions`: Windows PowerShell 命令状态集成必须在 ConPTY 下避免 payload 泄漏，并在无法可靠发出 lifecycle 事件时保持 launch profile 标签 fallback 可用。

## Impact

- Go shell 集成：`IntegratedShellLaunch`、PowerShell 临时脚本、Windows ConPTY 启动边界。
- Go 输出链路：PTY/ConPTY read loop、`TerminalOutputEvent`、终端历史追加。
- 前端终端解析：xterm.js OSC handler、terminal session 写入和历史回放。
- 测试：Go shell/session/history 单元测试，前端 xterm/terminal manager 测试，必要的 Windows 交叉编译验证。
