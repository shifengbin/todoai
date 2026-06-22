## 1. 测试覆盖

- [x] 1.1 调整现有“选择 TODO project 恢复视图”的前端测试，使其改为验证 workspace 打开/重开路径仍恢复保存的 TODO 视图。
- [x] 1.2 新增前端回归测试：在 `执行中` 视图选择终端后，切到 `未执行`，再点击当前 `未执行` 列表中的项目 item，视图仍保持 `未执行`。
- [x] 1.3 运行相关客户端测试并确认新增测试先能覆盖当前回跳行为。

## 2. 前端实现

- [x] 2.1 调整 `frontend/src/App.vue` 的 TODO 项目 item 选择路径，使 `selectTodoProject()` 应用后端状态时不恢复 TODO 工程保存的视图标签。
- [x] 2.2 保留 workspace 打开、重新打开和前端重新加载路径中的 TODO 工程视图恢复逻辑。
- [x] 2.3 确认选择终端、创建终端、终端聚焦、active TODO 工程和左侧栏宽度行为不发生回归。

## 3. 验证与交付

- [x] 3.1 运行前端自动化测试，至少覆盖 `frontend/src/App.test.js` 中的 TODO 工作区相关用例。
- [x] 3.2 运行 OpenSpec 严格校验：`openspec validate fix-todo-project-selection-view-jump --strict`。
- [x] 3.3 执行自动 review，检查实现是否符合规格、是否存在无关改动或测试缺口。
- [x] 3.4 运行 `wails build -tags webkit2_41` 生成可执行文件。
