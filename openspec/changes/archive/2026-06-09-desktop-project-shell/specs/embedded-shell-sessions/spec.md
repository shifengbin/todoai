## ADDED Requirements

### Requirement: Start Shell In Project Directory

The system SHALL start an embedded shell session in the active project's directory when the project is selected and no live session exists for that project.

#### Scenario: Shell starts with project working directory

- **WHEN** the user selects a project with path `/home/user/work/demo-app`
- **THEN** the shell session for that project starts with working directory `/home/user/work/demo-app`

### Requirement: Maintain One Runtime Shell Session Per Project

The system SHALL maintain at most one live shell session per opened project while the application is running.

#### Scenario: Switching projects keeps previous shell alive

- **WHEN** the user selects project A, starts a long-running command, then selects project B
- **THEN** project A's shell session remains running in the background

#### Scenario: Switching back restores the previous session

- **WHEN** the user switches from project A to project B and then back to project A
- **THEN** the shell area shows project A's existing session instead of creating a new shell

### Requirement: Route Terminal Input To Active Session

The system SHALL send user terminal input only to the currently active project's shell session.

#### Scenario: User types after switching projects

- **WHEN** the user switches from project A to project B and types in the terminal
- **THEN** the input is sent to project B's shell session and not project A's shell session

### Requirement: Route Terminal Output By Project

The system SHALL associate shell output with the project session that produced it so output is displayed in the correct terminal state.

#### Scenario: Background project produces output

- **WHEN** project A is running a command in the background while project B is active
- **THEN** project A's output is retained for project A and is not shown in project B's terminal

### Requirement: Resize Active Shell PTY

The system SHALL resize the active project's PTY when the terminal viewport dimensions change.

#### Scenario: Terminal viewport changes size

- **WHEN** the user resizes the application window and the terminal rows or columns change
- **THEN** the active project's PTY receives the updated terminal size

### Requirement: Handle Shell Exit

The system SHALL detect when a project shell exits and show that session as exited without closing the application.

#### Scenario: Shell process exits

- **WHEN** the active project's shell process exits
- **THEN** the application marks the shell session as exited and remains usable
