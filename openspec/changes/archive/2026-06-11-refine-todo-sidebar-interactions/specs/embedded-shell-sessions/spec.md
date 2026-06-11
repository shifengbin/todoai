## ADDED Requirements

### Requirement: Remove Todo Project Terminal Sessions

The system SHALL close and remove all runtime terminal sessions owned by a single TODO project context when that TODO-project association is removed. Removing one TODO project context SHALL NOT close terminal sessions for the same project under other TODOs.

#### Scenario: Removed todo project closes owned terminals

- **WHEN** TODO `fix-login` has project `frontend-app` with running terminal sessions
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** the system closes every running shell process owned by that TODO project context
- **AND** removes those terminal sessions from runtime state
- **AND** project `frontend-app` no longer appears under TODO `fix-login`

#### Scenario: Todo project cleanup preserves same project in other todos

- **WHEN** project `frontend-app` is associated with TODO `fix-login`
- **AND** project `frontend-app` is associated with TODO `upgrade-deps`
- **AND** both TODO project contexts have terminal sessions
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** terminal sessions under TODO `fix-login` and project `frontend-app` are closed and removed
- **AND** terminal sessions under TODO `upgrade-deps` and project `frontend-app` remain available

#### Scenario: Active todo project is removed

- **WHEN** the active terminal belongs to TODO `fix-login` and project `frontend-app`
- **AND** the user confirms removing project `frontend-app` from TODO `fix-login`
- **THEN** the active terminal is cleared or moved to a remaining terminal in a valid TODO project context
- **AND** the removed terminal no longer receives input or output

## MODIFIED Requirements

### Requirement: Resize Active Shell PTY

The system SHALL resize the active terminal's PTY when the terminal viewport dimensions change, including application window resize and workspace sidebar width changes.

#### Scenario: Terminal viewport changes size

- **WHEN** the user resizes the application window and the terminal rows or columns change
- **THEN** the active terminal's PTY receives the updated terminal size

#### Scenario: Sidebar resize changes terminal viewport

- **WHEN** the user drags the workspace sidebar divider and the active terminal rows or columns change
- **THEN** the active terminal's PTY receives the updated terminal size
