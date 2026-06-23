## MODIFIED Requirements

### Requirement: Label Terminal By Command State

The system SHALL display each terminal's shell name when that terminal is idle and SHALL display the currently executing command while that terminal is running a command. When the command finishes, the terminal label SHALL return to the shell name. Shell integrations that report command state MUST NOT emit `command-end` unless a corresponding command has started in that shell session.

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

#### Scenario: Zsh idle prompt does not end a command

- **WHEN** a zsh terminal starts and reaches its initial prompt without executing a user command
- **THEN** the zsh integration does not emit `command-end`
- **AND** the terminal command label is not cleared by the initial prompt

#### Scenario: Zsh command end follows command start

- **WHEN** terminal A runs command `npm test` in a zsh shell
- **AND** the zsh integration emits `command-start` for `npm test`
- **AND** the command finishes and the shell returns to the prompt
- **THEN** the zsh integration emits one `command-end` for that command
- **AND** terminal A's label returns to `zsh`
