## Why

当前终端 shell 探测只面向 Unix 环境：优先读取 `$SHELL`，再检查 `/bin/zsh`、`/bin/bash`、`/bin/sh` 等固定路径。Windows 用户首次加载终端设置时无法得到合理的 PowerShell 或 Cmd 候选项，手动保存 Windows shell 路径时也会受到 Unix 可执行权限判断影响。

这个变更让终端设置在 Windows 上能探测并验证本机 shell，为后续完整 Windows 嵌入式终端运行支持打下基础。

## What Changes

- 增加 Windows-aware 的 shell 探测顺序，覆盖 PowerShell 7 `pwsh.exe`、Windows PowerShell `powershell.exe`、`COMSPEC`/`cmd.exe`，并保留可配置候选路径。
- 调整 shell 路径可用性判断，让 Windows 使用存在性、非目录和可执行扩展/PATHEXT 语义，而 Unix 继续使用 execute bit。
- 让默认 shell 路径解析在 Windows 上使用同一套候选策略，避免回退到 `/bin/sh` 这类不可用路径。
- 保持现有设置持久化、手动保存、fallback 语义不变：已保存路径优先，不可用时提供自动探测 fallback。
- 明确当前范围只保证 Windows shell 探测和设置解析；完整 Windows 嵌入式 PTY 启动仍需要后续 ConPTY 或等价运行时支持。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `terminal-settings`: 终端 shell 自动探测和手动 shell 路径验证需要支持 Windows shell 与 Windows 可执行文件语义。
- `embedded-shell-sessions`: 新终端选择默认/fallback shell 时需要避免在 Windows 上解析出 Unix-only 路径，并明确 Windows PTY 启动能力边界。

## Impact

- Go 后端：`ShellDetector`、shell 路径可用性校验、默认 shell 路径解析、相关测试辅助注入点。
- Go 测试：增加 Windows 探测/校验的跨平台单元测试；整理 Unix-only 真实 PTY 进程测试，避免 Windows 目标测试编译失败。
- OpenSpec：更新 `terminal-settings` 和 `embedded-shell-sessions` 规格，加入 Windows shell 探测与启动边界要求。
- 依赖：本变更不引入新的 PTY 依赖；若后续实现完整 Windows 嵌入式终端启动，需要单独评估 ConPTY 方案。
