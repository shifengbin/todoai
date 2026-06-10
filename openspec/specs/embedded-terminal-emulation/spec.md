# embedded-terminal-emulation Specification

## Purpose
TBD - created by archiving change fix-terminal-emulation. Update Purpose after archive.
## Requirements
### Requirement: Start Shell With Xterm-Compatible Environment

The system SHALL start embedded PTY-backed shell processes with terminal capability variables that match the xterm.js renderer.

#### Scenario: Shell environment overrides inherited dumb terminal

- **WHEN** the desktop process environment contains `TERM=dumb` and an embedded shell is started
- **THEN** the shell process environment contains `TERM=xterm-256color`
- **AND** the shell process environment contains `COLORTERM=truecolor`

### Requirement: Render PTY Output Without Redundant Newline Conversion

The system SHALL render embedded shell output as PTY output without enabling client-side newline conversion intended for non-PTY data sources.

#### Scenario: Terminal session is created for PTY-backed output

- **WHEN** a project terminal session is created
- **THEN** the xterm.js terminal is configured without `convertEol` newline conversion

### Requirement: Preserve Shell Editing Feedback

The system SHALL preserve normal interactive shell editing feedback for typed input and deletion keys in the embedded terminal.

#### Scenario: User deletes typed input

- **WHEN** the user types `clear`, presses Backspace twice, types `ar`, and presses Enter in the active embedded terminal
- **THEN** the shell receives the edited command as `clear`
- **AND** the terminal display does not retain stale characters from the intermediate input

### Requirement: Clear Screen Uses Terminal Control Sequences

The system SHALL support clear-screen behavior from standard terminal commands in the embedded shell.

#### Scenario: User runs clear in zsh

- **WHEN** the user runs `clear` in an embedded zsh session
- **THEN** the terminal viewport is cleared using xterm-compatible control sequences
- **AND** previously typed command text is not left visible as residual output

### Requirement: Prevent Duplicate Project Shell Starts

The system SHALL maintain at most one live PTY process per project even when project activation is requested concurrently.

#### Scenario: Concurrent activation requests same project shell

- **WHEN** two shell start requests for the same available project overlap
- **THEN** the backend starts only one PTY process for that project
- **AND** both requests resolve to the same running shell session state

### Requirement: Capture Terminal Title Changes

系统 SHALL 捕获嵌入式 xterm 会话收到的终端标题变化，并 SHALL 将标题变化关联到产生该变化的 terminal。

#### Scenario: Interactive program updates terminal title

- **WHEN** terminal A 中运行的交互式程序发送 OSC 0 或 OSC 2 标题更新
- **THEN** 系统记录 terminal A 的最新运行时标题
- **AND** 该标题更新不会被关联到其他 terminal

#### Scenario: Inactive terminal updates title

- **WHEN** terminal A 处于后台
- **AND** terminal B 是当前激活终端
- **AND** terminal A 中运行的交互式程序发送标题更新
- **THEN** 系统更新 terminal A 的运行时标题状态
- **AND** terminal B 的运行时标题状态保持不变

#### Scenario: Program does not emit title changes

- **WHEN** terminal A 中运行的程序不发送终端标题更新
- **THEN** 系统保持 terminal A 的现有终端渲染和命令标签行为
- **AND** 不为 terminal A 生成交互式活动状态
