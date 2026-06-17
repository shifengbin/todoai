## 1. 后端 UI 状态持久化

- [x] 1.1 新增 TODO 工程 UI 状态数据结构，包含 `todoView`、`sidebarWidth` 和版本字段。
- [x] 1.2 新增 workspace-scoped store/manager，读写 `<workspace>/.data/todo-project-ui-state.json`，缺失或无效文件按空状态处理。
- [x] 1.3 在 `App` 打开、切换和关闭 workspace 时创建或切换 TODO 工程 UI 状态 store。
- [x] 1.4 暴露加载、保存和删除 TODO 工程 UI 状态的 Wails API，并校验无 workspace 时返回明确错误。
- [x] 1.5 在删除 TODO 工程和删除 TODO 时清理对应 TODO 工程 UI 状态。

## 2. 前端状态接线

- [x] 2.1 调整 `ProjectSidebar.vue`，支持由父组件传入当前 TODO 视图并在用户切换 `未执行`、`执行中`、`已完成` 时发出变更事件。
- [x] 2.2 调整 `App.vue`，在 active TODO 工程变化和 workspace 加载后恢复该 TODO 工程的 TODO 视图与左侧栏宽度。
- [x] 2.3 调整分割线拖拽逻辑，拖动过程中只更新内存宽度，拖拽结束后保存当前 TODO 工程的左侧栏宽度。
- [x] 2.4 处理默认值和异常路径：没有 active TODO 工程或没有持久化记录时使用 `not-started` 和默认左侧栏宽度。
- [x] 2.5 运行 Wails 绑定生成命令，确保前端 `wailsjs` 与新增 Go API 和模型一致。

## 3. 自动化测试

- [x] 3.1 添加后端测试覆盖 UI 状态文件读写、缺失/无效文件默认空状态和按 TODO 工程 ID 隔离。
- [x] 3.2 添加后端测试覆盖删除 TODO 工程和删除 TODO 时清理 UI 状态。
- [x] 3.3 添加前端测试覆盖切换 TODO 视图保存并在重新选择 TODO 工程时恢复。
- [x] 3.4 添加前端测试覆盖分割线宽度在拖拽结束后保存并按 TODO 工程恢复。
- [x] 3.5 运行 Go 测试和前端测试，确认相关测试通过。

## 4. Review 与打包

- [x] 4.1 运行自动 review 或静态检查，处理发现的代码质量和规范问题。
- [x] 4.2 运行 `openspec validate persist-todo-project-ui-state --strict`，确认变更规格有效。
- [x] 4.3 运行 `wails build -tags webkit2_41` 生成可执行文件。
