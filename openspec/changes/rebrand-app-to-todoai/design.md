## Context

应用当前在不同层面使用 `TUI Helper` 和 `tui-helper`：Wails 配置、窗口标题、HTML title、Debian 包名、Linux desktop 文件、README、本地配置目录，以及嵌入式终端 command-state 协议。用户已确认展示名称、包名和二进制名都要改为 `TodoAI` / `todoai`，并且没有现成图标，需要生成新的启动图标。

这不是单纯文案替换。包名和配置目录会影响已安装用户的命令路径和本地数据；command-state 协议运行在 shell 临时脚本、后端输出过滤和前端 xterm OSC handler 之间，改名时需要兼容旧标识。

## Goals / Non-Goals

**Goals:**

- 统一用户可见应用名为 `TodoAI`。
- 统一发布身份、包名、二进制名和 Linux 命令名为 `todoai`。
- 生成并落地新的应用启动图标，覆盖 Wails 通用图标和 Windows `.ico`。
- 将默认本地数据目录迁移到 `todoai`，并保留旧 `tui-helper` 数据可用性。
- 将应用私有 command-state 协议迁移到 `todoai` 标识，同时继续消费旧 `tui-helper` payload。
- 更新自动化测试和文档，证明打包元数据、数据迁移和协议兼容行为正确。

**Non-Goals:**

- 不重新设计主界面布局或 TODO 工作流。
- 不改变终端命令生命周期语义、agent 状态优先级或终端历史格式。
- 不提供旧 `tui-helper` 二进制名的长期 alias 或兼容安装包。
- 不引入新的图标生成运行时依赖；图标作为静态构建资产提交。

## Decisions

1. **展示名和发布名分离但统一映射。**
   - 决策：用户可见名称使用 `TodoAI`，包名、二进制名、desktop 文件名和图标资源名使用小写 `todoai`。
   - 理由：Linux 包名和命令名更适合小写无空格，窗口和启动器需要保留产品大小写。
   - 备选：全部使用 `TodoAI`。缺点是命令名和包名不符合常见 Linux 命名习惯。

2. **本地数据目录迁移采用“新目录优先，旧目录 fallback/搬迁”。**
   - 决策：默认路径改为用户配置目录下的 `todoai`。如果新路径不存在但旧 `tui-helper` 路径存在，启动时迁移旧目录内容到新目录；如果新旧都存在，优先使用新目录，避免覆盖用户新数据。
   - 理由：用户升级后不应丢失项目、TODO、设置和终端历史；同时要避免自动覆盖新版本已产生的数据。
   - 备选：只改默认目录，不迁移。缺点是升级后应用看起来像全新安装。

3. **command-state 协议使用新标识并保留旧标识解析。**
   - 决策：新生成的 zsh/bash/PowerShell 集成输出 `777;todoai;...`，后端过滤器和前端 OSC handler 同时接受 `todoai` 与 `tui-helper`。
   - 理由：运行中的旧 shell 临时脚本、测试历史和可能残留的终端输出仍可能发出旧 payload；兼容解析可以降低改名造成的终端 UI 回归。
   - 备选：一次性替换为只接受 `todoai`。缺点是更容易让旧会话泄漏私有 payload。

4. **图标生成后作为普通静态资产处理。**
   - 决策：使用生成图作为源，落地 `build/appicon.png`，并生成 `build/windows/icon.ico`；其他平台继续按 Wails 现有构建机制从这些资源读取。
   - 理由：符合 Wails 项目结构，不需要在构建时重新生成图标。
   - 备选：保留旧图标或只替换 PNG。缺点是 Windows 构建可能继续使用旧 `.ico`。

## Risks / Trade-offs

- [Risk] 外部脚本仍调用 `tui-helper` 命令。→ 在 README 和 proposal 中明确这是包名/命令名 breaking change。
- [Risk] 数据迁移覆盖新目录已有数据。→ 只有新目录缺失且旧目录存在时迁移；新旧都存在时不覆盖。
- [Risk] command-state 标识迁移导致私有 payload 可见。→ 前端和后端都接受新旧标识，并用测试覆盖 raw OSC、Windows 文本 fallback、无效 payload 和 split payload。
- [Risk] 生成图标在小尺寸下不清晰。→ 使用居中、无文字、少细节的图标，并在实现时检查 1024 PNG 和 `.ico` 多尺寸输出。
- [Risk] Wails 构建输出名改变影响 Debian 测试 fixture。→ 更新假 `wails` fixture、路径断言、control 文件和 desktop 文件断言。

## Migration Plan

1. 先提交静态元数据、配置路径和协议兼容变更，并让单元测试覆盖迁移与 payload 兼容。
2. 再替换图标资产并更新文档。
3. 验证 `go test ./...`、前端测试/构建，以及 Debian 打包脚本测试。
4. 回滚时可恢复旧包名和标题；数据目录迁移应不删除旧目录，因此旧版本仍可读取原目录数据。

## Open Questions

- 是否需要为旧命令名 `tui-helper` 提供发行包级 alias 或迁移提示？本次默认不提供。
