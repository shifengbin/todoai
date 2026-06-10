## Context

应用是 Wails 桌面应用，前端使用 Vue 展示项目、终端和底部状态栏，后端使用 Go 管理项目列表、终端会话和 Git 状态查询。当前 Git 状态栏由 `App.vue` 中的单个文本 computed 渲染，后端已有 `GetProjectGitStatus(projectID)` 能返回分支、改动数量、暂存、未暂存、未跟踪、ahead/behind 以及 `isRepo=false` 状态。

现有实现缺少两类能力：第一，状态栏没有把各类 Git 信息拆成可扫描的信息块；第二，当前项目不是 Git 仓库时，只能提示用户，没有可直接执行的初始化动作。

## Goals / Non-Goals

**Goals:**

- 将底部 Git 状态栏拆成多个可独立着色的圆角信息块。
- 在 Git 仓库状态下展示分支、总改动数，并在有值时展示 staged、unstaged、untracked、ahead、behind。
- 在非 Git 仓库状态下展示初始化按钮，点击后直接在当前项目目录执行 `git init`。
- 初始化成功后刷新当前项目 Git 状态，失败时显示非阻断错误。
- 保持状态栏固定高度，不影响终端区域布局。

**Non-Goals:**

- 不添加提交、暂存、拉取、推送等 Git 操作。
- 不添加 Git 仓库选择、远程配置或默认分支配置。
- 不把初始化命令写入或发送到当前终端会话。
- 不引入新的前端组件库或后端 Git 依赖库。

## Decisions

1. **用结构化状态片段替代单段状态文本。**

   前端将现有 `projectGitStatusText` 拆成适合模板渲染的数据结构，例如状态 chip 列表、按钮可见性和 loading/error 状态。这样可以直接渲染不同颜色的圆角信息块，也能继续在测试中断言可见文本。

   备选方案是继续拼接字符串并通过 CSS 或分隔符模拟信息块。该方案实现较快，但状态分支和样式绑定会混在一起，后续扩展 staged/untracked 等字段会更脆弱。

2. **后端新增 Wails 方法执行 Git 初始化。**

   新增 `InitializeProjectGitRepository(projectID)`，先通过项目管理器查找项目并确认路径可用，再执行 `git -C <project path> init`。方法成功后返回最新 Git 状态或由前端继续调用现有 `GetProjectGitStatus` 刷新。

   备选方案是向当前终端发送 `git init`。该方案依赖 active terminal、shell 运行状态和输入焦点，不适合作为状态栏按钮的可靠行为，也更难测试。

3. **Git 命令封装保持轻量。**

   继续使用 `os/exec` 调用系统 Git，不引入 Git SDK。初始化方法可复用类似当前 `gitStatusRunner` 的 runner 形态，以便单元测试覆盖成功和失败路径。

4. **按钮直接执行，不弹确认。**

   用户已确认 `Initialize Git Repository` 按钮点击后直接执行。执行中按钮禁用并显示初始化中状态，防止重复点击；成功后刷新状态栏；失败时复用当前全局错误展示。

## Risks / Trade-offs

- **Git 不存在或执行失败** -> 后端返回明确错误，前端显示非阻断错误并保持非仓库状态可见。
- **项目路径不可用** -> 后端在初始化前校验项目可用性并返回错误；前端不在路径不可用状态显示初始化按钮。
- **重复点击造成并发初始化** -> 前端用 loading 状态禁用按钮；后端 `git init` 本身对已初始化目录通常幂等，但仍以按钮禁用减少重复请求。
- **状态栏信息过多导致窄屏拥挤** -> 状态栏使用可换行或横向压缩的 chip 容器，并对长分支名使用省略或截断，保持固定高度和终端区域稳定。

## Migration Plan

该变更不需要数据迁移。实现后需要重新生成 Wails 前端绑定，使新增后端方法能被 Vue 调用。回滚时删除新增方法、绑定和前端按钮逻辑即可，现有 Git 状态查询数据结构保持兼容。

## Open Questions

无。按钮点击行为已确定为直接执行 `git init`。
