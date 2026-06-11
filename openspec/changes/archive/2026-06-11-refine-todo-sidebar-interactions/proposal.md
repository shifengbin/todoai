## Why

TODO 工作区的侧边栏已经承载任务、工程和终端的主要导航，但当前列表在长任务、多工程和终端清理场景下不够稳定：优先级背景只覆盖 TODO 名称区域，长列表滚动和侧边栏宽度不可控，TODO 详情编辑和工程移除缺少明确入口。随着 TODO 成为主要工作上下文，需要把侧边栏交互补齐，并保证移除工程时不会留下已失效终端。

## What Changes

- TODO item 的优先级背景 SHALL 覆盖整条 item header，而不是只覆盖名称或文案范围。
- TODO 列表 SHALL 在内容过长时在侧边栏内部滚动，不挤压终端区域。
- TODO item SHALL 不再显示 `高`、`中`、`低` 文案标签；优先级继续通过颜色表达。
- TODO item SHALL 提供眼睛图标按钮，用于打开查看和编辑 TODO 详情。
- TODO 详情 SHALL 支持编辑名称、描述、优先级和关联工程。
- 编辑详情保存时，若移除了带终端的工程，系统 SHALL 在保存时确认，并在用户确认后关闭对应终端。
- TODO 下的工程列表 SHALL 支持不进入详情直接移除工程；删除按钮旁 SHALL 弹出确认气泡。
- 直接移除 TODO 工程时，系统 SHALL 只移除当前 TODO 下的工程关联，并清理该关联下的终端，不影响同一工程在其它 TODO 下的关联和终端。
- 侧边栏 SHALL 支持鼠标拖动调整宽度，同时终端区域宽度随之调整并重新适配活动终端尺寸。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 修改 TODO 工作树的列表呈现、详情编辑、工程移除确认和关联终端清理要求。
- `embedded-shell-sessions`: 修改终端清理要求，增加按 TODO-project 关联粒度关闭终端的行为。

## Impact

- Go 后端需要新增 TODO 更新请求和按 TODO-project 关联移除工程的 API。
- Shell session 管理需要新增按 `todoProjectID` 清理终端的能力，并复用现有进程关闭和 session cleanup 语义。
- Wails bindings 需要重新生成，前端调用需要同步更新。
- `App.vue` 需要增加 TODO 详情编辑状态、保存确认、侧边栏拖拽宽度状态和终端 fit 调用。
- `ProjectSidebar.vue` 需要调整 TODO item 结构、优先级背景范围、眼睛图标按钮、工程行删除按钮和确认气泡。
- 前端 CSS 需要支持整行背景、滚动区域、确认气泡、可拖拽分隔条和不同侧边栏宽度下的文本截断。
- Go 和前端测试需要覆盖 TODO 更新、工程移除、终端清理、确认交互、滚动布局和可拖拽宽度行为。
