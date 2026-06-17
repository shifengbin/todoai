## MODIFIED Requirements

### Requirement: Collapse Todo Branches

系统 SHALL 允许用户独立展开和收起 `未执行` 与 `执行中` 视图中的 TODO 分支。收起 TODO SHALL 隐藏其项目和终端子项，但 SHALL 保留 TODO 行可见。若收起的 TODO 下存在终端，TODO 行 SHALL 反映被隐藏子终端的聚合活动状态；聚合优先级 SHALL 为 `needs-input` 高于 `needs-ack` 高于 `busy` 高于 `idle`。折叠 TODO 的非空聚合活动状态 SHALL 使用覆盖 TODO item 整行的呼吸式状态反馈表达，并 SHALL 区分 `busy`、`needs-ack` 与 `needs-input`。折叠 TODO 行 MUST NOT 复用终端行的转圈或警告活动图标来表达聚合状态。

#### Scenario: User collapses a todo

- **WHEN** TODO `修复登录问题` 已展开并显示项目子项
- **AND** 用户激活该 TODO 的收起控件
- **THEN** 该 TODO 下的项目和终端子项被隐藏
- **AND** TODO `修复登录问题` 仍显示在当前状态视图中

#### Scenario: Collapsed todo shows hidden terminal busy as row breathing

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行呼吸式状态反馈显示运行中的聚合活动状态
- **AND** TODO `修复登录问题` 行不显示终端行的转圈活动图标

#### Scenario: Collapsed todo shows hidden terminal confirmation as urgent row breathing

- **WHEN** TODO `修复登录问题` 下存在确认状态为 `needs-ack` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行急促呼吸式状态反馈显示待确认聚合活动状态
- **AND** 待确认的整行状态反馈与运行中的整行状态反馈可区分
- **AND** TODO `修复登录问题` 行不显示终端行的三角感叹号图标

#### Scenario: Collapsed todo shows hidden terminal needing input as row breathing

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行呼吸式状态反馈显示等待输入的聚合活动状态
- **AND** 等待输入的整行状态反馈与运行中的整行状态反馈可区分
- **AND** TODO `修复登录问题` 行不显示终端行的警告活动图标

#### Scenario: Collapsed todo prioritizes needs input over confirmation and busy

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 下存在确认状态为 `needs-ack` 的终端
- **AND** TODO `修复登录问题` 下还存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的整行呼吸式状态反馈
- **AND** TODO `修复登录问题` 行不以待确认或运行中状态作为最高优先级提示

#### Scenario: Collapsed todo prioritizes confirmation over busy

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 下还存在确认状态为 `needs-ack` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示待确认的整行急促呼吸式状态反馈
- **AND** TODO `修复登录问题` 行不以运行中状态作为最高优先级提示

#### Scenario: Expanded todo relies on terminal rows for activity state

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已展开
- **THEN** 该终端行显示运行中的活动提示
- **AND** TODO `修复登录问题` 行不重复显示收起态聚合活动提示

#### Scenario: Expanded todo shows confirmation on terminal row

- **WHEN** TODO `修复登录问题` 下存在确认状态为 `needs-ack` 的终端
- **AND** TODO `修复登录问题` 已展开
- **THEN** 该终端行显示三角感叹号确认态提示
- **AND** TODO `修复登录问题` 行不重复显示收起态聚合活动提示

#### Scenario: Collapsed todo without active hidden terminal has no breathing feedback

- **WHEN** TODO `修复登录问题` 已收起
- **AND** TODO `修复登录问题` 下不存在活动状态为 `busy` 或 `needs-input` 的终端
- **AND** TODO `修复登录问题` 下不存在确认状态为 `needs-ack` 的终端
- **THEN** TODO `修复登录问题` 行不显示整行呼吸式状态反馈
- **AND** TODO `修复登录问题` 行不显示终端活动图标

#### Scenario: User expands a todo

- **WHEN** TODO `修复登录问题` 已收起
- **AND** 用户激活该 TODO 的展开控件
- **THEN** 系统显示该 TODO 下的项目子项
