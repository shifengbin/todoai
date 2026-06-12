## Why

Debian 打包脚本当前把默认版本号写死在脚本里，重复执行会反复生成同一个版本的 `.deb` 产物，容易覆盖产物或安装时无法区分新包。让脚本在默认打包路径中自动递增 patch 版本，可以减少手动改版本号的漏改和出错。

## What Changes

- 为 deb 打包流程增加持久化的版本来源，用于记录当前/最近一次成功打包版本。
- 默认执行 `scripts/package-deb.sh` 时自动将 patch 版本加 1，并使用递增后的版本写入 Debian control metadata 和输出文件名。
- 仅在 Wails 构建和 `dpkg-deb` 打包成功后写回版本状态，失败构建不消耗版本号。
- 保留通过 `VERSION=... scripts/package-deb.sh` 手动指定版本的能力，便于需要指定版本号的发布场景。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `linux-deb-packaging`: deb 打包命令需要在默认路径下自动递增 patch 版本，并在成功打包后持久化该版本。

## Impact

- 影响 `scripts/package-deb.sh` 的版本解析、产物命名和成功后状态写回逻辑。
- 可能新增根目录版本状态文件，例如 `VERSION`，作为默认打包版本来源。
- 影响 README 中 Debian 打包命令和产物示例的描述。
- 不改变应用运行时 API、前端行为或 Wails 应用启动行为。
