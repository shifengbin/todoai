# todo-cli Specification

## Purpose
Define TodoAI command-line behavior for querying TODO data from project directories.
## Requirements
### Requirement: List Completed Todos From Project Directory

系统 SHALL 提供 `todoai list --done` 命令，用于从命令行列出当前项目相关的已完成 TODO。该命令 MUST 可在已登记项目根目录或其任意子目录执行。系统 SHALL 根据当前工作目录定位 TodoAI 已知项目，并 SHALL 只返回与该项目匹配且状态为 `completed` 的 TODO。成功时，stdout SHALL 输出 JSON 数组；每个数组元素 SHALL 包含 `taskName`、`worktreeBranch`、`baseBranch`。若历史项目快照缺少 worktree 分支或 base 分支，系统 SHALL 保留该 TODO 并用 `-` 显示缺失字段。该命令 SHALL 在执行时不启动 Wails GUI。

#### Scenario: List completed todos from project root

- **WHEN** 当前工作目录为已登记项目 `frontend-app` 的根目录
- **AND** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的完成时项目快照匹配项目 `frontend-app`
- **AND** 该项目快照的 worktree 分支为 `todo/fix-login`
- **AND** 该项目快照的 base 分支为 `main`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout 为 JSON 数组
- **AND** JSON 数组包含 `taskName` 为 `修复登录问题` 的元素
- **AND** 该元素的 `worktreeBranch` 为 `todo/fix-login`
- **AND** 该元素的 `baseBranch` 为 `main`
- **AND** 系统不启动 Wails GUI

#### Scenario: List completed todos from project child directory

- **WHEN** 当前工作目录为已登记项目 `frontend-app` 下的子目录 `src/components`
- **AND** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的完成时项目快照匹配项目 `frontend-app`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `修复登录问题` 的元素

#### Scenario: List completed todos from git worktree child directory

- **WHEN** 当前工作目录为项目 `frontend-app` 的 Git linked worktree 子目录 `build/bin`
- **AND** TodoAI 已知项目 `frontend-app` 记录的是该 linked worktree 的源仓库路径
- **AND** TODO `worktree 子目录任务` 的状态为 `completed`
- **AND** TODO `worktree 子目录任务` 的完成时项目快照匹配项目 `frontend-app`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `worktree 子目录任务` 的元素

#### Scenario: Open todos are excluded

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** TODO `待执行任务` 的状态为 `not-started`
- **AND** TODO `执行中任务` 的状态为 `in-progress`
- **AND** TODO `已完成任务` 的状态为 `completed`
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `已完成任务` 的元素
- **AND** stdout JSON 数组不包含 `taskName` 为 `待执行任务` 的元素
- **AND** stdout JSON 数组不包含 `taskName` 为 `执行中任务` 的元素

#### Scenario: Missing branch fields use placeholders

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** TODO `旧任务` 的状态为 `completed`
- **AND** TODO `旧任务` 的完成时项目快照匹配项目 `frontend-app`
- **AND** 该项目快照缺少 worktree 分支或 base 分支
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout JSON 数组包含 `taskName` 为 `旧任务` 的元素
- **AND** 该元素缺失的分支字段值为 `-`

#### Scenario: Unknown project directory returns error

- **WHEN** 当前工作目录不属于任何 TodoAI 已知项目或其子目录
- **THEN** 执行 `todoai list --done` 返回失败
- **AND** 输出说明无法定位 TodoAI 项目

#### Scenario: No completed todos returns empty state

- **WHEN** 当前工作目录匹配已登记项目 `frontend-app`
- **AND** 该项目没有匹配的 `completed` TODO
- **THEN** 执行 `todoai list --done` 返回成功
- **AND** stdout 为 JSON 空数组

### Requirement: Run Todo Lifecycle Commands From Task Folder

系统 SHALL 提供 `todoai start` 和 `todoai done` 命令，用于从当前任务文件夹或其子目录触发当前 TODO 的状态流转。命令 MUST 在不启动新 Wails GUI 的情况下执行，并 SHALL 通过本机 IPC 委托已运行的 TodoAI GUI 进程执行与页面“开始”和“完成”按钮相同的后端逻辑。GUI 进程成功执行后 SHALL 通过现有页面状态事件通知前端刷新。

#### Scenario: Start todo from task folder

- **WHEN** TodoAI GUI 已运行并打开包含 TODO `实现登录` 的 workspace
- **AND** 当前工作目录位于 TODO `实现登录` 的任务文件夹或其子目录
- **AND** TODO `实现登录` 的状态允许页面“开始”按钮执行
- **THEN** 执行 `todoai start` 返回成功
- **AND** 系统调用与页面“开始”按钮相同的后端逻辑
- **AND** TODO `实现登录` 按现有页面逻辑进入 `in-progress` 流程
- **AND** 页面收到状态更新并刷新
- **AND** 系统不启动新的 Wails GUI

#### Scenario: Complete todo from task folder

- **WHEN** TodoAI GUI 已运行并打开包含 TODO `实现登录` 的 workspace
- **AND** 当前工作目录位于 TODO `实现登录` 的任务文件夹或其子目录
- **AND** TODO `实现登录` 的状态允许页面“完成”按钮执行
- **THEN** 执行 `todoai done` 返回成功
- **AND** 系统调用与页面“完成”按钮相同的后端逻辑
- **AND** TODO `实现登录` 按现有页面逻辑进入完成流程
- **AND** 页面收到状态更新并刷新
- **AND** 系统不启动新的 Wails GUI

#### Scenario: Complete todo with lifecycle script

- **WHEN** TodoAI GUI 已运行并打开包含 TODO `实现登录` 的 workspace
- **AND** 当前工作目录位于 TODO `实现登录` 的任务文件夹或其子目录
- **AND** TODO `实现登录` 配置了完成生命周期脚本
- **THEN** 执行 `todoai done` 返回成功
- **AND** 系统启动与页面“完成”按钮相同的完成生命周期脚本流程
- **AND** TODO 的最终完成状态 SHALL 由现有生命周期脚本成功回调决定
- **AND** 页面 SHALL 按现有生命周期脚本状态事件和 workspace 状态事件刷新

### Requirement: Lifecycle CLI Commands Require Active GUI IPC

系统 SHALL 只通过已运行 GUI 进程执行 `todoai start` 和 `todoai done`。当 GUI IPC 不可用、认证失败、当前目录无法解析为当前 workspace 的任务文件夹，或任务状态不满足现有页面按钮约束时，命令 MUST 返回失败并输出明确错误。系统 MUST NOT 在这些失败场景中直接修改 TodoAI 持久化状态。

#### Scenario: GUI is not running

- **WHEN** 当前没有可连接的 TodoAI GUI IPC 服务
- **THEN** 执行 `todoai start` 返回失败
- **AND** stderr 输出 TodoAI 页面未运行或不可达的错误说明
- **AND** 系统不直接修改 TodoAI 持久化状态
- **AND** 系统不启动新的 Wails GUI

#### Scenario: Current directory is not a task folder

- **WHEN** TodoAI GUI 已运行
- **AND** 当前工作目录不位于当前 GUI workspace 的任何任务文件夹内
- **THEN** 执行 `todoai done` 返回失败
- **AND** stderr 输出无法定位当前任务的错误说明
- **AND** 系统不直接修改 TodoAI 持久化状态

#### Scenario: Todo state rejects lifecycle command

- **WHEN** TodoAI GUI 已运行并打开包含 TODO `实现登录` 的 workspace
- **AND** 当前工作目录位于 TODO `实现登录` 的任务文件夹或其子目录
- **AND** TODO `实现登录` 的状态不允许页面对应按钮执行
- **THEN** 执行 `todoai start` 或 `todoai done` 返回失败
- **AND** stderr 输出状态流转失败的错误说明
- **AND** TODO `实现登录` 的状态保持不变

