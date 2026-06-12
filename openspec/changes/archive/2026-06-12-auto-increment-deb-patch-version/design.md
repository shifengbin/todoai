## Context

当前 Debian 打包入口是 `scripts/package-deb.sh`。脚本会执行 Wails Linux 构建，组装 `DEBIAN/control`、desktop launcher 和图标，再通过 `dpkg-deb` 输出 `.deb` 文件。版本号目前由脚本内默认值提供，并允许调用方通过 `VERSION` 环境变量覆盖；这让默认打包路径很容易反复生成同一版本号。

该变更只涉及发布/打包流程，不影响 Go 后端、Vue 前端、Wails 运行时行为或用户数据格式。

## Goals / Non-Goals

**Goals:**

- 默认执行 deb 打包脚本时自动生成下一个 patch 版本。
- 使用同一个版本值写入 Debian metadata 和 `.deb` 输出文件名。
- 只有完整打包成功后才持久化新版本，失败时保留原版本。
- 继续支持 `VERSION=... scripts/package-deb.sh` 这种显式版本覆盖。

**Non-Goals:**

- 不引入完整发布管理、Git tag、changelog 或 CI 发布流程。
- 不修改应用内关于版本号的展示或 Wails 配置 schema。
- 不改变 major/minor 的自动升级规则；本次只处理 patch 递增。

## Decisions

### 使用根目录 `VERSION` 文件作为版本状态

新增根目录 `VERSION` 文件，保存最近一次成功打包使用的 semver，例如 `0.1.8`。默认执行脚本时读取该文件、计算下一个 patch 版本并用于本次产物。

选择文件而不是继续把版本写在脚本里，是因为版本状态是发布数据，不是脚本逻辑。文件也更容易被测试、review 和手动编辑。替代方案是从已有 `.deb` 文件名推断最大版本，但 build 目录可能被清理，且不同架构产物会让推断复杂化。

### 默认路径递增 patch，显式 `VERSION` 覆盖保持可用

未设置 `VERSION` 时，脚本从 `VERSION` 文件读取当前版本并递增最后一段：`0.1.8 -> 0.1.9 -> 0.1.10`。设置 `VERSION` 时，脚本直接使用调用方给出的版本，适合手动发版或修正版本号。

显式覆盖成功后也写回 `VERSION`，让后续默认打包从该版本继续递增。替代方案是覆盖时不写回，但这样手动发布 `0.2.0` 后下一次默认打包仍可能回到旧 patch 线。

### 成功后写回，失败不消耗版本号

脚本在 Wails 构建、文件组装和 `dpkg-deb` 全部成功后，才把本次版本写入 `VERSION`。任何失败都会保持原文件内容。

这避免了构建失败导致版本号跳号。代价是如果包已经生成但写回失败，产物版本可能没有被记录；脚本应把写回作为最后一个必须成功步骤，让调用方能看到失败。

### 使用 shell 内建解析 semver patch

脚本只需要支持 `X.Y.Z` 三段数字版本。实现时可以使用 Bash 正则校验并递增第三段，不引入新依赖。

Debian 支持更复杂版本格式，但当前项目的 README 和脚本都使用 `0.1.x` 风格。限制为三段数字可以让“patch 自动加 1”语义明确。

## Risks / Trade-offs

- [Risk] `VERSION` 文件缺失或格式错误会阻断默认打包。→ Mitigation: 脚本给出清晰错误；仓库提交初始 `VERSION` 文件。
- [Risk] 并发执行两次打包脚本可能计算出同一个下一个版本。→ Mitigation: 当前本地桌面应用发布流程以人工单次打包为主；如未来进入 CI 并发发布，再引入文件锁。
- [Risk] 显式覆盖版本写回可能把错误版本持久化。→ Mitigation: 脚本对 `VERSION` 值执行三段数字 semver 校验，并在输出中打印最终 `.deb` 路径。
- [Risk] README 示例版本可能过期。→ Mitigation: README 使用“示例”或泛化说明，不绑定固定 patch 号。
