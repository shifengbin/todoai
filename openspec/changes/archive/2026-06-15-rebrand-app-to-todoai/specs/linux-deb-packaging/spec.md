## MODIFIED Requirements

### Requirement: Include Package Metadata

The `.deb` package SHALL include application metadata needed for installation and desktop launching. The generated package metadata SHALL use `todoai` as the package, binary, desktop file, and icon identity, and SHALL use `TodoAI` as the desktop launcher display name.

#### Scenario: Package metadata is present

- **WHEN** the `.deb` package is built
- **THEN** it includes package name, version, architecture, description, maintainer metadata, and desktop launcher metadata

#### Scenario: TodoAI package metadata is present

- **WHEN** the `.deb` package is built
- **THEN** the Debian package name is `todoai`
- **AND** the installed executable path is `/usr/bin/todoai`
- **AND** the desktop launcher file is named `todoai.desktop`
- **AND** the desktop launcher display name is `TodoAI`
- **AND** the launcher icon name is `todoai`

### Requirement: Installed Application Launches

The installed package SHALL provide a launchable desktop application that starts the Wails app through the `todoai` executable.

#### Scenario: User launches installed application

- **WHEN** the user installs the `.deb` package and launches the application from the desktop environment
- **THEN** the Wails desktop application starts and displays the project shell UI

#### Scenario: User launches TodoAI executable

- **WHEN** the user installs the `.deb` package
- **AND** the user runs `/usr/bin/todoai`
- **THEN** the Wails desktop application starts
