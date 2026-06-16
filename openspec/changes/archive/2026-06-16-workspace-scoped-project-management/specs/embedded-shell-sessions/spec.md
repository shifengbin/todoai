## MODIFIED Requirements

### Requirement: Persist Terminal History Across Application Restart
The system SHALL persist terminal records and recent terminal output for active TODO project contexts in the current workspace data directory so that they are restored after that workspace is reopened. Restored terminals SHALL NOT restore shell processes, PTY state, or command execution, and SHALL be reported as non-running terminals. Terminal history SHALL NOT be shared globally across workspaces.

#### Scenario: Workspace reopen restores terminal records and output
- **WHEN** workspace `/work/customer-a` has TODO `fix-login` with status `in-progress`
- **AND** project `frontend-app` is associated with TODO `fix-login`
- **AND** the user creates terminal A under TODO `fix-login` and project `frontend-app`
- **AND** terminal A outputs `npm test`
- **AND** the user closes and reopens workspace `/work/customer-a`
- **THEN** terminal A appears under TODO `fix-login` and project `frontend-app`
- **AND** selecting terminal A shows output containing `npm test`
- **AND** terminal A is not reported as running
- **AND** no shell process is started automatically for terminal A

#### Scenario: Workspace reopen restores active terminal selection
- **WHEN** workspace `/work/customer-a` has TODO `fix-login` and project `frontend-app` with terminal A and terminal B
- **AND** terminal B is the active terminal in that TODO project context
- **AND** the user closes and reopens workspace `/work/customer-a`
- **THEN** terminal B remains the active terminal for TODO `fix-login` and project `frontend-app`
- **AND** terminal B's persisted output is shown when the workspace is restored

#### Scenario: Restored terminal history is capped
- **WHEN** terminal A produces output larger than the configured terminal history limit
- **AND** the user closes and reopens the owning workspace
- **THEN** terminal A is restored with only its most recent output up to the configured limit
- **AND** terminal A remains selectable under its TODO project context

#### Scenario: Missing terminal history file is treated as empty
- **WHEN** persisted TODO and project data exists in the current workspace
- **AND** terminal history storage is missing from that workspace data directory
- **AND** the user opens the workspace
- **THEN** the application loads the TODO workspace without error
- **AND** no shell process is started solely to recreate missing terminal history

#### Scenario: Terminal history is isolated by workspace
- **WHEN** terminal A has persisted output in workspace `/work/customer-a`
- **AND** the user opens workspace `/work/customer-b`
- **THEN** terminal A does not appear in `/work/customer-b`
- **AND** terminal A's output is not restored in `/work/customer-b`

