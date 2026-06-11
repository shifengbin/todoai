## ADDED Requirements

### Requirement: Start Windows Embedded Shell With ConPTY

The system SHALL start embedded shell sessions through a Windows ConPTY backend when the application runs on a Windows version that supports ConPTY.

#### Scenario: Windows ConPTY shell starts in project directory

- **WHEN** the application runs on Windows 10 1809 or later
- **AND** the configured terminal shell path resolves to an available Windows shell such as `pwsh.exe`, `powershell.exe`, or `cmd.exe`
- **AND** the user creates an embedded terminal for an available TODO project context
- **THEN** the system starts the shell through the Windows ConPTY backend
- **AND** the shell process working directory is the owning project's path
- **AND** the terminal state becomes `running`
- **AND** shell output is emitted to the owning terminal session

#### Scenario: Windows ConPTY terminal receives input

- **WHEN** a Windows ConPTY-backed terminal is running
- **AND** the user types in the active embedded terminal
- **THEN** the input is written to that ConPTY shell session
- **AND** the input is not written to other terminal sessions

#### Scenario: Windows ConPTY terminal resizes

- **WHEN** a Windows ConPTY-backed terminal is running
- **AND** the active terminal viewport rows or columns change
- **THEN** the Windows ConPTY backend receives the updated terminal size

#### Scenario: Windows ConPTY terminal closes on removal

- **WHEN** the user removes a running Windows ConPTY-backed terminal
- **THEN** the system closes the ConPTY process
- **AND** removes the terminal session from runtime state
- **AND** the removed terminal no longer receives input or output

### Requirement: Preserve Unsupported State For Windows Without ConPTY

The system SHALL keep the existing unsupported embedded terminal behavior when Windows ConPTY is unavailable.

#### Scenario: Windows version does not support ConPTY

- **WHEN** the application runs on a Windows version where ConPTY is unavailable
- **AND** the user creates or restarts an embedded terminal
- **THEN** the system marks the terminal shell as `unsupported`
- **AND** the application remains usable
- **AND** the terminal area displays the unsupported-platform message

#### Scenario: Windows ConPTY initialization reports unsupported

- **WHEN** the application runs on Windows
- **AND** the configured terminal shell path resolves to an available Windows shell
- **AND** ConPTY initialization fails because the backend is unsupported in the current environment
- **THEN** the system marks the terminal shell as `unsupported`
- **AND** the system does not automatically start another shell process for the same terminal

#### Scenario: Windows shell configuration errors remain startup errors

- **WHEN** the application runs on Windows with ConPTY support
- **AND** the configured shell path is invalid or cannot be started
- **AND** the failure is not an unsupported ConPTY backend failure
- **THEN** the system reports a shell startup error
- **AND** the system does not mark the failure as unsupported
