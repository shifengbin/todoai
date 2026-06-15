## MODIFIED Requirements

### Requirement: Reflect Launch Profile Command Label

The system SHALL update the newly created terminal's command label when it submits a launch profile command, even before a shell-specific command-start event is received. The launch profile command label by itself MUST NOT mark the terminal agent activity phase as `busy`; agent activity SHALL be derived by the unified agent status system from shell lifecycle, command-state, structured Claude/Codex events, and title fallback according to source priority.

#### Scenario: Windows launch profile displays command label immediately

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `codex`
- **AND** the terminal label does not remain `pwsh`, `powershell`, or `cmd`
- **AND** the terminal agent activity phase remains `idle` until a recognized activity signal is received

#### Scenario: Launch profile with parameters displays submitted command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `Codex GPT-5` with startup parameters `codex --model gpt-5`
- **THEN** the TODO terminal tree displays the new terminal label as `codex --model gpt-5`
- **AND** the displayed label is sanitized using the normal terminal command label rules
- **AND** the submitted command label is not by itself treated as an agent busy signal

#### Scenario: Shell command end clears command label when available

- **WHEN** a terminal has command label `codex`
- **AND** the shell integration emits a command-end event for that terminal
- **THEN** the system clears the terminal command label
- **AND** the terminal falls back to its shell display name while still running
- **AND** the unified agent status for that terminal is reset to `idle` unless a newer structured agent event keeps a higher-priority phase

### Requirement: Emit Command State For Windows PowerShell Sessions

The system SHALL provide command state events for Windows `pwsh` and `powershell` embedded shell sessions using an application-private command-state protocol that is consumed before terminal rendering and history persistence. The system MUST NOT expose the command-state protocol payload as visible terminal output. When a valid command-state event cannot be recovered, the system SHALL safely ignore the event and preserve existing launch profile command-label fallback behavior. Valid command-state events SHALL update the command label and SHALL feed the unified agent status system as shell-command lifecycle signals.

#### Scenario: PowerShell command start updates terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `pwsh.exe` or `powershell.exe`
- **AND** the user runs `npm test` in the embedded terminal
- **THEN** the shell integration emits a command-start event for `npm test`
- **AND** the corresponding terminal command label becomes `npm test`
- **AND** the command-state payload is not displayed in the terminal
- **AND** the command-start signal is available to the unified agent status reducer

#### Scenario: PowerShell command completion clears terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal is running command `npm test`
- **AND** the command completes and the shell returns to its prompt
- **THEN** the shell integration emits a command-end event for that terminal
- **AND** the terminal command label is cleared
- **AND** the command-state payload is not displayed in the terminal
- **AND** the command-end signal is available to reset agent activity when no newer structured agent status applies

#### Scenario: Invalid PowerShell command-state payload is safely ignored

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal produces an application-private command-start payload with invalid base64 command text
- **THEN** the system does not update the terminal command label from that invalid payload
- **AND** the invalid command-state payload is not displayed in the terminal
- **AND** the invalid command-state payload is not persisted in terminal history
- **AND** the invalid command-state payload does not change the unified agent activity phase

#### Scenario: Cmd fallback remains usable without lifecycle hook

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `cmd.exe`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the launch profile command to the cmd-backed shell session
- **AND** the system keeps the application-provided launch profile command label when no shell lifecycle hook event is available
- **AND** the shell session remains associated with `cmd.exe`
- **AND** agent activity falls back to structured Claude/Codex events or terminal title fallback when shell command-state is unavailable
