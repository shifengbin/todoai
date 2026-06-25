## MODIFIED Requirements

### Requirement: Display Active Project Git Status Bar

系统 SHALL 在主工作区底部显示固定高度状态栏，用于展示当前激活项目上下文的 Git 仓库信息。状态栏 SHALL 使用可区分的圆角信息块展示 Git 信息，并 SHALL 至少包含当前分支名称和当前上下文目录的已改动文件数量。当当前上下文是已准备 worktree 的 TODO project 时，状态栏 SHALL 展示该 TODO project worktree 目录的 Git 信息，而不是来源项目目录的 Git 信息。

#### Scenario: Active Git project shows branch and change count

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 当前分支是 `main`
- **AND** 仓库有 3 个已改动文件
- **THEN** 状态栏显示分支 `main`
- **AND** 状态栏显示 3 个已改动文件
- **AND** 分支和改动数量显示在独立圆角信息块中

#### Scenario: Active todo project worktree shows worktree branch and change count

- **WHEN** 用户选择 TODO `修复登录问题` 下的项目 `frontend-app`
- **AND** 该 TODO project 的 worktree 状态为 ready
- **AND** 该 TODO project 的 worktree 路径是 Git 仓库
- **AND** 该 worktree 当前分支是 `todo/fix-login/frontend-app`
- **AND** 该 worktree 有 2 个已改动文件
- **AND** 来源项目路径当前分支是 `main` 且没有已改动文件
- **THEN** 状态栏显示分支 `todo/fix-login/frontend-app`
- **AND** 状态栏显示 2 个已改动文件
- **AND** 状态栏不显示来源项目路径的 `main` 分支状态

#### Scenario: Active Git project shows detailed change counts

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 仓库有暂存、未暂存或未跟踪文件
- **THEN** 状态栏分别显示 staged、unstaged 和 untracked 数量
- **AND** 每类数量显示在可区分的圆角信息块中

#### Scenario: Active Git project shows ahead and behind counts

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 当前分支相对上游存在 ahead 或 behind 数量
- **THEN** 状态栏显示 ahead 或 behind 数量
- **AND** ahead 或 behind 数量显示在独立圆角信息块中

#### Scenario: Active todo project without ready worktree keeps status bar stable

- **WHEN** 用户选择 TODO `修复登录问题` 下的项目 `frontend-app`
- **AND** 该 TODO project 没有 ready worktree 路径
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏显示当前 TODO project 的 Git 状态不可用或路径不可用状态
- **AND** 系统不使用来源项目路径作为该 TODO project 的 Git 状态替代

#### Scenario: No active project keeps status bar stable

- **WHEN** 没有项目被选中
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏显示没有项目的空状态

### Requirement: Refresh Active Project Git Status

系统 SHALL 在当前项目上下文可能变化或仓库状态可能变化时刷新状态栏中的 Git 信息。刷新 SHALL 不改变终端区域的布局高度。系统 SHALL 避免在项目导入完成后立即触发 Git 状态查询，并 SHALL 将导入后的 Git 状态查询延迟到用户展开 TODO、选择 TODO 项目、显式选择项目或其他明确刷新时机。系统在 Windows 上执行后台 Git 状态刷新和 Git 初始化命令时 SHALL 不显示临时控制台窗口，并且 SHALL 防止窗口 focus 事件在短时间内重复触发同一项目上下文的 Git 状态刷新。当前上下文是 TODO project 时，刷新 SHALL 以 TODO project 为上下文边界；同一来源项目下的不同 TODO project worktree SHALL 分别查询、去重和展示 Git 状态。

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
- **AND** 当前激活 TODO project 的 worktree 路径可用
- **THEN** 系统查询当前激活 TODO project worktree 的 Git 状态
- **AND** 状态栏显示当前激活 TODO project worktree 的 Git 信息

#### Scenario: Selecting todo project refreshes git status

- **WHEN** 用户选择 TODO 下的一个项目
- **AND** 该 TODO project 的 worktree 状态为 ready
- **THEN** 系统查询该 TODO project worktree 的 Git 状态
- **AND** 状态栏显示该 TODO project worktree 的 Git 信息

#### Scenario: Same source project todo worktrees keep independent git status

- **WHEN** TODO `修复登录问题` 和 TODO `升级依赖` 都关联来源项目 `frontend-app`
- **AND** 两个 TODO project 分别拥有不同的 ready worktree 路径
- **AND** 用户先选择 TODO `修复登录问题` 下的 `frontend-app`
- **AND** 系统正在查询该 TODO project worktree 的 Git 状态
- **AND** 用户随后选择 TODO `升级依赖` 下的 `frontend-app`
- **THEN** 系统按新的 TODO project worktree 发起独立 Git 状态查询
- **AND** 前一个 TODO project 的延迟响应不会覆盖当前状态栏
- **AND** 当前状态栏只显示 TODO `升级依赖` 下 `frontend-app` worktree 的 Git 信息

#### Scenario: Terminal command completion refreshes git status

- **WHEN** 激活项目上下文的终端命令结束
- **THEN** 系统刷新当前激活项目上下文的 Git 状态
- **AND** 状态栏反映命令结束后的改动文件数量

#### Scenario: Window focus refreshes git status

- **WHEN** 应用窗口重新获得焦点
- **AND** 当前激活项目上下文可用
- **THEN** 系统刷新当前激活项目上下文的 Git 状态

#### Scenario: Windows Git refresh does not flash console windows

- **WHEN** Windows 用户打开应用或切换到一个可用 Git 项目上下文
- **THEN** 系统在后台查询该项目上下文的 Git 状态
- **AND** 查询过程不显示系统控制台窗口
- **AND** 应用窗口保持可用

#### Scenario: Focus jitter does not start repeated Git refreshes

- **WHEN** 应用窗口在短时间内重复获得焦点
- **AND** 当前激活项目上下文没有变化
- **THEN** 系统最多启动一次当前项目上下文的 focus Git 状态刷新
- **AND** 后续重复 focus 事件不会启动新的后台 Git 命令，直到去重窗口结束或已有请求完成

#### Scenario: Windows Git initialization does not flash console windows

- **WHEN** Windows 用户从状态栏初始化当前项目的 Git 仓库
- **THEN** 系统在后台执行 Git 初始化
- **AND** 初始化过程不显示系统控制台窗口
- **AND** 初始化完成后系统刷新当前项目的 Git 状态
