## ADDED Requirements

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
