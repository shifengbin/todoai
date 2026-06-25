## Why

当前候选项目列表只能一次性清空全部候选。用户如果只想移除一个误导入、路径失效或不再使用的候选项目，需要重新导入其余项目，操作成本过高。

## What Changes

- 在创建 TODO、编辑 TODO、向 TODO 添加工程的候选项目列表中支持清除单个候选项目。
- 清除单个候选项目前要求用户确认，避免误删候选记录。
- 清除后该候选项目从全局候选库和当前候选列表中移除。
- 如果该候选项目已在当前弹窗中被临时选中，清除时同步移除该临时选择，避免提交不存在的项目 ID。
- 清除单个候选项目只删除候选记录，不删除磁盘目录，不删除已加入 TODO 的工程副本，不关闭已有 TODO 工程终端。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `global-project-candidates`: 扩展全局项目候选管理，允许用户清除单个候选项目并保持既有安全边界。

## Impact

- 前端：`frontend/src/App.vue` 候选项目列表、交互处理和样式。
- 测试：`frontend/src/App.test.js` 覆盖单项清除确认、取消、当前选择同步移除和 TODO 工程副本保留。
- 后端：复用已有 `DeleteProject(projectID)` 行为，预计不新增 Wails API。
- 规格：更新 `global-project-candidates` 需求和场景。
