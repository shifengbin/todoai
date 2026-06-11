## ADDED Requirements

### Requirement: Handle Unsupported Windows Embedded Shell Backend

The system SHALL detect when the embedded shell backend is unsupported on Windows and SHALL present a stable unavailable state without repeatedly attempting to start a shell process.

#### Scenario: Windows embedded shell backend is unsupported

- **WHEN** a Windows user creates or restarts an embedded terminal
- **AND** the configured PTY backend does not support Windows shell sessions
- **THEN** the system marks the terminal shell as exited or unavailable
- **AND** the application remains usable
- **AND** the terminal area displays a clear unsupported-platform message

#### Scenario: Unsupported backend does not retry automatically

- **WHEN** a Windows terminal start fails because the embedded shell backend is unsupported
- **THEN** the system does not automatically start another shell process for the same terminal
- **AND** switching focus back to the application does not retry the failed terminal start

#### Scenario: Unsupported backend does not show system console windows

- **WHEN** a Windows terminal start fails because the embedded shell backend is unsupported
- **THEN** the failure is reported inside the application
- **AND** no system console window is displayed for the failed embedded terminal start
