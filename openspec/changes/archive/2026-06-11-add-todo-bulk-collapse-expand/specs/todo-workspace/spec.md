## ADDED Requirements

### Requirement: Expand Or Collapse All Todo Branches

系统 SHALL 允许用户在活动 TODO 列表中一键展开或收起所有 TODO 分支。批量收起 SHALL 隐藏所有活动 TODO 下的项目和终端子项，但 SHALL 保留 TODO 行可见。批量展开 SHALL 显示所有活动 TODO 下的项目子项。

#### Scenario: User collapses all active todo branches

- **WHEN** 活动 TODO 列表中存在多个已展开 TODO
- **AND** 用户激活全部收起控件
- **THEN** 所有活动 TODO 下的项目和终端子项被隐藏
- **AND** 所有活动 TODO 行仍显示在活动任务列表中

#### Scenario: User expands all active todo branches

- **WHEN** 活动 TODO 列表中存在多个已收起 TODO
- **AND** 用户激活全部展开控件
- **THEN** 所有活动 TODO 下的项目子项被显示

#### Scenario: Active context expands after bulk collapse

- **WHEN** 用户已批量收起所有活动 TODO 分支
- **AND** 当前 TODO、TODO 项目或终端上下文切换到某个已收起 TODO 下
- **THEN** 系统自动展开该 TODO 分支
- **AND** 其他已收起 TODO 分支保持收起
