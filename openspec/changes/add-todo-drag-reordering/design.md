## Context

当前 `ProjectSidebar.vue` 从 `todos` 属性派生 `未执行`、`执行中` 和 `已完成` 列表。两个开放列表在组件内按优先级或创建时间排序，`已完成` 列表按完成时间排序。工作区 UI 偏好由 Go 侧 `TodoProjectUIStateStore` 保存到工作区 `DataPath` 下的 `todo-project-ui-state.json`，目前包含侧栏宽度以及各 TODO project 的视图状态，并通过临时文件和重命名进行原子写入。

手动排序同时涉及 Vue 列表交互、Wails API、工作区 UI 状态模型和向后兼容迁移。TODO 行本身还是可展开的树节点，包含菜单、按钮、项目和终端子项；因此不能把整行设为拖拽触发区域，也不能在拖动期间永久修改展开状态。

## Goals / Non-Goals

**Goals:**

- 在现有优先级和时间排序之外增加明确的手动排序模式。
- 分别维护 `未执行` 和 `执行中` 的稳定顺序，并按工作区恢复排序模式与顺序。
- 使用专用手柄和成熟拖拽库提供可靠的同列表排序、占位反馈及边缘自动滚动。
- 在拖动期间临时视觉收起 TODO 子树，并在成功、取消和失败路径中恢复原展开状态。
- 保持旧 UI 状态文件、现有自动排序和 `已完成` 排序行为兼容。
- 对显式排序保存采用乐观更新和失败回滚，避免未持久化顺序停留在界面中。

**Non-Goals:**

- 不支持跨 `未执行`、`执行中` 或 `已完成` 列表拖放来改变 TODO 状态。
- 不改变优先级、创建时间或完成时间的业务含义。
- 不把手动顺序写入 `projects.json` 或 `Todo` 业务模型。
- 不提供多选拖拽、跨工作区拖拽或已完成 TODO 的手动排序。
- 不在本次变更中重构现有 TODO 树、浮层或批量展开/收起体系。

## Decisions

### 1. 手动顺序属于工作区 UI 状态

扩展 `TodoProjectUIStateFile`，新增工作区级排序模式和两个显式顺序列表。建议使用有类型的结构而不是任意状态键 map：

```go
type TodoManualOrders struct {
    NotStarted []string `json:"notStarted,omitempty"`
    InProgress []string `json:"inProgress,omitempty"`
}

type TodoProjectUIStateFile struct {
    Version               int                           `json:"version"`
    SidebarWidth          int                           `json:"sidebarWidth,omitempty"`
    TodoSortMode          string                        `json:"todoSortMode,omitempty"`
    TodoOrdersInitialized bool                          `json:"todoOrdersInitialized,omitempty"`
    TodoOrders            TodoManualOrders              `json:"todoOrders,omitempty"`
    TodoProjects          map[string]TodoProjectUIState `json:"todoProjects"`
}
```

排序模式和顺序是展示偏好，放入已经按工作区隔离的 UI 状态文件可以避免污染 `Todo` 业务模型，也不会让项目导入、归档或 CLI 数据处理依赖展示顺序。

备选方案包括给每个 `Todo` 增加 `manualOrder`，以及直接改变 `ProjectState.Todos` 数组顺序。前者把展示偏好写入业务数据并需要批量改写 Todo；后者在两个状态列表交错时形成隐式语义，状态切换和导入更容易破坏顺序，因此不采用。

### 2. 后端统一校验和规范化 UI 状态

新增一个工作区级列表 UI 状态保存入口，例如 `SaveTodoListUIState`，一次接收排序模式和两个顺序列表，并返回规范化后的状态。排序模式只接受 `priority`、`time`、`manual`；顺序规范化按以下步骤执行：

1. 根据当前 `ProjectState.Todos` 建立 `not-started` 和 `in-progress` 的有效 ID 集合。
2. 按请求中的首次出现顺序保留属于目标状态的 ID，并丢弃重复、未知、已完成或状态不匹配的 ID。
3. 将有效但未记录的 ID 按 `createdAt` 正序和原数组索引稳定追加到目标列表末尾。
4. 将排序模式和两个规范化列表作为一次原子写入保存。

加载 UI 状态时执行相同规范化。新建 TODO、删除 TODO 或状态变化后，前端也用同一规则立即派生显示顺序；遗漏的 ID 自然追加，失效 ID 自然移除。显式切换模式或拖拽时保存完整状态，后端再次基于最新 Todo 集合校验，从而处理保存期间发生的数据变化。

`TodoOrdersInitialized` 用于区分“从未进入手动模式”和“已经保存手动顺序但当前切回自动排序”。首次切换到手动模式且该标志为 `false` 时，前端根据切换前模式计算两个开放列表的当前显示顺序，并把两组 ID 与 `manual` 模式一起提交；后端在保存手动模式后把标志设为 `true`。后续切换回手动模式复用已保存顺序，避免覆盖用户此前的拖拽结果。

该标志是权威状态，不能根据规范化后的顺序数组是否非空反推。加载未初始化工作区时，后端仍会为了确定性返回补齐后的顺序数组，但自动排序模式下必须保持 `TodoOrdersInitialized=false`；只有已持久化的显式标志或 `manual` 模式才能表示已经初始化。

`TodoProjectUIStateStore` 使用单个互斥锁把读取、字段变更和原子替换收进同一事务，列表顺序、侧栏宽度和 TODO project 视图的并发保存不会再用旧快照互相覆盖。App 同时以工作区绑定读写锁固定一次请求所使用的 project manager 与 UI state store；工作区切换必须等待已开始的 UI 状态请求结束，切换窗口中的不一致绑定直接拒绝，避免 A 工作区请求污染 B 的文件。

### 3. App 负责状态所有权，ProjectSidebar 负责交互

`App.vue` 加载并持有工作区的 `todoSortMode`、`todoOrders` 和保存中状态，将它们作为属性传给 `ProjectSidebar.vue`。侧栏负责计算列表、初始化拖拽实例并发出模式或顺序变更事件，不直接调用 Wails 持久化 API。

建议事件边界：

- `todo-sort-mode-change`：携带目标模式；首次进入手动模式时同时携带按旧模式快照的两组顺序。
- `todo-order-change`：携带当前状态、原顺序和新顺序。

`App.vue` 对这两类变更执行乐观更新并调用统一保存 API。成功后采用后端返回的规范化状态；失败时恢复事件携带的原状态并调用现有错误提示。保存期间把 `todoOrderSaving` 传回侧栏，用于禁用排序切换、拖拽、批量操作和 TODO 行操作。

加载和保存都携带前端工作区 scope epoch 与请求序号。工作区切换会重置列表状态并递增 epoch；较早工作区的保存或加载响应必须忽略。同一工作区内，保存开始与结束都会使并发 Load 失效，避免旧读取在保存成功后覆盖新模式或顺序。

### 4. 使用 SortableJS 直接管理顶层 TODO 排序

前端新增 `sortablejs` 依赖并在 `ProjectSidebar.vue` 中直接创建和销毁实例。该库原生提供专用 `handle`、`draggable`、`disabled`、拖拽生命周期、占位样式和自动滚动，能够复用现有 DOM 结构且不需要引入额外 Vue 包装组件。

每次只为当前可见的开放状态列表创建一个实例，不设置跨列表 `group`，从结构上禁止跨状态拖放。每个顶层 TODO 容器提供稳定的 `data-id`，并配置：

- `handle` 指向仅在手动模式显示的 Lucide 拖拽手柄。
- `draggable` 只匹配顶层 TODO，子项目和终端行不参与排序。
- `disabled` 在非手动模式或保存中为 `true`。
- `ghostClass`、`chosenClass` 和动画参数提供稳定占位及拖动反馈。
- 自动滚动指向现有 TODO 列表滚动容器。
- 固定使用 SortableJS fallback 指针路径，避免不同 WebView 对原生 HTML5 DnD 的事件差异，并让鼠标、触摸与自动滚动共用同一生命周期。

实例随当前视图、排序模式和组件生命周期安全启停；组件卸载前必须调用 `destroy()`。

拖拽手柄保持可聚焦，并支持 `ArrowUp`、`ArrowDown` 在当前状态列表内移动 TODO。保存期间手柄使用 `aria-disabled` 保留焦点，Sortable 实例和键盘 handler 同时阻止再次重排。

### 5. 拖动时只做视觉收起，不改展开状态集合

拖动开始前保存原顺序，并设置瞬时 `isTodoReordering` 标志。该标志通过列表级 CSS 隐藏所有顶层 TODO 的子树，但不修改组件现有的展开 ID 集合。这样拖动结束、取消、失焦或组件卸载时只需清除标志，所有 TODO 会自动恢复到各自原来的展开或收起状态，无需重建展开快照。

拖拽生命周期如下：

1. `onChoose` 关闭侧栏菜单和确认气泡，设置视觉收起标志，使列表高度在实际换位前稳定。
2. `onStart` 记录原有序 ID，并锁定冲突操作。
3. SortableJS 正常落放会先派发 `onUnchoose`、再派发 `onEnd`；`onUnchoose` 立即恢复视觉状态，但把原顺序快照延迟到当前调用栈结束后清理，使随后的 `onEnd` 仍可比较前后 ID 顺序。
4. `onEnd` 不依赖浏览器是否提供索引元数据，而是比较原顺序与 DOM `data-id`；顺序未变化时不保存。
5. `pointercancel`、`touchcancel` 或未发生 drop 的 `dragend` 会按取消处理：先恢复原 DOM 顺序，再执行幂等清理且不发出保存事件。视图切换和组件卸载也同步清理会话。

如果保存失败，`App.vue` 恢复原顺序；展开状态因从未被修改而保持正确。

侧栏以自身宽度作为 CSS container；在 240px 及以下把 TODO header 调整为两行，第一行保留分支、手柄和标题，第二行承载完整操作区，同时压缩排序标签字号，保证 220px 最小侧栏宽度下标题和控件不重叠。

### 6. 保留三种排序算法的清晰边界

`ProjectSidebar.vue` 的开放 TODO 比较器继续负责 `priority` 和 `time`。选择 `manual` 时不使用比较器，而是按照规范化 ID 列表建立 rank map 后排序，未记录 ID 稳定追加。`未执行` 和 `执行中` 使用不同 rank map。`sortedCompletedTodos` 保持现状，不读取开放列表排序模式或手动顺序。

## Risks / Trade-offs

- [拖动开始时列表项高度变化可能影响命中位置] → 在 `onChoose` 阶段先应用列表级视觉收起，并使用稳定的顶层容器、占位样式和组件测试验证索引。
- [项目状态与 UI 状态分属两个文件，无法跨文件事务提交] → 手动顺序不参与业务正确性；前后端都根据最新 Todo 集合确定性规范化，缺失项追加、失效项删除，避免依赖跨文件原子事务。
- [保存期间连续操作可能覆盖较新的顺序] → 保存完成前禁用拖拽和相关排序操作，并以后端返回的规范化状态作为最终状态。
- [多个 UI 偏好并发保存或工作区切换可能覆盖错误文件] → store 对读改写执行单锁事务，App 对 workspace 绑定执行读写锁，前端再以 scope epoch 和请求序号丢弃晚到响应。
- [外部拖拽依赖增加包体和维护面] → 直接依赖单一 `sortablejs` 包，只使用核心同列表排序和默认自动滚动能力，并锁定版本及提交 lockfile。
- [jsdom 无法完整模拟浏览器拖拽和滚动] → 把 SortableJS 初始化封装为可替换边界，组件测试直接触发生命周期回调并验证状态；实现完成后在 Wails/Vite 浏览器中验证真实拖动、占位和自动滚动。
- [旧版本应用读取新版 UI 状态后可能忽略手动字段] → 新字段保持可选且不改变既有字段；回退旧应用不会破坏 TODO 业务数据，但旧版本再次保存 UI 状态时可能丢失手动偏好。

## Migration Plan

1. 将工作区 UI 状态版本提升到 `2`，保存时写入新版本和可选排序字段。
2. 加载时先读取版本；版本 `0` 或 `1` 继续走现有 legacy 迁移，保留侧栏宽度和 TODO project 视图，并补充默认 `priority` 模式与空顺序。
3. 版本 `2` 直接解析新结构，再执行模式和 ID 规范化。无效或缺失字段不得导致整个 UI 状态加载失败。
4. 首次进入手动模式时才固化现有顺序，升级本身不主动改变用户看到的列表顺序。
5. 回滚到旧应用时无需迁移 `projects.json`；旧应用忽略新字段并继续使用原有自动排序。

## Open Questions

无。拖拽动画时长、边缘滚动灵敏度等视觉参数可在实现和人工验证中调整，但不得改变本设计规定的状态与持久化行为。
