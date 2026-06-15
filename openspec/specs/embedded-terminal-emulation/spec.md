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

### Requirement: Preserve Unrecognized Launch Profile Text

The system SHALL preserve terminal output that only resembles base64 text, a Windows launch profile text leak, or an unsupported control-text shape. The system MUST only consume payloads that match the supported application-private command-state protocol.

#### Scenario: Unrecognized Windows launch profile text is preserved

- **WHEN** the application runs on Windows
- **AND** launching a non-empty launch profile command surfaces text that looks like base64 or an unsupported control-text fragment
- **AND** that text does not match the supported application-private command-state protocol
- **THEN** the system renders that text normally
- **AND** the system persists that text in terminal history normally

#### Scenario: Ordinary base64-like output is preserved

- **WHEN** terminal output contains base64-like text that does not match the supported application-private command-state protocol
- **THEN** the system renders that text normally
- **AND** the system persists that text in terminal history normally

### Requirement: Track Terminal Activity From Title Changes

The system SHALL track terminal activity from captured terminal title changes using a time-based fallback. Any title update SHALL record the latest runtime title and mark the terminal `busy` with source `title-fallback` when no newer higher-priority unified agent status applies. If no further title update is received for 1 second, the title fallback status SHALL return to `idle`. Title-derived status SHALL be treated as low-confidence fallback and MUST NOT override newer higher-priority unified agent status from shell lifecycle, command-state, Claude/Codex structured events, or machine-readable agent streams.

#### Scenario: Title change marks terminal busy

- **WHEN** terminal A receives a title update `codex thinking`
- **AND** terminal A has no newer higher-priority agent status
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase becomes `busy`
- **AND** the status source is `title-fallback`

#### Scenario: Repeated title changes keep terminal busy

- **WHEN** terminal A is `busy` from title fallback activity
- **AND** terminal A receives another title update before 1 second elapses
- **THEN** terminal A remains `busy`
- **AND** the 1-second idle timeout restarts from the latest title update

#### Scenario: No title change for one second returns terminal idle

- **WHEN** terminal A is `busy` from title fallback activity
- **AND** terminal A receives no title updates for 1 second
- **THEN** terminal A's unified agent activity phase becomes `idle`
- **AND** the status source remains `title-fallback`

#### Scenario: Title text is not semantically classified

- **WHEN** terminal A receives title updates containing a Windows path, Unix path, spinner character, Claude dot frame, or attention marker
- **THEN** the system records terminal A's latest runtime title
- **AND** those title strings follow the same title-change busy and 1-second idle timeout rule
- **AND** the system does not infer `needs-input` or long-lived `busy` from the title text itself

#### Scenario: Structured status is not overridden by title fallback

- **WHEN** terminal A has unified agent activity phase `needs-input` from a Claude hook notification
- **AND** terminal A receives a terminal title update `claude thinking`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase remains `needs-input`
- **AND** terminal A's unified agent status source remains the Claude hook source

### Requirement: Suppress Internal Command-State Payloads

The system SHALL consume application-private command-state payloads before embedded terminal output is rendered or persisted. The system MUST NOT display or replay base64 command-state payloads as terminal text. New command-state payloads SHALL use the `todoai` application identifier. The system MUST continue to consume legacy `tui-helper` command-state payloads. When a command-state payload is invalid but recognizable as application-private, the system SHALL drop the payload and preserve the previous terminal command state.

#### Scenario: Raw command-state OSC is not rendered or persisted

- **WHEN** an embedded terminal output chunk contains `ESC ] 777 ; todoai ; command-start ; bnBtIHRlc3Q= BEL`
- **THEN** the visible terminal output excludes the OSC payload
- **AND** the persisted terminal history excludes the OSC payload

#### Scenario: Windows ConPTY textual command-state payload is not rendered or persisted

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;todoai;command-start;Y29kZXg=` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload

#### Scenario: Invalid Windows textual command-state payload is dropped

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;todoai;command-start;not-base64` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload
- **AND** the terminal command label is not updated from the invalid payload

#### Scenario: Split command-state payload is consumed across output chunks

- **WHEN** one terminal output read ends with `ESC ] 777 ; todoai ; command-start ;`
- **AND** the following read contains `Y2xhdWRl BEL`
- **THEN** the system consumes the complete command-state payload
- **AND** neither output chunk renders or persists the partial payload

#### Scenario: Legacy raw command-state OSC is still consumed

- **WHEN** an embedded terminal output chunk contains `ESC ] 777 ; tui-helper ; command-start ; bnBtIHRlc3Q= BEL`
- **THEN** the visible terminal output excludes the OSC payload
- **AND** the persisted terminal history excludes the OSC payload

#### Scenario: Legacy Windows textual command-state payload is still consumed

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;tui-helper;command-start;Y29kZXg=` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload

#### Scenario: Non-application terminal output is preserved

- **WHEN** terminal output contains ordinary command output that does not match the application-private `todoai` or `tui-helper` command-state protocol
- **THEN** the system renders that output normally
- **AND** the system persists that output in terminal history normally

### Requirement: Focus Terminal After Tree Selection

系统 SHALL 在用户从左侧 TODO 终端树选择终端并成功激活对应右侧嵌入式终端后，将键盘输入焦点转移到该嵌入式终端。该自动聚焦 MUST 只由用户明确选择终端的交互触发，后台状态更新、初始化恢复或其他非选择终端路径不得因此抢占焦点。

#### Scenario: User selects a terminal from todo tree

- **WHEN** 用户在左侧 TODO 终端树中点击终端 B
- **AND** 系统成功将活动终端切换为终端 B
- **THEN** 右侧终端区域显示终端 B
- **AND** 终端 B 的嵌入式终端获得键盘输入焦点

#### Scenario: Terminal selection fails

- **WHEN** 用户在左侧 TODO 终端树中点击终端 B
- **AND** 系统未能成功将活动终端切换为终端 B
- **THEN** 系统不得将键盘输入焦点转移到终端 B
