## 1. 后端任务级上下文

- [x] 1.1 更新 `App.SelectTerminal` 的任务级终端分支，使其返回 active TODO task context，并清空 `ActiveTodoProjectID` 与 `ActiveProjectID`
- [x] 1.2 更新 `App.CreateTaskTerminal` 返回状态，使新建任务级终端成为 active terminal 时同步进入任务级上下文
- [x] 1.3 调整或新增 Go 单元测试，覆盖选择任务级终端和创建任务级终端后 active 字段的变化

## 2. 前端任务级上下文展示

- [x] 2.1 在 `App.vue` 中增加任务级终端上下文派生，优先显示 `TODO 标题 / 任务终端`
- [x] 2.2 在 `App.vue` 中由当前 workspace 和 TODO `workspaceDirName` 派生任务目录路径，并在任务级终端激活时展示该路径
- [x] 2.3 在 `ProjectSidebar.vue` 中统一 TODO active 判定，使 active task terminal 的父 TODO 使用选中背景
- [x] 2.4 保持任务级终端行自身 active 样式，并确保 TODO project 行不会因任务级终端选择而误显示为 active

## 3. 终端启动菜单浮层

- [x] 3.1 将任务级和 TODO project 终端启动菜单渲染到不受 TODO 列表滚动裁剪的 fixed/Teleport 浮层
- [x] 3.2 基于触发按钮位置和侧栏可视区域计算菜单上翻、下翻、最大高度和滚动行为
- [x] 3.3 确保点击菜单项、点击外部、打开其他浮层、滚动或窗口 resize 时关闭或重算菜单，避免位置陈旧

## 4. 自动化测试

- [x] 4.1 补充 `ProjectSidebar` 测试，覆盖 active task terminal 高亮父 TODO 且不高亮 TODO project
- [x] 4.2 补充 `App` 前端测试，覆盖任务级终端标题和任务目录路径展示
- [x] 4.3 补充终端启动菜单测试，覆盖长列表底部任务级和 TODO project 启动菜单不被裁剪并可选择启动项
- [x] 4.4 运行前端自动化测试，例如 `npm test -- --run`
- [x] 4.5 运行 Go 自动化测试，例如 `go test ./...`

## 5. 质量检查与打包

- [x] 5.1 执行自动 review，检查实现是否满足 proposal、design 和 specs，且未引入无关 UI/状态回归
- [x] 5.2 修复 review 或测试发现的问题并重新运行相关验证
- [x] 5.3 运行 `wails build -tags webkit2_41` 生成可执行文件
