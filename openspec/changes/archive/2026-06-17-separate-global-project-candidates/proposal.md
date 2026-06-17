## Why

当前项目库既是用户管理项目候选的入口，也是 TODO 工程关联的事实来源；删除项目库项目会移除 active TODO 中的工程关联，导致已加入工程的 TODO 上下文不够完整。需要将跨 workspace 共享的全局候选项目与 workspace 内 TODO 工程副本分离，确保候选项管理不会破坏已有 TODO 工程数据。

## What Changes

- 新增跨所有 workspace 共享的全局项目候选库，用于导入、搜索、选择和清空候选项目。
- 将 TODO 下的工程改为 workspace 内独立副本，添加时从全局候选复制项目名称、路径和来源 ID。
- 全局候选项目被删除或清空后，已加入 TODO 的工程继续保留名称、路径、终端上下文和 Git 状态能力。
- 移除左侧工作区中的 `项目` tab，左侧仅保留 TODO 工作树。
- 将原项目库的导入、清除和候选管理能力移动到创建 TODO / 添加工程相关弹窗中。
- 旧 workspace 的项目库在打开时自动合并到全局候选库，并为旧 TODO 工程关联补齐独立副本数据。
- **BREAKING**: 项目候选列表不再按 workspace 隔离；按路径去重后跨 workspace 共享。
- **BREAKING**: 删除或清空全局候选项目不再移除 TODO 工程关联或关闭已有 TODO 工程终端。

## Capabilities

### New Capabilities

- `global-project-candidates`: 管理跨所有 workspace 共享的全局项目候选库，包括导入、清空、路径去重和旧 workspace 项目合并。

### Modified Capabilities

- `project-workspace`: 项目库从 workspace 本地打开项目列表调整为全局候选项目管理，并移除从项目库选择/删除直接影响 TODO 工程的行为。
- `todo-workspace`: 左侧 workspace tab 调整为仅 TODO 树；创建 TODO、编辑 TODO 和添加工程弹窗使用全局候选，并在 TODO 内保存工程副本。

## Impact

- 后端项目数据模型需要拆分全局候选项目存储与 workspace TODO 工程存储。
- `ProjectState`、Wails 绑定模型和前端状态需要暴露全局候选项目以及包含名称/路径的 TODO 工程副本。
- `ProjectManager`、`App` workspace 绑定和迁移逻辑需要处理旧 workspace 项目合并与 TODO 工程副本补齐。
- `ProjectSidebar.vue` 需要移除项目 tab，并把候选项目管理入口整合到 TODO 工程选择弹窗。
- 测试需要覆盖全局候选跨 workspace 共享、TODO 工程副本不受候选清空影响、旧数据迁移和同一 TODO 按路径去重。
