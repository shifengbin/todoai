## 1. 状态聚合与展示

- [x] 1.1 在 `ProjectSidebar.vue` 中为 TODO 派生子终端活动状态，按 `needs-input > busy > idle` 计算最高优先级状态。
- [x] 1.2 在 TODO 收起态渲染聚合活动提示，复用现有终端活动图标语义，并补充 `aria-label`、`title` 或 `data-activity-state` 以支持可访问性和测试。
- [x] 1.3 确保 TODO 展开态不重复显示父级聚合提示，继续由子终端行展示详细活动状态。

## 2. 客户端自动化测试

- [x] 2.1 为 `ProjectSidebar.test.js` 增加收起 TODO 显示 `needs-input` 聚合状态的测试。
- [x] 2.2 为 `ProjectSidebar.test.js` 增加 `needs-input` 优先于 `busy` 的聚合优先级测试。
- [x] 2.3 为 `ProjectSidebar.test.js` 增加展开 TODO 不显示父级聚合提示、终端行仍显示活动状态的测试。

## 3. 验证与自动 Review

- [x] 3.1 运行前端组件测试，确认新增和既有 `ProjectSidebar` 测试通过。
- [x] 3.2 运行项目相关自动化测试或检查命令，确认没有破坏现有前端行为。
- [x] 3.3 执行自动 review，检查实现是否符合设计、spec 和现有代码风格。
