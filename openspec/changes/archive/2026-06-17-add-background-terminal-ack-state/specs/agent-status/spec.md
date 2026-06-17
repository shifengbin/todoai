## ADDED Requirements

### Requirement: Track Background Terminal Confirmation State

系统 SHALL 为后台终端维护运行时确认状态。确认状态 SHALL 是前端 UI 状态，不 SHALL 写入后端 agent phase，也不 SHALL 持久化到终端历史。仅当非当前激活终端的可视活动状态从 `busy` 切换为非活动状态时，系统 SHALL 标记该终端为 `needs-ack`。非活动状态包括 `idle`、`done`、`failed` 和 `exited` 对应的非忙碌展示状态。

#### Scenario: Background busy terminal becomes confirmation needed

- **WHEN** 终端 `terminal-b` 不是当前激活终端
- **AND** 终端 `terminal-b` 的可视活动状态为 `busy`
- **AND** 终端 `terminal-b` 收到使其不再忙碌的状态事件
- **THEN** 系统标记终端 `terminal-b` 的确认状态为 `needs-ack`
- **AND** 终端 `terminal-b` 的 agent phase 不被改写为 `needs-ack`

#### Scenario: Active busy terminal becomes idle without confirmation state

- **WHEN** 终端 `terminal-a` 是当前激活终端
- **AND** 终端 `terminal-a` 的可视活动状态为 `busy`
- **AND** 终端 `terminal-a` 收到使其不再忙碌的状态事件
- **THEN** 系统不标记终端 `terminal-a` 为 `needs-ack`

#### Scenario: Needs input terminal does not become confirmation needed

- **WHEN** 终端 `terminal-b` 不是当前激活终端
- **AND** 终端 `terminal-b` 的可视活动状态为 `needs-input`
- **AND** 终端 `terminal-b` 收到使其进入 `idle` 的状态事件
- **THEN** 系统不因该转换标记终端 `terminal-b` 为 `needs-ack`

#### Scenario: Selecting terminal clears confirmation state

- **WHEN** 终端 `terminal-b` 的确认状态为 `needs-ack`
- **AND** 用户选择终端 `terminal-b`
- **THEN** 系统清除终端 `terminal-b` 的确认状态
- **AND** 终端 `terminal-b` 按当前 agent phase 和 shell state 显示活动状态

#### Scenario: Busy restart clears stale confirmation state

- **WHEN** 终端 `terminal-b` 的确认状态为 `needs-ack`
- **AND** 终端 `terminal-b` 再次进入 `busy`
- **THEN** 系统清除终端 `terminal-b` 的确认状态
- **AND** 终端 `terminal-b` 显示运行中的活动状态
