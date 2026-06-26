## Context

终端 launch profile 当前保存在全局 terminal settings 中，前端在 TODO 项目和任务级终端入口的下拉菜单中展示启用的 profile。用户选择自定义 profile 时，前端先调用后端创建一个嵌入式终端，再把 profile 命令通过 `SendTerminalInput` 写入该终端。这个流程适合交互式工具，但不适合只需要一次性执行的后台命令。

现有后端已经区分 TODO 项目 worktree 终端和任务级终端：项目级终端必须在 prepared worktree 中启动，任务级终端在 TODO 任务工作区中启动。后台启动必须复用同样的上下文约束，但不能注册 `ProjectTerminal`、不能写 terminal history，也不能触发 xterm 或 terminal status 事件。

## Goals / Non-Goals

**Goals:**

- 允许用户在 launch profile 中配置是否后台启动。
- 后台启动 profile 在现有下拉菜单中可选，但选择后不新增 UI 终端、不改变当前终端、不收集输出。
- 后台命令在正确 TODO 上下文目录中一次性执行，并在命令结束后自动回收进程资源。
- 保持旧 settings 文件兼容，缺少后台启动字段的 profile 按前台启动处理。

**Non-Goals:**

- 不提供后台任务列表、进度、输出查看、停止按钮或重试队列。
- 不把后台命令接入 Claude/Codex agent status 监控。
- 不承诺后台命令脱离应用生命周期继续运行。
- 不改变内置 `Terminal` 选项；它仍然只创建普通嵌入式终端。

## Decisions

1. 在 `TerminalLaunchProfileSetting` 上新增 `Background bool` 字段。

   旧配置反序列化时缺少该字段会得到 `false`，等价于现有前台启动行为。保存时将后台状态随 name、command、enabled 一起持久化。相比新增独立的后台 profile 列表，这个方案能保留现有排序、启用/禁用和菜单结构。

2. 前端根据 profile 的 `background` 字段分流启动路径。

   `ProjectSidebar` 继续把用户选中的 profile 作为 option emit 给 `App.vue`。`App.vue` 若发现 `background === true`，则调用新的后台启动 API；否则沿用 `CreateTodoTerminal/CreateTaskTerminal` + `SendTerminalInput`。这样下拉菜单组件不需要理解后端细节，也避免改变普通 profile 行为。

3. 后端新增后台命令 API，不复用终端创建 API。

   项目级后台启动应校验 TODO 为 `in-progress`、确保 TODO 项目 worktree ready，并以 worktree path 作为工作目录。任务级后台启动应校验 TODO 为 `in-progress`，确保任务工作区目录存在，并以任务工作区作为工作目录。API 成功返回前只保证进程已成功启动；命令后续退出不回写 UI 状态。

4. 后台命令使用配置 shell 的一次性命令模式执行。

   后台命令不使用 `IntegratedShellLaunch`，因为交互式集成会注入 rc/profile 脚本，部分 shell 还会使用 `-i` 或 `-NoExit`，不符合“运行结束后自动退出”。实现应按 shell 类型选择一次性执行参数，例如 Unix shell 使用 `-lc <command>`，PowerShell 使用 `-NoLogo -ExecutionPolicy Bypass -Command <command>`，cmd 使用 `/C <command>`。未知 shell 可使用保守的一次性参数或返回明确错误。

5. 后台 runner 独立封装并可注入测试替身。

   后台命令启动应通过小接口封装工作目录、shell path、args、env 和 wait 行为。真实实现复用 `newBackgroundCommand`，以便 Windows 隐藏 console window；测试实现记录请求并可模拟启动错误。启动成功后真实 runner 在 goroutine 中等待进程退出并释放资源。

## Risks / Trade-offs

- 后台命令没有输出和状态反馈，用户难以确认长时间任务是否仍在运行。→ 第一版严格满足“不影响 UI”，只在启动失败时显示错误；后续如需要可新增独立后台任务视图。
- shell 一次性执行参数在非主流 shell 上可能不兼容。→ 支持当前 shell 检测覆盖的主流 shell；未知 shell 启动失败时返回明确错误，避免创建无法退出的交互式进程。
- 后台命令不带 terminal identity，Claude/Codex hook 无法关联 UI 终端。→ 这是刻意行为，避免后台命令影响终端列表和 agent status。
- 应用退出时后台进程生命周期未对用户可见。→ 本变更不提供后台任务管理；实现保持当前应用内启动和回收语义。

## Migration Plan

- settings schema 保持同一版本或通过普通兼容读取迁移；旧 profile 缺少 `background` 字段时按 `false` 处理。
- Wails 绑定更新后前端调用新的后台启动 API。
- 回滚时旧应用会忽略或丢弃未知 `background` 字段的行为取决于保存路径；不影响普通终端启动。

## Open Questions

- 无。后台进程管理、输出查看和跨应用生命周期保活均不在本变更范围内。
