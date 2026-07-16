## 1. 布局合同测试

- [x] 1.1 更新 `ProjectSidebar.test.js` 的侧栏布局测试，先断言共享工具栏位于 `.todo-workspace-scroll` 之外、状态页签之后且滚动容器只承载当前 TODO 列表，并确认测试在实现前失败。
- [x] 1.2 补充开放视图与已完成视图回归断言，验证固定工具栏分别显示排序及全部展开/收起控件、已选数量及批量删除控件，并保持现有交互和无障碍属性。
- [x] 1.3 补充样式合同断言，覆盖工具栏不可收缩、与列表右侧留白一致，以及滚动容器继续使用剩余高度并保持可滚动。

## 2. 固定共享工具栏

- [x] 2.1 调整 `ProjectSidebar.vue` 模板层级，将完整的 `.todo-tree-toolbar` 移到 `.todo-workspace-scroll` 之前，同时保留现有视图分支、事件处理器和列表内容。
- [x] 2.2 调整 `style.css` 的 flex 尺寸与间距，使状态页签和共享工具栏固定、TODO 列表独立滚动，并保持工具栏与列表边缘对齐。
- [x] 2.3 确认 `todoWorkspaceScrollElement`、弹窗滚动锁定和 SortableJS `scroll` 配置仍指向仅包含 TODO 列表的滚动容器。

## 3. 自动化验证与审查

- [x] 3.1 运行 `cd frontend && npm run test -- src/components/ProjectSidebar.test.js`，确认固定工具栏、三视图控件和拖拽滚动目标相关组件测试通过。
- [x] 3.2 运行 `cd frontend && npm run test` 和 `cd frontend && npm run build`，确认完整客户端自动化测试与生产构建通过。
- [x] 3.3 运行自动代码审查，检查变更是否符合 proposal、design 和 delta spec，重点审查滚动边界、三视图一致性、拖拽自动滚动及测试覆盖，并修复发现的问题。
- [x] 3.4 运行 `git diff --check`，确认补丁不存在空白错误。

## 4. 应用打包

- [x] 4.1 运行 `wails build -tags webkit2_41`，生成可执行文件并确认打包成功。
