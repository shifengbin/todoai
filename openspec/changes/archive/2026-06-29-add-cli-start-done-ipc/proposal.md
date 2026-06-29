## Why

用户需要在任务文件夹中通过 `todoai start` 和 `todoai done` 快速驱动当前任务流转，并让已打开的 TodoAI 页面立即刷新。现有页面按钮已经包含完整的开始/完成逻辑，CLI 入口必须复用同一套逻辑，避免命令行状态与桌面端状态不一致。

## What Changes

- 新增 `todoai start` 命令，可在当前任务文件夹中触发与页面“开始”按钮相同的任务开始逻辑。
- 新增 `todoai done` 命令，可在当前任务文件夹中触发与页面“完成”按钮相同的任务完成逻辑。
- CLI 通过跨平台 IPC 请求正在运行的 TodoAI GUI 进程执行操作，而不是在 CLI 进程中直接修改持久化文件。
- GUI 进程执行完成后继续通过现有 `workspace-state` 事件通知前端，确保页面立即刷新。
- 当 GUI 未运行、IPC 不可达、当前目录无法定位任务或任务状态不允许流转时，CLI 返回明确错误。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `todo-cli`: 增加从任务文件夹执行 `todoai start` / `todoai done` 并委托 GUI 执行同等按钮逻辑的命令行能力。

## Impact

- CLI 分发逻辑：新增 `start`、`done` 子命令，并保持 `list --done` 与 `claude-hook` 行为不变。
- GUI 后端：启动跨平台 IPC 服务，接收经过认证的本机命令请求并调用现有 App 方法。
- 工作区与任务定位：从 IPC 请求的当前目录解析对应任务和工作区上下文。
- 前端刷新：复用现有 Wails `workspace-state` 事件与前端 `applyState` 流程。
- 测试覆盖：增加 CLI 路由、IPC 请求处理、任务定位、成功状态流转、错误状态和刷新事件相关测试。
