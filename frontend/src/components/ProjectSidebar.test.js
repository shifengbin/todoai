import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ProjectSidebar from './ProjectSidebar.vue'

describe('ProjectSidebar', () => {
  it('renders projects and emits create/select actions', async () => {
    const wrapper = mount(ProjectSidebar, {
      props: {
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: false }
        ],
        activeProjectId: 'project-a'
      }
    })

    expect(wrapper.text()).toContain('alpha')
    expect(wrapper.text()).toContain('/work/alpha')
    expect(wrapper.text()).toContain('beta')
    expect(wrapper.text()).toContain('Unavailable')
    expect(wrapper.find('[data-testid="project-project-a"]').classes()).toContain('active')

    await wrapper.find('[data-testid="new-project"]').trigger('click')
    await wrapper.find('[data-testid="project-project-b"]').trigger('click')

    expect(wrapper.emitted('create-project')).toHaveLength(1)
    expect(wrapper.emitted('select-project')[0]).toEqual(['project-b'])
  })
})
