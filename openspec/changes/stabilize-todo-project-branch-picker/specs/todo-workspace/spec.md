## MODIFIED Requirements

### Requirement: Select Branch For Todo Projects

系统 SHALL 在创建 TODO、编辑 TODO 工程关联和为 TODO 添加项目时，允许用户为每个选中的项目选择或输入用于创建任务 worktree 的分支。分支控件 SHALL 默认使用主分支。用户选择或输入的分支信息 SHALL 保存到 TODO 工程副本中，并 SHALL 用于后续创建任务项目 worktree。分支候选 SHALL 作为输入辅助而不是提交前置条件；当候选数量较多、候选加载失败或候选不可用时，系统 SHALL 保持分支输入和表单提交流程可用。系统 MUST NOT 依赖会一次性渲染全部分支候选的原生下拉控件。

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

#### Scenario: Large branch lists remain stable

- **WHEN** Git 项目 `frontend-app` 返回大量本地和远端分支候选
- **AND** 用户在 TODO 项目分支控件中输入筛选文本
- **THEN** 系统只渲染与当前输入匹配的有限数量候选
- **AND** 分支输入框保持可编辑
- **AND** 应用不因候选数量过多而卡死或闪退

#### Scenario: Branch list loading failure allows manual input

- **WHEN** 用户为 TODO 选择 Git 项目 `frontend-app`
- **AND** 系统加载 `frontend-app` 的分支候选失败或超时
- **THEN** 系统保持 `frontend-app` 的分支输入框可编辑
- **AND** 用户可以输入 `feature/manual-branch`
- **AND** 用户提交后该 TODO 工程副本保存分支值 `feature/manual-branch`

#### Scenario: Editing todo uses stable branch picker

- **WHEN** TODO `修复登录问题` 已关联 Git 项目 `frontend-app`
- **AND** 用户打开 TODO 详情编辑
- **AND** 用户在 `frontend-app` 的分支控件中选择或输入 `release/2026`
- **AND** 用户保存 TODO 详情
- **THEN** 该 TODO 工程副本保存 base 分支选择 `release/2026`
