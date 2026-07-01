## MODIFIED Requirements

### Requirement: Show Todo Project Terminal Launch Menu

系统 SHALL 仅在 `in-progress` TODO 的可用项目行上提供终端启动菜单。启动菜单 SHALL 包含内置 `Terminal` 选项和已配置的终端启动配置。`not-started` TODO 的项目行 SHALL NOT 暴露新增终端启动菜单。worktree 准备失败的 TODO 项目行 SHALL NOT 暴露可用的终端启动入口。worktree 已清除但原项目路径可用的 TODO 项目行 SHALL 继续暴露终端启动菜单。

#### Scenario: In-progress todo project launch menu contains configured profiles

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下项目 `frontend-app` 的 worktree 状态为 ready
- **AND** 设置中包含启动配置 `codex` 和 `claude`
- **AND** 用户激活 TODO `修复登录问题` 下项目 `frontend-app` 的新增终端控件
- **THEN** 启动菜单显示 `Terminal` 作为第一项
- **AND** 启动菜单按配置顺序显示 `codex` 和 `claude`

#### Scenario: Cleared worktree todo project keeps launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下项目 `frontend-app` 的原项目路径可用
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **AND** 用户激活该 TODO project 的新增终端控件
- **THEN** 启动菜单显示 `Terminal` 作为第一项
- **AND** 启动菜单显示已启用的终端启动配置

#### Scenario: Not-started todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

#### Scenario: Unavailable todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径不可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

#### Scenario: Failed worktree todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下项目 `frontend-app` 的 worktree 状态为 failed
- **THEN** 该 TODO 项目行不暴露可用的新增终端启动菜单

### Requirement: Display Todo Project Current Worktree Branch

系统 SHALL 在左侧 TODO 工作树的 TODO project 行中显示项目名称和当前 worktree 真实分支或 worktree 清除状态。分支显示格式 SHALL 为 `项目名称(分支名称)`。分支名称 MUST 来自该 TODO project ready worktree 路径的当前 Git 状态，MUST NOT 使用创建 worktree 时保存的静态 `worktreeBranch` 字段作为实时显示来源。当 TODO project worktree 状态为 cleared，或分支刷新检测到 ready worktree 路径或保存的 worktree 分支已不存在时，系统 SHALL 在同一括号后缀位置显示 `项目名称(worktree已清除)`。若当前分支不可用、worktree 未 ready 且未被判定为 cleared、Git 状态不可用或 Git 查询失败，系统 SHALL 保持显示项目名称且不追加分支后缀。该显示要求 SHALL 只适用于左侧 TODO 项目列表，SHALL NOT 改变顶部工作区标题。

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

#### Scenario: Removed worktree directory shows cleared label

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project 保存的 worktree 状态为 ready
- **AND** 该 TODO project 保存的 worktree 路径在磁盘上不存在
- **WHEN** 系统刷新该 TODO project 的分支显示
- **THEN** 左侧 TODO 工作树中的该项目行显示 `frontend-app(worktree已清除)`
- **AND** 左侧 TODO 工作树中的该项目行不显示旧 worktree 分支名称

#### Scenario: Removed worktree branch shows cleared label

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project 保存的 worktree 分支为 `todo/fix-login/frontend-app`
- **AND** 系统确认该 worktree 分支已不存在
- **THEN** 左侧 TODO 工作树中的该项目行显示 `frontend-app(worktree已清除)`

#### Scenario: Cleared worktree state shows cleared label immediately

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **THEN** 左侧 TODO 工作树中的该项目行显示 `frontend-app(worktree已清除)`
- **AND** 系统不需要先成功读取该 worktree 的 Git 分支

#### Scenario: Unavailable branch omits branch suffix

- **WHEN** TODO `修复登录问题` 下显示 TODO project `frontend-app`
- **AND** 该 TODO project worktree 未 ready 且未被判定为 cleared，或 Git 状态不可用
- **THEN** 左侧 TODO 工作树中的该项目行显示项目名称 `frontend-app`
- **AND** 左侧 TODO 工作树中的该项目行不显示分支括号后缀
