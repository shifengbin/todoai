# linux-deb-packaging Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.
## Requirements
### Requirement: Build Debian Package

The system SHALL provide a build path that produces an installable Linux `.deb` package for the desktop application.

#### Scenario: Build creates deb artifact

- **WHEN** the Linux packaging build command completes successfully
- **THEN** a `.deb` package artifact exists in the configured build output directory

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

### Requirement: Auto-Increment Debian Package Patch Version

The system SHALL automatically advance the Debian package patch version for the default Linux packaging command.

#### Scenario: Default packaging increments patch version

- **WHEN** the Debian packaging command runs without an explicit version override and the persisted package version is `0.1.8`
- **THEN** the generated package metadata uses version `0.1.9`
- **AND** the generated `.deb` artifact filename includes `0.1.9`

#### Scenario: Successful packaging persists generated version

- **WHEN** the Debian packaging command completes successfully with generated version `0.1.9`
- **THEN** the persisted package version is updated to `0.1.9`

#### Scenario: Failed packaging preserves previous version

- **WHEN** the Debian packaging command fails before completing the `.deb` artifact
- **THEN** the persisted package version remains unchanged

#### Scenario: Explicit version override is supported

- **WHEN** the Debian packaging command runs with an explicit version override of `0.2.0`
- **THEN** the generated package metadata uses version `0.2.0`
- **AND** the generated `.deb` artifact filename includes `0.2.0`
- **AND** the persisted package version is updated to `0.2.0` after successful packaging

