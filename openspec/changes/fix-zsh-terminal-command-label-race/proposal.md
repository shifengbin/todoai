## Why

在 zsh 终端中通过 launch profile 启动 `codex` 等命令时，终端名称偶尔会从命令名退回 `zsh`。根因是 zsh 集成会在初始 prompt 或空闲 prompt 发出没有对应 `command-start` 的 `command-end`，前端收到后清空 `currentCommand`。

这个问题会让用户难以区分正在运行的 agent 终端，尤其是在同一个 TODO 下同时运行多个终端时更明显。

## What Changes

- 修正 zsh shell 集成，只在本轮确实收到过 `preexec`/`command-start` 后才从 `precmd` 发出 `command-end`。
- 保持真实命令生命周期不变：命令开始时仍发 `command-start`，命令结束并回到 prompt 时仍发 `command-end`。
- 增加回归覆盖，证明 zsh 初始 prompt 不会发空闲 `command-end`，launch profile 标签不会被无配对结束事件清空。
- 不改变 bash、PowerShell、cmd 或普通用户输入的行为。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `embedded-shell-sessions`: zsh 命令状态事件必须成对反映真实命令生命周期，空闲 prompt 不得发出会清空命令标签的 `command-end`。
- `embedded-terminal-emulation`: 前端显示的运行中 launch profile 命令标签不得被 zsh 初始/空闲 prompt 的无配对结束事件重置为 shell 名称。

## Impact

- Go 后端 zsh 集成脚本：`shell.go` 中 `zshIntegrationScript()`。
- 后端测试：覆盖 zsh 集成脚本的 command-start/command-end gating。
- 前端命令标签测试：覆盖 launch profile 标签与无配对 `command-end` 的时序。
- 无数据迁移、无配置格式变化、无新增依赖。
