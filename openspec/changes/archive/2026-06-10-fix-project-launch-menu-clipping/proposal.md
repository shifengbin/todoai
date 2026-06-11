## Why

左侧项目列表中的最后几个项目打开终端启动菜单时，菜单会被项目列表滚动容器裁剪，导致用户无法完整看到或选择下方选项。这个问题会直接影响从项目树快速创建终端的可用性，尤其是在窗口高度较小或启动配置较多时更明显。

## What Changes

- 调整项目侧栏终端启动菜单的展示行为，使菜单在靠近项目列表底部时仍能完整可见并可操作。
- 保持现有终端启动菜单内容和选择行为不变，包括内置 `Terminal` 选项和用户配置的启动 profile。
- 保持项目列表自身可滚动，避免为了弹出菜单牺牲长项目列表的浏览能力。
- 增加覆盖边界场景的验证，确保最后一个可用项目打开菜单时不会被遮挡。

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `project-workspace`: 补充项目终端启动菜单在侧栏滚动边界处必须完整可见并可操作的要求。

## Impact

- 前端组件：`frontend/src/components/ProjectSidebar.vue`
- 前端样式：`frontend/src/style.css`
- 前端测试：`frontend/src/components/ProjectSidebar.test.js`，必要时补充更适合布局边界的测试工具或测试策略
- 不涉及 Go 后端 API、持久化格式或终端启动协议变更
