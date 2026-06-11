## Why

当前工作区以项目为主线组织终端，用户在多个项目和多个终端之间切换时，很难判断某个终端属于哪项任务、为什么启动。需要把任务提升为主要工作上下文，让终端的用途跟随 TODO 记忆，并在任务完成或删除后清理相关终端。

## What Changes

- 新增 TODO 视图作为主工作流，按 `TODO -> 项目 -> 终端` 的层级展示和切换工作上下文。
- 在现有项目区域增加 `TODO` 和 `项目` 两个 tab：TODO tab 负责任务工作和终端入口，项目 tab 负责项目库管理。
- 支持创建 TODO，并为 TODO 选择一个或多个已导入项目；同一个项目可关联到多个 TODO。
- 将终端运行态隔离到 TODO 项目上下文中；同一项目出现在不同 TODO 下时，各自拥有独立终端集合。
- 支持在项目 tab 中选择父目录批量导入第一层子目录为项目，并继续支持删除项目。
- 完成或删除 TODO 时关闭并销毁该 TODO 下的所有终端，然后归档 TODO。
- 在 TODO tab 中增加归档查看入口，展示已完成或已删除归档的 TODO 及其关联项目快照。
- 项目 tab 不直接创建、选择或展示可操作终端，避免绕过 TODO 上下文。
- **BREAKING**: 项目选择不再作为终端工作区的唯一入口；终端操作必须发生在 TODO 上下文中。

## Capabilities

### New Capabilities

- `todo-workspace`: 管理 TODO、TODO 与项目的关联、TODO 主视图、完成/删除归档、归档查看，以及 TODO 级选中状态。

### Modified Capabilities

- `project-workspace`: 增加项目库 tab 和父目录批量导入；调整项目视图为管理用途，不再直接作为终端入口。
- `embedded-shell-sessions`: 将终端归属从项目级调整为 TODO 项目上下文级，并在 TODO 完成/删除时批量关闭对应终端。

## Impact

- Go 后端：项目持久化状态需要扩展为包含 TODO、TODO 项目关联、归档快照和新的活动上下文；Wails API 需要增加 TODO 管理、TODO 项目关联、父目录批量导入和 TODO 级终端操作。
- Shell 运行态：`ProjectTerminal` 和 `ShellSessionManager` 需要支持 `todoId`/`todoProjectId` 维度的终端隔离、选中和批量清理。
- Vue 前端：侧边栏需要增加 TODO/项目 tab、TODO 树、归档视图、TODO 创建和项目选择流程；终端面板和状态栏需要基于 TODO 上下文显示。
- Wails 生成绑定：新增或调整后端方法后需要重新生成前端 bindings。
- 测试：需要覆盖 TODO 持久化、项目批量导入、TODO 项目关联、终端隔离、归档清理、项目删除跨 TODO 清理，以及前端树交互。
