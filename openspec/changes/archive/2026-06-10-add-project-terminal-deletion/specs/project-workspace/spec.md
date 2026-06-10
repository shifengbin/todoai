## ADDED Requirements

### Requirement: Remove Opened Project
The system SHALL allow the user to remove an opened project from the application without deleting the project's directory or files from disk.

#### Scenario: User confirms project removal
- **WHEN** the user requests to delete opened project `/home/user/work/demo-app`
- **AND** confirms the deletion
- **THEN** the project list no longer contains `/home/user/work/demo-app`
- **AND** the persisted opened project list no longer contains `/home/user/work/demo-app`
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal
- **WHEN** the user requests to delete an opened project
- **AND** cancels the confirmation
- **THEN** the project list remains unchanged

#### Scenario: Active project is removed
- **WHEN** the active project is removed
- **THEN** the system selects the remaining opened project with the most recent selection time as the active project
- **AND** if no opened projects remain, the active project is empty

#### Scenario: Removed project is not found
- **WHEN** the user requests to delete a project that is not in the opened project list
- **THEN** the system reports an error and leaves the opened project list unchanged
