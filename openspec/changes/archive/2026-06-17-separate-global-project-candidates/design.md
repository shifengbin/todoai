## Context

现有实现把 `ProjectState.Projects` 作为当前 workspace 的项目库，同时让 `TodoProject` 只保存 `projectId` 引用该项目库。结果是项目库既承担候选项管理，又承担 TODO 工程数据源；删除项目库项目会删除 active TODO 中的工程关联并关闭相关终端。

新的产品语义要求项目候选跨所有 workspace 共享，但加入 TODO 后必须成为 workspace 内的独立工程副本。全局候选的清空、删除、导入只影响后续选择，不影响已经加入 TODO 的工程上下文。

## Goals / Non-Goals

**Goals:**

- 提供跨所有 workspace 共享的全局项目候选库。
- TODO 工程在添加时复制候选项目的名称、路径和来源 ID，并在 workspace 内独立持久化。
- 全局候选删除或清空后，已有 TODO 工程继续可见、可选、可启动终端、可刷新 Git 状态。
- 移除左侧 `项目` tab，把候选项目导入和清空入口放进创建 TODO / 添加工程相关弹窗。
- 自动迁移旧 workspace 项目库：按路径合并进全局候选，并为旧 `todoProjects` 补齐工程副本字段。

**Non-Goals:**

- 不做全局候选与 TODO 工程副本之间的后续同步；添加后名称和路径以 TODO 工程副本为准。
- 不删除磁盘上的真实项目目录。
- 不引入多用户同步或远程项目库。
- 不改变 TODO、终端、Git 状态的核心工作流，只改变项目数据来源。

## Decisions

### 1. 全局候选使用应用级配置文件

全局候选项目存储在应用配置目录下，例如 `global-project-candidates.json`。该目录已用于 recent workspace 和 settings，适合保存跨 workspace 的用户级偏好数据。workspace 打开时不再把候选项目绑定到 workspace 的 `.tui-helper/projects.json`。

备选方案是继续把候选项目放在每个 workspace 内，再在 UI 上做复制。该方案不能满足跨 workspace 共享候选的要求，也会让同一目录在多个 workspace 中重复导入。

### 2. TODO 工程保存副本而不是引用全局候选

`TodoProject` 需要从只保存 `projectId` 扩展为保存：

- `sourceProjectId`: 来源全局候选 ID，可为空或指向已删除候选。
- `name`: 添加时复制的项目名称。
- `path`: 添加时复制的绝对路径。
- `available`: 当前路径可用性，可在加载或状态刷新时计算。
- `createdAt` / `lastSelectedAt`: 保持现有排序和选择行为。

终端启动、Git 状态、TODO 树展示和已完成快照都以 TODO 工程副本的 `path/name` 为准。`activeTodoProjectId` 成为工作上下文的权威标识；`activeProjectId` 若保留，应作为兼容字段从 active TODO 工程推导，而不是作为查找全局候选的必需条件。

备选方案是只给 `TodoProject` 加快照字段，但继续以全局项目为事实来源。该方案在全局候选被删除后仍会让选择、终端和 Git 路径解析失败。

### 3. 按路径去重

全局候选库按规范化绝对路径去重。导入父目录时已存在路径的候选会跳过或更新可用性。同一个 TODO 下也按规范化路径去重，不允许同一路径出现多个工程副本。不同 TODO 可以各自保存同一路径的独立副本。

备选方案是按候选 ID 去重，但清空并重新导入同一路径会生成新 ID，导致同一 TODO 下重复添加同一路径。

### 4. 弹窗承接候选管理

左侧 sidebar 只保留 TODO 工作树。全局候选管理入口放到创建 TODO、编辑 TODO 和添加工程弹窗中：

- 搜索候选项目。
- 从父目录导入候选。
- 清空全局候选。
- 选择候选加入 TODO。

这样原 `项目` tab 不再占据主导航，同时候选管理发生在用户真正需要给 TODO 添加工程的上下文中。

### 5. 旧数据迁移在 workspace 加载时执行

旧 workspace 的 `projects` 字段仍可能存在。加载 workspace 时需要：

1. 读取旧 workspace 项目库。
2. 将这些项目按路径合并进全局候选库。
3. 为旧 `todoProjects` 根据旧 `projectId` 查找项目名称和路径，并补齐 `sourceProjectId/name/path`。
4. 保存迁移后的 workspace 状态，避免重复迁移造成副作用。

如果旧 `todoProject.projectId` 找不到项目，保留现有关联 ID，并以 `Missing project` 或路径不可用状态展示；不能阻止 workspace 打开。

## Risks / Trade-offs

- [Risk] 数据模型迁移影响面较大，`ProjectState`、Wails 绑定和前端计算属性都要同步更新。→ Mitigation: 先用后端单元测试锁定迁移和删除候选不影响 TODO 工程，再更新前端测试。
- [Risk] `activeProjectId` 旧语义依赖项目库 ID。→ Mitigation: 将 active TODO 工程作为路径解析事实来源，保留兼容字段但不再用它反查全局候选。
- [Risk] 清空全局候选后用户可能误以为会删除 TODO 工程。→ Mitigation: UI 文案明确表达“只清空候选，不影响已加入 TODO 的工程”。
- [Risk] 旧 workspace 中项目路径不可访问。→ Mitigation: 迁移仍保存名称和路径，加载时标记不可用，终端启动继续被阻止。

## Migration Plan

1. 添加全局候选存储和 manager，使用应用配置目录持久化。
2. 扩展 workspace `TodoProject` 数据结构，并兼容旧 JSON 中只有 `projectId` 的记录。
3. 在 workspace 绑定或加载时执行旧项目库合并与 TODO 工程副本补齐。
4. 调整后端创建 TODO、更新 TODO、添加工程、选择 TODO 工程、终端创建和 Git 状态路径解析。
5. 调整前端状态模型和弹窗交互，移除项目 tab。
6. 更新 Wails 绑定和测试。

回滚时可保留新增字段；旧版本会忽略未知 JSON 字段，但如果旧版本仍要求 `projects` 包含 TODO 工程引用，已迁移数据可能无法完整还原旧行为。

## Open Questions

无。当前约定为：全局候选跨所有 workspace，共享候选；选进 TODO 后保存 workspace 内副本；同一 TODO 下按路径去重。
