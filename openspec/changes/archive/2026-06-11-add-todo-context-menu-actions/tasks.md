## 1. TODO 右键菜单交互

- [x] 1.1 在 `ProjectSidebar.vue` 中新增 TODO 右键菜单状态、打开/关闭函数和菜单互斥关闭逻辑
- [x] 1.2 在 `not-started` 和 `in-progress` TODO 行上绑定右键菜单触发区域，并渲染查看详情、添加项目、复制描述、删除 TODO 菜单项
- [x] 1.3 将 TODO 行外动作收紧为状态按钮：`not-started` 仅保留开始按钮，`in-progress` 仅保留完成按钮
- [x] 1.4 将删除 TODO 流程改为从右键菜单触发确认气泡，并确保取消、外部点击和确认后的关闭行为正确
- [x] 1.5 为 TODO 右键菜单添加与现有侧边栏风格一致的样式，确保菜单项在窄侧边栏下可读

## 2. App 事件与弹窗关闭规则

- [x] 2.1 在 `ProjectSidebar` 与 `App.vue` 之间新增 `copy-todo-description` 事件，并在 App 中通过 `ClipboardSetText` 写入 TODO `description`
- [x] 2.2 确保复制空描述时写入空字符串，不拼接 TODO 标题或其他字段
- [x] 2.3 移除 TODO 创建、TODO 详情编辑和 TODO 添加项目弹窗的遮罩点击关闭能力
- [x] 2.4 保留 TODO 弹窗关闭按钮、取消按钮和提交成功后的关闭行为

## 3. 客户端自动化测试

- [x] 3.1 更新 `ProjectSidebar.test.js`，覆盖 TODO 右键菜单展示、菜单项事件、外部点击关闭和与其他侧边栏浮层互斥
- [x] 3.2 更新 `ProjectSidebar.test.js`，覆盖行外只保留开始/完成状态按钮，不再显示查看详情、添加项目或删除按钮
- [x] 3.3 更新 `App.test.js`，将查看详情、添加项目和删除 TODO 的测试入口改为右键菜单
- [x] 3.4 更新 `App.test.js`，覆盖复制 TODO 描述到剪贴板和空描述复制空字符串
- [x] 3.5 更新 `App.test.js`，覆盖 TODO 创建、详情编辑和添加项目弹窗点击遮罩不关闭，且显式关闭控件仍可关闭

## 4. 验证与 Review

- [x] 4.1 运行客户端自动化测试，确认 TODO 菜单、剪贴板和弹窗关闭规则通过
- [x] 4.2 运行 OpenSpec 校验，确认 proposal、design、specs、tasks 可解析
- [x] 4.3 执行自动 review，检查右键菜单可访问性、浮层互斥、剪贴板错误处理和测试覆盖

## 5. 打包

- [x] 5.1 运行 `wails build -tags webkit2_41`，生成可执行文件
