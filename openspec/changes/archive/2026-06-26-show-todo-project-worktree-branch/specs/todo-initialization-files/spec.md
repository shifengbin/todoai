## MODIFIED Requirements

### Requirement: Write Selected Initialization Files Into Todo Workspace

系统 SHALL 将 TODO 保存的初始化文件快照写入任务文件夹根目录。每个初始化文件 SHALL 使用快照中的文件名创建，文件内容 SHALL 等于快照中的文本内容。系统 SHALL 继续生成和维护 `README.md`。若任务文件夹中已经存在同名初始化文件，系统 SHALL NOT 覆盖该文件。对于包含关联 TODO project 的 TODO，系统 MUST 在该 TODO 的所有关联 TODO project worktree 都创建完成且状态为 ready 后，才写入初始化文件。若任一关联 TODO project worktree 尚未 ready、正在准备或准备失败，系统 SHALL 延迟写入初始化文件。对于没有关联 TODO project 的 TODO，系统 SHALL 在任务文件夹创建后写入初始化文件。

#### Scenario: Selected initialization files are created after all worktrees are ready

- **WHEN** TODO `修复登录问题` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `请先阅读任务说明`
- **AND** TODO `修复登录问题` 关联项目 `frontend-app` 和 `api-service`
- **AND** 用户将该 TODO 标记为 `in-progress`
- **AND** 系统已经为 `frontend-app` 和 `api-service` 创建 ready worktree
- **THEN** 系统创建该 TODO 的任务文件夹
- **AND** 任务文件夹中包含 `AGENTS.md`
- **AND** `AGENTS.md` 的内容为 `请先阅读任务说明`
- **AND** 任务文件夹中仍包含系统生成的 `README.md`

#### Scenario: Initialization files wait until all worktrees are ready

- **WHEN** TODO `修复登录问题` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `请先阅读任务说明`
- **AND** TODO `修复登录问题` 关联项目 `frontend-app` 和 `api-service`
- **AND** `frontend-app` 的 TODO project worktree 状态为 ready
- **AND** `api-service` 的 TODO project worktree 尚未 ready 或准备失败
- **THEN** 系统不写入 `AGENTS.md`
- **WHEN** `api-service` 的 TODO project worktree 后续达到 ready
- **THEN** 系统写入 `AGENTS.md`

#### Scenario: Todo without associated projects writes initialization files after task workspace creation

- **WHEN** TODO `整理文档` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `请先阅读任务说明`
- **AND** TODO `整理文档` 没有关联 TODO project
- **AND** 用户将该 TODO 标记为 `in-progress`
- **THEN** 系统创建该 TODO 的任务文件夹
- **AND** 任务文件夹中包含 `AGENTS.md`

#### Scenario: Existing initialization file is not overwritten

- **WHEN** TODO `修复登录问题` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `模板内容`
- **AND** 该 TODO 的所有关联 TODO project worktree 均已 ready
- **AND** 该 TODO 的任务文件夹中已经存在 `AGENTS.md`
- **AND** 现有 `AGENTS.md` 内容为 `用户修改内容`
- **AND** 系统再次确保该 TODO 的任务文件夹存在
- **THEN** 系统不覆盖 `AGENTS.md`
- **AND** `AGENTS.md` 的内容仍为 `用户修改内容`

#### Scenario: Todo without selected initialization files only creates readme

- **WHEN** 用户创建 TODO 时未选择任何初始化文件模板
- **AND** 用户将该 TODO 标记为 `in-progress`
- **AND** 该 TODO 的所有关联 TODO project worktree 均已 ready
- **THEN** 系统创建该 TODO 的任务文件夹
- **AND** 系统生成 `README.md`
- **AND** 系统不额外创建初始化文件
