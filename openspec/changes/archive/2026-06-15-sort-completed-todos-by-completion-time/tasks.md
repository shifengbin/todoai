## 1. 前端排序实现

- [x] 1.1 在 `ProjectSidebar.vue` 中为已完成 TODO 添加完成时间 timestamp helper，优先解析 `completedAt`，缺失时解析 `archivedAt`。
- [x] 1.2 将 `completedTodos` computed 改为先过滤 `completed` 状态，再按完成时间倒序排序。
- [x] 1.3 为无有效完成时间和完成时间相同的 TODO 添加稳定兜底排序，确保无有效时间项排在末尾且同时间项保持原始相对顺序。
- [x] 1.4 确认 `未执行` 和 `执行中` 视图继续使用现有优先级/创建时间排序逻辑，且排序控件不影响 `已完成` 视图。

## 2. 客户端自动化测试

- [x] 2.1 在 `ProjectSidebar.test.js` 增加 `已完成` 视图按 `completedAt` 倒序展示的测试。
- [x] 2.2 增加缺失 `completedAt` 时按 `archivedAt` 兜底排序的测试。
- [x] 2.3 增加无有效完成时间的已完成 TODO 排在有时间 TODO 之后的测试。
- [x] 2.4 运行前端测试命令，确认相关组件测试通过。

## 3. 质量检查

- [x] 3.1 运行 OpenSpec 校验，确认 `sort-completed-todos-by-completion-time` 的 proposal、design、specs 和 tasks 可被识别。
- [x] 3.2 进行自动代码 review，检查排序实现是否符合 spec、是否存在数据变异、边界时间解析或测试遗漏。

## 4. 打包验证

- [x] 4.1 运行 `wails build -tags webkit2_41`，确认生成可执行文件。
