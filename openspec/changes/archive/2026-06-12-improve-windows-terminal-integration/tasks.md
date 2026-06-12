## 1. Launch Profile 提交与标签

- [x] 1.1 将 launch profile 自动提交从普通换行改为交互式终端 Enter 序列，并限制在自动提交路径，不影响普通输入和粘贴。
- [x] 1.2 在提交 launch profile 命令时同步更新目标 terminal 的 `currentCommand`，复用现有命令标签清理规则。
- [x] 1.3 确保 shell `command-start` 和 `command-end` 事件仍能覆盖或清空应用侧设置的 launch profile 标签。
- [x] 1.4 更新 `frontend/src/App.test.js`，覆盖 Windows/通用 launch profile 使用 Enter 提交、带参数命令标签、无 launch profile 不提交命令。

## 2. Windows Shell 命令状态集成

- [x] 2.1 扩展 shell 启动集成识别 `pwsh` 和 `powershell`，为其生成临时 PowerShell 集成脚本或 profile 参数。
- [x] 2.2 在 PowerShell 集成中发出 OSC 777 `command-start` 和 `command-end`，协议与 zsh/bash 现有实现保持一致。
- [x] 2.3 保留用户 PowerShell profile 的加载路径，并在 shell session cleanup 时清理临时文件。
- [x] 2.4 为 PowerShell 集成启动参数、临时文件清理、非 PowerShell shell fallback 添加 Go 单元测试。

## 3. 终端标题活动状态

- [x] 3.1 调整标题 busy 判定，移除普通 `/` 和 `\` 路径分隔符导致的 busy 误判。
- [x] 3.2 保留明确 spinner 字符、`working`/`thinking`/`running` 等 busy 文本和 `!` needs-input 信号。
- [x] 3.3 更新前端测试，覆盖 Windows 路径标题、Unix 路径标题、spinner 标题、busy 文本标题、needs-input 标题和 idle baseline。

## 4. 验证与收尾

- [x] 4.1 运行 `go test ./...`，修复与 shell 启动、Windows ConPTY、settings 相关的失败。
- [x] 4.2 运行 `cd frontend && npm test`，修复客户端自动化测试失败。
- [x] 4.3 运行 OpenSpec 校验命令确认 proposal/design/specs/tasks 格式和需求 delta 可归档。
- [x] 4.4 执行自动代码 review，重点检查跨平台行为、临时文件清理、命令标签状态回退和测试覆盖。
- [x] 4.5 运行 `wails build -tags webkit2_41`，生成可执行文件。
