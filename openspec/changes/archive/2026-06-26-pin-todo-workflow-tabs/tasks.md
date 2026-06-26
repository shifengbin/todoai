## 1. Layout Implementation

- [x] 1.1 调整 `ProjectSidebar.vue` 模板结构，将 `.todo-view-tabs` 保持在 TODO 列表滚动区域之外。
- [x] 1.2 为 tab 下方内容添加或调整独立滚动容器，确保工具栏、空状态、TODO 树和已完成列表仍按当前视图正常渲染。
- [x] 1.3 更新 `frontend/src/style.css`，让侧边栏继续使用 flex column 布局，并让新的滚动容器使用 `flex: 1 1 auto; min-height: 0; overflow-y: auto;`。
- [x] 1.4 确认 `未执行`、`执行中`、`已完成` 三个 tab 的点击、激活态和可访问性属性不受 DOM 层级调整影响。

## 2. Frontend Tests

- [x] 2.1 更新 `ProjectSidebar` 布局测试，断言 `.todo-view-tabs` 不在滚动内容容器内，且三个 tab 仍按 `未执行`、`执行中`、`已完成` 顺序渲染。
- [x] 2.2 新增或更新 CSS 规则测试，覆盖 tab 区固定在滚动区域之外、下方内容容器负责滚动的关键布局规则。
- [x] 2.3 覆盖 `已完成` 视图，确认批量选择/删除工具栏仍在滚动内容区域内正常展示。
- [x] 2.4 运行 `cd frontend && npm test`，确保客户端自动化测试通过。

## 3. Review and Build

- [x] 3.1 进行自动 review，检查布局变更是否符合设计、规范和现有前端风格。
- [x] 3.2 Run `wails build -tags webkit2_41` to generate the executable file.
