## ADDED Requirements

### Requirement: Preserve Unicode Terminal Clipboard Text

系统 SHALL 在嵌入式终端复制和粘贴路径中保持 Unicode 文本保真。终端选中文本写入系统剪贴板时 SHALL 原样保存中文、emoji、校验符号和终端符号。系统剪贴板文本粘贴到终端时 SHALL 原样发送给对应 shell 会话，MUST NOT 在前端剪贴板层引入 mojibake 或替换字符。

#### Scenario: Copy Unicode terminal selection

- **WHEN** 活动终端中选中文本为 `中文 ✓ 🔧   `
- **AND** 用户通过终端复制快捷键或终端右键菜单触发复制
- **THEN** 系统剪贴板内容为 `中文 ✓ 🔧   `
- **AND** 系统剪贴板内容不包含 mojibake 或替换字符

#### Scenario: Paste Unicode clipboard text into terminal

- **WHEN** 系统剪贴板内容为 `printf '中文 ✓ 🔧   \n'`
- **AND** 用户通过终端粘贴快捷键或终端右键菜单触发粘贴
- **THEN** 系统向活动 shell 会话发送 `printf '中文 ✓ 🔧   \n'`
- **AND** 发送内容不包含 mojibake 或替换字符
