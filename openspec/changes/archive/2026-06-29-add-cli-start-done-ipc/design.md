## Context

TodoAI 是 Go + Vue + Wails 桌面应用。现有 CLI 入口已经能在 `main.go` 中绕过 GUI 执行 `claude-hook` 和 `list --done`，但任务开始/完成逻辑目前主要存在于 GUI 进程的 `App` 方法中：开始任务会调用 `ChangeTodoStatus(..., in-progress)`，并继续准备任务目录、worktree、初始化文件和初始化生命周期脚本；完成任务会调用 `CompleteTodo`，并处理完成生命周期脚本、终端清理、完成快照分支补全和生命周期状态清理。

本次变更要求 `todoai start` / `todoai done` 与页面按钮逻辑一致，并且执行后页面立即刷新。由于 CLI 是独立进程，直接修改 `projects.json` 无法操作 GUI 内存中的终端、生命周期脚本执行器，也无法直接发 Wails 前端事件。因此 CLI 需要把命令委托给正在运行的 GUI 进程执行。

## Goals / Non-Goals

**Goals:**

- `todoai start` / `todoai done` 在任务文件夹或其子目录执行时，不启动新的 Wails GUI。
- CLI 通过跨平台 IPC 请求已运行的 GUI 进程执行命令。
- GUI 进程复用现有开始/完成按钮对应的 App 方法，保证状态流转、生命周期脚本、worktree 准备、终端清理等行为一致。
- GUI 执行命令后复用现有 `workspace-state` 事件刷新页面。
- IPC 方案覆盖 Windows、macOS、Linux，并避免平台专属协议分支。

**Non-Goals:**

- 不支持 GUI 未运行时由 CLI 离线修改任务状态。
- 不新增远程网络 API；IPC 只绑定本机 loopback 地址。
- 不改变 `todoai list --done`、`todoai claude-hook` 或页面按钮本身的用户语义。
- 不在用户项目仓库中写入 IPC 元数据。

## Decisions

1. IPC 使用 `127.0.0.1` loopback HTTP/JSON。

   GUI 启动时使用 Go 标准库 `net/http` 在 `127.0.0.1:0` 监听随机端口，并将运行态信息写入 app config 目录，例如 `todoai-ipc.json`。运行态文件包含版本、地址、token、进程标识和创建时间。CLI 读取该文件后向 `POST /todo-command` 发送 JSON 请求。

   选择该方案是因为 Go 标准库在 Windows、macOS、Linux 上行为一致，不需要分别实现 Unix Domain Socket、Windows Named Pipe 或平台特定通知机制。备选方案是文件监听 `projects.json`，但它不能保证“同页面按钮逻辑”；另一个备选是 Unix Socket / Named Pipe，但会引入多平台分支和测试成本。

2. IPC 请求使用本机 token 认证。

   GUI 每次启动生成随机 token，并写入运行态文件。CLI 请求携带 token，GUI 校验后才执行命令。服务只绑定 `127.0.0.1`，运行态文件在 Unix 平台尽量使用 `0600` 权限，Windows 依赖用户配置目录 ACL。GUI 退出时关闭 listener，并仅在运行态文件仍属于当前 token 时删除文件，避免误删新实例信息。

   该设计不把 IPC 暴露给局域网，也降低同机其它进程误触发命令的风险。它不是强安全边界，但足以匹配本机桌面应用的控制面需求。

3. CLI 只做分发，不复制按钮业务逻辑。

   `runCLICommand` 增加 `start` 和 `done` 分支。CLI 归一化当前工作目录后，把 `command` 和 `workingDir` 发给 GUI。CLI 不直接调用 `ProjectManager.ChangeTodoStatus` 或 `ProjectManager.CompleteTodo`，也不直接写 `projects.json`。

   这样可以避免 CLI 与 GUI 出现两套状态流转实现。若 GUI 未运行、运行态文件缺失、连接失败或响应错误，CLI 返回非零退出码并在 stderr 输出明确错误。

4. GUI 在当前 workspace 中解析任务文件夹。

   IPC handler 在 GUI 进程中读取当前 workspace state，按当前 workspace 的 `tasks/<workspaceDirName>` 路径匹配请求中的 `workingDir`。匹配应支持任务文件夹本身及其任意子目录。找到 TODO 后：

   - `start` 调用 `ChangeTodoStatus(todoID, TodoStatusInProgress)`。
   - `done` 调用 `CompleteTodo(todoID)`。

   如果当前 GUI 没有打开 workspace、目录不属于当前 workspace 的任务文件夹、任务不存在或状态不满足现有按钮方法约束，则返回错误。该约束让 CLI 行为与用户当前看到的页面上下文保持一致。

5. 页面刷新复用现有 Wails 事件。

   IPC handler 调用 App 方法成功后，GUI 进程发送 `workspace-state` 事件给前端。对于 `done` 存在完成生命周期脚本的情况，行为保持与页面按钮一致：命令成功代表 GUI 已接受并启动完成流程，最终完成状态仍由生命周期脚本成功后的既有回调发出。

## Risks / Trade-offs

- 运行态文件过期或 GUI 崩溃后残留 → CLI 连接失败时返回“TodoAI 页面未运行或不可达”，并可在后续实现中清理明显不可达的运行态文件。
- 同一用户同时启动多个 GUI 实例 → app config 下的运行态文件以后写入者为准；CLI 委托给最近启动的实例。若该实例未打开对应 workspace，会返回上下文不匹配错误。
- 本机其它进程读取 token 后可发命令 → 绑定 loopback、随机 token、用户配置目录权限共同降低风险；不把该接口设计为远程控制 API。
- `done` 的完成生命周期脚本可能异步运行 → CLI 成功只表示与页面按钮相同地提交了完成动作，最终 completed 状态仍遵循现有生命周期脚本结果。
- HTTP IPC 比平台原生 IPC 略多一个本地端口 → 换来统一实现和统一测试，且只监听 loopback 随机端口，运维复杂度较低。

## Migration Plan

- 新增 IPC server 和 client 后保持默认 GUI 启动流程兼容。
- CLI 新命令只在显式执行 `todoai start` / `todoai done` 时启用，不影响现有用户数据。
- 若出现问题，可移除 CLI 分支和 IPC 启动逻辑，现有页面按钮、`list --done`、`claude-hook` 不受数据迁移影响。

## Open Questions

- 无。当前设计按“CLI 必须与页面开始/完成按钮使用同一套逻辑”执行。
