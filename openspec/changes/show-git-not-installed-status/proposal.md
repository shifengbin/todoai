## Why

当前 Git 状态栏在系统未安装 `git` 时只显示通用的 Git 状态不可用，用户无法区分是 Git 未安装、项目不是仓库，还是一次普通查询失败。先校验 Git 命令并给出明确状态，可以让用户快速理解环境问题，同时避免继续执行必然失败的 Git 命令。

## What Changes

- Git 状态查询在执行 `git status` 前先校验 `git` 命令是否存在。
- 当 `git` 未安装或不可执行时，状态栏显示明确的 `未安装 Git` 状态。
- `git` 未安装时，状态栏不显示初始化 Git 仓库操作，因为 `git init` 无法成功执行。
- Git 初始化操作在执行 `git init` 前同样校验 `git` 命令是否存在，并返回明确的未安装错误。
- 项目路径不可用、非 Git 仓库、普通 Git 查询失败等现有状态继续保持原有语义。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `project-workspace`: 修改项目 Git 状态栏在 Git 命令缺失时的状态展示、查询前校验和初始化入口可用性要求。

## Impact

- 后端 Git 状态查询和 Git 初始化命令执行路径。
- `GitStatus` Wails 绑定模型和前端状态消费逻辑。
- 主工作区底部状态栏 Git 状态芯片和初始化仓库按钮显示条件。
- 后端 Git 状态测试、App API 测试和前端组件测试。
