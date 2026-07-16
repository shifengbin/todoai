## 1. 前端测试先行

- [x] 1.1 在 `ProjectSidebar.test.js` 增加虚拟 TODO 仅在 `执行中` 且存在全局终端时显示、固定首位、空状态隐藏及不暴露真实 TODO 动作的失败测试
- [x] 1.2 在 `ProjectSidebar.test.js` 增加虚拟 TODO 展开收起、活动优先级聚合、批量折叠及拖拽后恢复展开状态的失败测试
- [x] 1.3 在 `ProjectSidebar.test.js` 增加全局终端父子选中、新增、选择、删除、键盘操作及不进入排序 ID 的失败测试
- [x] 1.4 在 `App.test.js` 增加顶部创建成功后切换到 `执行中`、失败时保持原视图和上下文、主区域不再显示 Global 标签组但保留 xterm pane 的失败测试

## 2. 实现侧栏虚拟 TODO

- [x] 2.1 为 `ProjectSidebar.vue` 增加独立的 workspace 全局终端 prop 与创建、选择、删除事件接口，并从活动终端推导虚拟父子选中状态
- [x] 2.2 提取通用终端活动聚合函数，让真实 TODO 与虚拟 TODO 共用 `needs-input > needs-ack > busy > idle` 规则
- [x] 2.3 在 `执行中` 真实 TODO 列表前渲染固定的 `Global 终端` 虚拟 TODO、全局终端子项及新增和删除入口，并补齐键盘、焦点和辅助文本
- [x] 2.4 实现虚拟 TODO 的独立展开状态、双击折叠、数量增加后自动展开，并接入批量展开收起与拖拽临时折叠恢复流程
- [x] 2.5 调整侧栏样式，使虚拟 TODO 复用真实 TODO 的选中和活动反馈，同时不显示优先级、状态动作、上下文菜单或拖拽手柄

## 3. 接入 App 与移除顶部标签组

- [x] 3.1 在 `App.vue` 向侧栏传递带关注状态的 `workspaceTerminals`，并连接现有全局终端创建、选择和删除 handler
- [x] 3.2 调整全局终端创建时序，仅在后端成功并应用状态后通过现有 UI 状态保存路径切换到 `in-progress`，失败时不修改视图或选择
- [x] 3.3 删除主终端区域 Global 标签组、条件布局、专用标签 helper 与相关样式，保留顶部创建按钮和所有终端 xterm pane
- [x] 3.4 确认删除最后一个全局终端后虚拟节点及选中反馈消失，并确认选择全局终端不改写真实 TODO、项目和 Git 上下文

## 4. 自动化验证与交付

- [x] 4.1 在 `frontend` 目录运行 `npm test`，修复虚拟 TODO、App 集成及现有前端回归测试失败
- [x] 4.2 在 `frontend` 目录运行 `npm run build`，确认 Vue 模板、样式和生产构建通过
- [x] 4.3 运行 `go test ./...`，确认现有 workspace terminal 后端行为与 Wails 集成未回归
- [x] 4.4 执行自动代码 review，核对实现与 proposal、design、delta specs 一致，重点检查真实 TODO 数据隔离、失败时序、无障碍交互和废弃样式清理，并修复发现的问题
- [x] 4.5 运行 `wails build -tags webkit2_41`，生成并核验可执行文件
