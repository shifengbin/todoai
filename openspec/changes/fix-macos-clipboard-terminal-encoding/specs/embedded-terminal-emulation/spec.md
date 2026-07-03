## MODIFIED Requirements

### Requirement: Start Shell With Xterm-Compatible Environment

The system SHALL start embedded PTY-backed shell processes with terminal capability variables that match the xterm.js renderer and a UTF-8 text locale suitable for Unicode terminal input and output. The system MUST preserve existing UTF-8 locale values and SHALL only fill or replace locale values that are missing or not UTF-8.

#### Scenario: Shell environment overrides inherited dumb terminal

- **WHEN** the desktop process environment contains `TERM=dumb` and an embedded shell is started
- **THEN** the shell process environment contains `TERM=xterm-256color`
- **AND** the shell process environment contains `COLORTERM=truecolor`

#### Scenario: Shell environment fills missing UTF-8 locale

- **WHEN** the desktop process environment does not contain `LANG` or `LC_CTYPE`
- **AND** an embedded shell is started
- **THEN** the shell process environment contains a UTF-8 `LANG` value
- **AND** the shell process environment contains a UTF-8 `LC_CTYPE` value

#### Scenario: Shell environment preserves existing UTF-8 locale

- **WHEN** the desktop process environment contains `LANG=zh_CN.UTF-8`
- **AND** the desktop process environment contains `LC_CTYPE=zh_CN.UTF-8`
- **AND** an embedded shell is started
- **THEN** the shell process environment preserves `LANG=zh_CN.UTF-8`
- **AND** the shell process environment preserves `LC_CTYPE=zh_CN.UTF-8`

## ADDED Requirements

### Requirement: Preserve UTF-8 Terminal Output Across Read Chunks

系统 SHALL 在 PTY 输出进入字符串过滤、Wails 事件和终端历史之前保持 UTF-8 多字节字符完整。若一次 PTY read 以不完整 UTF-8 字节序列结束，系统 SHALL 暂存该不完整序列并与下一次 read 合并后再解码。系统 MUST NOT 因 read chunk 边界引入替换字符。

#### Scenario: Split Chinese character is rendered intact

- **WHEN** shell 输出文本 `中文 ✓`
- **AND** PTY read 在 `中` 的 UTF-8 字节中间分割输出
- **THEN** 前端终端收到的数据包含 `中文 ✓`
- **AND** 终端历史保存的数据包含 `中文 ✓`
- **AND** 输出不包含因分片产生的替换字符

#### Scenario: Split UTF-8 output still allows command-state filtering

- **WHEN** shell 输出包含应用私有 command-state payload
- **AND** payload 前后相邻的普通输出包含跨 read 分割的 UTF-8 字符
- **THEN** 系统仍消费 command-state payload
- **AND** 普通 UTF-8 输出保持完整渲染和持久化

### Requirement: Use Robust Terminal Font Fallbacks

系统 SHALL 为 xterm 会话配置覆盖常见 macOS 终端字符的字体 fallback。字体栈 SHALL 保留等宽字体优先级，并 SHALL 包含中文、emoji、Powerline/Nerd Font 符号和 macOS 常见等宽字体 fallback。系统 MUST NOT 依赖单一字体存在才能显示普通中文或 emoji。

#### Scenario: New xterm session includes Unicode fallback fonts

- **WHEN** 系统创建新的嵌入式 xterm 会话
- **THEN** xterm 字体栈包含常见等宽字体
- **AND** xterm 字体栈包含中文字体 fallback
- **AND** xterm 字体栈包含 emoji 字体 fallback
- **AND** xterm 字体栈包含 Powerline 或 Nerd Font 符号 fallback

#### Scenario: Terminal keeps fixed theme while font fallback changes

- **WHEN** 系统创建新的嵌入式 xterm 会话
- **THEN** xterm 仍使用既有固定终端配色
- **AND** 字体 fallback 调整不改变终端颜色主题
