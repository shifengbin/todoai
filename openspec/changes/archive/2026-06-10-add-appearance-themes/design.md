## Context

当前应用的视觉系统由硬编码 CSS 颜色和硬编码 xterm theme 共同组成。应用外壳、侧栏、设置弹窗和状态栏主要是浅色；终端区域和 xterm 固定为深色且不参与应用外观主题。现有 `settings.json` 由 Go 侧 `SettingsManager` 维护，并通过 `LoadTerminalSettings` 暴露给 Vue 前端，已用于终端 shell 和 launch profiles。

主题切换会跨越 Go 设置模型、Wails 绑定、Vue 状态、CSS 样式变量、xterm 会话生命周期和测试，因此需要在实现前明确边界。

## Goals / Non-Goals

**Goals:**

- 提供 `light` 和 `dark` 两种应用级主题。
- 使用 `light` 作为默认值，兼容没有主题字段的旧设置文件。
- 将主题偏好持久化到现有 `settings.json`。
- 设置界面可选择主题，并在保存成功后立即应用。
- 主题覆盖应用外壳、侧栏、工作区、设置弹窗和状态栏等终端以外页面。
- 嵌入式终端内容区、终端占位层、终端右键菜单和 xterm 配色保持既有固定配色。

**Non-Goals:**

- 不支持 `system` 跟随系统主题。
- 不支持用户自定义颜色、导入导出主题或第三方主题包。
- 不支持切换嵌入式终端或 xterm 配色。
- 不改变终端 shell、launch profile、项目树或 Git 状态栏的业务行为。
- 不引入新的前端状态管理库或样式框架。

## Decisions

### 主题作为设置模型的一部分

在 `TerminalSettingsState` 和持久化设置中增加 `Theme string` 字段，允许值为 `light` 或 `dark`。缺失、空值或未知值统一规范化为 `light`。

选择复用现有设置模型，而不是新建独立外观配置文件，是因为当前应用已经通过同一个设置弹窗和 `settings.json` 管理本地偏好。这样能减少配置文件数量和加载顺序复杂度。

### 保存 API 使用明确的主题方法

在 Go 侧增加类似 `SaveTerminalTheme(theme string)` 的方法，并通过 Wails 暴露给前端。设置弹窗保存时先校验 shell 和 launch profiles，再保存主题；任何一步失败都保持弹窗打开并显示错误。

选择独立保存方法，而不是把所有设置合并成一个大 `SaveTerminalSettings`，是为了降低对现有 shell 和 launch profile 保存路径的影响，避免把现有测试和错误处理一次性重写。

### 前端以一个当前主题状态驱动 UI 和终端

`App.vue` 从 `LoadTerminalSettings` 获得主题，保存到响应式状态，并在应用根节点设置 `data-theme`。CSS 使用变量定义 Light/Dark 色板，各组件继续使用现有 class 结构。

选择 CSS variables，而不是复制一套 `.dark .selector` 覆盖样式，是因为当前 `style.css` 中颜色分布广，变量化后主题边界更清晰，也方便后续维护状态色。终端相关变量在 Light/Dark 下保持同值，确保页面主题不会间接改变终端区域。

### 终端配色保持固定

`xtermFactory.js` 继续使用固定 xterm theme，不接收应用外观主题。`TerminalSessionManager` 不保存应用主题，也不向已打开终端 session 下发主题更新。

选择将终端排除在 Light/Dark 主题之外，是为了保留终端程序、ANSI 色彩和既有终端交互的稳定性。应用主题只改变终端以外的页面外壳。

### 主题切换只在保存成功后生效

设置弹窗中的选择可以临时编辑，但全局主题在保存成功后更新。取消设置不改变当前应用主题。

选择保存后生效而不是预览式即时生效，是为了保持设置弹窗现有的保存语义：用户可以修改多项设置后统一确认，取消时不会留下部分 UI 状态。

## Risks / Trade-offs

- 颜色硬编码较多 → 通过第一步集中引入 CSS variables，并逐步替换应用内颜色，降低遗漏风险。
- 终端配色被 CSS 变量间接改动 → 在 terminal 相关变量上保持 Light/Dark 同值，并用 `xtermFactory.test.js` 和 `terminalManager.test.js` 覆盖不传递应用主题的行为。
- 设置保存分多次调用可能部分成功 → 前端按现有顺序处理错误；主题保存失败时弹窗保持打开，并显示错误。shell 或 launch profiles 已成功保存的情况沿用现有设置保存语义。
- Wails 绑定需要更新 → 实现时运行 Wails 生成或按项目现有绑定方式更新 `frontend/wailsjs/go/main/App.js`、`App.d.ts` 和 `models.ts`。
- 视觉回归难以完全由单元测试覆盖 → 单元测试覆盖状态与 API 行为，最终通过本地前端测试和构建验证样式无语法错误。
