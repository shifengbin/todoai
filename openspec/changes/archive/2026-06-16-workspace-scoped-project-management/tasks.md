## 1. Workspace 数据模型与迁移

- [x] 1.1 新增 workspace/recent 数据结构，包含当前 workspace、最近打开记录、规范化路径和 `.data` 路径解析。
- [x] 1.2 实现全局 `recent-workspaces.json` 读写、去重、倒序排序和清空逻辑。
- [x] 1.3 实现打开 workspace 时的目录校验、`.data` 创建和失败不切换当前 workspace 的保护。
- [x] 1.4 实现旧全局 `projects.json`、`terminal-history.json` 到 app-managed legacy workspace 的一次性兼容迁移，并保留 `settings.json` 为应用全局配置。
- [x] 1.5 为 workspace 打开、最近打开、清理最近打开、旧数据迁移和路径不可用场景添加 Go 单元测试。

## 2. 后端 Workspace 作用域切换

- [x] 2.1 调整 `App` 初始化，使项目库和终端历史 manager 根据当前 workspace `.data` 创建或重建，设置 manager 保持全局。
- [x] 2.2 修改项目库、TODO 和终端历史 API，在没有当前 workspace 时返回空状态或明确错误；设置 API 保持全局可用。
- [x] 2.3 实现关闭和切换 workspace 前的 runtime terminal shutdown、terminal history flush 和 active context 清理。
- [x] 2.4 确保 Git 状态查询、Git 初始化和 TODO 相关操作只作用于当前 workspace 数据。
- [x] 2.5 为 workspace 数据隔离、关闭清理终端、切换 workspace 后状态刷新添加 Go 单元测试。
- [x] 2.6 启动时恢复最近打开列表中的第一项；若该 workspace 不可访问则保持无 workspace 空态且不回退到其他最近项。

## 3. 文件菜单与 Wails 绑定

- [x] 3.1 增加 Wails 原生 `文件` 菜单，包含 `打开项目`、`最近打开`、`清理最近打开`、`关闭`。
- [x] 3.2 增加后端 workspace API 和菜单事件，用于打开目录、打开最近 workspace、清理最近打开、关闭当前 workspace、读取 workspace 状态。
- [x] 3.3 重新生成 Wails 前端绑定，确保新增 API 可从 Vue 调用。
- [x] 3.4 为菜单回调和 workspace API 的成功、取消、错误路径添加后端测试或可替代的集成覆盖。

## 4. 前端 Workspace 体验

- [x] 4.1 在 Vue 主状态中接入 `currentWorkspace` 和 `recentWorkspaces`，并让 `applyState` 能处理无 workspace 状态。
- [x] 4.2 实现无 workspace 空态，禁用或隐藏创建 TODO、导入工程、创建终端和 Git 初始化等 workspace 依赖操作。
- [x] 4.3 实现最近打开 workspace 选择面板，支持显示名称、路径、不可用错误和选择后打开。
- [x] 4.4 在打开、关闭、切换 workspace 后清理前端 terminal sessions、Git 状态、agent 状态和错误上下文。
- [x] 4.5 调整终端设置加载和保存流程，使设置来自应用全局配置，并且不随 workspace 切换重置。

## 5. 自动化测试与质量检查

- [x] 5.1 更新前端组件和状态管理测试，覆盖无 workspace 空态、最近打开面板、关闭/切换 workspace 后 UI 清理。
- [x] 5.2 运行客户端自动化测试并修复失败：`cd frontend && npm test -- --run`。
- [x] 5.3 运行 Go 测试并修复失败：`go test ./...`。
- [x] 5.4 运行 OpenSpec 校验并修复 proposal/spec/task 问题：`openspec status --change workspace-scoped-project-management`。
- [x] 5.5 执行自动 review，检查数据迁移、workspace 隔离、错误处理、并发关闭终端和用户数据不被删除的风险。
- [x] 5.6 完成 Wails 打包验证：`wails build -tags webkit2_41`。
