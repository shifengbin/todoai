## Why

当前 workspace 全局终端以独立标签组占用主终端区域顶部空间，并与左侧 TODO 工作流分离，用户难以在统一任务视图中观察其运行、待确认或等待输入状态。将它们投影为“执行中”分类里的虚拟 TODO，可以统一终端导航和活动反馈，同时继续保持全局终端不属于任何真实 TODO 的数据语义。

## What Changes

- 移除主终端区域顶部的 Global terminal 标签组，主区域仅保留实际 xterm pane。
- 在 TODO 侧栏的“执行中”分类首位展示固定的 `Global 终端` 虚拟 TODO，并在其下展示全部 workspace 全局终端。
- 虚拟 TODO 仅在至少存在一个全局终端时显示；无终端时继续通过顶部 `Global terminal` 按钮创建首个终端。
- 虚拟 TODO 支持与普通 TODO 一致的展开、收起、批量折叠和终端活动聚合，但不参与真实 TODO 的排序、拖拽、状态变更或持久化。
- 选择全局终端时高亮对应终端及虚拟父项，同时保留当前真实 TODO、项目和 Git 状态上下文。
- 从其他 TODO 分类创建全局终端成功后，自动切换到“执行中”、展开虚拟 TODO 并选中新终端；创建失败时保持原视图。
- 在虚拟 TODO 内保留新增和删除全局终端的入口，并复用现有终端生命周期与错误处理能力。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `workspace-global-terminals`: 将全局终端列表的展示位置从主终端区域顶部改为 TODO 侧栏中的虚拟 TODO，并定义创建、选择、折叠和隐藏行为。
- `todo-workspace`: 在“执行中”分类中引入固定首位且与真实 TODO 数据隔离的虚拟 TODO，定义其排序、活动聚合、批量折叠和视图切换规则。

## Impact

- 前端：调整 `frontend/src/App.vue` 与 `frontend/src/components/ProjectSidebar.vue` 的终端数据传递、布局、交互和活动聚合逻辑。
- 样式：移除主终端区 Global 标签组相关布局，并增加虚拟 TODO 与全局终端行样式。
- 测试：更新 `frontend/src/App.test.js`，扩展 `frontend/src/components/ProjectSidebar.test.js` 对虚拟 TODO 行为的覆盖。
- 后端与绑定：继续使用现有 `workspaceTerminal` 标记及创建、选择、删除 API，不改变 Go 数据模型或 Wails 接口。
