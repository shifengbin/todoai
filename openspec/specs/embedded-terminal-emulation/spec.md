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

### Requirement: Classify Terminal Activity From Titles

The system SHALL classify terminal activity from captured terminal title changes without treating ordinary Windows path separators as busy indicators. The system SHALL continue to surface explicit busy and needs-input title signals.

#### Scenario: Windows path title remains idle

- **WHEN** the application runs on Windows
- **AND** terminal A receives a title update `C:\Users\developer\repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `\`

#### Scenario: Unix path title remains idle

- **WHEN** terminal A receives a title update `/home/developer/repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `/`

#### Scenario: Explicit busy title signal marks terminal busy

- **WHEN** terminal A receives a title update `codex thinking`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Spinner title signal marks terminal busy

- **WHEN** terminal A receives a title update containing an explicit spinner character such as `⠋`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Needs input title signal marks terminal needs input

- **WHEN** terminal A receives a title update `codex !`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `needs-input`

#### Scenario: Stable title establishes idle baseline

- **WHEN** terminal A is running without a prior idle title
- **AND** terminal A receives a stable title update matching its shell or current command label
- **THEN** the system records that title as terminal A's idle title
- **AND** terminal A's activity state remains `idle`

### Requirement: Suppress Internal Command-State Payloads

The system SHALL consume application-private command-state payloads before embedded terminal output is rendered or persisted. The system MUST NOT display or replay base64 command-state payloads as terminal text.

#### Scenario: Raw command-state OSC is not rendered or persisted

- **WHEN** an embedded terminal output chunk contains `ESC ] 777 ; tui-helper ; command-start ; bnBtIHRlc3Q= BEL`
- **THEN** the visible terminal output excludes the OSC payload
- **AND** the persisted terminal history excludes the OSC payload

#### Scenario: Windows ConPTY textual command-state payload is not rendered or persisted

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;tui-helper;command-start;Y29kZXg=` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload

#### Scenario: Split command-state payload is consumed across output chunks

- **WHEN** one terminal output read ends with `ESC ] 777 ; tui-helper ; command-start ;`
- **AND** the following read contains `Y2xhdWRl BEL`
- **THEN** the system consumes the complete command-state payload
- **AND** neither output chunk renders or persists the partial payload

#### Scenario: Non-application terminal output is preserved

- **WHEN** terminal output contains ordinary command output that does not match the application-private `tui-helper` command-state protocol
- **THEN** the system renders that output normally
- **AND** the system persists that output in terminal history normally
