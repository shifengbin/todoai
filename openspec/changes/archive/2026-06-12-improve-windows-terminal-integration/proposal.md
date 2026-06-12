## Why

Windows 平台下嵌入式终端已经能通过 ConPTY 启动 shell，但启动配置和活动状态仍按 Unix shell 行为假设处理，导致 `codex`/`claude` 需要用户再按一次回车、终端行不显示启动命令、进入后被误标记为 busy。

这会让 Windows 用户无法稳定使用内置 launch profile 工作流，也会让 TODO 终端树的状态提示失真。

## What Changes

- 修正 launch profile 自动提交命令的 Enter 语义，使 Windows ConPTY/PowerShell/cmd 下无需用户手动再按回车。
- 在应用主动提交 launch profile 命令时更新终端命令标签，让 `codex`、`claude` 或带参数的启动命令能立即显示在终端列表中。
- 为 Windows shell 补齐可行的命令状态集成，优先覆盖 PowerShell/pwsh，保持 zsh/bash 行为不回退。
- 调整终端标题活动状态判定，避免 Windows 路径中的 `\` 或 `/` 被误判为 spinner，从而进入终端后直接显示 busy。
- 保持 Windows 无 ConPTY 或 shell 启动失败时的现有 unsupported/startup error 语义。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `embedded-shell-sessions`: Windows launch profile 命令提交、命令标签和 shell 命令状态集成需要满足跨平台终端会话行为。
- `embedded-terminal-emulation`: 终端标题变化到活动状态的映射需要避免 Windows 路径误判，同时继续支持交互式程序的 busy/needs-input 提示。

## Impact

- 后端 shell 启动与集成脚本：`shell.go`、`shell_windows.go` 及相关测试。
- 前端 launch profile 提交与终端运行时标签：`frontend/src/App.vue` 及组件测试。
- xterm 标题和 OSC 命令状态处理：`frontend/src/xtermFactory.js`、`frontend/src/terminalManager.js` 及前端单元测试。
- 平台影响：Windows 行为修复为主；Unix zsh/bash 现有集成应保持兼容。
