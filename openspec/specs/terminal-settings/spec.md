# terminal-settings Specification

## Purpose
Defines local terminal shell selection settings used when creating embedded shell sessions.
## Requirements
### Requirement: Persist Terminal Shell Setting

The system SHALL persist the embedded terminal shell setting locally and SHALL reload it when the application starts.

#### Scenario: First startup detects and persists shell

- **WHEN** the application loads terminal settings and no saved terminal shell setting exists
- **THEN** the system detects an available shell
- **AND** the system saves the detected shell path as the terminal shell setting
- **AND** the settings state exposes the saved shell path and display name

#### Scenario: Saved shell setting is restored

- **WHEN** the user has previously saved `/usr/bin/zsh` as the terminal shell setting
- **AND** the application starts again
- **THEN** the settings state exposes `/usr/bin/zsh` as the selected terminal shell path
- **AND** automatic detection is not used to replace the saved path

### Requirement: Change Terminal Shell Setting

The system SHALL allow the user to change the embedded terminal shell setting from the settings interface.

#### Scenario: User saves detected shell option

- **WHEN** the settings interface shows `/usr/bin/bash` as an available shell option
- **AND** the user selects `/usr/bin/bash` and saves
- **THEN** the terminal shell setting is persisted as `/usr/bin/bash`
- **AND** newly created embedded terminals use `/usr/bin/bash`

#### Scenario: User saves manual shell path

- **WHEN** the user enters `/opt/custom/bin/fish` as a manual shell path
- **AND** the path exists and is executable
- **AND** the user saves
- **THEN** the terminal shell setting is persisted as `/opt/custom/bin/fish`
- **AND** the settings state exposes `fish` as the display name

#### Scenario: User enters invalid manual shell path

- **WHEN** the user enters `/missing/shell` as a manual shell path
- **AND** the path does not exist or is not executable
- **THEN** the system rejects the setting
- **AND** the previously saved terminal shell setting remains unchanged

### Requirement: Re-detect Terminal Shell Setting

The system SHALL allow the user to re-run automatic terminal shell detection from the settings interface.

#### Scenario: User re-runs shell detection

- **WHEN** the user triggers shell detection from settings
- **THEN** the system detects an available shell
- **AND** the settings interface shows the detected shell as the current candidate

#### Scenario: User saves re-detected shell

- **WHEN** shell detection returns `/usr/bin/zsh`
- **AND** the user saves the detected result
- **THEN** the terminal shell setting is persisted as `/usr/bin/zsh`

### Requirement: Report Unavailable Saved Shell

The system SHALL report when the saved terminal shell path is unavailable and SHALL provide a detected fallback for continued terminal startup.

#### Scenario: Saved shell path is unavailable

- **WHEN** the saved terminal shell setting is `/old/bin/zsh`
- **AND** `/old/bin/zsh` does not exist or is not executable
- **AND** the application loads terminal settings
- **THEN** the settings state marks the saved shell as unavailable
- **AND** the settings state includes an automatically detected fallback shell

#### Scenario: Unavailable saved shell does not prevent terminal startup

- **WHEN** the saved terminal shell setting is unavailable
- **AND** an automatically detected fallback shell exists
- **AND** the user creates a new embedded terminal
- **THEN** the new embedded terminal starts with the fallback shell
- **AND** the saved terminal shell setting remains unchanged until the user saves a new setting

### Requirement: Persist Terminal Launch Profiles

The system SHALL persist configurable terminal launch profiles in terminal settings and SHALL expose their name, startup parameters, enabled state, and saved order when settings are loaded. The built-in `Terminal` launch option SHALL NOT be persisted as a configurable profile.

#### Scenario: Missing launch profiles use defaults

- **WHEN** the application loads terminal settings from an existing settings file that has no launch profiles field
- **THEN** the settings state includes launch profiles named `codex` and `claude`
- **AND** the `codex` profile has startup parameters `codex`
- **AND** the `claude` profile has startup parameters `claude`
- **AND** both default launch profiles are enabled

#### Scenario: Saved launch profiles are restored

- **WHEN** the user has previously saved launch profiles named `Codex GPT-5` and `Claude Plan`
- **AND** the application loads terminal settings
- **THEN** the settings state exposes those launch profile names in the saved order
- **AND** each launch profile exposes its saved startup parameters
- **AND** each launch profile exposes its saved enabled state

#### Scenario: Legacy launch profiles without enabled state remain enabled

- **WHEN** the application loads terminal settings from an existing settings file whose launch profiles do not include an enabled state
- **THEN** each launch profile is exposed as enabled
- **AND** the existing launch profile names, startup parameters, and order remain unchanged

#### Scenario: Empty launch profile list remains empty

- **WHEN** the user has previously saved an empty launch profile list
- **AND** the application loads terminal settings
- **THEN** the settings state exposes no custom launch profiles
- **AND** the built-in `Terminal` launch option remains available outside the configurable profile list

### Requirement: Change Terminal Launch Profiles

The system SHALL allow the user to add, edit, reorder, enable, disable, and remove configurable terminal launch profiles from the settings interface.

#### Scenario: User saves valid launch profiles

- **WHEN** the user configures a launch profile named `Codex` with startup parameters `codex --model gpt-5`
- **AND** the user saves settings
- **THEN** the launch profile is persisted with name `Codex`
- **AND** the launch profile is persisted with startup parameters `codex --model gpt-5`
- **AND** the launch profile is persisted as enabled

#### Scenario: User disables a launch profile

- **WHEN** settings contains an enabled launch profile named `Claude Plan`
- **AND** the user disables `Claude Plan` and saves settings
- **THEN** the launch profile remains persisted with its name and startup parameters
- **AND** the launch profile is persisted as disabled

#### Scenario: User enables a disabled launch profile

- **WHEN** settings contains a disabled launch profile named `Claude Plan`
- **AND** the user enables `Claude Plan` and saves settings
- **THEN** the launch profile remains persisted with its name and startup parameters
- **AND** the launch profile is persisted as enabled

#### Scenario: User removes a launch profile

- **WHEN** settings contains launch profiles named `codex` and `claude`
- **AND** the user removes the `claude` profile and saves settings
- **THEN** the settings state includes the `codex` launch profile
- **AND** the settings state does not include the `claude` launch profile

#### Scenario: Invalid launch profile is rejected

- **WHEN** the user configures a launch profile with an empty name or empty startup parameters
- **AND** the user saves settings
- **THEN** the system rejects the setting
- **AND** the previously saved launch profiles remain unchanged

#### Scenario: Launch profile name conflicts with built-in terminal

- **WHEN** the user configures a custom launch profile named `Terminal`
- **AND** the user saves settings
- **THEN** the system rejects the setting
- **AND** the built-in `Terminal` launch option remains unchanged

### Requirement: Display Enabled Terminal Launch Profiles

The system SHALL include only enabled configurable terminal launch profiles in terminal launch menus. The built-in `Terminal` launch option SHALL remain available regardless of configurable launch profile states.

#### Scenario: Disabled launch profile is hidden from launch menu

- **WHEN** terminal settings include an enabled launch profile named `codex`
- **AND** terminal settings include a disabled launch profile named `claude`
- **AND** the user opens a terminal launch menu
- **THEN** the launch menu includes `Terminal`
- **AND** the launch menu includes `codex`
- **AND** the launch menu does not include `claude`

#### Scenario: Launch menu works when all custom profiles are disabled

- **WHEN** all configurable launch profiles are disabled
- **AND** the user opens a terminal launch menu
- **THEN** the launch menu includes `Terminal`
- **AND** the launch menu does not include any custom launch profile

#### Scenario: Enabled launch profile can start with its command

- **WHEN** terminal settings include an enabled launch profile named `Codex GPT-5` with startup parameters `codex --model gpt-5`
- **AND** the user selects `Codex GPT-5` from the launch menu
- **THEN** the system creates a terminal
- **AND** the system submits `codex --model gpt-5` to the created terminal

### Requirement: Persist Appearance Theme Setting
The system SHALL persist the application appearance theme in terminal settings and SHALL expose the theme when terminal settings are loaded.

#### Scenario: Missing theme setting uses default
- **WHEN** the application loads terminal settings from an existing settings file that has no theme field
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system preserves existing terminal shell and launch profile settings

#### Scenario: Saved theme setting is restored
- **WHEN** the user has previously saved `dark` as the appearance theme
- **AND** the application loads terminal settings
- **THEN** the settings state exposes `dark` as the appearance theme

#### Scenario: Invalid saved theme is normalized
- **WHEN** the application loads terminal settings from a settings file with an unsupported theme value
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system does not reject the settings file

### Requirement: Change Appearance Theme Setting
The system SHALL allow the user to save the application appearance theme from the settings interface.

#### Scenario: User saves valid appearance theme
- **WHEN** the user selects `dark` as the appearance theme
- **AND** the user saves settings
- **THEN** the appearance theme is persisted as `dark`
- **AND** the settings state exposes `dark` as the appearance theme

#### Scenario: User saves unsupported appearance theme
- **WHEN** the application receives an unsupported appearance theme value
- **THEN** the system rejects the setting
- **AND** the previously saved appearance theme remains unchanged

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
