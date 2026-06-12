## ADDED Requirements

### Requirement: Persist Terminal History Across Application Restart

The system SHALL persist terminal records and recent terminal output for active TODO project contexts so that they are restored after the application is closed and reopened. Restored terminals SHALL NOT restore shell processes, PTY state, or command execution, and SHALL be reported as non-running terminals.

#### Scenario: Application restart restores terminal records and output

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** project `frontend-app` is associated with TODO `fix-login`
- **AND** the user creates terminal A under TODO `fix-login` and project `frontend-app`
- **AND** terminal A outputs `npm test`
- **AND** the user closes and reopens the application
- **THEN** terminal A appears under TODO `fix-login` and project `frontend-app`
- **AND** selecting terminal A shows output containing `npm test`
- **AND** terminal A is not reported as running
- **AND** no shell process is started automatically for terminal A

#### Scenario: Application restart restores active terminal selection

- **WHEN** TODO `fix-login` and project `frontend-app` have terminal A and terminal B
- **AND** terminal B is the active terminal in that TODO project context
- **AND** the user closes and reopens the application
- **THEN** terminal B remains the active terminal for TODO `fix-login` and project `frontend-app`
- **AND** terminal B's persisted output is shown when the workspace is restored

#### Scenario: Restored terminal history is capped

- **WHEN** terminal A produces output larger than the configured terminal history limit
- **AND** the user closes and reopens the application
- **THEN** terminal A is restored with only its most recent output up to the configured limit
- **AND** terminal A remains selectable under its TODO project context

#### Scenario: Missing terminal history file is treated as empty

- **WHEN** persisted TODO and project data exists
- **AND** terminal history storage is missing
- **AND** the user opens the application
- **THEN** the application loads the TODO workspace without error
- **AND** no shell process is started solely to recreate missing terminal history

### Requirement: Clean Persisted Terminal History

The system SHALL remove persisted terminal records and output history when their owning terminal, TODO, TODO project context, or project is removed from the active workspace.

#### Scenario: Removed terminal clears persisted history

- **WHEN** terminal A has persisted output history
- **AND** the user removes terminal A
- **AND** the user closes and reopens the application
- **THEN** terminal A does not appear in the TODO project terminal tree
- **AND** terminal A's output history is not restored

#### Scenario: Completed todo clears owned terminal history

- **WHEN** TODO `fix-login` owns terminal A with persisted output history
- **AND** the user completes TODO `fix-login`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored under active TODOs
- **AND** terminal A's output history is not restored

#### Scenario: Removed todo project clears owned terminal history

- **WHEN** TODO `fix-login` has project `frontend-app` with terminal A and persisted output history
- **AND** the user removes project `frontend-app` from TODO `fix-login`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored
- **AND** terminal A's output history is not restored

#### Scenario: Deleted project clears owned terminal history

- **WHEN** project `frontend-app` owns terminal A across one or more TODO contexts
- **AND** terminal A has persisted output history
- **AND** the user deletes project `frontend-app`
- **AND** the user closes and reopens the application
- **THEN** terminal A is not restored
- **AND** terminal A's output history is not restored
