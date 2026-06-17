## Context

TODO 工作区当前使用 `TodoProjectUIState` 同时保存 TODO 视图标签和左侧 TODO 栏宽度。前端在 `applyState()` 中只要发现 `activeTodoProjectId` 变化，就调用 UI 状态恢复逻辑；而添加终端和切换终端会通过后端同步改变 active TODO 工程，因此终端操作会意外恢复旧的 TODO 视图和左侧栏宽度。

这两个字段的归属不同：TODO 视图标签表达“当前 TODO 工程上次查看哪个状态视图”，可以按 TODO 工程保存；左侧 TODO 栏宽度表达 workspace 的布局偏好，应随 workspace 保存，不应跟随 TODO item、TODO 工程或终端变化。

## Goals / Non-Goals

**Goals:**

- 保持添加终端和切换终端时当前 TODO 视图不变。
- 保持添加终端和切换终端时左侧 TODO 栏宽度不变。
- 将左侧 TODO 栏宽度定义并实现为 workspace 级 UI 状态。
- 保留 TODO 视图标签按 TODO 工程保存和恢复。
- 兼容读取已有 `todo-project-ui-state.json` 中的 `sidebarWidth`，减少升级后的布局丢失。

**Non-Goals:**

- 不改变终端创建、终端选择和终端启动 profile 的后端业务语义。
- 不改变 TODO 工作流状态规则。
- 不引入全新的设置页面或用户可见配置项。

## Decisions

### 拆分 UI 状态归属

选择：TODO 视图标签继续保存在 `todoProjects[todoProjectId].todoView`，左侧 TODO 栏宽度保存为 workspace 级字段，例如 `sidebarWidth`。

原因：视图标签和 TODO 工程上下文相关，宽度是整个 TODO 工作区的布局偏好。把宽度放在 TODO 工程下会导致切换 TODO 工程、切换终端或后端 active context 刷新时出现布局跳变。

备选方案：继续按 TODO 工程保存宽度，只修复终端路径不触发恢复。这个方案改动更小，但仍会让用户主动切 TODO 工程时左侧栏宽度跳变，不符合“宽度属于整个工程而不是某个 TODO item”的要求。

### 显式恢复，而不是根据 active id 自动恢复

选择：前端 `applyState()` 不再因为 `activeTodoProjectId` 普通变化自动恢复 TODO 工程 UI 状态。只有打开 workspace、前端重新加载、重新打开 workspace，或用户主动点击 TODO 工程行时，才显式恢复相关 UI 状态。

原因：后端 active TODO 工程会被多种业务动作更新，包括添加终端和切换终端。仅凭 active id 变化无法判断用户意图。

备选方案：后端为每个返回状态标记触发来源。该方案会扩大 Wails API 和后端状态模型的改动范围，当前前端已经知道触发路径，没必要增加协议复杂度。

### 兼容旧数据

选择：保留旧 JSON 结构的读取能力。首次加载 workspace 级宽度时，如果新字段不存在，可以从当前 active TODO 工程旧 `sidebarWidth` 或任意有效旧宽度迁移到 workspace 级字段；后续保存写入新字段。

原因：旧版本已经把宽度写入 TODO 工程 UI 状态。完全丢弃旧字段会让用户升级后宽度重置。

备选方案：直接重置为默认宽度。实现最简单，但用户可感知，并且迁移成本不高。

## Risks / Trade-offs

- [Risk] 新旧字段同时存在时可能出现优先级歧义。→ Mitigation: 新 workspace 级字段优先；只有新字段缺失时才读取旧 TODO 工程宽度。
- [Risk] 保存 TODO 视图时误覆盖 workspace 宽度，或保存宽度时误覆盖 TODO 视图。→ Mitigation: 前端拆分持久化函数，分别保存 TODO 工程视图状态和 workspace 布局状态。
- [Risk] Wails 生成文件需要随 Go 模型/API 更新。→ Mitigation: 实现后运行项目现有 Wails 生成命令或等效测试路径，确保 `frontend/wailsjs` 与 Go 绑定一致。
- [Risk] 现有测试断言“按 TODO 工程恢复宽度”需要调整。→ Mitigation: 用新需求替换相关测试，覆盖 workspace 级宽度恢复和终端操作不改变宽度。
