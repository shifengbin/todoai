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
    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')

    expect(wrapper.emitted('create-project')).toHaveLength(1)
    expect(wrapper.emitted('import-projects')).toHaveLength(1)
    expect(wrapper.emitted('select-project')[0]).toEqual(['project-a'])
    expect(wrapper.emitted('delete-project')[0]).toEqual(['project-a'])
    expect(wrapper.emitted('create-terminal')).toBeUndefined()
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

  it('shows archived TODO snapshots without terminal launch controls', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="todo-view-archived"]').trigger('click')

    expect(wrapper.find('[data-testid="archived-todos"]').exists()).toBe(true)
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

function mountSidebar(options = {}) {
  return mount(ProjectSidebar, {
    props: {
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
      todos: [
        { id: 'todo-a', title: '修复登录问题', status: 'active' },
        {
          id: 'todo-archived',
          title: '已完成任务',
          status: 'archived',
          archivedReason: 'completed',
          archivedAt: '2026-06-10T10:00:00Z',
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
