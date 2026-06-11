## ADDED Requirements

### Requirement: Bulk Remove Opened Projects
The system SHALL allow the user to select one or more opened projects in the project library and remove the selected projects in one confirmed bulk action. Bulk removal SHALL remove only application records, active TODO project associations, and runtime terminal sessions for the selected projects; it SHALL NOT delete directories or files from disk. Bulk removal SHALL be all-or-nothing when any requested project ID is invalid.

#### Scenario: User selects projects for bulk removal
- **WHEN** the user opens the `项目` tab
- **AND** the project library contains `frontend-app` and `api-service`
- **AND** the user checks `frontend-app`
- **AND** the user checks `api-service`
- **THEN** the project library shows both projects as selected
- **AND** the bulk delete action is enabled
- **AND** the bulk delete action indicates that 2 projects are selected

#### Scenario: Bulk delete is unavailable without selection
- **WHEN** the user opens the `项目` tab
- **AND** no project is selected
- **THEN** the bulk delete action is disabled
- **AND** activating the bulk delete action does not remove any project

#### Scenario: User confirms bulk project removal
- **WHEN** the user selects projects `frontend-app` and `api-service` in the project library
- **AND** requests bulk deletion
- **THEN** the system shows a confirmation popover for deleting the 2 selected projects
- **WHEN** the user confirms the bulk deletion
- **THEN** the project list no longer contains `frontend-app`
- **AND** the project list no longer contains `api-service`
- **AND** the persisted opened project list no longer contains `frontend-app`
- **AND** the persisted opened project list no longer contains `api-service`
- **AND** active TODOs no longer contain associations to `frontend-app` or `api-service`
- **AND** runtime terminal sessions for `frontend-app` and `api-service` are closed
- **AND** the directories for `frontend-app` and `api-service` remain on disk
- **AND** the project selection is cleared

#### Scenario: User cancels bulk project removal
- **WHEN** the user selects projects `frontend-app` and `api-service` in the project library
- **AND** requests bulk deletion
- **AND** the system shows the bulk deletion confirmation popover
- **WHEN** the user cancels the confirmation
- **THEN** the project list still contains `frontend-app`
- **AND** the project list still contains `api-service`
- **AND** TODO project associations remain unchanged
- **AND** runtime terminal sessions remain unchanged

#### Scenario: Active project is removed by bulk deletion
- **WHEN** the active project is `frontend-app`
- **AND** the user bulk deletes `frontend-app` and `api-service`
- **THEN** the system selects the remaining opened project with the most recent selection time as the selected project for project management
- **AND** if no opened projects remain, the selected project is empty
- **AND** any active terminal owned by removed projects is cleared

#### Scenario: Bulk removal request includes a missing project
- **WHEN** the system receives a bulk removal request for `frontend-app` and missing project `missing-project`
- **THEN** the system reports an error
- **AND** the opened project list remains unchanged
- **AND** TODO project associations remain unchanged
- **AND** runtime terminal sessions remain unchanged

## MODIFIED Requirements

### Requirement: Remove Opened Project

The system SHALL allow the user to remove an opened project from the application without deleting the project's directory or files from disk. Removing a project SHALL require confirmation in a contextual popover anchored to the project row delete button. Removing a project SHALL remove that project from active TODO project associations and SHALL close runtime terminal sessions for that project across all TODO contexts. Archived TODO project snapshots SHALL remain unchanged.

#### Scenario: User opens project removal confirmation

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app` from the project library row
- **THEN** the system shows a confirmation popover next to that project row delete button
- **AND** the system does not use the browser native confirmation dialog
- **AND** the project list remains unchanged until the user confirms the popover

#### Scenario: User confirms project removal

- **WHEN** the user requests to delete opened project `/home/user/work/demo-app`
- **AND** the system shows the project removal confirmation popover
- **AND** the user confirms the deletion
- **THEN** the project list no longer contains `/home/user/work/demo-app`
- **AND** the persisted opened project list no longer contains `/home/user/work/demo-app`
- **AND** active TODOs no longer contain associations to `/home/user/work/demo-app`
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal

- **WHEN** the user requests to delete an opened project
- **AND** the system shows the project removal confirmation popover
- **AND** the user cancels the confirmation
- **THEN** the project list remains unchanged
- **AND** TODO project associations remain unchanged

#### Scenario: Active project is removed

- **WHEN** the active project is removed
- **THEN** the system selects the remaining opened project with the most recent selection time as the selected project for project management
- **AND** if no opened projects remain, the selected project is empty
- **AND** any active terminal owned by the removed project is cleared

#### Scenario: Removed project is not found

- **WHEN** the user requests to delete a project that is not in the opened project list
- **THEN** the system reports an error and leaves the opened project list unchanged
- **AND** TODO project associations remain unchanged
