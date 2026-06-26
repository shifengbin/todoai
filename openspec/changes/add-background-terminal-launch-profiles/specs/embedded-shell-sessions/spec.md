## ADDED Requirements

### Requirement: Start Background Launch Profile Command

The system SHALL start background launch profile commands without registering an embedded terminal session. Background launch commands SHALL use the configured terminal shell in one-shot command mode, SHALL run in the same working directory that a visible terminal for the selected context would use, and SHALL wait for process exit to release resources. Background command output and exit state SHALL NOT be displayed in the terminal UI.

#### Scenario: Background project command runs in todo project worktree

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** project `demo-app` under TODO `fix-login` has prepared worktree path `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** the user selects a background launch profile with startup parameters `npm run sync` from that TODO project launch menu
- **THEN** the system starts a background command using working directory `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** no embedded terminal session is added to the TODO project context

#### Scenario: Background task command runs in todo workspace directory

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** TODO `fix-login` has task workspace directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the user selects a background launch profile with startup parameters `npm run prepare` from the task-level launch menu
- **THEN** the system starts a background command using working directory `/home/user/work/customer-a/tasks/abc123`
- **AND** no task terminal session is added to the TODO context

#### Scenario: Background command does not affect terminal UI state

- **WHEN** terminal `terminal-a` is the active terminal
- **AND** the user selects a background launch profile from a terminal launch menu
- **THEN** `terminal-a` remains the active terminal
- **AND** the terminal list remains unchanged
- **AND** the terminal history remains unchanged
- **AND** the system does not emit terminal output, terminal status, or terminal agent status events for the background command

#### Scenario: Background command exits without UI change

- **WHEN** a background launch profile command is running
- **AND** the background process exits
- **THEN** the system releases the process resources
- **AND** no terminal is marked as exited
- **AND** no new terminal is displayed

#### Scenario: Not-started todo project cannot start background command

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a background launch command under TODO `fix-login` and project `demo-app`
- **THEN** the system rejects the background command request
- **AND** no background process is started
- **AND** no runtime terminal session is added to that TODO project context

#### Scenario: Todo project with failed worktree cannot start background command

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** project `demo-app` under TODO `fix-login` has worktree status `failed`
- **AND** the user requests a background launch command for that TODO project
- **THEN** the system rejects the background command request
- **AND** no background process is started
- **AND** no runtime terminal session is added to that TODO project context

#### Scenario: Background command start failure is reported without terminal creation

- **WHEN** the user selects a background launch profile whose command cannot be started
- **THEN** the system reports the startup error through the existing application error display
- **AND** no embedded terminal session is added
- **AND** the current active terminal remains unchanged
