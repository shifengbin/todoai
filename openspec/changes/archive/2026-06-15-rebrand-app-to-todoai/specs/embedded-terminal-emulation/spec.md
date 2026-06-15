## MODIFIED Requirements

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
