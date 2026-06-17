## MODIFIED Requirements

### Requirement: Display Workspace Tabs

系统 SHALL 在左侧工作区直接显示 TODO 工作树。左侧工作区 SHALL NOT 提供独立的 `项目` tab。全局项目候选管理 SHALL 出现在创建 TODO、编辑 TODO 或添加工程弹窗中，不作为左侧独立工作区视图。

#### Scenario: User views the sidebar workspace

- **WHEN** 用户打开一个 workspace
- **THEN** 系统在左侧工作区显示 TODO 工作树
- **AND** 终端入口按 TODO 上下文展示
- **AND** 系统不显示 `项目` tab
- **AND** 系统不显示独立项目库管理视图

### Requirement: Create Todo

系统 SHALL 允许用户通过应用内表单创建 TODO。创建表单 SHALL 包含必填 TODO 名称、选填 TODO 描述、任务优先级和选填工程。工程 SHALL 从全局项目候选中多选。任务优先级 SHALL 支持 `高`、`中`、`低` 三档，并 SHALL 默认选择 `中`。创建后的 TODO SHALL 默认保存为 `not-started` 状态，出现在 `未执行` 视图中，并 SHALL 可作为后续项目关联和终端创建的上下文。创建时选择的工程 SHALL 在当前 workspace 中保存为 TODO 工程副本，包含添加时的名称、路径和来源候选 ID。创建后的 TODO 分支 SHALL 默认收起，且创建操作 SHALL 不改变当前 TODO、当前项目、当前 TODO 项目或当前终端上下文。

#### Scenario: User creates a todo without a project

- **WHEN** 用户在 TODO 工作区中创建名称为 `修复登录问题`、描述为空、优先级为 `中` 且未选择工程的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态为 `not-started`
- **AND** 该 TODO 的优先级为 `中`
- **AND** 该 TODO 初始不包含关联项目
- **AND** TODO `修复登录问题` 的分支默认收起

#### Scenario: User creates a todo with description and priority

- **WHEN** 用户在 TODO 工作区中创建名称为 `修复登录问题`、描述为 `登录后跳回首页`、优先级为 `高` 的 TODO
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** 该 TODO 保存描述 `登录后跳回首页`
- **AND** 该 TODO 的优先级为 `高`
- **AND** 该 TODO 的状态为 `not-started`

#### Scenario: User creates a todo with optional projects

- **WHEN** 全局项目候选包含 `frontend-app` 和 `api-service`
- **AND** 当前 TODO 项目上下文为 TODO `升级依赖` 下的 `docs-site`
- **AND** 用户在创建 TODO 表单中输入名称 `修复登录问题`
- **AND** 用户选择优先级 `高`
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** `未执行` 视图包含 `修复登录问题`
- **AND** TODO `修复登录问题` 下保存工程副本 `frontend-app`
- **AND** TODO `修复登录问题` 下保存工程副本 `api-service`
- **AND** 工程副本保存添加时的名称和路径
- **AND** 当前 TODO 项目上下文仍为 TODO `升级依赖` 下的 `docs-site`
- **AND** TODO `修复登录问题` 的分支默认收起

#### Scenario: Todo title is required

- **WHEN** 用户尝试创建名称为空的 TODO
- **THEN** 系统不创建 TODO
- **AND** 系统显示不会改变当前工作区状态的错误信息

#### Scenario: Project selection can be searched while creating todo

- **WHEN** 全局项目候选包含 `frontend-app` 和 `api-service`
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

系统 SHALL 在当前 workspace 数据目录中持久化 TODO、TODO 描述、TODO 优先级、TODO 工作流状态、TODO 工程副本、TODO 选中状态和已完成状态，并 SHALL 在该 workspace 重新打开后恢复。不同 workspace 的 TODO 数据 SHALL NOT 全局共享。TODO 工程副本 SHALL 保存添加时的项目名称、路径和来源候选 ID，且 SHALL NOT 依赖全局候选继续存在。旧数据中缺少优先级的 TODO SHALL 按 `中` 优先级处理。旧数据中状态为 `active` 的 TODO SHALL 按 `not-started` 处理。旧数据中状态为 `archived` 且归档原因为 `completed` 的 TODO SHALL 按 `completed` 处理。旧数据中状态为 `archived` 且归档原因为 `deleted` 的 TODO SHALL 不在 TODO 工作区列表中展示。

#### Scenario: Todo workspace is restored after reopening workspace

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户填写描述 `登录后跳回首页`
- **AND** 用户选择优先级 `高`
- **AND** 用户将 TODO `修复登录问题` 标记为执行中
- **AND** 用户将项目 `frontend-app` 关联到该 TODO
- **AND** 用户关闭并重新打开 workspace `/work/customer-a`
- **THEN** `执行中` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的描述仍为 `登录后跳回首页`
- **AND** TODO `修复登录问题` 的优先级仍为 `高`
- **AND** TODO `修复登录问题` 的状态仍为 `in-progress`
- **AND** `frontend-app` 仍保存为该 TODO 下的工程副本
- **AND** 该工程副本仍包含添加时保存的路径

#### Scenario: Todo workspace is isolated by workspace

- **WHEN** 用户打开 workspace `/work/customer-a`
- **AND** 用户创建 TODO `修复登录问题`
- **AND** 用户打开 workspace `/work/customer-b`
- **THEN** TODO 工作区不显示 `修复登录问题`

#### Scenario: Legacy todo project copies are populated

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 包含旧 `todoProject` 引用 `project-a`
- **AND** 旧项目库中 `project-a` 的名称为 `frontend-app`
- **AND** 旧项目库中 `project-a` 的路径为 `/repo/frontend-app`
- **AND** 用户打开该 workspace
- **THEN** TODO `修复登录问题` 下的工程副本名称为 `frontend-app`
- **AND** 该工程副本路径为 `/repo/frontend-app`
- **AND** 该工程副本来源候选 ID 指向迁移后的全局候选或旧项目 ID

#### Scenario: Legacy todo without priority uses medium

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 不包含优先级字段
- **AND** 用户打开该 workspace
- **THEN** TODO 工作区显示 `修复登录问题`
- **AND** TODO `修复登录问题` 按 `中` 优先级展示

#### Scenario: Legacy active todo becomes not-started

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `active`
- **AND** 用户打开该 workspace
- **THEN** `未执行` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `not-started` 处理

#### Scenario: Legacy completed archived todo remains completed

- **WHEN** 当前 workspace 持久化数据中 TODO `修复登录问题` 的状态为 `archived`
- **AND** TODO `修复登录问题` 的归档原因为 `completed`
- **AND** 用户打开该 workspace
- **THEN** `已完成` 视图显示 `修复登录问题`
- **AND** TODO `修复登录问题` 的状态按 `completed` 处理

#### Scenario: Legacy deleted archived todo is hidden

- **WHEN** 当前 workspace 持久化数据中 TODO `废弃任务` 的状态为 `archived`
- **AND** TODO `废弃任务` 的归档原因为 `deleted`
- **AND** 用户打开该 workspace
- **THEN** TODO 工作区不在 `未执行` 视图显示 `废弃任务`
- **AND** TODO 工作区不在 `执行中` 视图显示 `废弃任务`
- **AND** TODO 工作区不在 `已完成` 视图显示 `废弃任务`

### Requirement: Associate Projects With Todo

系统 SHALL 允许用户从全局项目候选中通过可搜索多选控件选择一个或多个项目关联到 TODO。关联时系统 SHALL 在当前 workspace 中创建 TODO 工程副本，并 SHALL 保存添加时的项目名称、路径和来源候选 ID。系统 SHALL 允许用户从 TODO 中移除已关联工程副本。同一路径 SHALL 可关联到多个不同 TODO。项目选择 SHALL 支持按项目名称和路径筛选，且 SHALL 不要求用户手动输入完整项目名称。移除 TODO 下的工程关联 SHALL 只影响当前 TODO 下的该工程副本。

#### Scenario: User associates projects with a todo

- **WHEN** 全局项目候选包含 `frontend-app` 和 `api-service`
- **AND** 用户为 TODO `修复登录问题` 选择这两个项目
- **THEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** TODO `修复登录问题` 下显示工程副本 `api-service`
- **AND** 两个工程副本均保存添加时的路径

#### Scenario: Same project is associated with multiple todos

- **WHEN** 全局项目候选包含 `frontend-app`
- **AND** 用户将 `frontend-app` 关联到 TODO `修复登录问题`
- **AND** 用户将 `frontend-app` 关联到 TODO `升级依赖`
- **THEN** `frontend-app` 同时显示在两个 TODO 下
- **AND** 两个 TODO 下的 `frontend-app` 工程副本互不替代

#### Scenario: Duplicate association is ignored

- **WHEN** TODO `修复登录问题` 已关联路径为 `/repo/frontend-app` 的工程副本
- **AND** 用户再次将路径为 `/repo/frontend-app` 的全局候选关联到该 TODO
- **THEN** TODO `修复登录问题` 下只显示一个路径为 `/repo/frontend-app` 的工程副本

#### Scenario: User filters projects while associating a todo

- **WHEN** 全局项目候选包含名称为 `frontend-app` 的项目
- **AND** 全局项目候选包含路径为 `/work/api-service` 的项目
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户在工程筛选框输入 `api`
- **THEN** 工程选择列表显示 `/work/api-service` 对应候选项目
- **AND** 工程选择列表不显示 `frontend-app`

#### Scenario: User associates multiple filtered projects with a todo

- **WHEN** 全局项目候选包含 `frontend-app`、`api-service` 和 `docs-site`
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户选择 `frontend-app`
- **AND** 用户选择 `api-service`
- **AND** 用户确认添加
- **THEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** TODO `修复登录问题` 下显示工程副本 `api-service`
- **AND** TODO `修复登录问题` 下不新增 `docs-site`

#### Scenario: Already linked project is excluded from selectable projects

- **WHEN** TODO `修复登录问题` 已关联路径为 `/repo/frontend-app` 的工程副本
- **AND** 全局项目候选还包含路径为 `/repo/api-service` 的 `api-service`
- **AND** 用户为 TODO `修复登录问题` 打开添加工程控件
- **THEN** 工程选择列表不显示路径为 `/repo/frontend-app` 的候选项目
- **AND** 工程选择列表显示 `api-service`

#### Scenario: Selected projects can be removed while associating a todo

- **WHEN** 用户为 TODO `修复登录问题` 打开添加工程控件
- **AND** 用户选择工程 `frontend-app`
- **AND** 用户选择工程 `api-service`
- **THEN** 添加工程控件在筛选框上方以 tag 展示 `frontend-app`
- **AND** 添加工程控件在筛选框上方以 tag 展示 `api-service`
- **WHEN** 用户删除 `api-service` tag
- **AND** 用户确认添加
- **THEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** TODO `修复登录问题` 下不新增 `api-service`

#### Scenario: User removes project from todo list with popover confirmation

- **WHEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** 用户点击 `frontend-app` 工程行上的删除按钮
- **THEN** 系统在删除按钮旁显示删除确认气泡
- **WHEN** 用户在确认气泡中确认删除
- **THEN** TODO `修复登录问题` 下不再显示工程副本 `frontend-app`

#### Scenario: User cancels project removal popover

- **WHEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** 用户点击 `frontend-app` 工程行上的删除按钮
- **AND** 系统显示删除确认气泡
- **WHEN** 用户取消删除
- **THEN** TODO `修复登录问题` 下仍显示工程副本 `frontend-app`

#### Scenario: Removing project from one todo preserves other todos

- **WHEN** 工程路径 `/repo/frontend-app` 同时关联到 TODO `修复登录问题` 和 TODO `升级依赖`
- **AND** 用户从 TODO `修复登录问题` 下移除工程副本 `frontend-app`
- **THEN** TODO `修复登录问题` 下不再显示工程副本 `frontend-app`
- **AND** TODO `升级依赖` 下仍显示工程副本 `frontend-app`

### Requirement: Select Todo Project Context

系统 SHALL 允许用户选择 TODO 下的工程副本作为当前工作上下文。选择 TODO 工程上下文 SHALL 更新当前 TODO、当前项目路径和当前 TODO 工程关联，但 SHALL 只使用该 TODO 工程副本下的终端集合。选择 TODO 工程 SHALL 使用工程副本保存的名称、路径和可用性，不要求来源全局候选仍然存在。

#### Scenario: User selects a project under a todo

- **WHEN** 用户在 TODO `修复登录问题` 下选择工程副本 `frontend-app`
- **THEN** 当前 TODO 为 `修复登录问题`
- **AND** 当前项目路径为该工程副本保存的路径
- **AND** 终端区域只关联该 TODO 工程上下文中的终端

#### Scenario: Same project selected under different todos

- **WHEN** 项目路径 `/repo/frontend-app` 同时关联到 TODO `修复登录问题` 和 TODO `升级依赖`
- **AND** 用户选择 TODO `修复登录问题` 下的 `frontend-app`
- **THEN** 终端区域显示 `修复登录问题` 下 `frontend-app` 的终端集合
- **WHEN** 用户选择 TODO `升级依赖` 下的 `frontend-app`
- **THEN** 终端区域显示 `升级依赖` 下 `frontend-app` 的终端集合

#### Scenario: Todo project remains selectable after global candidate is cleared

- **WHEN** TODO `修复登录问题` 下存在工程副本 `frontend-app`
- **AND** 用户清空全局项目候选库
- **AND** 用户选择 TODO `修复登录问题` 下的 `frontend-app`
- **THEN** 当前 TODO 为 `修复登录问题`
- **AND** 当前项目路径为该工程副本保存的路径
- **AND** 系统不因为来源候选缺失而报错

### Requirement: Preserve Archived Project Snapshots

系统 SHALL 在 TODO 归档时保存关联工程副本的名称、路径和来源项目 ID 快照。后续全局项目候选变化 SHALL 不改变已归档 TODO 的项目快照。

#### Scenario: Archived todo keeps project snapshot after project deletion

- **WHEN** TODO `修复登录问题` 关联工程副本 `frontend-app`
- **AND** 用户完成该 TODO
- **AND** 用户随后从全局候选库删除 `frontend-app`
- **THEN** 归档视图中 TODO `修复登录问题` 仍显示 `frontend-app` 的归档名称和路径
