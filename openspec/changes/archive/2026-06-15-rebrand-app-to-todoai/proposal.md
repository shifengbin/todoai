## Why

当前应用仍以 `TUI Helper` / `tui-helper` 发布和显示，但产品定位已经转向以 TODO 驱动的 AI 工作流。将应用统一更名为 `TodoAI` 可以让窗口、安装包、启动器和图标传达一致的产品身份。

## What Changes

- 将用户可见应用名称从 `TUI Helper` / `tui-helper` 统一改为 `TodoAI`。
- 将发布身份中的包名、二进制名和构建输出名从 `tui-helper` 改为 `todoai`。
- 生成并替换新的 `TodoAI` 启动图标，用于 Wails、Windows、macOS 和 Linux 打包资源。
- 将默认配置目录迁移到 `todoai`，并兼容读取旧 `tui-helper` 配置，避免升级后丢失已有项目、TODO、设置和终端历史。
- 更新应用私有 command-state 协议的品牌标识为 `todoai`，同时继续识别旧 `tui-helper` payload 以兼容运行中或历史输出。
- 更新 README、打包脚本测试和相关构建元数据。
- **BREAKING**: 新安装包、二进制文件和 Linux 命令名改为 `todoai`；依赖旧 `tui-helper` 命令名的外部脚本需要调整。

## Capabilities

### New Capabilities

- `application-identity`: 定义应用名称、发布身份、图标资产和本地数据目录迁移行为。

### Modified Capabilities

- `linux-deb-packaging`: Debian 包名、二进制名、desktop launcher 元数据和图标安装路径改为 `todoai` / `TodoAI`。
- `embedded-terminal-emulation`: 应用私有 command-state 协议改为使用 `todoai` 标识，并继续兼容旧 `tui-helper` 标识。

## Impact

- Wails 配置和启动入口：`wails.json`、`main.go`、`frontend/index.html`。
- 打包资源：`build/appicon.png`、`build/windows/icon.ico`、Windows/macOS/Linux 构建模板和 README。
- Linux Debian 打包脚本及测试：`scripts/package-deb.sh`、`scripts/package-deb.test.sh`。
- 本地数据路径：`defaultProjectConfigPath`、设置路径、终端历史目录，以及相关测试。
- 终端集成和过滤：shell 临时脚本、OSC 777 command-state 解析、ConPTY 文本 fallback、xterm OSC handler 和相关前后端测试。
