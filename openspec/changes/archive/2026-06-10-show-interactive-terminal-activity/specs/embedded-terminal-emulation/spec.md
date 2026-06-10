## ADDED Requirements

### Requirement: Capture Terminal Title Changes
系统 SHALL 捕获嵌入式 xterm 会话收到的终端标题变化，并 SHALL 将标题变化关联到产生该变化的 terminal。

#### Scenario: Interactive program updates terminal title
- **WHEN** terminal A 中运行的交互式程序发送 OSC 0 或 OSC 2 标题更新
- **THEN** 系统记录 terminal A 的最新运行时标题
- **AND** 该标题更新不会被关联到其他 terminal

#### Scenario: Inactive terminal updates title
- **WHEN** terminal A 处于后台
- **AND** terminal B 是当前激活终端
- **AND** terminal A 中运行的交互式程序发送标题更新
- **THEN** 系统更新 terminal A 的运行时标题状态
- **AND** terminal B 的运行时标题状态保持不变

#### Scenario: Program does not emit title changes
- **WHEN** terminal A 中运行的程序不发送终端标题更新
- **THEN** 系统保持 terminal A 的现有终端渲染和命令标签行为
- **AND** 不为 terminal A 生成交互式活动状态
