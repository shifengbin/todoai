## ADDED Requirements

### Requirement: Default Select Single Imported Candidate In Active Picker

系统 SHALL 在用户通过单个目录导入全局项目候选后，将该导入目录对应的候选默认加入当前打开的项目选择控件。若该候选已经存在，系统 SHALL 选中已有候选。该默认选择 SHALL 只影响当前控件的临时选择状态，SHALL NOT 立即创建 TODO 工程副本，SHALL NOT 改变当前 TODO project 上下文，且 SHALL NOT 创建或激活终端。

#### Scenario: Single imported candidate is selected in create todo form

- **WHEN** 用户打开创建 TODO 表单
- **AND** 用户通过单个导入选择目录 `/repo/frontend-app`
- **THEN** 全局项目候选包含 `/repo/frontend-app`
- **AND** 创建 TODO 表单的项目选择中默认选中 `/repo/frontend-app`
- **AND** 当前 workspace 尚未创建新的 TODO 工程副本
- **AND** 当前 TODO project 上下文保持不变

#### Scenario: Existing single imported candidate is selected in create todo form

- **WHEN** 用户打开创建 TODO 表单
- **AND** 全局项目候选已包含 `/repo/frontend-app`
- **AND** 用户通过单个导入再次选择目录 `/repo/frontend-app`
- **THEN** 创建 TODO 表单的项目选择中默认选中已有候选 `/repo/frontend-app`
- **AND** 当前 workspace 尚未创建新的 TODO 工程副本
- **AND** 当前 TODO project 上下文保持不变

#### Scenario: Single imported candidate does not select without active picker

- **WHEN** 用户没有打开创建 TODO、编辑 TODO 或添加项目控件
- **AND** 用户通过单个导入选择目录 `/repo/frontend-app`
- **THEN** 全局项目候选包含 `/repo/frontend-app`
- **AND** 系统不创建 TODO 工程副本
- **AND** 当前 TODO project 上下文保持不变
- **AND** 当前终端保持不变
