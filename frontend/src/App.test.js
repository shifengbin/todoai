import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import { CreateTerminal, SelectTerminal, SendTerminalInput } from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText } from '../wailsjs/runtime/runtime'

const appApiMock = vi.hoisted(() => ({
  CreateTerminal: vi.fn(),
  CreateProjectFromDialog: vi.fn(),
  ListProjects: vi.fn(),
  ResizeTerminal: vi.fn(),
  SelectProject: vi.fn(),
  SelectTerminal: vi.fn(),
  SendTerminalInput: vi.fn(),
  StartShell: vi.fn()
}))

const runtimeMock = vi.hoisted(() => ({
  ClipboardGetText: vi.fn(),
  ClipboardSetText: vi.fn(),
  EventsOff: vi.fn(),
  EventsOn: vi.fn(),
  handlers: {}
}))

const xtermMock = vi.hoisted(() => ({ sessions: new Map() }))

vi.mock('../wailsjs/go/main/App', () => appApiMock)
vi.mock('../wailsjs/runtime/runtime', () => runtimeMock)
vi.mock('./xtermFactory', () => {
  return {
    createXtermSession(terminalId, onData, onShortcut, onCommandState) {
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
        onCommandState,
        terminal
      }
      xtermMock.sessions.set(terminalId, session)
      return session
    }
  }
})

describe('App project terminal tree', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    xtermMock.sessions.clear()
    runtimeMock.handlers = {}
    runtimeMock.EventsOn.mockImplementation((name, handler) => {
      runtimeMock.handlers[name] = handler
    })
    appApiMock.ListProjects.mockResolvedValue(projectState())
    appApiMock.SelectProject.mockResolvedValue(projectState())
    appApiMock.SelectTerminal.mockResolvedValue(projectState())
    appApiMock.CreateTerminal.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          terminal({ id: 'terminal-b', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.StartShell.mockResolvedValue({ projectId: 'project-a', terminalId: 'terminal-a', state: 'running' })
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
    xtermMock.sessions.get('terminal-a').terminal.selection = 'selected output'

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
    expect(SendTerminalInput).toHaveBeenCalledWith('terminal-a', 'echo hi\n')
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
  })

  it('creates an additional terminal under the active project', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="add-terminal-project-a"]').trigger('click')
    await flushPromises()

    expect(CreateTerminal).toHaveBeenCalledWith('project-a', 100, 32)
    expect(xtermMock.sessions.has('terminal-b')).toBe(true)
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').classes()).toContain('active')
  })

  it('selects a terminal from the project tree', async () => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })]
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    appApiMock.SelectProject.mockResolvedValue(twoTerminalState)
    appApiMock.SelectTerminal.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(SelectTerminal).toHaveBeenCalledWith('terminal-b')
    expect(xtermMock.sessions.has('terminal-b')).toBe(true)
    expect(wrapper.find('[data-testid="terminal-pane-terminal-b"]').classes()).toContain('active')
  })

  it('updates terminal labels from command-state events and restores the shell name when idle', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'npm run dev' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('npm run dev')

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-end' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('zsh')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).not.toContain('npm run dev')
  })

  it('restores the shell name when a running command exits with the shell', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'npm run dev' })
    await nextTick()
    runtimeMock.handlers['terminal-status']({ projectId: 'project-a', terminalId: 'terminal-a', state: 'exited' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('zsh')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).not.toContain('npm run dev')
  })

  it('preserves a running command label when switching terminals', async () => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })]
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    appApiMock.SelectProject.mockResolvedValue(twoTerminalState)
    appApiMock.SelectTerminal.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'npm run dev' })
    await nextTick()
    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('npm run dev')
  })
})

async function mountReadyApp() {
  const wrapper = mount(App)
  await flushPromises()
  return wrapper
}

async function openTerminalMenu(wrapper) {
  await wrapper.find('[data-testid="terminal-pane-terminal-a"]').trigger('contextmenu', {
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

function projectState(overrides = {}) {
  return {
    projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
    activeProjectId: 'project-a',
    terminals: [terminal({ id: 'terminal-a' })],
    activeTerminalId: 'terminal-a',
    ...overrides
  }
}

function terminal(overrides = {}) {
  return {
    id: 'terminal-a',
    projectId: 'project-a',
    shellName: 'zsh',
    currentCommand: '',
    state: 'running',
    ...overrides
  }
}
