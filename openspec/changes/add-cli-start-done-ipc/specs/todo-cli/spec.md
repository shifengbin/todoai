## ADDED Requirements

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
