## MODIFIED Requirements

### Requirement: Handle Git Status Unavailable

系统 SHALL 在当前项目不是 Git 仓库、项目路径不可用、Git 命令未安装或 Git 状态查询失败时保持状态栏可用，并显示不会阻断项目或终端操作的状态信息。系统在执行 Git 状态查询前 SHALL 校验 `git` 命令是否存在；若 `git` 命令不存在或不可执行，系统 SHALL 不执行 Git 状态查询命令。

#### Scenario: Active project is not a git repository

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径不是 Git 仓库
- **AND** 系统可以执行 `git` 命令
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 终端区域保持可用

#### Scenario: Active project path is unavailable

- **WHEN** 当前激活项目路径不可访问
- **THEN** 状态栏显示项目路径不可用的状态
- **AND** 系统不执行 Git 状态查询

#### Scenario: Git executable is not installed

- **WHEN** 当前激活项目可用
- **AND** 系统无法找到可执行的 `git` 命令
- **THEN** 系统不执行 Git 状态查询命令
- **AND** 状态栏显示 `未安装 Git`
- **AND** 项目选择和终端会话功能保持可用

#### Scenario: Git status command fails

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
- **AND** Git 状态查询失败
- **THEN** 状态栏显示 Git 状态不可用
- **AND** 项目选择和终端会话功能保持可用

### Requirement: Initialize Active Project Git Repository From Status Bar

系统 SHALL 在当前激活项目可用、系统可以执行 `git` 命令且该项目不是 Git 仓库时，在底部状态栏提供初始化 Git 仓库的操作。该操作 SHALL 直接在当前项目路径执行 Git 初始化，并在成功后刷新状态栏中的 Git 信息。系统在执行 Git 初始化前 SHALL 校验 `git` 命令是否存在；若 `git` 命令不存在或不可执行，系统 SHALL 不执行 Git 初始化命令，并 SHALL 返回明确的未安装 Git 错误。

#### Scenario: Non-Git project shows initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 系统可以执行 `git` 命令
- **AND** 该项目路径不是 Git 仓库
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 状态栏显示 `Initialize Git Repository` 操作

#### Scenario: Missing Git hides initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 系统无法找到可执行的 `git` 命令
- **THEN** 状态栏显示 `未安装 Git`
- **AND** 状态栏不显示 `Initialize Git Repository` 操作

#### Scenario: User initializes Git repository from status bar

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
- **AND** 该项目路径不是 Git 仓库
- **AND** 用户点击 `Initialize Git Repository`
- **THEN** 系统在当前项目路径执行 `git init`
- **AND** 状态栏刷新当前项目的 Git 信息
- **AND** 状态栏不再显示非 Git 仓库状态

#### Scenario: Git initialization is in progress

- **WHEN** 用户点击 `Initialize Git Repository`
- **AND** Git 初始化请求尚未完成
- **THEN** 初始化操作不可重复触发
- **AND** 状态栏显示初始化进行中的状态

#### Scenario: Git initialization is blocked when Git is missing

- **WHEN** 当前激活项目可用
- **AND** 系统无法找到可执行的 `git` 命令
- **AND** 用户或客户端请求初始化当前项目的 Git 仓库
- **THEN** 系统不执行 `git init`
- **AND** 系统返回未安装 Git 的错误

#### Scenario: Git initialization fails

- **WHEN** 当前激活项目可用
- **AND** 系统可以执行 `git` 命令
- **AND** 用户点击 `Initialize Git Repository`
- **AND** Git 初始化失败
- **THEN** 状态栏仍显示该项目不是 Git 仓库
- **AND** 系统显示不会阻断终端操作的错误信息
