# TUI Helper

## 简介

TUI Helper 是一个 Wails 桌面应用，前端使用 Vue，后端使用 Go。它会持久化本地项目目录列表，并为当前选中的项目提供嵌入式 shell 会话。

## 平台支持

TUI Helper 支持 Linux 桌面，并为 Windows 10 1809+ 和 Windows 11 提供 ConPTY 嵌入式终端支持。

在 Windows 上，后端会按以下顺序探测系统 shell：

- PowerShell 7：`pwsh.exe`
- Windows PowerShell：`powershell.exe`
- Cmd：通过 `COMSPEC` 或 `cmd.exe`

手动配置终端 shell 路径时，Windows 会按 Windows 可执行文件语义校验，支持 `.exe`、`.cmd`、`.bat`、`.com`，以及 `PATHEXT` 中声明的扩展名。类 Unix 系统继续使用 `$SHELL` 和已知 shell 路径，例如 `/bin/zsh`、`/bin/bash`、`/bin/sh`。

嵌入式终端后端按平台实现：Linux/macOS 使用 `creack/pty`，Windows 使用 ConPTY。低于 Windows 10 1809、缺少 ConPTY API 或运行环境不支持 ConPTY 时，应用会显示稳定的 unsupported 状态，不会自动重复启动 shell。

包元数据：

- Wails 主版本：v2
- 应用名称：TUI Helper
- 包名：`tui-helper`
- 版本：`0.1.0`
- 维护者：`FengbinShi <shifengbin@jiandan100.cn>`
- 图标：`build/appicon.png`

## 本地开发

在项目目录中运行 `wails dev` 可以启动实时开发模式。该命令会启动 Vite 开发服务器，前端改动可以快速热重载。

如果需要在浏览器中调试并访问 Go 方法，可以打开开发服务器地址 http://localhost:34115，然后在浏览器 devtools 中调用 Go 代码。

## 构建

后端测试：

```bash
go test ./...
```

前端测试和构建：

```bash
cd frontend
npm run test
npm run build
```

Wails 构建：

```bash
wails build -tags webkit2_41
```

Windows 兼容性检查：

```bash
GOOS=windows GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go test -c -o /tmp/tui-helper-windows.test.exe .
```

Debian 安装包：

```bash
scripts/package-deb.sh
```

Debian 打包脚本会构建 Linux Wails 二进制文件，组装包元数据，并输出 `build/bin/tui-helper_0.1.0_amd64.deb`。
