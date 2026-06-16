## MODIFIED Requirements

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
