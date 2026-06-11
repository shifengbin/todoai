import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { nextTick } from 'vue'
import { describe, expect, it } from 'vitest'
import ProjectSidebar from './ProjectSidebar.vue'

describe('ProjectSidebar', () => {
  it('renders the TODO tree and emits TODO-scoped terminal actions', async () => {
    const wrapper = mountSidebar({
      props: {
        launchProfiles: [{ name: 'codex', command: 'codex' }]
      }
    })

    expect(wrapper.find('[data-testid="workspace-tabs"]').classes()).toContain('tab-strip')
    expect(wrapper.find('[data-testid="sidebar-tab-todos"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="sidebar-tab-projects"]').text()).toBe('项目')
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
    await wrapper.find('[data-testid="edit-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="add-project-to-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')
    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-delete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('create-todo')).toHaveLength(1)
    expect(wrapper.emitted('edit-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('add-project-to-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('select-todo-project')[0]).toEqual(['todo-project-a'])
    expect(wrapper.emitted('create-terminal')[0]).toEqual(['todo-project-a', { name: 'codex', command: 'codex' }])
    expect(wrapper.emitted('select-terminal')[0]).toEqual(['terminal-a'])
    expect(wrapper.emitted('complete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.emitted('delete-todo')[0]).toEqual(['todo-a'])
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
    await wrapper.find('[data-testid="mark-todo-not-started-todo-b"]').trigger('click')

    expect(wrapper.emitted('change-todo-status')[0]).toEqual(['todo-a', 'in-progress'])
    expect(wrapper.emitted('change-todo-status')[1]).toEqual(['todo-b', 'not-started'])
  })

  it('groups TODO workflow tabs and item actions into single rows', () => {
    const wrapper = mountSidebar()
    const workflowTabs = wrapper.find('[data-testid="todo-workflow-tabs"]')
    const actionGroup = wrapper.find('[data-testid="todo-actions-todo-a"]')
    const styles = readFileSync('src/style.css', 'utf8')
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
    expect(actionTestIds).toEqual([
      'mark-todo-in-progress-todo-a',
      'edit-todo-todo-a',
      'add-project-to-todo-todo-a',
      'complete-todo-todo-a',
      'delete-todo-todo-a'
    ])
    expect(tabsRule).toContain('grid-template-columns: repeat(3, minmax(0, 1fr));')
    expect(actionsRule).toContain('display: inline-flex;')
    expect(actionsRule).toContain('flex-wrap: nowrap;')
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

  it('shows project library management without terminal actions', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')

    expect(wrapper.find('[data-testid="project-library"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="project-name-project-a"]').text()).toBe('alpha')
    expect(wrapper.text()).toContain('/work/alpha')
    expect(wrapper.find('[data-testid="add-terminal-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="new-project"]').trigger('click')
    await wrapper.find('[data-testid="import-parent-directory"]').trigger('click')
    await wrapper.find('[data-testid="project-project-a"]').trigger('click')

    expect(wrapper.emitted('create-project')).toHaveLength(1)
    expect(wrapper.emitted('import-projects')).toHaveLength(1)
    expect(wrapper.emitted('select-project')[0]).toEqual(['project-a'])
    expect(wrapper.emitted('create-terminal')).toBeUndefined()
  })

  it('confirms direct project removal from a popover', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-project-popover-project-a"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-project')).toBeUndefined()

    await wrapper.find('[data-testid="cancel-delete-project-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-project-popover-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-delete-project-project-a"]').trigger('click')

    expect(wrapper.emitted('delete-project')[0]).toEqual(['project-a'])
    expect(wrapper.find('[data-testid="delete-project-popover-project-a"]').exists()).toBe(false)
  })

  it('selects projects and confirms bulk deletion from a popover', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ]
      }
    })

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')

    const bulkDelete = wrapper.find('[data-testid="bulk-delete-projects"]')
    expect(bulkDelete.attributes('disabled')).toBeDefined()
    await bulkDelete.trigger('click')
    expect(wrapper.find('[data-testid="bulk-delete-projects-popover"]').exists()).toBe(false)

    await wrapper.find('[data-testid="select-project-project-a"]').trigger('click')
    await wrapper.find('[data-testid="select-project-project-b"]').trigger('click')

    expect(wrapper.emitted('select-project')).toBeUndefined()
    expect(wrapper.find('[data-testid="select-project-project-a"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="select-project-project-b"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="bulk-delete-projects"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('[data-testid="bulk-delete-projects"]').text()).toContain('2')

    await wrapper.find('[data-testid="bulk-delete-projects"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="bulk-delete-projects-popover"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-projects')).toBeUndefined()

    await wrapper.find('[data-testid="cancel-bulk-delete-projects"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="bulk-delete-projects-popover"]').exists()).toBe(false)

    await wrapper.find('[data-testid="bulk-delete-projects"]').trigger('click')
    await wrapper.find('[data-testid="confirm-bulk-delete-projects"]').trigger('click')

    expect(wrapper.emitted('delete-projects')[0]).toEqual([['project-a', 'project-b']])
    expect(wrapper.find('[data-testid="bulk-delete-projects-popover"]').exists()).toBe(false)
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

  it('shows completed TODO snapshots without terminal launch controls', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(wrapper.find('[data-testid="completed-todos"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('已完成任务')
    expect(wrapper.text()).toContain('completed')
    expect(wrapper.text()).toContain('/work/archived-alpha')
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').exists()).toBe(false)
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

  it('shows hidden terminal activity on a collapsed TODO', async () => {
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
    const activity = wrapper.find('[data-testid="todo-activity-todo-a"]')
    expect(todoRow.attributes('data-activity-state')).toBe('needs-input')
    expect(activity.classes()).toContain('needs-input')
    expect(activity.attributes('aria-label')).toBe('Needs input')
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
    expect(wrapper.find('[data-testid="todo-activity-todo-a"]').classes()).toContain('needs-input')
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

  it('renders TODO description and priority styling', () => {
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

  it('keeps completed TODOs in their existing order', async () => {
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

  it('opens TODO action confirmation popovers before emitting', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(wrapper.emitted('complete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="cancel-complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()
  })

  it('confirms and cancels TODO action popovers', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="cancel-complete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.emitted('complete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('complete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="confirm-delete-todo-todo-a"]').trigger('click')

    expect(wrapper.emitted('delete-todo')[0]).toEqual(['todo-a'])
    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(false)
  })

  it('closes TODO action popovers from outside clicks and other sidebar popovers', async () => {
    const wrapper = mountSidebar({
      props: {
        launchProfiles: [{ name: 'codex', command: 'codex' }]
      }
    })

    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
    window.dispatchEvent(new MouseEvent('click'))
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.emitted('delete-todo')).toBeUndefined()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
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

  it('keeps TODO project row layout wider than generic project rows', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const genericProjectRowIndex = styles.indexOf('.project-header-row')
    const todoProjectRowIndex = styles.indexOf('.todo-project-header-row {')

    expect(todoProjectRowIndex).toBeGreaterThan(genericProjectRowIndex)
    expect(styles.slice(todoProjectRowIndex, todoProjectRowIndex + 120)).toContain(
      'grid-template-columns: minmax(0, 1fr) 30px 30px;'
    )
    expect(styles).toContain('.todo-project-header-row:hover')
    expect(styles).toContain('.todo-project-node.is-active-project > .todo-project-header-row')
    expect(styles).toContain('.todo-project-header-row .project-row.active')
  })

  it('defines compact project library selection and delete styles', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const libraryProjectRowIndex = styles.indexOf('.library-project-header-row {')

    expect(libraryProjectRowIndex).toBeGreaterThan(-1)
    expect(styles.slice(libraryProjectRowIndex, libraryProjectRowIndex + 140)).toContain(
      'grid-template-columns: 26px minmax(0, 1fr) 30px;'
    )
    expect(styles).toContain('.project-select-checkbox')
    expect(styles).toContain('.project-delete-control')
    expect(styles).toContain('.project-delete-popover')
    expect(styles).toContain('.bulk-project-delete-control')
    expect(styles).toContain('.library-action-button-delete:disabled')
  })

  it('defines compact bulk TODO branch control styles', () => {
    const styles = readFileSync('src/style.css', 'utf8')

    expect(styles).toContain('.todo-tree-toolbar')
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

function mountSidebar(options = {}) {
  return mount(ProjectSidebar, {
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
      todoProjects: [{ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' }],
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
  })
}
