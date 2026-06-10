## ADDED Requirements

### Requirement: Remove Runtime Terminal Session
The system SHALL allow the user to remove a runtime terminal session from an opened project. If the terminal session has a running PTY process, the system SHALL close that process before removing the terminal session from runtime state.

#### Scenario: User removes a running terminal
- **WHEN** the user deletes a terminal session with a running shell process
- **THEN** the system closes that shell process
- **AND** the terminal no longer appears under its project

#### Scenario: User removes an exited terminal
- **WHEN** the user deletes a terminal session whose shell process has exited
- **THEN** the terminal no longer appears under its project
- **AND** no new shell process is started automatically

#### Scenario: Active terminal is removed
- **WHEN** the active terminal is removed from a project that still has other terminals
- **THEN** the system selects that project's most recently selected remaining terminal as the active terminal

#### Scenario: Last terminal is removed
- **WHEN** the last terminal under the active project is removed
- **THEN** the active terminal is empty
- **AND** no replacement terminal is created automatically

#### Scenario: Removed terminal is not found
- **WHEN** the user requests to delete a terminal that is not in runtime terminal state
- **THEN** the system reports an error and leaves runtime terminal state unchanged

### Requirement: Remove Project Terminal Sessions
The system SHALL close and remove all runtime terminal sessions owned by a project when that project is removed from the application.

#### Scenario: Project with running terminals is removed
- **WHEN** the user confirms deletion of a project that owns running terminal sessions
- **THEN** the system closes every running shell process owned by that project
- **AND** removes those terminal sessions from runtime state

#### Scenario: Project terminal cleanup preserves other projects
- **WHEN** the user deletes project A while project B has terminal sessions
- **THEN** project A's terminal sessions are removed
- **AND** project B's terminal sessions remain available
