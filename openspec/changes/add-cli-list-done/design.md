## Context

TodoAI 当前是 Wails 桌面应用，`main.go` 只在 `todoai claude-hook` 时绕过 GUI，其它参数会进入 Wails 启动流程。工作区状态保存在 TodoAI workspace 的 `.data/projects.json` 中，`Todo` 记录包含 `status`、`title` 和完成时保存的 `projectSnapshots`。完成任务时，系统已把关联项目的名称、路径、base 分支、worktree 分支写入快照。

新命令需要从项目仓库或其子目录运行，因此不能只依赖当前 GUI 打开的 workspace。命令应复用 TodoAI recent workspaces 和 workspace project state 来定位当前目录对应的项目。

## Goals / Non-Goals

**Goals:**

- 让 `todoai list --done` 在命令行中运行，不启动 Wails GUI。
- 支持在已登记项目根目录和任意子目录执行。
- 只列出当前项目相关且状态为 `completed` 的 TODO。
- 输出任务名称、worktree 分支、base 分支，并对缺失分支使用稳定占位符。
- 通过 Go 自动化测试覆盖路由、匹配、过滤、输出和错误状态。

**Non-Goals:**

- 不新增任务完成状态或新的持久化数据格式。
- 不在用户项目仓库中写入 TodoAI 索引文件。
- 不检查 worktree 分支是否已合并到 base 分支。
- 不改变桌面端已完成视图和现有 Wails API。
- 不为本次命令提供交互式 UI。

## Decisions

1. CLI 路由在 `main.go` 中先于 Wails 启动处理。

   `todoai list --done` 与现有 `todoai claude-hook` 一样属于无 GUI 命令。`main()` 应在创建 `App` 和调用 `wails.Run` 之前识别该命令，执行后用返回码退出。这样可以避免 CLI 查询误启动桌面窗口。

   备选方案是在 Wails app 初始化后复用 `App` 方法，但这会引入 GUI 生命周期和运行时依赖，不适合简单 CLI 查询。

2. 当前目录通过 recent workspaces 反查项目归属。

   命令从当前工作目录开始，读取 TodoAI 配置目录中的 `recent-workspaces.json`，遍历可用 workspace 的 `.data/projects.json`。若当前目录等于某个项目路径，或位于该项目路径子目录中，则视为匹配该项目。为兼容已完成快照位于 worktree 路径的情况，匹配时也应考虑 completed TODO 的 `projectSnapshots[].path`。

   备选方案是只向上查找 `.data/projects.json`，但这只能覆盖 workspace 目录内部，无法满足从项目仓库子目录执行。另一个方案是在项目仓库写元数据索引，但侵入性过高。

3. 输出数据来自 completed TODO 的项目快照。

   命令过滤 `Todo.Status == completed`，再从每个 TODO 的 `ProjectSnapshots` 中选择当前项目匹配的快照。输出 `Todo.Title`、`snapshot.WorktreeBranch`、`snapshot.BaseBranch`。缺失分支显示为 `-`，但不丢弃该 completed TODO。

   不使用 `TodoProjects`，因为完成任务后关联会从打开任务列表中移除；`ProjectSnapshots` 才是完成时的历史记录。

4. 输出使用稳定 JSON 数组。

   成功时 stdout 始终输出 JSON 数组。数组元素包含 `taskName`、`worktreeBranch`、`baseBranch`。没有匹配到 completed TODO 时返回成功并输出 `[]`；当前目录无法匹配任何 TodoAI 项目时返回失败并在 stderr 输出错误信息。

## Risks / Trade-offs

- Recent workspace 数据过期或缺失 -> 命令无法定位项目；错误信息应明确提示当前目录不是已知 TodoAI 项目。
- 同一路径可能被多个 workspace 记录 -> 使用 recent workspaces 的现有排序优先级，选择最近打开的匹配项，保持行为可预测。
- 历史 completed TODO 可能缺少快照或分支字段 -> 保留可识别记录，缺失字段显示 `-`。
- JSON 数组比表格更适合命令行复用，但对人工直接阅读略少格式化；使用缩进输出降低阅读成本。
