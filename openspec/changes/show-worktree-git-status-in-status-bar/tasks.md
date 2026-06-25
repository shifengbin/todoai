## 1. 后端 Git 状态上下文

- [x] 1.1 增加后端测试，覆盖 ready TODO project 使用 `worktreePath` 查询 Git 状态，而不是来源项目 `path`。
- [x] 1.2 增加后端测试，覆盖 TODO project worktree 未 ready、`worktreePath` 为空或目录不可用时返回稳定不可用状态且不执行 Git 查询。
- [x] 1.3 实现 TODO project Git 状态查询入口，按 `todoProjectId` 查找 TODO project 并对 ready worktree 目录复用现有 Git 状态解析。
- [x] 1.4 保留普通 `GetProjectGitStatus(projectID)` 行为，确保非 TODO project 上下文仍按项目路径查询。

## 2. Wails 绑定与前端上下文建模

- [x] 2.1 更新 Wails 前端绑定，使新增 TODO project Git 状态查询入口可从 Vue 调用。
- [x] 2.2 在 `App.vue` 中集中构造 Git status context key，区分 `project:<projectId>` 与 `todo-project:<todoProjectId>`。
- [x] 2.3 调整状态栏 active project 展示，在 TODO project 上下文中优先显示 ready `worktreePath`，并保留来源项目路径作为未准备时的回退展示。
- [x] 2.4 调整 Git 状态刷新、in-flight 去重、focus 去重和过期响应判断，全部使用 context key 而不是只使用 `projectId`。
- [x] 2.5 调整初始化 Git 仓库按钮逻辑，确保普通项目路径行为不变，TODO project worktree 不可用时不误用来源项目路径初始化或展示来源项目 Git 状态。

## 3. 前端自动化测试

- [x] 3.1 增加前端测试，覆盖选中 ready TODO project 时调用 TODO project Git 状态查询入口并显示 worktree 分支和改动数量。
- [x] 3.2 增加前端测试，覆盖同一来源项目的两个 TODO project worktree 独立刷新，旧请求响应不会覆盖当前状态栏。
- [x] 3.3 增加前端测试，覆盖 TODO project worktree 未 ready 或路径不可用时状态栏显示稳定不可用状态，且不调用来源项目 Git 状态查询。
- [x] 3.4 更新既有项目 Git 状态栏测试，确认普通项目上下文仍调用 `GetProjectGitStatus(projectID)` 并保留初始化 Git 仓库行为。

## 4. 验证与交付

- [x] 4.1 运行 Go 单元测试，确认后端 Git 状态查询和既有 worktree/终端行为未回归。
- [x] 4.2 运行前端自动化测试，确认状态栏、刷新去重和初始化按钮行为符合规格。
- [x] 4.3 运行自动 review 检查，例如 `git diff --check`，并修复发现的格式或空白问题。
- [x] 4.4 运行 `wails build -tags webkit2_41` 生成可执行文件。
