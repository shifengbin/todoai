## MODIFIED Requirements

### Requirement: Create Todo Task Terminal

The system SHALL allow the user to create task-level terminal sessions for an `in-progress` TODO whose task workspace directory exists. Each task-level terminal SHALL start an independent shell process in the TODO task workspace directory. Task-level terminals SHALL belong to the TODO but SHALL NOT belong to any TODO project. Selecting or creating a task-level terminal SHALL make the owning TODO the active task context and SHALL clear the active TODO project and active project context.

#### Scenario: User creates task terminal for in-progress todo

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** TODO `fix-login` has task workspace directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the user creates a task terminal under TODO `fix-login`
- **THEN** the system starts a new shell process with working directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the terminal records TODO `fix-login` as owner
- **AND** the terminal does not record a TODO project owner
- **AND** the task terminal becomes the active terminal
- **AND** TODO `fix-login` becomes the active task context
- **AND** the active TODO project and active project context are empty

#### Scenario: Not-started todo cannot create task terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user requests a task terminal under TODO `fix-login`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no task terminal is added to TODO `fix-login`

#### Scenario: Task terminal selects todo task context

- **WHEN** current TODO project context is TODO `fix-login` under project `frontend-app`
- **AND** the user selects a task terminal under TODO `fix-login`
- **THEN** the task terminal becomes the active terminal
- **AND** TODO `fix-login` becomes the active task context
- **AND** the active TODO project and active project context are empty
