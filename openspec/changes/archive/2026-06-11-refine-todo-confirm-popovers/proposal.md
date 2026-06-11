## Why

TODO 工作树里已经有项目移除的行内确认气泡，但完成 TODO 和删除 TODO 仍使用浏览器原生确认框，交互方式不一致且脱离当前行上下文。TODO 下的项目行选中/悬停背景也只覆盖项目信息区域，没有延伸到创建终端和删除按钮区域，导致整行状态表达不完整。

## What Changes

- TODO item 的完成按钮 SHALL 在按钮旁显示确认气泡，用户确认后才完成 TODO。
- TODO item 的删除按钮 SHALL 在按钮旁显示确认气泡，用户确认后才删除 TODO。
- TODO 完成/删除确认气泡 SHALL 支持取消、点击外部关闭，以及切换到其它浮层时关闭。
- TODO 完成/删除 SHALL 不再依赖浏览器原生 `window.confirm`。
- TODO 下的项目行背景 SHALL 在选中和悬停状态下覆盖整条项目 header，包括项目信息、创建终端按钮和删除按钮区域。
- 创建终端按钮和删除按钮 SHALL 在整行背景上保持可读，并保留各自 hover/focus 反馈。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-workspace`: 明确 TODO 完成/删除使用行内确认气泡，并补充 TODO 项目行整行背景覆盖要求。

## Impact

- `frontend/src/components/ProjectSidebar.vue` 需要新增 TODO 完成/删除确认气泡状态和模板。
- `frontend/src/App.vue` 需要移除 TODO 完成/删除流程中的原生确认调用。
- `frontend/src/style.css` 需要调整 TODO 项目行选中/悬停背景覆盖范围，并补充 TODO 操作确认气泡样式。
- `frontend/src/components/ProjectSidebar.test.js` 和 `frontend/src/App.test.js` 需要更新确认交互、取消行为和整行背景覆盖断言。
