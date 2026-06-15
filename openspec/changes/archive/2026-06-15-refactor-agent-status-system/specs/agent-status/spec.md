## ADDED Requirements

### Requirement: Maintain Unified Agent Activity Status

The system SHALL maintain a unified runtime agent activity status for each running terminal. The status SHALL include phase, source, confidence, reason, and update time. Supported phases SHALL include `idle`, `busy`, `needs-input`, `done`, `failed`, and `exited`.

#### Scenario: Terminal starts with idle agent status

- **WHEN** a new terminal is created and its shell process starts
- **THEN** the system creates an agent activity status for that terminal
- **AND** the status phase is `idle`
- **AND** the status source is `shell`

#### Scenario: Shell exit clears active agent status

- **WHEN** a terminal has agent phase `busy` or `needs-input`
- **AND** the owning shell process exits
- **THEN** the system updates that terminal's agent phase to `exited`
- **AND** lower-priority title or command events no longer display that terminal as busy

### Requirement: Prioritize Structured Agent Signals

The system SHALL prioritize structured agent status signals over terminal title-derived signals. A lower-priority title fallback event MUST NOT replace a newer structured status for the same terminal.

#### Scenario: Structured busy is not cleared by stable title

- **WHEN** a terminal receives a structured Codex or Claude event indicating agent phase `busy`
- **AND** the same terminal later receives a stable terminal title update such as `codex`
- **THEN** the terminal agent phase remains `busy`
- **AND** the status source remains the structured source

#### Scenario: Structured needs-input overrides busy title

- **WHEN** a terminal receives a title update that would classify as `busy`
- **AND** the same terminal receives a structured Claude notification indicating input is needed
- **THEN** the terminal agent phase becomes `needs-input`
- **AND** TODO activity summary treats that terminal as needing input

#### Scenario: Title fallback applies when no structured signal exists

- **WHEN** a terminal has no current structured Claude or Codex status
- **AND** the terminal receives a title update containing an explicit busy signal
- **THEN** the terminal agent phase becomes `busy`
- **AND** the status source is `title-fallback`
- **AND** the status confidence is `heuristic`

### Requirement: Map Claude Status Sources

The system SHALL map supported Claude status sources into the unified agent activity model. Claude background session JSON and Claude hook events SHALL be treated as structured or authoritative sources.

#### Scenario: Claude background session blocked on input

- **WHEN** Claude background session status data reports `state` as `blocked`
- **AND** `waitingFor` indicates input is needed or a permission prompt is pending
- **THEN** the matched terminal or TODO context agent phase becomes `needs-input`
- **AND** the status source is `claude-agents-json`

#### Scenario: Claude hook notification marks input needed

- **WHEN** a Claude hook event with notification type `permission_prompt` or `idle_prompt` is received for a terminal
- **THEN** the terminal agent phase becomes `needs-input`
- **AND** the status source is `claude-hook`

#### Scenario: Claude stop marks terminal idle

- **WHEN** a Claude `Stop` hook event is received for a running terminal
- **THEN** the terminal agent phase becomes `idle`
- **AND** the terminal command label remains unchanged unless a shell command-end event is received

#### Scenario: Claude session end marks done or exited

- **WHEN** a Claude `SessionEnd` hook event is received for a terminal whose shell is still running
- **THEN** the terminal agent phase becomes `done`
- **AND** the terminal shell state remains `running`

### Requirement: Map Codex Status Sources

The system SHALL map supported Codex JSONL, app-server, and hook events into the unified agent activity model. Codex machine-readable turn and item events SHALL take priority over terminal title fallback.

#### Scenario: Codex JSONL turn starts

- **WHEN** a `codex exec --json` stream emits `turn.started` for a terminal
- **THEN** the terminal agent phase becomes `busy`
- **AND** the status source is `codex-jsonl`

#### Scenario: Codex JSONL turn completes

- **WHEN** a `codex exec --json` stream emits `turn.completed` for a terminal
- **THEN** the terminal agent phase becomes `done`
- **AND** command execution items from that completed turn no longer keep the terminal busy

#### Scenario: Codex JSONL turn fails

- **WHEN** a `codex exec --json` stream emits `turn.failed` or `error` for a terminal
- **THEN** the terminal agent phase becomes `failed`
- **AND** the TODO activity summary does not report that terminal as still running

#### Scenario: Codex app-server item execution is busy

- **WHEN** a Codex app-server notification reports an `item/started` command execution or tool call for a terminal thread
- **THEN** the terminal agent phase becomes `busy`
- **AND** the status source is `codex-app-server`

#### Scenario: Codex stop hook marks idle

- **WHEN** a Codex `Stop` hook event is received for a running terminal
- **THEN** the terminal agent phase becomes `idle`
- **AND** title fallback cannot keep the terminal busy unless a later title event is the freshest available signal

### Requirement: Summarize Agent Activity For Todos

The system SHALL summarize terminal agent activity into TODO activity state. `needs-input` SHALL take precedence over `busy`, and terminal shell exit SHALL remove busy or needs-input activity from that terminal.

#### Scenario: Todo with needs-input terminal is prioritized

- **WHEN** TODO `fix-login` has terminal A with phase `busy`
- **AND** TODO `fix-login` has terminal B with phase `needs-input`
- **THEN** TODO `fix-login` activity summary is `needs-input`

#### Scenario: Todo with only busy terminals is busy

- **WHEN** TODO `fix-login` has one or more terminals with phase `busy`
- **AND** no terminal under that TODO has phase `needs-input`
- **THEN** TODO `fix-login` activity summary is `busy`

#### Scenario: Exited terminal does not keep todo active

- **WHEN** TODO `fix-login` has a terminal with phase `busy`
- **AND** that terminal shell exits
- **THEN** TODO `fix-login` activity summary no longer treats that terminal as busy
