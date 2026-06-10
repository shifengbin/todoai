## Context

当前项目侧边栏的项目加号直接触发 `create-terminal`，前端调用 Wails `CreateTerminal`，后端创建并启动一个独立 shell session。settings 目前只保存终端 shell 路径，缺少“常用交互式命令启动项”的配置。用户希望加号弹出启动菜单，内置普通终端入口，并把 `codex`、`claude` 这类启动项作为可配置列表追加在菜单后面。

这个变更横跨 Vue 侧边栏、settings UI、Go settings 持久化和终端输入路径，但不需要改变 PTY 的核心启动模型。

## Goals / Non-Goals

**Goals:**

- 项目加号打开启动菜单，而不是立即创建终端。
- 菜单第一项始终是内置 `Terminal`，不需要也不允许在 settings 中配置。
- settings 持久化自定义启动 profiles，默认包含 `codex` 和 `claude`。
- 每个自定义 profile 支持配置显示名称和启动参数。
- 选择自定义 profile 时，系统创建新终端并在 shell 启动后提交启动参数。
- 旧 settings 文件缺少启动 profiles 时自动补默认值。

**Non-Goals:**

- 不引入新的终端进程模型，不用 profile 命令替代 configured shell。
- 不做命令可执行性检测；启动参数由用户配置并交给 shell 执行。
- 不把内置 `Terminal` 做成可删除或可重命名项。
- 不支持 per-project profile；profiles 是全局 settings。

## Decisions

### 1. 使用 settings 中的 `launchProfiles` 保存自定义启动项

新增 `TerminalLaunchProfileSetting`，字段为 `name` 和 `command`。`TerminalSettingsState` 与持久化 settings 都增加 `launchProfiles`。内置 `Terminal` 不保存，前端展示菜单时始终把它放在第一项，再追加 settings 返回的 profiles。

备选方案是把 `Terminal` 也保存到列表里，但这会让用户误删或改坏普通终端入口。将其作为内置项可以保证基础创建终端能力始终存在。

### 2. 缺失 profiles 字段时补默认值，显式空数组保持为空

旧 settings 文件没有 `launchProfiles` 字段时，加载后返回默认 profiles：`codex` 和 `claude`。如果用户明确保存 `[]`，系统应保持空自定义列表，只显示内置 `Terminal`。Go 持久化层需要区分 JSON 字段缺失和显式空数组。

### 3. 通过现有终端输入通道执行 profile 命令

选择自定义 profile 后，前端先调用现有 `CreateTerminal`，应用返回 state 后获取新 active terminal id，再调用 `SendTerminalInput(terminalId, command + "\n")`。这样 profile 命令运行在用户配置的 shell 内，命令退出后仍回到同一个 shell。

备选方案是在后端新增 `CreateTerminalWithCommand`。这会让命令执行语义更集中，但需要扩展 Wails API、shell manager 和测试面；当前已有输入 API 能准确表达“创建终端后输入命令”。另一个备选是通过 shell 启动参数执行命令，但这会改变交互式 shell 生命周期，不适合 Codex/Claude 这类长交互程序。

### 4. 保存时校验 profile 名称和启动参数

自定义 profile 的 `name` 和 `command` 去除首尾空白后都必须非空。名称不能与内置 `Terminal` 冲突，且自定义 profile 名称不应重复。启动参数不做可执行性校验，保留用户传入的参数字符串。

## Risks / Trade-offs

- [Risk] shell 刚创建后立即写入命令可能遇到 session 尚未可写的时序问题 -> 在 `CreateTerminal` 成功返回后再发送输入，并复用现有 `SendTerminalInput` 的错误处理。
- [Risk] 旧 settings 保存 shell 路径时覆盖新增 profiles -> settings 保存逻辑必须加载旧状态、更新目标字段、保留其他字段。
- [Risk] 用户配置无效命令导致 shell 报错 -> 不在 settings 阶段阻断，命令执行错误由 shell 自身展示。
- [Risk] 菜单在 settings 尚未加载时显示过期 profiles -> App 启动时加载 terminal settings，失败时使用默认 profiles 并显示错误。

## Migration Plan

- 读取旧 settings 时，如果缺少 `launchProfiles` 字段，返回并保存默认 `codex`、`claude` profiles。
- 读取到显式空数组时不补默认值，保留用户配置。
- 保存 shell setting 时保留已有 launch profiles；保存 launch profiles 时保留已有 shell setting。
- 回滚时旧版本会忽略无法识别的新增字段风险较低；如旧版本重写 settings，profiles 可能丢失。

## Open Questions

无。
