## ADDED Requirements

### Requirement: Display Todo Priority Visuals

系统 SHALL 在 TODO 工作树中展示活动 TODO 的优先级，并 SHALL 使用与优先级一致的颜色标记 TODO 行。优先级 SHALL 包含 `高`、`中`、`低` 三档，TODO 行背景 SHALL 使用同优先级一致的同色系背景。

#### Scenario: Todo row shows high priority styling

- **WHEN** 活动 TODO `修复登录问题` 的优先级为 `高`
- **THEN** TODO 行显示 `高` 优先级标记
- **AND** TODO 行背景使用高优先级对应的红色系背景

#### Scenario: Todo row shows medium priority styling

- **WHEN** 活动 TODO `升级依赖` 的优先级为 `中`
- **THEN** TODO 行显示 `中` 优先级标记
- **AND** TODO 行背景使用中优先级对应的橙色系背景

#### Scenario: Todo row shows low priority styling

- **WHEN** 活动 TODO `整理文档` 的优先级为 `低`
- **THEN** TODO 行显示 `低` 优先级标记
- **AND** TODO 行背景使用低优先级对应的绿色系背景

## MODIFIED Requirements

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

系统 SHALL 允许用户从项目库中通过可搜索多选控件选择一个或多个项目关联到 TODO。同一个项目 SHALL 可关联到多个不同 TODO。项目选择 SHALL 支持按项目名称和路径筛选，且 SHALL 不要求用户手动输入完整项目名称。

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
