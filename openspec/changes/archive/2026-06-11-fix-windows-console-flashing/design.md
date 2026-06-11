## Context

当前应用是 Wails 桌面应用，前端在启动、项目切换、终端命令结束和窗口重新获得焦点时刷新当前项目的 Git 状态。导入父目录项目时，后端会把最后一个导入项目写入 `activeProjectId`，前端 `applyState` 会把这视为项目切换并立即触发 Git 查询。后端通过 `exec.CommandContext` 直接运行 `git -C <path> status --porcelain=v2 --branch` 和 `git -C <path> init`。在 Windows GUI 应用中启动控制台子进程时，如果没有隐藏窗口配置，系统可能为 `git.exe` 等短命令显示临时控制台窗口。

窗口获得焦点本身又会触发 Git 状态刷新。因此后台控制台窗口闪现并退出后，主窗口重新获得焦点，可能立即启动下一次 Git 查询，形成用户看到的连续闪烁。

另一个相关边界是嵌入式终端。项目已支持 Windows shell 探测和设置解析，但当前 PTY 后端仍使用 `creack/pty`，该库在 Windows 上对 `StartWithSize` 返回 `unsupported`，尚不能提供真正的 ConPTY 会话。应用需要把这个限制稳定地反馈给用户，而不是把 Windows shell 探测误表现为可启动终端。

## Goals / Non-Goals

**Goals:**

- Windows 上执行后台 Git 命令时不显示临时控制台窗口。
- 导入项目后不因为 `activeProjectId` 变化立即刷新 Git 状态。
- 用户展开 TODO、选择 TODO 项目或显式选择项目时再刷新当前可见/激活项目的 Git 状态。
- 窗口 focus 触发的 Git 状态刷新不会因为短时间焦点抖动而重复启动多次后台命令。
- Windows 上嵌入式 PTY 不可用时，终端创建或重启以稳定错误状态结束，并在界面中展示可理解提示。
- 保持 Linux/macOS 现有 Git 状态刷新、终端启动和测试行为。

**Non-Goals:**

- 不实现 Windows ConPTY 后端。
- 不改变 Git 状态解析格式或状态栏展示模型。
- 不引入每个项目独立的 Git 状态缓存，也不在展开 TODO 时批量查询所有子项目。
- 不移除窗口 focus 刷新能力，只限制重复触发和窗口闪烁。
- 不修改终端设置持久化格式。

## Decisions

### 封装后台命令创建并隐藏 Windows 控制台窗口

新增一个小的命令配置边界，用于创建 Git 后台命令后应用平台专属属性。Windows 分支设置 `SysProcAttr.HideWindow = true`，非 Windows 分支保持无操作。`runGitStatusCommand` 和 `runGitInitCommand` 复用该配置，避免未来只修一个 Git 命令路径。

备选方案是在每个 `exec.CommandContext` 调用点直接写 Windows build tag 代码。该方案改动少，但会把平台细节散落到业务函数里，后续新增后台命令时容易遗漏。

### 对 focus Git 刷新做短间隔去重

保留窗口获得焦点后刷新 Git 状态的行为，但增加前端级别的短间隔去重：如果同一项目已有 Git 状态请求进行中，或上一次 focus 刷新刚刚发生，则跳过本次 focus 刷新。项目切换和终端命令结束仍可触发刷新，避免影响用户期望的状态更新。

备选方案是完全移除 focus 刷新。这样最直接止血，但会降低用户从外部编辑器切回应用后看到最新 Git 状态的能力。

### 将导入项目状态应用与 Git 查询解耦

`applyState` 当前只知道前后 `activeProjectId` 是否变化，无法区分变化来源。新增一个调用层面的刷新策略参数或包装函数，让导入项目使用“应用状态但跳过 Git 刷新”的路径；用户显式选择项目、选择 TODO 项目、终端命令结束和允许的 focus 刷新继续使用刷新路径。

备选方案是在后端导入项目时不更新 `activeProjectId`。这能避免前端触发 Git 查询，但会改变项目管理器的选择语义，并影响导入后当前选中项目的既有体验。前端区分刷新来源更局部，也更容易测试。

### 由 TODO 展开触发当前项目 Git 懒加载

`ProjectSidebar` 目前把 TODO 收起/展开状态保存在组件内部，没有向 `App.vue` 暴露展开事件。新增一个轻量事件，例如 `todo-expanded`，在单个 TODO 从收起变为展开时发出。`App.vue` 收到事件后，如果展开的 TODO 包含当前激活的 TODO project 或展开操作导致用户选择了某个 TODO project，则刷新当前激活项目的 Git 状态。

不在本变更中为展开 TODO 的每个子项目批量查询 Git 状态。当前应用底部状态栏只有一个 `gitStatus`，对应 `activeProject`；批量查询需要引入 `gitStatusByProjectId` 缓存和新的 UI 展示契约，超出本次止血范围。

### Windows 终端不可用作为明确平台状态处理

在 Windows 上，当嵌入式 PTY 启动返回不支持时，后端应返回稳定、可识别的错误。前端捕获该错误后显示平台不支持提示，并保持终端列表和应用布局稳定。该行为不应自动循环重试；用户主动创建或重启终端时才会再次尝试。

备选方案是把 Windows shell 启动改为普通 `exec.Command` 并隐藏窗口。该方案不能提供 PTY、交互输入输出和终端尺寸控制，会破坏嵌入式终端语义，因此不适合作为兼容实现。

## Risks / Trade-offs

- [Risk] `HideWindow` 只能处理通过 Go `exec` 启动的子进程，不能隐藏由 Git hook 或外部程序再启动的独立窗口。 -> 当前修复覆盖应用直接启动的后台 Git 命令；若未来发现 hook 子进程闪窗，再针对 hook 环境做限制。
- [Risk] focus 去重可能跳过极短时间内的真实状态变化。 -> 项目切换和终端命令结束仍会刷新，focus 去重窗口应保持较短，仅用于抑制窗口闪烁循环。
- [Risk] 导入项目后状态栏可能暂时没有新导入项目的 Git 信息。 -> 这是刻意的懒加载行为；用户展开 TODO、选择 TODO 项目或显式选择项目时会刷新。
- [Risk] TODO 展开事件和自动展开 active terminal 的逻辑可能重复触发刷新。 -> 刷新层保留同项目请求去重，展开事件只作为懒加载信号。
- [Risk] Windows 用户仍无法使用嵌入式终端。 -> 本变更明确展示不可用状态，ConPTY 后端作为后续独立变更处理。
- [Risk] 平台专属行为在 Linux CI 中难以完整运行。 -> 添加可单元测试的平台配置函数，并保留 Windows 交叉编译检查。
