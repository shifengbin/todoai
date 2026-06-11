## 1. 后端后台命令窗口控制

- [x] 1.1 为 Git 后台命令创建平台封装，在 Windows 分支设置隐藏控制台窗口，非 Windows 分支保持现有行为。
- [x] 1.2 将 `git status` 和 `git init` 命令创建路径统一接入该封装，避免后台 Git 命令在 Windows GUI 应用中弹出控制台。
- [x] 1.3 增加 Go 单元测试覆盖 Git 命令配置边界，并通过 Windows 交叉编译验证平台分支可编译。

## 2. 前端 Git 刷新去重

- [x] 2.1 在 `App.vue` 中区分导入项目、focus、显式项目选择、TODO 项目选择、终端命令结束等 Git 刷新来源。
- [x] 2.2 让导入项目应用返回状态时跳过 Git 状态刷新，保持导入摘要和项目列表正常更新。
- [x] 2.3 为同一项目的 focus Git 刷新增加短间隔去重和请求进行中保护，防止焦点抖动反复启动后台命令。
- [x] 2.4 增加前端自动化测试，覆盖导入项目不调用 `GetProjectGitStatus`、短时间重复 focus 事件只触发一次 Git 状态刷新，以及项目切换仍能刷新新项目。

## 3. TODO 展开触发 Git 懒加载

- [x] 3.1 在 `ProjectSidebar.vue` 中为 TODO 从收起变为展开的操作发出事件，同时避免重复展开时重复发事件。
- [x] 3.2 在 `App.vue` 中处理 TODO 展开事件，仅当展开 TODO 包含当前激活 TODO project 且项目路径可用时刷新当前项目 Git 状态。
- [x] 3.3 确保选择 TODO 项目仍会刷新该项目 Git 状态，并与 TODO 展开事件共享同项目请求去重逻辑。
- [x] 3.4 增加前端自动化测试，覆盖展开当前 TODO 后触发 Git 查询、展开无关 TODO 不触发查询、选择 TODO 项目触发查询。

## 4. Windows 嵌入式终端降级

- [x] 4.1 在 Windows PTY 不支持时返回稳定、可识别的终端不可用错误，不启动额外 shell 进程。
- [x] 4.2 在前端捕获终端不可用错误，展示明确的 Windows 嵌入式终端暂不支持提示，并保持终端区域和应用布局稳定。
- [x] 4.3 确保终端启动失败后不会因应用 focus、终端容器重新挂载或状态刷新自动重试。
- [x] 4.4 增加 Go 和前端自动化测试，覆盖 Windows 后端不可用错误映射、界面提示和不自动重试行为。

## 5. 验证与自动 Review

- [x] 5.1 运行 `go test ./...`，确认后端测试通过。
- [x] 5.2 运行 `GOOS=windows GOARCH=amd64 go test -c -o /tmp/tui-helper-windows.test.exe .`，确认 Windows 目标可编译。
- [x] 5.3 运行前端测试，确认新增和既有 Vue 行为测试通过。
- [x] 5.4 运行 `openspec validate fix-windows-console-flashing --strict`。
- [x] 5.5 执行自动 review，检查实现是否符合 proposal、design、specs 和现有代码风格。
