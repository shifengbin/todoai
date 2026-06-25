## ADDED Requirements

### Requirement: Clear Single Global Project Candidate

系统 SHALL 允许用户从候选项目列表中清除单个全局项目候选。清除单个候选 SHALL 使用应用内自定义确认弹窗，SHALL NOT 使用系统原生确认框。清除单个候选 SHALL 只删除该候选项记录，SHALL NOT 删除磁盘目录，SHALL NOT 删除任何 workspace 中已经加入 TODO 的工程副本，SHALL NOT 关闭已有 TODO 工程终端。若被清除候选已经在当前未提交的创建 TODO、编辑 TODO 或添加工程弹窗中被选中，系统 SHALL 同步移除该临时选择。

#### Scenario: User clears one global candidate

- **WHEN** 全局候选项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户在候选项目列表中请求清除候选项目 `frontend-app`
- **AND** 系统显示应用内清除确认弹窗
- **AND** 用户在该弹窗中确认清除
- **THEN** 全局候选项目库不再包含 `frontend-app`
- **AND** 全局候选项目库仍包含 `api-service`

#### Scenario: User cancels clearing one global candidate

- **WHEN** 全局候选项目库包含 `frontend-app`
- **AND** 用户在候选项目列表中请求清除候选项目 `frontend-app`
- **AND** 系统显示应用内清除确认弹窗
- **AND** 用户在该弹窗中取消确认
- **THEN** 全局候选项目库仍包含 `frontend-app`

#### Scenario: Clearing selected candidate removes pending selection

- **WHEN** 用户在创建 TODO、编辑 TODO 或添加工程弹窗中已选择候选项目 `frontend-app`
- **AND** 用户确认清除候选项目 `frontend-app`
- **THEN** 当前弹窗的待提交项目选择不再包含 `frontend-app`
- **AND** 后续提交不会引用已清除的候选项目 ID

#### Scenario: Clearing one candidate preserves TODO project copy and terminals

- **WHEN** 全局候选项目库包含路径 `/repo/frontend-app`
- **AND** TODO `修复登录问题` 下已保存工程副本 `frontend-app`
- **AND** 该 TODO 工程已有运行中的终端
- **AND** 用户确认清除候选项目 `frontend-app`
- **THEN** 系统从候选库移除 `/repo/frontend-app`
- **AND** 系统不删除磁盘目录 `/repo/frontend-app`
- **AND** TODO `修复登录问题` 下仍显示工程 `frontend-app`
- **AND** 该 TODO 工程终端保持运行
