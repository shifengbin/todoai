## MODIFIED Requirements

### Requirement: Reflect Launch Profile Command Label

The system SHALL update the newly created terminal's command label when it submits a launch profile command, even before a shell-specific command-start event is received. The launch profile command label by itself MUST NOT mark the terminal as busy. Submitting a non-empty launch profile command MUST NOT render supported application-private command-state payloads in the terminal, and MUST NOT hide unrelated base64-like launch output by heuristic. Terminal activity state SHALL be derived from shell status, command-state events, and terminal title activity classification.

#### Scenario: Windows launch profile displays command label immediately

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `codex`
- **AND** the terminal label does not remain `pwsh`, `powershell`, or `cmd`

#### Scenario: Windows Claude launch profile displays command label without forcing busy

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `claude` with startup parameters `claude --dangerously-skip-permissions`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `claude --dangerously-skip-permissions`
- **AND** the terminal activity state remains `idle` until a recognized busy, needs-input, or command-state signal is received

#### Scenario: Windows arbitrary launch profile preserves unrelated startup text

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `Calculator` with startup parameters `calc`
- **THEN** the system submits the profile command to the new shell session
- **AND** the terminal displays normal shell or program output
- **AND** the terminal does not display supported application-private command-state payloads created by launch profile submission
- **AND** the terminal preserves unrelated base64-like launch output instead of hiding it by heuristic

#### Scenario: Launch profile with parameters displays submitted command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `Codex GPT-5` with startup parameters `codex --model gpt-5`
- **THEN** the TODO terminal tree displays the new terminal label as `codex --model gpt-5`
- **AND** the displayed label is sanitized using the normal terminal command label rules

#### Scenario: Shell command end clears command label when available

- **WHEN** a terminal has command label `codex`
- **AND** the shell integration emits a command-end event for that terminal
- **THEN** the system clears the terminal command label
- **AND** the terminal falls back to its shell display name while still running

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

#### Scenario: PowerShell command-state payload is hidden during launch profile startup

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the user launches any non-empty launch profile in a PowerShell-backed embedded terminal
- **AND** PowerShell emits an application-private command-state payload for the submitted command
- **THEN** the command-state payload is not displayed in the terminal
- **AND** the command-state payload is not persisted in terminal history
- **AND** the launch profile command label remains visible unless a valid command-state event replaces or clears it

#### Scenario: Invalid PowerShell command-state payload is safely ignored

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal produces an application-private command-start payload with invalid base64 command text
- **THEN** the system does not update the terminal command label from that invalid payload
- **AND** the invalid command-state payload is not displayed in the terminal
- **AND** the invalid command-state payload is not persisted in terminal history
