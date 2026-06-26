## Context

当前应用使用 Vue/Wails 前端展示 TODO 工作树，后端已有 `GetTodoProjectGitStatus` 可按 TODO project 的 worktree 路径查询真实 Git 状态。状态栏当前只维护激活上下文的 Git 状态；左侧 `ProjectSidebar` 只接收 TODO/project/terminal 数据，不知道每个 TODO project 的当前 Git 分支。

TODO 初始化文件当前在任务工作区创建或 README 刷新路径中写入，顺序没有被规格约束。新的要求是初始化必须在该 TODO 关联的 worktree 都创建完成后执行，避免初始化在 worktree 尚未准备好时提前发生。

## Goals / Non-Goals

**Goals:**

- 左侧 TODO 项目行显示 `项目名称(当前真实分支)`。
- 分支显示以 worktree 当前 Git 状态为准，并在相关 worktree 终端命令结束后刷新。
- 无激活项目上下文时，底部状态栏不显示 `No project` 或其他 Git 状态 chip。
- TODO 初始化文件写入晚于该 TODO 的所有关联 worktree ready 状态。

**Non-Goals:**

- 不修改顶部工作区标题。
- 不把创建 worktree 时保存的 `worktreeBranch` 当作实时分支来源。
- 不新增后端 Git 状态 API，除非现有 `GetTodoProjectGitStatus` 在实现中无法满足列表级查询。
- 不改变初始化文件模板管理、选择和快照保存规则。

## Decisions

1. 在 `App.vue` 中维护按 `todoProjectId` 索引的轻量 Git 状态缓存，并把分支信息作为 props 传给 `ProjectSidebar`。

   备选方案是让 `ProjectSidebar` 自己调用 Wails API。该方案会让展示组件持有后端调用和去重逻辑，破坏现有数据流。状态缓存留在 `App.vue` 更符合当前架构。

2. 只为需要展示的 TODO project 刷新分支，并复用 `GetTodoProjectGitStatus(todoProjectId)`。

   列表中可能存在多个 TODO project，全部无差别轮询会增加后台 Git 命令数量。实现应在展开 TODO、选择 TODO project、窗口 focus、相关终端 command-end 等明确时机刷新可见或相关项。

3. 命令结束刷新按终端归属判断，而不是只刷新当前 active terminal。

   当命令来自某个 TODO project 的终端时，只刷新该 TODO project 的列表分支缓存；如果该 TODO project 同时是当前激活上下文，再刷新状态栏。这样可以在用户从 worktree 终端切换分支后，让左侧列表及时显示最新分支。

4. 无激活项目上下文时保留状态栏布局，但隐藏 Git 状态 chip。

   这满足“不显示 git 状态”的要求，同时避免终端区域因为状态栏高度变化而抖动。

5. TODO 级控制台使用 TODO 任务文件夹根目录作为独立 Git 状态上下文。

   选择 TODO 级控制台时，状态栏不能沿用上一个项目或 TODO project 的 Git 状态。后端查询只承认 TODO 任务文件夹根目录自己的 `.git` 元数据；如果只有子目录是 Git 仓库，或任务文件夹本身不是 Git 仓库，则前端不显示 Git 状态 chip。

6. 初始化文件写入只在 worktree 准备完成后执行。

   对包含 TODO project 的 TODO，初始化文件写入应在所有关联 TODO project 的 worktree 状态均为 `ready` 后执行。若存在 pending 或 failed worktree，则延迟初始化文件写入，直到后续准备流程让全部关联 worktree 达到 `ready`。没有关联项目的 TODO 没有 worktree 前置条件，可在任务工作区创建后写入初始化文件。

## Risks / Trade-offs

- Git 状态查询数量增加 -> 只刷新可见或相关 TODO project，并对同一 TODO project 的并发请求去重。
- 分支查询失败 -> 左侧项目名应退回只显示项目名称，不显示错误文案以免污染树结构；状态栏仍按现有错误规则展示。
- TODO 级控制台可能处于非 Git 任务目录 -> 状态栏隐藏 Git chip，避免把子目录 worktree 或上一个项目状态误认为当前 TODO 状态。
- worktree 失败会延迟初始化文件 -> 用户需要先修复 worktree 准备问题；这符合“worktree 都创建好后执行”的顺序约束。
- 初始化写入从多个调用点收敛后可能影响现有任务文件夹创建测试 -> 增加覆盖“worktree ready 后写入”和“未全 ready 不写入”的后端测试。
