# 隐藏 TODO 列表滚动条实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 隐藏三个 TODO 视图共用滚动区的可见滚动条，同时保留原有滚动和拖拽自动滚动能力。

**架构：** 只调整 `.todo-workspace-scroll` 的跨浏览器 CSS，不改变 DOM、Vue 状态或 SortableJS 配置。通过组件样式契约测试确认滚动仍启用、Firefox/WebKit 隐藏规则齐全且不再预留滚动条槽位。

**技术栈：** Vue 3、CSS、Vitest、Vue Test Utils、Vite

---

### 任务 1：隐藏 TODO 工作区滚动条

**文件：**

- 修改：`frontend/src/components/ProjectSidebar.test.js`
- 修改：`frontend/src/style.css`

**步骤 1：编写失败测试**

在 `ProjectSidebar.test.js` 的布局测试附近新增独立测试：

```javascript
it('keeps the TODO workspace scrollable while hiding its scrollbar', () => {
  const styles = readFileSync('src/style.css', 'utf8')
  const scrollRule = styles.slice(styles.indexOf('.todo-workspace-scroll {'), styles.indexOf('.todo-workspace-scroll::-webkit-scrollbar {'))
  const webkitRule = styles.slice(styles.indexOf('.todo-workspace-scroll::-webkit-scrollbar {'), styles.indexOf('.app-shell.has-modal-overlay'))

  expect(scrollRule).toContain('overflow-y: auto;')
  expect(scrollRule).toContain('scrollbar-width: none;')
  expect(scrollRule).toContain('-ms-overflow-style: none;')
  expect(scrollRule).not.toContain('scrollbar-gutter: stable;')
  expect(webkitRule).toContain('display: none;')
})
```

**步骤 2：运行测试并确认失败**

运行：

```bash
cd frontend && npm run test -- src/components/ProjectSidebar.test.js -t "keeps the TODO workspace scrollable while hiding its scrollbar"
```

预期：测试因缺少 `scrollbar-width: none`、`-ms-overflow-style: none` 和 WebKit 规则而失败。

**步骤 3：添加最小样式实现**

将滚动容器规则调整为：

```css
.todo-workspace-scroll {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding-right: 10px;
    scrollbar-width: none;
    -ms-overflow-style: none;
}

.todo-workspace-scroll::-webkit-scrollbar {
    display: none;
    width: 0;
    height: 0;
}
```

保留弹窗打开时已有的 `overflow-y: hidden` 规则。

**步骤 4：运行聚焦测试并确认通过**

运行：

```bash
cd frontend && npm run test -- src/components/ProjectSidebar.test.js -t "keeps the TODO workspace scrollable while hiding its scrollbar"
```

预期：聚焦测试通过。

**步骤 5：运行完整前端验证**

运行：

```bash
cd frontend && npm run test
cd frontend && npm run build
git diff --check
```

预期：全部测试通过，生产构建成功，补丁无空白错误。当前工作区已有待归档变更，本任务不创建提交。
