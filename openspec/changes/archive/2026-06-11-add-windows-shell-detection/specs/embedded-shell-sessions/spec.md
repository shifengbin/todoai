## ADDED Requirements

### Requirement: Resolve Platform Terminal Shell For Session Startup
系统 SHALL 在创建新的嵌入式 shell session 前使用平台感知的终端 shell 路径解析，并 SHALL 避免在 Windows 上选择 Unix-only fallback 路径。

#### Scenario: Windows new terminal uses detected shell fallback
- **WHEN** 应用运行在 Windows 上
- **AND** 已保存的终端 shell 设置不可用
- **AND** 自动探测选择 `cmd.exe` 作为 fallback shell
- **AND** 用户为可用项目创建新的嵌入式终端
- **THEN** 新终端的 shell path 解析为 `cmd.exe`
- **AND** shell path 不解析为 `/bin/sh`、`/bin/bash` 或其他 Unix-only fallback

#### Scenario: Windows unsupported PTY startup surfaces startup error
- **WHEN** 应用运行在 Windows 上
- **AND** shell path 已解析为可用的 Windows shell
- **AND** 当前 PTY backend 不支持 Windows session startup
- **AND** 用户创建新的嵌入式终端
- **THEN** 系统报告 shell startup error
- **AND** 系统不通过改用 Unix-only shell path 隐藏该错误

#### Scenario: Non-Windows terminal startup remains unchanged
- **WHEN** 应用运行在非 Windows 系统上
- **AND** 已保存的终端 shell 设置可用
- **AND** 用户创建新的嵌入式终端
- **THEN** 新终端继续使用已保存的 shell path 启动
- **AND** shell process working directory 仍是所属项目路径
