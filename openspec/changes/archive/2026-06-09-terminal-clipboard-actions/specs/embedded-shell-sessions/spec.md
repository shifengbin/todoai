## ADDED Requirements

### Requirement: Copy Terminal Selection To Clipboard
The system SHALL allow users to copy selected text from the active embedded terminal to the system clipboard without using plain `Ctrl+C`.

#### Scenario: Copy selected terminal text with shortcut
- **WHEN** the user has selected text in the active terminal and presses `Ctrl+Shift+C`
- **THEN** the selected text is written to the system clipboard

#### Scenario: Preserve shell interrupt shortcut
- **WHEN** the user presses plain `Ctrl+C` in the active terminal
- **THEN** the input is sent to the active shell instead of being handled as a clipboard copy action

### Requirement: Paste Clipboard Text Into Active Shell
The system SHALL allow users to paste system clipboard text into the active project's embedded shell.

#### Scenario: Paste clipboard text with shortcut
- **WHEN** the user presses `Ctrl+Shift+V` in the active terminal and the system clipboard contains text
- **THEN** the clipboard text is sent to the active project's shell input

#### Scenario: Ignore empty clipboard paste
- **WHEN** the user triggers paste and the system clipboard has no text
- **THEN** no terminal input is sent to the active shell

### Requirement: Provide Terminal Clipboard Context Menu
The system SHALL provide a context menu in the terminal area with Copy and Paste actions.

#### Scenario: Open terminal context menu
- **WHEN** the user right-clicks the active terminal area
- **THEN** the system shows a terminal context menu with Copy and Paste actions at the pointer location

#### Scenario: Copy from context menu
- **WHEN** the user chooses Copy from the terminal context menu while text is selected in the active terminal
- **THEN** the selected text is written to the system clipboard and the menu closes

#### Scenario: Paste from context menu
- **WHEN** the user chooses Paste from the terminal context menu and the system clipboard contains text
- **THEN** the clipboard text is sent to the active project's shell input and the menu closes
