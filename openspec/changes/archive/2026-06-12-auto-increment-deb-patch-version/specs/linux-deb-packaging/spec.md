## ADDED Requirements

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
