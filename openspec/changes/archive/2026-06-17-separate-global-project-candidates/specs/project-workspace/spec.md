## MODIFIED Requirements

### Requirement: Create Project From Directory

The system SHALL allow the user to create a global project candidate by selecting a local directory through a native directory picker. The created candidate's default display name SHALL be the basename of the selected directory. Created project candidates SHALL be shared across all workspaces and SHALL be used only as selectable candidates for TODO engineering contexts.

#### Scenario: User creates a project from a directory

- **WHEN** the user clicks the new project action and selects `/home/user/work/demo-app`
- **THEN** the global project candidate list contains a project named `demo-app` with path `/home/user/work/demo-app`
- **AND** the candidate is available from any workspace

#### Scenario: User cancels directory selection

- **WHEN** the user opens the directory picker and cancels it
- **THEN** the global project candidate list remains unchanged

### Requirement: Persist Opened Projects

The system SHALL persist the project candidate list in the application-level configuration directory and SHALL reload it independently of the current workspace. Project candidate lists SHALL be shared globally across workspaces. Workspace TODO data SHALL remain isolated by workspace, and selecting a global candidate into a TODO SHALL create a workspace-local TODO project copy.

#### Scenario: Project list is restored after reopening workspace

- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates global candidate project `/home/user/repos/frontend-a`
- **AND** the user closes and reopens workspace `/home/user/work/customer-a`
- **THEN** the global candidate project `/home/user/repos/frontend-a` appears in TODO project selection controls

#### Scenario: Project list is shared by workspace

- **WHEN** the user opens workspace `/home/user/work/customer-a`
- **AND** the user creates global candidate project `/home/user/repos/frontend-a`
- **AND** the user opens workspace `/home/user/work/customer-b`
- **THEN** project `/home/user/repos/frontend-a` appears in TODO project selection controls for `/home/user/work/customer-b`

#### Scenario: No workspace has no TODO project creation

- **WHEN** no workspace is open
- **THEN** creating TODOs or adding projects to TODOs is unavailable
- **AND** the global project candidate list remains persisted and loadable

### Requirement: Select Active Project

The system SHALL select an active project for terminal and Git status context only through a TODO project context. Selecting or managing a global project candidate SHALL NOT create, select, reveal, or retarget a terminal session. Terminal activation SHALL occur only through a TODO project context.

#### Scenario: User selects a global project candidate for management

- **WHEN** the user clicks or focuses a project in the global candidate management UI
- **THEN** no shell session is created
- **AND** no terminal becomes active from that candidate interaction
- **AND** no TODO project context is selected

#### Scenario: User selects a project through a TODO context

- **WHEN** the user clicks project `demo-app` under TODO `fix-login`
- **THEN** project `demo-app` becomes the active project for the shell area
- **AND** the shell area is associated only with terminals under that TODO project context

### Requirement: Handle Duplicate Project Paths

The system SHALL avoid creating duplicate global project candidate entries for the same normalized absolute path. The system SHALL also avoid creating duplicate TODO project copies for the same normalized absolute path under the same TODO.

#### Scenario: User selects an already opened directory

- **WHEN** the user creates a global candidate from a directory that is already in the global candidate list
- **THEN** the existing candidate entry is updated or selected instead of adding a duplicate entry

#### Scenario: User adds an already linked path to the same todo

- **WHEN** TODO `修复登录问题` already contains a TODO project copy with path `/repo/frontend-app`
- **AND** the global candidate list contains another candidate with path `/repo/frontend-app`
- **AND** the user tries to add that candidate to TODO `修复登录问题`
- **THEN** TODO `修复登录问题` still contains only one TODO project with path `/repo/frontend-app`

### Requirement: Handle Missing Project Paths

The system SHALL detect when a persisted global candidate path or TODO project copy path no longer exists or is inaccessible. Missing global candidates SHALL remain visible as unavailable in candidate selection controls. Missing TODO project copies SHALL remain visible in TODO trees, and shell startup SHALL be prevented for that TODO project until the path is valid again.

#### Scenario: Persisted project path is missing

- **WHEN** the application starts and a persisted project path no longer exists
- **THEN** the global candidate or TODO project remains visible as unavailable
- **AND** selecting a TODO project with that path does not launch a shell

### Requirement: Remove Opened Project

The system SHALL allow the user to remove global project candidates from the application without deleting the project's directory or files from disk. Removing or clearing a global project candidate SHALL NOT remove workspace TODO project copies, SHALL NOT close runtime terminal sessions for existing TODO projects, and SHALL NOT change archived TODO project snapshots.

#### Scenario: User opens project removal confirmation

- **WHEN** the user requests to delete global candidate `/home/user/work/demo-app`
- **THEN** the system shows a confirmation popover next to that candidate delete button
- **AND** the system does not use the browser native confirmation dialog
- **AND** the global candidate list remains unchanged until the user confirms the popover

#### Scenario: User confirms project removal

- **WHEN** the user requests to delete global candidate `/home/user/work/demo-app`
- **AND** the system shows the project removal confirmation popover
- **AND** the user confirms the deletion
- **THEN** the global candidate list no longer contains `/home/user/work/demo-app`
- **AND** active TODOs that already copied `/home/user/work/demo-app` continue to contain their TODO project copies
- **AND** the directory `/home/user/work/demo-app` remains on disk

#### Scenario: User cancels project removal

- **WHEN** the user requests to delete a global candidate
- **AND** the system shows the project removal confirmation popover
- **AND** the user cancels the confirmation
- **THEN** the global candidate list remains unchanged
- **AND** TODO project copies remain unchanged

#### Scenario: Active todo project candidate is removed

- **WHEN** the active TODO project was copied from global candidate `/home/user/work/demo-app`
- **AND** that global candidate is removed
- **THEN** the active TODO project remains selected
- **AND** any terminal owned by that TODO project remains available according to its runtime state
- **AND** the active TODO project continues to use its copied path

#### Scenario: Removed project is not found

- **WHEN** the user requests to delete a candidate that is not in the global candidate list
- **THEN** the system reports an error and leaves the global candidate list unchanged
- **AND** TODO project copies remain unchanged
