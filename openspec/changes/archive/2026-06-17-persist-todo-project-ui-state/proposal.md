## Why

当前 TODO 工作区的视图标签和左侧栏宽度只存在于前端内存中，应用重启或重新打开 workspace 后会回到默认值。用户在不同 TODO 工程之间工作时，需要每个 TODO 工程恢复上次选择的 TODO 状态视图和分割线宽度，以减少重复调整。

## What Changes

- 为每个 TODO 工程持久化左侧 TODO 栏宽度。
- 为每个 TODO 工程持久化上次选择的 TODO 视图标签：`未执行`、`执行中` 或 `已完成`。
- 将这些 UI 状态保存到当前 workspace 的 `.data` 目录，并在重新打开软件或 workspace 后恢复。
- 切换 TODO 工程时恢复该 TODO 工程自己的 UI 状态；没有记录时使用现有默认值。
- 删除 TODO 工程时清理对应 UI 状态。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: TODO 工作区需要按 TODO 工程恢复上次选择的 TODO 视图标签和左侧栏宽度。
- `workspace-lifecycle`: workspace `.data` 需要包含 TODO 工程级 UI 状态，并在 workspace 重新打开时加载。

## Impact

- 后端需要新增 workspace 级 UI 状态持久化结构和读写 API，保存到当前 workspace 的 `.data` 目录。
- `App` 需要随 workspace 切换重建或切换该 UI 状态存储。
- 前端 `App.vue` 需要在 active TODO 工程变化时恢复左侧栏宽度，并在拖拽结束后保存。
- 前端 `ProjectSidebar.vue` 需要支持外部传入当前 TODO 视图、在用户切换视图时通知父组件保存。
- Wails 绑定需要暴露加载和保存 TODO 工程 UI 状态的接口。
- 测试需要覆盖重启恢复、TODO 工程切换隔离、默认值和删除 TODO 工程清理状态。
