## Why

当前应用已经能在 Windows 上探测 PowerShell、Cmd、pwsh 等 shell，但嵌入式终端仍依赖 `creack/pty`，在 Windows 上只能进入 unsupported 状态。Windows 用户无法在应用内直接使用项目终端，这与已有终端设置和 TODO 项目工作流不一致。

本变更为 Windows 10 1809+ 和 Windows 11 增加 ConPTY 后端，让受支持的 Windows 环境可以运行真正的内嵌终端，同时保留旧系统的稳定不可用兜底。

## What Changes

- 为 Windows 增加 ConPTY-based PTY 后端，实现与现有 `PtyProcess` 接口兼容的启动、读写、resize、wait 和 close 行为。
- 将平台相关 PTY 创建拆分为 Unix 与 Windows 实现：Unix/Linux/macOS 继续使用当前 `creack/pty`，Windows 使用 Go ConPTY 依赖。
- Windows 10 1809+ / Windows 11 上，已探测或已保存的 PowerShell、Cmd、pwsh 等 shell 可以作为嵌入式终端启动。
- 旧 Windows、ConPTY API 不存在、依赖初始化失败或平台能力不足时继续返回 `ErrEmbeddedShellUnsupported`，前端保持现有 unsupported 提示和不自动重试行为。
- 保持现有 Wails API、前端 xterm.js 集成、终端状态事件和设置持久化格式不变。

## Capabilities

### New Capabilities

### Modified Capabilities

- `embedded-shell-sessions`: Windows 受支持系统上的嵌入式 shell 会话从不可用变为可启动、可交互、可 resize；旧系统仍保持 unsupported 状态。

## Impact

- 后端终端启动层：`NewPtyProcess`、`PtyProcess` 平台实现、错误映射和 resize/close 行为。
- 依赖：新增一个 Go ConPTY 依赖，优先通过本地 adapter 封装，避免第三方 API 扩散到业务层。
- 测试：补充 Windows 后端 adapter 单元测试、现有 session manager 状态测试，并保留 Windows 目标交叉编译检查。
- 前端：原则上无需改变；现有 terminal-output、SendTerminalInput、ResizeTerminal 和 unsupported 状态 UI 继续复用。
