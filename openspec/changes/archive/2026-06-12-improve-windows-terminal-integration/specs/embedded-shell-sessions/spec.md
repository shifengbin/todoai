## MODIFIED Requirements

### Requirement: Run Terminal Launch Profile Command

The system SHALL execute the selected terminal launch profile startup parameters inside the newly created shell session for the selected `in-progress` TODO project context, and SHALL submit the command using the platform-correct interactive Enter sequence. The system SHALL NOT execute launch profile commands for `not-started` TODO project contexts because terminal creation is not allowed there.

#### Scenario: Launch profile submits command to new shell

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system submits `codex` followed by Enter to that new shell session
- **AND** on Windows ConPTY-backed shells the command starts without requiring the user to press Enter again
- **AND** the new terminal belongs to TODO `fix-login`

#### Scenario: Launch profile supports startup parameters

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex --model gpt-5`
- **THEN** the system submits `codex --model gpt-5` followed by Enter to the new shell session as a single command
- **AND** on Windows ConPTY-backed shells the command starts without requiring the user to press Enter again

#### Scenario: Plain terminal launch does not submit command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses the built-in `Terminal` launch option under a TODO project context
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system does not submit any automatic command to that shell session

#### Scenario: Not-started todo launch profile is not executed

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** the system does not submit `codex` to any shell session

## ADDED Requirements

### Requirement: Reflect Launch Profile Command Label

The system SHALL update the newly created terminal's command label when it submits a launch profile command, even before a shell-specific command-start event is received.

#### Scenario: Windows launch profile displays command label immediately

- **WHEN** the application runs on Windows
- **AND** TODO `fix-login` has status `in-progress`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the profile command to the new shell session
- **AND** the TODO terminal tree displays the new terminal label as `codex`
- **AND** the terminal label does not remain `pwsh`, `powershell`, or `cmd`

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

The system SHALL provide command state events for Windows `pwsh` and `powershell` embedded shell sessions using the same OSC 777 command-start and command-end protocol used by existing Unix shell integrations.

#### Scenario: PowerShell command start updates terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `pwsh.exe` or `powershell.exe`
- **AND** the user runs `npm test` in the embedded terminal
- **THEN** the shell integration emits a command-start event for `npm test`
- **AND** the corresponding terminal command label becomes `npm test`

#### Scenario: PowerShell command completion clears terminal label

- **WHEN** the application runs on Windows with ConPTY support
- **AND** a PowerShell-backed embedded terminal is running command `npm test`
- **AND** the command completes and the shell returns to its prompt
- **THEN** the shell integration emits a command-end event for that terminal
- **AND** the terminal command label is cleared

#### Scenario: Cmd fallback remains usable without lifecycle hook

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured terminal shell is `cmd.exe`
- **AND** the user chooses launch profile `codex` with startup parameters `codex`
- **THEN** the system submits the launch profile command to the cmd-backed shell session
- **AND** the system keeps the application-provided launch profile command label when no shell lifecycle hook event is available
- **AND** the shell session remains associated with `cmd.exe`
