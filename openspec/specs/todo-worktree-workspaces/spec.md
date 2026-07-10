# todo-worktree-workspaces Specification

## Purpose
TBD - created by archiving change add-todo-worktree-workspaces. Update Purpose after archive.
## Requirements
### Requirement: Prepare Todo Workspace Directory

系统 SHALL 为执行中且存在落盘产物或需要任务目录执行生命周期脚本的 TODO 创建独立任务工作区目录。任务工作区目录 SHALL 位于当前 workspace 根目录下的任务工作区根目录中。任务工作区目录名 SHALL 在首次创建时由 TODO 标题和 TODO 描述计算 MD5 得到，并 SHALL 持久化到 TODO 数据中。TODO 标题或描述后续编辑 SHALL NOT 重命名已经创建的任务工作区目录。没有关联 TODO project、没有初始化文件快照且没有生命周期脚本的 TODO 进入执行中时，系统 SHALL NOT 创建任务工作区目录。

#### Scenario: Todo workspace directory is created when todo starts with project

- **WHEN** 用户将 TODO `修复登录问题` 从 `not-started` 标记为 `in-progress`
- **AND** TODO `修复登录问题` 已关联项目 `frontend-app`
- **THEN** 系统在当前 workspace 的任务工作区根目录下创建该 TODO 的任务工作区目录
- **AND** 该目录名由 TODO `修复登录问题` 的标题和描述计算 MD5 得到
- **AND** 系统将该目录名保存到 TODO 数据中

#### Scenario: Todo title edit does not rename workspace directory

- **WHEN** TODO `修复登录问题` 已创建任务工作区目录
- **AND** 用户将 TODO 标题修改为 `修复登录跳转问题`
- **THEN** 系统不重命名该 TODO 的任务工作区目录
- **AND** 该 TODO 继续引用原任务工作区目录

#### Scenario: Todo without projects and files does not create workspace directory

- **WHEN** 用户将 TODO `整理文档` 从 `not-started` 标记为 `in-progress`
- **AND** TODO `整理文档` 没有关联 TODO project
- **AND** TODO `整理文档` 没有初始化文件快照
- **AND** TODO `整理文档` 没有生命周期脚本
- **THEN** 系统不创建该 TODO 的任务工作区目录
- **AND** 系统不为该 TODO 保存任务工作区目录名

#### Scenario: Task terminal does not create workspace directory without inputs

- **WHEN** TODO `整理文档` 的状态为 `in-progress`
- **AND** TODO `整理文档` 没有关联 TODO project
- **AND** TODO `整理文档` 没有初始化文件快照
- **AND** TODO `整理文档` 没有生命周期脚本
- **WHEN** 用户请求创建该 TODO 的任务级终端
- **THEN** 系统不创建该 TODO 的任务工作区目录
- **AND** 系统不启动任务级终端
- **AND** 系统返回任务工作区尚未创建的错误

### Requirement: Maintain Todo Workspace Readme

系统 SHALL 在包含至少一个 TODO project 的任务工作区目录中维护 `README.md`。`README.md` SHALL 包含任务名称、任务详情和项目信息。任务详情为空时，系统 SHALL 在 `## 任务详情` 下保留空白内容。项目信息 SHALL 按 TODO 项目列出项目名称、base 分支和当前 worktree 分支。系统 SHALL 在任务工作区创建、TODO 标题或描述更新、TODO 项目增加或移除、TODO 项目分支信息变化后重新生成完整 `README.md`。未关联任何 TODO project 的 TODO SHALL NOT 生成 `README.md`。

#### Scenario: Readme is created with title description and project branches

- **WHEN** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** 该 TODO 下项目 `frontend-app` 的 base 分支为 `main`
- **AND** 该 TODO 下项目 `frontend-app` 的当前 worktree 分支为 `todo/fix-login/frontend-app`
- **AND** 系统创建该 TODO 的任务工作区目录
- **THEN** 任务工作区目录包含 `README.md`
- **AND** `README.md` 包含标题 `# 任务: 修复登录问题`
- **AND** `README.md` 包含章节 `## 任务详情`
- **AND** `README.md` 包含描述 `登录后跳回首页`
- **AND** `README.md` 包含项目行 `1. frontend-app: base分支为main, 当前worktree分支为todo/fix-login/frontend-app;`

#### Scenario: Empty description remains blank in readme

- **WHEN** TODO `修复登录问题` 的描述为空
- **AND** TODO `修复登录问题` 已关联项目 `frontend-app`
- **AND** 系统生成该 TODO 任务工作区的 `README.md`
- **THEN** `README.md` 包含章节 `## 任务详情`
- **AND** `## 任务详情` 下不写入占位文本

#### Scenario: Readme updates after todo edit

- **WHEN** TODO `修复登录问题` 已创建任务工作区目录
- **AND** TODO `修复登录问题` 已关联项目 `frontend-app`
- **AND** 用户将 TODO 描述修改为 `登录后回到原页面`
- **THEN** 系统重新生成该任务工作区目录中的 `README.md`
- **AND** `README.md` 包含描述 `登录后回到原页面`
- **AND** 任务工作区目录名保持不变

#### Scenario: Todo without projects does not create readme

- **WHEN** TODO `整理文档` 没有关联 TODO project
- **AND** 系统准备该 TODO 的任务工作区
- **THEN** 系统不生成该 TODO 的 `README.md`

#### Scenario: Removing last project removes generated readme

- **WHEN** TODO `修复登录问题` 已创建任务工作区目录
- **AND** TODO `修复登录问题` 下只关联项目 `frontend-app`
- **AND** 任务工作区目录中已存在系统生成的 `README.md`
- **WHEN** 用户移除 `frontend-app`
- **THEN** 系统移除该任务工作区目录中的系统生成 `README.md`
- **AND** 系统保留该任务工作区目录和其它用户文件

### Requirement: Create Project Worktrees Inside Todo Workspace

系统 SHALL 为执行中 TODO 的每个关联 Git 项目在任务工作区目录下创建 Git worktree。每个 TODO 项目 SHALL 保存 base 分支、当前 worktree 分支、worktree 路径和准备状态。用户在 TODO 项目分支输入框中选择或输入的分支 SHALL 只作为该 TODO 项目的 base 分支。若 TODO 已经处于执行中，用户后续新增关联 Git 项目时，系统 SHALL 为新增 TODO 项目创建 Git worktree。项目级终端 SHALL 在对应 TODO 项目 worktree 准备成功后可创建；若该 TODO 项目曾经准备成功但后续 worktree 路径或保存的 worktree 分支被清除，系统 SHALL 将该 TODO 项目记录为 worktree 已清除状态，并 SHALL 允许在原项目路径可用时继续创建项目级终端。worktree 准备失败 SHALL 保持失败状态，并 SHALL 阻止为该 TODO 项目创建项目终端。

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

#### Scenario: Adding project to in-progress todo creates worktree

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 该 TODO 尚未关联项目
- **AND** 用户将 Git 项目 `frontend-app` 关联到该 TODO
- **THEN** 系统创建该 TODO 的任务工作区目录
- **AND** 系统在任务工作区目录下创建 `frontend-app` 的 worktree
- **AND** 该 TODO 项目保存 worktree 路径和当前 worktree 分支

#### Scenario: Missing git repository records worktree failure

- **WHEN** TODO `修复登录问题` 进入 `in-progress`
- **AND** 该 TODO 关联项目 `docs-site`
- **AND** `docs-site` 路径不是 Git 仓库
- **THEN** 系统不为 `docs-site` 创建 worktree
- **AND** 该 TODO 项目保存 worktree 准备失败状态
- **AND** 系统阻止为该 TODO 项目创建项目终端

#### Scenario: Removed ready worktree path records cleared state

- **WHEN** TODO `修复登录问题` 下项目 `frontend-app` 已保存 ready worktree 路径
- **AND** 该 worktree 路径在磁盘上不存在
- **AND** 系统刷新该 TODO project 的 Git 状态或用户请求创建项目终端
- **THEN** 系统将该 TODO project 保存为 worktree 已清除状态
- **AND** 系统不将该 TODO project 保存为 worktree 准备失败状态

#### Scenario: Removed ready worktree branch records cleared state

- **WHEN** TODO `修复登录问题` 下项目 `frontend-app` 已保存 ready worktree 分支 `todo/fix-login/frontend-app`
- **AND** 系统确认该 worktree 分支已不存在
- **THEN** 系统将该 TODO project 保存为 worktree 已清除状态
- **AND** 系统不将该 TODO project 保存为 worktree 准备失败状态

#### Scenario: Cleared worktree project remains terminal-eligible when source path is available

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下项目 `frontend-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目路径可用
- **THEN** 系统允许为该 TODO project 创建项目终端

#### Scenario: Cleared worktree project with missing source path cannot create terminal

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下项目 `frontend-app` 的 worktree 状态为 cleared
- **AND** 该 TODO project 保存的原项目路径不可用
- **THEN** 系统阻止为该 TODO project 创建项目终端

### Requirement: Open Todo Workspace Folders

系统 SHALL 允许用户从 TODO 行菜单打开任务工作区文件夹，并允许用户从 TODO 项目行菜单打开项目 worktree 文件夹。打开文件夹 SHALL 使用系统文件管理器。目录不存在或不可访问时，系统 SHALL 显示不改变工作区状态的错误信息。

#### Scenario: User opens todo workspace folder

- **WHEN** TODO `修复登录问题` 已创建任务工作区目录
- **AND** 用户从 TODO 行菜单选择 `打开任务文件夹`
- **THEN** 系统使用系统文件管理器打开该 TODO 的任务工作区目录

#### Scenario: User opens todo project worktree folder

- **WHEN** TODO `修复登录问题` 下项目 `frontend-app` 已创建 worktree
- **AND** 用户从该 TODO 项目行菜单选择 `打开项目文件夹`
- **THEN** 系统使用系统文件管理器打开该 TODO 项目的 worktree 目录

#### Scenario: Opening missing folder reports error

- **WHEN** TODO `修复登录问题` 尚未创建任务工作区目录
- **AND** 用户请求打开该 TODO 的任务文件夹
- **THEN** 系统不打开系统文件管理器
- **AND** 系统显示不会改变当前工作区状态的错误信息

### Requirement: Preserve Todo Workspace Until Manual Cleanup

系统 SHALL 在 TODO 完成、删除或 workspace 关闭时保留任务工作区目录和项目 worktree，除非用户显式执行清理操作。本变更 SHALL NOT 自动删除任务工作区目录或 Git worktree。

#### Scenario: Completing todo keeps workspace directory

- **WHEN** TODO `修复登录问题` 已创建任务工作区目录
- **AND** 用户将 TODO `修复登录问题` 标记为完成
- **THEN** 系统保留该任务工作区目录
- **AND** 系统保留该任务工作区目录下的项目 worktree

#### Scenario: Closing workspace keeps todo workspace directories

- **WHEN** workspace `/work/customer-a` 中存在 TODO 任务工作区目录
- **AND** 用户关闭当前 workspace
- **THEN** 系统不删除任何 TODO 任务工作区目录
- **AND** 系统不删除任何 TODO 项目 worktree

