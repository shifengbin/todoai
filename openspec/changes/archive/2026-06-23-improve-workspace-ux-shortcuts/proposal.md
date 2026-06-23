## Why

当前工作流里有三个高频摩擦点：单个导入项目后还需要再次手动选择，临时执行命令必须先进入某个 TODO 项目上下文，TODO 分支展开/收起只能点较小的箭头按钮。这个变更优化这些路径，让用户在创建任务、临时运行命令和浏览任务树时少做无关操作。

## What Changes

- 单个导入全局项目候选后，如果用户当前正在创建 TODO、编辑 TODO 或为 TODO 添加项目，系统默认把新导入候选加入当前控件的已选项目列表；用户仍需确认表单才会真正关联到 TODO。
- 新增 workspace 级全局终端会话能力，用户可以创建多个不绑定 TODO 项目的终端，会话默认工作目录为当前 workspace 根目录。
- 终端区域顶部显示全局终端分组和创建入口；当不存在全局终端时，该分组完全不渲染、不占用布局高度。
- 选择全局终端只激活该终端，不改变当前 TODO、TODO 项目或项目 Git 状态上下文。
- TODO header 行支持双击展开/收起，行内按钮和菜单保持原有点击行为，不触发展开/收起。

## Capabilities

### New Capabilities

- `workspace-global-terminals`: workspace 级全局终端的创建、选择、删除、工作目录和终端区展示行为。

### Modified Capabilities

- `todo-workspace`: 修改 TODO 项目选择控件中的单个导入默认选择行为，并新增 TODO header 双击展开/收起交互。
- `project-workspace`: 明确单个导入候选在当前选择控件中默认选中，但不自动关联 TODO、不改变当前工作上下文。
- `embedded-shell-sessions`: 允许终端会话归属 workspace 全局上下文，并保持与 TODO 项目终端隔离。

## Impact

- 后端 Go：项目导入返回状态、shell session 终端模型、workspace 根目录终端创建/选择/删除 API、历史持久化和状态恢复。
- 前端 Vue：项目选择弹窗导入后的默认勾选逻辑、终端区顶部全局终端分组、全局终端创建/选择/删除/重启交互、TODO header 双击事件。
- Wails bindings：新增 workspace 全局终端相关方法后需要重新生成前端绑定。
- 测试：补充 Go 单元测试和前端组件/应用测试，覆盖默认勾选、全局终端隔离、隐藏分组和双击手势。
