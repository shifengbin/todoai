import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ProjectSidebar from './ProjectSidebar.vue'

describe('ProjectSidebar', () => {
  afterEach(() => {
    vi.useRealTimers()
    document.body.innerHTML = ''
  })

  it('renders the TODO tree and emits TODO-scoped terminal actions', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [{ name: 'codex', command: 'codex', enabled: true }]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.find('[data-testid="workspace-tabs"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sidebar-tab-projects"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="project-library"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('修复登录问题')
    expect(wrapper.text()).toContain('alpha')
    expect(wrapper.text()).toContain('codex')
    expect(wrapper.find('[data-testid="todo-todo-a"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toBe('alpha')
    expect(
      wrapper
        .find('[data-testid="todo-project-todo-project-a"]')
        .element.closest('.todo-project-header-row')
        .querySelector('.branch-toggle-placeholder')
    ).toBeNull()
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').classes()).toContain('active')

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-edit-todo-a"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-add-project-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')
    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-delete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('create-todo')).toHaveLength(1)
    expect(wrapper.emitted('edit-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('add-project-to-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('select-todo-project')[0]).toEqual(['todo-project-a'])
    expect(wrapper.emitted('create-terminal')[0]).toEqual(['todo-project-a', { name: 'codex', command: 'codex', enabled: true }])
    expect(wrapper.emitted('select-terminal')[0]).toEqual(['terminal-a'])
    expect(wrapper.emitted('complete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('delete-todo')[0]).toEqual(['todo-a'])
  })

  it('keeps TODO project rows free of project folder menus', async () => {
    const wrapper = mountInProgressSidebar()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-project-menu-button-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-menu-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-menu-open-folder-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').exists()).toBe(true)
  })

  it('hides disabled launch profiles while keeping Terminal available', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: true },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]')
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).toContain('codex')
    expect(menu.text()).not.toContain('claude')

    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')

    expect(wrapper.emitted('create-terminal')[0]).toEqual(['todo-project-a', { name: 'codex', command: 'codex', enabled: true }])
  })

  it('uses the terminal launch menu when creating task terminals', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: true },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-task-terminal-todo-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="terminal-launch-menu-task-todo-a"]')
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).toContain('codex')
    expect(menu.text()).not.toContain('claude')

    await wrapper.find('[data-testid="terminal-launch-option-task-todo-a-1"]').trigger('click')

    expect(wrapper.emitted('create-task-terminal')[0]).toEqual(['todo-a', { name: 'codex', command: 'codex', enabled: true }])
  })

  it('shows only Terminal when all custom launch profiles are disabled', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: false },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    const options = wrapper.findAll('[data-testid^="terminal-launch-option-todo-project-a-"]')
    expect(options).toHaveLength(1)
    expect(options[0].text()).toBe('Terminal')
  })

  it('shows TODOs in not-started, in-progress, and completed views', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-not-started', title: '整理文档', status: 'not-started', createdAt: '2026-06-10T09:00:00Z' },
          { id: 'todo-in-progress', title: '修复登录问题', status: 'in-progress', createdAt: '2026-06-10T10:00:00Z' },
          { id: 'todo-completed', title: '已完成任务', status: 'completed', completedAt: '2026-06-10T11:00:00Z' },
          { id: 'todo-deleted', title: '已删除任务', status: 'deleted', archivedReason: 'deleted' }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(wrapper.find('[data-testid="todo-view-not-started"]').classes()).toContain('active')
    expect(visibleTodoTitles(wrapper, 'not-started-todos')).toEqual(['整理文档'])

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    expect(visibleTodoTitles(wrapper, 'in-progress-todos')).toEqual(['修复登录问题'])

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    expect(completedTodoTitles(wrapper)).toEqual(['已完成任务'])
    expect(wrapper.find('[data-testid="completed-todos"]').text()).not.toContain('已删除任务')
  })

  it('uses the controlled TODO view prop and emits changes', async () => {
    const wrapper = mountSidebar({
      props: {
        todoView: 'completed'
      }
    })

    expect(wrapper.find('[data-testid="todo-view-completed"]').classes()).toContain('active')
    expect(completedTodoTitles(wrapper)).toEqual(['已完成任务'])

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.emitted('update:todo-view')[0]).toEqual(['in-progress'])
    expect(wrapper.emitted('todo-view-change')[0]).toEqual(['in-progress'])
  })

  it('emits manual TODO status changes from TODO rows', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-a', title: '修复登录问题', status: 'not-started' },
          { id: 'todo-b', title: '升级依赖', status: 'in-progress' }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="mark-todo-in-progress-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.emitted('change-todo-status')[0]).toEqual(['todo-a', 'in-progress'])
    expect(wrapper.find('[data-testid="mark-todo-not-started-todo-b"]').exists()).toBe(false)
  })

  it('shows TODO management actions in the context menu', async () => {
    const wrapper = mountSidebar()

    await openTodoContextMenu(wrapper, 'todo-a')

    const menu = wrapper.find('[data-testid="todo-context-menu-todo-a"]')
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('View details')
    expect(menu.text()).toContain('Add project')
    expect(menu.text()).toContain('Copy title and description')
    expect(menu.text()).toContain('Delete TODO')

    await wrapper.find('[data-testid="todo-menu-copy-title-description-todo-a"]').trigger('click')

    expect(wrapper.emitted('copy-todo-description')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
  })

  it('positions the TODO context menu near the right-click point without leaving the viewport', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('contextmenu', {
      clientX: window.innerWidth + 400,
      clientY: window.innerHeight + 300
    })
    await nextTick()

    const menu = wrapper.find('[data-testid="todo-context-menu-todo-a"]')
    const left = Number.parseFloat(menu.element.style.left)
    const top = Number.parseFloat(menu.element.style.top)

    expect(menu.exists()).toBe(true)
    expect(left).toBeGreaterThanOrEqual(0)
    expect(top).toBeGreaterThanOrEqual(0)
    expect(left).toBeLessThan(window.innerWidth)
    expect(top).toBeLessThan(window.innerHeight)
  })

  it('positions the TODO context menu near the three-dot button without leaving the viewport', async () => {
    const wrapper = mountSidebar()
    const button = wrapper.find('[data-testid="todo-menu-button-todo-a"]')
    vi.spyOn(button.element, 'getBoundingClientRect').mockReturnValue({
      left: window.innerWidth + 250,
      right: window.innerWidth + 280,
      top: window.innerHeight + 120,
      bottom: window.innerHeight + 150,
      width: 30,
      height: 30,
      x: window.innerWidth + 250,
      y: window.innerHeight + 120,
      toJSON: () => {}
    })

    await button.trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="todo-context-menu-todo-a"]')
    const left = Number.parseFloat(menu.element.style.left)
    const top = Number.parseFloat(menu.element.style.top)

    expect(menu.exists()).toBe(true)
    expect(left).toBeGreaterThanOrEqual(0)
    expect(top).toBeGreaterThanOrEqual(0)
    expect(left).toBeLessThan(window.innerWidth)
    expect(top).toBeLessThan(window.innerHeight)
  })

  it('emits TODO management actions from the context menu', async () => {
    const wrapper = mountSidebar()

    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-edit-todo-a"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-add-project-todo-a"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.emitted('edit-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('add-project-to-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()
  })

  it('closes TODO context menus from outside clicks and other sidebar popovers', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [{ name: 'codex', command: 'codex' }]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    window.dispatchEvent(new MouseEvent('click'))
    await nextTick()

    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
    expect(wrapper.emitted('edit-todo')).toBeUndefined()
    expect(wrapper.emitted('add-project-to-todo')).toBeUndefined()
    expect(wrapper.emitted('copy-todo-description')).toBeUndefined()

    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]').exists()).toBe(true)
  })

  it('groups TODO workflow tabs and status item actions into single rows', () => {
    const wrapper = mountSidebar()
    const workflowTabs = wrapper.find('[data-testid="todo-workflow-tabs"]')
    const actionGroup = wrapper.find('[data-testid="todo-actions-todo-a"]')
    const styles = readFileSync('src/style.css', 'utf8')
    const sidebarHeaderRule = styles.slice(styles.indexOf('.sidebar-header {'), styles.indexOf('.sidebar-title'))
    const tabsRule = styles.slice(styles.indexOf('.todo-view-tabs {'), styles.indexOf('.todo-tree-toolbar'))
    const actionsRule = styles.slice(styles.indexOf('.todo-actions {'), styles.indexOf('.todo-action-button'))

    expect(workflowTabs.exists()).toBe(true)
    expect(Array.from(workflowTabs.element.children).map((node) => node.getAttribute('data-testid'))).toEqual([
      'todo-view-not-started',
      'todo-view-in-progress',
      'todo-view-completed'
    ])
    expect(actionGroup.exists()).toBe(true)
    const actionTestIds = Array.from(actionGroup.element.children).map((node) => {
      return node.getAttribute('data-testid') || node.querySelector('[data-testid]')?.getAttribute('data-testid')
    })
    expect(actionTestIds).toEqual(['todo-menu-button-todo-a', 'mark-todo-in-progress-todo-a'])
    expect(wrapper.find('[data-testid="edit-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-project-to-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="delete-todo-todo-a"]').exists()).toBe(false)
    expect(sidebarHeaderRule).toContain('flex-shrink: 0;')
    expect(wrapper.find('[data-testid="workspace-tabs"]').exists()).toBe(false)
    expect(tabsRule).toContain('grid-template-columns: repeat(3, minmax(0, 1fr));')
    expect(actionsRule).toContain('display: inline-flex;')
    expect(actionsRule).toContain('flex-wrap: nowrap;')
  })

  it('places task terminal creation on TODO rows and hides empty task terminal groups', async () => {
    const wrapper = mountInProgressSidebar()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    const actionGroup = wrapper.find('[data-testid="todo-actions-todo-a"]')
    const actionTestIds = Array.from(actionGroup.element.children).map((node) => {
      return node.getAttribute('data-testid') || node.querySelector('[data-testid]')?.getAttribute('data-testid')
    })
    expect(actionTestIds).toEqual(['todo-menu-button-todo-a', 'add-task-terminal-todo-a', 'complete-todo-todo-a'])
    expect(wrapper.find('[data-testid="task-terminal-list-todo-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="add-task-terminal-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-task-todo-a-0"]').trigger('click')

    expect(wrapper.emitted('create-task-terminal')[0]).toEqual(['todo-a', null])
  })

  it('shows task terminal groups only when task terminals exist', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        terminals: [
          {
            id: 'task-terminal-a',
            todoId: 'todo-a',
            shellName: 'bash',
            currentCommand: '',
            state: 'running'
          }
        ],
        activeTerminalId: 'task-terminal-a'
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.find('[data-testid="task-terminal-list-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="task-terminal-task-terminal-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="task-terminal-list-todo-a"]').find('[data-testid="add-task-terminal-todo-a"]').exists()).toBe(false)
  })

  it('uses the top toolbar for completed bulk deletion without open TODO controls', async () => {
    const wrapper = mountSidebar()

    expect(wrapper.find('[data-testid="todo-tree-toolbar"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-tree-toolbar"]').attributes('role')).toBe('toolbar')
    expect(wrapper.find('[data-testid="sort-active-todos-priority"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="collapse-all-todos"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="expand-all-todos"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-tree-toolbar"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-tree-toolbar"]').attributes('role')).toBe('toolbar')
    expect(wrapper.find('[data-testid="sort-active-todos-priority"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sort-active-todos-time"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="collapse-all-todos"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="expand-all-todos"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="bulk-delete-completed-todos"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="completed-todo-toolbar"]').exists()).toBe(false)
  })

  it('opens completed TODO details from the completed row menu', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-button-todo-completed"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="completed-todo-menu-todo-completed"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="completed-todo-menu-edit-todo-completed"]').exists()).toBe(true)

    await wrapper.find('[data-testid="completed-todo-menu-edit-todo-completed"]').trigger('click')

    expect(wrapper.emitted('edit-todo')[0]).toEqual(['todo-completed'])
    expect(wrapper.find('[data-testid="completed-todo-menu-todo-completed"]').exists()).toBe(false)
  })

  it('confirms completed TODO deletion from the completed row menu', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-button-todo-completed"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-delete-todo-completed"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-completed"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="confirm-delete-todo-todo-completed"]').trigger('click')

    expect(wrapper.emitted('delete-todo')[0]).toEqual(['todo-completed'])
    expect(wrapper.find('[data-testid="delete-todo-popover-todo-completed"]').exists()).toBe(false)
  })

  it('selects completed TODOs and confirms bulk deletion from the completed view only', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-a', title: '活动任务', status: 'not-started' },
          { id: 'todo-completed-a', title: '已完成 A', status: 'completed', completedAt: '2026-06-10T10:00:00Z' },
          { id: 'todo-completed-b', title: '已完成 B', status: 'completed', completedAt: '2026-06-10T11:00:00Z' }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(wrapper.find('[data-testid="bulk-delete-completed-todos"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    const bulkDelete = wrapper.find('[data-testid="bulk-delete-completed-todos"]')
    expect(bulkDelete.attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="collapse-all-todos"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="expand-all-todos"]').exists()).toBe(false)

    await wrapper.find('[data-testid="select-completed-todo-todo-completed-a"]').trigger('click')
    await wrapper.find('[data-testid="select-completed-todo-todo-completed-b"]').trigger('click')

    expect(wrapper.find('[data-testid="select-completed-todo-todo-completed-a"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="select-completed-todo-todo-completed-b"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="bulk-delete-completed-todos"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('[data-testid="bulk-delete-completed-todos"]').text()).toContain('2')

    await wrapper.find('[data-testid="bulk-delete-completed-todos"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="bulk-delete-completed-todos-popover"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-completed-todos')).toBeUndefined()

    await wrapper.find('[data-testid="cancel-bulk-delete-completed-todos"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="bulk-delete-completed-todos-popover"]').exists()).toBe(false)

    await wrapper.find('[data-testid="bulk-delete-completed-todos"]').trigger('click')
    await wrapper.find('[data-testid="confirm-bulk-delete-completed-todos"]').trigger('click')

    expect(wrapper.emitted('delete-completed-todos')[0]).toEqual([['todo-completed-a', 'todo-completed-b']])
    expect(wrapper.find('[data-testid="bulk-delete-completed-todos-popover"]').exists()).toBe(false)
  })

  it('collapses only TODO branches in the current open status view', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-not-started', title: '整理文档', status: 'not-started' },
          { id: 'todo-in-progress', title: '修复登录问题', status: 'in-progress' }
        ],
        todoProjects: [
          { id: 'todo-project-not-started', todoId: 'todo-not-started', projectId: 'project-a' },
          { id: 'todo-project-in-progress', todoId: 'todo-in-progress', projectId: 'project-a' }
        ],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="collapse-all-todos"]').trigger('click')
    expect(wrapper.find('[data-testid="todo-project-list-todo-not-started"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    expect(wrapper.find('[data-testid="todo-project-list-todo-in-progress"]').exists()).toBe(true)
  })

  it('keeps global candidate management out of the sidebar workspace', () => {
    const wrapper = mountSidebar()

    expect(wrapper.find('[data-testid="todo-workspace"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="sidebar-tab-projects"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="project-library"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="new-project"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="import-parent-directory"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="bulk-delete-projects"]').exists()).toBe(false)
    expect(wrapper.emitted('create-project')).toBeUndefined()
    expect(wrapper.emitted('import-projects')).toBeUndefined()
    expect(wrapper.emitted('select-project')).toBeUndefined()
    expect(wrapper.emitted('create-terminal')).toBeUndefined()
  })

  it('renders compact TODO project rows from the workspace copy when the global candidate is gone', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        projects: [],
        todoProjects: [
          {
            id: 'todo-project-a',
            todoId: 'todo-a',
            projectId: 'project-a',
            sourceProjectId: 'project-a',
            name: 'alpha-copy',
            path: '/work/alpha-copy',
            available: true
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    const todoProjectRow = wrapper.find('[data-testid="todo-project-todo-project-a"]')
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toBe('alpha-copy')
    expect(todoProjectRow.find('.lucide-folder-git-2').exists()).toBe(true)
    expect(todoProjectRow.find('.lucide-terminal-square').exists()).toBe(false)
    expect(todoProjectRow.text()).not.toContain('/work/alpha-copy')
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').exists()).toBe(true)
  })

  it('keeps TODO project status visible without restoring path text', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        projects: [],
        todoProjects: [
          {
            id: 'todo-project-a',
            todoId: 'todo-a',
            projectId: 'project-a',
            sourceProjectId: 'project-a',
            name: 'alpha-copy',
            path: '/work/alpha-copy',
            available: false,
            worktreeStatus: 'failed',
            worktreeError: 'Worktree failed'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    const todoProjectRow = wrapper.find('[data-testid="todo-project-todo-project-a"]')
    expect(todoProjectRow.text()).toContain('alpha-copy')
    expect(todoProjectRow.text()).toContain('Unavailable')
    expect(todoProjectRow.text()).toContain('Worktree failed')
    expect(todoProjectRow.text()).not.toContain('/work/alpha-copy')
  })

  it('collapses and expands a TODO branch independently', async () => {
    const wrapper = mountSidebar()

    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="toggle-todo-todo-a"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.emitted('select-todo-project')).toBeUndefined()

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-todo-todo-a"]').attributes('aria-expanded')).toBe('true')
  })

  it('emits todo-expanded only when a collapsed TODO branch expands', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')
    expect(wrapper.emitted('todo-expanded')).toBeUndefined()

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('todo-expanded')).toEqual([['todo-a']])
  })

  it('collapses all active TODO branches from the active list toolbar', async () => {
    const wrapper = mountSidebar({
      props: multiTodoProps()
    })

    const collapseAll = wrapper.find('[data-testid="collapse-all-todos"]')
    expect(collapseAll.attributes('aria-label')).toBe('Collapse all TODOs')
    expect(collapseAll.attributes('title')).toBe('Collapse all TODOs')

    await collapseAll.trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-todo-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="toggle-todo-todo-a"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.find('[data-testid="toggle-todo-todo-b"]').attributes('aria-expanded')).toBe('false')
  })

  it('expands all active TODO branches from the active list toolbar', async () => {
    const wrapper = mountSidebar({
      props: multiTodoProps()
    })

    await wrapper.find('[data-testid="collapse-all-todos"]').trigger('click')
    await wrapper.find('[data-testid="expand-all-todos"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-todo-todo-a"]').attributes('aria-expanded')).toBe('true')
    expect(wrapper.find('[data-testid="toggle-todo-todo-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('disables bulk TODO branch controls when there are no active TODOs', () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-completed',
            title: '已完成任务',
            status: 'completed',
            completedAt: '2026-06-10T10:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(wrapper.find('[data-testid="collapse-all-todos"]').attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="expand-all-todos"]').attributes('disabled')).toBeDefined()
  })

  it('shows completed TODO snapshot branches with merge status icons without terminal launch controls', async () => {
    const wrapper = mountSidebar({
      props: {
        completedMergeStatuses: {
          'todo-completed::project-a::/work/archived-alpha::0': { id: 'todo-completed::project-a::/work/archived-alpha::0', status: 'merged' },
          'todo-completed::project-b::/work/archived-beta::1': { id: 'todo-completed::project-b::/work/archived-beta::1', status: 'unmerged' }
        },
        todos: [
          { id: 'todo-a', title: '修复登录问题', status: 'active' },
          {
            id: 'todo-completed',
            title: '已完成任务',
            status: 'completed',
            completedAt: '2026-06-10T10:00:00Z',
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'archived-alpha',
                path: '/work/archived-alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              },
              {
                projectId: 'project-b',
                name: 'archived-beta',
                path: '/work/archived-beta',
                worktreeBranch: 'feature/payments',
                baseBranch: 'release/2026'
              },
              {
                projectId: 'project-legacy',
                name: 'legacy-alpha',
                path: '/work/legacy-alpha'
              }
            ]
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(wrapper.find('[data-testid="completed-todos"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('已完成任务')
    expect(wrapper.text()).toContain('completed')
    expect(wrapper.text()).toContain('feature/login -> main')
    expect(wrapper.text()).toContain('feature/payments -> release/2026')
    expect(wrapper.text()).toContain('Unknown branch -> Unknown base')
    expect(wrapper.text()).not.toContain('/work/archived-alpha')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-a-0"]').classes()).toContain('merged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-a-0"]').attributes('title')).toBe('Merged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-b-1"]').classes()).toContain('unmerged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-b-1"]').attributes('title')).toBe('Not merged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-legacy-2"]').classes()).toContain('unknown')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-legacy-2"]').attributes('title')).toBe('Merge status unknown')
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').exists()).toBe(false)
  })

  it('renders completed snapshot merge status as checking while async results load', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-a', title: '修复登录问题', status: 'active' },
          {
            id: 'todo-completed',
            title: '已完成任务',
            status: 'completed',
            completedAt: '2026-06-10T10:00:00Z',
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'archived-alpha',
                path: '/work/archived-alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    const status = wrapper.find('[data-testid="completed-project-merge-status-todo-completed-project-a-0"]')
    expect(status.classes()).toContain('checking')
    expect(status.attributes('title')).toBe('Checking merge status')
  })

  it('shows terminal activity in the TODO tree', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            activityState: 'needs-input',
            state: 'running'
          }
        ]
      }
    })

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    const activity = wrapper.find('[data-testid="terminal-activity-terminal-a"]')
    expect(terminalRow.text()).toContain('codex')
    expect(terminalRow.classes()).toContain('activity-needs-input')
    expect(terminalRow.attributes('data-activity-state')).toBe('needs-input')
    expect(terminalRow.attributes('aria-label')).toContain('Needs input')
    expect(activity.classes()).toContain('needs-input')
  })

  it('derives terminal activity from unified agent status when activityState is absent', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            agentStatus: { phase: 'needs-input' },
            state: 'running'
          }
        ]
      }
    })

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.attributes('data-activity-state')).toBe('needs-input')
    expect(terminalRow.attributes('aria-label')).toContain('Needs input')
  })

  it('shows terminal confirmation state in the TODO tree', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            attentionState: 'needs-ack',
            state: 'running'
          }
        ]
      }
    })

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    const activity = wrapper.find('[data-testid="terminal-activity-terminal-a"]')
    expect(terminalRow.classes()).toContain('activity-needs-ack')
    expect(terminalRow.attributes('data-activity-state')).toBe('needs-ack')
    expect(terminalRow.attributes('aria-label')).toContain('Review needed')
    expect(activity.classes()).toContain('needs-ack')
    expect(activity.find('svg').exists()).toBe(true)
  })

  it('shows hidden terminal activity as row state on a collapsed TODO', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            activityState: 'needs-input',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    const todoRow = wrapper.find('[data-testid="todo-todo-a"]')
    const todoHeader = todoHeaderFor(wrapper, 'todo-a')
    expect(todoRow.attributes('data-activity-state')).toBe('needs-input')
    expect(todoRow.attributes('title')).toBe('Needs input')
    expect(todoHeader.attributes('data-activity-state')).toBe('needs-input')
    expect(todoHeader.classes()).toContain('todo-activity-needs-input')
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
  })

  it('shows hidden busy terminal activity as a lighter collapsed TODO row state', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            activityState: 'busy',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    const todoHeader = todoHeaderFor(wrapper, 'todo-a')
    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('busy')
    expect(todoHeader.attributes('data-activity-state')).toBe('busy')
    expect(todoHeader.classes()).toContain('todo-activity-busy')
    expect(todoHeader.classes()).not.toContain('todo-activity-needs-input')
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
  })

  it('shows hidden confirmation state as urgent collapsed TODO row state', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            attentionState: 'needs-ack',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    const todoHeader = todoHeaderFor(wrapper, 'todo-a')
    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('needs-ack')
    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('title')).toBe('Review needed')
    expect(todoHeader.attributes('data-activity-state')).toBe('needs-ack')
    expect(todoHeader.classes()).toContain('todo-activity-needs-ack')
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
  })

  it('prioritizes confirmation over busy for a collapsed TODO', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-busy',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            activityState: 'busy',
            state: 'running'
          },
          {
            id: 'terminal-needs-ack',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            attentionState: 'needs-ack',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('needs-ack')
    expect(todoHeaderFor(wrapper, 'todo-a').classes()).toContain('todo-activity-needs-ack')
    expect(todoHeaderFor(wrapper, 'todo-a').classes()).not.toContain('todo-activity-busy')
  })

  it('prioritizes needs input over busy for a collapsed TODO', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-busy',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            activityState: 'busy',
            state: 'running'
          },
          {
            id: 'terminal-needs-input',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            activityState: 'needs-input',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('needs-input')
    expect(todoHeaderFor(wrapper, 'todo-a').attributes('data-activity-state')).toBe('needs-input')
    expect(todoHeaderFor(wrapper, 'todo-a').classes()).toContain('todo-activity-needs-input')
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
  })

  it('prioritizes needs input over confirmation for a collapsed TODO', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-needs-ack',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            attentionState: 'needs-ack',
            state: 'running'
          },
          {
            id: 'terminal-needs-input',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            activityState: 'needs-input',
            state: 'running'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('needs-input')
    expect(todoHeaderFor(wrapper, 'todo-a').classes()).toContain('todo-activity-needs-input')
    expect(todoHeaderFor(wrapper, 'todo-a').classes()).not.toContain('todo-activity-needs-ack')
  })

  it('does not keep done failed or exited agent statuses busy in collapsed TODO summaries', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-done',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            agentStatus: { phase: 'done' },
            state: 'running'
          },
          {
            id: 'terminal-failed',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            agentStatus: { phase: 'failed' },
            state: 'running'
          },
          {
            id: 'terminal-exited',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            agentStatus: { phase: 'exited' },
            state: 'exited'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('idle')
    expect(todoHeaderFor(wrapper, 'todo-a').attributes('data-activity-state')).toBeUndefined()
  })

  it('shows activity on terminal rows instead of the parent TODO while expanded', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            activityState: 'busy',
            state: 'running'
          }
        ]
      }
    })

    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBeUndefined()
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')
  })

  it('shows confirmation on terminal rows instead of the parent TODO while expanded', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            attentionState: 'needs-ack',
            state: 'running'
          }
        ]
      }
    })

    expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBeUndefined()
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('needs-ack')
  })

  it('does not label a collapsed TODO without terminals as idle', async () => {
    const wrapper = mountSidebar({
      props: {
        terminals: []
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')

    const todoRow = wrapper.find('[data-testid="todo-todo-a"]')
    expect(todoRow.attributes('data-activity-state')).toBeUndefined()
    expect(todoRow.attributes('title')).toBeUndefined()
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').exists()).toBe(false)
  })

  it('renders TODO description summary and priority styling without a default tooltip', () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '修复登录问题',
            description: '登录后跳回首页',
            priority: 'high',
            status: 'active'
          }
        ]
      }
    })

    const todoRow = wrapper.find('[data-testid="todo-todo-a"]')
    const todoHeader = todoRow.element.closest('.todo-header-row')
    expect(todoHeader.classList.contains('todo-header-row-priority-high')).toBe(true)
    expect(wrapper.find('[data-testid="todo-priority-todo-a"]').exists()).toBe(false)
    expect(todoHeader.textContent).not.toContain('高')
    expect(wrapper.find('[data-testid="todo-description-todo-a"]').text()).toBe('登录后跳回首页')
    expect(wrapper.find('[data-testid="todo-description-tooltip-todo-a"]').exists()).toBe(false)
  })

  it('shows the full TODO description tooltip after the hover delay', async () => {
    vi.useFakeTimers()
    const description = '登录后跳回首页，需要保留原始跳转地址'
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '修复登录问题',
            description,
            priority: 'high',
            status: 'active'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('mouseenter', { clientX: 120, clientY: 80 })
    vi.advanceTimersByTime(599)
    await nextTick()
    expect(wrapper.find('[data-testid="todo-description-tooltip-todo-a"]').exists()).toBe(false)

    vi.advanceTimersByTime(1)
    await nextTick()

    const tooltip = document.body.querySelector('[data-testid="todo-description-tooltip-todo-a"]')
    expect(tooltip).not.toBeNull()
    expect(tooltip.textContent).toBe(description)
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('renders the TODO description tooltip in a non-visual top-level layer', async () => {
    vi.useFakeTimers()
    const description = '登录后跳回首页，需要保留原始跳转地址'
    const wrapper = mountSidebar({
      attachTo: document.body,
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '修复登录问题',
            description,
            priority: 'high',
            status: 'active'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('mouseenter', { clientX: 120, clientY: 80 })
    vi.advanceTimersByTime(600)
    await nextTick()

    const tooltip = document.body.querySelector('[data-testid="todo-description-tooltip-todo-a"]')
    expect(tooltip).not.toBeNull()
    expect(tooltip.parentElement.classList.contains('todo-description-tooltip-layer')).toBe(true)
    expect(tooltip.parentElement.classList.contains('app-shell')).toBe(false)
    expect(tooltip.parentElement.parentElement).toBe(document.body)
    expect(wrapper.find('[data-testid="todo-todo-a"]').element.contains(tooltip)).toBe(false)
    wrapper.unmount()
  })

  it('hides the TODO description tooltip on mouse leave', async () => {
    vi.useFakeTimers()
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '修复登录问题',
            description: '登录后跳回首页',
            priority: 'high',
            status: 'active'
          }
        ]
      }
    })

    const todoRow = wrapper.find('[data-testid="todo-todo-a"]')
    await todoRow.trigger('mouseenter', { clientX: 120, clientY: 80 })
    vi.advanceTimersByTime(600)
    await nextTick()
    expect(document.body.querySelector('[data-testid="todo-description-tooltip-todo-a"]')).not.toBeNull()

    await todoRow.trigger('mouseleave')

    expect(document.body.querySelector('[data-testid="todo-description-tooltip-todo-a"]')).toBeNull()
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('does not show a TODO description tooltip without a description', async () => {
    vi.useFakeTimers()
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '整理文档',
            description: '',
            priority: 'low',
            status: 'active'
          }
        ]
      }
    })

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('mouseenter')
    vi.advanceTimersByTime(1000)
    await nextTick()

    expect(document.body.querySelector('[data-testid="todo-description-tooltip-todo-a"]')).toBeNull()
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('orders active TODOs by priority from high to low', () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-low',
            title: '整理文档',
            priority: 'low',
            status: 'active',
            createdAt: '2026-06-10T09:00:00Z'
          },
          {
            id: 'todo-high',
            title: '修复登录问题',
            priority: 'high',
            status: 'active',
            createdAt: '2026-06-10T10:00:00Z'
          },
          {
            id: 'todo-medium',
            title: '升级依赖',
            priority: 'medium',
            status: 'active',
            createdAt: '2026-06-10T08:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(activeTodoTitles(wrapper)).toEqual(['修复登录问题', '升级依赖', '整理文档'])
  })

  it('shows active TODO sort controls with priority selected by default', () => {
    const wrapper = mountSidebar()

    expect(wrapper.find('[data-testid="sort-active-todos-priority"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('[data-testid="sort-active-todos-time"]').attributes('aria-pressed')).toBe('false')
  })

  it('orders active TODOs with the same priority by creation time', () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-newer-high',
            title: '排查线上报警',
            priority: 'high',
            status: 'active',
            createdAt: '2026-06-10T11:00:00Z'
          },
          {
            id: 'todo-older-high',
            title: '修复登录问题',
            priority: 'high',
            status: 'active',
            createdAt: '2026-06-10T09:00:00Z'
          },
          {
            id: 'todo-low',
            title: '整理文档',
            priority: 'low',
            status: 'active',
            createdAt: '2026-06-10T08:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(activeTodoTitles(wrapper)).toEqual(['修复登录问题', '排查线上报警', '整理文档'])
  })

  it('switches active TODOs to creation time order', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-high-newer',
            title: '修复登录问题',
            priority: 'high',
            status: 'active',
            createdAt: '2026-06-10T11:00:00Z'
          },
          {
            id: 'todo-low-older',
            title: '整理文档',
            priority: 'low',
            status: 'active',
            createdAt: '2026-06-10T08:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    expect(activeTodoTitles(wrapper)).toEqual(['修复登录问题', '整理文档'])

    await wrapper.find('[data-testid="sort-active-todos-time"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="sort-active-todos-priority"]').attributes('aria-pressed')).toBe('false')
    expect(wrapper.find('[data-testid="sort-active-todos-time"]').attributes('aria-pressed')).toBe('true')
    expect(activeTodoTitles(wrapper)).toEqual(['整理文档', '修复登录问题'])
  })

  it('keeps completed TODOs unaffected by active TODO priority sorting', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-low',
            title: '整理文档',
            priority: 'low',
            status: 'active',
            createdAt: '2026-06-10T08:00:00Z'
          },
          {
            id: 'todo-high',
            title: '修复登录问题',
            priority: 'high',
            status: 'active',
            createdAt: '2026-06-10T09:00:00Z'
          },
          {
            id: 'todo-archived-low',
            title: '旧的低优先级归档',
            priority: 'low',
            status: 'completed',
            completedAt: '2026-06-10T12:00:00Z'
          },
          {
            id: 'todo-archived-high',
            title: '旧的高优先级归档',
            priority: 'high',
            status: 'completed',
            completedAt: '2026-06-10T11:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoTitles(wrapper)).toEqual(['旧的低优先级归档', '旧的高优先级归档'])
  })

  it('orders completed TODOs by newest completedAt first', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-older-completed',
            title: '整理文档',
            status: 'completed',
            completedAt: '2026-06-14T09:00:00Z'
          },
          {
            id: 'todo-newer-completed',
            title: '修复登录问题',
            status: 'completed',
            completedAt: '2026-06-15T09:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoTitles(wrapper)).toEqual(['修复登录问题', '整理文档'])
  })

  it('orders completed TODOs by archivedAt when completedAt is missing or invalid', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-completed-at',
            title: '较早任务',
            status: 'completed',
            completedAt: '2026-06-15T09:00:00Z'
          },
          {
            id: 'todo-archived-at',
            title: '旧任务',
            status: 'completed',
            archivedAt: '2026-06-15T10:00:00Z'
          },
          {
            id: 'todo-invalid-completed-at',
            title: '异常旧任务',
            status: 'completed',
            completedAt: 'not-a-date',
            archivedAt: '2026-06-15T11:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoTitles(wrapper)).toEqual(['异常旧任务', '旧任务', '较早任务'])
  })

  it('orders completed TODOs without valid completion time last', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-missing-completion-time',
            title: '缺失时间任务',
            status: 'completed'
          },
          {
            id: 'todo-invalid-completion-time',
            title: '无效时间任务',
            status: 'completed',
            completedAt: 'not-a-date'
          },
          {
            id: 'todo-valid-completion-time',
            title: '有时间任务',
            status: 'completed',
            completedAt: '2026-06-15T09:00:00Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoTitles(wrapper)).toEqual(['有时间任务', '缺失时间任务', '无效时间任务'])
  })

  it('shows completed TODO duration from startedAt to completedAt', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-completed-duration',
            title: '修复登录问题',
            status: 'completed',
            startedAt: '2026-06-22T01:00:00Z',
            completedAt: '2026-06-22T02:15:30Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoMeta(wrapper, 'todo-completed-duration')).toContain('Duration 1h 15m')
  })

  it('does not infer completed TODO duration from createdAt when startedAt is missing', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-completed-history',
            title: '历史任务',
            status: 'completed',
            createdAt: '2026-06-01T01:00:00Z',
            completedAt: '2026-06-22T02:15:30Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoMeta(wrapper, 'todo-completed-history')).not.toContain('Duration')
  })

  it('does not show negative completed TODO duration', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-completed-invalid-duration',
            title: '异常任务',
            status: 'completed',
            startedAt: '2026-06-22T03:00:00Z',
            completedAt: '2026-06-22T02:15:30Z'
          }
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      }
    })

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(completedTodoMeta(wrapper, 'todo-completed-invalid-duration')).not.toContain('Duration')
    expect(completedTodoMeta(wrapper, 'todo-completed-invalid-duration')).not.toContain('Duration -')
  })

  it('opens TODO action confirmation popovers before emitting', async () => {
    const wrapper = mountInProgressSidebar()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(wrapper.emitted('complete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="cancel-complete-todo-todo-a"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()
  })

  it('anchors the TODO delete confirmation popover to the TODO action control', async () => {
    const wrapper = mountInProgressSidebar()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await nextTick()

    const popover = wrapper.find('[data-testid="delete-todo-popover-todo-a"]')

    expect(popover.exists()).toBe(true)
    const actionControl = popover.element.closest('.todo-action-confirm-control')
    expect(actionControl).not.toBeNull()
    expect(actionControl.closest('[data-testid="todo-actions-todo-a"]')).not.toBeNull()
  })

  it('confirms and cancels TODO action popovers', async () => {
    const wrapper = mountInProgressSidebar()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="cancel-complete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.emitted('complete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('complete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)

    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-delete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('delete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(false)
  })

  it('closes TODO action popovers from outside clicks and other sidebar popovers', async () => {
    const wrapper = mountInProgressSidebar({
      props: {
        launchProfiles: [{ name: 'codex', command: 'codex' }]
      }
    })

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    window.dispatchEvent(new MouseEvent('click'))
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]').exists()).toBe(true)

    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-delete-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="remove-todo-project-popover-todo-project-a"]').exists()).toBe(true)
  })

  it('confirms direct TODO project removal from a popover', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="remove-todo-project-popover-todo-project-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="cancel-remove-todo-project-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="remove-todo-project-popover-todo-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-remove-todo-project-todo-project-a"]').trigger('click')

    expect(wrapper.emitted('remove-todo-project')[0]).toEqual(['todo-project-a'])
  })

  it('expands a collapsed TODO when the active terminal moves under it', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          { id: 'todo-a', title: '修复登录问题', status: 'active' },
          { id: 'todo-b', title: '升级依赖', status: 'active' }
        ],
        todoProjects: [
          { id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' },
          { id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-a' }
        ],
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            todoId: 'todo-a',
            todoProjectId: 'todo-project-a',
            shellName: 'zsh',
            currentCommand: '',
            state: 'running'
          },
          {
            id: 'terminal-b',
            projectId: 'project-a',
            todoId: 'todo-b',
            todoProjectId: 'todo-project-b',
            shellName: 'bash',
            currentCommand: '',
            state: 'running'
          }
        ],
        activeTodoId: 'todo-a',
        activeTodoProjectId: 'todo-project-a',
        activeTerminalId: 'terminal-a'
      }
    })

    await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
    await wrapper.setProps({
      activeTodoId: 'todo-b',
      activeTodoProjectId: 'todo-project-b',
      activeTerminalId: 'terminal-b'
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-todo-todo-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('expands only the active TODO after bulk collapse when active context changes', async () => {
    const wrapper = mountSidebar({
      props: multiTodoProps()
    })

    await wrapper.find('[data-testid="collapse-all-todos"]').trigger('click')
    await wrapper.setProps({
      activeTodoId: 'todo-b',
      activeTodoProjectId: 'todo-project-b',
      activeTerminalId: 'terminal-b'
    })
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-todo-todo-a"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.find('[data-testid="toggle-todo-todo-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('uses compact TODO project row sizing without changing generic project rows', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const genericProjectRowIndex = styles.indexOf('.project-header-row')
    const todoProjectRowIndex = styles.indexOf('.todo-project-header-row {')
    const todoProjectButtonIndex = styles.indexOf('.todo-project-header-row .project-row')

    expect(todoProjectRowIndex).toBeGreaterThan(genericProjectRowIndex)
    expect(styles.slice(todoProjectRowIndex, todoProjectRowIndex + 120)).toContain(
      'grid-template-columns: minmax(0, 1fr) 30px 30px;'
    )
    expect(todoProjectButtonIndex).toBeGreaterThan(todoProjectRowIndex)
    expect(styles.slice(todoProjectButtonIndex, todoProjectButtonIndex + 260)).toContain('min-height: 30px;')
    expect(styles).toContain('.todo-project-header-row:hover')
    expect(styles).toContain('.todo-project-node.is-active-project > .todo-project-header-row')
    expect(styles).toContain('.todo-project-header-row .project-row.active')
  })

  it('keeps removed project library styles out of the sidebar CSS', () => {
    const styles = readFileSync('src/style.css', 'utf8')

    expect(styles).toContain('.candidate-management-toolbar')
    expect(styles).not.toContain('.library-project-header-row')
    expect(styles).not.toContain('.project-select-checkbox')
    expect(styles).not.toContain('.project-delete-control')
    expect(styles).not.toContain('.project-delete-popover')
    expect(styles).not.toContain('.bulk-project-delete-control')
    expect(styles).not.toContain('.library-action-button')
  })

  it('defines compact bulk TODO branch control styles', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const toolbarRule = styles.slice(styles.indexOf('.todo-tree-toolbar {'), styles.indexOf('.todo-sort-toggle'))

    expect(styles).toContain('.todo-tree-toolbar')
    expect(toolbarRule).toContain('min-height: 34px;')
    expect(styles).toContain('.todo-tree-action')
    expect(styles).toContain('.todo-tree-action:disabled')
  })

  it('defines priority row styles for each priority level', () => {
    const styles = readFileSync('src/style.css', 'utf8')

    expect(styles).toContain('--todo-priority-high-bg')
    expect(styles).toContain('--todo-priority-medium-bg')
    expect(styles).toContain('--todo-priority-low-bg')
    expect(styles).toContain('.todo-header-row-priority-high')
    expect(styles).toContain('.todo-header-row-priority-medium')
    expect(styles).toContain('.todo-header-row-priority-low')
  })

  it('defines row-level breathing styles for collapsed TODO activity states', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const busyRule = styles.slice(styles.indexOf('.todo-header-row.todo-activity-busy {'), styles.indexOf('.todo-header-row.todo-activity-needs-input {'))
    const needsInputRule = styles.slice(styles.indexOf('.todo-header-row.todo-activity-needs-input {'), styles.indexOf('@keyframes todo-activity-busy-breathe'))
    const reducedMotionRule = styles.slice(styles.indexOf('@media (prefers-reduced-motion: reduce)'), styles.indexOf('.todo-title-line {'))

    expect(styles).toContain('.todo-header-row.todo-activity-busy')
    expect(styles).toContain('.todo-header-row.todo-activity-needs-input')
    expect(busyRule).toContain('animation: todo-activity-busy-breathe')
    expect(busyRule).toContain('1.7s')
    expect(busyRule).toContain('rgba(15, 118, 110, 0.62)')
    expect(needsInputRule).toContain('animation: todo-activity-needs-input-breathe')
    expect(needsInputRule).toContain('1.3s')
    expect(needsInputRule).toContain('rgba(154, 91, 23, 0.82)')
    expect(busyRule).not.toContain('background:')
    expect(busyRule).not.toMatch(/(^|\n)\s*color:/)
    expect(needsInputRule).not.toContain('background:')
    expect(needsInputRule).not.toMatch(/(^|\n)\s*color:/)
    expect(styles).toContain('@keyframes todo-activity-busy-breathe')
    expect(styles).toContain('@keyframes todo-activity-needs-input-breathe')
    expect(styles).toContain('rgba(15, 118, 110, 0.18), 0 0 0 4px rgba(15, 118, 110, 0.34)')
    expect(styles).toContain('rgba(154, 91, 23, 0.24), 0 0 0 4px rgba(154, 91, 23, 0.44)')
    expect(reducedMotionRule).toContain('animation: none;')
    expect(reducedMotionRule).toContain('.todo-header-row.todo-activity-busy')
    expect(reducedMotionRule).toContain('.todo-header-row.todo-activity-needs-input')
  })

  it('defines terminal and collapsed TODO confirmation styles', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const ackRule = styles.slice(styles.indexOf('.todo-header-row.todo-activity-needs-ack {'), styles.indexOf('@keyframes todo-activity-busy-breathe'))
    const reducedMotionRule = styles.slice(styles.indexOf('@media (prefers-reduced-motion: reduce)'), styles.indexOf('.todo-title-line {'))
    const taskTerminalGroupRule = styles.slice(styles.indexOf('.task-terminal-group {'), styles.indexOf('.terminal-list {'))
    const terminalConnectorIndex = styles.indexOf('.terminal-row::before {')
    const taskTerminalConnectorIndex = styles.indexOf('.task-terminal-row::before {')
    const taskTerminalConnectorRule = styles.slice(taskTerminalConnectorIndex, styles.indexOf('.terminal-row:hover'))
    const terminalRowRule = styles.slice(styles.indexOf('.terminal-row {'), styles.indexOf('.terminal-row::before'))
    const taskTerminalRowRule = styles.slice(styles.indexOf('.task-terminal-row {'), styles.indexOf('.task-terminal-row:hover'))

    expect(styles).toContain('.terminal-row.activity-needs-ack')
    expect(styles).toContain('.terminal-activity.needs-ack')
    expect(taskTerminalGroupRule).not.toContain('.task-terminal-header')
    expect(taskTerminalGroupRule).toContain('padding-left: 20px;')
    expect(taskTerminalConnectorIndex).toBeGreaterThan(-1)
    expect(taskTerminalConnectorIndex).toBeGreaterThan(terminalConnectorIndex)
    expect(taskTerminalConnectorRule).toContain('left: -27px;')
    expect(taskTerminalConnectorRule).toContain('width: 28px;')
    expect(styles).not.toContain('.todo-project-node.has-terminals::before')
    expect(terminalRowRule).toContain('min-height: 30px;')
    expect(taskTerminalRowRule).toContain('min-height: 30px;')
    expect(taskTerminalRowRule).not.toContain('background:')
    expect(styles).toContain('.todo-header-row.todo-activity-needs-ack')
    expect(ackRule).toContain('animation: todo-activity-needs-ack-breathe')
    expect(ackRule).toContain('0.9s')
    expect(ackRule).toContain('rgba(124, 58, 237')
    expect(ackRule).not.toContain('rgba(15, 118, 110')
    expect(ackRule).not.toContain('background:')
    expect(ackRule).not.toMatch(/(^|\n)\s*color:/)
    expect(styles).toContain('@keyframes todo-activity-needs-ack-breathe')
    expect(reducedMotionRule).toContain('.todo-header-row.todo-activity-needs-ack')
    expect(reducedMotionRule).toContain('animation: none;')
  })

  it('defines a wide TODO description tooltip style', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const tooltipLayerRule = styles.slice(styles.indexOf('.todo-description-tooltip-layer {'), styles.indexOf('.todo-description-tooltip {'))
    const tooltipRule = styles.slice(styles.indexOf('.todo-description-tooltip {'), styles.indexOf('.todo-actions {'))

    expect(tooltipLayerRule).toContain('position: fixed;')
    expect(tooltipLayerRule).toContain('inset: 0;')
    expect(tooltipLayerRule).toContain('z-index: 30;')
    expect(tooltipLayerRule).toContain('pointer-events: none;')
    expect(tooltipRule).toContain('position: fixed;')
    expect(tooltipRule).toContain('width: min(520px, 72vw);')
    expect(tooltipRule).not.toContain('max-width: min(320px, calc(100% - 34px));')
  })

  it('opens TODO context menu from the three-dot action button', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-menu-button-todo-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="todo-context-menu-todo-a"]')
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('View details')
    expect(menu.text()).toContain('Add project')
    expect(menu.text()).toContain('Copy title and description')
    expect(menu.text()).toContain('Delete TODO')
  })

  it('shares menu actions between right-click and three-dot button', async () => {
    const wrapper = mountSidebar()

    // Open via right-click and verify actions work.
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-edit-todo-a"]').trigger('click')
    expect(wrapper.emitted('edit-todo')[0]).toEqual(['todo-a'])

    // Open via three-dot button and verify same menu appears.
    await wrapper.find('[data-testid="todo-menu-button-todo-a"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(true)
    await wrapper.find('[data-testid="todo-menu-add-project-todo-a"]').trigger('click')
    expect(wrapper.emitted('add-project-to-todo')[0]).toEqual(['todo-a'])
  })

  it('closes the three-dot menu on outside click', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-menu-button-todo-a"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(true)

    window.dispatchEvent(new MouseEvent('click'))
    await nextTick()
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
  })

  it('menu copy item shows title and description text', async () => {
    const wrapper = mountSidebar({
      props: {
        todos: [
          {
            id: 'todo-a',
            title: '修复登录问题',
            description: '登录后跳回首页',
            priority: 'high',
            status: 'active'
          }
        ]
      }
    })

    await openTodoContextMenu(wrapper, 'todo-a')

    const menu = wrapper.find('[data-testid="todo-context-menu-todo-a"]')
    expect(menu.text()).toContain('Copy title and description')
    expect(menu.text()).not.toContain('Copy description')
  })
})

function multiTodoProps() {
  return {
    projects: [
      { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
      { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
    ],
    todos: [
      { id: 'todo-a', title: '修复登录问题', status: 'active' },
      { id: 'todo-b', title: '升级依赖', status: 'active' }
    ],
    todoProjects: [
      { id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' },
      { id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b' }
    ],
    terminals: [
      {
        id: 'terminal-a',
        projectId: 'project-a',
        todoId: 'todo-a',
        todoProjectId: 'todo-project-a',
        shellName: 'zsh',
        currentCommand: '',
        state: 'running'
      },
      {
        id: 'terminal-b',
        projectId: 'project-b',
        todoId: 'todo-b',
        todoProjectId: 'todo-project-b',
        shellName: 'bash',
        currentCommand: '',
        state: 'running'
      }
    ],
    activeProjectId: 'project-a',
    activeTodoId: 'todo-a',
    activeTodoProjectId: 'todo-project-a',
    activeTerminalId: 'terminal-a'
  }
}

function activeTodoTitles(wrapper) {
  return wrapper
    .find('[data-testid="not-started-todos"]')
    .findAll('.todo-row .project-name')
    .map((node) => node.text())
}

function todoHeaderFor(wrapper, todoId) {
  const element = wrapper.find(`[data-testid="todo-${todoId}"]`).element.closest('.todo-header-row')
  return {
    attributes(name) {
      return element?.getAttribute(name) ?? undefined
    },
    classes() {
      return Array.from(element?.classList || [])
    }
  }
}

function visibleTodoTitles(wrapper, listTestId) {
  return wrapper
    .find(`[data-testid="${listTestId}"]`)
    .findAll('.todo-row .project-name')
    .map((node) => node.text())
}

function completedTodoTitles(wrapper) {
  return wrapper
    .find('[data-testid="completed-todos"]')
    .findAll('.completed-todo-title span')
    .map((node) => node.text())
}

function completedTodoMeta(wrapper, todoId) {
  return wrapper.find(`[data-testid="completed-todo-${todoId}"] .archived-todo-meta`).text()
}

async function openTodoContextMenu(wrapper, todoId) {
  await wrapper.find(`[data-testid="todo-${todoId}"]`).trigger('contextmenu', {
    clientX: 48,
    clientY: 64
  })
  await nextTick()
}

function mountSidebar(options = {}) {
  const mountOptions = {
    props: {
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
      todos: [
        { id: 'todo-a', title: '修复登录问题', status: 'active' },
        {
          id: 'todo-completed',
          title: '已完成任务',
          status: 'completed',
          completedAt: '2026-06-10T10:00:00Z',
          projectSnapshots: [{ projectId: 'project-a', name: 'archived-alpha', path: '/work/archived-alpha' }]
        }
      ],
      todoProjects: [
        {
          id: 'todo-project-a',
          todoId: 'todo-a',
          projectId: 'project-a',
          worktreeStatus: 'ready',
          worktreePath: '/work/customer-a/tasks/abc123/alpha'
        }
      ],
      terminals: [
        {
          id: 'terminal-a',
          projectId: 'project-a',
          todoId: 'todo-a',
          todoProjectId: 'todo-project-a',
          shellName: 'zsh',
          currentCommand: 'codex',
          state: 'running'
        }
      ],
      activeProjectId: 'project-a',
      activeTodoId: 'todo-a',
      activeTodoProjectId: 'todo-project-a',
      activeTerminalId: 'terminal-a',
      launchProfiles: [],
      ...(options.props || {})
    }
  }
  if (options.attachTo) {
    mountOptions.attachTo = options.attachTo
  }
  return mount(ProjectSidebar, mountOptions)
}

function mountInProgressSidebar(options = {}) {
  return mountSidebar({
    ...options,
    props: {
      todos: [
        { id: 'todo-a', title: '修复登录问题', status: 'in-progress' },
        {
          id: 'todo-completed',
          title: '已完成任务',
          status: 'completed',
          completedAt: '2026-06-10T10:00:00Z',
          projectSnapshots: [{ projectId: 'project-a', name: 'archived-alpha', path: '/work/archived-alpha' }]
        }
      ],
      ...(options.props || {})
    }
  })
}
