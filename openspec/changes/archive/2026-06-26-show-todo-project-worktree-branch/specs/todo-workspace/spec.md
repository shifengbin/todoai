## ADDED Requirements

### Requirement: Display Todo Project Current Worktree Branch

系统 SHALL 在左侧 TODO 工作树的 TODO project 行中显示项目名称和当前 worktree 真实分支。分支显示格式 SHALL 为 `项目名称(分支名称)`。分支名称 MUST 来自该 TODO project ready worktree 路径的当前 Git 状态，MUST NOT 使用创建 worktree 时保存的静态 `worktreeBranch` 字段作为实时显示来源。若当前分支不可用、worktree 未 ready、Git 状态不可用或 Git 查询失败，系统 SHALL 保持显示项目名称且不追加分支后缀。该显示要求 SHALL 只适用于左侧 TODO 项目列表，SHALL NOT 改变顶部工作区标题。

#### Scenario: Todo project row shows current worktree branch

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project 的 worktree 状态为 ready
- **AND** 该 TODO project worktree 当前 Git 分支为 `feature/login`
- **THEN** 左侧 TODO 工作树中的该项目行显示 `frontend-app(feature/login)`
- **AND** 顶部工作区标题不追加该分支名称

#### Scenario: Todo project row uses live branch instead of stored worktree branch

- **WHEN** TODO `修复登录问题` 下的 TODO project `frontend-app` 保存的 `worktreeBranch` 为 `todo/fix-login/frontend-app`
- **AND** 该 TODO project worktree 当前 Git 分支为 `feature/status-bar`
- **THEN** 左侧 TODO 工作树中的该项目行显示 `frontend-app(feature/status-bar)`
- **AND** 左侧 TODO 工作树中的该项目行不显示 `frontend-app(todo/fix-login/frontend-app)`

#### Scenario: Worktree terminal command completion refreshes todo project branch label

- **WHEN** TODO `修复登录问题` 下的 TODO project `frontend-app` 当前显示为 `frontend-app(feature/login)`
- **AND** 用户在该 TODO project 的 worktree 终端中执行切换分支命令
- **AND** 该终端命令结束后 worktree 当前 Git 分支为 `feature/payments`
- **THEN** 左侧 TODO 工作树中的该项目行刷新为 `frontend-app(feature/payments)`

#### Scenario: Unavailable branch omits branch suffix

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project worktree 未 ready 或 Git 状态不可用
- **THEN** 左侧 TODO 工作树中的该项目行显示项目名称 `frontend-app`
- **AND** 左侧 TODO 工作树中的该项目行不显示分支括号后缀
