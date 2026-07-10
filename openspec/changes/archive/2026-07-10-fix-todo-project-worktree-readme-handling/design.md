## Context

当前 TODO 项目关联有两条后端入口：

- `AddProjectsToTodoWithBranches`：保存关联后，如果 TODO 已经是 `in-progress`，会调用 `prepareTodoWorkspace` 创建任务工作区、准备 Git worktree，并刷新状态。
- `AddProjectSelectionsToTodo`：下拉菜单使用该入口；它只保存结构化项目选择，不触发 `prepareTodoWorkspace`。

任务工作区准备流程目前会先创建任务目录并持久化 `WorkspaceDirName`，随后无条件尝试写入 `README.md`。因此没有关联项目、也没有初始化文件的 TODO 进入执行中后仍会产生只包含任务信息的 README。

## Goals / Non-Goals

**Goals:**

- 让下拉菜单添加项目和旧添加项目入口在执行中 TODO 上具备一致的 worktree 准备行为。
- 避免没有关联项目且没有其它落盘文件的 TODO 创建任务目录或 README。
- 保留无项目但选中了初始化文件的 TODO 写入初始化文件能力，同时不为它生成 README。
- 添加第一个项目到已执行中的 TODO 时，补创建任务工作区、准备 worktree，并生成 README。

**Non-Goals:**

- 不新增或改名 Wails 前端 API。
- 不改变项目候选、分支选择、终端创建和 worktree 清理策略。
- 不自动清理历史上已经创建的无项目 README 或任务目录。

## Decisions

### 1. 在后端统一执行中 TODO 的项目添加后处理

`AddProjectSelectionsToTodo` SHALL 在保存项目选择后复用 `AddProjectsToTodoWithBranches` 的后处理语义：如果目标 TODO 是 `in-progress`，调用 `prepareTodoWorkspace(todoID)`，然后重新加载最新状态并返回。

原因：下拉菜单只是 UI 入口差异，业务语义仍是“向 TODO 添加项目”。把逻辑放在后端可以覆盖当前和未来所有调用该 API 的前端入口。

替代方案：前端在下拉提交后额外调用准备接口。该方案需要新增 API 或暴露内部流程，并把业务一致性放到 UI 层，风险更高。

### 2. 将 README 创建条件绑定到“存在 TODO 项目”

README 的项目信息章节依赖 TODO project 的 base 分支和 worktree 分支。系统 SHALL 只在 TODO 至少关联一个项目时生成和维护 README。无项目 TODO 即使因为初始化文件创建了任务目录，也 SHALL NOT 生成 README。

原因：用户反馈的问题是没有添加项目时不需要 README；同时 README 的主要价值是记录项目/worktree 元数据。

替代方案：继续为无项目 TODO 写 README，但在内容中省略项目信息。该方案保留了当前噪声文件，不满足需求。

### 3. 将任务目录创建条件收紧为“存在落盘产物”

`prepareTodoWorkspace` SHALL 在无关联项目且无初始化文件快照时直接返回，不创建目录、不持久化 `WorkspaceDirName`。当无关联项目但存在初始化文件快照时，系统仍 SHALL 创建任务目录并写入初始化文件，但跳过 README。

原因：这避免空目录和无意义 README，同时不破坏初始化文件能力。

替代方案：无项目 TODO 永远不创建任务目录。该方案会破坏已存在的初始化文件规格，范围过大。

## Risks / Trade-offs

- [Risk] 历史无项目 TODO 可能已经存在 `WorkspaceDirName` 和 README。→ 本变更不做数据迁移；后续刷新无项目 TODO 时不再重写 README，历史文件由用户自行保留或清理。
- [Risk] `AddProjectSelectionsToTodo` 调用 `prepareTodoWorkspace` 后需要返回重新加载的状态，否则前端看不到 worktree 状态。→ 实现时沿用 `AddProjectsToTodoWithBranches` 的 reload 模式，并增加后端测试覆盖。
- [Risk] 无项目但有初始化文件的 TODO 将有任务目录但没有 README。→ 规格明确该行为，测试覆盖初始化文件存在且 README 不存在的场景。

