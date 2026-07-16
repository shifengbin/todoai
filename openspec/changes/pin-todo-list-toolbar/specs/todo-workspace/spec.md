## MODIFIED Requirements

### Requirement: Scroll Todo Workspace List

系统 SHALL 在 TODO 列表内容超过侧边栏可见高度时，让当前状态视图的 TODO 列表在侧边栏内部独立滚动。滚动 SHALL 不改变终端区域高度或宽度。工作流状态 tab 行（`未执行`、`执行中`、`已完成` 三个按钮）和当前状态视图的共享顶部工具栏 SHALL 位于 TODO 列表滚动区域之外；用户滚动 TODO 列表时，工作流状态 tab 行和共享顶部工具栏 SHALL 保持可见且可操作。共享顶部工具栏 SHALL 在 `未执行` 和 `执行中` 视图显示排序与全部展开/收起控件，并 SHALL 在 `已完成` 视图显示已选数量与批量删除控件。

#### Scenario: Long todo list scrolls inside sidebar

- **WHEN** 当前状态视图的 TODO 列表内容超过侧边栏高度
- **THEN** 当前状态视图的 TODO 列表区域可以在侧边栏内部滚动
- **AND** 侧边栏 header、工作流状态 tab 和共享顶部工具栏保持可见
- **AND** 终端区域尺寸不因 TODO 列表长度被挤压

#### Scenario: Workflow tabs remain visible while todo list scrolls

- **WHEN** TODO 列表内容超过侧边栏高度
- **AND** 用户向下滚动 TODO 列表
- **THEN** `未执行`、`执行中`、`已完成` 三个工作流状态 tab 保持在 TODO 列表滚动区域之外
- **AND** 三个工作流状态 tab 保持可见且可点击

#### Scenario: Open todo toolbar remains visible while todo list scrolls

- **WHEN** 用户查看 `未执行` 或 `执行中` 视图
- **AND** 当前视图的 TODO 列表内容超过侧边栏高度
- **AND** 用户向下滚动 TODO 列表
- **THEN** 排序切换与全部展开/收起控件保持在 TODO 列表滚动区域之外
- **AND** 排序切换与全部展开/收起控件保持可见且可操作

#### Scenario: Completed todo toolbar remains visible while todo list scrolls

- **WHEN** 用户查看 `已完成` 视图
- **AND** 已完成 TODO 列表内容超过侧边栏高度
- **AND** 用户向下滚动已完成 TODO 列表
- **THEN** 已选数量与批量删除控件保持在 TODO 列表滚动区域之外
- **AND** 已选数量与批量删除控件保持可见且可操作
