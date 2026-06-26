## Purpose

定义 TODO 生命周期脚本模板、选择、执行、状态展示和重试行为，使 TODO 开始与完成流程可以按项目工作目录异步运行用户配置的脚本。

## Requirements

### Requirement: Manage Global Todo Lifecycle Script Pairs

系统 SHALL 允许用户通过全局管理菜单维护 TODO 生命周期脚本对模板。每个模板记录 SHALL 包含显示名称、描述、初始化脚本文本、完成脚本文本和是否默认选择。系统 SHALL 持久化这些模板，并在应用重启或 workspace 切换后继续可用。系统 SHALL 允许不同模板记录使用相同显示名称。系统 SHALL 拒绝保存空显示名称、初始化脚本和完成脚本均为空的模板记录，以及多个默认选择模板记录。

#### Scenario: User saves lifecycle script pair templates

- **WHEN** 用户打开全局管理菜单中的脚本管理
- **AND** 用户新增名称为 `Node setup`、描述为 `安装依赖并清理缓存`、包含初始化脚本和完成脚本的模板
- **AND** 用户保存脚本管理表单
- **THEN** 系统持久化该脚本对模板
- **AND** 用户切换 workspace 或重启应用后仍能看到该脚本对模板

#### Scenario: Invalid lifecycle script pair is rejected

- **WHEN** 用户在脚本管理中保存一个空显示名称的脚本对模板
- **THEN** 系统拒绝保存
- **AND** 系统展示校验错误
- **AND** 原有脚本对模板保持不变

#### Scenario: Multiple default lifecycle script pairs are rejected

- **WHEN** 用户在脚本管理中将两个脚本对模板同时标记为默认选择
- **AND** 用户保存脚本管理表单
- **THEN** 系统拒绝保存
- **AND** 系统提示只能有一个默认脚本对
- **AND** 原有脚本对模板保持不变

### Requirement: Select Lifecycle Script Pair When Creating Todo

系统 SHALL 在创建 TODO 表单中提供脚本对下拉筛选控件。该控件 SHALL 允许用户按名称或描述筛选脚本对，SHALL 展示脚本对描述，SHALL 支持不选择脚本对。存在默认脚本对时，系统 SHALL 在打开创建 TODO 表单时默认选中它。创建 TODO 时系统 SHALL 保存所选脚本对快照，后续全局模板修改 SHALL NOT 改变已创建 TODO 的脚本内容。

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

#### Scenario: Todo stores selected lifecycle script pair snapshot

- **WHEN** 用户选择脚本对 `Node setup` 创建 TODO
- **AND** 用户随后在全局脚本管理中修改 `Node setup` 的初始化脚本文本
- **THEN** 已创建 TODO 仍保留创建时选择的初始化脚本和完成脚本快照

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

系统 SHALL 在 Windows、macOS 和 Linux 上支持生命周期脚本后台执行。系统 SHALL 使用当前操作系统可用的 shell 执行用户提供的脚本文本，SHALL 将进程工作目录设置为对应 TODO 工作目录，SHALL NOT 尝试将脚本文本转换为其他操作系统语法。Windows 后台脚本执行 SHALL NOT 弹出系统控制台窗口。脚本执行状态 SHALL 通过应用状态或事件返回前端，且 SHALL NOT 阻塞用户继续操作 UI。

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
