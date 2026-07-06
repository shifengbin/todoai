## Context

当前应用存在三类终端上下文：workspace 全局终端、TODO project 终端、任务级终端。TODO project 终端通过 `activeTodoId`、`activeTodoProjectId`、`activeProjectId` 表达当前项目上下文；任务级终端只拥有 `todoId`，没有 `todoProjectId` 和 `projectId`。

现状中，选择任务级终端只更新 `activeTerminalId`，后端保留上一 TODO project/project 上下文。前端标题区域优先使用 `activeTodoProject` 或 `activeProject` 显示标题和路径，因此任务级终端激活后仍可能显示上一个项目的标题和目录。左侧 TODO 行的选中背景也只绑定 `activeTodoId`，所以任务级终端所属 TODO 不会稳定呈现为当前任务。

终端启动菜单当前渲染在 TODO 工作树内部，并使用 `position: absolute` 结合上下空间计算。即便能判断上翻和限高，菜单仍受侧栏滚动容器 `overflow-y: auto` 裁剪，长列表末尾的菜单会显示不全。

## Goals / Non-Goals

**Goals:**

- 选择或创建任务级终端后，应用进入该 TODO 的任务级上下文。
- 任务级上下文清空当前 TODO project/project 上下文，避免继续显示上一个项目。
- 顶部标题显示 TODO 任务级信息，路径显示该 TODO 的任务目录。
- 左侧 TODO 工作树在任务级终端激活时高亮父 TODO，并保持任务级终端行自身 active 状态。
- 终端启动菜单不被 TODO 列表滚动容器裁剪，长列表末尾仍可使用。
- 保持任务级终端归属模型不变：任务级终端属于 TODO，不属于任何 TODO project。

**Non-Goals:**

- 不为任务级终端新增 TODO project 或项目关联。
- 不改变 workspace 全局终端选择语义。
- 不改变 TODO project 终端的工作目录、标题和 Git 状态语义。
- 不引入新的 UI 组件库或外部定位依赖。

## Decisions

### 1. 后端将任务级终端选择为任务上下文

`SelectTerminal` 处理 `terminal.TodoID != "" && terminal.TodoProjectID == ""` 时，应加载当前 workspace state，确认 TODO 仍存在，然后返回：

- `ActiveTerminalID = terminal.ID`
- `ActiveTodoID = terminal.TodoID`
- `ActiveTodoProjectID = ""`
- `ActiveProjectID = ""`

`CreateTaskTerminal` 创建后也应返回同样的任务级上下文，因为新建任务终端会立即成为 active terminal。

这个选择比仅前端派生更稳，因为 Wails state 是前端单一状态来源，后端返回的 active 字段会被测试、刷新和后续操作复用。替代方案是前端看到 active task terminal 后临时覆盖标题和高亮，但会让状态与展示不一致，也不能解决刷新后 active TODO 不正确的问题。

### 2. 前端增加任务级标题和路径派生

`App.vue` 应优先判断 active terminal 是否为任务级终端。若是，标题显示为 `<TODO 标题> / 任务终端`，路径由当前 workspace 和 active TODO 的 `workspaceDirName` 派生为任务目录。

任务目录路径不需要新增 terminal 字段。TODO 数据已包含 `workspaceDirName`，任务级终端创建时会确保该字段存在；选择已存在任务级终端时，若字段缺失或 workspace 不可用，标题仍显示任务级上下文，路径显示为空或可用的缺失占位。

### 3. TODO 行 active 判定包含任务级 active terminal

`ProjectSidebar.vue` 应使用一个统一判断表达 TODO 是否处于当前上下文：

- `todo.id === activeTodoId`
- 或 active terminal 是该 TODO 下的任务级终端

该判断用于 TODO header 选中背景和必要的展开逻辑。任务级终端行自身仍通过 `terminal.id === activeTerminalId` 显示 active。

### 4. 终端启动菜单使用脱离滚动容器的浮层定位

终端启动菜单应使用 Vue `Teleport` 渲染到不受 TODO 列表滚动裁剪的层，例如组件本地创建的 fixed layer 或 `body`。菜单位置通过触发按钮 `getBoundingClientRect()` 计算，并使用侧栏/列表可视区域边界进行 clamp：

- 默认向下展开；
- 下方空间不足且上方空间更充足时向上展开；
- 上下空间都不足时设置最大高度并允许菜单内部滚动；
- 点击菜单项、点击外部、切换其他浮层或组件卸载时关闭菜单。

该方案优于修改滚动容器 `overflow`，因为侧栏仍需要稳定滚动，解除 overflow 会引入列表内容溢出和其它浮层叠层问题。

## Risks / Trade-offs

- 任务级终端清空 `activeTodoProjectId` 后，依赖当前 TODO project 的操作应进入禁用或空态。缓解：保留 `activeTerminalIsTaskTerminal` 分支，标题、终端 pane 和 Git 状态按 TODO 任务路径处理；项目专属操作继续依赖 active TODO project。
- 旧测试明确要求任务级终端不改变 TODO project context。缓解：用新的产品语义更新测试名称和断言。
- fixed 浮层需要处理滚动、窗口 resize 和组件卸载。缓解：打开时计算位置，滚动或 resize 时关闭菜单，避免位置陈旧。
- 任务目录路径依赖 `workspaceDirName`。缓解：任务级终端创建路径已经会持久化该字段；对旧数据缺失字段时展示任务上下文但不显示不可确认路径。
