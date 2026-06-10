## ADDED Requirements

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
