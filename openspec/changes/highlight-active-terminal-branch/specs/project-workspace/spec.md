## ADDED Requirements

### Requirement: Highlight Active Terminal Branch Guide
The system SHALL visually highlight the visible project-terminal branch guide for the project that owns the active terminal.

#### Scenario: Active terminal branch guide is highlighted
- **WHEN** a project has visible terminal child rows
- **AND** one of those terminal rows is the active terminal
- **THEN** the visible vertical branch guide for that project's terminal list uses the same active color as the active terminal row's horizontal branch guide

#### Scenario: Inactive terminal branch guides remain neutral
- **WHEN** a project has visible terminal child rows
- **AND** none of those terminal rows is the active terminal
- **THEN** the visible vertical branch guide for that project's terminal list uses the default neutral branch guide color
