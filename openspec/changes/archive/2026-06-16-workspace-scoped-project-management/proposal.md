## Why

当前 TODO、项目库和终端历史保存在应用全局配置目录，导致不同业务项目之间的数据混在一起。用户希望以“管理项目/workspace”为单位打开和关闭工作内容，并把该 workspace 相关数据保存到所选目录下的 `.data` 中，方便隔离、迁移和回到最近工作。终端 settings 作为应用级偏好继续全局保存。

## What Changes

- 新增 workspace 生命周期：用户通过文件菜单打开一个目录作为当前 workspace，系统记录该目录，并从该目录的 `.data` 加载项目库、TODO、终端历史和 workspace 级数据。
- 新增最近打开管理：应用全局保存最近打开的 workspace 列表，用户可从文件菜单查看并重新打开最近 workspace，也可清空最近打开记录。
- 新增关闭当前 workspace：关闭后清空当前 workspace 上下文和运行时终端，不删除 workspace 目录或 `.data` 数据。
- 调整项目库和 TODO 管理边界：项目库中的导入工程、TODO、TODO 与工程关联、选中状态和终端历史不再全局共享，而是属于当前打开的 workspace。
- 保持 settings 全局：终端 shell、启动配置和外观设置不随 workspace 切换，仍保存在应用全局配置中。
- 调整启动和空态行为：应用启动时不应无条件加载全局 `projects.json` 作为当前工作内容；没有打开 workspace 时，TODO、项目库和终端区域显示无 workspace 空态。
- 增加原生文件菜单：`文件` 下提供 `打开项目`、`最近打开`、`清理最近打开`、`关闭` 四个选项。
- **BREAKING**: 持久化位置从应用全局配置目录迁移到 workspace 目录下 `.data`。旧全局数据需要兼容迁移或导入策略，避免升级后已有数据不可见。

## Capabilities

### New Capabilities

- `workspace-lifecycle`: 管理 workspace 的打开、关闭、最近打开记录、清理最近打开记录和无 workspace 状态。

### Modified Capabilities

- `project-workspace`: 项目库中的导入工程列表改为当前 workspace 级持久化，打开/移除工程只影响当前 workspace。
- `todo-workspace`: TODO、TODO 状态、TODO 与工程关联和选中上下文改为当前 workspace 级持久化。
- `embedded-shell-sessions`: 终端记录和最近输出历史改为当前 workspace 级持久化，关闭或切换 workspace 时需要清理运行时终端上下文。
- `terminal-settings`: 终端 shell、启动配置和外观设置明确为应用全局偏好，不随 workspace 切换。
- `application-identity`: 旧应用全局数据迁移行为需要扩展到新的 workspace 数据模型，避免升级后丢失既有项目和 TODO。

## Impact

- Go 后端需要新增 workspace/recent 管理器，并让 `ProjectManager`、`TerminalHistoryStore` 的存储路径随当前 workspace 切换；`SettingsManager` 保持全局。
- Wails 入口需要增加原生菜单，并把菜单点击事件连接到打开目录、最近打开、清理最近打开和关闭 workspace 的后端/前端流程。
- Vue 前端需要接收 workspace 状态，渲染无 workspace 空态、最近打开选择界面，并在 workspace 切换后刷新 TODO、项目库、终端和设置状态。
- 现有 Wails 生成绑定需要新增 workspace API 后重新生成。
- 测试需要覆盖 workspace 数据隔离、最近打开、关闭/切换清理运行时终端、旧数据迁移和无 workspace 空态。
