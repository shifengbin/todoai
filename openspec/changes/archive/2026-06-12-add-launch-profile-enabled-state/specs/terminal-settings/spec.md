## MODIFIED Requirements

### Requirement: Persist Terminal Launch Profiles

The system SHALL persist configurable terminal launch profiles in terminal settings and SHALL expose their name, startup parameters, enabled state, and saved order when settings are loaded. The built-in `Terminal` launch option SHALL NOT be persisted as a configurable profile.

#### Scenario: Missing launch profiles use defaults

- **WHEN** the application loads terminal settings from an existing settings file that has no launch profiles field
- **THEN** the settings state includes launch profiles named `codex` and `claude`
- **AND** the `codex` profile has startup parameters `codex`
- **AND** the `claude` profile has startup parameters `claude`
- **AND** both default launch profiles are enabled

#### Scenario: Saved launch profiles are restored

- **WHEN** the user has previously saved launch profiles named `Codex GPT-5` and `Claude Plan`
- **AND** the application loads terminal settings
- **THEN** the settings state exposes those launch profile names in the saved order
- **AND** each launch profile exposes its saved startup parameters
- **AND** each launch profile exposes its saved enabled state

#### Scenario: Legacy launch profiles without enabled state remain enabled

- **WHEN** the application loads terminal settings from an existing settings file whose launch profiles do not include an enabled state
- **THEN** each launch profile is exposed as enabled
- **AND** the existing launch profile names, startup parameters, and order remain unchanged

#### Scenario: Empty launch profile list remains empty

- **WHEN** the user has previously saved an empty launch profile list
- **AND** the application loads terminal settings
- **THEN** the settings state exposes no custom launch profiles
- **AND** the built-in `Terminal` launch option remains available outside the configurable profile list

### Requirement: Change Terminal Launch Profiles

The system SHALL allow the user to add, edit, reorder, enable, disable, and remove configurable terminal launch profiles from the settings interface.

#### Scenario: User saves valid launch profiles

- **WHEN** the user configures a launch profile named `Codex` with startup parameters `codex --model gpt-5`
- **AND** the user saves settings
- **THEN** the launch profile is persisted with name `Codex`
- **AND** the launch profile is persisted with startup parameters `codex --model gpt-5`
- **AND** the launch profile is persisted as enabled

#### Scenario: User disables a launch profile

- **WHEN** settings contains an enabled launch profile named `Claude Plan`
- **AND** the user disables `Claude Plan` and saves settings
- **THEN** the launch profile remains persisted with its name and startup parameters
- **AND** the launch profile is persisted as disabled

#### Scenario: User enables a disabled launch profile

- **WHEN** settings contains a disabled launch profile named `Claude Plan`
- **AND** the user enables `Claude Plan` and saves settings
- **THEN** the launch profile remains persisted with its name and startup parameters
- **AND** the launch profile is persisted as enabled

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

## ADDED Requirements

### Requirement: Display Enabled Terminal Launch Profiles

The system SHALL include only enabled configurable terminal launch profiles in terminal launch menus. The built-in `Terminal` launch option SHALL remain available regardless of configurable launch profile states.

#### Scenario: Disabled launch profile is hidden from launch menu

- **WHEN** terminal settings include an enabled launch profile named `codex`
- **AND** terminal settings include a disabled launch profile named `claude`
- **AND** the user opens a terminal launch menu
- **THEN** the launch menu includes `Terminal`
- **AND** the launch menu includes `codex`
- **AND** the launch menu does not include `claude`

#### Scenario: Launch menu works when all custom profiles are disabled

- **WHEN** all configurable launch profiles are disabled
- **AND** the user opens a terminal launch menu
- **THEN** the launch menu includes `Terminal`
- **AND** the launch menu does not include any custom launch profile

#### Scenario: Enabled launch profile can start with its command

- **WHEN** terminal settings include an enabled launch profile named `Codex GPT-5` with startup parameters `codex --model gpt-5`
- **AND** the user selects `Codex GPT-5` from the launch menu
- **THEN** the system creates a terminal
- **AND** the system submits `codex --model gpt-5` to the created terminal
