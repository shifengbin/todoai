## MODIFIED Requirements

### Requirement: Persist Opened Projects
The system SHALL persist the opened project list inside the current workspace data directory and SHALL reload it when that workspace is opened again. Opened project lists SHALL NOT be shared globally across workspaces.

#### Scenario: Project list is restored after reopening workspace
- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates projects in that workspace
- **AND** the user closes and reopens workspace `/home/user/work/customer-a`
- **THEN** the previously opened projects appear in the left-side project list

#### Scenario: Project list is isolated by workspace
- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates project `/home/user/repos/frontend-a`
- **AND** the user opens workspace `/home/user/work/customer-b`
- **THEN** project `/home/user/repos/frontend-a` does not appear in the project list for `/home/user/work/customer-b`

#### Scenario: No workspace has no opened project list
- **WHEN** no workspace is open
- **THEN** the project list is empty
- **AND** creating or importing opened projects is unavailable until a workspace is opened

