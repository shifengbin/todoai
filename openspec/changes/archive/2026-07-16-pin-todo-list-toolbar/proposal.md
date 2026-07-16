## Why

TODO 排序、批量展开/收起以及已完成 TODO 批量操作共用的工具栏当前位于列表滚动区域内，长列表向下滚动后这些高频控件会离开可视区域，用户必须返回列表顶部才能继续操作。需要让工具栏与已经固定的工作流状态页签保持一致，在列表滚动期间始终可见。

## What Changes

- 将三个 TODO 状态视图共用的顶部工具栏固定在列表滚动区域之外。
- 在 `未执行` 和 `执行中` 视图中，排序切换与全部展开/收起控件在列表滚动时保持可见且可操作。
- 在 `已完成` 视图中，已选数量与批量删除控件在列表滚动时保持可见且可操作。
- 保持 TODO 列表独立滚动，并保留手动排序时的拖拽边缘自动滚动能力。
- 增加布局合同与交互回归测试，验证工具栏层级、三种视图内容切换和拖拽滚动目标。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 扩展 TODO 工作区滚动要求，使三个状态视图共用的顶部工具栏位于列表滚动区域之外并在滚动期间保持可见。

## Impact

- 前端组件：`frontend/src/components/ProjectSidebar.vue`
- 前端样式：`frontend/src/style.css`
- 前端测试：`frontend/src/components/ProjectSidebar.test.js`
- 不改变后端 API、Wails 绑定、持久化数据结构或第三方依赖。
