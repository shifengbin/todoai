import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { describe, expect, it } from 'vitest'
import ProjectSidebar from './ProjectSidebar.vue'

describe('ProjectSidebar', () => {
  it('renders a project terminal tree and emits tree actions', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: false }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          {
            id: 'terminal-b',
            projectId: 'project-a',
            shellName: 'zsh',
            currentCommand: 'npm run dev',
            state: 'running'
          }
        ],
        activeProjectId: 'project-a',
        activeTerminalId: 'terminal-b'
      }
    })

    expect(wrapper.text()).toContain('alpha')
    expect(wrapper.text()).toContain('/work/alpha')
    expect(wrapper.text()).toContain('beta')
    expect(wrapper.text()).toContain('Unavailable')
    expect(wrapper.text()).toContain('zsh')
    expect(wrapper.text()).toContain('npm run dev')
    expect(wrapper.find('[data-testid="project-project-a"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').classes()).toContain('active')

    await wrapper.find('[data-testid="new-project"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-project-a-0"]').trigger('click')
    await wrapper.find('[data-testid="project-project-b"]').trigger('click')
    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')

    expect(wrapper.emitted('create-project')).toHaveLength(1)
    expect(wrapper.emitted('create-terminal')[0]).toEqual(['project-a', null])
    expect(wrapper.emitted('select-project')[0]).toEqual(['project-b'])
    expect(wrapper.emitted('select-terminal')[0]).toEqual(['terminal-a'])
  })

  it('opens a terminal launch menu and emits the selected launch profile', async () => {
    const wrapper = mountSidebar({
      props: {
        launchProfiles: [
          { name: 'codex', command: 'codex' },
          { name: 'claude', command: 'claude' }
        ]
      }
    })

    await wrapper.find('[data-testid="add-terminal-project-a"]').trigger('click')

    const menu = wrapper.find('[data-testid="terminal-launch-menu-project-a"]')
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).toContain('codex')
    expect(menu.text()).toContain('claude')

    await wrapper.find('[data-testid="terminal-launch-option-project-a-1"]').trigger('click')

    expect(wrapper.emitted('create-terminal')[0]).toEqual(['project-a', { name: 'codex', command: 'codex' }])
    expect(wrapper.find('[data-testid="terminal-launch-menu-project-a"]').exists()).toBe(false)
  })

  it('closes the terminal launch menu on outside click and unavailable project changes', async () => {
    const wrapper = mountSidebar({
      props: { launchProfiles: [{ name: 'codex', command: 'codex' }] }
    })

    await wrapper.find('[data-testid="add-terminal-project-a"]').trigger('click')
    expect(wrapper.find('[data-testid="terminal-launch-menu-project-a"]').exists()).toBe(true)

    window.dispatchEvent(new MouseEvent('click'))
    await nextTick()
    expect(wrapper.find('[data-testid="terminal-launch-menu-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="add-terminal-project-a"]').trigger('click')
    await wrapper.setProps({
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: false }]
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-launch-menu-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-terminal-project-a"]').exists()).toBe(false)
  })

  it('collapses and expands a project terminal branch', async () => {
    const wrapper = mountSidebar()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="toggle-project-project-a"]').trigger('click')

    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="toggle-project-project-a"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.emitted('select-project')).toBeUndefined()

    await wrapper.find('[data-testid="toggle-project-project-a"]').trigger('click')

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-project-project-a"]').attributes('aria-expanded')).toBe('true')
  })

  it('emits delete actions without selecting rows', async () => {
    const wrapper = mountSidebar()

    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await wrapper.find('[data-testid="delete-terminal-terminal-a"]').trigger('click')

    expect(wrapper.emitted('delete-project')[0]).toEqual(['project-a'])
    expect(wrapper.emitted('delete-terminal')[0]).toEqual(['terminal-a'])
    expect(wrapper.emitted('select-project')).toBeUndefined()
    expect(wrapper.emitted('select-terminal')).toBeUndefined()
  })

  it('marks only the project branch that owns the active terminal', () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' }
        ],
        activeProjectId: 'project-b',
        activeTerminalId: 'terminal-b'
      }
    })

    const inactiveProjectNode = wrapper.find('[data-testid="project-project-a"]').element.closest('.project-node')
    const activeProjectNode = wrapper.find('[data-testid="project-project-b"]').element.closest('.project-node')

    expect(inactiveProjectNode.classList.contains('has-active-terminal')).toBe(false)
    expect(activeProjectNode.classList.contains('has-active-terminal')).toBe(true)
  })

  it('shows a busy activity indicator without replacing the terminal label', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
            shellName: 'zsh',
            currentCommand: 'codex',
            activityState: 'busy',
            state: 'running'
          }
        ]
      }
    })

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    const activity = wrapper.find('[data-testid="terminal-activity-terminal-a"]')
    expect(terminalRow.text()).toContain('codex')
    expect(terminalRow.classes()).toContain('activity-busy')
    expect(terminalRow.attributes('data-activity-state')).toBe('busy')
    expect(terminalRow.attributes('aria-label')).toContain('Running')
    expect(activity.exists()).toBe(true)
    expect(activity.classes()).toContain('busy')
    expect(activity.attributes('aria-label')).toBe('Running')
  })

  it('shows a needs-input activity indicator without showing the busy state', () => {
    const wrapper = mountSidebar({
      props: {
        terminals: [
          {
            id: 'terminal-a',
            projectId: 'project-a',
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
    expect(terminalRow.classes()).not.toContain('activity-busy')
    expect(terminalRow.attributes('data-activity-state')).toBe('needs-input')
    expect(terminalRow.attributes('aria-label')).toContain('Needs input')
    expect(activity.classes()).toContain('needs-input')
    expect(activity.attributes('aria-label')).toBe('Needs input')
  })

  it('keeps collapsed branch state independent per project', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' }
        ]
      }
    })

    await wrapper.find('[data-testid="toggle-project-project-a"]').trigger('click')

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-project-project-a"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.find('[data-testid="toggle-project-project-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('expands a collapsed branch when the active project changes to that project', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' }
        ],
        activeProjectId: 'project-a'
      }
    })

    await wrapper.find('[data-testid="toggle-project-project-b"]').trigger('click')
    await wrapper.setProps({ activeProjectId: 'project-b' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-project-project-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('expands a collapsed branch when the active terminal changes to a terminal under that project', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' }
        ],
        activeTerminalId: 'terminal-a'
      }
    })

    await wrapper.find('[data-testid="toggle-project-project-b"]').trigger('click')
    await wrapper.setProps({ activeTerminalId: 'terminal-b' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-project-project-b"]').attributes('aria-expanded')).toBe('true')
  })

  it('expands a collapsed branch when a new active terminal appears under that project', async () => {
    const wrapper = mountSidebar({
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        terminals: [
          { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
          { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' }
        ],
        activeTerminalId: 'terminal-a'
      }
    })

    await wrapper.find('[data-testid="toggle-project-project-b"]').trigger('click')
    await wrapper.setProps({
      terminals: [
        { id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' },
        { id: 'terminal-b', projectId: 'project-b', shellName: 'bash', currentCommand: '', state: 'running' },
        { id: 'terminal-c', projectId: 'project-b', shellName: 'fish', currentCommand: '', state: 'running' }
      ],
      activeTerminalId: 'terminal-c'
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-c"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="toggle-project-project-b"]').attributes('aria-expanded')).toBe('true')
  })
})

function mountSidebar(options = {}) {
  return mount(ProjectSidebar, {
    props: {
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
      terminals: [{ id: 'terminal-a', projectId: 'project-a', shellName: 'zsh', currentCommand: '', state: 'running' }],
      activeProjectId: 'project-a',
      activeTerminalId: 'terminal-a',
      ...(options.props || {})
    }
  })
}
