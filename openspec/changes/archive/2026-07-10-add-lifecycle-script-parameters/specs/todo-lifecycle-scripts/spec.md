## MODIFIED Requirements

### Requirement: Manage Global Todo Lifecycle Script Pairs

系统 SHALL 允许用户通过全局管理菜单维护 TODO 生命周期脚本对模板。每个模板记录 SHALL 包含显示名称、描述、初始化脚本文本、完成脚本文本、参数定义列表和是否默认选择。每个参数定义 SHALL 包含参数名、显示名称、描述、默认值和是否必填。系统 SHALL 在脚本管理列表中默认以单行摘要展示脚本模板，并允许用户点击后展开为多行编辑态。系统 SHALL 允许脚本编辑态中的参数区域独立折叠和展开。系统 SHALL 在参数区域提供使用方法和注意事项帮助提示，说明脚本中使用 `{{参数名}}` 引用参数、不要额外加引号、参数会按当前 shell 自动转义、未定义占位符保持原样且参数明文保存；该帮助提示 SHALL NOT 被脚本管理弹窗或滚动容器裁剪。系统 SHALL 持久化这些模板，并在应用重启或 workspace 切换后继续可用。系统 SHALL 允许不同模板记录使用相同显示名称。系统 SHALL 拒绝保存空显示名称、初始化脚本和完成脚本均为空的模板记录、多个默认选择模板记录、空参数名、同一模板内重复参数名，以及不匹配 `[A-Za-z_][A-Za-z0-9_]*` 的参数名。

#### Scenario: User saves lifecycle script pair templates

- **WHEN** 用户打开全局管理菜单中的脚本管理
- **AND** 用户新增名称为 `Node setup`、描述为 `安装依赖并清理缓存`、包含初始化脚本和完成脚本的模板
- **AND** 用户保存脚本管理表单
- **THEN** 系统持久化该脚本对模板
- **AND** 用户切换 workspace 或重启应用后仍能看到该脚本对模板

#### Scenario: User saves lifecycle script parameters

- **WHEN** 用户打开全局管理菜单中的脚本管理
- **AND** 用户为脚本对 `Create branch` 添加参数名为 `branch_name`、显示名称为 `分支名`、默认值为 `feature/demo`、必填的参数
- **AND** 用户保存脚本管理表单
- **THEN** 系统持久化该脚本对模板的参数定义
- **AND** 用户切换 workspace 或重启应用后仍能看到该参数定义

#### Scenario: Script management shows parameter usage help

- **WHEN** 用户打开全局管理菜单中的脚本管理
- **AND** 用户查看脚本参数区域的帮助提示
- **THEN** 系统展示 `{{参数名}}` 和 `git checkout -b {{branch_name}}` 示例
- **AND** 系统提示不要额外加引号、参数会按当前 shell 自动转义、未定义占位符保持原样且参数明文保存

#### Scenario: Script management rows expand for editing

- **WHEN** 用户打开全局管理菜单中的脚本管理
- **THEN** 系统以单行摘要展示已有脚本对模板
- **AND** 系统不展示完整脚本文本编辑框
- **WHEN** 用户展开某个脚本对模板
- **THEN** 系统展示该脚本对的显示名称、描述、初始化脚本、完成脚本和参数区域编辑控件

#### Scenario: Script parameter editor collapses independently

- **WHEN** 用户展开脚本管理中的某个脚本对模板
- **THEN** 系统默认收起该脚本的参数明细
- **AND** 系统展示参数数量、帮助按钮和添加参数按钮
- **WHEN** 用户展开参数区域或添加参数
- **THEN** 系统展示参数明细编辑控件

#### Scenario: Invalid lifecycle script pair is rejected

- **WHEN** 用户在脚本管理中保存一个空显示名称的脚本对模板
- **THEN** 系统拒绝保存
- **AND** 系统展示校验错误
- **AND** 原有脚本对模板保持不变

#### Scenario: Invalid lifecycle script parameter is rejected

- **WHEN** 用户在脚本管理中为脚本对保存参数名为 `branch-name` 的参数
- **THEN** 系统拒绝保存
- **AND** 系统展示参数名校验错误
- **AND** 原有脚本对模板保持不变

#### Scenario: Duplicate lifecycle script parameter is rejected

- **WHEN** 用户在脚本管理中为同一个脚本对保存两个参数名均为 `branch_name` 的参数
- **THEN** 系统拒绝保存
- **AND** 系统展示参数名重复错误
- **AND** 原有脚本对模板保持不变

#### Scenario: Multiple default lifecycle script pairs are rejected

- **WHEN** 用户在脚本管理中将两个脚本对模板同时标记为默认选择
- **AND** 用户保存脚本管理表单
- **THEN** 系统拒绝保存
- **AND** 系统提示只能有一个默认脚本对
- **AND** 原有脚本对模板保持不变

### Requirement: Select Lifecycle Script Pair When Creating Todo

系统 SHALL 在创建 TODO 表单中提供脚本对下拉筛选控件。该控件 SHALL 允许用户按名称或描述筛选脚本对，SHALL 展示脚本对描述，SHALL 支持不选择脚本对。存在默认脚本对时，系统 SHALL 在打开创建 TODO 表单时默认选中它。所选脚本对包含参数定义时，系统 SHALL 在创建 TODO 表单中展示参数输入项，SHALL 使用参数默认值初始化输入值，并 SHALL 在参数输入区域提供使用方法和注意事项帮助提示，SHALL 在创建 TODO 前校验必填参数。创建 TODO 时系统 SHALL 保存所选脚本对快照、参数定义快照和参数值快照，后续全局模板修改 SHALL NOT 改变已创建 TODO 的脚本内容、参数定义或参数值。

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

#### Scenario: Todo form shows parameter usage help

- **WHEN** 用户在创建 TODO 表单中选择包含参数定义的生命周期脚本对
- **AND** 用户查看参数输入区域的帮助提示
- **THEN** 系统展示脚本中使用 `{{参数名}}` 引用参数的说明
- **AND** 系统展示不要额外加引号、自动转义、未定义占位符保持原样和明文保存的注意事项

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

### Requirement: Run Lifecycle Scripts Cross Platform Without Blocking UI

系统 SHALL 在 Windows、macOS 和 Linux 上支持生命周期脚本后台执行。系统 SHALL 使用当前操作系统可用的 shell 执行用户提供的脚本文本，SHALL 将进程工作目录设置为对应 TODO 工作目录，SHALL 在执行前将已声明参数对应的 `{{param_name}}` 占位符渲染为当前 shell 可安全执行的字符串字面量，SHALL NOT 要求用户使用 shell 专属环境变量语法读取参数，SHALL NOT 尝试将脚本文本转换为其他操作系统语法。未声明参数对应的占位符文本 SHALL 保持原样。Windows 后台脚本执行 SHALL NOT 弹出系统控制台窗口。脚本执行状态 SHALL 通过应用状态或事件返回前端，且 SHALL NOT 阻塞用户继续操作 UI。

#### Scenario: Unix lifecycle script runs in todo directory

- **WHEN** 应用运行在 macOS 或 Linux
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统使用当前可用 Unix shell 后台执行脚本文本
- **AND** 脚本进程工作目录为 TODO `fix-login` 的工作目录
- **AND** UI 在脚本运行期间仍可响应用户操作

#### Scenario: Windows lifecycle script runs without console window

- **WHEN** 应用运行在 Windows
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统使用当前可用 Windows shell 后台执行脚本文本
- **AND** 脚本进程工作目录为 TODO `fix-login` 的工作目录
- **AND** 系统不显示额外的系统控制台窗口
- **AND** UI 在脚本运行期间仍可响应用户操作

#### Scenario: Lifecycle script parameter placeholder is rendered for current shell

- **WHEN** TODO `fix-login` 的生命周期脚本包含 `git checkout -b {{branch_name}}`
- **AND** TODO `fix-login` 的参数 `branch_name` 值为 `feature/login`
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统在执行前将 `{{branch_name}}` 渲染为当前 shell 的字符串字面量
- **AND** 用户无需在脚本中使用 `$branch_name`、`$env:branch_name` 或 `%branch_name%`

#### Scenario: Lifecycle script parameter values are escaped

- **WHEN** TODO `fix-login` 的生命周期脚本包含 `echo {{message}}`
- **AND** TODO `fix-login` 的参数 `message` 值包含空格、引号或 shell 特殊字符
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统在执行前将 `{{message}}` 渲染为当前 shell 的单个字符串字面量
- **AND** 参数值不会被拆分为多个参数或解释为额外命令

#### Scenario: Unknown placeholder is left unchanged

- **WHEN** TODO `fix-login` 的生命周期脚本包含 `echo {{unknown_value}}`
- **AND** TODO `fix-login` 没有声明参数 `unknown_value`
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统不渲染 `{{unknown_value}}`
- **AND** 脚本文本中的 `{{unknown_value}}` 保持原样传给 shell

#### Scenario: Unsupported shell reports script failure

- **WHEN** 当前操作系统没有可用 shell 执行生命周期脚本
- **AND** TODO `fix-login` 的生命周期脚本被触发
- **THEN** 系统记录该阶段脚本失败状态
- **AND** UI 展示失败原因
- **AND** 用户可以在修复 shell 配置后重新触发脚本执行
