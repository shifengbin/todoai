## ADDED Requirements

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

## MODIFIED Requirements

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
