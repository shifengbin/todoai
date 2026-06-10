## MODIFIED Requirements

### Requirement: Show Project Terminal Launch Menu

The system SHALL show a terminal launch menu when the user activates the add-terminal control for an available project. The system SHALL keep the launch menu visible and operable within the left-side project list's visible area when the menu is opened near the list boundary.

#### Scenario: Launch menu contains terminal and configured profiles

- **WHEN** settings contains launch profiles named `codex` and `claude`
- **AND** the user activates the add-terminal control for an available project
- **THEN** the launch menu shows `Terminal` as the first option
- **AND** the launch menu shows `codex` and `claude` after `Terminal` in the configured order

#### Scenario: Unavailable project has no launch menu

- **WHEN** a project path is unavailable
- **THEN** the project row does not expose an add-terminal launch menu action

#### Scenario: Launch menu opens near the project list bottom

- **WHEN** the user activates the add-terminal control for the last visible available project in the left-side project list
- **AND** there is not enough visible space below the project row to show the launch menu
- **THEN** the launch menu remains within the visible project list area
- **AND** all launch options remain reachable without being clipped by the project list
