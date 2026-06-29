## Why

用户需要在项目仓库或其子目录中快速查看当前工程已完成的 TodoAI 任务，不必打开桌面界面再切换到已完成视图。现有已完成任务已经保存了完成时项目快照和分支信息，命令行入口可以复用这些数据提供轻量查询能力。

## What Changes

- 新增 `todoai list --done` 命令行能力。
- 该命令可在已登记项目根目录或其子目录执行，并解析当前目录所属的 TodoAI 项目。
- 命令仅返回当前项目相关、状态为 `completed` 的 TODO。
- 输出 JSON 数组，元素包含任务名称、worktree 分支、base 分支。
- 历史快照缺少分支信息时，命令保留该任务并以 `-` 显示缺失分支。
- 未找到当前目录对应项目时，命令返回错误并说明无法定位 TodoAI 项目。

## Capabilities

### New Capabilities

- `todo-cli`: TodoAI 命令行查询能力，包括从项目目录列出当前项目已完成 TODO。

### Modified Capabilities

None.

## Impact

- CLI entrypoint in `main.go`; `todoai list --done` must run without starting the Wails GUI.
- Backend Go helpers for locating the current project from the working directory and reading workspace project state.
- Existing TodoAI workspace persistence under `.data/projects.json`; no new persistent storage format is expected.
- Automated Go tests for CLI routing, project-directory matching, completed TODO filtering, JSON output, branch output, and error/empty states.
