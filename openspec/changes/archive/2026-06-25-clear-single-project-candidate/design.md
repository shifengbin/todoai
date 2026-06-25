## Context

全局项目候选由 `projects` 状态承载，并在创建 TODO、编辑 TODO、向 TODO 添加工程的三个候选列表中复用。后端已经提供 `DeleteProject(projectID)`，该行为只移除候选记录，不删除 TODO 工程副本，也不关闭 TODO 工程终端。

当前前端只提供 `clearGlobalProjectCandidates()` 批量清空入口。候选列表项本身是一个选择按钮，点击后会把项目 ID 写入当前弹窗的 `projectSelections`。如果单项清除一个已选候选，但不清理当前弹窗选择，后续提交会因为 `project not found` 失败。

## Goals / Non-Goals

**Goals:**

- 在所有候选项目列表中提供单项清除入口。
- 清除前展示应用内自定义确认弹窗，确认后复用现有 `DeleteProject(projectID)` 移除候选记录。
- 清除后同步移除当前创建 TODO、编辑 TODO、添加工程弹窗中的临时选择。
- 保持清除候选的安全边界：不删除磁盘目录，不删除 TODO 工程副本，不关闭 TODO 工程终端。
- 用前端自动化测试覆盖用户可见行为和边界。

**Non-Goals:**

- 不新增候选项目数据模型字段。
- 不新增 Wails 后端 API。
- 不恢复侧栏中的项目库管理入口。
- 不提供批量勾选后删除候选的新交互；既有清空全部能力保持不变。

## Decisions

1. **复用现有 `DeleteProject(projectID)`，不新增 `ClearProjectCandidate` API。**

   该后端方法已经表达“删除单个全局候选记录”的语义，并已有测试保证不会关闭 TODO project 终端。新增 API 会增加 Wails 绑定和测试面，但不会带来新的业务能力。

2. **候选列表项拆为行容器，行内包含选择按钮和清除按钮。**

   现有候选项是一个 `<button>`，不能在内部嵌套另一个删除按钮。实现时应引入外层行容器，左侧按钮继续负责选择/取消选择候选，右侧 `Trash2` 图标按钮负责清除候选，并使用 `@click.stop` 避免触发选择。

3. **清除候选后统一清理当前弹窗选择状态。**

   清除函数在后端状态更新成功后，应从 `todoForm.projectSelections`、`todoDetail.projectSelections` 和 `projectPicker.projectSelections` 中移除对应项目 ID。这样即使用户清除了已选候选，也不会提交一个已不存在的候选 ID。

4. **候选清除不触发即时 Git 状态刷新。**

   批量清空候选已经使用 `refreshGitStatus: false`。单项清除也应保持一致，因为这个操作是候选库管理动作，不需要立即查询被移除候选或剩余候选的 Git 状态。

5. **单项候选清除使用应用内确认弹窗。**

   系统原生 `window.confirm` 样式不可控，和现有 Wails 桌面应用界面不一致。单项候选清除应使用与 Git 初始化确认弹窗类似的 overlay/dialog 结构，展示候选名称和路径，提供取消与清除动作。

## Risks / Trade-offs

- 清除按钮与选择按钮距离过近可能造成误点 -> 使用应用内确认弹窗，并让清除按钮有明确 title/aria-label。
- 拆分候选项 DOM 可能影响现有样式和测试选择器 -> 保留现有候选选择按钮的 `data-testid`，为清除按钮增加新的稳定 `data-testid`。
- 同步清理三个弹窗状态可能遗漏某个入口 -> 抽取小函数按 project ID 清理所有 `projectSelections`，并用三个入口测试覆盖。
- 删除全局候选后 TODO 详情仍需显示已加入 TODO 的工程 -> 保持使用 TODO project 快照补足展示的既有逻辑，并增加回归测试。
