## Why

当前应用外壳以浅色为主，用户无法根据环境光线或偏好切换终端以外页面的视觉模式。添加持久化的 Light/Dark 主题可以让侧栏、工作区、设置弹窗和状态栏等应用页面根据偏好切换，同时保持嵌入式终端既有配色不变。

## What Changes

- 增加应用级外观主题设置，支持 `light` 和 `dark` 两个固定选项。
- 默认主题为 `light`，兼容现有未保存主题的设置文件。
- 设置界面增加外观主题选择，并在保存后持久化用户偏好。
- 应用启动时加载已保存主题，并应用到终端以外的全局 UI。
- 主题切换不改变嵌入式终端内容区、xterm 配色或终端右键菜单配色。
- 不引入跟随系统主题、更多自定义颜色或主题导入导出。

## Capabilities

### New Capabilities

- `application-appearance`: 定义应用级 Light/Dark 外观主题的加载、切换、持久化和终端同步行为。

### Modified Capabilities

- `terminal-settings`: 终端设置状态扩展为包含外观主题偏好，并允许设置界面保存该偏好。

## Impact

- Go 设置模型与持久化文件：`settings.go`、`settings_test.go`。
- Wails 后端 API：`app.go` 及生成的前端绑定。
- Vue 应用状态和设置弹窗：`frontend/src/App.vue`、相关测试。
- 终端创建和会话管理：`frontend/src/xtermFactory.js`、`frontend/src/terminalManager.js`、相关测试，用于验证终端配色不随应用主题变化。
- 全局样式变量与 Light/Dark 颜色集：`frontend/src/style.css`。
