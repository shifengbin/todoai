## 1. 后端设置模型

- [x] 1.1 在 `settings.go` 中增加外观主题常量、规范化函数和 `TerminalSettingsState.Theme` 字段
- [x] 1.2 将 `theme` 写入并读取现有 `settings.json`，缺失或无效值默认规范化为 `light`
- [x] 1.3 增加 `SaveTheme`/`SaveTerminalTheme` 保存路径，拒绝 `light`/`dark` 之外的值并保留旧值
- [x] 1.4 补充 `settings_test.go` 覆盖缺失主题、恢复 `dark`、无效主题默认值、保存有效主题、拒绝无效主题

## 2. Wails API 与绑定

- [x] 2.1 在 `app.go` 暴露保存外观主题的方法，并保持 `LoadTerminalSettings` 返回主题字段
- [x] 2.2 更新 `frontend/wailsjs/go/main/App.js`、`App.d.ts` 和 `frontend/wailsjs/go/models.ts` 绑定
- [x] 2.3 补充或更新 Go 应用层测试，验证主题设置不会影响 shell 解析和项目状态

## 3. 前端主题状态与设置界面

- [x] 3.1 在 `App.vue` 中维护当前主题和设置弹窗临时主题值
- [x] 3.2 应用启动和打开设置时从 `LoadTerminalSettings` 同步主题
- [x] 3.3 在设置弹窗增加 `Light`/`Dark` 外观主题选择
- [x] 3.4 保存设置时调用主题保存 API，成功后更新全局主题，失败时保持弹窗打开并显示错误
- [x] 3.5 取消设置修改时不改变当前全局主题
- [x] 3.6 补充 `App.test.js` 覆盖主题加载、保存切换、取消不生效和保存失败

## 4. 样式系统

- [x] 4.1 在 `style.css` 中定义 light/dark CSS variables 色板
- [x] 4.2 将终端以外的应用外壳、侧栏、工作区、设置弹窗、状态栏和菜单硬编码颜色替换为变量
- [x] 4.3 确认终端以外页面的状态 chip、活动指示、错误/警告/主按钮在 light/dark 下仍有可读对比度

## 5. 终端配色保持不变

- [x] 5.1 保持 `xtermFactory.js` 使用固定 xterm theme，不接收应用外观主题
- [x] 5.2 保持 `terminalManager.js` 不向终端 session 传递或更新应用外观主题
- [x] 5.3 在主题保存成功后不重启 shell session，且不更新已打开终端配色
- [x] 5.4 补充 `xtermFactory.test.js` 和 `terminalManager.test.js` 覆盖终端配色不随应用主题变化

## 6. 验证与打包

- [x] 6.1 运行 `go test ./...`
- [x] 6.2 运行 `cd frontend && npm run test`
- [x] 6.3 运行 `cd frontend && npm run build`
- [x] 6.4 运行 `wails build -tags webkit2_41` 验证桌面应用编译
- [x] 6.5 如需发布 Debian 包，运行 `scripts/package-deb.sh` 验证打包流程
