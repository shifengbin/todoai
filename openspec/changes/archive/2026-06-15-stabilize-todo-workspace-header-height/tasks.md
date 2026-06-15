## 1. 前端布局实现

- [x] 1.1 调整 `ProjectSidebar.vue` 的 TODO toolbar 渲染，让 toolbar 容器在 `未执行`、`执行中`、`已完成` 三个状态视图中都保留。
- [x] 1.2 确保排序、批量收起和批量展开控件只在 `未执行` 与 `执行中` 视图中渲染或可交互。
- [x] 1.3 调整 `frontend/src/style.css` 中 TODO toolbar 样式，使空 toolbar 占位与开放视图 toolbar 保持同等高度。
- [x] 1.4 确认 `.todo-view-tabs` 三列宽度分配不随状态视图切换改变。

## 2. 客户端自动化测试

- [x] 2.1 在 `ProjectSidebar.test.js` 增加测试，验证切换到 `已完成` 后仍保留 TODO toolbar 布局占位。
- [x] 2.2 增加测试，验证 `已完成` 视图不暴露可操作的排序、批量收起和批量展开控件。
- [x] 2.3 增加或更新样式测试，验证状态切换栏仍使用三等分宽度布局，且 toolbar 样式包含固定/最小高度约束。
- [x] 2.4 运行前端测试命令，确认相关组件测试通过。

## 3. 质量检查

- [x] 3.1 运行 OpenSpec 校验，确认 `stabilize-todo-workspace-header-height` 的 proposal、design、specs 和 tasks 可被识别。
- [x] 3.2 进行自动代码 review，检查布局实现是否符合 spec、是否引入不可访问的隐藏控件或改变顶部状态按钮宽度。

## 4. 打包验证

- [x] 4.1 运行 `wails build -tags webkit2_41`，确认生成可执行文件。
