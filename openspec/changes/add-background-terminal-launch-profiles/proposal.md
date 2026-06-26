## Why

当前终端启动下拉菜单中的自定义 profile 都会创建一个可见嵌入式终端，再把命令发送到该终端执行。部分命令只需要在对应 TODO 上下文中后台运行，不应该占用 UI 终端列表、切换当前终端或留下退出后的终端记录。

## What Changes

- 在终端 launch profile 配置中增加“后台启动”选项，默认关闭，并兼容旧配置。
- 后台启动的 profile 仍显示在现有终端下拉菜单中，但选择后不会新增 UI 终端、不会切换当前终端、不会写入终端历史。
- 后台启动命令在对应上下文目录中一次性执行：TODO 项目菜单使用该 TODO 项目的 prepared worktree，任务级菜单使用 TODO 任务工作区。
- 后台程序启动成功后由应用等待并自动回收；命令结束不产生 UI 状态变化。启动失败沿用现有错误显示方式。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `terminal-settings`: launch profile 配置和下拉菜单行为增加后台启动模式。
- `embedded-shell-sessions`: 增加不注册嵌入式终端的后台 profile 命令启动语义。

## Impact

- 后端 Go：terminal settings schema、settings 读写/迁移、App 后台启动 API、后台命令执行器。
- 前端 Vue：Terminal Settings profile 表单、ProjectSidebar 下拉菜单、profile 点击分发、Wails 绑定。
- 测试：settings 单元测试、App 后台启动上下文测试、前端 launch profile 菜单/点击行为测试。
