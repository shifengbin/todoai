## MODIFIED Requirements

### Requirement: Create Workspace Global Terminals

系统 SHALL 允许用户在当前 workspace 中创建多个全局终端。全局终端 SHALL 不属于任何 TODO、TODO project 或全局项目候选。每个全局终端的 shell 进程 SHALL 以当前 workspace 根目录作为工作目录。成功创建全局终端后，系统 SHALL 激活该终端；若 TODO 侧栏当前不在 `执行中` 视图，系统 SHALL 切换到 `执行中`、展开 `Global 终端` 虚拟 TODO 并显示新终端。创建失败时，系统 MUST 保持原 TODO 视图、原终端选择和现有全局终端列表不变。

#### Scenario: User creates a workspace global terminal

- **WHEN** 用户打开 workspace `/home/user/work/customer-a`
- **AND** 用户创建一个全局终端
- **THEN** 系统创建一个新的终端会话
- **AND** 该终端不包含 TODO ID 或 TODO project ID
- **AND** 该终端的 shell 工作目录为 `/home/user/work/customer-a`
- **AND** 该终端成为当前活动终端

#### Scenario: User creates a workspace global terminal from another todo view

- **WHEN** 用户正在 TODO 侧栏的 `未执行` 或 `已完成` 视图
- **AND** 用户通过顶部 `Global terminal` 按钮成功创建全局终端
- **THEN** 系统切换到 `执行中` 视图
- **AND** 系统展开 `Global 终端` 虚拟 TODO
- **AND** 新创建的终端显示为选中的子终端

#### Scenario: User creates multiple workspace global terminals

- **WHEN** 用户打开 workspace `/home/user/work/customer-a`
- **AND** 用户创建全局终端 A
- **AND** 用户创建全局终端 B
- **THEN** 全局终端 A 和全局终端 B 是独立终端会话
- **AND** 两个终端都显示在 `Global 终端` 虚拟 TODO 下
- **AND** 两个终端都不显示在任何真实 TODO 或 TODO 项目终端列表中

#### Scenario: Workspace global terminal creation fails

- **WHEN** 用户正在 TODO 侧栏的 `已完成` 视图
- **AND** 用户请求创建全局终端
- **AND** 全局终端创建失败
- **THEN** TODO 侧栏仍显示 `已完成` 视图
- **AND** 当前活动终端和已有全局终端列表保持不变
- **AND** 系统显示非阻断错误信息

### Requirement: Display Workspace Global Terminal Group

系统 SHALL 在 TODO 侧栏的 `执行中` 视图中以名为 `Global 终端` 的单个虚拟 TODO 展示当前 workspace 的全局终端列表。该虚拟 TODO SHALL 在至少存在一个全局终端时固定显示在所有真实 `in-progress` TODO 之前，并 SHALL 提供新增全局终端入口。系统 SHALL NOT 在 `未执行` 或 `已完成` 视图中显示该虚拟 TODO。主终端区域 SHALL NOT 再渲染独立的全局终端标签组，也 SHALL NOT 为其保留布局高度。当当前 workspace 不存在全局终端时，系统 SHALL 隐藏虚拟 TODO，但 SHALL 保留顶部 `Global terminal` 创建按钮。

#### Scenario: Global terminal virtual todo appears when terminals exist

- **WHEN** 当前 workspace 存在至少一个全局终端
- **AND** 用户打开 TODO 侧栏的 `执行中` 视图
- **THEN** `Global 终端` 虚拟 TODO 显示在所有真实 TODO 之前
- **AND** 该虚拟 TODO 显示全局终端列表
- **AND** 用户可以从该虚拟 TODO 创建新的全局终端

#### Scenario: Global terminal virtual todo is hidden outside in-progress view

- **WHEN** 当前 workspace 存在至少一个全局终端
- **AND** 用户打开 TODO 侧栏的 `未执行` 或 `已完成` 视图
- **THEN** `Global 终端` 虚拟 TODO 不显示在当前列表中
- **AND** 全局终端会话继续保持原运行状态

#### Scenario: Global terminal virtual todo is hidden when empty

- **WHEN** 当前 workspace 不存在全局终端
- **THEN** TODO 侧栏不显示 `Global 终端` 虚拟 TODO
- **AND** 顶部 `Global terminal` 创建按钮保持可见且可用
- **AND** TODO 侧栏不为虚拟 TODO 保留空白高度

#### Scenario: Main terminal surface no longer renders global terminal tabs

- **WHEN** 当前 workspace 存在一个或多个全局终端
- **THEN** 主终端区域不显示独立的 Global terminal 标签组
- **AND** 主终端区域不为该标签组保留布局高度
- **AND** 每个全局终端对应的 xterm pane 仍可被激活和使用

### Requirement: Select Workspace Global Terminal

系统 SHALL 允许用户从 `Global 终端` 虚拟 TODO 中选择全局终端作为当前活动终端。选择全局终端 SHALL 激活对应 xterm pane，并 SHALL 高亮所选终端行及其虚拟父项，但 SHALL NOT 改变当前真实 TODO、当前 TODO project、当前项目候选或当前 Git 状态上下文。

#### Scenario: User selects global terminal without changing todo context

- **WHEN** 当前 TODO project 上下文为 TODO `修复登录问题` 下的 `frontend-app`
- **AND** 当前项目 Git 状态显示 `frontend-app`
- **AND** 用户选择一个全局终端
- **THEN** 该全局终端成为当前活动终端
- **AND** 该全局终端行和 `Global 终端` 虚拟 TODO 显示选中状态
- **AND** 当前 TODO project 上下文仍为 TODO `修复登录问题` 下的 `frontend-app`
- **AND** 当前项目 Git 状态上下文仍为 `frontend-app`

#### Scenario: User returns to todo project terminal

- **WHEN** 用户已选择全局终端
- **AND** TODO `修复登录问题` 下的 `frontend-app` 存在终端 A
- **AND** 用户选择终端 A
- **THEN** 终端 A 成为当前活动终端
- **AND** `Global 终端` 虚拟 TODO 不再显示选中状态
- **AND** 当前 TODO project 上下文更新为 TODO `修复登录问题` 下的 `frontend-app`

### Requirement: Manage Workspace Global Terminal Lifecycle

系统 SHALL 支持全局终端的输入、输出、resize、复制、粘贴、删除和 exited 后重启。`Global 终端` 虚拟 TODO SHALL 提供新增全局终端入口，每个全局终端子项 SHALL 提供删除入口。删除全局终端 SHALL 只关闭和移除该全局终端，不影响 TODO project 终端；删除最后一个全局终端后，系统 SHALL 隐藏虚拟 TODO。

#### Scenario: User deletes a global terminal

- **WHEN** 全局终端 A 正在运行
- **AND** TODO project 终端 B 正在运行
- **AND** 用户从 `Global 终端` 虚拟 TODO 删除全局终端 A
- **THEN** 系统关闭全局终端 A 的 shell 进程
- **AND** 全局终端 A 不再显示
- **AND** TODO project 终端 B 继续运行并保持显示

#### Scenario: User deletes the last global terminal

- **WHEN** `Global 终端` 虚拟 TODO 仅包含全局终端 A
- **AND** 用户成功删除全局终端 A
- **THEN** 系统隐藏 `Global 终端` 虚拟 TODO
- **AND** 顶部 `Global terminal` 创建按钮保持可用
- **AND** 系统不保留失效的虚拟 TODO 选中状态

#### Scenario: User restarts an exited global terminal

- **WHEN** 全局终端 A 的 shell 已退出
- **AND** 用户请求重启全局终端 A
- **THEN** 系统在当前 workspace 根目录重新启动全局终端 A 的 shell
- **AND** 全局终端 A 保持全局终端归属
- **AND** 全局终端 A 继续显示在 `Global 终端` 虚拟 TODO 下
