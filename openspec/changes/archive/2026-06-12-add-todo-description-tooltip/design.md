## Context

当前应用是 Wails + Go 后端 + Vue 前端。TODO 数据模型已经包含 `description`，`ProjectSidebar.vue` 目前在 TODO 行内显示一行描述摘要，长描述会被截断。用户想在不打开 TODO 详情的情况下读到完整描述，因此该变更应只增强 TODO 列表的前端浏览体验。

## Goals / Non-Goals

**Goals:**

- 保留 TODO 行中的现有描述摘要，维持列表可扫视能力。
- 鼠标悬浮在有描述的 TODO 行上一段时间后显示完整描述 tooltip。
- 鼠标移开 TODO 行后立即隐藏 tooltip，并清理尚未触发的延迟计时器。
- tooltip 支持较长描述的多行阅读，并在浅色和深色主题下保持可读。
- 不影响 TODO 右键菜单、三点菜单、完成/删除确认气泡、项目展开收起和活动状态提示。

**Non-Goals:**

- 不修改 TODO 创建、编辑、持久化或后端数据结构。
- 不新增前端依赖或通用 tooltip 组件库。
- 不改变已完成 TODO 列表和 TODO 详情弹窗行为。
- 不在触摸设备上新增长按 tooltip 行为。

## Decisions

### 1. 使用组件状态和定时器控制 tooltip

`ProjectSidebar.vue` 使用本地响应式状态记录当前悬浮的 TODO id，并使用 `setTimeout` 延迟显示 tooltip。`mouseenter` 时若 TODO 有描述则启动计时器；`mouseleave` 时清理计时器并隐藏 tooltip；组件卸载时也清理计时器。

替代方案是使用浏览器原生 `title`。原生 title 的延迟、样式、换行和测试都不可控，并且会与现有活动状态 title 语义产生竞争。另一个替代方案是纯 CSS `transition-delay`，实现更少，但难以可靠测试“延迟前不显示”和“移开立即隐藏”。

### 2. tooltip 只展示完整描述，行内摘要继续保留

TODO 行继续显示当前的一行描述摘要；tooltip 作为完整内容预览层出现。这样用户能先通过列表发现哪些 TODO 有描述，再在需要时停留查看完整内容。

替代方案是移除行内描述，只保留 tooltip。该方案会让列表更紧凑，但降低描述存在性的可发现性，不符合“无需打开即可知道描述内容”的目标。

### 3. tooltip 挂在 TODO 行结构内部

tooltip 放在 TODO header/row 附近，跟随对应 TODO 行定位，避免引入全局 portal 或窗口坐标计算。样式使用绝对定位、合理宽度上限、换行和层级，确保不挤压列表布局。

如果后续发现滚动容器裁剪 tooltip，再考虑提升为全局浮层并计算定位；当前变更先保持局部实现，减少复杂度。

## Risks / Trade-offs

- [tooltip 被侧边栏滚动区域裁剪] -> 先将 tooltip 放在 TODO 行附近并设置合适层级；测试和人工检查发现裁剪时再升级为全局浮层定位。
- [tooltip 遮挡 TODO 操作按钮或上下文菜单] -> tooltip 只在悬浮延迟后显示，右键菜单和按钮点击流程保持优先；打开其他侧边栏浮层时可隐藏 tooltip。
- [定时器泄漏或切换 TODO 后显示错误内容] -> mouseleave、切换目标和组件卸载时统一清理 timer，并在触发时确认目标 TODO 仍是当前悬浮项。
- [测试依赖真实时间导致不稳定] -> 前端单测使用 fake timers 或可控计时推进，覆盖延迟前、延迟后和离开后的状态。

