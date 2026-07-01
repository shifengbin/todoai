## 1. 数据模型与后端状态判断

- [x] 1.1 为 `TodoProjectSnapshot` 增加可选的持久化状态字段，记录已确认状态和确认原因，并保持旧 `projects.json` 兼容。
- [x] 1.2 扩展 completed merge status 请求结构，携带 `todoId`、`snapshotIndex` 和快照 fingerprint，避免异步旧请求写错快照。
- [x] 1.3 增加 Git/worktree 检查逻辑，区分已合并、未合并、worktree 路径不存在、worktree 分支不存在和 unknown 错误。
- [x] 1.4 在 `GetCompletedTodoProjectMergeStatuses` 流程中实现幂等写回：仅当快照仍匹配且状态可确认为对号时持久化。

## 2. 前端展示与查询跳过

- [x] 2.1 更新 Wails 前端模型，暴露 completed 项目快照和请求结构的新字段。
- [x] 2.2 调整 `completedMergeStatusEntries` 和刷新逻辑，优先使用快照持久化确认状态，并跳过这些快照的后端查询。
- [x] 2.3 调整已完成项目快照状态映射，让持久化确认、真实合并、worktree 目录清理和 worktree 分支清理都显示对号。
- [x] 2.4 保持未合并、Git 不可用、非 Git 仓库、超时和历史分支信息缺失时显示黄色警告。

## 3. 后端测试

- [x] 3.1 添加 Go 单元测试，覆盖 worktree 路径不存在时返回对号状态并写回 completed 快照。
- [x] 3.2 添加 Go 单元测试，覆盖 worktree 分支不存在时返回对号状态并写回 completed 快照。
- [x] 3.3 添加 Go 单元测试，覆盖已合并状态写回、未合并状态不写回、unknown 状态不写回。
- [x] 3.4 添加 Go 单元测试，覆盖 fingerprint 不匹配或 TODO 不是 completed 时不写回。

## 4. 前端测试与质量检查

- [x] 4.1 添加客户端自动化测试，确认持久化确认状态的 completed 快照直接显示对号且不调用 `GetCompletedTodoProjectMergeStatuses`。
- [x] 4.2 添加客户端自动化测试，确认新查询请求包含 `todoId`、`snapshotIndex` 和 fingerprint。
- [x] 4.3 添加客户端自动化测试，确认未合并和缺少分支信息仍显示黄色警告。
- [x] 4.4 执行 Go 与前端测试套件，确认相关回归测试通过。
- [x] 4.5 执行自动 review，检查实现是否符合设计、规格和现有代码风格。

## 5. 验证与打包

- [x] 5.1 运行 OpenSpec 校验，确认 `confirm-completed-worktree-cleanup-status` 的规格和任务状态有效。
- [x] 5.2 手动检查已完成视图关键流程：目录删除、分支删除、已合并、未合并和历史缺失分支信息。
- [x] 5.3 执行 `wails build -tags webkit2_41`，确认生成可执行文件。
