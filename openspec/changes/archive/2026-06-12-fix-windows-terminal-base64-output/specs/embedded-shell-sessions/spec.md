## MODIFIED Requirements

### Requirement: Emit Command State For Windows PowerShell Sessions

The system SHALL provide command state events for Windows `pwsh` and `powershell` embedded shell sessions using an application-private command-state protocol that is consumed before terminal rendering and history persistence. The system MUST NOT expose the command-state protocol payload as visible terminal output. When a valid command-state event cannot be recovered, the system SHALL safely ignore the event and preserve existing launch profile command-label fallback behavior.

#### Scenario: PowerShell command start updates terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `pwsh.exe` or `powershell.exe`
- **AND** the user runs `npm test` in the embedded terminal
- **THEN** the shell integration emits a command-start event for `npm test`
- **AND** the corresponding terminal command label becomes `npm test`
- **AND** the command-state payload is not displayed in the terminal

#### Scenario: PowerShell command completion clears terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal is running command `npm test`
- **AND** the command completes and the shell returns to its prompt
- **THEN** the shell integration emits a command-end event for that terminal
- **AND** the terminal command label is cleared
- **AND** the command-state payload is not displayed in the terminal

#### Scenario: Invalid PowerShell command-state payload is safely ignored

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal produces an application-private command-start payload with invalid base64 command text
- **THEN** the system does not update the terminal command label from that invalid payload
- **AND** the invalid command-state payload is not displayed in the terminal
- **AND** the invalid command-state payload is not persisted in terminal history

#### Scenario: Cmd fallback remains usable without lifecycle hook

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `cmd.exe`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the launch profile command to the cmd-backed shell session
- **AND** the system keeps the application-provided launch profile command label when no shell lifecycle hook event is available
- **AND** the shell session remains associated with `cmd.exe`
