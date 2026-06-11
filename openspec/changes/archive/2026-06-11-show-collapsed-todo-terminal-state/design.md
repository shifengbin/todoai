## Context

当前 TODO 工作树在 `ProjectSidebar.vue` 中已经接收 `terminals` prop，并在终端行上展示前端运行时派生的 `activityState`。`activityState` 由前端根据终端标题和输出推断，取值包括 `idle`、`busy`、`needs-input`，不属于 Go 后端持久化模型。

TODO 分支收起后，项目和终端子项会被隐藏，用户只能看到父级 TODO 行。现有父级 TODO 行只展示标题、描述和项目数量，无法反映隐藏终端是否正在运行或等待输入。

## Goals / Non-Goals

**Goals:**

- 收起 TODO 分支时，在父级 TODO item 上显示隐藏子终端的最高优先级活动状态。
- 复用现有终端活动状态语义：`needs-input` 优先于 `busy`，`busy` 优先于 `idle`。
- 展开 TODO 分支时继续由子终端行展示详细状态，避免父级 TODO 行重复显示。
- 增加组件自动化测试覆盖收起态聚合显示和状态优先级。

**Non-Goals:**

- 不修改后端 `ProjectTerminal` 模型或持久化数据。
- 不新增 Wails API 或 shell session 生命周期行为。
- 不改变终端活动状态推断算法。
- 不在归档 TODO 或项目库视图中显示运行时终端活动状态。

## Decisions

### 在 `ProjectSidebar` 内部聚合状态

`ProjectSidebar` 已经持有 TODO、TODO 项目和终端列表，并且已有 `terminalsByTodoProject`、`todoProjectTerminals` 等派生数据。新增 TODO 级聚合函数可以直接扫描 `props.terminals` 中 `terminal.todoId === todo.id` 的终端，无需 App 额外传入映射。

备选方案是在 `App.vue` 预计算 `todoActivityStates` prop。该方案数据流更显式，但会把只用于 sidebar 展示的 UI 派生状态扩散到父组件，并与 sidebar 已有终端分组逻辑重复。

### 只在收起态显示 TODO 级活动提示

展开态下，用户已经能看到每个终端行的活动图标、标签和命令名称。父级 TODO 行重复显示相同状态会增加视觉噪音。收起态显示聚合提示可以补足隐藏子项后的信息缺口。

### 使用固定优先级合并多终端状态

多个终端存在不同活动状态时，TODO 行显示最高优先级状态：

1. `needs-input`
2. `busy`
3. `idle`

这样等待用户输入的终端会优先提示，其次提示仍在运行的终端。`idle` 不需要额外强调，可作为 `data-activity-state` 的默认值供测试或可访问语义使用。

## Risks / Trade-offs

- [Risk] `activityState` 是运行时前端字段，刷新后可能暂时回到 `idle`。 → 保持现有行为，不引入持久化；终端输出和标题事件会继续更新活动状态。
- [Risk] 父级 TODO 行空间有限，新增提示可能挤压标题和操作按钮。 → 复用小尺寸图标提示，避免长文案直接占据主行；详细文案放在 `aria-label` 或 `title`。
- [Risk] 状态聚合若直接遍历所有终端，理论上随终端数量线性增长。 → 当前桌面应用的终端数量较小，且聚合仅用于 sidebar 渲染；如未来数量增大，可再引入按 TODO 分组的 computed map。
