## Context

当前应用是 Wails + Go 后端 + Vue 前端。TODO 工作树由 `ProjectSidebar.vue` 渲染，`App.vue` 负责调用后端完成、删除和状态同步。TODO 下项目移除已经在侧边栏内使用按钮旁确认气泡，并通过 `closeFloatingMenus()` 和终端启动菜单互斥；但 TODO 完成和删除仍在 `App.vue` 中使用 `window.confirm`。项目行选中状态目前主要由 `.project-row.active` 表达，只覆盖项目信息按钮，右侧创建终端和删除关联按钮区域没有共享同一背景。

## Goals / Non-Goals

**Goals:**

- 将 TODO 完成和删除确认迁移到 `ProjectSidebar.vue` 的行内确认气泡。
- 让 TODO 操作气泡与终端启动菜单、项目移除气泡共用同一类关闭行为。
- 从 `App.vue` 的 TODO 完成/删除路径移除 `window.confirm`。
- 让 TODO 下项目行的 hover/active 背景覆盖整条项目 header，包括创建终端和删除按钮区域。
- 保持现有后端完成、删除、归档和终端清理语义不变。

**Non-Goals:**

- 不改变 TODO 完成或删除的后端 API。
- 不改变项目库视图的项目删除确认方式。
- 不引入新的 UI 组件库或外部依赖。
- 不把所有确认交互抽象成全局通用组件。

## Decisions

### 1. TODO 完成/删除确认在 `ProjectSidebar.vue` 内完成

侧边栏新增一个 TODO 操作确认状态，例如 `{ todoId, action }`，其中 `action` 为 `complete` 或 `delete`。点击完成或删除按钮只打开对应确认气泡；用户在气泡中确认后，侧边栏才 emit `complete-todo` 或 `delete-todo`。

这样可以让确认气泡靠近触发按钮，且 `App.vue` 只负责执行已经确认的业务动作。替代方案是在 `App.vue` 里维护确认状态，但确认气泡属于侧边栏行布局，放在 `ProjectSidebar.vue` 更接近现有项目移除气泡模式。

### 2. 浮层互斥沿用现有关闭函数

`closeFloatingMenus()` 应关闭终端启动菜单、TODO 项目移除气泡和新的 TODO 操作确认气泡。打开任一浮层前先关闭其它浮层；点击外部、取消、确认成功或目标切换时关闭当前气泡。

替代方案是为每类浮层写独立 window click handler。这样容易出现多个浮层同时打开，也会重复事件处理逻辑。

### 3. 使用专门的 TODO 操作气泡样式，但复用项目移除气泡视觉语言

TODO 完成和删除的气泡文案不同：完成是普通破坏性较低动作，删除是更强破坏性动作。可以共用结构和基础样式，但删除确认按钮使用 delete 语义颜色，完成确认按钮使用 accent 语义颜色。按钮仍需可通过 `data-testid` 区分。

不在本次变更中创建通用 `ConfirmPopover` 组件。当前只有同一文件内的少量重复，保留局部实现更直接；若后续确认气泡继续增加，再抽象更合理。

### 4. TODO 项目行背景挂到整条 header

TODO 项目行的 hover/active 视觉应由 `.todo-project-header-row` 或其状态类控制，而不是只由内部 `.project-row` 控制。内部 `.project-row` 保持透明背景，负责文本布局和点击区域；右侧按钮在整行背景上保持透明默认态和明确 hover/focus 态。

项目库视图仍保留原有 `.project-row.active` 视觉，不受 TODO 项目行背景迁移影响。

## Risks / Trade-offs

- [气泡点击后事件冒泡导致立即关闭] → 气泡根节点继续使用 `@click.stop`，触发按钮使用 `@click.stop`。
- [多个浮层同时打开] → 所有打开函数先调用统一关闭函数或关闭不相关浮层。
- [移除 `window.confirm` 后测试仍依赖原生确认 mock] → 更新 App 测试，改为通过侧边栏气泡确认后验证后端调用。
- [整行背景影响按钮可读性] → 按钮默认透明，hover/focus 使用现有 accent/delete 色，并在亮暗主题下保持对比。
- [CSS 影响项目库行] → 仅针对 `.todo-project-header-row` 添加整行背景规则，避免改变 `.library-project-header-row`。

## Migration Plan

1. 在 `ProjectSidebar.vue` 增加 TODO 操作确认状态、打开/关闭/确认函数和模板。
2. 从 `App.vue` 的 `completeTodo` 和 `deleteTodo` 中移除 `window.confirm` 分支。
3. 调整 `style.css` 中 TODO 项目行 hover/active 背景覆盖范围和确认气泡样式。
4. 更新前端测试覆盖气泡确认、取消、外部关闭、互斥关闭和项目行整行背景规则。
5. 运行前端单元测试和前端构建。

回滚时只需要恢复前端模板、样式和测试；后端状态结构和持久化数据不变。

## Open Questions

无。确认方式、背景覆盖范围和不改后端语义均已确定。
