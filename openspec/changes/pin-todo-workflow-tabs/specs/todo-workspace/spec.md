## MODIFIED Requirements

### Requirement: Scroll Todo Workspace List

系统 SHALL 在 TODO 列表内容超过侧边栏可见高度时，让 TODO tab 内容在侧边栏内部滚动。滚动 SHALL 不改变终端区域高度或宽度。工作流状态 tab 行（`未执行`、`执行中`、`已完成` 三个按钮）SHALL 位于 TODO 列表滚动区域之外；用户滚动 TODO 列表时，工作流状态 tab 行 SHALL 保持可见且可点击。

#### Scenario: Long todo list scrolls inside sidebar

- **WHEN** 活动 TODO 列表内容超过侧边栏高度
- **THEN** TODO tab 内部出现可滚动区域
- **AND** 侧边栏 header 和 tab 控件保持可见
- **AND** 终端区域尺寸不因 TODO 列表长度被挤压

#### Scenario: Workflow tabs remain visible while todo list scrolls

- **WHEN** TODO 列表内容超过侧边栏高度
- **AND** 用户向下滚动 TODO 列表
- **THEN** `未执行`、`执行中`、`已完成` 三个工作流状态 tab 保持在 TODO 列表滚动区域之外
- **AND** 三个工作流状态 tab 保持可见且可点击
