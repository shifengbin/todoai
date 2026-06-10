## Why

当前底部 Git 状态栏只显示一段纯文本，分支、改动数量和异常状态不够醒目，用户很难快速扫描当前项目的仓库状态。对于不是 Git 仓库的项目，状态栏只能提示问题，不能直接完成初始化，打断了从项目打开到版本管理的工作流。

## What Changes

- 将底部 Git 状态栏从单段文本升级为多个彩色圆角信息块，分别展示分支、改动数量、暂存、未暂存、未跟踪以及 ahead/behind 信息。
- 在当前项目不是 Git 仓库时，状态栏显示非仓库提示，并提供 `Initialize Git Repository` 按钮。
- 用户点击初始化按钮后，系统直接在当前项目路径执行 `git init`，成功后刷新 Git 状态栏。
- 初始化执行中禁用按钮并显示进行中状态，失败时以非阻断错误信息反馈。
- 保持没有项目、项目路径不可用、Git 查询失败等状态的稳定底部布局。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `project-workspace`: 扩展当前项目 Git 状态栏的展示结构，并增加非 Git 仓库的一键初始化行为。

## Impact

- 前端：`frontend/src/App.vue` 的 Git 状态栏状态建模、模板和交互；`frontend/src/style.css` 的状态栏 chip/button 样式；`frontend/src/App.test.js` 的状态栏和初始化交互测试。
- 后端：新增 Wails 公开方法用于初始化当前项目 Git 仓库；复用现有项目查找和路径可用性校验。
- Wails 绑定：新增前端可调用的 `InitializeProjectGitRepository` 绑定。
- 测试：增加 Go 单元测试覆盖初始化路径、不可用项目和缺失项目；增加前端测试覆盖非仓库按钮、点击后刷新、loading/错误状态。
