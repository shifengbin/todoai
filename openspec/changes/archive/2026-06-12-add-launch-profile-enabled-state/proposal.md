## Why

当前 Launch profiles 只能新增、编辑、排序或删除。用户如果只是暂时不想在启动菜单中看到某个 profile，只能删除配置并在之后手动重建。

添加启用状态可以保留 profile 的名称和启动命令，同时让启动菜单只展示当前需要使用的入口。

## What Changes

- 为可配置的 terminal launch profile 增加启用状态。
- 设置界面允许用户切换每个 launch profile 是否启用。
- 禁用的 launch profile 继续保存在设置中，并可在设置界面重新启用、编辑、排序或删除。
- 左侧终端启动菜单隐藏禁用的自定义 launch profile。
- 内置 `Terminal` 启动选项保持可用，不作为可禁用的自定义 profile 持久化。
- 旧设置文件中没有启用状态的 launch profile 按启用处理，避免升级后现有 profile 被隐藏。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `terminal-settings`: 扩展 terminal launch profile 的持久化和设置界面行为，支持启用状态以及启动菜单过滤。

## Impact

- Go settings model and persistence: `TerminalLaunchProfileSetting`、加载/保存/迁移和验证逻辑。
- Wails generated bindings: launch profile TypeScript model。
- Vue settings UI: Launch profiles 行增加启用切换，并在保存时提交状态。
- Vue sidebar launch menu: 只展示启用的自定义 launch profiles。
- Tests: settings persistence, frontend settings interactions, and launch menu filtering。
