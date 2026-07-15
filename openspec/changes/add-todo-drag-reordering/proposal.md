## Why

当前开放 TODO 只能按优先级或创建时间自动排序，用户无法表达与这些字段无关的实际执行顺序。增加可持久化的手动拖拽排序，可以让任务列表直接反映用户的工作计划，并在切换工作区或重启应用后保持一致。

## What Changes

- 在 `未执行` 和 `执行中` 视图的排序控件中新增“手动”模式，并仅在该模式显示 TODO 拖拽手柄。
- 为两个开放状态分别保存有序 TODO ID 列表；首次进入手动模式时沿用切换前的显示顺序，新建 TODO 或切换状态的 TODO 追加到目标列表末尾。
- 按工作区持久化最后选择的排序模式和手动顺序，并兼容没有这些字段的旧 UI 状态文件。
- 拖动开始后临时视觉收起当前列表中的所有 TODO 子树，拖动结束或取消后恢复原有展开状态；长列表拖动时支持自动滚动。
- 拖拽只允许发生在当前状态列表内部；`已完成` 视图继续按完成时间排序且不提供拖拽。
- 保存失败时回滚到拖动前顺序并显示错误，不丢失原有展开状态。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `todo-workspace`: 扩展开放 TODO 排序要求，增加手动模式、拖拽交互、工作区级顺序持久化及状态变化时的顺序维护规则。

## Impact

- 前端：`frontend/src/components/ProjectSidebar.vue`、`frontend/src/App.vue`、相关样式和 Vitest 组件测试。
- 后端：工作区级 `TodoProjectUIStateFile`、Wails App 方法及对应 Go 单元测试。
- 绑定：Wails 生成的 JavaScript/TypeScript API 与模型可能需要更新。
- 依赖：前端将引入或复用一个兼容 Vue 3、支持拖拽排序和自动滚动的成熟库。
- 数据：`todo-project-ui-state.json` 增加可选字段，旧文件缺少字段时继续使用优先级排序，不构成破坏性变更。
