## MODIFIED Requirements

### Requirement: Persist Terminal Shell Setting
The system SHALL persist the embedded terminal shell setting in the application-global settings file and SHALL reload it regardless of the current workspace. Terminal shell settings SHALL be shared across workspaces.

#### Scenario: First settings load detects and persists shell
- **WHEN** the application loads terminal settings
- **AND** no saved terminal shell setting exists
- **THEN** the system detects an available shell
- **AND** the system saves the detected shell path as the terminal shell setting
- **AND** the settings state exposes the saved shell path and display name

#### Scenario: Saved global shell setting is restored
- **WHEN** the user has previously saved `/usr/bin/zsh` as the terminal shell setting
- **AND** the application loads terminal settings
- **THEN** the settings state exposes `/usr/bin/zsh` as the selected terminal shell path
- **AND** automatic detection is not used to replace the saved path

#### Scenario: Shell setting is shared across workspaces
- **WHEN** workspace `/work/customer-a` is open
- **AND** the user saves `/usr/bin/zsh` as the terminal shell setting
- **AND** the user opens workspace `/work/customer-b`
- **THEN** the settings state still exposes `/usr/bin/zsh`

#### Scenario: Shell setting is available without workspace
- **WHEN** no workspace is open
- **THEN** the user can load and save the terminal shell setting

### Requirement: Persist Terminal Launch Profiles
The system SHALL persist configurable terminal launch profiles in the application-global terminal settings and SHALL expose their name, startup parameters, enabled state, and saved order regardless of the current workspace. The built-in `Terminal` launch option SHALL NOT be persisted as a configurable profile. Terminal launch profiles SHALL be shared across workspaces.

#### Scenario: Missing launch profiles use defaults
- **WHEN** the application loads terminal settings from the global settings file that has no launch profiles field
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
- **WHEN** the application loads terminal settings from the global settings file whose launch profiles do not include an enabled state
- **THEN** each launch profile is exposed as enabled
- **AND** the existing launch profile names, startup parameters, and order remain unchanged

#### Scenario: Empty launch profile list remains empty
- **WHEN** the user has previously saved an empty launch profile list
- **AND** the application loads terminal settings
- **THEN** the settings state exposes no custom launch profiles
- **AND** the built-in `Terminal` launch option remains available outside the configurable profile list

#### Scenario: Launch profiles are shared across workspaces
- **WHEN** workspace `/work/customer-a` is open
- **AND** the user saves launch profile `Customer Codex`
- **AND** the user opens workspace `/work/customer-b`
- **THEN** terminal launch menus include `Customer Codex`

### Requirement: Persist Appearance Theme Setting
The system SHALL persist the application appearance theme in the application-global terminal settings and SHALL expose the theme regardless of the current workspace. Appearance theme settings SHALL be shared across workspaces.

#### Scenario: Missing theme setting uses default
- **WHEN** the application loads terminal settings from the global settings file that has no theme field
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system preserves existing terminal shell and launch profile settings

#### Scenario: Saved theme setting is restored
- **WHEN** the user has previously saved `dark` as the appearance theme
- **AND** the application loads terminal settings
- **THEN** the settings state exposes `dark` as the appearance theme

#### Scenario: Invalid saved theme is normalized
- **WHEN** the application loads terminal settings from the global settings file with an unsupported theme value
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system does not reject the settings file

#### Scenario: Appearance theme is shared across workspaces
- **WHEN** workspace `/work/customer-a` is open
- **AND** the user saves `dark` as the appearance theme
- **AND** the user opens workspace `/work/customer-b`
- **THEN** the application appearance theme is still `dark`
