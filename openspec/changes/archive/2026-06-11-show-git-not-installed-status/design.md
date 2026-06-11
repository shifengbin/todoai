## Context

项目当前通过后端执行 `git -C <path> status --porcelain=v2 --branch` 获取当前项目 Git 状态，并通过 `GitStatus` Wails 模型传给 Vue 状态栏。当前错误分支只区分非 Git 仓库、路径不可用、超时和普通命令失败；当系统未安装 `git` 时，前端只能显示通用的 Git 状态不可用。

状态栏还会在可用项目且非 Git 仓库时显示 `Initialize Git Repository` 操作。若 `git` 命令不存在，初始化入口即使出现也无法成功执行，因此缺少一个明确的“工具链不可用”状态。

## Goals / Non-Goals

**Goals:**

- 在执行 `git status` 和 `git init` 前校验 `git` 命令是否存在。
- 用结构化状态表示 Git 未安装或不可执行，而不是把它混入普通查询失败。
- 状态栏在 Git 缺失时显示 `未安装 Git`，并隐藏初始化仓库操作。
- 保留现有的非 Git 仓库、项目路径不可用、查询失败、初始化失败语义。
- 让命令可用性检查可被单元测试覆盖，不依赖测试环境真实安装状态。

**Non-Goals:**

- 不提供安装 Git 的引导、下载链接或自动安装能力。
- 不改变 Git 状态刷新时机、去重策略或 Windows 后台命令隐藏策略。
- 不引入新的外部依赖。
- 不改变终端会话、TODO 工作区或项目导入行为。

## Decisions

1. **在后端查询入口做 Git 可用性检查。**

   选择：`queryGitStatus` / `initializeGitRepository` 调用底层命令前，先通过 `exec.LookPath("git")` 或等价封装检查命令是否存在。为测试保留注入点，例如增加带 checker 参数的内部函数。

   原因：命令存在性是运行环境事实，后端最接近实际执行点，也能避免前端误判 PATH。前端只消费结构化状态，不负责探测本机命令。

   备选：继续执行命令并解析 `exec: "git": executable file not found`。该方案依赖平台错误文本，且仍然违反“执行前先校验”的要求。

2. **用 `GitStatus` 承载 Git 缺失状态。**

   选择：在 `GitStatus` 增加布尔字段，例如 `GitUnavailable` / `gitUnavailable`。当项目路径可用但 Git 命令不存在时，`GetProjectGitStatus` 返回带该字段的状态对象，不把它作为普通错误抛给前端。

   原因：前端状态栏已经围绕 `GitStatus` 展示路径不可用、非仓库和仓库状态。Git 缺失也是可展示的环境状态，不应触发全局错误。

   备选：继续通过 rejected promise 和错误字符串让前端识别。该方案需要前端匹配错误文本，且会和普通 Git 查询失败混在一起。

3. **初始化入口依赖 Git 可用状态。**

   选择：`showInitializeGitRepository` 在 Git 缺失时返回 false。即使前端因为旧状态或竞态调用初始化 API，后端也在执行 `git init` 前返回明确错误。

   原因：按钮代表用户可执行操作。Git 缺失时展示初始化入口会产生必然失败的动作；后端检查则保证 API 边界仍然安全。

   备选：显示按钮并在点击后报错。该方案暴露无效操作，不如直接在状态栏展示根因。

4. **状态栏优先级明确化。**

   选择：状态优先级为无项目、路径不可用、Git 未安装、加载中、普通错误、非 Git 仓库、正常仓库状态。路径不可用不需要探测 Git；Git 未安装优先于非 Git 仓库，因为无法可靠判断仓库状态。

   原因：优先级避免多个状态同时满足时出现不稳定 UI，也保证用户先看到最可行动的环境问题。

## Risks / Trade-offs

- [Risk] 在测试或打包环境中 PATH 与用户 shell PATH 不一致，`LookPath("git")` 可能与用户期望不同。→ Mitigation: 使用应用进程实际 PATH，与后端命令执行环境保持一致；不额外模拟 shell 初始化。
- [Risk] 新增 `GitStatus` 字段需要同步 Wails 前端绑定。→ Mitigation: 实现后运行项目既有绑定生成或手动更新生成文件，并用前端测试覆盖字段消费。
- [Risk] Git 缺失状态如果作为错误返回，会继续被前端显示为通用失败。→ Mitigation: 后端状态查询返回结构化 `GitStatus`，初始化 API 才返回明确错误。
