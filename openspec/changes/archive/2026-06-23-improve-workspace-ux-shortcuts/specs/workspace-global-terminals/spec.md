## ADDED Requirements

### Requirement: Create Workspace Global Terminals

系统 SHALL 允许用户在当前 workspace 中创建多个全局终端。全局终端 SHALL 不属于任何 TODO、TODO project 或全局项目候选。每个全局终端的 shell 进程 SHALL 以当前 workspace 根目录作为工作目录。

#### Scenario: User creates a workspace global terminal

- **WHEN** 用户打开 workspace `/home/user/work/customer-a`
- **AND** 用户创建一个全局终端
- **THEN** 系统创建一个新的终端会话
- **AND** 该终端不包含 TODO ID 或 TODO project ID
- **AND** 该终端的 shell 工作目录为 `/home/user/work/customer-a`

#### Scenario: User creates multiple workspace global terminals

- **WHEN** 用户打开 workspace `/home/user/work/customer-a`
- **AND** 用户创建全局终端 A
- **AND** 用户创建全局终端 B
- **THEN** 全局终端 A 和全局终端 B 是独立终端会话
- **AND** 两个终端都显示在全局终端分组中
- **AND** 两个终端都不显示在任何 TODO 项目终端列表中

### Requirement: Display Workspace Global Terminal Group

系统 SHALL 在终端区域顶部展示全局终端分组。全局终端分组 SHALL 显示当前 workspace 的全局终端列表和创建入口。当当前 workspace 不存在全局终端时，系统 SHALL 不渲染该分组，且 SHALL 不为该分组保留空白高度。

#### Scenario: Global terminal group appears when terminals exist

- **WHEN** 当前 workspace 存在至少一个全局终端
- **THEN** 终端区域顶部显示全局终端分组
- **AND** 该分组显示全局终端列表
- **AND** 用户可以从该区域创建新的全局终端

#### Scenario: Global terminal group is hidden when empty

- **WHEN** 当前 workspace 不存在全局终端
- **THEN** 终端区域不显示全局终端分组
- **AND** 终端区域不显示全局终端空状态
- **AND** 终端区域不为全局终端分组保留布局高度

### Requirement: Select Workspace Global Terminal

系统 SHALL 允许用户选择全局终端作为当前活动终端。选择全局终端 SHALL 激活对应 xterm pane，但 SHALL NOT 改变当前 TODO、当前 TODO project、当前项目候选或当前 Git 状态上下文。

#### Scenario: User selects global terminal without changing todo context

- **WHEN** 当前 TODO project 上下文为 TODO `修复登录问题` 下的 `frontend-app`
- **AND** 当前项目 Git 状态显示 `frontend-app`
- **AND** 用户选择一个全局终端
- **THEN** 该全局终端成为当前活动终端
- **AND** 当前 TODO project 上下文仍为 TODO `修复登录问题` 下的 `frontend-app`
- **AND** 当前项目 Git 状态上下文仍为 `frontend-app`

#### Scenario: User returns to todo project terminal

- **WHEN** 用户已选择全局终端
- **AND** TODO `修复登录问题` 下的 `frontend-app` 存在终端 A
- **AND** 用户选择终端 A
- **THEN** 终端 A 成为当前活动终端
- **AND** 当前 TODO project 上下文更新为 TODO `修复登录问题` 下的 `frontend-app`

### Requirement: Manage Workspace Global Terminal Lifecycle

系统 SHALL 支持全局终端的输入、输出、resize、复制、粘贴、删除和 exited 后重启。删除全局终端 SHALL 只关闭和移除该全局终端，不影响 TODO project 终端。

#### Scenario: User deletes a global terminal

- **WHEN** 全局终端 A 正在运行
- **AND** TODO project 终端 B 正在运行
- **AND** 用户删除全局终端 A
- **THEN** 系统关闭全局终端 A 的 shell 进程
- **AND** 全局终端 A 不再显示
- **AND** TODO project 终端 B 继续运行并保持显示

#### Scenario: User restarts an exited global terminal

- **WHEN** 全局终端 A 的 shell 已退出
- **AND** 用户请求重启全局终端 A
- **THEN** 系统在当前 workspace 根目录重新启动全局终端 A 的 shell
- **AND** 全局终端 A 保持全局终端归属
