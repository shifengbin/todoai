## Why

当前任务级终端被选中后，左侧 TODO 行仍然像未选中状态，顶部标题和路径也可能继续显示上一个 TODO project 上下文，导致用户无法确认当前终端实际属于哪个任务目录。与此同时，长任务列表中靠后的终端启动菜单会被侧栏滚动容器裁剪，影响创建任务级或项目级终端。

## What Changes

- 选择任务级终端时，将当前上下文切换为任务级上下文：当前 TODO 指向该终端所属 TODO，当前 TODO project 和全局项目上下文清空。
- 顶部工作区标题和路径在任务级终端激活时显示任务级信息，包括 TODO 标题、任务级终端标识和该 TODO 的任务目录路径。
- 左侧 TODO 工作树在任务级终端激活时对其父 TODO 展示选中背景，使任务级终端和父任务的关系可见。
- 终端启动菜单脱离 TODO 列表滚动裁剪，靠近侧栏可视边界时仍可完整显示或以受限高度滚动。
- 不改变终端 session 的归属规则：任务级终端仍属于 TODO，不属于任何 TODO project。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `embedded-shell-sessions`: 修改任务级终端选择语义，使任务级终端成为当前任务上下文，并清空当前 TODO project/project 上下文。
- `todo-workspace`: 修改 TODO 工作树和工作区标题展示要求，使任务级终端激活时父 TODO 高亮并显示任务目录标题/路径；同时要求终端启动菜单不被长列表滚动容器裁剪。

## Impact

- 后端：`SelectTerminal` 的任务级终端分支、任务级上下文状态返回、相关 Go 单元测试。
- 前端：`App.vue` 的顶部标题/路径派生、活动上下文应用逻辑、`ProjectSidebar.vue` 的 TODO 激活样式判断和终端启动菜单定位/渲染。
- 测试：补充任务级终端选择、任务级标题路径、长列表下启动菜单可见性的前端测试；调整已有“任务级终端不改变 TODO project context”的后端测试。
