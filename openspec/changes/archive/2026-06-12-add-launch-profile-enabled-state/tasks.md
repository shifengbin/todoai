## 1. 后端持久化与迁移

- [x] 1.1 为 `TerminalLaunchProfileSetting` 增加 `enabled` 状态，并确保默认 launch profiles 归一化为启用。
- [x] 1.2 调整 settings JSON 加载逻辑，区分旧配置缺少 `enabled` 与显式 `enabled: false`，缺字段时按启用迁移。
- [x] 1.3 调整 `SaveLaunchProfiles` 归一化和保存逻辑，保留每个 profile 的启用状态且继续执行 name/command/重复名称验证。
- [x] 1.4 更新后端测试，覆盖默认 profiles 启用、旧配置缺字段迁移为启用、显式禁用可保存并重载、无效禁用 profile 仍被拒绝。

## 2. 前端设置与启动菜单

- [x] 2.1 更新 Wails TypeScript 模型，使 launch profile 包含 `enabled` 字段。
- [x] 2.2 更新 `App.vue` 的默认 profiles、clone、add、normalize 和保存逻辑，新建 profile 默认启用，旧 state 缺字段时按启用处理。
- [x] 2.3 在 Settings 的 Launch profiles 行中增加启用切换控件，禁用 profile 仍可编辑、排序、删除和重新启用。
- [x] 2.4 更新 `ProjectSidebar` launch menu，只展示启用的自定义 profiles，并始终保留内置 `Terminal`。
- [x] 2.5 保持启动已启用 profile 的行为不变，选择 profile 后继续创建终端并提交对应启动命令。

## 3. 自动化测试

- [x] 3.1 更新 `frontend/src/App.test.js`，覆盖设置页渲染启用状态、保存禁用状态、新增 profile 默认启用、旧 state 缺字段默认启用。
- [x] 3.2 更新 `frontend/src/components/ProjectSidebar.test.js` 或 App 集成测试，覆盖禁用 profile 从启动菜单隐藏、全部自定义 profile 禁用时仅显示 `Terminal`、启用 profile 仍可启动。
- [x] 3.3 运行后端测试：`go test ./...`。
- [x] 3.4 运行客户端自动化测试：`npm test -- --run`（在 `frontend/` 目录）。

## 4. 质量检查与打包

- [x] 4.1 执行自动 review，检查数据迁移、旧配置兼容、菜单过滤和测试覆盖是否符合 spec。
- [x] 4.2 运行 OpenSpec 校验或状态检查，确认 change apply-ready 且 artifact 状态正确。
- [x] 4.3 运行 `wails build -tags webkit2_41`，生成可执行文件。
