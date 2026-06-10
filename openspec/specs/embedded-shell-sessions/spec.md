# embedded-shell-sessions Specification

## Purpose
TBD - created by archiving change desktop-project-shell. Update Purpose after archive.

## Requirements
### Requirement: Start Shell In Project Directory

The system SHALL start an embedded shell session in the owning project's directory when a terminal session is created for an available project.

#### Scenario: Shell starts with project working directory

- **WHEN** the user creates a terminal for a project with path `/home/user/work/demo-app`
- **THEN** the shell session for that terminal starts with working directory `/home/user/work/demo-app`

### Requirement: Create Additional Project Terminal

The system SHALL allow the user to create additional terminal sessions for an available opened project. Each created terminal session SHALL start an independent shell process in the owning project's directory.

#### Scenario: User creates another terminal for a project

- **WHEN** the user creates a new terminal under project `/home/user/work/demo-app`
- **THEN** the system starts a new shell process with working directory `/home/user/work/demo-app`
- **AND** the new terminal is independent from existing terminal sessions for that project

### Requirement: Maintain Multiple Runtime Shell Sessions Per Project

The system SHALL support multiple live shell sessions per opened project while the application is running.

#### Scenario: Same project runs commands in separate terminals

- **WHEN** the user creates terminal A and terminal B under the same project
- **AND** starts a long-running command in terminal A
- **THEN** terminal B remains available for independent shell input
- **AND** terminal A's command continues running in the background

#### Scenario: Switching terminals keeps previous terminal alive

- **WHEN** the user switches from terminal A to terminal B and then back to terminal A
- **THEN** the shell area shows terminal A's existing session instead of creating a new shell

### Requirement: Route Terminal Input To Active Session

The system SHALL send user terminal input only to the currently active terminal's shell session.

#### Scenario: User types after switching terminals

- **WHEN** the user switches from terminal A to terminal B and types in the terminal
- **THEN** the input is sent to terminal B's shell session and not terminal A's shell session

### Requirement: Route Terminal Output By Project

The system SHALL associate shell output with the terminal session that produced it so output is displayed in the correct terminal state under the owning project.

#### Scenario: Background terminal produces output

- **WHEN** terminal A is running a command in the background while terminal B is active
- **THEN** terminal A's output is retained for terminal A
- **AND** terminal A's output is not shown in terminal B's terminal

### Requirement: Resize Active Shell PTY

The system SHALL resize the active terminal's PTY when the terminal viewport dimensions change.

#### Scenario: Terminal viewport changes size

- **WHEN** the user resizes the application window and the terminal rows or columns change
- **THEN** the active terminal's PTY receives the updated terminal size

### Requirement: Handle Shell Exit

The system SHALL detect when a terminal shell exits and show that terminal as exited without closing the application or other terminal sessions.

#### Scenario: Shell process exits

- **WHEN** the active terminal's shell process exits
- **THEN** the application marks that terminal session as exited
- **AND** the application remains usable

### Requirement: Label Terminal By Command State

The system SHALL display each terminal's shell name when that terminal is idle and SHALL display the currently executing command while that terminal is running a command. When the command finishes, the terminal label SHALL return to the shell name.

#### Scenario: Terminal starts a command

- **WHEN** terminal A is idle with label `zsh`
- **AND** the user starts command `npm run dev`
- **THEN** terminal A's label becomes `npm run dev`

#### Scenario: Terminal command finishes

- **WHEN** terminal A is labeled `npm run dev` because that command is running
- **AND** the command finishes and the shell returns to the prompt
- **THEN** terminal A's label becomes `zsh`

#### Scenario: Shell command state is unavailable

- **WHEN** a terminal's shell does not report command start or command end state
- **THEN** the terminal label remains the shell name

### Requirement: Use Configured Terminal Shell

The system SHALL start newly created embedded shell sessions with the configured terminal shell path when a usable setting exists.

#### Scenario: New terminal uses saved shell setting

- **WHEN** the terminal shell setting is saved as `/usr/bin/zsh`
- **AND** the user creates a new terminal for project `/home/user/work/demo-app`
- **THEN** the shell process starts with shell path `/usr/bin/zsh`
- **AND** the shell process working directory is `/home/user/work/demo-app`

#### Scenario: Existing terminal keeps original shell after setting changes

- **WHEN** terminal A was created with shell path `/usr/bin/bash`
- **AND** the user changes the terminal shell setting to `/usr/bin/zsh`
- **THEN** terminal A keeps using `/usr/bin/bash`
- **AND** a terminal created after the setting change uses `/usr/bin/zsh`

#### Scenario: New terminal uses fallback when saved shell is unavailable

- **WHEN** the saved terminal shell setting is unavailable
- **AND** automatic detection selects `/bin/sh` as the fallback shell
- **AND** the user creates a new embedded terminal
- **THEN** the shell process starts with shell path `/bin/sh`

### Requirement: Copy Terminal Selection To Clipboard

The system SHALL allow users to copy selected text from the active embedded terminal to the system clipboard without using plain `Ctrl+C`.

#### Scenario: Copy selected terminal text with shortcut

- **WHEN** the user has selected text in the active terminal and presses `Ctrl+Shift+C`
- **THEN** the selected text is written to the system clipboard

#### Scenario: Preserve shell interrupt shortcut

- **WHEN** the user presses plain `Ctrl+C` in the active terminal
- **THEN** the input is sent to the active shell instead of being handled as a clipboard copy action

### Requirement: Paste Clipboard Text Into Active Shell

The system SHALL allow users to paste system clipboard text into the active terminal's shell.

#### Scenario: Paste clipboard text with shortcut

- **WHEN** the user presses `Ctrl+Shift+V` in the active terminal and the system clipboard contains text
- **THEN** the clipboard text is sent to the active terminal's shell input

#### Scenario: Ignore empty clipboard paste

- **WHEN** the user triggers paste and the system clipboard has no text
- **THEN** no terminal input is sent to the active terminal's shell

### Requirement: Provide Terminal Clipboard Context Menu

The system SHALL provide a context menu in the active terminal area with Copy and Paste actions.

#### Scenario: Open terminal context menu

- **WHEN** the user right-clicks the active terminal area
- **THEN** the system shows a terminal context menu with Copy and Paste actions at the pointer location

#### Scenario: Copy from context menu

- **WHEN** the user chooses Copy from the terminal context menu while text is selected in the active terminal
- **THEN** the selected text is written to the system clipboard and the menu closes

#### Scenario: Paste from context menu

- **WHEN** the user chooses Paste from the terminal context menu and the system clipboard contains text
- **THEN** the clipboard text is sent to the active terminal's shell input and the menu closes

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
