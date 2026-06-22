## 1. 数据模型与后端状态流

- [x] 1.1 在 Go `Todo` 模型中新增可选 `startedAt` 字段，并确保 workspace JSON 读写兼容旧数据。
- [x] 1.2 修改 `ChangeTodoStatus`，在 TODO 从 `not-started` 成功进入 `in-progress` 时写入 UTC RFC3339 开始执行时间。
- [x] 1.3 保持 `CompleteTodo` 写入 `completedAt` 的现有行为，并确认完成后的 TODO 保留 `startedAt`。
- [x] 1.4 更新或重新生成 Wails 前端模型，使 `Todo` 包含 `startedAt`。

## 2. 前端展示

- [x] 2.1 在 TODO 侧边栏中添加持续时长解析与格式化逻辑，使用 `startedAt` 与 `completedAt` 计算完成耗时。
- [x] 2.2 在 `已完成` 视图的 TODO meta 信息中展示持续时长，缺少有效开始时间或出现负数时长时使用降级展示。
- [x] 2.3 确认持续时长展示不影响已完成列表按完成时间倒序排序、项目快照展示和批量删除交互。

## 3. 自动化测试

- [x] 3.1 添加后端测试，覆盖开始执行时写入 `startedAt`、完成后保留 `startedAt`、非 `not-started` 状态不能重新写入开始时间。
- [x] 3.2 添加客户端自动化测试，覆盖已完成 TODO 展示持续时长。
- [x] 3.3 添加客户端自动化测试，覆盖历史 completed TODO 缺少 `startedAt` 时不使用 `createdAt` 推断持续时长。
- [x] 3.4 添加客户端自动化测试，覆盖完成时间早于开始时间时不展示负数持续时长。

## 4. 验证与交付

- [x] 4.1 运行 Go 测试，确认后端 TODO 工作流仍通过。
- [x] 4.2 运行前端测试，确认 TODO 侧边栏展示和交互仍通过。
- [x] 4.3 运行 OpenSpec 校验，确认 change artifacts 和规格格式有效。
- [x] 4.4 执行自动 review，检查代码质量、边界条件、测试覆盖和是否符合既有实现风格。
- [x] 4.5 运行 `wails build -tags webkit2_41`，生成可执行文件。
