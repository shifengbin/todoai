## MODIFIED Requirements

### Requirement: Start Shell In Project Directory

系统 SHALL 在为可用 TODO project context 创建终端 session 时使用该 context 解析出的工作目录启动 embedded shell。若 TODO project 的 worktree 状态为 ready 且 worktree 目录可用，工作目录 SHALL 为该 worktree 目录。若 TODO project 的 worktree 状态为 cleared 且原项目目录可用，工作目录 SHALL 回退为该 TODO project 保存的原项目目录。若 TODO project 尚未准备 worktree、worktree 准备失败或无法解析可用工作目录，系统 SHALL 拒绝创建终端并 SHALL NOT 启动 shell 进程。

#### Scenario: Shell starts with todo project worktree directory

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 ready
- **AND** 该 TODO project worktree 路径为 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **THEN** 该终端 shell session 使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动
- **AND** 该终端归属于 TODO `fix-login` 和项目 `demo-app` 的 TODO project context

#### Scenario: Shell starts with source project directory for cleared worktree

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 原项目目录可用
- **THEN** 该终端 shell session 使用工作目录 `/home/user/work/demo-app` 启动
- **AND** 该终端仍归属于 TODO `fix-login` 和项目 `demo-app` 的 TODO project context

#### Scenario: Todo project without prepared or cleared worktree cannot start shell

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 没有已准备的 worktree 路径
- **AND** 该 TODO project 未被标记为 worktree 已清除
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程

#### Scenario: Cleared worktree with unavailable source project cannot start shell

- **WHEN** 用户为 TODO `fix-login` 下项目 `demo-app` 创建终端
- **AND** 该 TODO project 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录不可用
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程

### Requirement: Create Additional Project Terminal

系统 SHALL 允许用户仅为 `in-progress` TODO 中可用的 TODO project context 创建额外终端 session。若该 TODO project worktree 为 ready，新的 shell 进程 SHALL 在该 worktree 目录启动；若该 TODO project worktree 为 cleared 且原项目目录可用，新的 shell 进程 SHALL 在原项目目录启动。每个新终端 session SHALL 归属于该 TODO project context，并 SHALL 与该 context 下已有终端相互独立。系统 SHALL 拒绝为 `not-started` TODO project context、未准备且未清除 worktree 的 TODO project context、worktree 准备失败的 TODO project context 或项目路径不可用的 TODO project context 创建终端。

#### Scenario: User creates another terminal for an in-progress todo project

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 有已准备 worktree 路径 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** 用户在 TODO `fix-login` 下项目 `demo-app` 创建新终端
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动新 shell 进程
- **AND** 新终端独立于该 TODO project context 中已有终端
- **AND** 新终端不显示在引用同一项目的其它 TODO 下

#### Scenario: User creates another terminal for a cleared worktree todo project

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 用户在 TODO `fix-login` 下项目 `demo-app` 创建新终端
- **THEN** 系统使用工作目录 `/home/user/work/demo-app` 启动新 shell 进程
- **AND** 新终端仍归属于该 TODO project context
- **AND** 新终端不显示在引用同一项目的其它 TODO 下

#### Scenario: Not-started todo project cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `not-started`
- **AND** 用户或客户端请求在 TODO `fix-login` 下项目 `/home/user/work/demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project with failed worktree cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 `failed`
- **AND** 用户请求为项目 `demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project without prepared or cleared worktree cannot create terminal

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 未准备完成
- **AND** 该 TODO project 未被标记为 worktree 已清除
- **AND** 用户请求为项目 `demo-app` 创建新终端
- **THEN** 系统拒绝终端创建请求
- **AND** 不启动 shell 进程
- **AND** 不向该 TODO project context 添加运行时终端 session

### Requirement: Start Background Launch Profile Command

系统 SHALL 在不注册 embedded terminal session 的情况下启动后台终端启动配置命令。后台启动命令 SHALL 使用配置的终端 shell 的一次性命令模式，SHALL 运行在同一 selected context 的可见终端会使用的工作目录中，并 SHALL 等待进程退出以释放资源。后台命令输出和退出状态 SHALL NOT 显示在终端 UI 中。

#### Scenario: Background project command runs in todo project worktree

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 有已准备 worktree 路径 `/home/user/work/customer-a/tasks/abc123/demo-app`
- **AND** 用户从该 TODO project 启动菜单选择启动参数为 `npm run sync` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123/demo-app` 启动后台命令
- **AND** 不向该 TODO project context 添加 embedded terminal session

#### Scenario: Background project command runs in source project directory for cleared worktree

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目目录为 `/home/user/work/demo-app`
- **AND** 用户从该 TODO project 启动菜单选择启动参数为 `npm run sync` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/demo-app` 启动后台命令
- **AND** 不向该 TODO project context 添加 embedded terminal session

#### Scenario: Background task command runs in todo workspace directory

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 有任务工作区目录 `/home/user/work/customer-a/tasks/abc123`
- **AND** 用户从任务级启动菜单选择启动参数为 `npm run prepare` 的后台启动配置
- **THEN** 系统使用工作目录 `/home/user/work/customer-a/tasks/abc123` 启动后台命令
- **AND** 不向该 TODO context 添加任务终端 session

#### Scenario: Background command does not affect terminal UI state

- **WHEN** 终端 `terminal-a` 是当前活动终端
- **AND** 用户从终端启动菜单选择后台启动配置
- **THEN** `terminal-a` 仍是当前活动终端
- **AND** 终端列表保持不变
- **AND** 终端历史保持不变
- **AND** 系统不为该后台命令发出终端输出、终端状态或终端 agent 状态事件

#### Scenario: Background command exits without UI change

- **WHEN** 后台启动配置命令正在运行
- **AND** 后台进程退出
- **THEN** 系统释放进程资源
- **AND** 没有终端被标记为 exited
- **AND** 不显示新终端

#### Scenario: Not-started todo project cannot start background command

- **WHEN** TODO `fix-login` 的状态为 `not-started`
- **AND** 用户或客户端请求在 TODO `fix-login` 下项目 `demo-app` 启动后台命令
- **THEN** 系统拒绝后台命令请求
- **AND** 不启动后台进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Todo project with failed worktree cannot start background command

- **WHEN** TODO `fix-login` 的状态为 `in-progress`
- **AND** TODO `fix-login` 下项目 `demo-app` 的 worktree 状态为 `failed`
- **AND** 用户请求为该 TODO project 启动后台命令
- **THEN** 系统拒绝后台命令请求
- **AND** 不启动后台进程
- **AND** 不向该 TODO project context 添加运行时终端 session

#### Scenario: Background command start failure is reported without terminal creation

- **WHEN** 用户选择无法启动命令的后台启动配置
- **THEN** 系统通过现有应用错误显示报告启动错误
- **AND** 不添加 embedded terminal session
- **AND** 当前活动终端保持不变
