## Why

Windows 用户打开应用后会看到大量控制台窗口反复闪现并立即退出，严重干扰正常使用。当前应用会在启动、项目切换和窗口获得焦点时刷新 Git 状态，并在 Windows GUI 程序中直接启动 `git.exe` 等控制台子进程；这些子进程窗口退出后又可能触发新的焦点刷新，形成闪烁循环。

## What Changes

- 后台执行 Git 状态查询和 Git 初始化时，在 Windows 上隐藏子进程控制台窗口。
- 导入项目后不立即获取 Git 状态，避免批量导入时对最后导入项目触发不必要查询。
- 将 Git 状态查询延迟到用户显式选择项目、展开 TODO 或选择 TODO 项目后再触发。
- 调整窗口 focus 触发的 Git 状态刷新，避免短时间内重复刷新造成命令风暴。
- Windows 上嵌入式 PTY 后端不可用时，终端启动应稳定失败并向界面呈现可理解状态，不应反复尝试或弹出系统控制台窗口。
- 保持现有 Linux/macOS 终端和 Git 状态行为不变。
- 不在本变更中实现 Windows ConPTY 后端。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `project-workspace`: Git 状态刷新在 Windows 上不得显示后台控制台窗口；导入项目不应立即触发 Git 查询；Git 查询应延迟到用户显式选择项目、展开 TODO、选择 TODO 项目或其他明确刷新时机，并防止 focus 事件触发重复刷新循环。
- `embedded-shell-sessions`: Windows 下嵌入式 shell 后端不可用时，终端启动应稳定降级并提示不可用，而不是反复启动失败。

## Impact

- 后端 Git 命令执行路径：`git_status.go`。
- 后端 shell 启动路径和 Windows 平台分支：`shell.go`、可能新增 Windows 专用文件。
- 前端 Git 状态刷新触发逻辑：`frontend/src/App.vue`。
- TODO 展开事件传递：`frontend/src/components/ProjectSidebar.vue`。
- 前端终端不可用或启动失败状态展示：`frontend/src/App.vue` 及相关测试。
- 需要补充 Go 单元测试和 Vue 组件/应用测试；不引入新的外部依赖。
