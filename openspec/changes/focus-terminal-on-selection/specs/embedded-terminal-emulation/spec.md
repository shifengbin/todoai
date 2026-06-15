## ADDED Requirements

### Requirement: Focus Terminal After Tree Selection

系统 SHALL 在用户从左侧 TODO 终端树选择终端并成功激活对应右侧嵌入式终端后，将键盘输入焦点转移到该嵌入式终端。该自动聚焦 MUST 只由用户明确选择终端的交互触发，后台状态更新、初始化恢复或其他非选择终端路径不得因此抢占焦点。

#### Scenario: User selects a terminal from todo tree

- **WHEN** 用户在左侧 TODO 终端树中点击终端 B
- **AND** 系统成功将活动终端切换为终端 B
- **THEN** 右侧终端区域显示终端 B
- **AND** 终端 B 的嵌入式终端获得键盘输入焦点

#### Scenario: Terminal selection fails

- **WHEN** 用户在左侧 TODO 终端树中点击终端 B
- **AND** 系统未能成功将活动终端切换为终端 B
- **THEN** 系统不得将键盘输入焦点转移到终端 B
