# terminal-settings Specification

## Purpose
Defines local terminal shell selection settings used when creating embedded shell sessions.
## Requirements
### Requirement: Persist Terminal Shell Setting

The system SHALL persist the embedded terminal shell setting locally and SHALL reload it when the application starts.

#### Scenario: First startup detects and persists shell

- **WHEN** the application loads terminal settings and no saved terminal shell setting exists
- **THEN** the system detects an available shell
- **AND** the system saves the detected shell path as the terminal shell setting
- **AND** the settings state exposes the saved shell path and display name

#### Scenario: Saved shell setting is restored

- **WHEN** the user has previously saved `/usr/bin/zsh` as the terminal shell setting
- **AND** the application starts again
- **THEN** the settings state exposes `/usr/bin/zsh` as the selected terminal shell path
- **AND** automatic detection is not used to replace the saved path

### Requirement: Change Terminal Shell Setting

The system SHALL allow the user to change the embedded terminal shell setting from the settings interface.

#### Scenario: User saves detected shell option

- **WHEN** the settings interface shows `/usr/bin/bash` as an available shell option
- **AND** the user selects `/usr/bin/bash` and saves
- **THEN** the terminal shell setting is persisted as `/usr/bin/bash`
- **AND** newly created embedded terminals use `/usr/bin/bash`

#### Scenario: User saves manual shell path

- **WHEN** the user enters `/opt/custom/bin/fish` as a manual shell path
- **AND** the path exists and is executable
- **AND** the user saves
- **THEN** the terminal shell setting is persisted as `/opt/custom/bin/fish`
- **AND** the settings state exposes `fish` as the display name

#### Scenario: User enters invalid manual shell path

- **WHEN** the user enters `/missing/shell` as a manual shell path
- **AND** the path does not exist or is not executable
- **THEN** the system rejects the setting
- **AND** the previously saved terminal shell setting remains unchanged

### Requirement: Re-detect Terminal Shell Setting

The system SHALL allow the user to re-run automatic terminal shell detection from the settings interface.

#### Scenario: User re-runs shell detection

- **WHEN** the user triggers shell detection from settings
- **THEN** the system detects an available shell
- **AND** the settings interface shows the detected shell as the current candidate

#### Scenario: User saves re-detected shell

- **WHEN** shell detection returns `/usr/bin/zsh`
- **AND** the user saves the detected result
- **THEN** the terminal shell setting is persisted as `/usr/bin/zsh`

### Requirement: Report Unavailable Saved Shell

The system SHALL report when the saved terminal shell path is unavailable and SHALL provide a detected fallback for continued terminal startup.

#### Scenario: Saved shell path is unavailable

- **WHEN** the saved terminal shell setting is `/old/bin/zsh`
- **AND** `/old/bin/zsh` does not exist or is not executable
- **AND** the application loads terminal settings
- **THEN** the settings state marks the saved shell as unavailable
- **AND** the settings state includes an automatically detected fallback shell

#### Scenario: Unavailable saved shell does not prevent terminal startup

- **WHEN** the saved terminal shell setting is unavailable
- **AND** an automatically detected fallback shell exists
- **AND** the user creates a new embedded terminal
- **THEN** the new embedded terminal starts with the fallback shell
- **AND** the saved terminal shell setting remains unchanged until the user saves a new setting

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

### Requirement: Persist Appearance Theme Setting
The system SHALL persist the application appearance theme in terminal settings and SHALL expose the theme when terminal settings are loaded.

#### Scenario: Missing theme setting uses default
- **WHEN** the application loads terminal settings from an existing settings file that has no theme field
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system preserves existing terminal shell and launch profile settings

#### Scenario: Saved theme setting is restored
- **WHEN** the user has previously saved `dark` as the appearance theme
- **AND** the application loads terminal settings
- **THEN** the settings state exposes `dark` as the appearance theme

#### Scenario: Invalid saved theme is normalized
- **WHEN** the application loads terminal settings from a settings file with an unsupported theme value
- **THEN** the settings state exposes `light` as the appearance theme
- **AND** the system does not reject the settings file

### Requirement: Change Appearance Theme Setting
The system SHALL allow the user to save the application appearance theme from the settings interface.

#### Scenario: User saves valid appearance theme
- **WHEN** the user selects `dark` as the appearance theme
- **AND** the user saves settings
- **THEN** the appearance theme is persisted as `dark`
- **AND** the settings state exposes `dark` as the appearance theme

#### Scenario: User saves unsupported appearance theme
- **WHEN** the application receives an unsupported appearance theme value
- **THEN** the system rejects the setting
- **AND** the previously saved appearance theme remains unchanged

