## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Classify Terminal Activity From Titles

The system SHALL classify terminal activity from captured terminal title changes without treating ordinary Windows path separators, stable program titles, or single-frame Claude title markers as busy indicators. The system SHALL continue to surface explicit busy, needs-input, and recognized animated busy title signals.

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

- **WHEN** terminal A has an established idle title
- **AND** terminal A receives a title update `codex thinking`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Spinner title signal marks terminal busy

- **WHEN** terminal A has an established idle title
- **AND** terminal A receives a title update containing an explicit spinner character such as `⠋`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Claude dot animation marks terminal busy

- **WHEN** terminal A is running Claude
- **AND** terminal A has an established idle title
- **AND** terminal A receives title updates showing the Claude dot marker moving through left, middle, and right positions
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Single Claude dot frame remains idle

- **WHEN** terminal A is running Claude without a prior idle title
- **AND** terminal A receives a title update containing only one frame of the Claude dot marker
- **AND** no subsequent left, middle, and right dot animation sequence is observed
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state remains `idle`

#### Scenario: Needs input title signal marks terminal needs input

- **WHEN** terminal A receives a title update `codex !`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `needs-input`

#### Scenario: Stable title establishes idle baseline

- **WHEN** terminal A is running without a prior idle title
- **AND** terminal A receives a stable title update matching its shell, current command label, launch profile label, or stable program title
- **THEN** the system records that title as terminal A's idle title
- **AND** terminal A's activity state remains `idle`

#### Scenario: Returning to idle baseline clears busy state

- **WHEN** terminal A is marked `busy` from title activity
- **AND** terminal A receives a title update matching its idle baseline
- **THEN** terminal A's activity state becomes `idle`

### Requirement: Suppress Internal Command-State Payloads

The system SHALL consume application-private command-state payloads before embedded terminal output is rendered or persisted. The system MUST NOT display or replay base64 command-state payloads as terminal text. When a command-state payload is invalid but recognizable as application-private, the system SHALL drop the payload and preserve the previous terminal command state.

#### Scenario: Raw command-state OSC is not rendered or persisted

- **WHEN** an embedded terminal output chunk contains `ESC ] 777 ; tui-helper ; command-start ; bnBtIHRlc3Q= BEL`
- **THEN** the visible terminal output excludes the OSC payload
- **AND** the persisted terminal history excludes the OSC payload

#### Scenario: Windows ConPTY textual command-state payload is not rendered or persisted

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;tui-helper;command-start;Y29kZXg=` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload

#### Scenario: Invalid Windows textual command-state payload is dropped

- **WHEN** the application runs on Windows
- **AND** ConPTY output surfaces `777;tui-helper;command-start;not-base64` as ordinary terminal text
- **THEN** the visible terminal output excludes that application-private payload
- **AND** the persisted terminal history excludes that application-private payload
- **AND** the terminal command label is not updated from the invalid payload

#### Scenario: Split command-state payload is consumed across output chunks

- **WHEN** one terminal output read ends with `ESC ] 777 ; tui-helper ; command-start ;`
- **AND** the following read contains `Y2xhdWRl BEL`
- **THEN** the system consumes the complete command-state payload
- **AND** neither output chunk renders or persists the partial payload

#### Scenario: Non-application terminal output is preserved

- **WHEN** terminal output contains ordinary command output that does not match the application-private `tui-helper` command-state protocol
- **THEN** the system renders that output normally
- **AND** the system persists that output in terminal history normally
