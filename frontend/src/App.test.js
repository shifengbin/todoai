import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import { SendTerminalInput } from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText } from '../wailsjs/runtime/runtime'

const appApiMock = vi.hoisted(() => ({
  CreateProjectFromDialog: vi.fn(),
  ListProjects: vi.fn(),
  ResizeTerminal: vi.fn(),
  SelectProject: vi.fn(),
  SendTerminalInput: vi.fn(),
  StartShell: vi.fn()
}))

const runtimeMock = vi.hoisted(() => ({
  ClipboardGetText: vi.fn(),
  ClipboardSetText: vi.fn(),
  EventsOff: vi.fn(),
  EventsOn: vi.fn()
}))

const xtermMock = vi.hoisted(() => ({ sessions: new Map() }))

vi.mock('../wailsjs/go/main/App', () => appApiMock)
vi.mock('../wailsjs/runtime/runtime', () => runtimeMock)
vi.mock('./xtermFactory', () => {
  return {
    createXtermSession(projectId, onData, onShortcut) {
      const terminal = {
        cols: 100,
        rows: 32,
        openedIn: null,
        selection: '',
        open(container) {
          this.openedIn = container
        },
        write: vi.fn(),
        hasSelection() {
          return Boolean(this.selection)
        },
        getSelection() {
          return this.selection
        }
      }
      const session = {
        fitAddon: { fit: vi.fn() },
        onData,
        onShortcut,
        terminal
      }
      xtermMock.sessions.set(projectId, session)
      return session
    }
  }
})

describe('App terminal clipboard context menu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    xtermMock.sessions.clear()
    appApiMock.ListProjects.mockResolvedValue({
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
      activeProjectId: 'project-a'
    })
    appApiMock.StartShell.mockResolvedValue({ projectId: 'project-a', state: 'running' })
    appApiMock.SendTerminalInput.mockResolvedValue()
    runtimeMock.ClipboardGetText.mockResolvedValue('')
    runtimeMock.ClipboardSetText.mockResolvedValue(true)
  })

  it('shows terminal copy and paste actions on right click', async () => {
    const wrapper = await mountReadyApp()

    await openTerminalMenu(wrapper)

    const menu = wrapper.find('[data-testid="terminal-context-menu"]')
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('Copy')
    expect(menu.text()).toContain('Paste')
  })

  it('disables context-menu copy when the terminal has no selection', async () => {
    const wrapper = await mountReadyApp()

    await openTerminalMenu(wrapper)

    expect(wrapper.find('[data-testid="terminal-menu-copy"]').attributes('disabled')).toBeDefined()
  })

  it('copies selected terminal text from the context menu', async () => {
    const wrapper = await mountReadyApp()
    xtermMock.sessions.get('project-a').terminal.selection = 'selected output'

    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-copy"]').trigger('click')
    await flushPromises()

    expect(ClipboardSetText).toHaveBeenCalledWith('selected output')
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
  })

  it('pastes clipboard text into the active shell from the context menu', async () => {
    runtimeMock.ClipboardGetText.mockResolvedValue('echo hi\n')
    const wrapper = await mountReadyApp()

    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-paste"]').trigger('click')
    await flushPromises()

    expect(ClipboardGetText).toHaveBeenCalled()
    expect(SendTerminalInput).toHaveBeenCalledWith('project-a', 'echo hi\n')
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
  })
})

async function mountReadyApp() {
  const wrapper = mount(App)
  await flushPromises()
  return wrapper
}

async function openTerminalMenu(wrapper) {
  await wrapper.find('.terminal-pane').trigger('contextmenu', {
    clientX: 48,
    clientY: 64
  })
  await nextTick()
}

async function flushPromises() {
  await nextTick()
  await Promise.resolve()
  await nextTick()
  await Promise.resolve()
  await nextTick()
}
