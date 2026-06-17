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
The system SHALL use `todoai` as the default local application configuration directory and SHALL preserve existing data from the legacy `tui-helper` directory during upgrade. When migrating to workspace-scoped storage, the system SHALL make legacy global project data and terminal history available through an app-managed legacy workspace instead of continuing to treat global `projects.json` as the active workspace data source. The system SHALL keep terminal settings as application-global data in the app config directory.

#### Scenario: New install uses todoai config directory
- **WHEN** no legacy `tui-helper` configuration directory exists
- **AND** the application resolves its default application config directory
- **THEN** the path is under the user config directory's `todoai` child directory

#### Scenario: Legacy config migrates to todoai directory
- **WHEN** the legacy `tui-helper` configuration directory exists
- **AND** the new `todoai` configuration directory does not exist
- **AND** the application resolves its default application config directory
- **THEN** the application makes the legacy project data, settings, and terminal history available under the new `todoai` configuration directory

#### Scenario: Existing todoai config is not overwritten
- **WHEN** both the legacy `tui-helper` configuration directory and the new `todoai` configuration directory exist
- **AND** the application resolves its default application config directory
- **THEN** the application uses the new `todoai` configuration directory
- **AND** the application does not overwrite files in the new `todoai` configuration directory with legacy files

#### Scenario: Legacy global data becomes recent workspace
- **WHEN** legacy global `projects.json` or `terminal-history.json` exists in the application config directory
- **AND** workspace-scoped migration has not already been completed
- **THEN** the system copies the legacy workspace files into an app-managed legacy workspace `.data` directory
- **AND** the app-managed legacy workspace appears in the recent workspace list
- **AND** the original legacy files remain on disk

#### Scenario: Legacy global settings remain global
- **WHEN** legacy global `settings.json` exists in the application config directory
- **AND** workspace-scoped migration runs
- **THEN** the system keeps `settings.json` in the application config directory
- **AND** the system does not copy `settings.json` into the app-managed legacy workspace `.data` directory

#### Scenario: Migrated legacy workspace is not overwritten
- **WHEN** the app-managed legacy workspace already exists
- **AND** the application starts again
- **THEN** the system does not overwrite the existing legacy workspace `.data` files
- **AND** the recent workspace list still contains the app-managed legacy workspace

