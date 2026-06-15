## Why

Windows 下仍存在几类影响日常使用的终端和侧栏交互问题：TODO 删除相关浮层/菜单位置不稳定，通过创建终端下拉菜单启动非纯 `Terminal` 的 launch profile 时可能显示一长串类似 base64 的未知控制文本，Claude 刚启动或空闲时可能被误判为忙碌状态。这些问题会干扰用户判断当前终端状态，也会让删除确认操作显得脱离当前 TODO 上下文。

这些现象都发生在 Windows WebView + ConPTY + xterm.js 的边界上，需要把浮层定位、终端私有控制序列过滤和交互式程序活动状态分类一起收敛成可测试的行为。

## What Changes

- 修正 Windows 下 TODO 删除菜单/确认气泡的定位行为，使其锚定到触发操作的 TODO 上下文附近，并避免跑到应用全局右上角。
- 收窄终端输出过滤与诊断覆盖：继续隐藏应用私有命令状态序列，但不再对 Windows launch profile 启动时出现的类似 base64 未知文本做特殊启发式隐藏。
- 调整 Claude 等交互式终端标题活动分类规则，避免 Claude 空闲或初始标题变化被长期误判为 `busy`。
- 保留明确忙碌信号和需要输入信号：真实的工作中、思考中、spinner 或 `!` 提示仍应反映到 TODO 终端树。
- 增加针对 Windows 形态的单元测试和必要的手工验证步骤。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: TODO 删除相关菜单/确认气泡必须在对应 TODO 上下文附近显示，不能脱离到全局角落。
- `embedded-terminal-emulation`: 终端渲染必须继续隐藏应用私有控制序列，并且交互式程序标题活动分类不能把 Claude 初始/空闲标题误判为忙碌。
- `embedded-shell-sessions`: Windows PowerShell/ConPTY 会话中的命令状态集成和启动 profile 标签行为必须不向用户暴露内部 payload，并与活动状态判断保持一致。

## Impact

- 前端：`frontend/src/components/ProjectSidebar.vue`、`frontend/src/style.css`、`frontend/src/App.vue`、`frontend/src/xtermFactory.js`、相关 Vitest。
- 后端：`command_state_output_filter.go`、`shell.go`、Windows/PowerShell 集成相关 Go 测试。
- 运行环境：Windows WebView2、ConPTY、PowerShell/pwsh、launch profile 启动路径、Claude 交互式终端标题行为。
- 不引入新的外部依赖，不改变已保存项目、TODO、终端历史和 launch profile 数据格式。
