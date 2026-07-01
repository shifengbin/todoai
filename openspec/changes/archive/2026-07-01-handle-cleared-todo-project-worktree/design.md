## Context

TODO project 当前保存 `worktreeBranch`、`worktreePath` 和 `worktreeStatus`，其中状态只有 `pending`、`ready`、`failed`。项目行分支显示已经改为通过 `GetTodoProjectGitStatus` 查询 ready worktree 的真实 Git 分支；当 worktree 路径不可用时前端会清掉分支后缀。项目终端创建路径则更严格：`CreateTodoTerminal` 会要求 worktree 状态为 ready 且目录存在，否则拒绝创建 shell。

这个行为和新需求冲突：worktree 被外部清理后，用户需要在任务列表项目 item 上看到明确标记，并且仍能用原项目目录打开终端。该场景不等同于 worktree 准备失败，因为失败表示系统从未成功创建隔离 worktree，仍应阻止项目终端。

## Goals / Non-Goals

**Goals:**

- 为 TODO project 表达“worktree 已清除”的稳定状态，区别于 `failed`。
- 项目行在原分支后缀位置显示 `worktree已清除`。
- 对已清除 worktree 的 `in-progress` TODO project，交互式项目终端和背景启动配置使用保存的原项目目录作为工作目录。
- 保持 `not-started` TODO、项目路径不可用和 worktree 准备失败的拒绝行为。

**Non-Goals:**

- 不新增自动恢复或重新创建已清除 worktree 的功能。
- 不自动删除 Git worktree 或任务工作区目录。
- 不改变 completed TODO 快照的合并确认规则。
- 不改变“打开项目文件夹”的行为，除非实现阶段明确决定把它纳入同一 fallback 规则。

## Decisions

1. **新增 `cleared` worktree 状态。**

   - 选择：在 Go 模型中新增 `WorktreeStatusCleared = "cleared"`，并通过现有 `worktreeStatus` 字段暴露到前端。
   - 原因：`failed` 已表示准备失败，复用会导致 UI 禁用终端并显示错误；空状态又会触发重新准备 worktree，不符合“已被清理”的语义。
   - 备选：仅在前端通过 Git status 查询结果派生标记。该方式无法稳定持久化，也难以让终端创建入口共享同一判断。

2. **清除状态以惰性检测为主。**

   - 选择：在读取 Git 状态、创建项目终端、启动背景项目命令等会触碰 worktree 的入口检测 ready worktree 是否仍可用；若 ready worktree 路径不存在或保存的 worktree 分支确认不存在，则记录为 `cleared`。
   - 原因：worktree 可能被外部删除，应用无法可靠实时监听所有文件系统和 Git 分支变化。惰性检测与现有分支刷新、终端创建路径契合。
   - 备选：每次 `ListProjects` 都扫描所有 worktree。该方式会让普通状态加载承担较重磁盘和 Git 查询成本。

3. **终端工作目录通过单一 helper 解析。**

   - 选择：新增或抽取 `todoProjectTerminalWorkingDir(todoProject, project)` 之类的后端 helper：ready 且 worktree 目录存在时返回 `worktreePath`；cleared 时返回 `project.Path`；pending/failed/不可用路径返回错误。
   - 原因：交互式终端、背景启动配置、shell manager 注册终端都需要一致的 cwd 选择，避免不同入口出现一处 fallback、一处报错。
   - 备选：只在 `CreateTodoTerminal` 里改路径。这样背景启动配置仍会使用旧的 worktree 强校验，行为不一致。

4. **前端显示优先级明确。**

   - 选择：项目行显示顺序为：项目不可用状态独立显示；worktree failed 显示失败错误；worktree cleared 或 Git 状态刷新确认清除时，名称后缀显示 `worktree已清除`；ready 且拿到真实分支时显示真实分支。
   - 原因：用户提出“显示分支的位置标记 worktree 已清除”，因此 `frontend-app(worktree已清除)` 比额外状态行更贴近现有布局；失败错误仍保留在状态行，避免把错误信息挤进名称。
   - 备选：只在项目名下方加状态文本。该方式不满足“分支的位置”这个 UI 要求。

## Risks / Trade-offs

- **风险：清除状态检测滞后** → 通过分支刷新、终端创建和背景命令入口都执行检测，覆盖用户最常触发的路径。
- **风险：fallback 到原项目目录会打破 worktree 隔离** → 只允许在明确的 `cleared` 状态下 fallback，并在规格中保留 failed/pending 的拒绝行为；UI 标记让用户知道当前不在隔离 worktree。
- **风险：旧数据中 ready 状态但目录缺失** → 不做一次性迁移，首次触碰时标记为 cleared，避免打开 workspace 时大量扫描。
- **风险：前端短时间内只拿到 Git status 而未刷新 ProjectState** → Git status 返回或前端分支状态缓存需要携带清除信号，立即显示 `worktree已清除`；后端再持久化状态用于后续加载。
