## Context

当前终端设置由 Go 后端负责。`SettingsManager` 在首次加载设置时调用 `ShellDetector` 自动探测 shell，并把结果保存到 `settings.json`；后续新终端通过 `ResolveShellPath` 读取已保存路径或 fallback。现有探测逻辑只面向 Unix：读取 `$SHELL`，再检查 `/bin/zsh`、`/usr/bin/zsh`、`/bin/bash`、`/usr/bin/bash`、`/bin/fish`、`/usr/bin/fish`、`/bin/sh`、`/usr/bin/sh`。路径可用性也使用 Unix execute bit。

生产代码当前可以交叉编译到 Windows，但 `creack/pty@v1.1.24` 的 Windows `StartWithSize` 返回 unsupported。因此本设计只覆盖 Windows shell 探测、保存和路径解析，不承诺 Windows 上的嵌入式 PTY 进程已经可运行。

## Goals / Non-Goals

**Goals:**

- 在 Windows 上自动探测合理的终端 shell。
- 在 Windows 上正确判断手动输入的 shell 路径是否可用。
- 让 `DefaultShellPath` 与设置探测共享平台语义，避免 Windows 上回退到 Unix 路径。
- 保持现有设置文件结构和 Wails API 不变。
- 补充可在 Linux CI 上运行的 Windows 行为单元测试。

**Non-Goals:**

- 不实现 Windows ConPTY 或替换现有 PTY 依赖。
- 不改变已保存设置优先于自动探测的语义。
- 不增加 per-project、per-terminal 或前端侧 shell 探测。
- 不为 PowerShell/Cmd 增加命令状态集成脚本。

## Decisions

### 将平台差异封装在后端 shell 探测层

`ShellDetector` 应扩展为平台感知组件，保留可注入的环境变量读取、候选路径和路径检查能力。实现上可以通过 `runtime.GOOS` 的默认值决定平台分支，并在测试里注入 Windows 平台名和 lookup/path-check 函数。

替代方案是在调用点直接判断 `runtime.GOOS`。这样会让 `SettingsManager`、`DefaultShellPath` 和测试各自复制平台规则，后续维护风险更高。

### Windows 优先探测本机 shell

Windows 默认候选顺序应优先 PowerShell 7 `pwsh.exe`，再 Windows PowerShell `powershell.exe`，再 `COMSPEC` 或 `cmd.exe`。固定路径应基于环境变量组合，例如 `SystemRoot/System32/WindowsPowerShell/v1.0/powershell.exe` 和 `SystemRoot/System32/cmd.exe`，同时允许 `exec.LookPath` 发现 PATH 中的 `pwsh.exe`。

`SHELL` 在 Windows 上不应排在最前，因为 Git Bash、MSYS 或开发工具可能设置它，但用户请求的是 Windows 系统 shell 探测。可以把 `SHELL` 作为低优先级兼容候选，避免完全忽略用户环境。

### 可用性校验按平台解释

Unix 继续使用 `os.Stat` + execute bit。Windows 使用 `os.Stat` 确认路径存在且不是目录，再按扩展名判断是否可执行：优先接受 `.exe`、`.cmd`、`.bat`、`.com`，并尊重 `PATHEXT`。这会让手动保存 `powershell.exe`、`cmd.exe` 或 `.cmd` wrapper 能通过验证。

替代方案是在所有平台都只检查文件存在。这样会放松 Unix 当前行为，可能让普通文本文件被保存为 shell。

### 默认 shell 路径复用探测策略

`DefaultShellPath` 不应继续硬编码 `/bin/sh` 作为所有平台兜底。它应调用平台感知探测器，探测失败时 Unix 仍可回退 `/bin/sh`，Windows 则回退到 `COMSPEC`、`cmd.exe` 或空错误路径以避免返回不可用的 Unix 路径。

替代方案是只改 `SettingsManager` 的首次探测。这样设置页可以显示 Windows shell，但某些错误路径或加载失败情况下新终端仍可能使用 Unix fallback。

### 明确 Windows PTY 启动边界

本变更不引入新依赖，`NewPtyProcess` 仍使用当前 `creack/pty`。规格和任务需要明确 Windows shell 探测通过后，Windows 上启动嵌入式终端仍可能因 PTY 后端 unsupported 而失败。后续完整运行支持应作为单独变更评估 ConPTY。

替代方案是在同一变更中引入 ConPTY。范围会跨依赖选型、进程 IO、resize、关闭和真实 Windows 验证，风险远大于 shell 探测。

## Risks / Trade-offs

- Windows 上 `SHELL` 与本机 shell 优先级冲突 -> 明确把 `SHELL` 放在低优先级兼容候选，优先 PowerShell/Cmd。
- `exec.LookPath` 在交叉平台单元测试中不可直接模拟 -> 为探测器保留 lookup/path-check 注入点，用 table tests 覆盖 Windows 顺序。
- Windows 可执行扩展判断可能漏掉少见 wrapper -> 支持 `PATHEXT`，并覆盖常见 `.exe/.cmd/.bat/.com`。
- 探测成功但 PTY 启动失败 -> 在设计、规格和任务中明确当前范围；后续用 ConPTY 变更解决运行时支持。
- 现有 `shell_test.go` 含 Unix-only `syscall.Kill` -> 拆出 Unix-only 测试文件或用 build tag，保证 Windows 目标测试至少能编译。
