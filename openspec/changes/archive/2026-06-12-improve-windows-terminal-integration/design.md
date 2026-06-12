## Context

当前嵌入式终端由 Go 后端启动 PTY/ConPTY shell，前端通过 xterm.js 渲染输出并发送输入。Unix shell 集成只覆盖 zsh/bash：启动 shell 时注入 rc 脚本，脚本通过 OSC 777 发出 `command-start` 和 `command-end`，前端据此更新 `currentCommand`。

Windows 路径目前优先探测 `pwsh.exe`、`powershell.exe`，再退回 `cmd.exe`，并通过 ConPTY 启动。Windows 侧没有等价的命令状态集成；launch profile 命令由前端在创建终端后通过输入流写入；终端活动状态还会从 xterm 标题变化推断 busy/needs-input。

这次变更跨越 Go shell 启动、前端 launch profile 提交、xterm 标题状态和 TODO 终端树展示，因此需要把平台差异收敛到明确的终端会话契约上。

## Goals / Non-Goals

**Goals:**

- Windows ConPTY 下选择 `codex`、`claude` 或自定义 launch profile 后，命令应被提交执行，不需要用户再手动按回车。
- 启动 profile 创建的终端应立即显示启动命令标签，Windows 上不能因为缺少 shell hook 而长期显示 `pwsh`、`powershell` 或 `cmd`。
- PowerShell/pwsh 会话应尽量提供与 zsh/bash 一致的 `command-start`/`command-end` 事件。
- 终端标题活动状态应避免把 Windows 路径分隔符误判为 busy，同时继续识别明确的 spinner、working/thinking/running 类标题和等待输入信号。
- 保持现有 unsupported ConPTY、shell 启动失败、zsh/bash 集成和终端历史行为兼容。

**Non-Goals:**

- 不把 launch profile 改成替换 shell 进程的直接命令执行；启动命令仍运行在配置的 shell 内。
- 不要求 cmd.exe 提供完整命令生命周期 hook；cmd 至少不能阻塞 launch profile 标签和提交体验。
- 不引入新的终端后端或外部依赖。
- 不改变用户已保存的 launch profile 配置格式。

## Decisions

### 使用终端 Enter 语义提交 launch profile

前端提交 launch profile 时应发送交互式终端的 Enter 序列，而不是假设 `\n` 在所有 shell/PTY 中都等价。Windows ConPTY/PowerShell 下优先使用 `\r` 能更贴近真实键盘回车，Unix shell 也能接受 PTY 中的 carriage return。

备选方案是后端在 `WriteInput` 中按平台转换换行，但这会影响所有用户输入、粘贴和程序输入流，风险更大。将转换限制在自动提交 launch profile 的路径上，变更面更小。

### 应用侧先设置 launch profile 命令标签

launch profile 是由应用主动提交的命令，前端在提交前已经知道 profile 名称和命令文本。因此创建终端并提交命令时，应同步把该终端的 `currentCommand` 设置为可显示的命令标签。后续如果 shell 集成发出 `command-start` 或 `command-end`，继续以 shell 事件为准更新/清空状态。

备选方案是只为 Windows shell 补 hook，然后完全依赖 hook 更新标签。但 cmd 不适合稳定 hook，PowerShell hook 也可能受用户 profile 或 PSReadLine 可用性影响；应用侧乐观标签能覆盖启动 profile 的主要路径，并避免 UI 长时间显示 shell 名称。

### 为 PowerShell/pwsh 增加独立集成

`IntegratedShellLaunch` 应识别 `pwsh`/`powershell`，通过临时 profile 或启动参数注入最小集成脚本。脚本需要尽量保留用户原始 profile 行为，并发出与 zsh/bash 相同的 OSC 777 协议：

```
PowerShell prompt / command hooks
        │
        ├─ command-start  ──▶ OSC 777 tui-helper;command-start;<base64>
        └─ command-end    ──▶ OSC 777 tui-helper;command-end
```

优先设计为“可用则增强”：PowerShell/pwsh 应提供命令状态；cmd 不强制完整 hook，但仍应支持 profile 命令提交和应用侧标签。

### 标题活动判定不把路径分隔符当 busy

当前 busy 信号把 `/` 和 `\` 直接作为 spinner 字符处理，这在 Windows 标题中会把 `C:\repo\app` 误判为 busy。应把 busy 判定限定为明确 spinner 字符、重复变化的 spinner 状态，或包含 `working`、`thinking`、`running`、`processing`、`executing`、`busy` 等明确语义的标题。

`!` 仍可作为 needs-input 信号，但路径、shell 稳定标题和首次 idle 标题必须保持 idle。

## Risks / Trade-offs

- PowerShell hook 与用户 profile 交互复杂 -> 通过临时 profile/包装脚本隔离集成逻辑，并保留加载用户 profile 的路径；测试覆盖 profile 注入参数和清理。
- 应用侧乐观标签可能在命令很快退出时短暂显示旧命令 -> shell `command-end` 到达时清空；无 hook 的 cmd 会保留 launch profile 标签，优先保证用户能识别终端来源。
- 将 launch profile Enter 改为 `\r` 可能影响已有测试期望 -> 更新测试为“提交交互式 Enter”，并验证不影响普通手动输入和粘贴路径。
- 去掉 `/`/`\` busy 信号可能降低对某些简单 spinner 的识别 -> 保留更明确的 spinner 字符和文字信号，避免 Windows 高频误报。

## Migration Plan

- 不需要数据迁移；现有 settings 和 launch profile 配置继续有效。
- 发布后新建终端使用新提交和标签行为；已有非运行历史终端不回放命令状态。
- 如 PowerShell 集成失败，应退回为普通 shell session 加应用侧 launch profile 标签，而不是阻断终端启动。

## Open Questions

- cmd.exe 是否需要后续单独设计完整命令状态 hook；本次先不把它作为完成条件。
- PowerShell command-start 的实现应优先使用 PSReadLine hook 还是 prompt/trap 组合，需要在实现阶段通过单元测试和 Windows 手测确认。
