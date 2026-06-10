## ADDED Requirements

### Requirement: Run Terminal Launch Profile Command

The system SHALL execute the selected terminal launch profile startup parameters inside the newly created shell session.

#### Scenario: Launch profile submits command to new shell

- **WHEN** the user chooses a launch profile with startup parameters `codex`
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system submits `codex` followed by Enter to that new shell session

#### Scenario: Launch profile supports startup parameters

- **WHEN** the user chooses a launch profile with startup parameters `codex --model gpt-5`
- **THEN** the system submits `codex --model gpt-5` followed by Enter to the new shell session as a single command

#### Scenario: Plain terminal launch does not submit command

- **WHEN** the user chooses the built-in `Terminal` launch option
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system does not submit any automatic command to that shell session

### Requirement: Keep Launch Profile Commands In Configured Shell

The system SHALL run launch profile startup parameters inside the configured terminal shell instead of replacing the shell process with the startup command.

#### Scenario: Launch profile command exits

- **WHEN** a terminal launch profile command exits after running in a new terminal
- **THEN** the terminal remains associated with its configured shell session unless the shell itself exits
