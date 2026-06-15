## Why

TODO 工作区从 `未执行` 或 `执行中` 切换到 `已完成` 时，顶部控制区高度会变窄，造成列表内容上移和界面抖动。该区域需要在三个 TODO 状态视图之间保持稳定，同时不能改变顶部状态切换栏的宽度分配。

## What Changes

- `已完成` 视图 SHALL 保留与 `未执行`、`执行中` 视图相同的 TODO 顶部控制区高度。
- TODO 顶部三段状态切换栏 SHALL 在切换 `未执行`、`执行中`、`已完成` 时保持相同宽度分配。
- `已完成` 视图 SHALL 不展示或启用开放 TODO 专用的排序、批量收起、批量展开操作。
- `未执行` 和 `执行中` 视图现有排序与批量展开/收起行为不变。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: TODO 状态视图切换时，顶部控制区高度和状态切换栏宽度必须保持稳定。

## Impact

- 前端：`frontend/src/components/ProjectSidebar.vue` 的 TODO 顶部控制区渲染逻辑。
- 样式：`frontend/src/style.css` 的 TODO toolbar / tabs 布局约束。
- 测试：`frontend/src/components/ProjectSidebar.test.js` 增加布局稳定性覆盖。
- 数据/API：不改变 Go 后端、Wails API、持久化数据或 TODO 工作流状态语义。
