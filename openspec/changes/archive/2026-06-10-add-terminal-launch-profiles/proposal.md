## Why

项目侧边栏的项目加号目前只能直接创建普通终端，用户如果经常启动 Codex、Claude 等交互式命令，需要先建终端再手动输入命令。将加号改为可配置启动菜单，可以把常用交互式工具变成一次点击，同时保留普通终端入口。

## What Changes

- 项目加号从直接创建终端改为打开启动菜单。
- 启动菜单始终包含内置 `Terminal` 选项，选择后只创建普通终端。
- settings 增加可配置的终端启动 profiles 列表，默认包含 `codex` 和 `claude`。
- 每个自定义启动 profile 支持配置显示名称和启动参数。
- 选择自定义启动 profile 时，系统创建新终端，并在终端启动后执行该 profile 的启动参数。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `terminal-settings`: 增加持久化和编辑终端启动 profiles 的设置要求。
- `project-workspace`: 项目树加号需要展示启动菜单，并用所选启动项创建终端。
- `embedded-shell-sessions`: 支持创建终端后执行所选启动 profile 的启动参数。

## Impact

- 前端：项目侧边栏加号交互、启动菜单 UI、settings 弹窗的启动 profile 编辑区、相关组件测试。
- 后端：settings 状态结构、持久化配置读写、默认启动 profile 补齐与校验、相关 Go 测试。
- Wails API：可能扩展 terminal settings 结构；如新增保存启动 profiles 的方法，需要重新生成 Wails bindings。
- 兼容性：已有 settings 文件应继续可读，缺少启动 profiles 时自动使用默认 `codex` 和 `claude`。
