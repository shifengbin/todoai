## MODIFIED Requirements

### Requirement: Display Active Project Git Status Bar

系统 SHALL 在主工作区底部显示固定高度状态栏，用于展示当前激活项目上下文的 Git 仓库信息。状态栏 SHALL 使用可区分的圆角信息块展示 Git 信息，并 SHALL 至少包含当前分支名称和当前上下文目录的已改动文件数量。当当前上下文是已准备 worktree 的 TODO project 时，状态栏 SHALL 展示该 TODO project worktree 目录的 Git 信息，而不是来源项目目录的 Git 信息。当当前上下文是 TODO 级控制台时，状态栏 SHALL 只查询该 TODO 任务文件夹根目录本身的 Git 状态，SHALL NOT 沿用上一个项目或 TODO project 的 Git 状态，也 SHALL NOT 扫描该 TODO 任务文件夹的子目录仓库。当没有当前激活项目上下文时，状态栏 SHALL 保持固定底部高度，但 SHALL NOT 显示 Git 状态 chip 或 `No project` 空状态文案。

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

#### Scenario: No active project hides git status chips

- **WHEN** 没有项目或 TODO project 被选中
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏不显示 Git 状态 chip
- **AND** 状态栏不显示 `No project` 文案

#### Scenario: Active todo task terminal shows todo workspace git status

- **WHEN** 用户选择 TODO `修复登录问题` 的 TODO 级控制台
- **AND** 该 TODO 任务文件夹根目录本身是 Git 仓库
- **AND** 当前分支是 `todo/root`
- **AND** 仓库有 1 个已改动文件
- **THEN** 状态栏显示分支 `todo/root`
- **AND** 状态栏显示 1 个已改动文件
- **AND** 状态栏不显示上一个项目或 TODO project 的 Git 状态

#### Scenario: Active todo task terminal without root git repository hides git status chips

- **WHEN** 用户选择 TODO `修复登录问题` 的 TODO 级控制台
- **AND** 该 TODO 任务文件夹根目录本身不是 Git 仓库
- **AND** 该 TODO 任务文件夹子目录中存在项目 worktree 或其他 Git 仓库
- **THEN** 状态栏仍占据固定底部高度
- **AND** 状态栏不显示 Git 状态 chip
- **AND** 状态栏不显示 `Not a git repository`
- **AND** 状态栏不显示上一个项目或 TODO project 的 Git 状态
