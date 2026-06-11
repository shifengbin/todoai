## Context

当前应用是 Wails + Go 后端 + Vue 前端。左侧 `ProjectSidebar.vue` 展示 `TODO -> 项目 -> 终端` 树，`App.vue` 管理后端状态、表单 overlay、终端区域和 xterm session。TODO 创建已经支持描述、优先级和多工程关联，但后续浏览和维护能力还不完整：TODO 行优先级背景只应用在标题内容区域，优先级标签与背景重复表达，TODO 详情没有编辑入口，工程移除只能通过间接编辑设想完成，侧边栏宽度固定，长列表和终端区域之间缺少明确的尺寸协作。

终端运行时状态按 TODO project context 隔离。现有清理能力包含按 TODO 清理和按 project 全局清理，但编辑或删除单个 TODO 下的工程关联需要更细的 `todoProjectID` 粒度，避免误关同一工程在其它 TODO 下的终端。

## Goals / Non-Goals

**Goals:**

- 让 TODO item 的优先级颜色覆盖整条 header，并移除 item 上的优先级文字标签。
- 让 TODO 列表在侧边栏内部滚动，避免长列表挤压终端区域。
- 在 TODO item 上提供眼睛图标入口，用于查看和编辑 TODO 详情。
- 支持编辑 TODO 名称、描述、优先级和关联工程，并在保存时确认会关闭终端的移除操作。
- 支持在 TODO 工程行上直接移除工程，通过删除按钮旁的确认气泡完成确认。
- 新增按 `todoProjectID` 清理终端的后端能力，只影响被移除关联下的终端。
- 支持鼠标拖动调整侧边栏宽度，并在布局变化后重新适配活动终端。

**Non-Goals:**

- 不改变 TODO 完成、删除和归档快照语义。
- 不自动为新增关联工程创建终端；终端仍由用户显式创建。
- 不引入新的前端组件库或外部依赖。
- 不实现侧边栏宽度跨应用重启持久化，除非实现时已有低成本的本地设置模式可复用。

## Decisions

### 1. TODO 更新使用结构化请求并由后端计算关联 diff

新增 `UpdateTodoRequest`，字段包含 `id`、`title`、`description`、`priority`、`projectIds`。前端详情表单提交完整目标状态，后端负责 trim、校验、规范化优先级、验证项目存在、计算新增和移除的项目关联。

替代方案是在前端分别调用更新字段、添加工程和删除工程多个 API。这个方案会产生半完成状态，且前端需要理解关联 diff 和 active context 修正。用结构化请求可以让 TODO 元数据和工程关联在一个状态保存流程内完成。

### 2. 按 TODO-project 关联粒度清理终端

新增 shell manager 能力，例如 `DeleteTodoProjectTerminals(todoProjectID string)`。移除 TODO 下某个工程关联时，只关闭并移除该 `todoProjectID` 下的终端。`DeleteTodoTerminals(todoID)` 和 `DeleteProjectTerminals(projectID)` 保持现有语义，分别用于归档 TODO 和从项目库删除项目。

替代方案是复用 `DeleteProjectTerminals(projectID)`，但它会关闭该项目在所有 TODO 下的终端，违反当前 TODO 上下文隔离。

### 3. App 层编排状态更新和终端清理

Project manager 负责保存 TODO 和 TODO-project 状态，并返回被移除的 `todoProjectID` 集合。App 层在状态保存成功后调用 shell manager 清理对应终端，再返回 `withShellState(state)`。这样前端拿到的状态已经移除了失效终端。

如果保存失败，不清理终端；如果终端清理中已有 shell 退出，清理函数按现有 cleanup 语义幂等处理。

### 4. 两种移除入口复用同一后端路径

详情编辑保存和工程行直接删除都应走同一套核心逻辑。详情编辑可以一次移除多个工程；保存时若移除项下存在终端，前端统一弹确认。工程行直接删除只有单个目标，点击删除按钮后在按钮旁显示确认气泡，确认后调用移除 API。

确认气泡比 `window.confirm` 更贴近列表上下文，也避免阻塞整个应用 UI。

### 5. 侧边栏宽度由 App 层响应式控制

将 `.app-shell` 的固定 `280px` 列宽改为响应式 CSS 变量或 inline grid style。App 层保存当前拖动宽度并提供拖拽分隔条，宽度限制在合理范围内，例如 220px 到 520px。拖动过程中和结束后调用 `terminalManager.fitActive()`，让 xterm 重新计算 rows/cols，并通过现有 resize 回调同步 PTY。

该逻辑放在 App 层，因为它同时影响 `ProjectSidebar` 和 terminal surface。

## Risks / Trade-offs

- [移除工程后 active context 指向不存在的 TODO-project] → 后端在保存时修正 `ActiveTodoProjectID`、`ActiveProjectID` 和 `ActiveTerminalID`，优先选择同一 TODO 下最近的剩余工程，若不存在则清空 TODO-project 和 active terminal。
- [保存 TODO 编辑时误关其它 TODO 的终端] → 终端清理只接受 `todoProjectID`，测试覆盖同一 project 关联多个 TODO 的隔离场景。
- [确认气泡在滚动容器内被裁剪] → 气泡优先放在工程行内部并随行滚动；若空间不足，使用已有启动菜单的上下 placement 思路。
- [拖动侧边栏时频繁触发 PTY resize] → 拖动中用 `requestAnimationFrame` 或轻量节流调用 fit，拖动结束再执行一次完整 fit。
- [TODO item 操作按钮被优先级背景淹没] → 背景挂在整条 header 容器，按钮保持透明或轻微 hover 状态，并保留可读 focus/hover 样式。

## Migration Plan

1. 添加后端 TODO 更新请求、关联 diff 和按 `todoProjectID` 清理终端能力。
2. 更新 App API、Wails bindings 和前端调用。
3. 调整 `ProjectSidebar` 的 TODO 行结构、优先级背景、眼睛按钮、工程删除按钮和确认气泡。
4. 添加 TODO 详情编辑 overlay，并复用现有工程搜索多选和 tag 移除模式。
5. 添加侧边栏拖拽分隔条和终端 fit 逻辑。
6. 补充 Go 与前端测试，运行 Go 测试、前端测试和前端 build。

回滚时应同时回退前后端 API 和生成的 bindings。已持久化的 TODO 字段结构保持兼容，不需要数据迁移。

## Open Questions

无。详情编辑保存时确认、工程行气泡确认、按 TODO-project 粒度清理终端、侧边栏拖拽宽度均已确定。
