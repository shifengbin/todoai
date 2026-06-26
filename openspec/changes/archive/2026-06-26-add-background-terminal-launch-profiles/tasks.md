## 1. 测试先行

- [x] 1.1 为 terminal settings 增加 Go 单元测试，覆盖 `background` 字段保存、读取、旧配置默认前台、enabled 缺失兼容。
- [x] 1.2 为 App 后台启动 API 增加 Go 单元测试，覆盖 TODO 项目 worktree 目录、任务工作区目录、not-started TODO 拒绝、worktree failed 拒绝、启动失败不新增终端。
- [x] 1.3 为前端设置面板增加自动化测试，覆盖后台启动控件渲染、编辑保存 payload、旧 profile 默认前台。
- [x] 1.4 为前端下拉菜单增加自动化测试，覆盖后台 profile 点击后调用后台 API，且不调用 `CreateTodoTerminal`/`CreateTaskTerminal`/`SendTerminalInput`、不创建 xterm session、不改变 active terminal。

## 2. 后端实现

- [x] 2.1 扩展 `TerminalLaunchProfileSetting` 和 settings normalize/unmarshal 逻辑，持久化 `background` 字段并兼容旧配置。
- [x] 2.2 新增后台命令 request/runner 抽象，支持注入测试替身，真实实现复用 `newBackgroundCommand` 并在 goroutine 中 `Wait()` 回收。
- [x] 2.3 实现配置 shell 的一次性执行参数解析，覆盖 zsh/bash/sh、PowerShell、cmd，并避免使用交互式 `IntegratedShellLaunch`。
- [x] 2.4 新增 TODO 项目后台启动 API，复用 in-progress 校验、worktree ready 校验和 worktree 工作目录。
- [x] 2.5 新增任务级后台启动 API，复用 in-progress 校验和任务工作区目录创建逻辑。

## 3. 前端实现

- [x] 3.1 更新 terminal settings profile 表单，增加后台启动开关，并在 clone/normalize/save 流程中保留 `background`。
- [x] 3.2 更新 `ProjectSidebar` launch option 传递，确保 profile 的 `background` 状态随点击事件传到 `App.vue`。
- [x] 3.3 更新 `App.vue` 启动分流：前台 profile 维持现有创建终端并发送命令，后台 profile 调用对应后台启动 API 并保持 UI 状态不变。
- [x] 3.4 重新生成 Wails 前端绑定和模型，使新增后台启动 API 与 `TerminalLaunchProfileSetting.background` 可供前端使用。

## 4. 验证与交付

- [x] 4.1 运行 `go test ./...`。
- [x] 4.2 运行 `cd frontend && npm run test`。
- [x] 4.3 运行 `cd frontend && npm run build`。
- [x] 4.4 运行 OpenSpec 状态/校验命令，确认 change 可 apply 且 spec delta 结构有效。
- [x] 4.5 执行自动 review，检查后台启动不影响 UI 状态、终端历史、agent status 和 worktree 隔离约束。
- [x] 4.6 运行 `wails build -tags webkit2_41`。
