## 1. 后端终端上下文

- [x] 1.1 扩展终端模型以标识 workspace 全局终端，并保持旧终端历史数据向后兼容。
- [x] 1.2 新增创建 workspace 全局终端的应用 API，使用当前 workspace 根目录作为 shell 工作目录。
- [x] 1.3 调整终端选择逻辑，使选择全局终端只更新 active terminal，不调用项目或 TODO project 选择逻辑。
- [x] 1.4 调整终端删除、workspace 关闭和历史恢复逻辑，确保全局终端与 TODO project 终端隔离。
- [x] 1.5 为全局终端创建、多个全局终端、选择隔离、删除隔离、workspace 关闭清理补充 Go 单元测试。

## 2. Wails 绑定与前端状态

- [x] 2.1 重新生成 Wails 前端绑定，暴露全局终端创建和更新后的终端模型。
- [x] 2.2 在前端区分 workspace 全局终端和 TODO project 终端，左侧 TODO 树只接收 TODO project 终端。
- [x] 2.3 在终端区域顶部实现全局终端分组、创建入口、选择和删除操作，并在没有全局终端时完全隐藏分组。
- [x] 2.4 调整 active terminal 激活、重启、resize、clipboard 和 context menu 逻辑，使全局终端复用现有 xterm 能力。

## 3. 项目导入与 TODO 交互

- [x] 3.1 调整单个项目导入处理，在创建 TODO、编辑 TODO 或添加项目控件打开时默认选中新导入候选。
- [x] 3.2 确保默认选中新候选只改变当前控件临时状态，取消选择或关闭控件不会创建 TODO 工程副本。
- [x] 3.3 在 TODO header 行添加双击展开/收起手势，保持行内按钮、菜单和确认气泡不触发展开/收起。
- [x] 3.4 为导入默认选中、取消默认选择、双击 header、双击按钮不切换补充前端自动化测试。

## 4. 验证与交付

- [x] 4.1 运行 Go 测试，确认后端终端、项目和 TODO 行为没有回归。
- [x] 4.2 运行前端自动化测试，确认客户端交互和状态渲染符合规格。
- [x] 4.3 执行自动代码 review，检查实现是否符合设计、规格和现有代码风格。
- [x] 4.4 运行 `openspec status --change improve-workspace-ux-shortcuts`，确认变更处于可应用状态。
- [x] 4.5 运行 `wails build -tags webkit2_41` 生成可执行文件。
