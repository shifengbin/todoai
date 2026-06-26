## Context

TODO 工作区侧边栏当前由 `.project-sidebar` 包含固定的 `.sidebar-header` 和一个可滚动的 `.project-list`。工作流 tab、工具栏和 TODO 树都在 `.project-list` 内，因此列表内容超过高度后，`未执行`、`执行中`、`已完成` 三个 tab 会随内容滚动离开可见区域。

现有规范已经要求长列表滚动时 tab 控件保持可见。本设计将实现方式定为结构拆分：tab 保持在滚动区域之外，只有 tab 下方内容滚动。

## Goals / Non-Goals

**Goals:**

- 让 `未执行`、`执行中`、`已完成` 三个工作流 tab 在 TODO 列表滚动时始终可见且可点击。
- 将 TODO 列表滚动限制在 tab 下方的独立区域内。
- 保持现有 tab 切换、排序、折叠/展开、已完成批量操作等行为不变。
- 保持侧边栏 header 和终端区域尺寸行为不变。

**Non-Goals:**

- 不改变 TODO 工作流状态、排序规则或持久化逻辑。
- 不新增后端 API、数据结构或 Wails 绑定。
- 不要求 TODO 工具栏固定在顶部；本次只固定三个工作流 tab。

## Decisions

1. **使用 DOM 结构拆分而不是 `position: sticky`。**

   将 `.todo-view-tabs` 放在可滚动列表区域之外，新增或调整包裹层让 `.todo-tree-toolbar` 与 TODO 树继续在下方滚动。这样 tabs 的固定效果不依赖 sticky 定位、z-index 或滚动容器 padding，行为更稳定。

   备选方案是给 `.todo-view-tabs` 使用 `position: sticky; top: 0`。该方案改动较小，但 sticky 与容器 padding、后续 toolbar、背景覆盖之间更容易出现遮挡或缝隙。

2. **保持现有组件事件和数据流不变。**

   tab 按钮继续调用 `setTodoView`，工具栏继续使用现有 computed 状态和事件。布局变化只影响模板层级和 CSS，不改变业务逻辑。

3. **用 CSS 测试覆盖布局合同。**

   现有前端测试已经通过读取 `src/style.css` 验证部分布局规则。本变更继续沿用该方式，检查 tab 区不可滚动、下方列表区域可滚动，并补充 DOM 顺序断言，避免回归为 tabs 放回滚动区域。

## Risks / Trade-offs

- **风险：滚动容器切换后空状态或已完成视图高度计算异常。** → 保持侧边栏为 flex column 布局，并让新的滚动内容容器使用 `flex: 1 1 auto; min-height: 0; overflow-y: auto;`。
- **风险：工具栏从固定位置变为随列表滚动时用户预期不一致。** → 当前用户明确选择只把三个状态按钮固定在上面；工具栏保持在列表内容区域内，避免扩大交互范围。
- **风险：测试环境无法真实计算滚动位置。** → 单元测试验证 DOM 层级和关键 CSS 规则，必要的视觉滚动行为由浏览器布局自然保证。
