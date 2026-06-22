## Why

用户在 `执行中` 视图选择某个终端后，再切换到 `未执行` 并点击项目 item，会被自动切回 `执行中`。这会把一次普通的项目选择误表现为视图切换，打断用户当前查看的 TODO 状态列表。

## What Changes

- 调整 TODO 工作区中点击项目 item 的行为：选择项目上下文时保持用户当前 TODO 视图标签，不再自动恢复该 TODO 工程保存过的视图标签。
- 保留打开 workspace、重新加载 workspace 等恢复场景中按 TODO 工程恢复保存视图的能力。
- 保留终端选择、终端聚焦、active TODO 工程和 active terminal 的现有业务语义。
- 增加覆盖“执行中选择终端 -> 切到未执行 -> 点击项目 item 不回跳执行中”的前端回归测试。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `todo-workspace`: 收窄 TODO 工程视图恢复规则，明确普通项目 item 点击不得覆盖当前 TODO 视图标签。

## Impact

- 影响前端 TODO 项目选择流程：`frontend/src/App.vue`。
- 影响前端 TODO 工作区测试：`frontend/src/App.test.js`。
- 不改变 Go 后端 API、Wails 绑定、终端会话模型或持久化格式。
