## ADDED Requirements

### Requirement: Detect Windows Terminal Shell
系统 SHALL 在 Windows 上使用 Windows-aware 的候选顺序自动探测终端 shell，并 SHALL 返回可保存到终端设置中的路径、显示名、来源和可用状态。

#### Scenario: Windows detects PowerShell 7 first
- **WHEN** 应用在 Windows 上加载终端设置
- **AND** 没有已保存的终端 shell 设置
- **AND** `pwsh.exe` 可通过 PATH 或候选路径找到
- **THEN** 系统将 PowerShell 7 探测为终端 shell
- **AND** 探测结果的 source 是 `detected`
- **AND** 探测结果标记为 available

#### Scenario: Windows falls back to Windows PowerShell
- **WHEN** 应用在 Windows 上加载终端设置
- **AND** 没有已保存的终端 shell 设置
- **AND** `pwsh.exe` 不可用
- **AND** `powershell.exe` 可通过 PATH 或系统目录候选路径找到
- **THEN** 系统将 Windows PowerShell 探测为终端 shell
- **AND** 不返回 Unix-only shell 路径

#### Scenario: Windows falls back to Cmd
- **WHEN** 应用在 Windows 上加载终端设置
- **AND** PowerShell 7 和 Windows PowerShell 都不可用
- **AND** `COMSPEC` 指向可用的 `cmd.exe`
- **THEN** 系统将 Cmd 探测为终端 shell
- **AND** 不返回 `/bin/sh`、`/bin/bash` 或其他 Unix-only fallback

#### Scenario: Unix shell detection remains unchanged
- **WHEN** 应用在非 Windows 系统上加载终端设置
- **AND** 没有已保存的终端 shell 设置
- **THEN** 系统继续优先使用 `$SHELL`
- **AND** 在 `$SHELL` 不可用时继续检查已知 Unix shell 候选路径

### Requirement: Validate Windows Terminal Shell Path
系统 SHALL 在 Windows 上按 Windows 可执行文件语义验证终端 shell 路径，并 SHALL 在非 Windows 系统上保留现有 Unix execute bit 校验。

#### Scenario: Windows accepts executable shell extension
- **WHEN** 用户在 Windows 上手动保存终端 shell 路径
- **AND** 路径存在、不是目录
- **AND** 路径扩展名是 `.exe`、`.cmd`、`.bat`、`.com` 或 `PATHEXT` 中声明的扩展名
- **THEN** 系统接受该终端 shell 设置
- **AND** 新设置被保存为 selected terminal shell

#### Scenario: Windows rejects non-executable shell path
- **WHEN** 用户在 Windows 上手动保存终端 shell 路径
- **AND** 路径不存在、是目录，或扩展名不是 Windows 可执行扩展
- **THEN** 系统拒绝该终端 shell 设置
- **AND** 之前已保存的终端 shell 设置保持不变

#### Scenario: Unix validation still requires execute permission
- **WHEN** 用户在非 Windows 系统上手动保存终端 shell 路径
- **AND** 路径存在但没有 execute permission
- **THEN** 系统拒绝该终端 shell 设置
- **AND** 之前已保存的终端 shell 设置保持不变
