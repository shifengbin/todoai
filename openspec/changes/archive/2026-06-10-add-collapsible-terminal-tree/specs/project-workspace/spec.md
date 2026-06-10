## ADDED Requirements

### Requirement: Collapse Project Terminal Branches
The system SHALL allow the user to expand and collapse each project's terminal child rows independently in the left-side project terminal tree.

#### Scenario: User collapses a project branch
- **WHEN** a project branch is expanded and has terminal child rows
- **AND** the user activates that project's collapse control
- **THEN** the terminal child rows for that project are hidden
- **AND** the project row remains visible

#### Scenario: User expands a project branch
- **WHEN** a project branch is collapsed and has terminal child rows
- **AND** the user activates that project's expand control
- **THEN** the terminal child rows for that project are shown beneath the project row

#### Scenario: Collapsing one project does not affect another project
- **WHEN** project A and project B both have terminal child rows
- **AND** the user collapses project A
- **THEN** project A's terminal child rows are hidden
- **AND** project B's expanded or collapsed state is unchanged

### Requirement: Reveal Active Project Terminal Branch
The system SHALL expand the branch for the project that becomes active through project selection, terminal selection, or terminal creation.

#### Scenario: User selects a collapsed project
- **WHEN** a project branch is collapsed
- **AND** the user selects that project
- **THEN** that project's branch is expanded

#### Scenario: User selects a terminal under a project
- **WHEN** a terminal becomes active under a project
- **THEN** the owning project's branch is expanded

#### Scenario: User creates a terminal under a project
- **WHEN** the user creates a terminal under an available project
- **THEN** the owning project's branch is expanded
- **AND** the new terminal row is visible under that project

### Requirement: Show Project Terminal Hierarchy
The system SHALL visually distinguish project parent rows from terminal child rows and SHALL communicate terminal ownership through tree indentation or branch guides.

#### Scenario: Project has terminal children
- **WHEN** a project has visible terminal child rows
- **THEN** the terminal rows appear visually nested under the project row
- **AND** the sidebar communicates that the terminal rows belong to that project

#### Scenario: Project branch is collapsed
- **WHEN** a project with terminal child rows is collapsed
- **THEN** the project row indicates that hidden terminal children can be expanded
