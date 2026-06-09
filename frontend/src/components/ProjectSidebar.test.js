import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ProjectSidebar from './ProjectSidebar.vue'

describe('ProjectSidebar', () => {
  it('renders a project terminal tree and emits tree actions', async () => {
    const wrapper = mount(ProjectSidebar, {
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
    await wrapper.find('[data-testid="project-project-b"]').trigger('click')
    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')

    expect(wrapper.emitted('create-project')).toHaveLength(1)
    expect(wrapper.emitted('create-terminal')[0]).toEqual(['project-a'])
    expect(wrapper.emitted('select-project')[0]).toEqual(['project-b'])
    expect(wrapper.emitted('select-terminal')[0]).toEqual(['terminal-a'])
  })
})
