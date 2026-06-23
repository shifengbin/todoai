## ADDED Requirements

### Requirement: Default Select Imported Candidate In Todo Project Pickers

系统 SHALL 在创建 TODO、编辑 TODO 或为 TODO 添加项目的项目选择控件中，默认选中刚通过单个目录导入创建的全局项目候选。默认选中 SHALL 可由用户取消，且 SHALL 只有在用户提交当前控件后才关联到 TODO。

#### Scenario: Imported candidate is selected while creating todo

- **WHEN** 用户打开创建 TODO 表单
- **AND** 用户单个导入项目候选 `frontend-app`
- **THEN** 创建 TODO 表单显示 `frontend-app` 为已选项目
- **WHEN** 用户提交创建 TODO 表单
- **THEN** 新 TODO 下保存 `frontend-app` 的 TODO 工程副本

#### Scenario: Imported candidate is selected while editing todo

- **WHEN** 用户打开 TODO `修复登录问题` 的编辑表单
- **AND** 用户单个导入项目候选 `frontend-app`
- **THEN** 编辑表单显示 `frontend-app` 为已选项目
- **WHEN** 用户保存编辑表单
- **THEN** TODO `修复登录问题` 下保存 `frontend-app` 的 TODO 工程副本

#### Scenario: Imported candidate is selected while adding projects to todo

- **WHEN** 用户为 TODO `修复登录问题` 打开添加项目控件
- **AND** 用户单个导入项目候选 `frontend-app`
- **THEN** 添加项目控件显示 `frontend-app` 为已选项目
- **WHEN** 用户确认添加
- **THEN** TODO `修复登录问题` 下保存 `frontend-app` 的 TODO 工程副本

#### Scenario: User cancels imported candidate selection

- **WHEN** 用户打开创建 TODO 表单
- **AND** 用户单个导入项目候选 `frontend-app`
- **AND** 创建 TODO 表单显示 `frontend-app` 为已选项目
- **WHEN** 用户从已选项目中移除 `frontend-app`
- **AND** 用户提交创建 TODO 表单
- **THEN** 新 TODO 不包含 `frontend-app` 的 TODO 工程副本

### Requirement: Toggle Todo Branch By Double Clicking Header

系统 SHALL 允许用户通过双击 `未执行` 与 `执行中` 视图中的 TODO header 行来展开或收起 TODO 分支。双击 TODO header 行 SHALL 与展开按钮使用相同的折叠状态规则。双击 header 行内操作按钮、菜单或确认气泡 SHALL NOT 触发展开或收起。

#### Scenario: User double clicks expanded todo header

- **WHEN** TODO `修复登录问题` 已展开
- **AND** 用户双击该 TODO 的 header 行
- **THEN** TODO `修复登录问题` 被收起
- **AND** 该 TODO 下的项目和终端子项被隐藏

#### Scenario: User double clicks collapsed todo header

- **WHEN** TODO `修复登录问题` 已收起
- **AND** 用户双击该 TODO 的 header 行
- **THEN** TODO `修复登录问题` 被展开
- **AND** 系统显示该 TODO 下的项目子项

#### Scenario: Double clicking todo action does not toggle branch

- **WHEN** TODO `修复登录问题` 已展开
- **AND** 用户双击该 TODO header 行内的菜单按钮或完成按钮
- **THEN** TODO `修复登录问题` 保持展开
- **AND** 被双击的按钮保持其原有交互行为
