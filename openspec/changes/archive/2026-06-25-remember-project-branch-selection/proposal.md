## Why

用户在同一 workspace 中为同一个项目连续创建多个 TODO 时，通常会反复选择相同的 base 分支。当前分支选择只保存到单个 TODO 工程副本中，下次创建 TODO 时不能复用上次选择，增加重复操作并容易选错分支。

## What Changes

- 在 workspace 级别记录每个项目上次成功保存的 base 分支选择。
- 创建 TODO、编辑 TODO 和为已有 TODO 添加项目时，项目分支控件优先使用 workspace 级上次选择作为默认值。
- 只有创建、编辑或添加项目保存成功后才更新记录；用户取消表单或仅在表单中临时修改分支不会更新记录。
- 保持全局项目候选库不记录 workspace 级分支偏好，避免不同 workspace 之间互相影响。
- 保持 TODO 工程副本继续保存当次选择的 base 分支，作为该 TODO 后续 worktree 创建和历史展示的事实数据。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `todo-workspace`: 创建 TODO、编辑 TODO 和添加项目时的项目分支默认值与保存后的 workspace 级分支偏好记录。

## Impact

- 后端 workspace 项目状态模型和 `projects.json` 持久化格式增加兼容字段。
- 创建 TODO、更新 TODO、添加项目到 TODO 的后端保存路径需要在成功保存后同步分支偏好。
- 前端 TODO 创建、TODO 详情和添加项目弹窗需要读取并使用 workspace 返回的项目分支偏好。
- 需要补充 Go 状态持久化测试和 Vue 表单默认值/提交后的行为测试。
