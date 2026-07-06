## 1. 菜单与剪贴板测试

- [x] 1.1 为应用菜单添加测试，覆盖 macOS 标准 App/Edit/Window 菜单角色和既有“文件”菜单动作共存。
- [x] 1.2 为 TODO 右键复制添加前端测试，断言中文、emoji 和终端符号标题/描述原样传入剪贴板写入函数。
- [x] 1.3 为 TODO 右键复制失败添加前端测试，断言剪贴板错误会通过现有错误提示展示且 TODO 数据不变。
- [x] 1.4 为终端复制/粘贴添加前端测试，覆盖 Unicode 文本从终端选择写入剪贴板、从剪贴板发送到 shell。

## 2. macOS 编辑菜单与 TODO 剪贴板

- [x] 2.1 调整应用菜单构建，在 macOS 下补齐 Wails 标准 App/Edit/Window 菜单角色并保留现有“文件”菜单。
- [x] 2.2 调整 TODO 标题/描述复制逻辑，等待剪贴板写入完成并在失败时调用现有错误提示。
- [x] 2.3 确认创建 TODO 名称和描述输入框不拦截系统粘贴动作，并通过测试覆盖 Unicode 粘贴场景。

## 3. UTF-8 环境与终端输出

- [x] 3.1 添加 UTF-8 locale 环境处理测试，覆盖缺失 locale 自动补齐、已有 UTF-8 locale 保留、非 UTF-8 locale 替换。
- [x] 3.2 实现应用进程和嵌入式 shell 的 UTF-8 locale 补齐，确保 `LANG` 和 `LC_CTYPE` 在缺失或非 UTF-8 时使用 UTF-8 值。
- [x] 3.3 添加 PTY 输出 UTF-8 分片测试，覆盖中文字符跨 read 分割后前端事件和终端历史仍保留完整文本。
- [x] 3.4 实现 PTY 输出 UTF-8 chunk decoder，在进入 command-state filter、Wails 事件和历史保存前合并不完整多字节字符。
- [x] 3.5 添加 command-state payload 与 UTF-8 分片相邻时的测试，确认内部 payload 仍被消费且普通输出完整。

## 4. 终端字体 fallback

- [x] 4.1 为 xterm session 字体配置添加前端测试，断言字体栈包含等宽、中文、emoji 和 Powerline/Nerd Font 符号 fallback。
- [x] 4.2 调整 xterm 字体栈，保留现有字体并补充 macOS 常见字体和符号字体 fallback。
- [x] 4.3 确认字体 fallback 调整不改变既有 xterm 配色和 `convertEol` 行为。

## 5. 验证与收尾

- [x] 5.1 运行 Go 单元测试，确认菜单、环境、PTY 输出和历史相关测试通过。
- [x] 5.2 运行前端自动化测试，确认 TODO 剪贴板、终端剪贴板和 xterm 字体相关测试通过。
- [x] 5.3 执行自动 review，检查变更是否符合设计、规格和现有代码风格。
- [x] 5.4 在 macOS 手工验证 `Cmd+V` 粘贴 TODO 名称/描述、右键复制中文 TODO、终端显示 `printf '中文 ✓ 🔧   \n'`。
- [x] 5.5 运行 `wails build -tags webkit2_41` 生成可执行文件。
