## Context

当前创建 TODO、编辑 TODO 和为 TODO 添加项目时，前端会为每个选中项目提交 `baseBranch`，后端将该值保存到对应的 TODO 工程副本中。该数据只属于单个 TODO 工程副本，用于后续创建 worktree、生成 README 和展示完成快照。

全局项目候选已经从 workspace TODO 数据中分离，候选项目跨 workspace 共享；TODO 和 TODO 工程副本保存在当前 workspace 的 `.data/projects.json` 中。因此分支偏好必须保存在 workspace 项目状态中，而不是写入全局候选项目库。

## Goals / Non-Goals

**Goals:**

- 为每个 workspace 记录每个项目上次成功保存的 base 分支选择。
- 创建 TODO、编辑 TODO 中新增项目、为已有 TODO 添加项目时，默认使用该 workspace 中该项目的上次选择。
- 只有保存成功后更新偏好，取消表单或未提交的输入不影响下次默认值。
- 保持 TODO 工程副本中的 `baseBranch` 仍是该 TODO 的事实数据。
- 兼容旧 workspace 数据，缺少偏好字段时正常加载。

**Non-Goals:**

- 不改变全局项目候选库的数据归属，不让分支偏好跨 workspace 共享。
- 不新增 Git 分支探测或分支有效性校验规则。
- 不自动修改已有 TODO 工程副本的 `baseBranch`。
- 不改变 worktree 创建、README 生成或 completed TODO 快照的语义。

## Decisions

### 在 `ProjectState` 中新增 workspace 级偏好

在后端 `ProjectState` 增加 workspace-local 字段，例如：

```go
ProjectBranchPreferences map[string]ProjectBranchPreference `json:"projectBranchPreferences,omitempty"`

type ProjectBranchPreference struct {
    BaseBranch string `json:"baseBranch"`
}
```

key 使用全局候选的 `projectId`。该字段随 workspace 的 `projects.json` 保存，不写入应用级全局候选文件。

替代方案是把字段加到 `Project`。这会污染全局候选项目，因为当前项目候选跨 workspace 共享；不同 workspace 选择不同 base 分支时会互相覆盖。另一个替代方案是前端 localStorage，但它会绕过后端事实状态，且换设备、清缓存或后端测试时不可见。

### 以 map entry 存在性区分“无记录”和“明确空值”

偏好记录需要允许空分支值。如果用户清空某项目分支并保存，下一次选择该项目时应默认空值，而不是继续回退到当前 Git 分支。

因此前端默认值解析应检查 `projectBranchPreferences` 是否包含该 `projectId`：

```text
if preference map contains projectId:
  use preference.baseBranch
else:
  use current active git branch fallback
```

不能用 `preference?.baseBranch || fallback` 这类逻辑，否则明确保存的空字符串会被误判为无记录。

### 后端在保存成功路径更新偏好

偏好更新应放在 `ProjectManager` 的成功保存流程中：

- `CreateTodo`：创建 TODO 和 TODO 工程副本成功后，根据提交的项目选择更新偏好。
- `UpdateTodo`：TODO 编辑保存成功后，根据最终保存的项目选择更新偏好。
- `AssociateProjectSelectionsWithTodo` / 相关添加项目路径：项目成功关联到 TODO 后，根据提交的项目选择更新偏好。

这些路径均应先通过现有校验，例如项目存在、TODO 存在、重复项目处理等。只有状态最终成功持久化后，偏好才视为更新完成。若保存失败，返回错误且不留下部分写入。

### 前端默认值优先级

前端维护当前 `ProjectState.projectBranchPreferences`。当用户在创建 TODO 或添加项目弹窗中新选项目时，默认分支优先级为：

1. 当前 workspace 中该 `projectId` 的偏好记录。
2. 现有逻辑可提供的当前 active project Git 分支。
3. 空字符串。

编辑已有 TODO 时，已关联项目仍应显示该 TODO 工程副本自己的 `baseBranch`，因为用户是在编辑已保存事实数据；只有新增到详情里的项目使用偏好默认值。

### Wails 模型生成随 Go 类型更新

后端 `ProjectState` 增加字段后，需要更新 Wails 生成的前端模型，使 TypeScript 侧能读取 `projectBranchPreferences`。如果项目当前没有自动生成步骤纳入测试，应手动确认生成文件变更与 Go JSON 字段一致。

## Risks / Trade-offs

- [Risk] 以 `projectId` 为 key 时，删除再重新导入同一路径可能得到新 ID，旧偏好不会自动复用。→ 当前选择和 TODO 关联都以项目 ID 为主，先保持一致；如未来需要路径级迁移，可单独设计。
- [Risk] 空字符串偏好容易被前端通用 truthy 判断误处理。→ 测试覆盖“保存空分支后下次默认空值，不回退到 Git 分支”。
- [Risk] 创建/编辑/添加项目有多条入口，遗漏某条会导致偏好不一致。→ 把更新逻辑封装为后端 helper，并覆盖三条保存路径。
- [Risk] 旧数据缺少字段。→ 加载时将 nil map 视为空偏好；只有保存新选择后写入字段。

## Migration Plan

旧 workspace 的 `projects.json` 缺少 `projectBranchPreferences` 时无需迁移脚本。加载后按空偏好处理；首次成功保存项目分支选择时写入新字段。回滚到旧版本时，该字段会作为未知 JSON 字段被忽略，不影响已有 TODO 和 TODO 工程副本。

## Open Questions

无。
