## Why

TODO 分支收起后，用户无法看到其子终端是否正在运行或等待输入，容易遗漏需要处理的终端。需要在保持列表紧凑的同时，将被隐藏终端的重要活动状态反馈到父级 TODO item。

## What Changes

- 在 TODO 分支收起时，在父级 TODO item 上显示其子终端的聚合活动状态。
- 聚合状态按 `needs-input > busy > idle` 的优先级计算，确保等待输入的终端优先获得用户注意。
- 展开 TODO 分支时继续由现有终端行展示活动状态，父级 TODO item 不重复显示聚合提示。
- 为聚合状态补充可访问标签和可自动化验证的状态属性。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 扩展 TODO 分支收起行为，使父级 TODO item 能反映被隐藏子终端的聚合活动状态。

## Impact

- 主要影响 `frontend/src/components/ProjectSidebar.vue` 中 TODO 树的状态派生与展示。
- 需要扩展 `frontend/src/components/ProjectSidebar.test.js` 的组件自动化测试。
- 不修改 Go 后端、Wails API、持久化模型或终端会话生命周期。
