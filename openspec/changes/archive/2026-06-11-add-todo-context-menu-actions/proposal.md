## Why

当前 TODO 行内同时展示状态推进、查看详情、添加项目和删除等多个入口，操作密度高，且 TODO 创建/编辑弹窗会因点击空白遮罩而意外关闭。这个变更将 TODO 管理动作收进右键菜单，并收紧 TODO 弹窗关闭方式，减少误触和行内按钮噪音。

## What Changes

- TODO 创建、详情编辑和添加项目弹窗 SHALL 不再因点击空白遮罩关闭，仅通过关闭按钮、取消按钮或成功提交后的流程关闭。
- 活动 TODO 行 SHALL 支持右键菜单。
- 右键菜单 SHALL 包含查看详情、添加项目、复制 TODO 描述和删除 TODO 等管理动作。
- TODO 行外部 SHALL 只保留状态推进入口：`not-started` TODO 保留开始按钮，`in-progress` TODO 保留完成按钮。
- 删除 TODO SHALL 从右键菜单触发，并继续使用确认气泡或等效确认流程，取消删除不改变 TODO 或其终端状态。
- 复制 TODO 描述 SHALL 将该 TODO 的 `description` 字段写入系统剪贴板；描述为空时不写入非描述内容。

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `todo-workspace`: 调整 TODO 弹窗关闭规则、活动 TODO 行动作入口布局、TODO 右键菜单行为和复制描述交互。

## Impact

- Vue 前端：`App.vue` 中 TODO 弹窗关闭方式和剪贴板写入处理；`ProjectSidebar.vue` 中 TODO 行动作、右键菜单和删除入口。
- 客户端测试：更新 `App.test.js` 与 `ProjectSidebar.test.js`，覆盖弹窗遮罩点击不关闭、右键菜单动作、行外按钮可见性和复制描述。
- Wails runtime：复用现有 `ClipboardSetText`，不新增后端 API。
- Go 后端和持久化格式：无预期变更。
