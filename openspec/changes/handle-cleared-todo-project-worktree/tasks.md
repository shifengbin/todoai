## 1. 后端状态与检测

- [x] 1.1 在 Go 模型中新增 `WorktreeStatusCleared = "cleared"`，并确保现有 JSON/Wails 模型能向前端暴露该状态。
- [x] 1.2 抽取 TODO project worktree 清除检测逻辑：ready worktree 路径缺失或保存的 worktree 分支确认不存在时，持久记录为 cleared，且不覆盖 failed 状态。
- [x] 1.3 更新 `GetTodoProjectGitStatus`，在分支刷新时识别 worktree 已清除并返回前端可立即展示的清除信号。
- [x] 1.4 抽取项目终端工作目录解析 helper：ready 使用 worktree 路径，cleared 使用原项目路径，pending/failed/路径不可用返回错误。
- [x] 1.5 更新 `CreateTodoTerminal` 和 `StartTodoProjectBackgroundCommand` 使用统一工作目录解析，保持 not-started、failed 和项目路径不可用场景的拒绝行为。

## 2. 前端展示与交互

- [x] 2.1 更新 `App.vue` 的 TODO project 分支刷新缓存，支持从 Git 状态结果或 TODO project 状态识别 worktree 已清除。
- [x] 2.2 更新 `ProjectSidebar.vue` 的项目名显示逻辑，在分支括号位置显示 `worktree已清除`，并保留 ready worktree 的实时分支显示。
- [x] 2.3 更新项目终端启动按钮和菜单状态：cleared 且原项目路径可用时可启动，failed、not-started 或项目路径不可用时不可启动。
- [x] 2.4 如后端模型或 Wails API 字段变化，重新生成并检查 `frontend/wailsjs/go/*` 绑定文件。

## 3. 自动化测试

- [x] 3.1 增加 Go 测试：ready worktree 路径缺失时记录 cleared，且创建项目终端 cwd 回退到原项目目录。
- [x] 3.2 增加 Go 测试：cleared worktree 的后台启动配置 cwd 回退到原项目目录，failed/not-started 场景仍拒绝。
- [x] 3.3 增加前端测试：项目行显示 `frontend-app(worktree已清除)`，不显示旧 worktree 分支。
- [x] 3.4 增加前端测试：cleared 项目行显示终端启动菜单，failed 或不可用项目行不暴露可用启动入口。
- [x] 3.5 回归现有分支刷新测试，确认 ready worktree 仍显示实时 Git 分支且命令结束后刷新。

## 4. 质量检查

- [x] 4.1 运行 `go test ./...`，修复后端测试失败。
- [x] 4.2 运行 `cd frontend && npm test`，修复客户端自动化测试失败。
- [x] 4.3 运行 OpenSpec 状态或校验命令，确认 change 仍满足 apply 条件。
- [x] 4.4 执行自动 review，重点检查 cleared 与 failed 语义边界、cwd fallback 是否只在明确 cleared 状态生效、以及前端状态缓存是否会误显示清除标记。

## 5. 打包验证

- [x] 5.1 运行 `wails build -tags webkit2_41`，生成可执行文件。
