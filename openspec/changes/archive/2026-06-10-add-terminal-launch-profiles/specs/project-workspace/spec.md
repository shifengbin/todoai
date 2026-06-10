## ADDED Requirements

### Requirement: Show Project Terminal Launch Menu

The system SHALL show a terminal launch menu when the user activates the add-terminal control for an available project.

#### Scenario: Launch menu contains terminal and configured profiles

- **WHEN** settings contains launch profiles named `codex` and `claude`
- **AND** the user activates the add-terminal control for an available project
- **THEN** the launch menu shows `Terminal` as the first option
- **AND** the launch menu shows `codex` and `claude` after `Terminal` in the configured order

#### Scenario: Unavailable project has no launch menu

- **WHEN** a project path is unavailable
- **THEN** the project row does not expose an add-terminal launch menu action

### Requirement: Create Terminal From Launch Menu

The system SHALL create a new terminal for the selected project using the launch option chosen from the project terminal launch menu.

#### Scenario: User chooses terminal launch option

- **WHEN** the user opens the launch menu for project `demo-app`
- **AND** chooses `Terminal`
- **THEN** the system creates a new terminal under project `demo-app`
- **AND** the new terminal starts as a normal shell session without an automatic startup command

#### Scenario: User chooses configured launch profile

- **WHEN** the user opens the launch menu for project `demo-app`
- **AND** chooses the `codex` launch profile
- **THEN** the system creates a new terminal under project `demo-app`
- **AND** the new terminal is selected as the active terminal
- **AND** the owning project branch is expanded so the new terminal row is visible
