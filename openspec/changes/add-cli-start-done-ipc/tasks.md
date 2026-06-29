## 1. IPC 基础设施

- [x] 1.1 定义 TodoAI IPC 运行态文件结构、请求/响应结构、运行态文件路径和随机 token 生成逻辑。
- [x] 1.2 实现 GUI 进程的 loopback HTTP server，监听 `127.0.0.1:0`，启动后写入 app config 下的运行态文件。
- [x] 1.3 实现 IPC server 关闭逻辑，在 App shutdown 时关闭 listener，并仅在 token 匹配时删除当前运行态文件。
- [x] 1.4 实现 IPC 请求 token 校验、JSON 解码、超时和统一错误响应。
- [x] 1.5 实现 CLI 侧 IPC client，读取运行态文件并向 GUI 发送 `start` / `done` 命令请求。

## 2. CLI 命令分发

- [x] 2.1 在现有 CLI runner 中新增 `todoai start` 和 `todoai done` 分支，并保持空参数启动 GUI、`claude-hook`、`list --done` 行为不变。
- [x] 2.2 CLI 请求中传递归一化后的当前工作目录，成功时返回 0，失败时返回非 0 并向 stderr 输出明确错误。
- [x] 2.3 为 GUI 未运行、运行态文件缺失、IPC 连接失败、认证失败和不支持命令提供稳定错误信息。

## 3. GUI 命令执行

- [x] 3.1 在 GUI 进程中实现按当前 workspace 的 `tasks/<workspaceDirName>` 匹配工作目录到 TODO 的 helper，支持任务文件夹及其子目录。
- [x] 3.2 IPC `start` handler 解析 TODO 后调用 `ChangeTodoStatus(todoID, TodoStatusInProgress)`，复用页面“开始”按钮后端逻辑。
- [x] 3.3 IPC `done` handler 解析 TODO 后调用 `CompleteTodo(todoID)`，复用页面“完成”按钮后端逻辑。
- [x] 3.4 IPC handler 成功后调用现有 `emitWorkspaceState` 推送最新 `ProjectState`，确保页面立即刷新。
- [x] 3.5 保持完成生命周期脚本的现有异步语义：IPC 成功表示完成流程已被页面同逻辑接受，最终 completed 状态由既有脚本回调决定。

## 4. 自动化测试

- [x] 4.1 添加 IPC 运行态文件、token 校验、server start/stop 和 stale runtime file 场景的 Go 单元测试。
- [x] 4.2 添加 CLI `start` / `done` 分发测试，验证命令不启动 Wails GUI，并正确调用 fake IPC server。
- [x] 4.3 添加工作目录到 TODO 的解析测试，覆盖任务文件夹、任务子目录、未知目录和当前 workspace 不匹配。
- [x] 4.4 添加 GUI IPC handler 测试，验证 `start` 与 `done` 调用现有 App 方法并触发 workspace state 刷新路径。
- [x] 4.5 添加状态拒绝场景测试，验证不允许的 `start` / `done` 返回错误且不直接修改持久化状态。
- [x] 4.6 运行后端自动化测试：`go test ./...`。
- [x] 4.7 运行客户端自动化测试：`cd frontend && npm test`。

## 5. 验证与收尾

- [x] 5.1 运行 OpenSpec 校验，确认 `add-cli-start-done-ipc` 的 proposal、design、specs、tasks 可用于 apply。
- [x] 5.2 执行自动代码 review，重点检查跨平台 IPC、token 校验、GUI 业务复用、错误处理和测试覆盖。
- [x] 5.3 根据 review 结果修复问题或记录明确的剩余风险。
- [x] 5.4 运行最终回归测试：`go test ./...` 和 `cd frontend && npm test`。
- [x] 5.5 运行 `wails build -tags webkit2_41` 生成可执行文件。
