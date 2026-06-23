## Context

应用当前以 workspace 为数据边界，以 TODO project 作为主要工作上下文。全局项目候选只用于创建或编辑 TODO 时选择项目，选择全局候选不应创建终端或切换工作上下文。终端会话目前主要归属 TODO project context，`ProjectTerminal` 通过 `projectId`、`todoId`、`todoProjectId` 标识所属上下文，前端用 `activeTerminalId` 驱动 xterm pane 激活。

这次优化横跨项目候选选择、终端上下文和 TODO 树交互。设计上需要保持现有语义：导入候选不等于关联 TODO，临时全局终端不等于选择项目，展开/收起手势不等于打开详情或选择终端。

## Goals / Non-Goals

**Goals:**

- 单个导入项目候选后，在当前项目选择控件中默认选中新候选，减少重复点击。
- 支持多个 workspace 级全局终端，默认工作目录为当前 workspace 根目录。
- 在终端区域顶部提供全局终端分组；没有全局终端时该区域完全隐藏。
- 全局终端和 TODO project 终端共享 xterm、输入、resize、clipboard、重启、删除等基础能力，但状态上下文相互隔离。
- TODO header 行支持双击展开/收起，行内操作按钮继续按原行为工作。

**Non-Goals:**

- 不把单个导入的项目自动关联到 TODO；仍由用户提交创建、保存或添加操作后才生效。
- 不恢复左侧独立项目库视图。
- 不让全局终端参与项目 Git 状态刷新，也不改变当前 TODO project 选择。
- 不重构所有终端上下文为通用多态模型；只做满足 workspace 全局终端的最小模型扩展。

## Decisions

### 使用 workspace terminal 上下文，而不是伪项目

全局终端 SHALL 使用明确的 workspace 级上下文标识，例如 `workspaceTerminal` 或等价字段，而不是创建一个指向 workspace 根目录的全局项目候选。这样可以避免污染项目候选列表，也避免全局终端选择触发项目 Git 状态、TODO project 激活或项目删除联动。

备选方案是用伪项目复用现有 `CreateTerminal(projectID)`。该方案实现短，但会让“临时命令入口”看起来像项目，且容易被项目删除、Git 状态刷新和 TODO 选择逻辑误处理，因此不采用。

### 终端基础设施复用，选择语义分离

后端 shell manager 继续负责 shell 启动、输入、resize、输出和删除。新增 workspace 终端创建 API 使用当前 workspace 根目录作为 `WorkingDir`，并给终端写入全局上下文标识。选择全局终端时只更新 active terminal，不调用 `SelectProject` 或 `SelectTodoProject`。

前端仍由统一 `terminalManager` 挂载和激活 xterm pane。终端列表需要区分 workspace 全局终端与 TODO project 终端：全局终端只显示在终端区域顶部的 Global 分组，TODO project 终端仍显示在左侧 TODO 树中。

### Global 分组按存在性渲染

终端区域顶部的 Global 分组只在存在至少一个全局终端时渲染。创建入口应在终端区域工具栏中可用，但空列表状态不应占据额外高度。这样满足“没有全局终端就消失”的体验要求。

### 单个导入默认选中只作用于当前打开的选择控件

`CreateProjectFromDialog` 返回更新后的项目列表和导入摘要后，前端根据 import summary 定位导入目录对应的项目 ID，并把它加入当前打开控件的选择数组：创建 TODO、编辑 TODO 或添加项目弹窗。若 summary 表示项目已存在，前端使用 skipped path 精确匹配已有候选并默认选中。若没有这些控件打开，则只更新候选列表，不产生默认选择。

### TODO 双击绑定在 header 行

双击事件绑定在 TODO header 行，内部按钮和菜单继续使用 stop propagation 或事件目标过滤，避免双击删除、完成、菜单等控件时误触发展开/收起。双击只更新折叠状态，不选择 TODO、不打开详情、不影响终端。

## Risks / Trade-offs

- 全局终端上下文字段会影响历史持久化和 Wails 模型生成 -> 通过后端测试覆盖创建、选择、删除、恢复，并重新生成前端绑定。
- 选择全局终端不改变项目上下文，可能让 header 仍显示上一个 TODO project -> 终端区域应通过 Global 分组和 active terminal 状态表达当前激活的是全局终端，项目 header 不应被全局终端重定向。
- 导入默认选中依赖识别导入目录对应候选 -> 优先使用 import summary 中的 added 项目；若 summary 表示重复导入，则使用 skipped path 精确匹配已有候选，避免误选无关候选。
- 双击与单击按钮可能在浏览器事件顺序上互相影响 -> 按钮区域停止冒泡，并用前端测试覆盖双击 header 与双击按钮的差异。

## Migration Plan

新增字段应向后兼容旧终端历史：缺少 workspace 全局标识的旧记录仍按现有 TODO project 或 project 终端处理。打开旧 workspace 时不会自动创建全局终端。回滚时，全局终端历史可被忽略，不影响 TODO 和项目候选数据。

## Open Questions

- 全局终端标签是否只显示 shell/current command，还是需要显示 workspace 名称；初版按现有终端标签规则显示 shell/current command。
