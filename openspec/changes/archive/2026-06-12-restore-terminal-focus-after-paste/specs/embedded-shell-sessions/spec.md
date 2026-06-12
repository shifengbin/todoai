## MODIFIED Requirements

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
- **AND** focus returns to that active terminal

#### Scenario: Paste empty clipboard from context menu

- **WHEN** the user chooses Paste from the terminal context menu and the system clipboard has no text
- **THEN** no terminal input is sent to the active terminal's shell
- **AND** the menu closes
- **AND** focus returns to that active terminal
