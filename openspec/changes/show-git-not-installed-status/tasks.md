## 1. 后端 Git 可用性状态

- [x] 1.1 增加可测试的 Git 命令可用性检查封装，默认使用应用进程环境中的 `git` 查找结果
- [x] 1.2 扩展 `GitStatus`，增加用于表达 Git 未安装或不可执行的结构化字段
- [x] 1.3 修改 Git 状态查询路径，在执行 `git status` 前先校验 Git 命令，缺失时返回结构化状态且不调用命令 runner
- [x] 1.4 修改 Git 初始化路径，在执行 `git init` 前先校验 Git 命令，缺失时返回明确的未安装 Git 错误且不调用命令 runner
- [x] 1.5 保持项目路径不可用、非 Git 仓库、Git 查询超时和普通命令失败的现有行为不变

## 2. Wails 绑定与前端状态栏

- [x] 2.1 更新 Wails 前端绑定模型，使 `GitStatus` 暴露 Git 未安装状态字段
- [x] 2.2 调整状态栏 Git 状态优先级：无项目、路径不可用、Git 未安装、加载中、普通错误、非 Git 仓库、正常仓库状态
- [x] 2.3 在 Git 未安装时显示 `未安装 Git` 状态芯片
- [x] 2.4 在 Git 未安装时隐藏 `Initialize Git Repository` 操作
- [x] 2.5 确认 Git 未安装状态不会阻断项目选择、TODO 项目选择或终端区域操作

## 3. 自动化测试

- [x] 3.1 增加后端测试，覆盖 Git 缺失时 Git 状态查询不执行 runner 并返回结构化状态
- [x] 3.2 增加后端测试，覆盖 Git 缺失时 Git 初始化不执行 runner 并返回未安装 Git 错误
- [x] 3.3 增加 App 层测试，覆盖可用项目在 Git 缺失时返回状态栏可消费的 Git 状态
- [x] 3.4 更新客户端组件测试，覆盖状态栏显示 `未安装 Git`
- [x] 3.5 更新客户端组件测试，覆盖 Git 缺失时不显示初始化 Git 仓库按钮
- [x] 3.6 回归客户端测试，确认非 Git 仓库、路径不可用和普通 Git 查询失败仍显示原有状态

## 4. 验证与 Review

- [x] 4.1 运行 Go 测试：`go test ./...`
- [x] 4.2 运行前端自动化测试：`cd frontend && npm run test`
- [x] 4.3 运行前端构建：`cd frontend && npm run build`
- [x] 4.4 运行 OpenSpec 校验，确认 proposal、design、specs、tasks 可被解析
- [x] 4.5 执行自动 review，检查 Git 缺失状态契约、状态栏优先级、错误处理和测试覆盖
- [x] 4.6 运行 Wails 打包生成可执行文件：`wails build -tags webkit2_41`
