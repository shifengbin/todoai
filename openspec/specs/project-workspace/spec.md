# project-workspace Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.
## Requirements
### Requirement: Create Project From Directory

The system SHALL allow the user to create a project by selecting a local directory through a native directory picker. The created project's default display name SHALL be the basename of the selected directory.

#### Scenario: User creates a project from a directory

- **WHEN** the user clicks the new project action and selects `/home/user/work/demo-app`
- **THEN** the project list contains a project named `demo-app` with path `/home/user/work/demo-app`

#### Scenario: User cancels directory selection

- **WHEN** the user opens the directory picker and cancels it
- **THEN** the project list remains unchanged

### Requirement: Persist Opened Projects

The system SHALL persist the opened project list locally and reload it when the application starts.

#### Scenario: Project list is restored after restart

- **WHEN** the user creates projects and then closes and reopens the application
- **THEN** the previously opened projects appear in the left-side project list

### Requirement: Select Active Project

The system SHALL allow the user to select an opened project from the left-side list and SHALL expose the selected project as the active project for the shell area.

#### Scenario: User selects a project

- **WHEN** the user clicks a project in the left-side project list
- **THEN** that project becomes active and the shell area is associated with that project's directory

### Requirement: Handle Duplicate Project Paths

The system SHALL avoid creating duplicate project entries for the same absolute path.

#### Scenario: User selects an already opened directory

- **WHEN** the user creates a project from a directory that is already in the project list
- **THEN** the existing project entry is selected instead of adding a duplicate entry

### Requirement: Handle Missing Project Paths

The system SHALL detect when a persisted project path no longer exists or is inaccessible and SHALL prevent shell startup for that project until the path is valid again.

#### Scenario: Persisted project path is missing

- **WHEN** the application starts and a persisted project path no longer exists
- **THEN** the project remains visible as unavailable and selecting it does not launch a shell

