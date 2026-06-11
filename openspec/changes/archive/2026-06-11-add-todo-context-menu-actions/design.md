## Context

TODO 工作区目前把活动 TODO 的状态推进、查看详情、添加项目和删除入口都放在 TODO 行内。`ProjectSidebar.vue` 负责 TODO 树、按钮和浮层状态，`App.vue` 负责调用 Wails API、管理 TODO 表单和剪贴板。创建 TODO、编辑 TODO 和添加项目弹窗都复用 `settings-overlay`，当前点击遮罩会直接关闭弹窗。

这次变更只调整 TODO 工作区的前端交互，不改变 TODO 数据模型、Go 后端方法签名或持久化格式。

## Goals / Non-Goals

**Goals:**

- 让 TODO 创建、详情编辑和添加项目弹窗只能通过关闭按钮、取消按钮或提交成功关闭。
- 在活动 TODO 行上提供右键菜单，集中承载查看详情、添加项目、复制描述和删除 TODO。
- 让 TODO 行外部只保留状态推进按钮：未执行任务显示开始，执行中任务显示完成。
- 复用现有 Wails runtime 剪贴板能力复制 TODO 描述。
- 更新客户端自动化测试覆盖新的入口和关闭规则。

**Non-Goals:**

- 不新增 TODO 状态。
- 不改变完成、删除、编辑或关联项目的后端语义。
- 不改变已完成 TODO 列表的行为。
- 不新增全局菜单系统或第三方菜单依赖。
- 不改变终端右键菜单的行为。

## Decisions

1. **TODO 右键菜单状态放在 `ProjectSidebar.vue`。**

   选择：Sidebar 维护当前打开的 TODO 菜单 `{ todoId, x, y }`，在 TODO 行 `contextmenu` 时打开，并在菜单点击、外部点击、切换浮层或 props 更新后关闭。

   原因：TODO 行、现有浮层和行内按钮都在 Sidebar 内，菜单布局和互斥关系放在同一组件里最直接。App 不需要知道菜单位置，只接收业务事件。

   备选：在 App 中实现全局 TODO 菜单。该方案能统一终端菜单和 TODO 菜单，但会让 App 了解 Sidebar DOM 事件细节，耦合更高。

2. **管理动作通过 Sidebar 事件回到 App。**

   选择：Sidebar 为菜单项分别触发 `edit-todo`、`add-project-to-todo`、`copy-todo-description` 和删除确认流程。App 处理 `copy-todo-description` 时查找 TODO，并调用 `ClipboardSetText(todo.description || '')`。

   原因：App 已经持有 TODO 数据和 Wails runtime clipboard 导入；Sidebar 保持为展示和事件组件，不直接依赖 Wails runtime。

   备选：把完整 TODO 描述作为事件 payload 从 Sidebar 传给 App。该方案可行，但会让 Sidebar 决定复制字段的业务语义；传 `todoId` 更贴近现有事件风格。

3. **删除 TODO 移入右键菜单。**

   选择：外部只保留状态按钮，所以删除与查看详情、添加项目一样作为管理动作进入右键菜单。删除仍需确认，取消删除不调用后端。

   原因：用户明确要求外部只保留对状态的修改按钮。删除不是状态推进动作，继续放在行外会破坏规则。

   备选：删除按钮仍保留在行外。该方案减少改动，但不符合“外部只保留状态修改按钮”的交互目标。

4. **TODO 弹窗遮罩不再承担关闭动作。**

   选择：去掉 TODO 创建、详情编辑和添加项目 overlay 的点击关闭处理，保留 dialog 内 `@click.stop` 不影响。关闭按钮、Cancel 按钮和成功提交后的关闭逻辑保持不变。

   原因：这是最小改动，能直接消除误触关闭。终端设置弹窗不属于本次 TODO 工作区需求，保持现状。

## Risks / Trade-offs

- [Risk] 右键菜单可能与已有终端启动菜单、确认气泡同时显示。→ Mitigation: 打开 TODO 菜单前关闭现有 Sidebar 浮层，打开其他 Sidebar 浮层时也关闭 TODO 菜单。
- [Risk] 既有测试大量直接点击 `edit-todo-*` 和 `add-project-to-todo-*`。→ Mitigation: 测试 helper 改为先触发 TODO 行右键菜单，再点击菜单项，并保留对事件 payload 的断言。
- [Risk] 空描述复制行为可能让用户误以为复制失败。→ Mitigation: 规格明确只复制 `description` 字段，空描述写入空字符串且不拼接标题。
- [Risk] 菜单用固定坐标定位时可能贴近窗口边缘。→ Mitigation: 初版沿用现有终端右键菜单的固定定位模式；如出现溢出，再单独优化菜单边界约束。

## Migration Plan

无需数据迁移。该变更仅影响前端交互和客户端测试；回滚时恢复行内管理按钮和 overlay 点击关闭即可。

## Open Questions

None.
