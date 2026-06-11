## Why

当前 TODO 工作区只区分活动和归档，用户无法显式表达任务尚未执行或正在执行。归档视图还会混入已删除 TODO，导致“已完成”语义不清晰。

该变更将 TODO 工作流调整为可手动管理的三栏状态，让用户按未执行、执行中、已完成组织任务，并让删除行为从完成历史中消失。

## What Changes

- 将 TODO tab 顶部的 `Active` / `Archived` 视图改为三栏：`未执行`、`执行中`、`已完成`。
- 新建 TODO 默认进入 `未执行`，并在创建后默认保持收起状态。
- 增加手动状态切换能力，允许用户在 `未执行` 和 `执行中` 之间切换。
- 完成 TODO 后进入 `已完成`，并继续关闭和移除该 TODO 下的运行时终端。
- 删除 TODO 后不再显示在已完成或归档视图中，并继续关闭和移除该 TODO 下的运行时终端。
- `未执行` 和 `执行中` 列表复用当前活动 TODO 的排序规则和排序控件。
- 旧数据迁移语义：现有 `active` TODO 视为 `未执行`，现有 `archived` 且原因为 `completed` 的 TODO 视为 `已完成`，现有 `archived` 且原因为 `deleted` 的 TODO 不在任何 TODO 列表中展示。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `todo-workspace`: 修改 TODO 工作区的状态模型、列表分栏、完成展示、删除展示和创建后的默认折叠行为。

## Impact

- 后端 TODO 状态常量、创建、更新、完成、删除、持久化兼容和活动 TODO 校验逻辑。
- Wails 绑定模型和前端状态消费逻辑。
- `ProjectSidebar` 的 TODO 视图切换、列表过滤、排序、批量展开/收起和操作按钮。
- 前端与后端测试中对 active/archived、完成、删除、排序和折叠的断言。
