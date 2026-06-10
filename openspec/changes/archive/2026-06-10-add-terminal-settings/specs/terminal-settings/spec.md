## ADDED Requirements

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
