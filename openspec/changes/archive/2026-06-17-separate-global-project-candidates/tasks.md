## 1. 后端数据模型与全局候选存储

- [x] 1.1 新增全局项目候选持久化结构和 manager，使用应用级配置目录保存跨 workspace 候选项目。
- [x] 1.2 扩展 `ProjectState` / `TodoProject` 数据模型，使 TODO 工程副本保存 `sourceProjectId`、`name`、`path` 和路径可用性。
- [x] 1.3 实现全局候选导入、按规范化绝对路径去重、删除和清空能力，且不影响 workspace 内 TODO 工程副本。
- [x] 1.4 实现旧 workspace 项目库迁移：将旧 `projects` 按路径合并到全局候选，并为旧 `todoProjects` 补齐工程副本字段。

## 2. 后端 TODO 工程语义

- [x] 2.1 调整创建 TODO、更新 TODO 和添加工程逻辑，使其从全局候选复制为 workspace 内 TODO 工程副本。
- [x] 2.2 调整同一 TODO 下工程去重逻辑，按规范化路径防止重复添加，不同 TODO 保持独立副本。
- [x] 2.3 调整选择 TODO 工程、终端创建、Git 状态刷新和完成 TODO 快照逻辑，使其读取 TODO 工程副本的名称和路径。
- [x] 2.4 调整删除/清空全局候选行为，确保不会移除 TODO 工程副本、不会关闭已有 TODO 工程终端。

## 3. Wails 绑定与前端状态

- [x] 3.1 更新 Wails 生成模型和前端类型使用，暴露全局候选项目与扩展后的 TODO 工程副本字段。
- [x] 3.2 调整 `App.vue` 状态计算和 API 调用，使创建 TODO、编辑 TODO、添加工程都使用全局候选并提交候选 ID。
- [x] 3.3 移除左侧 `项目` tab 和项目库视图，只保留 TODO 工作树。
- [x] 3.4 将导入单个工程、导入父目录、清空全局候选、候选搜索和多选加入 TODO 的交互整合到创建 TODO / 编辑 TODO / 添加工程弹窗。
- [x] 3.5 调整 TODO 树、终端区域和 Git 状态栏展示，使其使用 TODO 工程副本的名称、路径和可用性。

## 4. 自动化测试

- [x] 4.1 添加后端测试覆盖全局候选跨 workspace 共享、按路径去重、清空候选不影响 TODO 工程副本。
- [x] 4.2 添加后端测试覆盖旧 workspace 项目库迁移和旧 `todoProjects` 工程副本补齐。
- [x] 4.3 添加后端测试覆盖创建/更新/添加 TODO 工程按路径去重，以及全局候选删除后仍可选择 TODO 工程。
- [x] 4.4 添加前端自动化测试覆盖移除项目 tab、弹窗候选导入/清空入口、候选搜索和添加工程副本展示。
- [x] 4.5 运行 Go 测试和前端测试，确认相关测试通过。

## 5. 生成、Review 与打包

- [x] 5.1 运行 Wails 绑定生成命令，确认 `frontend/wailsjs` 与 Go 模型一致。
- [x] 5.2 运行自动 review 或静态检查，处理发现的代码质量和规范问题。
- [x] 5.3 运行 `openspec validate separate-global-project-candidates --strict`，确认变更规格仍有效。
- [x] 5.4 运行 `wails build -tags webkit2_41` 生成可执行文件。
