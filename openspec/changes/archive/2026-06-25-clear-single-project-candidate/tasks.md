## 1. 前端交互实现

- [x] 1.1 在 `App.vue` 增加清除单个候选项目的处理函数，复用 `DeleteProject(projectId)` 并在确认取消时不调用后端。
- [x] 1.2 清除成功后从 `todoForm.projectSelections`、`todoDetail.projectSelections` 和 `projectPicker.projectSelections` 中移除对应项目 ID。
- [x] 1.3 将三个候选项目列表的候选项拆为行容器，保留原选择按钮 `data-testid`，新增单项清除图标按钮和稳定 `data-testid`。
- [x] 1.4 调整候选项目列表样式，保证选择区域、项目名称、路径和清除按钮在桌面与窄宽度下不重叠。

## 2. 测试覆盖

- [x] 2.1 在 `frontend/src/App.test.js` 增加确认清除单个候选项目的测试，验证只移除目标候选并保留其他候选。
- [x] 2.2 增加取消确认的测试，验证不会调用 `DeleteProject` 且候选仍显示。
- [x] 2.3 增加已选候选被清除后的测试，验证当前弹窗待提交选择同步移除且提交不引用已清除项目 ID。
- [x] 2.4 增加回归测试，验证清除候选后 TODO 已保存工程副本仍显示，相关终端不被移除。

## 3. 验证与质量

- [x] 3.1 运行客户端自动化测试，至少覆盖 `frontend/src/App.test.js`。
- [x] 3.2 运行 Go 测试，确认既有 `DeleteProject`/`DeleteProjects` 安全边界未回归。
- [x] 3.3 运行 OpenSpec 校验，确认 `global-project-candidates` delta 可归档。
- [x] 3.4 执行自动 review，检查交互、测试和规格是否符合本 change。
- [x] 3.5 运行 `wails build -tags webkit2_41` 生成可执行文件。

## 4. 反馈调整

- [x] 4.1 将单项候选清除确认从系统原生弹框改为应用内自定义确认弹窗，并补充自动化测试。
- [x] 4.2 增加候选列表滚动条与删除按钮之间的右侧间距，降低误触风险。
