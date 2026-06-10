## 1. 后端 Git 状态模型与解析

- [x] 1.1 添加失败的 Go 单元测试，覆盖 porcelain v2 输出中的分支名、changed/staged/unstaged/untracked/ahead/behind 解析。
- [x] 1.2 实现 `GitStatus` 模型、porcelain v2 解析函数和改动数量统计逻辑。
- [x] 1.3 添加失败的 Go 单元测试，覆盖非 Git 仓库、缺少 Git 可执行文件或 Git 命令失败时的返回行为。
- [x] 1.4 实现项目路径下的 Git 状态查询，使用短超时，并将非 Git 仓库转换为 `isRepo=false`。

## 2. Wails 应用 API

- [x] 2.1 添加失败的 `App` API 测试，覆盖 `GetProjectGitStatus` 对可用项目、不可用项目和不存在项目的行为。
- [x] 2.2 实现 `App.GetProjectGitStatus(projectID string)`，通过项目 ID 查找路径并返回运行时 Git 状态。
- [x] 2.3 更新或重新生成 Wails 前端绑定，使前端可调用新的 Git 状态 API 并获得对应类型。

## 3. 前端状态栏行为

- [x] 3.1 添加失败的 `App.vue` 测试，覆盖 active Git 项目显示分支和改动数量。
- [x] 3.2 添加失败的 `App.vue` 测试，覆盖未选择项目、非 Git 仓库、项目路径不可用和 Git 查询失败的状态栏显示。
- [x] 3.3 实现前端 Git 状态 state、loading/error 状态和底部状态栏渲染。
- [x] 3.4 添加失败的 `App.vue` 测试，覆盖项目切换、终端命令结束和窗口 focus 时刷新 Git 状态。
- [x] 3.5 实现状态栏刷新触发逻辑，并防止过期请求覆盖当前激活项目的状态。

## 4. 布局与验证

- [x] 4.1 样式化固定高度底部状态栏，确保终端 surface 高度稳定，错误信息不会造成布局抖动。
- [x] 4.2 运行 Go 测试，覆盖 Git 状态解析、查询和 App API 行为。
- [x] 4.3 运行前端单元测试，覆盖 App 状态栏显示和刷新行为。
- [x] 4.4 运行项目构建，确认 Wails 绑定、前端编译和 Go 编译通过。
