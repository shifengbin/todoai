## 1. 后端状态与持久化

- [x] 1.1 在 `ProjectState` 中新增 workspace 级项目分支偏好字段，并定义可序列化的偏好结构。
- [x] 1.2 在加载 workspace 项目状态时兼容缺少偏好字段的旧数据，并确保保存时不会写入全局项目候选文件。
- [x] 1.3 实现统一 helper，根据成功保存的项目选择更新 `projectBranchPreferences`，并保留空字符串作为有效上次选择。
- [x] 1.4 在创建 TODO、编辑 TODO、为已有 TODO 添加项目的成功保存路径中调用偏好更新逻辑。

## 2. 前端默认值与模型

- [x] 2.1 更新 Wails 前端模型，使 `ProjectState` 暴露 `projectBranchPreferences`。
- [x] 2.2 在前端状态应用逻辑中保存并传递 workspace 级项目分支偏好。
- [x] 2.3 调整项目分支默认值解析：优先使用偏好 map 中存在的项目记录，缺失时回退到既有默认分支逻辑。
- [x] 2.4 确保编辑已有 TODO 时，已关联项目显示自己的 `TodoProject.baseBranch`，新增项目才使用 workspace 偏好默认值。
- [x] 2.5 确认取消创建、取消编辑或关闭添加项目弹窗不会触发前端本地偏好写入。

## 3. 自动化测试

- [x] 3.1 补充 Go 测试，覆盖创建 TODO、编辑 TODO、添加项目成功后更新 workspace 分支偏好。
- [x] 3.2 补充 Go 测试，覆盖旧数据缺少偏好字段、空字符串偏好持久化、偏好不写入全局候选文件。
- [x] 3.3 补充前端测试，覆盖创建 TODO 和添加项目时默认使用上次分支选择。
- [x] 3.4 补充前端测试，覆盖空字符串偏好不回退、取消表单不影响下次默认值、编辑已有 TODO 保留自身分支。
- [x] 3.5 执行后端测试 `go test ./...`。
- [x] 3.6 执行客户端自动化测试 `cd frontend && npm run test`。

## 4. 质量检查与构建

- [x] 4.1 执行自动 review 或等效代码自检，确认实现符合 spec、design 和现有代码风格。
- [x] 4.2 执行前端构建 `cd frontend && npm run build`。
- [x] 4.3 执行 Wails 打包 `wails build -tags webkit2_41`。
