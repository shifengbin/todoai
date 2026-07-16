## MODIFIED Requirements

### Requirement: Display Todo Workflow Status Views

系统 SHALL 在 TODO tab 中提供 `未执行`、`执行中`、`已完成` 三个状态视图。`未执行` 视图 SHALL 只显示状态为 `not-started` 的真实 TODO，`执行中` 视图 SHALL 显示状态为 `in-progress` 的真实 TODO，并 SHALL 在存在 workspace 全局终端时额外显示 `Global 终端` 虚拟 TODO；`已完成` 视图 SHALL 只显示状态为 `completed` 的真实 TODO。`Global 终端` 虚拟 TODO SHALL NOT 被视为真实 TODO 工作流数据。

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

#### Scenario: User views workspace global terminals in progress

- **WHEN** 当前 workspace 存在至少一个全局终端
- **AND** 用户打开 `执行中` 视图
- **THEN** TODO tab 在真实 `in-progress` TODO 之前显示 `Global 终端` 虚拟 TODO
- **AND** 用户切换到 `未执行` 或 `已完成` 视图后不显示该虚拟 TODO

#### Scenario: User views completed todos

- **WHEN** TODO `修复登录问题` 的状态为 `completed`
- **AND** 用户打开 `已完成` 视图
- **THEN** TODO tab 显示 `修复登录问题`

#### Scenario: Deleted todos are hidden from workflow views

- **WHEN** 用户删除 TODO `修复登录问题`
- **THEN** TODO `修复登录问题` 不显示在 `未执行` 视图
- **AND** TODO `修复登录问题` 不显示在 `执行中` 视图
- **AND** TODO `修复登录问题` 不显示在 `已完成` 视图

## ADDED Requirements

### Requirement: Keep Workspace Global Terminal Virtual Todo Outside Real Todo Ordering

系统 SHALL 将 `Global 终端` 虚拟 TODO 固定在 `执行中` 视图首位。该虚拟 TODO SHALL NOT 进入优先级排序、创建时间排序、手动排序、手动顺序持久化或真实 TODO 拖拽重排。该虚拟 TODO SHALL NOT 暴露优先级、开始、完成、编辑、删除或拖拽手柄；真实 TODO 的排序和保存数据 SHALL 仅包含真实 TODO ID。

#### Scenario: Virtual todo remains first across sort modes

- **WHEN** 当前 workspace 存在全局终端
- **AND** `执行中` 视图包含多个真实 TODO
- **WHEN** 用户在优先级、创建时间和手动排序模式之间切换
- **THEN** `Global 终端` 虚拟 TODO 始终显示在第一项
- **AND** 真实 TODO 按当前排序模式在其后排序

#### Scenario: Virtual todo is excluded from manual order persistence

- **WHEN** 用户在 `执行中` 视图保存真实 TODO 的手动顺序
- **THEN** 保存的有序 ID 仅包含真实 `in-progress` TODO ID
- **AND** 保存数据不包含 `Global 终端` 虚拟 TODO 的标识

#### Scenario: Virtual todo cannot be dragged

- **WHEN** 用户在 `执行中` 视图启用手动排序
- **THEN** `Global 终端` 虚拟 TODO 不显示拖拽手柄
- **AND** 真实 TODO 不能被拖放到虚拟 TODO 之前

### Requirement: Control Workspace Global Terminal Virtual Todo Branch

系统 SHALL 允许用户通过展开控件或双击 header 行展开和收起 `Global 终端` 虚拟 TODO。展开后系统 SHALL 显示全局终端子项及新增入口；收起后系统 SHALL 隐藏子项，并 SHALL 按 `needs-input` 高于 `needs-ack` 高于 `busy` 高于 `idle` 的优先级聚合全部全局终端活动状态。`执行中` 视图的批量展开、批量收起以及拖拽期间的临时收起 SHALL 包含该虚拟 TODO，且 SHALL 在拖拽结束或取消后恢复其原展开状态。

#### Scenario: User expands and collapses the virtual todo

- **WHEN** 当前 workspace 存在多个全局终端
- **AND** 用户展开 `Global 终端` 虚拟 TODO
- **THEN** 系统显示全部全局终端子项和新增入口
- **WHEN** 用户收起该虚拟 TODO
- **THEN** 系统隐藏全部全局终端子项和新增入口

#### Scenario: Collapsed virtual todo aggregates terminal attention

- **WHEN** `Global 终端` 虚拟 TODO 已收起
- **AND** 一个全局终端的活动状态为 `busy`
- **AND** 另一个全局终端的活动状态为 `needs-ack`
- **AND** 另一个全局终端的活动状态为 `needs-input`
- **THEN** 虚拟 TODO 行显示 `needs-input` 的整行活动反馈

#### Scenario: Bulk collapse includes the virtual todo

- **WHEN** `Global 终端` 虚拟 TODO 和一个或多个真实 TODO 已展开
- **AND** 用户在 `执行中` 视图触发批量收起
- **THEN** 虚拟 TODO 与所有真实 TODO 的子树均被收起
- **WHEN** 用户触发批量展开
- **THEN** 虚拟 TODO 与所有真实 TODO 的可用子树均被展开

#### Scenario: Todo dragging temporarily collapses the virtual todo

- **WHEN** `Global 终端` 虚拟 TODO 已展开
- **AND** 用户开始拖动一个真实 `in-progress` TODO
- **THEN** 系统临时收起虚拟 TODO 和当前视图中的真实 TODO 子树
- **WHEN** 用户完成或取消拖动
- **THEN** 系统恢复虚拟 TODO 与真实 TODO 拖动前的展开状态

### Requirement: Highlight Workspace Global Terminal Virtual Context

系统 SHALL 在当前活动终端为 workspace 全局终端时高亮该终端子项及 `Global 终端` 虚拟 TODO 父项。该虚拟选中状态 SHALL 由活动终端推导，SHALL NOT 写入当前真实 TODO ID、TODO project ID、TODO 排序数据或 TODO UI 持久化状态。

#### Scenario: Active global terminal highlights virtual parent and child

- **WHEN** `Global 终端` 虚拟 TODO 下存在全局终端 A
- **AND** 全局终端 A 成为当前活动终端
- **THEN** 全局终端 A 显示终端选中状态
- **AND** `Global 终端` 虚拟 TODO 显示父级选中状态
- **AND** 当前真实 TODO 和 TODO project ID 保持不变

#### Scenario: Selecting a real todo terminal clears virtual highlight

- **WHEN** 当前活动终端为全局终端 A
- **AND** 用户选择真实 TODO 下的终端 B
- **THEN** 终端 B 显示终端选中状态
- **AND** `Global 终端` 虚拟 TODO 不再显示父级选中状态
