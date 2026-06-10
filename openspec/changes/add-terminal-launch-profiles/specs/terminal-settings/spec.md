## ADDED Requirements

### Requirement: Persist Terminal Launch Profiles

The system SHALL persist configurable terminal launch profiles in terminal settings and SHALL expose them when settings are loaded. The built-in `Terminal` launch option SHALL NOT be persisted as a configurable profile.

#### Scenario: Missing launch profiles use defaults

- **WHEN** the application loads terminal settings from an existing settings file that has no launch profiles field
- **THEN** the settings state includes launch profiles named `codex` and `claude`
- **AND** the `codex` profile has startup parameters `codex`
- **AND** the `claude` profile has startup parameters `claude`

#### Scenario: Saved launch profiles are restored

- **WHEN** the user has previously saved launch profiles named `Codex GPT-5` and `Claude Plan`
- **AND** the application loads terminal settings
- **THEN** the settings state exposes those launch profile names in the saved order
- **AND** each launch profile exposes its saved startup parameters

#### Scenario: Empty launch profile list remains empty

- **WHEN** the user has previously saved an empty launch profile list
- **AND** the application loads terminal settings
- **THEN** the settings state exposes no custom launch profiles
- **AND** the built-in `Terminal` launch option remains available outside the configurable profile list

### Requirement: Change Terminal Launch Profiles

The system SHALL allow the user to add, edit, reorder, and remove configurable terminal launch profiles from the settings interface.

#### Scenario: User saves valid launch profiles

- **WHEN** the user configures a launch profile named `Codex` with startup parameters `codex --model gpt-5`
- **AND** the user saves settings
- **THEN** the launch profile is persisted with name `Codex`
- **AND** the launch profile is persisted with startup parameters `codex --model gpt-5`

#### Scenario: User removes a launch profile

- **WHEN** settings contains launch profiles named `codex` and `claude`
- **AND** the user removes the `claude` profile and saves settings
- **THEN** the settings state includes the `codex` launch profile
- **AND** the settings state does not include the `claude` launch profile

#### Scenario: Invalid launch profile is rejected

- **WHEN** the user configures a launch profile with an empty name or empty startup parameters
- **AND** the user saves settings
- **THEN** the system rejects the setting
- **AND** the previously saved launch profiles remain unchanged

#### Scenario: Launch profile name conflicts with built-in terminal

- **WHEN** the user configures a custom launch profile named `Terminal`
- **AND** the user saves settings
- **THEN** the system rejects the setting
- **AND** the built-in `Terminal` launch option remains unchanged
