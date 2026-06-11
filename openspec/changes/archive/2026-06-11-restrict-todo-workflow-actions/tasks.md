## 1. 后端状态约束

- [x] 1.1 收紧 `ProjectManager.ChangeTodoStatus`，只允许 `not-started` TODO 切换为 `in-progress`
- [x] 1.2 收紧 `ProjectManager.CompleteTodo`，只允许 `in-progress` TODO 完成，并拒绝 `not-started` TODO 完成
- [x] 1.3 在 `App.CreateTodoTerminal` 创建终端前校验所属 TODO 为 `in-progress`
- [x] 1.4 保持删除、编辑和项目关联逻辑对 `not-started` 与 `in-progress` TODO 的现有可用性

## 2. 前端交互

- [x] 2.1 调整 TODO 行动作：`not-started` 显示开始和删除，不显示完成或退回未执行
- [x] 2.2 调整 TODO 行动作：`in-progress` 显示完成和删除，不显示退回未执行
- [x] 2.3 调整 TODO 项目行添加终端按钮，只在 TODO 为 `in-progress` 且项目可用时显示
- [x] 2.4 确认查看/编辑 TODO 和添加项目入口在 `not-started` 与 `in-progress` 状态下继续可用

## 3. 自动化测试

- [x] 3.1 更新 Go 单元测试，覆盖 `not-started -> in-progress` 成功和 `in-progress -> not-started` 被拒绝
- [x] 3.2 更新 Go 单元测试，覆盖 `not-started` 不能完成、`in-progress` 可以完成
- [x] 3.3 更新 App 层测试，覆盖 `not-started` TODO 不能创建终端，`in-progress` TODO 可以创建终端
- [x] 3.4 更新客户端自动化测试，覆盖未执行/执行中 TODO 动作按钮和添加终端按钮的显示规则
- [x] 3.5 更新客户端自动化测试，覆盖执行中 TODO 完成路径和未执行 TODO 不显示完成入口

## 4. 验证与 Review

- [x] 4.1 运行 Go 测试，确认后端状态约束和终端创建约束通过
- [x] 4.2 运行前端自动化测试，确认客户端交互和按钮可见性通过
- [x] 4.3 运行 OpenSpec 校验，确认 proposal、design、specs、tasks 可被解析
- [x] 4.4 执行自动 review，检查状态机一致性、API 兜底校验、UI 可访问性和测试覆盖

## 5. 打包

- [ ] 5.1 运行 Wails 打包任务，生成可执行文件
