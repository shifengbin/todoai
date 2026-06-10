## ADDED Requirements

### Requirement: Initialize Active Project Git Repository From Status Bar

系统 SHALL 在当前激活项目可用且不是 Git 仓库时，在底部状态栏提供初始化 Git 仓库的操作。该操作 SHALL 直接在当前项目路径执行 Git 初始化，并在成功后刷新状态栏中的 Git 信息。

#### Scenario: Non-Git project shows initialize action

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径不是 Git 仓库
- **THEN** 状态栏显示该项目不是 Git 仓库
- **AND** 状态栏显示 `Initialize Git Repository` 操作

#### Scenario: User initializes Git repository from status bar

- **WHEN** 当前激活项目可用
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

#### Scenario: Git initialization fails

- **WHEN** 当前激活项目可用
- **AND** 用户点击 `Initialize Git Repository`
- **AND** Git 初始化失败
- **THEN** 状态栏仍显示该项目不是 Git 仓库
- **AND** 系统显示不会阻断终端操作的错误信息

## MODIFIED Requirements

### Requirement: Display Active Project Git Status Bar

系统 SHALL 在主工作区底部显示固定高度状态栏，用于展示当前激活项目的 Git 仓库信息。状态栏 SHALL 使用可区分的圆角信息块展示 Git 信息，并 SHALL 至少包含当前分支名称和当前项目的已改动文件数量。

#### Scenario: Active Git project shows branch and change count

- **WHEN** 用户选择一个可用项目
- **AND** 该项目路径是 Git 仓库
- **AND** 当前分支是 `main`
- **AND** 仓库有 3 个已改动文件
- **THEN** 状态栏显示分支 `main`
- **AND** 状态栏显示 3 个已改动文件
- **AND** 分支和改动数量显示在独立圆角信息块中

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

#### Scenario: No active project keeps status bar stable

- **WHEN** 没有项目被选中
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏显示没有项目的空状态
