# todo-initialization-files Specification

## Purpose
TBD - created by archiving change add-todo-initialization-files. Update Purpose after archive.
## Requirements
### Requirement: Manage Global Todo Initialization File Templates

系统 SHALL 允许用户通过菜单栏“全局管理 > 文件管理”维护 TODO 初始化文件模板。每个模板记录 SHALL 包含显示名称、描述、文件名、文本内容和是否默认选择。系统 SHALL 通过用户上传文本文件设置模板记录的文件名和文本内容。系统 SHALL 持久化这些模板，并在应用重启或 workspace 切换后继续可用。系统 SHALL 允许不同模板记录使用相同显示名称。系统 SHALL 拒绝保存空显示名称、空文件名、重复文件名、绝对路径文件名、包含路径穿越的文件名和包含目录分隔符的文件名。系统 SHALL NOT 在终端 Settings 弹窗中展示或保存 TODO 初始化文件模板管理表单。

#### Scenario: User saves initialization file templates from file management

- **WHEN** 用户打开菜单栏“全局管理 > 文件管理”
- **AND** 用户保存两个初始化文件模板
- **AND** 第一个模板显示名称为 `Agent Rules`、描述为 `任务执行约束`、上传文件名为 `AGENTS.md`、上传内容为 `请先阅读任务说明`、默认选择为 true
- **AND** 第二个模板显示名称为 `Prompt`、描述为 `可选提示词`、上传文件名为 `prompt.md`、上传内容为 `生成实现计划`、默认选择为 false
- **THEN** 系统持久化这两个模板
- **AND** 后续读取全局设置时返回相同的名称、描述、文件名、内容和默认选择状态

#### Scenario: Duplicate display names are saved as separate records

- **WHEN** 用户打开菜单栏“全局管理 > 文件管理”
- **AND** 用户保存两个显示名称都为 `Prompt` 的初始化文件模板
- **AND** 第一个模板上传文件名为 `prompt.md`
- **AND** 第二个模板上传文件名为 `notes.md`
- **THEN** 系统持久化这两个模板记录
- **AND** 两个记录不会因为显示名称相同而互相覆盖

#### Scenario: Invalid initialization file template filename is rejected

- **WHEN** 用户在文件管理中保存文件名为空、为绝对路径、包含 `..` 路径穿越或包含目录分隔符的初始化文件模板
- **THEN** 系统拒绝保存该配置
- **AND** 系统返回明确的校验错误

#### Scenario: Duplicate initialization file template filenames are rejected

- **WHEN** 用户在文件管理中保存两个文件名都为 `AGENTS.md` 的初始化文件模板
- **THEN** 系统拒绝保存该配置
- **AND** 系统提示初始化文件名不能重复

### Requirement: Select Initialization Files When Creating Todo

系统 SHALL 在创建 TODO 时展示全局初始化文件模板。每个模板 SHALL 展示名称、描述和文件名。默认选择为 true 的模板 SHALL 在创建 TODO 表单中自动勾选，默认选择为 false 的模板 SHALL 保持未勾选并允许用户手动勾选。系统 SHALL 在创建 TODO 时保存用户选中模板的文件快照，快照 SHALL 包含名称、描述、文件名和文本内容。

#### Scenario: Default templates are preselected in create todo form

- **WHEN** 全局文件管理中存在一个默认选择的初始化文件模板 `AGENTS.md`
- **AND** 存在一个非默认选择的初始化文件模板 `prompt.md`
- **AND** 用户打开创建 TODO 表单
- **THEN** 系统展示两个模板的名称、描述和文件名
- **AND** `AGENTS.md` 模板自动勾选
- **AND** `prompt.md` 模板未勾选

#### Scenario: User manually selects optional initialization file

- **WHEN** 用户打开创建 TODO 表单
- **AND** `prompt.md` 模板未默认勾选
- **AND** 用户手动勾选 `prompt.md`
- **AND** 用户提交创建 TODO
- **THEN** 系统创建 TODO
- **AND** 该 TODO 保存 `prompt.md` 的初始化文件快照

#### Scenario: Todo stores selected file snapshot

- **WHEN** 用户创建 TODO 时选中模板 `AGENTS.md`
- **AND** 模板创建时的内容为 `旧内容`
- **AND** TODO 创建完成后用户把全局模板 `AGENTS.md` 的内容修改为 `新内容`
- **THEN** 该 TODO 保存的初始化文件快照仍包含文件名 `AGENTS.md`
- **AND** 该 TODO 保存的初始化文件快照内容仍为 `旧内容`

### Requirement: Write Selected Initialization Files Into Todo Workspace

系统 SHALL 将 TODO 保存的初始化文件快照写入任务文件夹根目录。每个初始化文件 SHALL 使用快照中的文件名创建，文件内容 SHALL 等于快照中的文本内容。若任务文件夹中已经存在同名初始化文件，系统 SHALL NOT 覆盖该文件。对于包含关联 TODO project 的 TODO，系统 MUST 在该 TODO 的所有关联 TODO project worktree 都创建完成且状态为 ready 后，才写入初始化文件，并 SHALL 继续生成和维护 `README.md`。若任一关联 TODO project worktree 尚未 ready、正在准备或准备失败，系统 SHALL 延迟写入初始化文件。对于没有关联 TODO project 的 TODO，系统 SHALL 在任务文件夹创建后写入初始化文件，但 SHALL NOT 生成 `README.md`。对于没有关联 TODO project 且没有选中初始化文件的 TODO，系统 SHALL NOT 只为 `README.md` 创建任务文件夹。

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

#### Scenario: Todo without associated projects writes initialization files without readme

- **WHEN** TODO `整理文档` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `请先阅读任务说明`
- **AND** TODO `整理文档` 没有关联 TODO project
- **AND** 用户将该 TODO 标记为 `in-progress`
- **THEN** 系统创建该 TODO 的任务文件夹
- **AND** 任务文件夹中包含 `AGENTS.md`
- **AND** 任务文件夹中不包含 `README.md`

#### Scenario: Existing initialization file is not overwritten

- **WHEN** TODO `修复登录问题` 保存了初始化文件快照，文件名为 `AGENTS.md`，内容为 `模板内容`
- **AND** 该 TODO 的所有关联 TODO project worktree 均已 ready
- **AND** 该 TODO 的任务文件夹中已经存在 `AGENTS.md`
- **AND** 现有 `AGENTS.md` 内容为 `用户修改内容`
- **AND** 系统再次确保该 TODO 的任务文件夹存在
- **THEN** 系统不覆盖 `AGENTS.md`
- **AND** `AGENTS.md` 的内容仍为 `用户修改内容`

#### Scenario: Todo without associated projects and selected initialization files creates no task files

- **WHEN** 用户创建 TODO 时未选择任何初始化文件模板
- **AND** TODO 没有关联 TODO project
- **AND** 用户将该 TODO 标记为 `in-progress`
- **THEN** 系统不创建该 TODO 的任务文件夹
- **AND** 系统不生成 `README.md`
- **AND** 系统不额外创建初始化文件

