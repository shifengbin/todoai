## 1. 客户端测试

- [x] 1.1 在 `ProjectSidebar` 组件测试中新增活动 TODO 按 `高`、`中`、`低` 优先级展示的用例。
- [x] 1.2 新增同优先级活动 TODO 按 `createdAt` 创建时间正序展示的用例。
- [x] 1.3 新增归档 TODO 列表不应用活动 TODO 排序规则的用例。

## 2. 前端实现

- [x] 2.1 在 `ProjectSidebar.vue` 中为 TODO 优先级定义稳定排序权重，并将缺失或未知优先级按 `medium` 处理。
- [x] 2.2 更新 `activeTodos` 派生列表，使其先过滤活动 TODO，再按优先级权重和 `createdAt` 正序排序。
- [x] 2.3 保持 `archivedTodos` 派生列表现有过滤和展示顺序不变。

## 3. 验证与 review

- [x] 3.1 运行前端自动化测试，确认 `ProjectSidebar` 排序行为和既有组件行为通过。
- [x] 3.2 运行相关项目验证命令，确认变更没有破坏现有 Go/Vue 测试。
- [x] 3.3 执行自动 review，检查排序实现是否符合规格、边界处理是否清晰、测试覆盖是否充分。

## 4. 排序切换控件补充

- [x] 4.1 在 `ProjectSidebar` 组件测试中新增排序切换控件默认选中优先级的用例。
- [x] 4.2 在 `ProjectSidebar` 组件测试中新增用户切换到时间排序后按 `createdAt` 正序展示的用例。
- [x] 4.3 在 `ProjectSidebar.vue` 中新增活动 TODO 排序模式状态和切换控件。
- [x] 4.4 更新 `activeTodos` 派生列表，使优先级模式和时间模式分别使用对应排序规则。
- [x] 4.5 运行前端测试、Go 测试和空白检查，并复查排序控件实现。
