## MODIFIED Requirements

### Requirement: Collapse Todo Branches

系统 SHALL 允许用户独立展开和收起 TODO 分支。收起 TODO SHALL 隐藏其项目和终端子项，但 SHALL 保留 TODO 行可见。若收起的 TODO 下存在终端，TODO 行 SHALL 反映被隐藏子终端的聚合活动状态；聚合优先级 SHALL 为 `needs-input` 高于 `busy` 高于 `idle`。

#### Scenario: User collapses a todo

- **WHEN** TODO `修复登录问题` 已展开并显示项目子项
- **AND** 用户激活该 TODO 的收起控件
- **THEN** 该 TODO 下的项目和终端子项被隐藏
- **AND** TODO `修复登录问题` 仍显示在活动任务列表中

#### Scenario: Collapsed todo shows hidden terminal needing input

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的终端活动提示

#### Scenario: Collapsed todo prioritizes needs input over busy

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 下还存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的终端活动提示
- **AND** TODO `修复登录问题` 行不以运行中状态作为最高优先级提示

#### Scenario: Expanded todo relies on terminal rows for activity state

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已展开
- **THEN** 该终端行显示运行中的活动提示
- **AND** TODO `修复登录问题` 行不重复显示收起态聚合活动提示

#### Scenario: User expands a todo

- **WHEN** TODO `修复登录问题` 已收起
- **AND** 用户激活该 TODO 的展开控件
- **THEN** 系统显示该 TODO 下的项目子项
