## Context

当前 TODO 数据模型只有 `active` 和 `archived` 两类状态，完成和删除都通过 `archivedReason` 区分。前端 TODO tab 也只提供 `Active` 和 `Archived` 两个视图，并将所有非 `active` TODO 都放入归档列表。

新需求要求用户手动管理 TODO 的执行状态：新建后默认未执行，用户可以在未执行和执行中之间切换，完成后进入已完成；删除不再进入归档或已完成视图。这会影响 Go 持久化模型、Wails 绑定、前端列表过滤、排序、折叠和测试。

## Goals / Non-Goals

**Goals:**

- 将 TODO 持久状态扩展为 `not-started`、`in-progress`、`completed`。
- 将 TODO tab 的 `Active` / `Archived` 切换改为 `未执行`、`执行中`、`已完成` 三个视图。
- 支持用户手动在 `未执行` 和 `执行中` 之间切换 TODO 状态。
- 新建 TODO 默认进入 `未执行`，并在侧边栏中保持收起。
- 已完成视图只展示完成的 TODO；删除 TODO 不在任何 TODO 列表中展示。
- 兼容旧持久化数据中的 `active`、`archived/completed` 和 `archived/deleted`。

**Non-Goals:**

- 不根据终端是否存在或是否 busy 自动切换 TODO 状态。
- 不提供从 `已完成` 恢复到未执行或执行中的能力。
- 不改变终端进程生命周期规则：完成或删除 TODO 仍会关闭该 TODO 下的终端。
- 不改变 TODO 优先级模型、项目关联模型或终端启动配置模型。

## Decisions

1. **状态使用持久字段，而不是从终端推导。**

   选择：后端保存 `not-started`、`in-progress`、`completed`。前端只根据 `todo.status` 分栏。

   原因：用户明确要求手动切换执行状态。如果用终端关系推导，创建终端会隐式改变 TODO 所属栏，且无法表达“我已经开始但暂时没有终端”的状态。

   备选：继续保存 `active/archived`，前端按终端关系拆分。该方案改动小，但不满足手动状态语义。

2. **删除 TODO 从可见 TODO 集合中移除。**

   选择：新的删除行为移除 TODO 及其关联项目，并关闭对应终端；删除后的 TODO 不进入 `completed`，也不在已完成视图中出现。

   原因：用户要求“删除不要显示在归档中”。删除和完成代表不同意图，完成历史不应混入删除记录。

   备选：保留隐藏的 `deleted` 状态。该方案便于审计，但会增加过滤和迁移复杂度，目前没有恢复或审计需求。

3. **旧数据在加载或规范化时兼容。**

   选择：读取旧数据时将 `active` 视为 `not-started`，将 `archived` 且 `archivedReason=completed` 视为 `completed`，将 `archived` 且 `archivedReason=deleted` 从 TODO 列表中过滤或在下次保存时清理。

   原因：避免已有用户数据在升级后丢失已完成历史，同时让旧删除记录不再污染已完成视图。

4. **新建 TODO 不自动切换当前工作上下文。**

   选择：创建 TODO 后保留当前 TODO/项目/终端上下文，新 TODO 在未执行视图中默认收起。

   原因：现有侧边栏 watcher 会自动展开当前 active TODO；如果后端继续把新建 TODO 设为当前 TODO，就会抵消“添加后默认收起”的要求。保留当前上下文也避免创建任务时打断正在工作的终端。

   备选：创建后仍选中新 TODO，并在前端为“刚创建 TODO”做特殊折叠优先级。该方案需要额外临时状态，规则更脆弱。

5. **排序控件复用当前 active TODO 排序规则。**

   选择：`未执行` 和 `执行中` 两个视图都使用同一套优先级/创建时间排序控件；`已完成` 不受该排序影响。

   原因：需求明确要求未执行和执行中按现在 active 顺序处理，复用现有排序避免引入第二套规则。

## Risks / Trade-offs

- [Risk] 修改状态常量会影响所有以 `TodoStatusActive` 判断可操作 TODO 的路径。→ Mitigation: 引入 `isOpenTodoStatus` / `openTodoByID` 一类辅助函数，统一定义未执行和执行中都是可操作 TODO。
- [Risk] 旧 `archived/deleted` 数据如果只在前端隐藏，后端测试仍可能看到历史记录。→ Mitigation: 在持久化加载规范化层明确处理旧删除记录，并用后端测试覆盖。
- [Risk] 新建后不自动选中 TODO 会改变现有“创建带项目 TODO 后当前上下文切到第一个项目”的行为。→ Mitigation: 在 spec 中显式声明创建不会改变当前工作上下文，以满足默认收起和不中断工作流。
- [Risk] Wails 生成绑定文件需要随 Go API 变更更新。→ Mitigation: 实现后运行 Wails 绑定生成或项目既有生成命令，并提交生成文件。
