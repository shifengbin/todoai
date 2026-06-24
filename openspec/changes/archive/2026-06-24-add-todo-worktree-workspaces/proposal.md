## Why

当前任务虽然可以关联多个项目并隔离运行时终端，但同一项目在多个任务中仍共享同一个工作目录和 Git 工作树，容易产生改动冲突。为任务创建独立目录，并将每个任务项目放入独立 Git worktree，可以让不同任务的文件改动、分支和终端上下文真正隔离。

## What Changes

- 为每个任务创建独立任务工作区目录，目录位于当前 workspace 下的任务工作区根目录中。
- 任务工作区目录名在首次创建时由任务标题和描述的 MD5 生成，任务后续编辑不重命名目录。
- 创建任务工作区后自动生成并维护 `README.md`，记录任务标题、任务描述、每个项目的 base 分支和当前 worktree 分支。
- 在 TODO 关联项目的选择流程中，为每个项目提供分支下拉/输入控件；未选择时默认使用主分支。
- 为任务中的每个关联 Git 项目在任务目录下创建 Git worktree。
- 支持基于已有分支创建任务隔离 worktree 分支；也支持输入不存在的分支名，并从主分支创建该分支作为 worktree 分支。
- 在 TODO 行菜单中新增“打开任务文件夹”，在 TODO 项目行菜单中新增“打开项目文件夹”。
- 支持任务级终端，工作目录为任务工作区目录；项目级终端工作目录改为对应任务项目 worktree 目录。
- 任务完成后保留任务目录和 worktree，由用户手动清理。

## Capabilities

### New Capabilities

- `todo-worktree-workspaces`: 任务工作区目录、Git worktree 创建、README 维护、文件夹打开和手动清理策略。

### Modified Capabilities

- `todo-workspace`: TODO 关联项目流程需要记录项目分支选择，并在 TODO/项目菜单暴露打开文件夹入口。
- `embedded-shell-sessions`: 终端启动目录需要支持任务级终端和任务项目 worktree 目录。
- `workspace-lifecycle`: workspace 数据布局需要允许任务工作区目录保存在 workspace 根目录下，而非 `.data` 内部。

## Impact

- 后端 Go：TODO/项目数据模型、workspace 路径管理、Git worktree 操作、README 生成、文件夹打开 API、终端创建和恢复逻辑。
- 前端 Vue：TODO 创建/编辑/添加项目表单的分支选择控件，TODO 和 TODO 项目行菜单，任务级终端 UI。
- Wails 绑定：新增或调整打开文件夹、创建任务终端、准备任务工作区/worktree 的应用方法。
- 持久化：TODO 项目副本需要保存 base 分支、worktree 分支、worktree 路径和准备状态；任务需要保存任务工作区目录名或路径。
- Git 依赖：需要检测 Git 可用性、仓库有效性、分支存在性、worktree 创建失败和分支 checkout 冲突。
