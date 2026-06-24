## ADDED Requirements

### Requirement: Select Branch For Todo Projects

系统 SHALL 在创建 TODO 和为 TODO 添加项目时，允许用户为每个选中的项目选择或输入用于创建任务 worktree 的分支。分支控件 SHALL 默认使用主分支。用户选择的分支信息 SHALL 保存到 TODO 工程副本中，并 SHALL 用于后续创建任务项目 worktree。

#### Scenario: User selects branch while creating todo with project

- **WHEN** 全局项目候选包含 Git 项目 `frontend-app`
- **AND** 用户打开创建 TODO 表单
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户在 `frontend-app` 的分支控件中选择 `develop`
- **AND** 用户提交创建 TODO
- **THEN** TODO 下保存工程副本 `frontend-app`
- **AND** 该工程副本保存 base 分支选择 `develop`

#### Scenario: Project branch defaults to main branch

- **WHEN** 全局项目候选包含 Git 项目 `frontend-app`
- **AND** 用户创建 TODO 时选择工程 `frontend-app`
- **AND** 用户未选择或输入分支
- **THEN** TODO 下保存工程副本 `frontend-app`
- **AND** 该工程副本保存默认主分支作为 base 分支

#### Scenario: User enters new branch while adding project to todo

- **WHEN** TODO `修复登录问题` 已存在
- **AND** 用户为该 TODO 添加 Git 项目 `frontend-app`
- **AND** 用户在 `frontend-app` 的分支控件中输入 `feature/login-fix`
- **AND** 用户确认添加
- **THEN** TODO `修复登录问题` 下保存工程副本 `frontend-app`
- **AND** 该工程副本保存用户输入的分支值 `feature/login-fix`

### Requirement: Display Todo Task Terminals In Tree

系统 SHALL 在 TODO 工作树中显示任务级终端入口和任务级终端列表。任务级终端 SHALL 显示在 TODO 下、项目列表之前，并 SHALL 不归属于任何 TODO 项目。任务级终端的活动状态 SHALL 参与收起 TODO 行的聚合活动状态。

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

### Requirement: Provide Folder Actions For Todo Workspace

系统 SHALL 在 TODO 行下拉菜单中提供打开任务文件夹入口，并 SHALL 在 TODO 项目行下拉菜单中提供打开项目文件夹入口。系统 SHALL 仅在对应目录已创建时执行打开操作；目录不可用时 SHALL 显示非阻断错误。

#### Scenario: Todo menu includes open task folder

- **WHEN** TODO `修复登录问题` 显示在 TODO 工作树中
- **AND** 用户打开该 TODO 的下拉菜单
- **THEN** 菜单包含 `打开任务文件夹`

#### Scenario: Todo project menu includes open project folder

- **WHEN** TODO `修复登录问题` 下显示项目 `frontend-app`
- **AND** 用户打开该 TODO 项目的下拉菜单
- **THEN** 菜单包含 `打开项目文件夹`
