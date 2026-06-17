## Context

当前 workspace 相关业务数据已经保存到 `<workspace>/.data`，包括 `projects.json` 和 `terminal-history.json`。TODO 工作区的状态视图标签由 `ProjectSidebar.vue` 内部 `todoView` 管理，默认值为 `not-started`；左侧栏宽度由 `App.vue` 的 `sidebarWidth` 管理，默认值为 `280`。这两个状态都只在前端内存中，应用重启、刷新前端或重新打开 workspace 后会丢失。

这次状态需要以 TODO 工程为单位保存。当前数据模型中 TODO 工程是 workspace 内独立副本，`TodoProject.ID` 是该副本的稳定标识，适合作为 UI 状态的 key。由于用户明确要求这些状态保存在 `.data` 中，它们属于 workspace-scoped 数据，而不是应用全局 settings。

## Goals / Non-Goals

**Goals:**

- 按 `TodoProject.ID` 持久化 TODO 视图标签和左侧栏宽度。
- 将状态保存到当前 workspace 的 `.data` 目录。
- 重新打开软件或 workspace 后恢复当前 active TODO 工程的 UI 状态。
- 切换 active TODO 工程时恢复目标 TODO 工程的 UI 状态。
- 删除 TODO 工程时清理对应 UI 状态。
- 保持没有状态记录时的现有默认体验。

**Non-Goals:**

- 不把这些 UI 状态同步到全局候选项目，也不按工程路径跨 TODO 共享。
- 不持久化所有前端临时状态，例如弹窗开关、搜索词、右键菜单或批量选择。
- 不改变 TODO 的业务状态流转规则。
- 不改变应用全局 settings 的作用域。

## Decisions

### 1. 新增独立的 workspace UI 状态文件

使用 `<workspace>/.data/todo-project-ui-state.json` 保存 TODO 工程 UI 状态，结构为：

```json
{
  "version": 1,
  "todoProjects": {
    "todo-project-id": {
      "todoView": "not-started",
      "sidebarWidth": 360
    }
  }
}
```

备选方案是把字段加入 `projects.json` 的 `TodoProject`。该方案会把 UI 偏好和业务数据混在一起，也会让未来清理、重置或迁移 UI 状态变复杂。独立文件与现有 `terminal-history.json` 的思路一致：workspace-scoped，但和核心项目配置解耦。

### 2. 以后端作为持久化边界

新增 Go 侧 store/manager 负责读写 `todo-project-ui-state.json`，并通过 Wails 暴露加载、保存和删除接口。前端不直接写文件，也不使用 `localStorage`。

备选方案是前端用浏览器存储。该方案不满足“存放在 `.data` 文件夹中”，也会让 workspace 切换、备份和迁移不透明。

### 3. 按 `TodoProject.ID` 隔离状态

UI 状态 key 使用 `TodoProject.ID`。同一个真实工程路径出现在不同 TODO 下时，它们是两个 TODO 工程副本，状态互相独立。

备选方案是按路径保存。路径更接近真实项目，但用户这次确认的是“以一个 TODO 工程为单位”，因此按路径会错误共享不同 TODO 下的状态。

### 4. 前端在状态变化点保存

`ProjectSidebar.vue` 改为接收当前 `todoView`，并在用户点击 `未执行`、`执行中`、`已完成` 时向父组件发出变更事件。`App.vue` 根据当前 `activeTodoProjectId` 保存新值。

左侧栏宽度继续在拖动过程中更新内存状态；只在拖拽结束时保存最终宽度，避免鼠标移动时高频写磁盘。

### 5. 恢复顺序以 active TODO 工程为准

workspace 加载完成后，前端拿到 `activeTodoProjectId`，再加载或应用对应 UI 状态。若当前没有 active TODO 工程，保留默认 `todoView = not-started` 和 `sidebarWidth = 280`。当 active TODO 工程变化时，优先应用目标工程的记录；若记录缺失，回退默认值。

## Risks / Trade-offs

- [Risk] 前端状态和后端保存之间存在短暂不同步。→ Mitigation: 标签切换立即保存，宽度在拖拽结束保存；保存失败时显示现有错误信息，但不阻塞用户继续使用当前内存状态。
- [Risk] 删除 TODO 工程后 UI 状态文件残留无效 key。→ Mitigation: 删除 TODO 工程和删除 TODO 时调用清理逻辑，测试覆盖对应场景。
- [Risk] 无效 JSON 或旧版本文件可能影响 workspace 打开。→ Mitigation: 缺失或解析失败时返回默认空状态，不阻止 workspace 打开。
- [Risk] 过小或过大的 `sidebarWidth` 造成布局异常。→ Mitigation: 读取和保存时都使用前端现有最小/最大宽度约束，后端也可做保守范围校验。

## Migration Plan

1. 新增 `todo-project-ui-state.json` 读写结构，缺失文件视为空状态。
2. `App` 在打开、切换和关闭 workspace 时同步切换 UI 状态 store。
3. 前端加载 workspace 状态后恢复 active TODO 工程 UI 状态。
4. 标签切换和分割线拖拽结束时保存当前 TODO 工程状态。
5. 删除 TODO 工程时清理对应状态。

回滚时可以保留 `todo-project-ui-state.json`，旧版本会忽略该文件。删除该文件只会丢失 UI 偏好，不影响 TODO、工程、终端历史或 settings。

## Open Questions

无。当前约定为：按 `TodoProject.ID` 保存，文件位于当前 workspace 的 `.data` 目录，记录 TODO 视图标签和左侧栏宽度。
