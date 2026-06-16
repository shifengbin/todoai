## Context

TODO 工作区目前把 TODO 分为 `未执行`、`执行中`、`已完成`。活动 TODO 可以通过详情弹窗编辑标题、描述、优先级和关联工程；已完成 TODO 只在 `已完成` 列表中展示完成时间和项目快照，不能打开详情，也不能从列表中删除。

后端完成 TODO 时会把关联工程保存到 `Todo.ProjectSnapshots`，随后移除 `TodoProjects` 运行时关联并清理终端上下文。因此完成态详情必须读取快照数据，而不是重新使用活动 TODO 的工程关联数据。

## Goals / Non-Goals

**Goals:**

- 允许用户在 `已完成` 视图中打开 completed TODO 的详情。
- 复用现有 TODO 详情弹窗布局，但在 completed TODO 下切换为只读模式。
- 在只读详情中显示标题、描述、优先级和完成时项目快照。
- 允许用户从 `已完成` 视图单个删除 completed TODO。
- 允许用户在 `已完成` 视图多选 completed TODO 并批量删除。

**Non-Goals:**

- 不允许编辑 completed TODO 的标题、描述、优先级或项目。
- 不支持 `未执行` 或 `执行中` TODO 的批量删除。
- 不恢复 completed TODO 的终端、shell 进程或活动项目关联。
- 不改变 completed TODO 的排序规则。

## Decisions

### 复用详情弹窗并增加只读模式

完成态详情复用现有 TODO 详情弹窗，避免新建第二套详情 UI。弹窗根据 TODO 状态决定模式：`not-started` 和 `in-progress` 保持编辑模式，`completed` 使用只读模式。

只读模式隐藏保存按钮，并禁用或替换可编辑控件。标题、描述和优先级以不可编辑方式展示；项目区域展示 `ProjectSnapshots` 中的项目名称和路径。由于 completed TODO 不再保留 `TodoProjects`，项目快照只作为历史上下文展示，不提供新增、移除或选择动作。

### 删除 API 允许 completed TODO

现有删除流程只允许 open TODO。该变更将单个删除扩展到 completed TODO，使 completed TODO 被删除后从可见 TODO 列表移除。删除 completed TODO 不需要关闭终端，但调用应用层删除时仍可保留终端清理调用的幂等行为。

### 批量删除仅面向已完成视图

批量删除入口只在 `已完成` 视图中显示，并且只允许选择 completed TODO。后端应提供批量删除能力或等价的批量操作封装，并校验所有目标都是 completed TODO，避免前端绕过限制批量删除 open TODO。

已完成视图的批量删除不复用 `未执行` 和 `执行中` 的批量展开/收起控件。完成态批量删除属于历史清理动作，应有独立的选择状态和确认流程。

## Risks / Trade-offs

- [Risk] 复用详情弹窗可能让 completed TODO 误进入保存路径。→ 只读模式隐藏保存按钮，并在保存逻辑中拒绝 completed TODO 的编辑提交。
- [Risk] 项目快照和活动工程关联数据结构不同，直接复用工程编辑列表会误导用户。→ completed 模式单独渲染项目快照，只显示名称和路径。
- [Risk] 批量删除如果只在前端限制，API 可能被绕过。→ 后端批量删除校验目标状态，仅允许 completed TODO 的批量删除。
- [Risk] 删除完成记录后无法从应用内恢复。→ 删除前使用确认流程，并在批量删除确认中显示删除数量。

## Migration Plan

无需数据迁移。现有 completed TODO 已包含标题、描述、优先级和项目快照；缺少项目快照的旧数据按空快照展示。

回滚时移除完成态详情入口和批量删除入口，并恢复后端删除状态校验即可。已被删除的 completed TODO 不做自动恢复。

## Open Questions

无。
