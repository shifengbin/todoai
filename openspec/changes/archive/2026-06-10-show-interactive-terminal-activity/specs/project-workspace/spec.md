## ADDED Requirements

### Requirement: Display Interactive Terminal Activity In Project Tree
系统 SHALL 在左侧项目终端树中展示交互式终端程序的运行时活动状态。该状态 SHALL 基于终端标题变化推导，并 SHALL 不替代现有 shell 命令标签。

#### Scenario: Interactive terminal records an idle launch title
- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 刚启动并收到静态运行时标题 `codex - alpha`
- **THEN** terminal A 的项目树终端行不显示执行中的动态指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Interactive terminal is busy after launch
- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 已记录空闲运行时标题 `codex - alpha`
- **AND** terminal A 后续收到不同于空闲标题的执行中运行时标题
- **THEN** terminal A 的项目树终端行显示执行中的动态指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Interactive terminal needs user input
- **WHEN** terminal A 的当前命令标签是 `codex`
- **AND** terminal A 收到表示需要注意的运行时标题，例如包含 `!`
- **THEN** terminal A 的项目树终端行显示需要用户输入的注意指示
- **AND** terminal A 不显示执行中的动态指示

#### Scenario: Interactive terminal returns to idle title
- **WHEN** terminal A 的项目树终端行正在显示执行中或需要输入状态
- **AND** terminal A 的运行时标题恢复为当前命令标签 `codex` 或已记录的空闲标题
- **THEN** terminal A 的项目树终端行停止活动指示
- **AND** terminal A 的主标签仍显示 `codex`

#### Scenario: Shell command state still controls command label
- **WHEN** terminal A 的 shell command-start 事件报告当前命令为 `codex`
- **AND** terminal A 后续收到运行时标题变化
- **THEN** terminal A 的主标签由 shell 命令状态保持为 `codex`
- **AND** 运行时标题只影响活动指示状态

#### Scenario: Shell command exits
- **WHEN** terminal A 正在显示交互式活动状态
- **AND** terminal A 的 shell command-end 事件到达
- **THEN** terminal A 清除交互式活动指示
- **AND** terminal A 的主标签恢复为 shell 名称
