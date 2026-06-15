# application-identity Specification

## Purpose
TBD - created by archiving change rebrand-app-to-todoai. Update Purpose after archive.
## Requirements
### Requirement: Display TodoAI Application Name

The system SHALL display `TodoAI` as the application name in user-visible desktop surfaces controlled by the app.

#### Scenario: Main window uses TodoAI title

- **WHEN** the desktop application starts
- **THEN** the native application window title is `TodoAI`

#### Scenario: Frontend document uses TodoAI title

- **WHEN** the frontend document is loaded
- **THEN** the document title is `TodoAI`

### Requirement: Publish TodoAI Application Identity

The system SHALL use `todoai` as the application package, binary, and build output identity where a lowercase machine-readable application name is required.

#### Scenario: Wails build uses todoai identity

- **WHEN** the Wails application is built
- **THEN** the configured project name is `todoai`
- **AND** the configured output filename is `todoai`

### Requirement: Provide TodoAI Launcher Icon Assets

The system SHALL provide static launcher icon assets for the TodoAI application.

#### Scenario: Common launcher icon exists

- **WHEN** application build assets are inspected
- **THEN** `build/appicon.png` contains the TodoAI launcher icon

#### Scenario: Windows launcher icon exists

- **WHEN** Windows application build assets are inspected
- **THEN** `build/windows/icon.ico` contains the TodoAI launcher icon

### Requirement: Migrate Local Application Data Directory

The system SHALL use `todoai` as the default local application configuration directory and SHALL preserve existing data from the legacy `tui-helper` directory during upgrade.

#### Scenario: New install uses todoai config directory

- **WHEN** no legacy `tui-helper` configuration directory exists
- **AND** the application resolves its default project config path
- **THEN** the path is under the user config directory's `todoai` child directory

#### Scenario: Legacy config migrates to todoai directory

- **WHEN** the legacy `tui-helper` configuration directory exists
- **AND** the new `todoai` configuration directory does not exist
- **AND** the application resolves its default project config path
- **THEN** the application makes the legacy project data, settings, and terminal history available under the new `todoai` configuration directory

#### Scenario: Existing todoai config is not overwritten

- **WHEN** both the legacy `tui-helper` configuration directory and the new `todoai` configuration directory exist
- **AND** the application resolves its default project config path
- **THEN** the application uses the new `todoai` configuration directory
- **AND** the application does not overwrite files in the new `todoai` configuration directory with legacy files

