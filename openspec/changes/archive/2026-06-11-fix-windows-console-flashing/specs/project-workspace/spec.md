## MODIFIED Requirements

### Requirement: Refresh Active Project Git Status

系统 SHALL 在当前项目可能变化或仓库状态可能变化时刷新状态栏中的 Git 信息。刷新 SHALL 不改变终端区域的布局高度。系统 SHALL 避免在项目导入完成后立即触发 Git 状态查询，并 SHALL 将导入后的 Git 状态查询延迟到用户展开 TODO、选择 TODO 项目、显式选择项目或其他明确刷新时机。系统在 Windows 上执行后台 Git 状态刷新和 Git 初始化命令时 SHALL 不显示临时控制台窗口，并且 SHALL 防止窗口 focus 事件在短时间内重复触发同一项目的 Git 状态刷新。

#### Scenario: Project selection refreshes git status

- **WHEN** 用户从项目树选择另一个可用项目
- **THEN** 系统查询新激活项目的 Git 状态
- **AND** 状态栏显示新激活项目的 Git 信息

#### Scenario: Importing projects defers git status refresh

- **WHEN** 用户从父目录导入一个或多个项目
- **THEN** 系统更新项目列表和导入摘要
- **AND** 系统不立即查询任何导入项目的 Git 状态
- **AND** 状态栏保持导入前的 Git 状态或空状态

#### Scenario: Expanding todo refreshes active todo project git status

- **WHEN** 用户展开一个包含当前激活 TODO project 的 TODO
- **AND** 当前激活项目路径可用
- **THEN** 系统查询当前激活项目的 Git 状态
- **AND** 状态栏显示当前激活项目的 Git 信息

#### Scenario: Selecting todo project refreshes git status

- **WHEN** 用户选择 TODO 下的一个项目
- **THEN** 系统查询该项目的 Git 状态
- **AND** 状态栏显示该项目的 Git 信息

#### Scenario: Terminal command completion refreshes git status

- **WHEN** 激活项目的终端命令结束
- **THEN** 系统刷新当前激活项目的 Git 状态
- **AND** 状态栏反映命令结束后的改动文件数量

#### Scenario: Window focus refreshes git status

- **WHEN** 应用窗口重新获得焦点
- **AND** 当前激活项目可用
- **THEN** 系统刷新当前激活项目的 Git 状态

#### Scenario: Windows Git refresh does not flash console windows

- **WHEN** Windows 用户打开应用或切换到一个可用 Git 项目
- **THEN** 系统在后台查询该项目的 Git 状态
- **AND** 查询过程不显示系统控制台窗口
- **AND** 应用窗口保持可用

#### Scenario: Focus jitter does not start repeated Git refreshes

- **WHEN** 应用窗口在短时间内重复获得焦点
- **AND** 当前激活项目没有变化
- **THEN** 系统最多启动一次当前项目的 focus Git 状态刷新
- **AND** 后续重复 focus 事件不会启动新的后台 Git 命令，直到去重窗口结束或已有请求完成

#### Scenario: Windows Git initialization does not flash console windows

- **WHEN** Windows 用户从状态栏初始化当前项目的 Git 仓库
- **THEN** 系统在后台执行 Git 初始化
- **AND** 初始化过程不显示系统控制台窗口
- **AND** 初始化完成后系统刷新当前项目的 Git 状态
