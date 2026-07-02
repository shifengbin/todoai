## MODIFIED Requirements

### Requirement: Display Todo Task Terminals In Tree

系统 SHALL 在 TODO 工作树中显示任务级终端入口和任务级终端列表。任务级终端 SHALL 显示在 TODO 下、项目列表之前，并 SHALL 不归属于任何 TODO 项目。任务级终端的活动状态 SHALL 参与收起 TODO 行的聚合活动状态。当任务级终端是当前活动终端时，系统 SHALL 将其父 TODO 渲染为当前任务上下文，并 SHALL 在工作区标题中显示该 TODO 的任务级上下文和任务目录路径。

#### Scenario: Todo shows task terminal list

- **WHEN** TODO `修复登录问题` 有任务级终端 `zsh`
- **AND** TODO `修复登录问题` 关联项目 `frontend-app`
- **THEN** TODO 工作树在 `修复登录问题` 下显示任务级终端 `zsh`
- **AND** TODO 工作树在任务级终端后显示项目 `frontend-app`
- **AND** 任务级终端 `zsh` 不显示在 `frontend-app` 的项目终端列表中

#### Scenario: Collapsed todo aggregates task terminal activity

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的任务级终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行呼吸式状态反馈显示运行中的聚合活动状态

#### Scenario: Active task terminal highlights owning todo

- **WHEN** TODO `修复登录问题` 下存在任务级终端 `zsh`
- **AND** 用户选择任务级终端 `zsh`
- **THEN** 任务级终端 `zsh` 显示为当前活动终端
- **AND** TODO `修复登录问题` 行显示当前任务上下文的选中背景
- **AND** TODO project `frontend-app` 不因任务级终端选择而显示为当前选中 TODO project

#### Scenario: Active task terminal displays task heading

- **WHEN** 当前 workspace 路径为 `/home/user/work/customer-a`
- **AND** TODO `修复登录问题` 的任务目录为 `/home/user/work/customer-a/tasks/abc123`
- **AND** 用户选择 TODO `修复登录问题` 下的任务级终端 `zsh`
- **THEN** 工作区标题显示 `修复登录问题 / 任务终端`
- **AND** 工作区路径显示 `/home/user/work/customer-a/tasks/abc123`

## ADDED Requirements

### Requirement: Keep Terminal Launch Menu Visible

系统 SHALL 在 TODO 工作树中显示任务级终端和 TODO project 终端的启动菜单时避免被任务列表滚动容器裁剪。启动菜单 SHALL 锚定在触发按钮附近；当触发按钮靠近侧栏可视区域底部时，菜单 SHALL 自动上翻或限制最大高度并允许菜单内部滚动。打开终端启动菜单 SHALL NOT 改变当前 TODO 视图标签或左侧 TODO 栏宽度。

#### Scenario: Task terminal launch menu remains visible near list bottom

- **WHEN** TODO 工作树包含足够多的任务，使 TODO `修复登录问题` 靠近侧栏可视区域底部
- **AND** 用户点击 TODO `修复登录问题` 行上的任务级终端启动按钮
- **THEN** 系统在触发按钮附近显示任务级终端启动菜单
- **AND** 菜单内容不被 TODO 列表滚动容器裁剪
- **AND** 用户可以选择菜单中的 `Terminal` 启动项

#### Scenario: Todo project launch menu remains usable near list bottom

- **WHEN** TODO 工作树包含足够多的任务，使 TODO project `frontend-app` 靠近侧栏可视区域底部
- **AND** 用户点击 `frontend-app` 行上的终端启动按钮
- **THEN** 系统在触发按钮附近显示 TODO project 终端启动菜单
- **AND** 菜单自动上翻或限制高度以保持在侧栏可视区域内
- **AND** 用户可以选择菜单中的任一可用启动项
