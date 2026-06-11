## Why

当 TODO 列表中包含多个任务、项目和终端时，用户需要逐个收起或展开 TODO 分支才能整理视图，操作成本较高。为活动 TODO 列表增加一键收起和一键展开，可以让用户快速切换“总览任务”和“查看详情”的工作状态。

## What Changes

- 在 TODO tab 的活动 TODO 列表中增加批量收起和批量展开入口。
- 批量收起时隐藏所有活动 TODO 下的项目和终端子项，但保留 TODO 行可见。
- 批量展开时恢复显示所有活动 TODO 下的项目子项。
- 保持现有单个 TODO 分支展开/收起行为不变。
- 保持当前 TODO、TODO 项目或终端被选中时自动展开对应 TODO 的行为不变。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: 为活动 TODO 树增加一键收起和一键展开所有 TODO 分支的行为要求。

## Impact

- 影响前端 TODO 侧边栏组件的展示状态和交互控件。
- 影响 TODO 侧边栏相关单元测试。
- 不改变 Go 数据模型、Wails API、持久化格式或外部依赖。
