## MODIFIED Requirements

### Requirement: Create Additional Project Terminal

The system SHALL allow the user to create additional terminal sessions only for an available project within an `in-progress` TODO project context. Each created terminal session SHALL start an independent shell process in the owning project's directory and SHALL belong only to that TODO project context. The system SHALL reject terminal creation for `not-started` TODO project contexts.

#### Scenario: User creates another terminal for an in-progress todo project

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user creates a new terminal under TODO `fix-login` and project `/home/user/work/demo-app`
- **THEN** the system starts a new shell process with working directory `/home/user/work/demo-app`
- **AND** the new terminal is independent from existing terminal sessions in that TODO project context
- **AND** the new terminal is not shown under other TODOs that reference the same project

#### Scenario: Not-started todo project cannot create terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a new terminal under TODO `fix-login` and project `/home/user/work/demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no runtime terminal session is added to that TODO project context

### Requirement: Run Terminal Launch Profile Command

The system SHALL execute the selected terminal launch profile startup parameters inside the newly created shell session for the selected `in-progress` TODO project context. The system SHALL NOT execute launch profile commands for `not-started` TODO project contexts because terminal creation is not allowed there.

#### Scenario: Launch profile submits command to new shell

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex` under TODO `fix-login` and project `demo-app`
- **THEN** the system creates a new shell session in the selected project's directory
- **AND** the system submits `codex` followed by Enter to that new shell session
- **AND** the new terminal belongs to TODO `fix-login`

#### Scenario: Launch profile supports startup parameters

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** the user chooses a launch profile with startup parameters `codex --model gpt-5`
- **THEN** the system submits `codex --model gpt-5` followed by Enter to the new shell session as a single command

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
