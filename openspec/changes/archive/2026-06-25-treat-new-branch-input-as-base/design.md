## Context

TODO 项目分支输入框当前提交的是 `projects[].baseBranch`，前端允许用户选择已存在分支，也允许手动输入不存在分支。后端 `GitWorktreeService.PrepareWorktree` 会在 TODO 进入执行状态时根据该字段准备项目 worktree。

现有后端行为有两条路径：已存在分支会保存为 base 分支，并派生 `todo-workspace/...` 隔离 worktree 分支；不存在分支会直接从主分支创建并作为 worktree 分支使用，同时把 base 分支记录为主分支。这导致同一个输入框在不同情况下具有不同含义。

本变更把输入框语义收敛为一条规则：用户输入或选择的分支只表示 base 分支。worktree 分支始终由系统生成，避免用户输入分支直接承担 TODO worktree 隔离职责。

## Goals / Non-Goals

**Goals:**

- 用户输入不存在分支时，系统先基于主分支创建该分支，并将它保存为 TODO 项目的 base 分支。
- 无论 base 分支原本是否存在，TODO 项目都使用系统生成的隔离 worktree 分支。
- 保持前端 API、持久化字段和已存在分支行为兼容。
- 用后端测试覆盖 Git 命令顺序、保存的 base 分支和保存的 worktree 分支。

**Non-Goals:**

- 不修改分支输入控件、候选列表或前端 payload。
- 不自动 fetch 远端分支。
- 不改变主分支解析策略，继续使用当前 `DefaultBranch` 规则。
- 不迁移已有 TODO 项目记录。

## Decisions

### 先创建缺失的 base 分支，再复用隔离 worktree 流程

选择：当 `requestedBranch` 不存在时，先执行等价于 `git branch <requestedBranch> <defaultBranch>` 的创建步骤。创建成功后，把该分支作为 base 分支继续走与已存在分支相同的隔离 worktree 分支流程。

原因：这让 `requestedBranch` 在所有场景下都只表示 base 分支，`worktreeBranchName(projectName, taskWorkspaceDir)` 在所有成功 worktree 准备场景下都负责生成当前 worktree 分支。

备选方案：继续使用 `git worktree add -b <requestedBranch> ... <defaultBranch>` 并把 base 分支改写为 `requestedBranch`。该方案仍会让用户输入分支成为 worktree 分支，不能满足“输入框里的分支只能作为 base 分支”的要求。

### 保持 `PrepareWorktree` 的结果结构不变

选择：不新增字段。成功时 `WorktreePrepareResult.BaseBranch` 返回实际 base 分支，`WorktreePrepareResult.WorktreeBranch` 返回系统生成的隔离 worktree 分支。

原因：`RecordTodoWorkspace` 已按这两个字段更新 `TodoProject`，前端和 README 也都依赖现有字段名。修改字段会扩大协议和迁移范围，但无法增加必要表达能力。

备选方案：增加 `CreatedBaseBranch` 或类似标记。当前需求不需要 UI 展示“是否新建 base 分支”，测试可通过 Git 命令和结果字段覆盖，暂不增加模型复杂度。

### 缺失 base 分支创建失败时记录准备失败

选择：如果创建用户输入的 base 分支失败，`PrepareWorktree` 返回 `WorktreeStatusFailed`，错误信息明确指向 base 分支创建失败，并且不继续创建 worktree。

原因：base 分支是后续隔离 worktree 的起点。失败后继续创建 worktree 会隐藏真实原因，也可能从错误起点创建分支。

备选方案：回退到主分支创建 worktree。该方案会违背用户输入分支作为 base 分支的语义，并产生难以解释的保存结果。

## Risks / Trade-offs

- [Risk] 用户输入的分支名不合法时会在创建 base 分支阶段失败。-> Mitigation: 复用 Git 返回错误并记录为 worktree 准备失败，前端不需要提前实现完整 Git ref 校验。
- [Risk] 新建 base 分支后，后续隔离 worktree 分支创建失败会留下已创建的 base 分支。-> Mitigation: 这是显式用户输入的 base 分支，保留比自动删除更可预期；失败信息会记录到 TODO 项目。
- [Risk] 已有测试和规格期望旧语义。-> Mitigation: 修改规格中的完整要求块，并更新后端测试使新语义成为回归保护。

## Migration Plan

无需数据迁移。已有 TODO 项目记录保持当前保存值。变更生效后，新的 worktree 准备流程会按新语义保存 base 分支和 worktree 分支。

回滚时恢复 `PrepareWorktree` 中不存在分支的旧路径，并恢复对应规格和测试期望。

## Open Questions

无。
