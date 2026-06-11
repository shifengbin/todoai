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

系统 SHALL 允许用户完成活动 TODO。完成 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为已完成状态。

#### Scenario: User completes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户确认完成该 TODO
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `completed`

### Requirement: Delete Todo

系统 SHALL 允许用户删除活动 TODO。删除 TODO SHALL 关闭并销毁该 TODO 下所有运行时终端，并 SHALL 将 TODO 归档为删除状态。

#### Scenario: User deletes a todo

- **WHEN** TODO `修复登录问题` 下存在运行中的终端
- **AND** 用户确认删除该 TODO
- **THEN** 系统关闭该 TODO 下所有运行中 shell 进程
- **AND** 系统从运行时状态移除该 TODO 下所有终端
- **AND** TODO `修复登录问题` 不再显示在活动任务列表中
- **AND** TODO `修复登录问题` 显示在归档列表中且归档原因为 `deleted`

#### Scenario: User cancels todo deletion

- **WHEN** 用户请求删除 TODO `修复登录问题`
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

