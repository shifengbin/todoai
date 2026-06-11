## Context

TODO 工作区当前在前端侧边栏中以 `TODO -> 项目 -> 终端` 树形结构展示活动任务。组件已经用本地 `collapsedTodoIds` 集合管理每个活动 TODO 的展开/收起状态，并在选中 TODO、TODO 项目或终端时自动展开对应 TODO 分支。

本次变更只需要在现有树形交互上增加批量操作入口，不需要改变 TODO 数据结构、持久化格式或 Wails API。

## Goals / Non-Goals

**Goals:**

- 在活动 TODO 列表中提供一键收起所有 TODO 分支的入口。
- 在活动 TODO 列表中提供一键展开所有 TODO 分支的入口。
- 复用现有 `collapsedTodoIds` 状态，使批量操作与单个分支折叠行为一致。
- 保持选中 TODO、TODO 项目或终端时自动展开对应分支的现有行为。

**Non-Goals:**

- 不为折叠状态增加持久化。
- 不改变归档 TODO 列表的展示方式。
- 不改变项目库 tab 或终端创建流程。
- 不引入新的前端或后端依赖。

## Decisions

- 批量操作仅作用于活动 TODO 列表。归档 TODO 当前不是可展开的树形视图，将批量操作限定在 active view 可以避免无效控件和额外状态分支。
- 批量收起通过把所有 `activeTodos` 的 id 写入 `collapsedTodoIds` 实现。这样每个 TODO 行仍可见，其项目和终端子项按现有条件隐藏。
- 批量展开通过从 `collapsedTodoIds` 中移除所有 `activeTodos` 的 id 实现。这样保留未来可能存在的其他非活动 TODO 折叠状态，并避免无条件清空所有状态带来的范围扩大。
- 批量控件放在 TODO active view 的列表工具区中，靠近 Active/Archived 视图切换后方或活动列表顶部，使用户在进入活动 TODO 树时即可发现。
- 自动展开逻辑继续优先。批量收起后，如果用户选择某个 TODO、TODO 项目或终端，现有 watcher 仍会展开对应 TODO，保证当前上下文可见。

## Risks / Trade-offs

- 批量控件增加侧边栏横向拥挤风险 -> 使用紧凑图标按钮，并通过 title/aria-label 提供语义。
- 批量收起后当前活动 TODO 可能被隐藏子项 -> 保留现有自动展开 watcher，当前上下文变化时恢复可见。
- 活动 TODO 为空时批量控件可能没有效果 -> 控件可以在无活动 TODO 时禁用或隐藏，测试覆盖空列表不应出现误触副作用。
