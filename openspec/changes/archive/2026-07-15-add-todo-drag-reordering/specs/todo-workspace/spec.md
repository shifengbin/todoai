## ADDED Requirements

### Requirement: Persist Manual Todo Order

系统 SHALL 在工作区级 UI 状态中分别持久化 `未执行` 和 `执行中` TODO 的手动顺序，并 SHALL 同时持久化最后选择的开放 TODO 排序模式。系统 SHALL 在首次进入手动模式且尚无已保存顺序时，以切换前的显示顺序初始化两个开放状态列表。系统 SHALL 在加载顺序时丢弃重复、未知或状态不匹配的 TODO ID，并 SHALL 将未记录但有效的 TODO 追加到所属状态列表末尾。

#### Scenario: First manual sort preserves the current visible order

- **WHEN** 工作区尚未保存手动顺序
- **AND** 用户当前按优先级或创建时间查看开放 TODO
- **AND** 用户首次选择手动排序
- **THEN** 系统以切换前的显示顺序初始化 `未执行` 和 `执行中` 的手动顺序
- **AND** 列表不会因切换到手动排序而立即跳动

#### Scenario: Manual sort mode and order survive reopening the workspace

- **WHEN** 用户在工作区中选择手动排序并调整了 TODO 顺序
- **AND** 用户关闭后重新打开该工作区
- **THEN** 系统恢复手动排序模式
- **AND** `未执行` 和 `执行中` 列表分别恢复各自保存的顺序

#### Scenario: Workspaces keep independent manual orders

- **WHEN** 用户在工作区 A 和工作区 B 中分别保存不同的手动顺序
- **AND** 用户在两个工作区之间切换
- **THEN** 每个工作区恢复自己的排序模式和手动顺序

#### Scenario: New todo is appended to the not-started manual order

- **WHEN** 工作区已经保存 `未执行` 列表的手动顺序
- **AND** 用户创建一个新 TODO
- **THEN** 新 TODO 追加到 `未执行` 手动列表末尾
- **AND** 现有 TODO 的相对顺序保持不变

#### Scenario: Status change appends the todo to the target manual order

- **WHEN** TODO 从 `未执行` 切换为 `执行中`
- **THEN** 系统从 `未执行` 手动顺序中移除该 TODO
- **AND** 系统将该 TODO 追加到 `执行中` 手动列表末尾
- **AND** 两个列表中其他 TODO 的相对顺序保持不变

#### Scenario: Invalid persisted IDs are normalized

- **WHEN** 已保存顺序包含重复、已删除或状态不匹配的 TODO ID
- **AND** 当前状态还包含未记录的有效 TODO
- **THEN** 系统按首次出现顺序保留有效 ID
- **AND** 系统丢弃无效 ID
- **AND** 系统将未记录的有效 TODO 追加到所属列表末尾

#### Scenario: Legacy UI state remains compatible

- **WHEN** 工作区 UI 状态文件不包含排序模式和手动顺序字段
- **THEN** 系统成功加载该文件
- **AND** 开放 TODO 默认使用优先级排序
- **AND** 用户首次进入手动模式时按当前显示顺序初始化手动顺序

### Requirement: Drag To Reorder Open Todos

系统 SHALL 仅在手动排序模式下允许用户通过 TODO 行的专用拖拽手柄重排当前 `未执行` 或 `执行中` 列表。一次拖拽 SHALL NOT 跨越状态列表。拖动开始时，系统 SHALL 临时收起当前状态视图中所有可见 TODO 的子树而不改变其既有展开状态；拖动结束或取消后，系统 SHALL 恢复每个 TODO 在拖动前的展开状态。系统 SHALL 在长列表拖动接近滚动区域边缘时自动滚动。

#### Scenario: Drag handles appear only in manual open-todo views

- **WHEN** 用户在 `未执行` 或 `执行中` 视图选择手动排序
- **THEN** 每个可重排 TODO 行显示专用拖拽手柄
- **WHEN** 用户选择优先级排序或时间排序
- **THEN** TODO 行不显示拖拽手柄
- **WHEN** 用户查看 `已完成` 视图
- **THEN** 已完成 TODO 不显示拖拽手柄

#### Scenario: Dragging reorders only the current status list

- **WHEN** 用户通过手柄将当前列表中的 TODO `整理文档` 拖到 TODO `修复登录问题` 前面
- **THEN** 当前状态列表立即按新顺序显示
- **AND** 另一个开放状态列表的顺序保持不变
- **AND** 系统不允许将 TODO 拖入另一个状态列表

#### Scenario: Todos collapse visually during dragging and restore after drop

- **WHEN** 当前状态视图中存在多个展开和收起状态不同的 TODO
- **AND** 用户开始拖动其中一个 TODO
- **THEN** 系统临时收起当前状态视图中所有可见 TODO 的子树
- **WHEN** 用户完成拖放
- **THEN** 系统恢复每个 TODO 在拖动前的展开或收起状态
- **AND** 系统保存新的手动顺序

#### Scenario: Cancelling a drag restores expansion without changing order

- **WHEN** 用户开始拖动 TODO 后取消拖动或拖动失焦
- **THEN** 系统恢复拖动前的 TODO 顺序
- **AND** 系统恢复每个 TODO 在拖动前的展开或收起状态
- **AND** 系统不保存新的手动顺序

#### Scenario: Dragging near the list edge scrolls the list

- **WHEN** 手动排序列表内容超过当前可视高度
- **AND** 用户拖动 TODO 接近列表顶部或底部边缘
- **THEN** 列表沿对应方向自动滚动
- **AND** 用户能够把 TODO 放置到初始可视区域之外的位置

#### Scenario: Saving a reordered list fails

- **WHEN** 用户完成拖放并触发手动顺序保存
- **AND** 后端保存失败
- **THEN** 系统回滚到拖动前的 TODO 顺序
- **AND** 系统恢复每个 TODO 在拖动前的展开或收起状态
- **AND** 系统显示保存失败错误

#### Scenario: Reordering controls are stable while a save is pending

- **WHEN** 新的手动顺序正在保存
- **THEN** 系统阻止再次拖拽
- **AND** 系统阻止切换排序模式、批量展开或收起以及触发 TODO 行操作

## MODIFIED Requirements

### Requirement: Sort Active Todos

系统 SHALL 在 TODO tab 的 `未执行` 和 `执行中` 列表中提供排序切换控件，并 SHALL 支持按任务优先级排序、按创建时间排序和手动排序。没有有效已保存排序模式的工作区 SHALL 默认选择优先级排序；具有有效已保存排序模式的工作区 SHALL 恢复该模式。优先级排序 SHALL 为 `高`、`中`、`低`，相同优先级的 TODO SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。时间排序 SHALL 按 `createdAt` 创建时间正序展示，先创建的 TODO 排在前面。手动排序 SHALL 分别使用 `未执行` 和 `执行中` 列表保存的有序 TODO ID。`已完成` 列表 SHALL 不受 `未执行` 和 `执行中` 的排序规则影响。

#### Scenario: Open todo sort control defaults to priority

- **WHEN** 用户打开尚未保存有效排序模式的工作区
- **AND** 用户查看 TODO tab 的 `未执行` 或 `执行中` 列表
- **THEN** 系统显示 TODO 排序切换控件
- **AND** 排序切换控件默认选中优先级排序

#### Scenario: Saved open todo sort mode is restored

- **WHEN** 工作区保存的开放 TODO 排序模式为手动排序
- **AND** 用户重新打开该工作区
- **THEN** 排序切换控件选中手动排序
- **AND** 开放 TODO 按已保存的手动顺序显示

#### Scenario: Not-started todos are ordered by priority

- **WHEN** `未执行` 列表包含优先级为 `低` 的 TODO `整理文档`
- **AND** `未执行` 列表包含优先级为 `高` 的 TODO `修复登录问题`
- **AND** `未执行` 列表包含优先级为 `中` 的 TODO `升级依赖`
- **THEN** TODO tab 的 `未执行` 列表依次显示 `修复登录问题`、`升级依赖`、`整理文档`

#### Scenario: In-progress todos with same priority are ordered by creation time

- **WHEN** `执行中` 列表包含优先级同为 `高` 的 TODO `修复登录问题` 和 `排查线上报警`
- **AND** TODO `修复登录问题` 的 `createdAt` 早于 TODO `排查线上报警`
- **THEN** TODO tab 的 `执行中` 列表中 `修复登录问题` 排在 TODO `排查线上报警` 前面

#### Scenario: User switches open todos to creation time order

- **WHEN** 当前状态视图包含创建时间更晚且优先级为 `高` 的 TODO `修复登录问题`
- **AND** 当前状态视图包含创建时间更早且优先级为 `低` 的 TODO `整理文档`
- **AND** 用户选择时间排序
- **THEN** 当前状态视图中 `整理文档` 排在 `修复登录问题` 前面

#### Scenario: User switches open todos to manual order

- **WHEN** 当前状态视图已经保存 TODO `整理文档` 排在 TODO `修复登录问题` 前面的手动顺序
- **AND** 用户选择手动排序
- **THEN** 当前状态视图中 `整理文档` 排在 `修复登录问题` 前面

#### Scenario: Completed todo order is unaffected

- **WHEN** 用户查看 `已完成` 列表
- **THEN** 系统不按 `未执行` 或 `执行中` 的优先级、创建时间或手动排序规则重排 `已完成` 列表
