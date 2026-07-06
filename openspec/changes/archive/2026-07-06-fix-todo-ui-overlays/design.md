## Context

Todo 相关 UI 由 `tui-helper/frontend/src/App.vue` 和 `ProjectSidebar.vue` 共同承载，样式集中在 `tui-helper/frontend/src/style.css`。分支选择器状态由 `openProjectBranchPickerKey` 和 `projectBranchPickerQueries` 管理，三处入口共享 `openProjectBranchPicker`、`updateProjectBranchInput`、`selectProjectBranchCandidate`、`closeProjectBranchPicker`。任务列表滚动区域使用 `.todo-workspace-scroll`，Todo 行右侧操作按钮由 `.todo-actions` 放在 grid 最右侧。

当前行为有两个问题：
- 手动输入自定义分支时，只会更新输入值并保持候选框打开；若用户点击弹窗内其他位置，候选框没有通用关闭路径。
- 任务列表出现滚动条时，滚动条占用最右侧空间，容易压到 `.todo-actions` 中的按钮。

## Goals / Non-Goals

**Goals:**
- 让分支候选框在输入框失焦时关闭，并保留用户手动输入的分支值。
- 让创建 Todo、编辑 Todo、给 Todo 添加项目三处共享一致的分支候选框关闭行为。
- 让任务列表滚动容器为滚动条预留稳定空间，避免遮盖右侧操作按钮。
- 用现有前端测试方式覆盖交互和样式约束。

**Non-Goals:**
- 不改变分支候选数据来源、过滤规则或候选数量限制。
- 不改变 Todo、TodoProject 或后端工作区数据结构。
- 不调整任务列表排序、折叠、状态切换或终端创建逻辑。
- 不引入新的 UI 组件库或外部依赖。

## Decisions

1. 分支候选框关闭采用输入框失焦触发。

   选择在三处分支输入框上调用既有 `closeProjectBranchPicker`，而不是新增全局 click 监听。原因是当前问题发生在输入完成后离开输入区域，失焦语义正好匹配；同时候选项已有 `@mousedown.prevent`，点击候选时不会先触发失焦导致选项无法选择。

   备选方案是增加 `window.click` 监听并判断点击目标是否在 picker 内。该方案更通用，但需要额外 DOM target 判断，并且 `App.vue` 中多个弹窗已有 `@click.stop`，局部失焦处理更小、更符合当前问题范围。

2. 任务列表滚动条避让复用现有样式策略。

   `.todo-project-options` 已使用 `padding-right: 10px` 和 `scrollbar-gutter: stable` 避免候选操作按钮贴近滚动条。任务列表滚动容器采用同样思路，保持布局稳定，并减少跨浏览器滚动条宽度差异对右侧按钮的影响。

   当 settings 类弹窗打开时，`app-shell` 标记为 modal overlay 状态，并临时隐藏 `.todo-workspace-scroll` 的纵向滚动条。这样可以避免原生滚动条或 gutter 在半透明遮罩和 Todo 弹窗区域中继续可见。

3. 测试以行为和样式契约为主。

   `App.test.js` 覆盖手动输入自定义分支后失焦关闭下拉框且提交值保留，并覆盖 Todo 弹窗打开时隐藏侧边栏任务滚动条。`ProjectSidebar.test.js` 扩展现有布局样式断言，确认任务滚动区域包含滚动条 gutter 和右侧留白。

## Risks / Trade-offs

- [Risk] 输入框失焦会关闭候选框，用户若只是想临时查看候选再继续输入，需要重新聚焦输入框。→ Mitigation: 聚焦输入框仍会重新打开候选框，且现有 stale filter 重置逻辑保持可用。
- [Risk] `scrollbar-gutter` 在部分旧运行环境可能支持不完整。→ Mitigation: 同时保留 `padding-right` 作为实际避让空间，`scrollbar-gutter` 只用于增强稳定性。
- [Risk] 三处分支输入共用逻辑，遗漏任一入口会导致行为不一致。→ Mitigation: 实现和测试应同时覆盖创建 Todo、编辑 Todo、给 Todo 添加项目的分支输入入口，至少保证共享代码和关键创建路径有测试。
