## Context

应用当前通过 Wails 后端管理嵌入式 shell session，前端使用 xterm.js 渲染终端并通过 `SendTerminalInput`、`ResizeTerminal`、`terminal-output`、`terminal-status` 与后端通信。后端核心边界已经存在：`ShellSessionManager` 只依赖 `ShellStarter` 创建 `PtyProcess`，而 `PtyProcess` 暴露 `Read`、`Write`、`Resize`、`Wait`、`Close`。

当前真实 PTY 实现直接使用 `creack/pty`。该依赖在 Windows 上不能提供 ConPTY session，启动失败会被映射为 `ErrEmbeddedShellUnsupported`，终端进入 `unsupported` 状态。此前 Windows shell 探测和保存已经支持 PowerShell、Cmd、pwsh 等路径，但探测成功不等于嵌入式终端可运行。

本变更利用现有 `PtyProcess` 边界增加 Windows ConPTY 后端，让受支持 Windows 系统具备与 Unix PTY 类似的内嵌终端能力。

## Goals / Non-Goals

**Goals:**

- Windows 10 1809+ 和 Windows 11 上可以启动真正的嵌入式终端 shell。
- 继续复用现有 `ShellSessionManager`、Wails API、前端 xterm.js 和终端状态事件。
- Unix/Linux/macOS 行为保持不变，继续使用 `creack/pty`。
- Windows ConPTY 不可用时继续返回 `ErrEmbeddedShellUnsupported`，前端保持稳定 unavailable 状态且不自动重试。
- 将第三方 ConPTY 依赖封装在本地 adapter 中，避免业务层直接绑定实验性 API。
- 保留 Windows 目标交叉编译检查，并补充可在非 Windows CI 中覆盖的错误映射和接口级测试。

**Non-Goals:**

- 不实现外部终端 fallback。
- 不改变终端设置持久化格式、shell 探测优先级或 launch profile 数据结构。
- 不为 PowerShell/Cmd 增加命令开始/结束集成脚本；命令状态不可用时继续使用已有 fallback 标签行为。
- 不支持低于 Windows 10 1809 的系统；这些环境继续显示 unsupported 状态。
- 不重写前端终端渲染、布局或快捷键逻辑。

## Decisions

### 保持 `PtyProcess` 接口，按平台拆分实现

`ShellSessionManager` 已经只依赖 `ShellStarter` 和 `PtyProcess`，这是合适的平台边界。实现上应把当前 `NewPtyProcess` 拆为平台文件：Unix 使用 `creack/pty`，Windows 使用 ConPTY adapter。这样启动、输出读取、输入写入、resize、退出等待和关闭流程都能继续走现有 session manager。

备选方案是新增更大的 `TerminalBackend` 抽象。它长期更显式，但当前会扩大改动面，并迫使已有 session 状态、回调和测试迁移。现有接口已经能表达本次需要的进程生命周期，因此不增加新抽象。

### Windows 后端使用 Go ConPTY 依赖并本地封装

Windows 实现使用 `github.com/UserExistsError/conpty` 并只在 Windows adapter 内部引用。adapter 对外只返回本项目的 `PtyProcess`。这样如果后续依赖 API 变化或需要替换依赖，影响范围限制在平台实现文件。

备选方案一是使用 `github.com/charmbracelet/x/conpty`。该包维护信号更好，但当前可用版本要求 `go 1.24`，会把本项目的 `go 1.23` directive 一并升级，影响面超出本变更需要。备选方案二是直接基于 `x/sys/windows` 调用 `CreatePseudoConsole` 等 API。这样依赖更少，但需要自行处理管道、进程属性、resize、关闭顺序和错误细节，初次实现风险更高。

### ConPTY 不可用继续映射为 unsupported

Windows adapter 应将 ConPTY API 不存在、系统版本不足、初始化失败且语义为“不支持”的错误映射为 `ErrEmbeddedShellUnsupported`。`ShellSessionManager.StartTerminal` 已经能把该错误转为 `ShellStateUnsupported`，并避免自动重试。非能力问题，例如 shell path 不存在、工作目录错误或启动参数错误，应继续作为启动错误暴露，不应伪装为 unsupported。

备选方案是把所有 Windows 启动失败都归为 unsupported。这样用户看到的状态稳定，但会隐藏配置错误，降低调试能力。

### 保持前端和 Wails API 不变

现有前端已经把 xterm 输入输出、resize、terminal status 和 unsupported overlay 建模清楚。ConPTY 后端只需要满足相同事件和方法契约，不需要新增前端 API。Windows 成功启动时终端状态为 `running`，失败且不支持时仍为 `unsupported`。

备选方案是前端先做平台能力探测，再决定是否允许创建终端。这样可以提前禁用按钮，但会新增前后端能力 API；当前后端启动阶段已经可以可靠判断能力，先保持行为集中在后端更简单。

### 测试分层覆盖

通用 session 行为继续通过注入 fake `ShellStarter` 测试，不依赖真实 PTY。Windows adapter 使用平台文件和小边界覆盖错误映射、resize 参数传递、close/wait 生命周期。Linux CI 通过 `GOOS=windows GOARCH=amd64 go build ./...` 和 Windows 测试包编译确认平台代码可编译；真实 ConPTY 交互作为 Windows 手动验证项。

备选方案是在非 Windows CI 中跳过所有 Windows 后端验证。这样实现快，但容易让平台文件、依赖 API 或 build tag 在合入后才暴露问题。

## Risks / Trade-offs

- [Risk] 选用的 ConPTY 依赖 API 可能仍处于实验状态。 -> 通过本地 adapter 封装第三方类型，只把本项目 `PtyProcess` 暴露给业务层。
- [Risk] 没有 Windows CI 时无法自动验证真实交互。 -> 保留交叉编译和可注入单元测试，并在 tasks 中加入 Windows 手动验证清单。
- [Risk] ConPTY close/wait 顺序处理不当可能导致进程残留或 goroutine 卡住。 -> adapter 需要明确 `Close` 幂等、`Wait` 只等待 shell 退出、删除终端时先 close process。
- [Risk] PowerShell/Cmd 不一定支持当前 zsh command-state 集成。 -> 本变更不承诺 Windows shell 命令状态集成；现有“状态不可用时显示 shell 名称”的行为继续生效。
- [Risk] 旧 Windows 或受限环境中 ConPTY API 不可用。 -> 保留 `ErrEmbeddedShellUnsupported` 和前端 unsupported UI，应用仍可使用其他功能。
- [Risk] Windows shell 路径和参数 quoting 与 Unix 不同。 -> 继续复用已有 platform-aware shell detection 和 `IntegratedShellLaunch` 输出，不在 ConPTY adapter 中重新解析用户命令。
