## Context

当前应用把项目库、TODO 和终端历史保存在应用配置目录下，例如 `projects.json` 和 `terminal-history.json`。这让所有业务上下文共享同一份数据，无法把 TODO 和导入工程按“管理项目/workspace”隔离。终端设置 `settings.json` 作为应用级偏好继续保存在应用配置目录。

现有代码中 `Project` 表示被导入的代码工程目录，而本次需求中的“打开项目”表示打开一个管理 workspace。为避免概念冲突，设计上区分：

- workspace：用户从文件菜单打开的目录，承载一组 TODO、导入工程和终端上下文。
- project：workspace 内导入的代码工程目录，继续沿用现有 `Project` 模型。

目标运行形态：

```text
global app config
├── recent-workspaces.json
└── settings.json

workspace directory selected by user
└── .data/
    ├── projects.json
    └── terminal-history.json
```

## Goals / Non-Goals

**Goals:**

- 用户可以通过原生文件菜单打开、关闭 workspace，并从最近打开列表重新打开 workspace。
- 当前 workspace 的项目库、TODO、TODO-项目关联和终端历史保存在该 workspace 的 `.data` 目录；终端设置保持应用全局。
- 切换或关闭 workspace 时清理运行时终端，避免旧 workspace 的 PTY、输出和选中状态泄漏到新 workspace。
- 没有打开 workspace 时，应用展示明确空态，workspace 依赖操作不可执行。
- 旧全局数据有可恢复迁移路径，升级后不会不可见或被覆盖。

**Non-Goals:**

- 不实现 workspace 之间的数据合并、同步或共享。
- 不删除用户选择的 workspace 目录，也不在关闭 workspace 时删除 `.data`。
- 不改变 workspace 内 `Project`、`Todo`、`TodoProject` 的核心业务模型。
- 不把最近打开记录放进 workspace；最近打开仍是应用全局偏好。
- 不把终端 shell、启动配置和外观设置按 workspace 隔离；这些 settings 是全局偏好。

## Decisions

### 1. 新增 WorkspaceManager 管理当前 workspace 和最近打开

新增 Go 侧 workspace 管理器，负责：

- 规范化和校验 workspace 路径。
- 创建并解析 `<workspace>/.data`。
- 维护当前 workspace 元数据：路径、名称、最近打开时间。
- 读写全局 `recent-workspaces.json`。
- 打开 workspace 时返回当前 workspace 和最近打开快照。
- 关闭 workspace 时清空当前 workspace，但保留最近打开记录。

替代方案是把 workspace 状态塞进现有 `ProjectManager`。不采用该方案，因为 `ProjectManager` 已经表示 workspace 内项目库和 TODO 数据，把外层 workspace 生命周期放进去会让“管理项目”和“导入工程”继续混淆。

### 2. App 根据当前 workspace 重建或切换数据管理器

`App` 保留长期存在的 workspace manager。打开 workspace 后，`App` 使用该 workspace 的 `.data` 路径创建或切换：

- `ProjectManager(<workspace>/.data/projects.json)`
- `TerminalHistoryStore(<workspace>/.data)`
- `ShellSessionManager` 的历史存储引用

关闭或切换 workspace 前，先关闭所有运行时终端，再替换 workspace-scoped managers。`SettingsManager` 不随 workspace 切换，始终读取应用全局 `settings.json`。没有当前 workspace 时，项目 API 返回空状态或明确错误，settings API 仍可用，前端不应尝试启动终端。

应用启动时先完成旧数据迁移，再尝试恢复最近打开列表中的第一项。只有第一项仍可访问时才自动打开并绑定其 `.data`；如果第一项不可访问，则保持无 workspace 空态，并保留最近打开记录，不自动跳到第二项，避免打开用户未预期的 workspace。

替代方案是让每个 manager 支持动态 `SetPath`。不优先采用该方案，因为 manager 内部有锁和缓存边界，重建 manager 更直接，也更容易在测试中验证不同 workspace 的隔离。

### 3. 原生菜单触发动作，最近打开用应用内选择面板

Wails v2 支持 `options.App.Menu` 和 `menu.AddText` 回调。文件菜单提供：

- `打开项目`
- `最近打开`
- `清理最近打开`
- `关闭`

`打开项目` 可直接调用后端目录选择并打开 workspace，然后通过事件通知前端刷新。`清理最近打开` 和 `关闭` 同理。`最近打开` 不做固定 native submenu，而是触发前端显示应用内最近列表面板；这样列表可以动态更新、展示完整路径、处理不可用路径并复用现有应用状态更新流程。

### 4. Workspace 状态进入前端主状态流

前端增加 workspace state：

- `currentWorkspace`
- `recentWorkspaces`
- `workspaceRequired` 或等价空态判断

打开、关闭、从最近打开选择后，前端刷新项目库、TODO、终端和 Git 状态；终端设置保持全局，不因 workspace 切换重置。没有 workspace 时：

- 左侧 TODO/项目区域显示无 workspace 空态。
- 终端区域显示选择或打开 workspace 的空态。
- 创建 TODO、导入工程、创建终端、Git 状态查询等 workspace 依赖操作禁用或后端拒绝。

### 5. 旧全局数据迁移为一个 app-managed legacy workspace

升级时如果发现旧的全局 `projects.json` 或 `terminal-history.json`，并且尚未迁移到 workspace 模型，系统创建应用配置目录下的 `legacy-workspace/.data`，复制旧 workspace 数据文件，并把该 legacy workspace 加入最近打开记录。旧文件不删除。旧全局 `settings.json` 保留在应用配置目录，继续作为全局设置使用。

这个策略避免在用户尚未选择 workspace 时强行把旧数据写入任意目录，也避免启动时继续把全局 `projects.json` 当作当前工作内容。用户可以从最近打开中打开 legacy workspace，再按需要迁移或重新选择工作目录。

## Risks / Trade-offs

- [Risk] 用户可能不理解 workspace 和导入工程的区别。 -> Mitigation: 文件菜单和空态使用“打开项目/workspace”语义，项目库继续展示导入工程路径，避免把两层对象放在同一列表。
- [Risk] 切换 workspace 时仍有运行中 PTY。 -> Mitigation: 后端在关闭/切换前统一 shutdown runtime terminals，并清空前端 xterm session、Git 状态和 agent 状态。
- [Risk] `.data` 目录写入失败会导致打开 workspace 部分成功。 -> Mitigation: 打开流程先创建/校验 `.data` 和必要文件路径，失败时保持原 workspace 不变并返回错误。
- [Risk] 旧数据迁移生成的 legacy workspace 不在用户业务目录下。 -> Mitigation: 仅作为兼容入口，记录在最近打开中，不覆盖旧文件；后续可以增加导出/迁移到当前 workspace 的能力。
- [Risk] 最近打开记录包含不可访问路径。 -> Mitigation: 最近列表展示路径，选择不可访问路径时报告错误并保持当前 workspace 不变；用户可清空最近打开。

## Migration Plan

1. 保留现有 `tui-helper` 到 `todoai` 应用配置目录迁移。
2. 在 `todoai` 配置目录中引入 `recent-workspaces.json`。
3. 首次运行新版本时，如果发现未迁移的全局项目或历史文件，则复制到 `<app-config>/legacy-workspace/.data`，并把 `<app-config>/legacy-workspace` 写入最近打开；全局 `settings.json` 保持原位。
4. 新建或打开的用户 workspace 只读写自身 `.data`。
5. 回滚时，旧版本仍可读取保留在应用配置目录的旧全局文件；新 workspace `.data` 会被旧版本忽略。

## Open Questions

- 本设计把“关闭”解释为关闭当前 workspace；如果没有当前 workspace，则关闭动作不可执行。若后续需要“从最近打开中移除某个 workspace”，应作为最近打开面板中的单独操作扩展。
