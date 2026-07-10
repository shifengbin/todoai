## 1. 测试刻画

- [x] 1.1 添加后端测试：执行中 TODO 通过 `AddProjectSelectionsToTodo` 添加项目后应准备 Git worktree 并生成包含项目信息的 `README.md`
- [x] 1.2 更新后端测试：无关联项目且无初始化文件的 TODO 进入执行中后不创建任务目录、不生成 `README.md`
- [x] 1.3 更新后端测试：无关联项目但有初始化文件的 TODO 进入执行中后写入初始化文件但不生成 `README.md`

## 2. 后端实现

- [x] 2.1 调整 `AddProjectSelectionsToTodo`，在目标 TODO 为 `in-progress` 时调用任务工作区准备逻辑并返回重新加载后的状态
- [x] 2.2 调整任务工作区准备逻辑，无关联项目且无初始化文件快照时直接跳过目录创建和 `WorkspaceDirName` 持久化
- [x] 2.3 调整 README 写入逻辑，未关联任何 TODO project 时不生成或重写 `README.md`
- [x] 2.4 确认添加第一个项目到已执行中 TODO 时会创建任务目录、准备 worktree、写入初始化文件并生成 README

## 3. 自动化验证

- [x] 3.1 运行 Go 测试：`go test ./...`
- [x] 3.2 运行客户端自动化测试：`cd frontend && npm test -- --run`
- [x] 3.3 执行自动 review，检查代码质量、规格一致性和回归风险，并处理发现的问题

## 4. 打包

- [x] 4.1 运行 Wails 打包：`wails build -tags webkit2_41`
