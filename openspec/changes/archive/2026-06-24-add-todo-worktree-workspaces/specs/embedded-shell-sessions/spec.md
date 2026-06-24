## MODIFIED Requirements

### Requirement: Start Shell In Project Directory

The system SHALL start an embedded shell session in the owning TODO project's prepared worktree directory when a terminal session is created for an available TODO project context. If the TODO project has not prepared a worktree directory, the system SHALL reject terminal creation for that TODO project and SHALL NOT start a shell process.

#### Scenario: Shell starts with todo project worktree directory

- **WHEN** the user creates a terminal for TODO `fix-login` under project `demo-app`
- **AND** the TODO project worktree path is `/home/user/work/customer-a/tasks/abc123/demo-app`
- **THEN** the shell session for that terminal starts with working directory `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** the terminal is owned by the TODO project context for `fix-login` and `demo-app`

#### Scenario: Todo project without prepared worktree cannot start shell

- **WHEN** the user creates a terminal for TODO `fix-login` under project `demo-app`
- **AND** the TODO project has no prepared worktree path
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started

### Requirement: Create Additional Project Terminal

The system SHALL allow the user to create additional terminal sessions only for an available project within an `in-progress` TODO project context whose worktree is prepared. Each created terminal session SHALL start an independent shell process in the owning TODO project's worktree directory and SHALL belong only to that TODO project context. The system SHALL reject terminal creation for `not-started` TODO project contexts or TODO project contexts without a prepared worktree.

#### Scenario: User creates another terminal for an in-progress todo project

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** project `demo-app` under TODO `fix-login` has prepared worktree path `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** the user creates a new terminal under TODO `fix-login` and project `demo-app`
- **THEN** the system starts a new shell process with working directory `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** the new terminal is independent from existing terminal sessions in that TODO project context
- **AND** the new terminal is not shown under other TODOs that reference the same project

#### Scenario: Not-started todo project cannot create terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user or client requests a new terminal under TODO `fix-login` and project `/home/user/work/demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no runtime terminal session is added to that TODO project context

#### Scenario: Todo project with failed worktree cannot create terminal

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** project `demo-app` under TODO `fix-login` has worktree status `failed`
- **AND** the user requests a new terminal under TODO `fix-login` and project `demo-app`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no runtime terminal session is added to that TODO project context

## ADDED Requirements

### Requirement: Create Todo Task Terminal

The system SHALL allow the user to create task-level terminal sessions for an `in-progress` TODO whose task workspace directory exists. Each task-level terminal SHALL start an independent shell process in the TODO task workspace directory. Task-level terminals SHALL belong to the TODO but SHALL NOT belong to any TODO project.

#### Scenario: User creates task terminal for in-progress todo

- **WHEN** TODO `fix-login` has status `in-progress`
- **AND** TODO `fix-login` has task workspace directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the user creates a task terminal under TODO `fix-login`
- **THEN** the system starts a new shell process with working directory `/home/user/work/customer-a/tasks/abc123`
- **AND** the terminal records TODO `fix-login` as owner
- **AND** the terminal does not record a TODO project owner

#### Scenario: Not-started todo cannot create task terminal

- **WHEN** TODO `fix-login` has status `not-started`
- **AND** the user requests a task terminal under TODO `fix-login`
- **THEN** the system rejects the terminal creation request
- **AND** no shell process is started
- **AND** no task terminal is added to TODO `fix-login`

#### Scenario: Task terminal does not change todo project context

- **WHEN** current TODO project context is TODO `fix-login` under project `frontend-app`
- **AND** the user selects a task terminal under TODO `fix-login`
- **THEN** the task terminal becomes the active terminal
- **AND** current TODO project context remains TODO `fix-login` under project `frontend-app`
