## Why

当前 TODO 项目分支输入框允许用户输入不存在的分支，但 worktree 准备逻辑会把这个输入分支直接作为 worktree 分支保存和检出。这与分支字段“只表示 base 分支”的产品语义不一致，也会让用户输入的新 base 分支缺少后续隔离 worktree 分支。

本变更统一分支选择和手动输入的语义：输入框中的分支永远只作为 TODO 项目的 base 分支；如果该 base 分支不存在，系统先基于主分支创建它，再基于该 base 分支创建隔离 worktree 分支。

## What Changes

- 修改不存在分支的 worktree 准备规则：先从主分支创建用户输入的 base 分支，再按已存在 base 分支流程创建隔离 worktree 分支。
- 保持已存在分支流程不变：用户选择或输入已存在分支时，该分支保存为 base 分支，并创建 `todo-workspace/...` 隔离 worktree 分支。
- 保持前端提交 API 和 payload 不变：`projects[].baseBranch` 继续传递输入框中的分支字符串。
- 更新 TODO worktree 规格和回归测试，明确手动输入的新分支不能直接作为 worktree 分支。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `todo-worktree-workspaces`: 修改“输入不存在分支”时的 worktree 创建要求，使输入分支作为 base 分支保存，并为 TODO 项目创建隔离 worktree 分支。

## Impact

- 后端：`GitWorktreeService.PrepareWorktree` 中不存在分支的处理流程。
- 数据：`TodoProject.BaseBranch` 保存用户输入的新 base 分支；`TodoProject.WorktreeBranch` 保存隔离 worktree 分支。无需数据迁移。
- 前端：分支输入控件和提交 payload 保持兼容，预计无需修改。
- 测试：更新 `git_worktree_test.go` 中不存在分支场景，并补充或调整规格相关断言。
