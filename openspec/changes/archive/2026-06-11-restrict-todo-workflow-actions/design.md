## Context

TODO 工作区已经引入 `not-started`、`in-progress`、`completed` 三个持久状态，但当前交互仍把 `not-started` 和 `in-progress` 当作同一类“开放 TODO”：用户可以把执行中的 TODO 退回未执行，也可以在未执行 TODO 下添加终端。

新的产品语义要求工作流是单向的：未执行任务先开始，执行中的任务才承载终端和完成动作。查看/编辑、添加项目和删除仍是 TODO 管理动作，不受这次状态收紧影响。

## Goals / Non-Goals

**Goals:**

- 将状态流转收紧为 `not-started -> in-progress -> completed`。
- 让未执行 TODO 只暴露开始和删除等允许动作，不显示完成和添加终端入口。
- 让执行中 TODO 暴露完成、删除和添加终端入口，不再支持退回未执行。
- 在 Go 后端对状态切换、完成和创建 TODO 终端做同样约束。
- 保持现有 API 方法签名和持久化字段不变。

**Non-Goals:**

- 不新增 TODO 状态。
- 不根据终端活动自动改变 TODO 状态。
- 不改变删除 TODO、编辑 TODO、关联项目和终端生命周期清理的既有语义。
- 不提供从 `completed` 恢复到开放状态的能力。

## Decisions

1. **把工作流定义成单向状态机。**

   选择：只允许 `ChangeTodoStatus(todoID, "in-progress")` 将 `not-started` TODO 标记为执行中；不再允许 `in-progress -> not-started`。

   原因：`in-progress` 表示用户已经开始执行，退回未执行会让状态历史和终端上下文含义变弱。删除仍可处理放弃或误建的任务。

   备选：保留后端双向切换，只隐藏前端按钮。该方案兼容性高，但 API 仍能绕过 UI 产生被产品规则禁止的状态。

2. **完成只属于执行中 TODO。**

   选择：`CompleteTodo` 只接受 `in-progress` TODO；前端只在执行中视图显示完成按钮。

   原因：未执行任务还没有开始，不应直接完成。这样用户路径更清晰：开始后才能完成。

   备选：后端仍允许未执行完成，前端隐藏按钮。该方案减少测试调整，但状态模型不一致。

3. **终端只能挂在执行中 TODO 上。**

   选择：TODO 项目行的添加终端按钮只在 TODO 为 `in-progress` 且项目可用时显示；`App.CreateTodoTerminal` 在创建 shell session 前校验所属 TODO 状态。

   原因：终端代表实际执行上下文，允许未执行 TODO 创建终端会使“未执行”失去含义。

   备选：创建终端时自动把 TODO 改为执行中。该方案减少点击，但违背“状态由用户操作决定”的既有设计原则，也会引入隐式状态变化。

4. **保留非执行动作。**

   选择：未执行和执行中的 TODO 都继续允许查看/编辑、添加项目和删除。

   原因：这些是任务管理动作，不代表执行状态推进。用户需要在开始前补充项目和描述。

## Risks / Trade-offs

- [Risk] 现有测试依赖 `in-progress -> not-started` 或未执行可建终端。→ Mitigation: 更新后端和前端测试，让测试断言新的拒绝行为和按钮可见性。
- [Risk] 只改前端会留下 API 绕过路径。→ Mitigation: 后端同步校验状态，并用 App/ProjectManager 测试覆盖。
- [Risk] 当前 base specs 尚未归档三状态变更，delta 可能需要同时覆盖旧文案和已完成变更中的行为。→ Mitigation: 本变更规格以最终三状态行为为准，并在实现时依赖现有三状态代码路径。

## Migration Plan

无需数据迁移。持久化状态值保持 `not-started`、`in-progress`、`completed` 不变，仅收紧允许的操作。

## Open Questions

None.
