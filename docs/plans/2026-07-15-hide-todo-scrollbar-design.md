# 隐藏 TODO 列表滚动条设计

## 背景

三个 TODO 视图共用 `.todo-workspace-scroll` 作为垂直滚动容器。当前容器通过 `overflow-y: auto` 提供滚动，并通过 `scrollbar-gutter: stable` 预留可见滚动条空间。

## 目标

- 未开始、进行中和已完成视图均不显示滚动条。
- 保留鼠标滚轮、触控、键盘和拖拽边缘自动滚动能力。
- 不影响其他侧栏、弹窗或终端区域的滚动条。

## 方案

仅修改 `.todo-workspace-scroll`：保留 `overflow-y: auto`，增加 Firefox 的 `scrollbar-width: none` 和旧版兼容的 `-ms-overflow-style: none`，并通过 `.todo-workspace-scroll::-webkit-scrollbar` 隐藏 WebKit/Chromium 滚动条。

移除 `scrollbar-gutter: stable`，避免隐藏后仍保留滚动条槽位；保留现有右侧内边距，继续为 TODO 行和侧栏边缘提供稳定间距。弹窗打开时现有的 `overflow-y: hidden` 规则保持不变。

## 验证

先在 `ProjectSidebar.test.js` 中增加失败断言，确认滚动容器仍使用 `overflow-y: auto`、包含跨浏览器隐藏声明、WebKit 伪元素规则存在且不再预留滚动条槽位。随后修改样式并运行组件测试、完整前端测试和生产构建。
