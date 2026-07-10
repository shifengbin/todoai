## MODIFIED Requirements

### Requirement: Associate Projects With Todo

系统 SHALL 允许用户从全局项目候选中通过可搜索多选控件选择一个或多个项目关联到 TODO。关联时系统 SHALL 在当前 workspace 中创建 TODO 工程副本，并 SHALL 保存添加时的项目名称、路径和来源候选 ID。系统 SHALL 允许用户从 TODO 中移除已关联工程副本。同一路径 SHALL 可关联到多个不同 TODO。项目选择 SHALL 支持按项目名称和路径筛选，且 SHALL 不要求用户手动输入完整项目名称。移除 TODO 下的工程关联 SHALL 只影响当前 TODO 下的该工程副本。若 TODO 已经处于 `in-progress`，系统 SHALL 在项目关联保存后准备该 TODO 的任务工作区和新增项目的 Git worktree。

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

#### Scenario: Associating project with in-progress todo prepares worktree

- **WHEN** TODO `修复登录问题` 的状态为 `in-progress`
- **AND** 该 TODO 尚未关联任何项目
- **AND** 用户通过添加工程控件选择 Git 项目 `frontend-app`
- **THEN** TODO `修复登录问题` 下显示工程副本 `frontend-app`
- **AND** 系统为 `frontend-app` 准备 Git worktree
- **AND** 系统在任务工作区中生成包含 `frontend-app` 项目信息的 `README.md`

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

