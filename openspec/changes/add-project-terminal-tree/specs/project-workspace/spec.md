## ADDED Requirements

### Requirement: Display Project Terminal Tree
The system SHALL display opened projects as top-level rows in the left sidebar and SHALL display each project's terminal sessions as child rows under the owning project.

#### Scenario: Project has multiple terminals
- **WHEN** a project has terminal sessions named `zsh`, `npm run dev`, and `go test ./...`
- **THEN** the left sidebar shows the project row with those terminal rows nested beneath it

### Requirement: Select Active Terminal From Project Tree
The system SHALL allow the user to select a terminal row under a project and SHALL expose that terminal as the active terminal for the shell area.

#### Scenario: User selects a terminal under a project
- **WHEN** the user clicks terminal `go test ./...` under project `demo-app`
- **THEN** project `demo-app` becomes the active project
- **AND** terminal `go test ./...` becomes the active terminal shown in the shell area

## MODIFIED Requirements

### Requirement: Select Active Project

The system SHALL allow the user to select an opened project from the left-side project tree and SHALL expose the selected project as the active project for the shell area. If the selected project has an existing terminal session, the system SHALL make that project's most recently active terminal active. If the selected project has no terminal session and the project path is available, the system SHALL create and select a default terminal for that project.

#### Scenario: User selects a project

- **WHEN** the user clicks a project in the left-side project tree
- **THEN** that project becomes active
- **AND** the shell area is associated with an active terminal in that project's directory
