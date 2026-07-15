## Why

创建 TODO 的用户只需要填写所选生命周期脚本的参数值，参数区域上方的使用方法问号与该流程无关，并增加了不必要的视觉元素。参数模板的编写说明仍应保留在全局生命周期脚本管理界面中，供配置脚本的用户查阅。

## What Changes

- 移除创建 TODO 表单中生命周期脚本参数标题旁的问号帮助按钮，以及由该按钮触发的悬浮提示。
- 保留创建 TODO 表单中的参数标题、输入项、参数描述、默认值初始化和必填校验行为。
- 保留全局生命周期脚本管理界面中的参数使用方法问号及其悬浮提示。
- 调整前端测试和生命周期脚本规范，使其不再要求创建 TODO 表单提供参数使用方法帮助入口。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-lifecycle-scripts`: 移除创建 TODO 表单必须提供生命周期脚本参数使用方法帮助提示的要求，同时保持参数录入和全局脚本管理帮助行为不变。

## Impact

- 前端：`frontend/src/App.vue` 中创建 TODO 表单的生命周期脚本参数标题区域。
- 测试：`frontend/src/App.test.js` 中创建 TODO 时参数帮助提示的交互测试。
- 规范：`openspec/specs/todo-lifecycle-scripts/spec.md` 对创建 TODO 参数区域的要求和场景。
- 不影响 Go 后端、Wails API、数据模型、参数快照、参数校验或生命周期脚本执行逻辑，也不新增依赖。
