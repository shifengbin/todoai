## 1. 测试覆盖

- [x] 1.1 添加前端测试：用户拖动左侧 TODO 栏后创建 TODO，普通状态刷新不改变当前侧栏宽度。
- [x] 1.2 添加前端测试：当前视图为 `执行中` 时创建 TODO，普通状态刷新不改变当前 TODO 视图。
- [x] 1.3 添加前端测试：点击 `not-started` TODO 的开始按钮后，后端状态返回后仍停留在 `执行中` 视图并显示该 TODO。
- [x] 1.4 补充回归测试：打开 workspace 或主动选择 TODO 工程时仍恢复已保存的 TODO 工程 UI 状态。
- [x] 1.5 补充回归测试：同一个 TODO 下不同 TODO 工程分别恢复自己的左侧宽度，添加工程后切换 active TODO 工程时恢复目标工程宽度。

## 2. 前端实现

- [x] 2.1 调整 `frontend/src/App.vue` 的 `applyState` 选项，使普通 ProjectState 刷新默认不调用 TODO 工程 UI 状态恢复。
- [x] 2.2 在打开 workspace、加载 TODO 工程 UI 状态完成、主动选择 TODO 工程等显式恢复路径传入恢复选项。
- [x] 2.3 保留用户手动切换 TODO 视图和拖动侧栏宽度后的持久化逻辑，并确保显式恢复后继续适配活动终端尺寸。
- [x] 2.4 确认开始 TODO 的状态变更流程不会被后端返回的普通 ProjectState 覆盖当前 `执行中` 视图。
- [x] 2.5 在 `applyState` 中识别 active TODO 工程变化，按新的 `todoProjectId` 恢复对应 UI 状态。

## 3. 验证与交付

- [x] 3.1 运行前端自动化测试，至少覆盖 `frontend/src/App.test.js`。
- [x] 3.2 运行相关 Go 测试，确认后端 ProjectState 和 TODO 工作流行为未回归。
- [x] 3.3 执行自动 review，检查实现是否符合规格、设计和项目代码规范。
- [x] 3.4 运行 `openspec status --change "fix-todo-ui-state-refresh"`，确认 change 处于可实施/可归档状态。
- [x] 3.5 运行 `wails build -tags webkit2_41` 生成可执行文件。
