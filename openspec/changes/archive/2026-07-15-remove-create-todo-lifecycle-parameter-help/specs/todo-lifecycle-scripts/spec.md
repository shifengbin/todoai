## MODIFIED Requirements

### Requirement: Select Lifecycle Script Pair When Creating Todo

系统 SHALL 在创建 TODO 表单中提供脚本对下拉筛选控件。该控件 SHALL 允许用户按名称或描述筛选脚本对，SHALL 展示脚本对描述，SHALL 支持不选择脚本对。存在默认脚本对时，系统 SHALL 在打开创建 TODO 表单时默认选中它。所选脚本对包含参数定义时，系统 SHALL 在创建 TODO 表单中展示参数输入项，SHALL 使用参数默认值初始化输入值，SHALL NOT 在参数输入区域展示参数使用方法问号按钮或由该按钮触发的悬浮提示，并 SHALL 在创建 TODO 前校验必填参数。创建 TODO 时系统 SHALL 保存所选脚本对快照、参数定义快照和参数值快照，后续全局模板修改 SHALL NOT 改变已创建 TODO 的脚本内容、参数定义或参数值。

#### Scenario: Default lifecycle script pair is preselected

- **WHEN** 全局脚本管理中存在默认脚本对 `Node setup`
- **AND** 用户打开创建 TODO 表单
- **THEN** 创建 TODO 表单默认选中 `Node setup`
- **AND** 表单展示该脚本对的描述

#### Scenario: User creates todo without lifecycle script pair

- **WHEN** 用户打开创建 TODO 表单
- **AND** 用户清空脚本对选择
- **AND** 用户创建 TODO
- **THEN** 系统创建 TODO
- **AND** 该 TODO 不包含生命周期脚本快照
- **AND** 后续开始和完成该 TODO 时不执行生命周期脚本

#### Scenario: Todo form shows lifecycle script parameters

- **WHEN** 全局脚本管理中存在脚本对 `Create branch`
- **AND** 该脚本对包含参数名为 `branch_name`、显示名称为 `分支名`、默认值为 `feature/demo` 的参数
- **AND** 用户打开创建 TODO 表单并选择 `Create branch`
- **THEN** 创建 TODO 表单展示 `分支名` 参数输入项
- **AND** 参数输入项默认值为 `feature/demo`

#### Scenario: Todo form omits parameter usage help

- **WHEN** 用户在创建 TODO 表单中选择包含参数定义的生命周期脚本对
- **THEN** 创建 TODO 表单展示对应的参数输入项
- **AND** 参数区域不展示参数使用方法问号按钮
- **AND** 参数区域不提供由该按钮触发的参数使用方法悬浮提示

#### Scenario: Required lifecycle script parameter is validated

- **WHEN** 用户选择包含必填参数 `branch_name` 的生命周期脚本对
- **AND** 用户清空 `branch_name` 参数值
- **AND** 用户创建 TODO
- **THEN** 系统拒绝创建 TODO
- **AND** 系统展示必填参数校验错误

#### Scenario: Todo stores selected lifecycle script pair snapshot

- **WHEN** 用户选择脚本对 `Node setup` 创建 TODO
- **AND** 用户随后在全局脚本管理中修改 `Node setup` 的初始化脚本文本
- **THEN** 已创建 TODO 仍保留创建时选择的初始化脚本和完成脚本快照

#### Scenario: Todo stores lifecycle script parameter snapshot

- **WHEN** 用户选择脚本对 `Create branch` 创建 TODO
- **AND** 用户将参数 `branch_name` 设置为 `feature/login`
- **AND** 用户随后在全局脚本管理中修改 `Create branch` 的参数默认值为 `feature/other`
- **THEN** 已创建 TODO 仍保留创建时的 `branch_name` 参数定义
- **AND** 已创建 TODO 仍保留参数值 `feature/login`
