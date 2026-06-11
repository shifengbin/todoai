# todo-workspace Specification

## Purpose
TBD - created by archiving change add-todo-centric-workspace. Update Purpose after archive.
## Requirements
### Requirement: Display Workspace Tabs

系统 SHALL 在左侧工作区提供 `TODO` 和 `项目` 两个 tab。`TODO` tab SHALL 作为终端工作主入口，`项目` tab SHALL 作为项目库管理入口。

#### Scenario: User switches between workspace tabs

- **WHEN** 用户在左侧工作区点击 `TODO` tab
- **THEN** 系统显示 TODO 工作树
- **AND** 终端入口按 TODO 上下文展示
- **WHEN** 用户点击 `项目` tab
- **THEN** 系统显示项目库管理视图
- **AND** 项目库视图不显示可操作终端入口

### Requirement: Create Todo

系统 SHALL 允许用户通过应用内表单创建 TODO。创建表单 SHALL 包含必填 TODO 名称、选填 TODO 描述、任务优先级和选填工程。工程 SHALL 支持多选。任务优先级 SHALL 支持 `高`、`中`、`低` 三档，并 SHALL 默认选择 `中`。创建后的 TODO SHALL 出现在 TODO tab 的活动任务列表中，并 SHALL 可作为后续项目关联和终端创建的上下文。

#### Scenario: User creates a todo without a project

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为空、优先级为 `中` 且未选择工程的 TODO
- **THEN** TODO 活动列表包含 `修复登录问题`
- **AND** 该 TODO 可被选中
- **AND** 该 TODO 的优先级为 `中`
- **AND** 该 TODO 初始不包含关联项目

#### Scenario: User creates a todo with description and priority

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为 `登录后跳回首页`、优先级为 `高` 的 TODO
- **THEN** TODO 活动列表包含 `修复登录问题`
- **AND** 该 TODO 保存描述 `登录后跳回首页`
- **AND** 该 TODO 的优先级为 `高`

#### Scenario: User creates a todo with optional projects

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户在创建 TODO 表单中输入名称 `修复登录问题`
- **AND** 用户选择优先级 `高`
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** TODO 活动列表包含 `修复登录问题`
- **AND** TODO `修复登录问题` 下显示项目 `frontend-app`
- **AND** TODO `修复登录问题` 下显示项目 `api-service`
- **AND** 当前 TODO 项目上下文为 `修复登录问题` 下的 `frontend-app`

#### Scenario: Todo title is required

- **WHEN** 用户尝试创建名称为空的 TODO
- **THEN** 系统不创建 TODO
- **AND** 系统显示不会改变当前工作区状态的错误信息

#### Scenario: Project selection can be searched while creating todo

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户打开创建 TODO 表单
- **AND** 用户在工程筛选框输入 `front`
- **THEN** 工程选择列表显示 `frontend-app`
- **AND** 工程选择列表不显示 `api-service`

#### Scenario: Selected projects can be removed while creating todo

- **WHEN** 用户在创建 TODO 表单中选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** 创建 TODO 表单在工程筛选框上方以 tag 展示 `frontend-app`
- **AND** 创建 TODO 表单在工程筛选框上方以 tag 展示 `api-service`
- **WHEN** 用户删除 `frontend-app` tag
- **THEN** 创建 TODO 表单不再选中 `frontend-app`
- **AND** 创建 TODO 表单仍选中 `api-service`

### Requirement: Persist Todos

系统 SHALL 持久化 TODO、TODO 描述、TODO 优先级、TODO 与项目的关联、TODO 选中状态和归档状态，并 SHALL 在应用重启后恢复。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。

#### Scenario: Todo workspace is restored after restart

- **WHEN** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户关闭并重新打开应用
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** `frontend-app` 仍显示为该 TODO 下的关联项目

#### Scenario: Legacy todo without priority uses medium

- **WHEN** 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开应用
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

### Requirement: Associate Projects With Todo

系统 SHALL 允许用户从项目库中通过可搜索多选控件选择一个或多个项目关联到 TODO，并 SHALL 允许用户从 TODO 中移除已关联项目。同一个项目 SHALL 可关联到多个不同 TODO。项目选择 SHALL 支持按项目名称和路径筛选，且 SHALL 不要求用户手动输入完整项目名称。移除 TODO 下的工程关联 SHALL 只影响当前 TODO 下的该工程关联。

#### Scenario: User associates projects with a todo

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 用户为 TODO `修复登录问题` 选择这两个项目
- **THEN** TODO `修复登录问题` 下显示项目 `frontend-app`
- **AND** TODO `修复登录问题` 下显示项目 `api-service`

#### Scenario: Same project is associated with multiple todos

- **WHEN** 项目库包含 `frontend-app`
- **AND** 用户将 `frontend-app` 关联到 TODO `修复登录问题`
- **AND** 用户将 `frontend-app` 关联到 TODO `升级依赖`
- **THEN** `frontend-app` 同时显示在两个 TODO 下
- **AND** 两个 TODO 下的 `frontend-app` 关联互不替代

#### Scenario: Duplicate association is ignored

- **WHEN** TODO `修复登录问题` 已关联项目 `frontend-app`
- **AND** 用户再次将 `frontend-app` 关联到该 TODO
- **THEN** TODO `修复登录问题` 下只显示一个 `frontend-app` 关联

#### Scenario: User filters projects while associating a todo

- **WHEN** 项目库包含名称为 `frontend-app` 的项目
- **AND** 项目库包含路径为 `/work/api-service` 的项目
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户在工程筛选框输入 `api`
- **THEN** 工程选择列表显示 `/work/api-service` 对应项目
- **AND** 工程选择列表不显示 `frontend-app`

#### Scenario: User associates multiple filtered projects with a todo

- **WHEN** 项目库包含 `frontend-app`、`api-service` 和 `docs-site`
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户选择 `frontend-app`
- **AND** 用户选择 `api-service`
- **AND** 用户确认添加
- **THEN** TODO `修复登录问题` 下显示项目 `frontend-app`
- **AND** TODO `修复登录问题` 下显示项目 `api-service`
- **AND** TODO `修复登录问题` 下不新增 `docs-site`

#### Scenario: Already linked project is excluded from selectable projects

- **WHEN** TODO `修复登录问题` 已关联项目 `frontend-app`
- **AND** 项目库还包含 `api-service`
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **THEN** 工程选择列表不显示 `frontend-app`
- **AND** 工程选择列表显示 `api-service`

#### Scenario: Selected projects can be removed while associating a todo

- **WHEN** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** 添加工程控件在筛选框上方以 tag 展示 `frontend-app`
- **AND** 添加工程控件在筛选框上方以 tag 展示 `api-service`
- **WHEN** 用户删除 `api-service` tag
- **AND** 用户确认添加
- **THEN** TODO `修复登录问题` 下显示项目 `frontend-app`
- **AND** TODO `修复登录问题` 下不新增 `api-service`

#### Scenario: User removes project from todo list with popover confirmation

- **WHEN** TODO `修复登录问题` 下显示工程 `frontend-app`
- **AND** 用户点击 `frontend-app` 工程行上的删除按钮
- **THEN** 系统在删除按钮旁显示删除确认气泡
- **WHEN** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 下不再显示工程 `frontend-app`

#### Scenario: User cancels project removal popover

- **WHEN** TODO `修复登录问题` 下显示工程 `frontend-app`
- **AND** 用户点击 `frontend-app` 工程行上的删除按钮
- **AND** 系统显示删除确认气泡
- **WHEN** 用户取消删除
- **THEN** TODO `修复登录问题` 下仍显示工程 `frontend-app`

#### Scenario: Removing project from one todo preserves other todos

- **WHEN** 工程 `frontend-app` 同时关联到 TODO `修复登录问题` 和 TODO `升级依赖`
- **AND** 用户从 TODO `修复登录问题` 下移除工程 `frontend-app`
- **THEN** TODO `修复登录问题` 下不再显示工程 `frontend-app`
- **AND** TODO `升级依赖` 下仍显示工程 `frontend-app`

### Requirement: Display Todo Project Terminal Tree

系统 SHALL 在 TODO tab 中按 `TODO -> 项目 -> 终端` 层级展示活动任务、任务关联项目和运行时终端。

#### Scenario: Todo has projects and terminals

- **WHEN** TODO `修复登录问题` 关联项目 `frontend-app`
- **AND** `frontend-app` 在该 TODO 下有终端 `codex` 和 `npm run dev`
- **THEN** TODO tab 显示 `修复登录问题` 作为顶层任务
- **AND** `frontend-app` 显示为该 TODO 下的项目
- **AND** `codex` 和 `npm run dev` 显示为该 TODO 项目下的终端

#### Scenario: Project without terminals remains visible

- **WHEN** TODO `修复登录问题` 关联项目 `frontend-app`
- **AND** 该 TODO 项目下尚无终端
- **THEN** TODO tab 仍显示 `frontend-app`
- **AND** 用户可从该 TODO 项目行创建终端

### Requirement: Collapse Todo Branches

系统 SHALL 允许用户独立展开和收起 TODO 分支。收起 TODO SHALL 隐藏其项目和终端子项，但 SHALL 保留 TODO 行可见。

#### Scenario: User collapses a todo

- **WHEN** TODO `修复登录问题` 已展开并显示项目子项
- **AND** 用户激活该 TODO 的收起控件
- **THEN** 该 TODO 下的项目和终端子项被隐藏
- **AND** TODO `修复登录问题` 仍显示在活动任务列表中

#### Scenario: User expands a todo

- **WHEN** TODO `修复登录问题` 已收起
- **AND** 用户激活该 TODO 的展开控件
- **THEN** 系统显示该 TODO 下的项目子项

### Requirement: Select Todo Project Context

系统 SHALL 允许用户选择 TODO 下的项目作为当前工作上下文。选择 TODO 项目上下文 SHALL 更新当前 TODO、当前项目和当前 TODO 项目关联，但 SHALL 只使用该 TODO 项目下的终端集合。

#### Scenario: User selects a project under a todo

- **WHEN** 用户在 TODO `修复登录问题` 下选择项目 `frontend-app`
- **THEN** 当前 TODO 为 `修复登录问题`
- **AND** 当前项目为 `frontend-app`
- **AND** 终端区域只关联该 TODO 项目上下文中的终端

#### Scenario: Same project selected under different todos

- **WHEN** 项目 `frontend-app` 同时关联到 TODO `修复登录问题` 和 TODO `升级依赖`
- **AND** 用户选择 TODO `修复登录问题` 下的 `frontend-app`
- **THEN** 终端区域显示 `修复登录问题` 下 `frontend-app` 的终端集合
- **WHEN** 用户选择 TODO `升级依赖` 下的 `frontend-app`
- **THEN** 终端区域显示 `升级依赖` 下 `frontend-app` 的终端集合

### Requirement: Show Todo Project Terminal Launch Menu

系统 SHALL 在 TODO 项目行上提供终端启动菜单。启动菜单 SHALL 包含内置 `Terminal` 选项和已配置的终端启动配置。

#### Scenario: Todo project launch menu contains configured profiles

- **WHEN** 设置中包含启动配置 `codex` 和 `claude`
- **AND** 用户激活 TODO `修复登录问题` 下项目 `frontend-app` 的新增终端控件
- **THEN** 启动菜单显示 `Terminal` 作为第一项
- **AND** 启动菜单按配置顺序显示 `codex` 和 `claude`

#### Scenario: Unavailable todo project has no launch menu

- **WHEN** TODO `修复登录问题` 下的项目 `frontend-app` 路径不可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

### Requirement: Complete Todo

系统 SHALL 允许用户通过 TODO item 完成按钮旁的确认气泡完成活动 TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为已完成状态。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User completes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户在确认气泡中确认完成
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `completed`

#### Scenario: User cancels todo completion

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在活动任务列表中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO item 删除按钮旁的确认气泡删除活动 TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为删除状态。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `deleted`

#### Scenario: User cancels todo deletion

- **WHEN** 用户点击 TODO `修复登录问题` 的删除按钮
- **AND** 系统在删除按钮旁显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在活动任务列表中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供归档查看功能。归档列表 SHALL 显示已完成和已删除的 TODO，并 SHALL 展示归档时保存的项目快照。

#### Scenario: User views archived todos

- **WHEN** TODO `修复登录问题` 已归档
- **AND** 用户在 TODO tab 中打开归档视图
- **THEN** 归档视图显示 TODO `修复登录问题`
- **AND** 归档视图显示该 TODO 的归档原因和归档时间
- **AND** 归档视图显示该 TODO 归档时关联项目的名称和路径快照

#### Scenario: Archived todo does not restore terminals

- **WHEN** 用户打开归档视图
- **AND** 用户查看 TODO `修复登录问题`
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程

### Requirement: Preserve Archived Project Snapshots

系统 SHALL 在 TODO 归档时保存关联项目的名称、路径和项目 ID 快照。后续项目库变化 SHALL 不改变已归档 TODO 的项目快照。

#### Scenario: Archived todo keeps project snapshot after project deletion

- **WHEN** TODO `修复登录问题` 关联项目 `frontend-app`
- **AND** 用户完成该 TODO
- **AND** 用户随后从项目库删除 `frontend-app`
- **THEN** 归档视图中 TODO `修复登录问题` 仍显示 `frontend-app` 的归档名称和路径

### Requirement: Display Todo Priority Visuals

系统 SHALL 在 TODO 工作树中使用与优先级一致的颜色标记活动 TODO item。优先级 SHALL 包含 `高`、`中`、`低` 三档。TODO item header 背景 SHALL 使用与优先级一致的同色系背景，并 SHALL 覆盖该 item 的整条 header，包括展开控件、标题区域和操作按钮区域。TODO item 上 SHALL 不显示 `高`、`中`、`低` 文案标签。

#### Scenario: Todo row shows high priority styling

- **WHEN** 活动 TODO `修复登录问题` 的优先级为 `高`
- **THEN** TODO item header 使用高优先级对应的红色系背景覆盖整条 header
- **AND** TODO item 上不显示 `高` 文案标签

#### Scenario: Todo row shows medium priority styling

- **WHEN** 活动 TODO `升级依赖` 的优先级为 `中`
- **THEN** TODO item header 使用中优先级对应的橙色系背景覆盖整条 header
- **AND** TODO item 上不显示 `中` 文案标签

#### Scenario: Todo row shows low priority styling

- **WHEN** 活动 TODO `整理文档` 的优先级为 `低`
- **THEN** TODO item header 使用低优先级对应的绿色系背景覆盖整条 header
- **AND** TODO item 上不显示 `低` 文案标签

### Requirement: Edit Todo Details

系统 SHALL 允许用户从活动 TODO item 的眼睛图标进入 TODO 详情，并 SHALL 在详情中查看和编辑 TODO 名称、描述、优先级和关联工程。保存编辑 SHALL 持久化更新后的 TODO 字段和工程关联。

#### Scenario: User opens todo details from item

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的眼睛图标
- **THEN** 系统打开 TODO 详情编辑界面
- **AND** 详情界面显示 TODO 名称、描述、优先级和已关联工程

#### Scenario: User edits todo metadata

- **WHEN** 用户打开 TODO `修复登录问题` 的详情
- **AND** 用户将名称改为 `修复登录跳转`
- **AND** 用户将描述改为 `登录后跳回首页`
- **AND** 用户将优先级改为 `高`
- **AND** 用户保存详情
- **THEN** 活动 TODO 列表显示 `修复登录跳转`
- **AND** TODO 描述保存为 `登录后跳回首页`
- **AND** TODO 优先级保存为 `高`

#### Scenario: User edits todo projects without removed terminals

- **WHEN** TODO `修复登录问题` 已关联工程 `frontend-app`
- **AND** 项目库还包含工程 `api-service`
- **AND** 用户在详情中新增工程 `api-service`
- **AND** 用户保存详情
- **THEN** TODO `修复登录问题` 下显示工程 `frontend-app`
- **AND** TODO `修复登录问题` 下显示工程 `api-service`
- **AND** 系统不自动为 `api-service` 创建终端

#### Scenario: Saving detail edit confirms removed project terminals

- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 存在终端
- **AND** 用户在详情中移除工程 `frontend-app`
- **AND** 用户点击保存
- **THEN** 系统在保存前提示移除该工程会关闭其在当前 TODO 下的终端
- **WHEN** 用户确认保存
- **THEN** 系统移除 TODO `修复登录问题` 下的工程 `frontend-app`
- **AND** 系统关闭并移除 `frontend-app` 在该 TODO 下的终端

#### Scenario: User cancels terminal-closing detail save

- **WHEN** TODO `修复登录问题` 下的工程 `frontend-app` 存在终端
- **AND** 用户在详情中移除工程 `frontend-app`
- **AND** 用户点击保存
- **AND** 用户取消关闭终端确认
- **THEN** 详情编辑界面保持打开
- **AND** TODO `修复登录问题` 下的工程 `frontend-app` 仍保持关联
- **AND** `frontend-app` 在该 TODO 下的终端保持运行时状态不变

### Requirement: Scroll Todo Workspace List

系统 SHALL 在 TODO 列表内容超过侧边栏可见高度时，让 TODO tab 内容在侧边栏内部滚动。滚动 SHALL 不改变终端区域高度或宽度。

#### Scenario: Long todo list scrolls inside sidebar

- **WHEN** 活动 TODO 列表内容超过侧边栏高度
- **THEN** TODO tab 内部出现可滚动区域
- **AND** 侧边栏 header 和 tab 控件保持可见
- **AND** 终端区域尺寸不因 TODO 列表长度被挤压

### Requirement: Resize Workspace Sidebar

系统 SHALL 允许用户通过鼠标拖动侧边栏和终端区域之间的分隔条调整侧边栏宽度。调整侧边栏宽度 SHALL 同时改变终端区域宽度，并 SHALL 重新适配活动终端尺寸。

#### Scenario: User drags sidebar wider

- **WHEN** 用户向右拖动侧边栏分隔条
- **THEN** 侧边栏宽度增加
- **AND** 终端区域宽度相应减少
- **AND** 活动终端重新适配新的可见尺寸

#### Scenario: Sidebar resize respects width limits

- **WHEN** 用户拖动侧边栏分隔条超过允许的最小或最大宽度
- **THEN** 系统将侧边栏宽度限制在允许范围内
- **AND** 终端区域仍保持可用宽度

### Requirement: Manage Todo Action Confirmation Popovers

系统 SHALL 在 TODO item 的完成和删除操作上使用按钮旁确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。

#### Scenario: Complete confirmation popover opens beside complete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的完成按钮
- **THEN** 系统在完成按钮旁显示完成确认气泡
- **AND** 系统不立即完成 TODO `修复登录问题`

#### Scenario: Delete confirmation popover opens beside delete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的删除按钮
- **THEN** 系统在删除按钮旁显示删除确认气泡
- **AND** 系统不立即删除 TODO `修复登录问题`

#### Scenario: Opening another sidebar popover closes todo action popover

- **WHEN** TODO `修复登录问题` 的完成确认气泡已显示
- **AND** 用户打开 TODO 下项目 `frontend-app` 的终端启动菜单
- **THEN** 系统关闭 TODO `修复登录问题` 的完成确认气泡
- **AND** 系统显示 `frontend-app` 的终端启动菜单

#### Scenario: Outside click closes todo action popover without action

- **WHEN** TODO `修复登录问题` 的删除确认气泡已显示
- **AND** 用户点击确认气泡外部
- **THEN** 系统关闭删除确认气泡
- **AND** TODO `修复登录问题` 保持在活动任务列表中

### Requirement: Display Todo Project Row State

系统 SHALL 在 TODO 工作树中用整行背景表达 TODO 下项目行的悬停和选中状态。TODO 项目行背景 SHALL 覆盖该项目 header 的项目信息区域、创建终端按钮区域和删除按钮区域。TODO 项目行上的创建终端按钮和删除按钮 SHALL 在整行背景上保持可读，并 SHALL 保留各自的 hover 和 focus 反馈。

#### Scenario: Active todo project row background covers action buttons

- **WHEN** TODO `修复登录问题` 下的项目 `frontend-app` 是当前选中的 TODO 项目上下文
- **THEN** `frontend-app` 项目行的选中背景覆盖项目名称和路径区域
- **AND** 选中背景覆盖创建终端按钮区域
- **AND** 选中背景覆盖删除按钮区域

#### Scenario: Hovered todo project row background covers action buttons

- **WHEN** 用户将鼠标悬停在 TODO `修复登录问题` 下的项目 `frontend-app` 行上
- **THEN** `frontend-app` 项目行的悬停背景覆盖项目名称和路径区域
- **AND** 悬停背景覆盖创建终端按钮区域
- **AND** 悬停背景覆盖删除按钮区域

