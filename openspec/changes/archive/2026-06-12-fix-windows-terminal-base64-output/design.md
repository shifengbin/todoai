## Context

当前 shell 集成通过 OSC 777 传递命令生命周期：zsh/bash 脚本使用 `printf '\033]777;tui-helper;...\a'`，PowerShell 临时脚本使用 `[Console]::Out.Write()` 输出同样的 payload。前端 xterm.js 注册 `OSC 777` handler，解码 base64 命令并更新 terminal command label。

Windows 路径多了一层 ConPTY。ConPTY 可能不把私有 OSC 777 原样透传给 xterm parser，导致 `777;tui-helper;command-start;<base64>` 或其中的 base64 片段作为普通文本出现在终端里。后端 `readOutput` 目前在任何解析前就把原始输出发给前端并追加到 terminal history，因此一旦泄漏，历史回放也会重复显示。

## Goals / Non-Goals

**Goals:**

- Windows ConPTY 下应用内部 command-state payload 不显示给用户。
- command-state payload 不写入持久化 terminal history。
- 能从有效 command-state payload 中继续产生 `command-start` / `command-end` UI 状态，保持 PowerShell 命令标签体验。
- 保留现有 zsh/bash 集成、launch profile 应用侧标签 fallback、cmd fallback 行为。
- 对 chunk 边界、无效 base64、未知 OSC payload 做可预测处理。

**Non-Goals:**

- 不实现 iTerm2/Kitty/Sixel 图片协议或 codex/claude 的全部高级终端能力。
- 不替换 ConPTY 后端，不引入新的终端依赖。
- 不改变用户保存的 terminal shell 或 launch profile 配置格式。
- 不承诺所有第三方 OSC payload 都能被解析；本变更只处理应用私有 `tui-helper` command-state payload。

## Decisions

### 后端作为 command-state payload 的主解析边界

在 `ShellSessionManager.readOutput` 和 history 追加之间增加每 session 的流式 command-state 过滤器。过滤器识别应用私有 payload，输出两类结果：

- cleaned terminal data：发给前端 terminal renderer，并写入 terminal history。
- command-state events：通过新的或扩展的 Wails runtime event 传给前端，由现有 `handleTerminalCommandState` 更新 command label。

这样可以在同一边界解决显示污染和历史污染。备选方案是只依赖 xterm.js `registerOscHandler`；它无法阻止后端历史保存原始 payload，也无法覆盖 ConPTY 已经把 payload 变成普通文本的情况。

### 解析范围保持窄而明确

过滤器只消费应用私有协议：

- raw OSC：`ESC ] 777 ; tui-helper ; command-start ; <base64> BEL`
- raw OSC：`ESC ] 777 ; tui-helper ; command-end BEL`
- 等价 ST 终止形式：`ESC \`
- Windows ConPTY 可见文本 fallback：严格匹配 `777;tui-helper;command-start;<base64>` 或 `777;tui-helper;command-end`

未知 OSC、普通文本、第三方协议不做删除。备选方案是泛化过滤所有 base64-looking 文本；这会误删正常命令输出，风险不可接受。

### 流式处理 chunk 边界

PTY/ConPTY read buffer 可能把 OSC payload 拆成多段。过滤器需要为每个 session 保存少量 pending buffer，直到看到 BEL、ST、换行或达到上限。超过上限或无法识别为应用私有 payload 时，应释放为普通输出，避免终端卡住或吞掉真实内容。

备选方案是在每次 read 的字符串内做正则替换。实现简单，但遇到分块 payload 时会漏删，正是当前问题最容易复现的路径之一。

### 前端保留兼容解析路径

前端继续保留 xterm.js OSC 777 handler，作为历史数据、测试 fixture 或未经过后端过滤路径的兼容处理。正常 runtime 输出应优先使用后端提取出的 command-state event；前端写入 terminal 的 data 应已经是 cleaned data。

备选方案是完全移除前端 OSC handler。这样会减少重复逻辑，但会让非标准启动路径或旧历史回放的兼容性变差。

### PowerShell 集成安全降级

PowerShell 临时脚本可以继续发出 command-state payload，但有效性由后端过滤器判定。有效 payload 产生 command-state event；无效或无法完整解析的 payload 不得显示给用户，也不得覆盖已有 launch profile 标签。cmd.exe 仍不要求 lifecycle hook，继续依赖应用侧 launch profile 标签。

备选方案是 Windows 下彻底禁用 PowerShell command-state 集成。这样能避免泄漏，但会回退手动命令的标签体验；在可以安全过滤和转发事件的前提下，保留集成价值更高。

## Risks / Trade-offs

- [Risk] 流式过滤器误吞普通输出。 -> 只匹配 `777;tui-helper` 私有协议，并设置 pending 上限；未知内容释放为普通输出。
- [Risk] 后端和前端同时解析造成重复 command-state。 -> runtime 主路径由后端发 command-state event，cleaned data 不再包含已消费 payload；前端 handler只处理未被后端消费的兼容路径。
- [Risk] ConPTY 泄漏文本形态和预期不同。 -> 先覆盖 raw OSC 和明确可见 fallback，并保留 Windows 手测任务记录实际输出形态。
- [Risk] 历史中已有旧 payload。 -> 本变更保证新输出不再污染历史；清理历史存量不纳入范围。
- [Risk] 无效 base64 command-start。 -> 忽略 command-state 更新，但仍过滤私有 payload，避免把内部协议显示给用户。

## Migration Plan

- 不需要数据迁移或配置迁移。
- 新 runtime 输出经过过滤后写入 history；已有 history 中的旧泄漏内容不会自动重写。
- 如果需要回滚，可恢复为只由 xterm handler 解析 OSC 777；回滚风险是 Windows 下 payload 可能再次可见。

## Open Questions

- Windows ConPTY 实际泄漏文本是否包含 ESC/BEL，还是只包含 payload 文本，需要在实现阶段用 Windows 手测确认并补充 fixture。
- 是否把 command-state event 设计为独立 Wails event，还是扩展 `TerminalOutputEvent` 携带 events 数组；实现阶段应选择改动更小、测试更清楚的方式。
