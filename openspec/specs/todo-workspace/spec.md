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

系统 SHALL 允许用户通过应用内表单创建 TODO。创建表单 SHALL 包含必填 TODO 名称、选填 TODO 描述、任务优先级和选填工程。工程 SHALL 支持多选。任务优先级 SHALL 支持 `高`、`中`、`低` 三档，并 SHALL 默认选择 `中`。创建后的 TODO SHALL 默认保存为 `not-started` 状态，出现在 `未执行` 视图中，并 SHALL 可作为后续项目关联和终端创建的上下文。创建后的 TODO 分支 SHALL 默认收起，且创建操作 SHALL 不改变当前 TODO、当前项目、当前 TODO 项目或当前终端上下文。

#### Scenario: User creates a todo without a project

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为空、优先级为 `中` 且未选择工程的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 该 TODO 的优先级为 `中`
- **AND** 该 TODO 初始不包含关联项目
- **AND** TODO `修复登录问题` 的分支默认收起

#### Scenario: User creates a todo with description and priority

- **WHEN** 用户在 TODO tab 中创建名称为 `修复登录问题`、描述为 `登录后跳回首页`、优先级为 `高` 的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** 该 TODO 保存描述 `登录后跳回首页`
- **AND** 该 TODO 的优先级为 `高`
- **AND** 该 TODO 的状态为 `not-started`

#### Scenario: User creates a todo with optional projects

- **WHEN** 项目库包含 `frontend-app` 和 `api-service`
- **AND** 当前 TODO 项目上下文为 TODO `升级依赖` 下的 `docs-site`
- **AND** 用户在创建 TODO 表单中输入名称 `修复登录问题`
- **AND** 用户选择优先级 `高`
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 下保存项目 `frontend-app`
- **AND** TODO `修复登录问题` 下保存项目 `api-service`
- **AND** 当前 TODO 项目上下文仍为 TODO `升级依赖` 下的 `docs-site`
- **AND** TODO `修复登录问题` 的分支默认收起

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

系统 SHALL 持久化 TODO、TODO 描述、TODO 优先级、TODO 工作流状态、TODO 与项目的关联、TODO 选中状态和已完成状态，并 SHALL 在应用重启后恢复。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。旧数据中状态为 `active` 的 TODO SHALL 按 `not-started` 处理。旧数据中状态为 `archived` 且归档原因为 `completed` 的 TODO SHALL 按 `completed` 处理。旧数据中状态为 `archived` 且归档原因为 `deleted` 的 TODO SHALL 不在 TODO 工作区列表中展示。

#### Scenario: Todo workspace is restored after restart

- **WHEN** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户关闭并重新打开应用
- **THEN** `执行中` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** TODO `修复登录问题` 的状态仍为 `in-progress`
- **AND** `frontend-app` 仍保存为该 TODO 下的关联项目

#### Scenario: Legacy todo without priority uses medium

- **WHEN** 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开应用
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

#### Scenario: Legacy active todo becomes not-started

- **WHEN** 持久化数据中 TODO `修复登录问题` 的状态为 `active`
- **AND** 用户打开应用
- **THEN** `未执行` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `not-started` 处理

#### Scenario: Legacy completed archived todo remains completed

- **WHEN** 持久化数据中 TODO `修复登录问题` 的状态为 `archived`
- **AND** TODO `修复登录问题` 的归档原因为 `completed`
- **AND** 用户打开应用
- **THEN** `已完成` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `completed` 处理

#### Scenario: Legacy deleted archived todo is hidden

- **WHEN** 持久化数据中 TODO `废弃任务` 的状态为 `archived`
- **AND** TODO `废弃任务` 的归档原因为 `deleted`
- **AND** 用户打开应用
- **THEN** TODO tab 不在 `未执行` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `执行中` 视图显示 `废弃任务`
- **AND** TODO tab 不在 `已完成` 视图显示 `废弃任务`

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

系统 SHALL 允许用户独立展开和收起 `未执行` 与 `执行中` 视图中的 TODO 分支。收起 TODO SHALL 隐藏其项目和终端子项，但 SHALL 保留 TODO 行可见。若收起的 TODO 下存在终端，TODO 行 SHALL 反映被隐藏子终端的聚合活动状态；聚合优先级 SHALL 为 `needs-input` 高于 `busy` 高于 `idle`。折叠 TODO 的非空聚合活动状态 SHALL 使用覆盖 TODO item 整行的呼吸式状态反馈表达，并 SHALL 区分 `busy` 与 `needs-input`。折叠 TODO 行 MUST NOT 复用终端行的转圈或警告活动图标来表达聚合状态。

#### Scenario: User collapses a todo

- **WHEN** TODO `修复登录问题` 已展开并显示项目子项
- **AND** 用户激活该 TODO 的收起控件
- **THEN** 该 TODO 下的项目和终端子项被隐藏
- **AND** TODO `修复登录问题` 仍显示在当前状态视图中

#### Scenario: Collapsed todo shows hidden terminal busy as row breathing

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行呼吸式状态反馈显示运行中的聚合活动状态
- **AND** TODO `修复登录问题` 行不显示终端行的转圈活动图标

#### Scenario: Collapsed todo shows hidden terminal needing input as row breathing

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行使用整行呼吸式状态反馈显示等待输入的聚合活动状态
- **AND** 等待输入的整行状态反馈与运行中的整行状态反馈可区分
- **AND** TODO `修复登录问题` 行不显示终端行的警告活动图标

#### Scenario: Collapsed todo prioritizes needs input over busy

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 下还存在活动状态为 `needs-input` 的终端
- **AND** TODO `修复登录问题` 已收起
- **THEN** TODO `修复登录问题` 行显示等待输入的整行呼吸式状态反馈
- **AND** TODO `修复登录问题` 行不以运行中状态作为最高优先级提示

#### Scenario: Expanded todo relies on terminal rows for activity state

- **WHEN** TODO `修复登录问题` 下存在活动状态为 `busy` 的终端
- **AND** TODO `修复登录问题` 已展开
- **THEN** 该终端行显示运行中的活动提示
- **AND** TODO `修复登录问题` 行不重复显示收起态聚合活动提示

#### Scenario: Collapsed todo without active hidden terminal has no breathing feedback

- **WHEN** TODO `修复登录问题` 已收起
- **AND** TODO `修复登录问题` 下不存在活动状态为 `busy` 或 `needs-input` 的终端
- **THEN** TODO `修复登录问题` 行不显示整行呼吸式状态反馈
- **AND** TODO `修复登录问题` 行不显示终端活动图标

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

系统 SHALL 仅在 `in-progress` TODO 的可用项目行上提供终端启动菜单。启动菜单 SHALL 包含内置 `Terminal` 选项和已配置的终端启动配置。`not-started` TODO 的项目行 SHALL NOT 暴露新增终端启动菜单。

#### Scenario: In-progress todo project launch menu contains configured profiles

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 设置中包含启动配置 `codex` 和 `claude`
- **AND** 用户激活 TODO `修复登录问题` 下项目 `frontend-app` 的新增终端控件
- **THEN** 启动菜单显示 `Terminal` 作为第一项
- **AND** 启动菜单按配置顺序显示 `codex` 和 `claude`

#### Scenario: Not-started todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

#### Scenario: Unavailable todo project has no launch menu

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下的项目 `frontend-app` 路径不可用
- **THEN** 该 TODO 项目行不暴露新增终端启动菜单

### Requirement: Complete Todo

系统 SHALL 允许用户通过 TODO item 完成按钮旁的确认气泡完成 `in-progress` TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 状态保存为 `completed`。用户取消确认 SHALL 不改变 TODO 或其终端状态。系统 SHALL NOT 允许 `not-started` TODO 直接完成。

#### Scenario: User completes an in-progress todo

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户在确认气泡中确认完成
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 显示在 `已完成` 视图
- **AND** TODO `修复登录问题` 的状态为 `completed`

#### Scenario: Not-started todo cannot be completed

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户或客户端请求完成 TODO `修复登录问题`
- **THEN** 系统拒绝完成请求
- **AND** TODO `修复登录问题` 的状态保持为 `not-started`

#### Scenario: User cancels todo completion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户点击 TODO `修复登录问题` 的完成按钮
- **AND** 系统在完成按钮旁显示完成确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: Delete Todo

系统 SHALL 允许用户通过 TODO 右键菜单中的删除动作和确认气泡删除 `not-started`、`in-progress` 或 `completed` TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 从可见 TODO 工作区列表中移除。用户取消确认 SHALL 不改变 TODO 或其终端状态。

#### Scenario: User deletes a not-started todo

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 不再显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User deletes an in-progress todo

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

#### Scenario: User deletes a completed todo

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户在 `已完成` 视图中打开 TODO `修复登录问题` 的操作菜单
- **AND** 用户选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 不再显示在 `已完成` 视图
- **AND** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程

#### Scenario: User cancels todo deletion

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 TODO `修复登录问题` 的右键菜单中选择删除 TODO
- **AND** 系统显示删除确认气泡
- **AND** 用户取消确认
- **THEN** TODO `修复登录问题` 保持在 `执行中` 视图中
- **AND** 该 TODO 下的终端保持运行时状态不变

### Requirement: View Archived Todos

系统 SHALL 在 TODO tab 中提供已完成查看功能。已完成列表 SHALL 只显示状态为 `completed` 的 TODO，并 SHALL 按完成时间倒序展示，最近完成的 TODO 排在前面。完成时间 SHALL 优先使用 `completedAt`，当 `completedAt` 缺失时 SHALL 使用 `archivedAt` 作为兼容旧数据的兜底。缺失有效完成时间的已完成 TODO SHALL 排在有完成时间的 TODO 之后。已完成列表 SHALL 展示完成时保存的项目快照。已删除 TODO SHALL NOT 显示在已完成列表中。

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 已完成
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图显示 TODO `修复登录问题`
- **AND** `已完成` 视图显示该 TODO 的完成时间
- **AND** `已完成` 视图显示该 TODO 完成时关联项目的名称和路径快照

#### Scenario: Completed todos are ordered by newest completion time

- **WHEN** `已完成` 视图包含 TODO `整理文档`
- **AND** TODO `整理文档` 的 `completedAt` 为 `2026-06-14T09:00:00Z`
- **AND** `已完成` 视图包含 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `修复登录问题` 排在 TODO `整理文档` 前面

#### Scenario: Completed todo order falls back to archivedAt

- **WHEN** `已完成` 视图包含 TODO `旧任务`
- **AND** TODO `旧任务` 不包含 `completedAt`
- **AND** TODO `旧任务` 的 `archivedAt` 为 `2026-06-15T10:00:00Z`
- **AND** `已完成` 视图包含 TODO `较早任务`
- **AND** TODO `较早任务` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `旧任务` 排在 TODO `较早任务` 前面

#### Scenario: Completed todo without completion time is ordered last

- **WHEN** `已完成` 视图包含 TODO `缺失时间任务`
- **AND** TODO `缺失时间任务` 不包含有效的 `completedAt` 或 `archivedAt`
- **AND** `已完成` 视图包含 TODO `有时间任务`
- **AND** TODO `有时间任务` 的 `completedAt` 为 `2026-06-15T09:00:00Z`
- **THEN** `已完成` 视图中 TODO `有时间任务` 排在 TODO `缺失时间任务` 前面

#### Scenario: Deleted todo is not shown as completed

- **WHEN** TODO `废弃任务` 已被删除
- **AND** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** `已完成` 视图不显示 TODO `废弃任务`

#### Scenario: Completed todo does not restore terminals

- **WHEN** 用户打开 `已完成` 视图
- **AND** 用户查看 TODO `修复登录问题`
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程

### Requirement: Stabilize Todo Workspace Header Layout

系统 SHALL 在 TODO tab 的 `未执行`、`执行中`、`已完成` 三个状态视图之间保持顶部控制区高度一致。顶部状态切换栏 SHALL 在三个状态视图之间保持相同宽度分配。开放 TODO 专用的排序、批量收起和批量展开控件 SHALL 只在 `未执行` 与 `执行中` 视图中可见且可交互，`已完成` 视图 SHALL 保留等高布局占位但不暴露这些开放 TODO 操作。

#### Scenario: Completed view keeps header height stable

- **WHEN** 用户在 TODO tab 中查看 `未执行` 视图
- **AND** 用户点击 `已完成` 状态按钮
- **THEN** TODO 工作区顶部控制区高度保持不变
- **AND** TODO 列表内容不会因顶部控制区高度变小而上移

#### Scenario: Completed view keeps workflow tab widths stable

- **WHEN** 用户在 TODO tab 中查看 `未执行` 视图
- **AND** 用户点击 `已完成` 状态按钮
- **THEN** `未执行`、`执行中`、`已完成` 三个状态按钮的宽度分配保持不变

#### Scenario: Completed view does not expose open todo controls

- **WHEN** 用户在 TODO tab 中打开 `已完成` 视图
- **THEN** 系统不显示可操作的开放 TODO 排序控件
- **AND** 系统不显示可操作的批量收起 TODO 控件
- **AND** 系统不显示可操作的批量展开 TODO 控件

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

### Requirement: Preview Todo Description On Hover

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行中保留描述摘要，并 SHALL 允许用户通过鼠标悬浮查看完整 TODO 描述。仅当 TODO 存在非空描述时，系统 SHALL 在鼠标悬浮于 TODO 行一段时间后显示完整描述 tooltip；鼠标移开后 tooltip SHALL 消失。tooltip SHALL 不要求用户打开 TODO 详情，且 SHALL 不改变 TODO 行的展开收起、菜单和状态操作行为。

#### Scenario: Todo row shows description summary

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页，需要保留原始跳转地址`
- **THEN** TODO 行显示该描述的摘要

#### Scenario: User previews full todo description after hover delay

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页，需要保留原始跳转地址`
- **AND** 用户将鼠标悬浮在 TODO `修复登录问题` 行上但尚未达到显示延迟
- **THEN** 系统不显示完整描述 tooltip
- **WHEN** 悬浮时间达到显示延迟
- **THEN** 系统显示包含完整描述 `登录后跳回首页，需要保留原始跳转地址` 的 tooltip

#### Scenario: Todo description tooltip hides on mouse leave

- **WHEN** TODO `修复登录问题` 的完整描述 tooltip 已显示
- **AND** 用户将鼠标移出 TODO `修复登录问题` 行
- **THEN** 系统隐藏该 tooltip

#### Scenario: Todo without description has no tooltip

- **WHEN** 活动 TODO 列表显示 TODO `整理文档`
- **AND** TODO `整理文档` 的描述为空
- **AND** 用户将鼠标悬浮在 TODO `整理文档` 行上
- **THEN** 系统不显示描述 tooltip

### Requirement: Edit Todo Details

系统 SHALL 允许用户从活动 TODO item 的右键菜单进入 TODO 详情，并 SHALL 在详情中查看和编辑 TODO 名称、描述、优先级和关联工程。保存编辑 SHALL 持久化更新后的 TODO 字段和工程关联。

#### Scenario: User opens todo details from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择查看详情
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

系统 SHALL 在 TODO item 的完成按钮和右键菜单删除操作上使用确认气泡。系统 SHALL 在同一时间最多显示一个侧边栏浮层；打开 TODO 操作确认气泡 SHALL 关闭 TODO 右键菜单、终端启动菜单和 TODO 项目移除确认气泡，打开其它侧边栏浮层 SHALL 关闭 TODO 操作确认气泡。确认气泡 SHALL 支持取消、点击外部关闭和确认成功后关闭。删除确认气泡 SHALL 锚定在对应 TODO item 的操作上下文附近，且 SHALL NOT 脱离当前 TODO item 显示到侧边栏右上角或应用全局右上角。

#### Scenario: Complete confirmation popover opens beside complete action

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击该 TODO item 上的完成按钮
- **THEN** 系统在完成按钮旁显示完成确认气泡
- **AND** 系统不立即完成 TODO `修复登录问题`

#### Scenario: Delete confirmation popover opens from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择删除 TODO
- **THEN** 系统关闭 TODO 右键菜单
- **AND** 系统在 TODO `修复登录问题` 的操作上下文附近显示删除确认气泡
- **AND** 系统不立即删除 TODO `修复登录问题`

#### Scenario: Delete confirmation popover stays anchored to the todo item

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在该 TODO item 的右键菜单中选择删除 TODO
- **THEN** 删除确认气泡显示在 TODO `修复登录问题` 所属操作区域附近
- **AND** 删除确认气泡不显示在侧边栏右上角
- **AND** 删除确认气泡不显示在应用窗口全局右上角

#### Scenario: Windows delete confirmation popover remains near trigger context

- **WHEN** 应用运行在 Windows
- **AND** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户通过右键菜单或三点菜单选择删除 TODO
- **THEN** 系统在 TODO `修复登录问题` 的操作上下文附近显示删除确认气泡
- **AND** 删除确认气泡不锚定到侧边栏或应用窗口的全局角落

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

### Requirement: Sort Active Todos

系统 SHALL 在 TODO tab 的 `未执行` 和 `执行中` 列表中提供排序切换控件，并 SHALL 支持按任务优先级排序和按创建时间排序。系统 SHALL 默认选择优先级排序。优先级排序 SHALL 为 `高`、`中`、`低`，相同优先级的 TODO SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。时间排序 SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。`已完成` 列表 SHALL 不受 `未执行` 和 `执行中` 的排序规则影响。

#### Scenario: Open todo sort control defaults to priority

- **WHEN** 用户打开 TODO tab 的 `未执行` 或 `执行中` 列表
- **THEN** 系统显示 TODO 排序切换控件
- **AND** 排序切换控件默认选中优先级排序

#### Scenario: Not-started todos are ordered by priority

- **WHEN** `未执行` 列表包含优先级为 `低` 的 TODO `整理文档`
- **AND** `未执行` 列表包含优先级为 `高` 的 TODO `修复登录问题`
- **AND** `未执行` 列表包含优先级为 `中` 的 TODO `升级依赖`
- **THEN** TODO tab 的 `未执行` 列表依次显示 `修复登录问题`、`升级依赖`、`整理文档`

#### Scenario: In-progress todos with same priority are ordered by creation time

- **WHEN** `执行中` 列表包含优先级同为 `高` 的 TODO `修复登录问题` 和 `排查线上报警`
- **AND** TODO `修复登录问题` 的 `createdAt` 早于 TODO `排查线上报警`
- **THEN** TODO tab 的 `执行中` 列表中 `修复登录问题` 排在 `排查线上报警` 前面

#### Scenario: User switches open todos to creation time order

- **WHEN** 当前状态视图包含创建时间更晚且优先级为 `高` 的 TODO `修复登录问题`
- **AND** 当前状态视图包含创建时间更早且优先级为 `低` 的 TODO `整理文档`
- **AND** 用户选择时间排序
- **THEN** 当前状态视图中 `整理文档` 排在 `修复登录问题` 前面

#### Scenario: Completed todo order is unaffected

- **WHEN** 用户查看 `已完成` 列表
- **THEN** 系统不按 `未执行` 或 `执行中` 的优先级排序规则重排 `已完成` 列表

### Requirement: Expand Or Collapse All Todo Branches

系统 SHALL 允许用户在 `未执行` 和 `执行中` 列表中一键展开或收起当前状态视图里的所有 TODO 分支。批量收起 SHALL 隐藏当前状态视图中所有 TODO 下的项目和终端子项，但 SHALL 保留 TODO 行可见。批量展开 SHALL 显示当前状态视图中所有 TODO 下的项目子项。

#### Scenario: User collapses all visible todo branches

- **WHEN** 当前 `未执行` 或 `执行中` 列表中存在多个已展开 TODO
- **AND** 用户激活全部收起控件
- **THEN** 当前状态视图中的所有 TODO 下的项目和终端子项被隐藏
- **AND** 当前状态视图中的所有 TODO 行仍显示

#### Scenario: User expands all visible todo branches

- **WHEN** 当前 `未执行` 或 `执行中` 列表中存在多个已收起 TODO
- **AND** 用户激活全部展开控件
- **THEN** 当前状态视图中的所有 TODO 下的项目子项被显示

#### Scenario: Active context expands after bulk collapse

- **WHEN** 用户已批量收起当前状态视图中的所有 TODO 分支
- **AND** 当前 TODO、TODO 项目或终端上下文切换到某个已收起 TODO 下
- **THEN** 系统自动展开该 TODO 分支
- **AND** 其他已收起 TODO 分支保持收起

### Requirement: Display Todo Workflow Status Views

系统 SHALL 在 TODO tab 中提供 `未执行`、`执行中`、`已完成` 三个状态视图。`未执行` 视图 SHALL 只显示状态为 `not-started` 的 TODO，`执行中` 视图 SHALL 只显示状态为 `in-progress` 的 TODO，`已完成` 视图 SHALL 只显示状态为 `completed` 的 TODO。

#### Scenario: User views not-started todos

- **WHEN** TODO `整理文档` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户打开 `未执行` 视图
- **THEN** TODO tab 显示 `整理文档`
- **AND** TODO tab 不显示 `修复登录问题`

#### Scenario: User views in-progress todos

- **WHEN** TODO `整理文档` 的状态为 `not-started`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户打开 `执行中` 视图
- **THEN** TODO tab 显示 `修复登录问题`
- **AND** TODO tab 不显示 `整理文档`

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户打开 `已完成` 视图
- **THEN** TODO tab 显示 `修复登录问题`

#### Scenario: Deleted todos are hidden from workflow views

- **WHEN** 用户删除 TODO `修复登录问题`
- **THEN** TODO `修复登录问题` 不显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

### Requirement: Manually Change Todo Workflow Status

系统 SHALL 允许用户手动在 `未执行` 和 `执行中` 状态之间切换 TODO。系统 SHALL NOT 根据终端创建、终端删除或终端活动状态自动改变 TODO 的工作流状态。

#### Scenario: User marks a not-started todo as in-progress

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **THEN** TODO `修复登录问题` 的状态保存为 `in-progress`
- **AND** TODO `修复登录问题` 显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `未执行` 视图

#### Scenario: User moves an in-progress todo back to not-started

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户将 TODO `修复登录问题` 标记为未执行
- **THEN** TODO `修复登录问题` 的状态保存为 `not-started`
- **AND** TODO `修复登录问题` 显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图

#### Scenario: Terminal activity does not change todo workflow status

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 TODO `修复登录问题` 下创建终端
- **THEN** TODO `修复登录问题` 的状态仍为 `not-started`

### Requirement: Restrict Todo Workflow Actions

系统 SHALL 按 TODO 当前工作流状态限制用户可执行的状态动作。`not-started` TODO SHALL 在行外仅暴露开始状态按钮，SHALL NOT 暴露完成入口或退回入口。`in-progress` TODO SHALL 在行外仅暴露完成状态按钮，SHALL NOT 暴露开始入口或退回 `not-started` 的入口。查看/编辑 TODO、添加项目和删除 TODO SHALL 在 `not-started` 与 `in-progress` 状态下通过 TODO 右键菜单保持可用。

#### Scenario: Not-started todo exposes only start as row action

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户在 `未执行` 视图查看该 TODO
- **THEN** 系统在 TODO 行外显示开始 TODO 的入口
- **AND** 系统不在 TODO 行外显示完成 TODO 的入口
- **AND** 系统不在 TODO 行外显示退回未执行的入口
- **AND** 系统不在 TODO 行外显示查看详情、添加项目或删除 TODO 入口
- **AND** 系统通过 TODO 右键菜单提供查看详情、添加项目和删除 TODO 入口

#### Scenario: In-progress todo exposes only complete as row action

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户在 `执行中` 视图查看该 TODO
- **THEN** 系统在 TODO 行外显示完成 TODO 的入口
- **AND** 系统不在 TODO 行外显示开始 TODO 的入口
- **AND** 系统不在 TODO 行外显示退回未执行的入口
- **AND** 系统不在 TODO 行外显示查看详情、添加项目或删除 TODO 入口
- **AND** 系统通过 TODO 右键菜单提供查看详情、添加项目和删除 TODO 入口

#### Scenario: User starts a not-started todo

- **WHEN** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **THEN** TODO `修复登录问题` 的状态保存为 `in-progress`
- **AND** TODO `修复登录问题` 显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `未执行` 视图

#### Scenario: In-progress todo cannot move back to not-started

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 用户或客户端请求将 TODO `修复登录问题` 标记为未执行
- **THEN** 系统拒绝该状态变更
- **AND** TODO `修复登录问题` 的状态保持为 `in-progress`

### Requirement: Manage Todo Dialog Dismissal

系统 SHALL 防止 TODO 创建、TODO 详情编辑和 TODO 添加项目弹窗因点击空白遮罩而关闭。这些弹窗 SHALL 仅通过弹窗关闭按钮、取消按钮或成功提交后的流程关闭。

#### Scenario: Create todo dialog ignores backdrop click

- **WHEN** 用户打开创建 TODO 弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** 创建 TODO 弹窗保持打开
- **AND** 用户已输入的 TODO 表单内容保持不变

#### Scenario: Todo detail dialog ignores backdrop click

- **WHEN** 用户打开 TODO `修复登录问题` 的详情编辑弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** TODO 详情编辑弹窗保持打开
- **AND** 用户已输入的详情表单内容保持不变

#### Scenario: Add project dialog ignores backdrop click

- **WHEN** 用户为 TODO `修复登录问题` 打开添加项目弹窗
- **AND** 用户点击弹窗外部空白遮罩
- **THEN** 添加项目弹窗保持打开
- **AND** 用户已选择的项目保持选中

#### Scenario: Todo dialog closes from explicit controls

- **WHEN** 用户打开创建 TODO、TODO 详情编辑或添加项目弹窗
- **AND** 用户点击弹窗关闭按钮或取消按钮
- **THEN** 系统关闭该弹窗

### Requirement: Use Todo Context Menu

系统 SHALL 在 `not-started` 和 `in-progress` TODO 行上提供右键菜单，并 SHALL 在 TODO item 上提供一个三点图标按钮用于打开同一组菜单动作。TODO 菜单 SHALL 包含查看详情、添加项目、复制标题和描述、删除 TODO 动作。菜单 SHALL 显示在触发它的指针或三点按钮附近，且 SHALL NOT 显示在应用窗口全局右上角。菜单动作完成、用户点击菜单外部或打开其他侧边栏浮层后，系统 SHALL 关闭 TODO 菜单。

#### Scenario: User opens todo context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开右键菜单
- **THEN** 系统在指针附近显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: User opens todo menu from item action button

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户点击 TODO `修复登录问题` item 上的三点图标按钮
- **THEN** 系统在三点图标按钮附近显示 TODO 菜单
- **AND** 菜单包含查看详情入口
- **AND** 菜单包含添加项目入口
- **AND** 菜单包含复制标题和描述入口
- **AND** 菜单包含删除 TODO 入口

#### Scenario: Windows todo menu does not appear in global corner

- **WHEN** 应用运行在 Windows
- **AND** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户通过右键或三点图标按钮打开 TODO 菜单
- **THEN** TODO 菜单显示在触发位置附近
- **AND** TODO 菜单不显示在应用窗口全局右上角

#### Scenario: User opens todo detail from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择查看详情
- **THEN** 系统打开 TODO `修复登录问题` 的详情编辑界面
- **AND** 系统关闭 TODO 菜单

#### Scenario: User opens add project dialog from context menu

- **WHEN** 活动 TODO 列表显示 TODO `修复登录问题`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择添加项目
- **THEN** 系统打开 TODO `修复登录问题` 的添加项目弹窗
- **AND** 系统关闭 TODO 菜单

#### Scenario: User copies todo title and description from context menu

- **WHEN** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择复制标题和描述
- **THEN** 系统将 `修复登录问题` 和 `登录后跳回首页` 写入系统剪贴板
- **AND** 剪贴板内容第一行为 `修复登录问题`
- **AND** 剪贴板内容第二行开始为 `登录后跳回首页`
- **AND** 系统关闭 TODO 菜单

#### Scenario: Empty todo description copies title only

- **WHEN** TODO `修复登录问题` 的描述为空
- **AND** 用户在 TODO `修复登录问题` 行上打开 TODO 菜单
- **AND** 用户选择复制标题和描述
- **THEN** 系统将 `修复登录问题` 写入系统剪贴板
- **AND** 剪贴板内容不包含额外空白描述行

#### Scenario: Outside click closes todo context menu

- **WHEN** TODO `修复登录问题` 的 TODO 菜单已显示
- **AND** 用户点击 TODO 菜单外部
- **THEN** 系统关闭 TODO 菜单
- **AND** 系统不执行菜单动作

### Requirement: View Completed Todo Details

系统 SHALL 允许用户从 `已完成` 视图打开 completed TODO 的详情。completed TODO 详情 SHALL 复用 TODO 详情弹窗，并 SHALL 以只读模式显示 TODO 标题、描述、优先级和完成时项目快照。只读详情 SHALL 隐藏保存按钮，SHALL 禁止编辑 TODO 字段或项目，且 SHALL NOT 恢复终端、启动 shell 进程或重新建立项目关联。

#### Scenario: User opens completed todo details

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** TODO `修复登录问题` 的描述为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级为 `高`
- **AND** TODO `修复登录问题` 的完成时项目快照包含名称为 `frontend-app`、路径为 `/work/frontend-app` 的项目
- **AND** 用户在 `已完成` 视图中打开 TODO `修复登录问题` 的详情
- **THEN** 系统打开 TODO 详情弹窗
- **AND** 详情弹窗显示标题 `修复登录问题`
- **AND** 详情弹窗显示描述 `登录后跳回首页`
- **AND** 详情弹窗显示优先级 `高`
- **AND** 详情弹窗显示项目快照 `frontend-app`
- **AND** 详情弹窗显示项目快照路径 `/work/frontend-app`

#### Scenario: Completed todo details are read-only

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统隐藏详情弹窗的保存按钮
- **AND** 系统不允许编辑 TODO 标题
- **AND** 系统不允许编辑 TODO 描述
- **AND** 系统不允许修改 TODO 优先级
- **AND** 系统不允许新增或移除项目

#### Scenario: Completed todo details do not restore runtime context

- **WHEN** 用户打开 completed TODO `修复登录问题` 的详情
- **THEN** 系统不重新创建该 TODO 的终端
- **AND** 系统不启动任何 shell 进程
- **AND** 系统不把完成时项目快照恢复为活动 TODO 项目关联

### Requirement: Bulk Delete Completed Todos

系统 SHALL 在 `已完成` 视图中提供 completed TODO 的多选批量删除能力。批量删除 SHALL 仅适用于 `已完成` 视图中的 completed TODO，SHALL 在删除前要求用户确认，确认后 SHALL 将选中的 completed TODO 从可见 TODO 工作区列表中移除。系统 SHALL NOT 在 `未执行` 或 `执行中` 视图中提供 TODO 批量删除能力。

#### Scenario: User bulk deletes completed todos

- **WHEN** `已完成` 视图显示 completed TODO `修复登录问题`
- **AND** `已完成` 视图显示 completed TODO `整理文档`
- **AND** 用户选择 TODO `修复登录问题`
- **AND** 用户选择 TODO `整理文档`
- **AND** 用户触发批量删除
- **AND** 系统显示批量删除确认
- **AND** 用户确认批量删除
- **THEN** `已完成` 视图不再显示 TODO `修复登录问题`
- **AND** `已完成` 视图不再显示 TODO `整理文档`

#### Scenario: User cancels completed todo bulk deletion

- **WHEN** `已完成` 视图显示 completed TODO `修复登录问题`
- **AND** 用户选择 TODO `修复登录问题`
- **AND** 用户触发批量删除
- **AND** 系统显示批量删除确认
- **AND** 用户取消批量删除
- **THEN** `已完成` 视图仍显示 TODO `修复登录问题`

#### Scenario: Open todo views do not expose bulk delete

- **WHEN** 用户在 TODO tab 中打开 `未执行` 视图
- **THEN** 系统不显示 TODO 批量删除入口
- **WHEN** 用户在 TODO tab 中打开 `执行中` 视图
- **THEN** 系统不显示 TODO 批量删除入口

#### Scenario: Bulk delete rejects non-completed todos

- **WHEN** 用户或客户端请求批量删除 TODO `修复登录问题`
- **AND** TODO `修复登录问题` 的状态为 `in-progress`
- **THEN** 系统拒绝批量删除请求
- **AND** TODO `修复登录问题` 仍显示在 `执行中` 视图

