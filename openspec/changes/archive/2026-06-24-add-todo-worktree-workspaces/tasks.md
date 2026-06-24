## 1. 数据模型与持久化

- [x] 1.1 扩展 TODO 数据模型，保存任务工作区目录名或路径，并兼容旧数据缺失字段。
- [x] 1.2 扩展 TODO 项目副本数据模型，保存 base 分支、worktree 分支、worktree 路径、worktree 状态和错误信息。
- [x] 1.3 更新项目状态加载、保存、规范化和迁移逻辑，确保旧 workspace 打开时不自动创建目录或 worktree。
- [x] 1.4 更新 Wails 生成模型所需的 Go 类型和前端 TypeScript 模型。

## 2. 任务工作区与 README

- [x] 2.1 实现任务工作区路径解析，使用 workspace 根目录下的任务工作区根目录和 `md5(title+description)` 目录名。
- [x] 2.2 实现任务工作区创建逻辑，并在 TODO 首次进入 `in-progress` 时持久化目录引用。
- [x] 2.3 实现 `README.md` 完整重写逻辑，输出任务名称、任务详情和每个项目的 base/worktree 分支。
- [x] 2.4 在 TODO 标题、描述、项目关联和项目分支信息变化后触发 README 更新。
- [x] 2.5 确保任务完成、删除、关闭 workspace 和切换 workspace 时不删除任务工作区目录。

## 3. Git Worktree 能力

- [x] 3.1 添加 Git 仓库检测、默认主分支识别、分支存在性检测和 worktree checkout 冲突检测。
- [x] 3.2 实现基于已存在分支创建任务隔离 worktree 分支的逻辑。
- [x] 3.3 实现用户输入不存在分支时从主分支创建该分支并创建 worktree 的逻辑。
- [x] 3.4 在 TODO 进入 `in-progress` 时为已关联 Git 项目批量准备 worktree。
- [x] 3.5 在执行中 TODO 新增项目时立即准备该项目 worktree。
- [x] 3.6 持久化每个 TODO 项目的 worktree 成功或失败状态，并在失败时保留可展示错误。

## 4. 终端与文件夹 API

- [x] 4.1 扩展终端模型和终端历史，区分 workspace 全局终端、任务级终端和 TODO 项目终端。
- [x] 4.2 添加创建任务级终端 API，限制为 `in-progress` TODO 且任务工作区目录存在。
- [x] 4.3 修改 TODO 项目终端启动目录，使其使用 TODO 项目 worktree 路径而不是原项目路径。
- [x] 4.4 在 TODO 项目 worktree 未准备或准备失败时拒绝创建项目终端。
- [x] 4.5 添加打开任务文件夹和打开项目 worktree 文件夹的后端 API，并处理目录缺失错误。
- [x] 4.6 确保关闭或切换 workspace 时停止任务级终端和项目终端，但保留目录和 worktree。

## 5. 前端交互

- [x] 5.1 在创建 TODO 表单的项目选择区域为每个选中项目添加分支下拉/输入控件。
- [x] 5.2 在为 TODO 添加项目弹窗中为每个选中项目添加分支下拉/输入控件。
- [x] 5.3 在 TODO 树中显示任务级终端入口和任务级终端列表，并让任务终端活动参与收起 TODO 聚合状态。
- [x] 5.4 在 TODO 行下拉菜单中添加 `打开任务文件夹` 操作。
- [x] 5.5 在 TODO 项目行下拉菜单中添加 `打开项目文件夹` 操作。
- [x] 5.6 在 TODO 项目行展示 worktree 准备失败状态，并阻止不可用项目终端入口。
- [x] 5.7 更新前端 Wails 调用、状态应用逻辑和终端激活逻辑，支持任务级终端。

## 6. 自动化测试与验证

- [x] 6.1 添加 Go 单元测试覆盖任务目录命名、README 生成、编辑后不改名和空描述 README。
- [x] 6.2 添加 Go 单元测试覆盖 Git 分支选择、输入新分支、worktree 创建失败和冲突处理。
- [x] 6.3 添加 Go 单元测试覆盖任务级终端创建、项目终端 worktree cwd 和关闭 workspace 保留目录。
- [x] 6.4 添加前端自动化测试覆盖分支选择控件、任务终端展示、打开文件夹菜单和 worktree 失败状态。
- [x] 6.5 运行后端测试 `go test ./...`。
- [x] 6.6 运行客户端自动化测试 `npm test -- --run`。
- [x] 6.7 执行自动 review，检查实现是否满足 proposal、design、specs 和项目代码规范。
- [x] 6.8 运行 `wails build -tags webkit2_41` 生成可执行文件。
