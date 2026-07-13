## Purpose

定义 TODO 生命周期脚本模板、选择、执行、状态展示和重试行为，使 TODO 开始与完成流程可以按项目工作目录异步运行用户配置的脚本。
## Requirements
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

### Requirement: Execute Initialization Script Asynchronously When Starting Todo

系统 SHALL 在 TODO 从 `not-started` 进入 `in-progress` 后，在该 TODO 工作目录中异步执行所选脚本对的初始化脚本。初始化脚本 SHALL NOT 阻止 TODO 进入 `in-progress`。初始化脚本运行期间系统 SHALL 展示运行中状态；初始化脚本成功后系统 SHALL 清除该阶段状态且 SHALL NOT 展示成功状态；初始化脚本失败时系统 SHALL 展示失败状态并提供重新执行入口。

#### Scenario: Starting todo immediately enters in-progress while init script runs

- **WHEN** TODO `fix-login` 选择了包含初始化脚本的脚本对
- **AND** 用户点击开始 TODO
- **THEN** TODO `fix-login` 立即变为 `in-progress`
- **AND** 系统在 TODO 工作目录中后台执行初始化脚本
- **AND** UI 展示初始化脚本运行中状态

#### Scenario: Successful initialization script clears visible status

- **WHEN** TODO `fix-login` 的初始化脚本正在运行
- **AND** 初始化脚本以成功退出码结束
- **THEN** 系统清除 TODO `fix-login` 的初始化脚本状态
- **AND** UI 不展示初始化成功状态

#### Scenario: Failed initialization script can be retried

- **WHEN** TODO `fix-login` 的初始化脚本执行失败
- **THEN** TODO `fix-login` 保持 `in-progress`
- **AND** UI 展示初始化脚本失败状态
- **AND** UI 提供重新执行初始化脚本的操作
- **WHEN** 用户重新执行初始化脚本
- **THEN** 系统在同一个 TODO 工作目录中再次后台执行初始化脚本

#### Scenario: Duplicate initialization run is prevented

- **WHEN** TODO `fix-login` 的初始化脚本正在运行
- **AND** 用户再次触发初始化脚本执行
- **THEN** 系统不启动第二个初始化脚本进程
- **AND** UI 保持当前运行中状态

### Requirement: Execute Completion Script Before Completing Todo

系统 SHALL 在用户完成包含完成脚本快照的 `in-progress` TODO 时异步执行完成脚本，并在完成脚本成功后执行现有 TODO 完成归档流程。完成脚本运行期间 TODO SHALL 保持 `in-progress`，UI SHALL 展示完成脚本运行中状态。完成脚本成功后系统 SHALL 清除该阶段状态且 SHALL NOT 展示成功状态。完成脚本失败时 TODO SHALL 保持 `in-progress`，UI SHALL 展示失败状态并提供重新执行入口。未选择完成脚本的 TODO SHALL 沿用现有完成行为。

#### Scenario: Completion script runs before todo is archived

- **WHEN** TODO `fix-login` 处于 `in-progress`
- **AND** TODO `fix-login` 选择了包含完成脚本的脚本对
- **AND** 用户点击完成 TODO
- **THEN** 系统在 TODO 工作目录中后台执行完成脚本
- **AND** TODO `fix-login` 保持 `in-progress`
- **AND** UI 展示完成脚本运行中状态

#### Scenario: Successful completion script completes todo

- **WHEN** TODO `fix-login` 的完成脚本正在运行
- **AND** 完成脚本以成功退出码结束
- **THEN** 系统完成并归档 TODO `fix-login`
- **AND** 系统清除完成脚本状态
- **AND** UI 不展示完成脚本成功状态

#### Scenario: Failed completion script keeps todo retryable

- **WHEN** TODO `fix-login` 的完成脚本执行失败
- **THEN** TODO `fix-login` 保持 `in-progress`
- **AND** UI 展示完成脚本失败状态
- **AND** UI 提供重新执行完成脚本的操作
- **WHEN** 用户重新执行完成脚本
- **THEN** 系统在同一个 TODO 工作目录中再次后台执行完成脚本
- **AND** 完成脚本成功后系统完成并归档 TODO `fix-login`

#### Scenario: Todo without completion script uses existing completion flow

- **WHEN** TODO `fix-login` 处于 `in-progress`
- **AND** TODO `fix-login` 没有完成脚本快照
- **AND** 用户点击完成 TODO
- **THEN** 系统立即执行现有 TODO 完成归档流程
- **AND** UI 不展示生命周期脚本状态

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

### Requirement: Surface Only Actionable Lifecycle Script Status

系统 SHALL 只在脚本运行中或失败时展示生命周期脚本状态。系统 SHALL 在脚本成功后隐藏状态。失败状态 SHALL 包含阶段、脚本名称和截断后的错误摘要，并 SHALL 保留到用户重试成功、TODO 被删除或 TODO 成功完成归档。

#### Scenario: Running status is visible

- **WHEN** TODO `fix-login` 的初始化脚本或完成脚本正在运行
- **THEN** UI 展示对应阶段的运行中状态
- **AND** UI 禁用该阶段的重复触发操作

#### Scenario: Success status is hidden

- **WHEN** TODO `fix-login` 的生命周期脚本成功结束
- **THEN** UI 不展示该脚本的成功状态
- **AND** TODO 行和详情中不保留成功徽标

#### Scenario: Failed status remains until resolved

- **WHEN** TODO `fix-login` 的生命周期脚本执行失败
- **THEN** UI 展示失败状态和错误摘要
- **AND** 用户切换 TODO 视图后仍能看到失败状态
- **WHEN** 用户重新执行该脚本且执行成功
- **THEN** 系统清除失败状态

### Requirement: Inspect And Copy Complete Lifecycle Script Failure Output

系统 SHALL 在开始 TODO 触发的初始化脚本或完成 TODO 触发的完成脚本执行失败时，于当前应用运行期间保留该次执行未经截断的 stdout/stderr 完整合并输出。系统 SHALL 继续在普通失败状态中展示截断后的错误摘要，并 SHALL 仅在用户请求查看或复制时按需读取完整输出。失败状态 SHALL 在 Retry 操作旁提供复制错误日志按钮，复制内容 SHALL 与该次执行的完整原始合并输出完全一致且 SHALL NOT 追加阶段、脚本名称或其他格式化文本；当进程没有产生输出时，系统 SHALL 回退使用执行错误信息。系统 SHALL 在用户悬浮失败状态时以保留换行、尺寸受限且可滚动的悬浮详情展示同一份完整输出。系统 SHALL 在该阶段重新执行、执行成功、TODO 被删除、TODO 成功完成归档或 workspace 切换时清除已经失效的完整输出，且 SHALL NOT 跨应用重启持久化该输出。

#### Scenario: Starting todo exposes complete initialization failure output

- **WHEN** 用户点击开始 TODO 并触发初始化脚本
- **AND** 初始化脚本执行失败且产生超过 4096 字节的多行合并输出
- **THEN** 失败状态继续展示截断后的单行错误摘要
- **AND** 用户悬浮失败状态后可以在可滚动详情中查看未经截断的完整多行输出
- **AND** 详情包含该次输出的开头、结尾和原始换行

#### Scenario: Completing todo allows exact failure output to be copied

- **WHEN** 用户点击完成 TODO 并触发完成脚本
- **AND** 完成脚本执行失败并产生原始合并输出
- **THEN** 系统在 Retry 操作旁展示复制错误日志按钮
- **WHEN** 用户点击复制错误日志按钮
- **THEN** 系统将该次完成脚本的完整原始合并输出写入剪贴板
- **AND** 复制内容不包含系统追加的阶段、脚本名称或其他说明文本

#### Scenario: Failure without process output copies error message

- **WHEN** 初始化脚本或完成脚本在产生进程输出前执行失败
- **AND** 当前失败没有 stdout/stderr 合并输出
- **WHEN** 用户查看或复制该失败的错误日志
- **THEN** 系统展示或复制该次执行错误信息

#### Scenario: Retrying replaces obsolete failure output

- **WHEN** 生命周期脚本失败状态具有完整错误输出
- **AND** 用户点击 Retry 重新执行同一阶段脚本
- **THEN** 系统立即使上一次失败的完整输出失效
- **AND** 运行中状态不提供上一次失败日志的查看或复制入口
- **WHEN** 本次重新执行再次失败
- **THEN** 系统只提供本次失败的完整原始输出

#### Scenario: Runtime cleanup removes complete failure output

- **WHEN** 生命周期脚本失败状态具有完整错误输出
- **AND** 对应脚本随后执行成功、TODO 被删除、TODO 成功完成归档或用户切换 workspace
- **THEN** 系统清除该失败状态对应的完整输出
- **AND** 后续请求不能读取已经失效的日志

#### Scenario: Complete failure output is not persisted

- **WHEN** 生命周期脚本执行失败并保留了完整原始输出
- **AND** 用户关闭并重新启动应用
- **THEN** 系统不从持久化数据恢复该次完整输出

