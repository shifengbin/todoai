## ADDED Requirements

### Requirement: Preserve Launch Profile Label Across Idle Command-End Events

The system SHALL keep a launch profile command label visible for a running terminal when an idle or unpaired command-end event is received before that launch profile command has reported a matching command-start. The system MUST still clear the command label when a real command that reported command-start later reports command-end.

#### Scenario: Launch profile label survives initial zsh command-end race

- **WHEN** the user creates terminal A from a launch profile with command `codex`
- **AND** the system sets terminal A's launch profile label to `codex`
- **AND** terminal A receives a `command-end` event from the shell before receiving any matching `command-start` event
- **THEN** terminal A's visible label remains `codex`
- **AND** terminal A's label does not fall back to `zsh`

#### Scenario: Real command end still clears command label

- **WHEN** terminal A has label `codex` after receiving a `command-start` event for `codex`
- **AND** terminal A receives a later `command-end` event for that running command
- **THEN** terminal A's visible label returns to its shell name
