## MODIFIED Requirements

### Requirement: Create Project Worktrees Inside Todo Workspace

系统 SHALL 为执行中 TODO 的每个关联 Git 项目在任务工作区目录下创建 Git worktree。每个 TODO 项目 SHALL 保存 base 分支、当前 worktree 分支、worktree 路径和准备状态。用户在 TODO 项目分支输入框中选择或输入的分支 SHALL 只作为该 TODO 项目的 base 分支。项目级终端 SHALL 仅在对应 TODO 项目 worktree 准备成功后可创建。

#### Scenario: Existing branch creates isolated worktree branch

- **WHEN** TODO `修复登录问题` 进入 `in-progress`
- **AND** 该 TODO 关联 Git 项目 `frontend-app`
- **AND** 用户为 `frontend-app` 选择已存在分支 `develop`
- **THEN** 系统将 `develop` 保存为该 TODO 项目的 base 分支
- **AND** 系统从 `develop` 创建该 TODO 项目的隔离 worktree 分支
- **AND** 系统在任务工作区目录下创建 `frontend-app` 的 worktree
- **AND** 该 TODO 项目保存 worktree 路径和当前 worktree 分支

#### Scenario: New input branch is created from main branch as base branch

- **WHEN** TODO `修复登录问题` 进入 `in-progress`
- **AND** 该 TODO 关联 Git 项目 `frontend-app`
- **AND** 用户输入不存在的分支 `feature/login-fix`
- **THEN** 系统从主分支创建 `feature/login-fix`
- **AND** 系统将 `feature/login-fix` 保存为该 TODO 项目的 base 分支
- **AND** 系统从 `feature/login-fix` 创建该 TODO 项目的隔离 worktree 分支
- **AND** 系统在任务工作区目录下创建 `frontend-app` 的 worktree
- **AND** 该 TODO 项目的当前 worktree 分支保存为隔离 worktree 分支

#### Scenario: Missing git repository records worktree failure

- **WHEN** TODO `修复登录问题` 进入 `in-progress`
- **AND** 该 TODO 关联项目 `docs-site`
- **AND** `docs-site` 路径不是 Git 仓库
- **THEN** 系统不为 `docs-site` 创建 worktree
- **AND** 该 TODO 项目保存 worktree 准备失败状态
- **AND** 系统阻止为该 TODO 项目创建项目终端
