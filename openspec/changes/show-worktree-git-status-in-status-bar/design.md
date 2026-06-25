## Context

当前应用已经将 TODO project 的项目级终端切换到任务 worktree 目录运行，后端 `TodoProject` 也保存了 `worktreePath`、`worktreeStatus` 和 `worktreeBranch`。但状态栏仍通过 `GetProjectGitStatus(projectID)` 查询原项目路径，前端也只用 `projectId` 作为 Git 状态请求和缓存匹配 key。

这会导致同一个来源项目在多个 TODO 中各自拥有 worktree 时，状态栏展示来源项目目录的分支和改动数量，而不是当前 TODO project 所在 worktree 的真实 Git 状态。

## Goals / Non-Goals

**Goals:**

- 当前上下文是 TODO project 且 worktree ready 时，状态栏查询并展示该 TODO project 的 worktree Git 状态。
- 当前上下文不是 TODO project 时，保留普通项目 Git 状态查询行为。
- 同一来源项目的不同 TODO worktree 使用不同前端状态 key，避免请求去重、过期响应和显示状态互相污染。
- worktree 缺失、未准备或路径不可用时，状态栏显示稳定不可用状态，不启动错误路径上的 Git 查询。

**Non-Goals:**

- 不改变 Git porcelain v2 解析逻辑。
- 不增加提交、暂存、拉取、推送或 worktree 清理操作。
- 不迁移已持久化的 TODO project 数据。
- 不改变完成 TODO 的 merge status 检查逻辑。

## Decisions

### 使用显式 Git status context，而不是只传 projectId

前端应根据当前选中上下文构造 Git 状态上下文：

- 普通项目上下文：`type=project`，使用 `projectId`。
- TODO project 上下文：`type=todo-project`，使用 `todoProjectId`，后端从 TODO project 读取 `worktreePath`。

前端状态匹配、in-flight 去重、focus 去重和过期响应判断都使用 context key，例如 `project:project-a` 或 `todo-project:todo-project-a`。

备选方案是继续调用 `GetProjectGitStatus(projectId)`，但让后端从当前 active TODO project 推断 worktree。该方案会把查询结果依赖隐藏在服务端全局选择状态里，也无法清晰表达并发请求来自哪个 TODO project，测试和过期响应处理更脆弱。

### 后端新增 TODO project Git 状态查询入口

后端保留现有 `GetProjectGitStatus(projectID)`，新增面向 TODO project 的查询入口，例如 `GetTodoProjectGitStatus(todoProjectID)`。该方法：

1. 校验 workspace 已打开。
2. 查找 TODO project。
3. 若 TODO project 不可用、worktree 未 ready 或 `worktreePath` 为空，返回带 `pathUnavailable` 或稳定不可用语义的 `GitStatus`。
4. 若 `worktreePath` 可用，则对 `worktreePath` 执行现有 `gitStatus` 查询。
5. 返回的 `projectId` 继续用于来源 project ID 兼容展示，同时可以增加上下文字段或由前端用请求 context key 绑定结果。

备选方案是把 `GetProjectGitStatus` 改为接收可选 path 或 todoProjectId。直接传 path 会让前端获得更大的文件系统查询能力，不适合 Wails API 边界；改动现有签名也会扩大绑定和测试影响。

### 前端显示路径使用 worktree path

在 TODO project 上下文中，工作区标题和状态栏判断应优先使用 `todoProject.worktreePath`，当 worktree 未准备时再展示 TODO project 的原始 `path` 或不可用状态。这样标题、终端 cwd 和状态栏 Git 信息保持一致。

该选择不改变 TODO project copy 的持久化 `path` 含义；`path` 仍表示来源项目路径，`worktreePath` 表示执行态任务目录。

### 初始化 Git 仍只面向普通项目路径

状态栏 `Initialize Git Repository` 行为保持用于当前可用项目路径。对于 TODO project worktree 上下文，如果 worktree 未准备或不是 Git 仓库，优先显示不可用或非仓库状态，不在本变更中新增“初始化 worktree”为 Git 仓库的流程。正常 worktree 应由 worktree 准备流程创建为 Git 仓库。

## Risks / Trade-offs

- [Risk] 前端同时存在 `projectId` 和 `todoProjectId` 后，旧的状态匹配条件遗漏某处会导致状态串显示。→ Mitigation: 集中创建 Git status context 和 context key，所有刷新入口复用同一 helper，并添加同源项目不同 TODO worktree 的前端测试。
- [Risk] worktree 目录被用户手动删除后，直接执行 Git 查询会产生普通失败。→ Mitigation: 后端在查询前检查 `worktreePath` 可用性，不可用时返回路径不可用状态。
- [Risk] 新增 Wails API 后前端绑定不同步。→ Mitigation: 实现时重新生成或同步更新 Wails 绑定，并用前端测试调用新增入口。
- [Risk] 状态栏在 worktree 准备失败时可能从“原项目 Git 状态”变为“worktree 不可用”。→ Mitigation: 这是符合用户当前上下文的行为；失败原因仍由 TODO project 行展示，状态栏只表达当前 Git 状态不可查询。

## Migration Plan

不需要数据迁移。已有 TODO project 的 `worktreePath` 和 `worktreeStatus` 字段继续使用。回滚时删除新增查询入口和前端 context key 调整，状态栏会恢复为按来源项目路径查询。

## Open Questions

无。
