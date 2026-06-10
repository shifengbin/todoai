import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import {
  CreateTerminal,
  DeleteProject,
  DeleteTerminal,
  DetectTerminalShell,
  GetProjectGitStatus,
  LoadTerminalSettings,
  SaveTerminalShell,
  SelectTerminal,
  SendTerminalInput,
  StartShell
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText } from '../wailsjs/runtime/runtime'

const appApiMock = vi.hoisted(() => ({
  CreateTerminal: vi.fn(),
  CreateProjectFromDialog: vi.fn(),
  DeleteProject: vi.fn(),
  DeleteTerminal: vi.fn(),
  DetectTerminalShell: vi.fn(),
  GetProjectGitStatus: vi.fn(),
  ListProjects: vi.fn(),
  LoadTerminalSettings: vi.fn(),
  ResizeTerminal: vi.fn(),
  SelectProject: vi.fn(),
  SelectTerminal: vi.fn(),
  SaveTerminalShell: vi.fn(),
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
    createXtermSession(terminalId, onData, onShortcut, onCommandState, onTitleChange) {
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
        onTitleChange,
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
    appApiMock.LoadTerminalSettings.mockResolvedValue(settingsState())
    appApiMock.DetectTerminalShell.mockResolvedValue(shellSetting({ path: '/usr/bin/bash', displayName: 'bash' }))
    appApiMock.SaveTerminalShell.mockResolvedValue(
      settingsState({ selected: shellSetting({ path: '/usr/bin/bash', displayName: 'bash', source: 'detected' }) })
    )
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
    appApiMock.DeleteProject.mockResolvedValue(projectState({ projects: [], activeProjectId: '', terminals: [], activeTerminalId: '' }))
    appApiMock.DeleteTerminal.mockResolvedValue(projectState({ terminals: [], activeTerminalId: '' }))
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus())
    appApiMock.StartShell.mockResolvedValue({ projectId: 'project-a', terminalId: 'terminal-a', state: 'running' })
    appApiMock.SendTerminalInput.mockResolvedValue()
    runtimeMock.ClipboardGetText.mockResolvedValue('')
    runtimeMock.ClipboardSetText.mockResolvedValue(true)
    vi.stubGlobal('confirm', vi.fn(() => true))
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

  it('confirms and deletes a project from the project tree', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('alpha'))
    expect(DeleteProject).toHaveBeenCalledWith('project-a')
    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(false)
  })

  it('does not delete a project when confirmation is cancelled', async () => {
    window.confirm.mockReturnValue(false)
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(true)
  })

  it('deletes a terminal from the project tree without confirmation', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="delete-terminal-terminal-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(DeleteTerminal).toHaveBeenCalledWith('terminal-a')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
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

  it('marks an interactive terminal busy from title changes without replacing the command label', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('codex')
    expect(terminalRow.text()).not.toContain('codex working')
    expect(terminalRow.attributes('data-activity-state')).toBe('busy')
  })

  it('keeps an interactive terminal idle when it receives the initial launch title', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('codex')
    expect(terminalRow.attributes('data-activity-state')).toBe('idle')
  })

  it('marks an interactive terminal as needing input from attention title changes', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('! codex')
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('needs-input')
  })

  it('restores idle activity when an interactive title returns to the command label', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('idle')
  })

  it('clears interactive title activity when a shell command ends', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-end' })
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('zsh')
    expect(terminalRow.attributes('data-activity-state')).toBe('idle')
  })

  it('preserves interactive title activity when switching terminals', async () => {
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

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()
    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')
  })

  it('clears previous interactive title activity when a new shell command starts', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'npm test' })
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('npm test')
    expect(terminalRow.attributes('data-activity-state')).toBe('idle')
  })

  it('clears interactive title activity when the shell exits', async () => {
    const wrapper = await mountReadyApp()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-start', command: 'codex' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex - alpha')
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
    await nextTick()
    runtimeMock.handlers['terminal-status']({ projectId: 'project-a', terminalId: 'terminal-a', state: 'exited' })
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('zsh')
    expect(terminalRow.attributes('data-activity-state')).toBe('idle')
  })

  it('shows the active project git branch and changed file count', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'main', changedCount: 3 }))

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('main')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('3 changed')
  })

  it('shows a stable empty git status when no project is selected', async () => {
    appApiMock.ListProjects.mockResolvedValue(projectState({ projects: [], activeProjectId: '', terminals: [], activeTerminalId: '' }))

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('No project')
  })

  it('shows when the active project is not a git repository', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Not a git repository')
  })

  it('shows when the active project path is unavailable without querying git', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-a', name: 'alpha', path: '/missing/alpha', available: false }],
        terminals: [],
        activeTerminalId: ''
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Project path unavailable')
  })

  it('shows when git status cannot be loaded', async () => {
    appApiMock.GetProjectGitStatus.mockRejectedValue(new Error('git status failed'))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Git status unavailable')
  })

  it('refreshes git status when the active project changes', async () => {
    const twoProjectState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(twoProjectState)
    appApiMock.SelectProject.mockResolvedValue(
      projectState({
        projects: twoProjectState.projects,
        activeProjectId: 'project-b',
        terminals: [terminal({ id: 'terminal-b', projectId: 'project-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.GetProjectGitStatus
      .mockResolvedValueOnce(gitStatus({ projectId: 'project-a', branch: 'main' }))
      .mockResolvedValueOnce(gitStatus({ projectId: 'project-b', branch: 'feature/git-status', changedCount: 2 }))
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="project-project-b"]').trigger('click')
    await flushPromises()

    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-b')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('feature/git-status')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('2 changed')
  })

  it('refreshes git status when a terminal command ends', async () => {
    const wrapper = await mountReadyApp()
    GetProjectGitStatus.mockClear()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-end' })
    await flushPromises()

    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
  })

  it('refreshes git status when the window receives focus', async () => {
    await mountReadyApp()
    GetProjectGitStatus.mockClear()

    window.dispatchEvent(new Event('focus'))
    await flushPromises()

    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
  })

  it('ignores stale git status responses from a previous active project', async () => {
    const twoProjectState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ]
    })
    let resolveProjectA
    let resolveProjectB
    appApiMock.ListProjects.mockResolvedValue(twoProjectState)
    appApiMock.SelectProject.mockResolvedValue(
      projectState({
        projects: twoProjectState.projects,
        activeProjectId: 'project-b',
        terminals: [terminal({ id: 'terminal-b', projectId: 'project-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.GetProjectGitStatus.mockImplementation((projectId) => {
      return new Promise((resolve) => {
        if (projectId === 'project-a') {
          resolveProjectA = resolve
        }
        if (projectId === 'project-b') {
          resolveProjectB = resolve
        }
      })
    })
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="project-project-b"]').trigger('click')
    await flushPromises()
    resolveProjectB(gitStatus({ projectId: 'project-b', branch: 'feature/git-status', changedCount: 2 }))
    await flushPromises()
    resolveProjectA(gitStatus({ projectId: 'project-a', branch: 'main', changedCount: 0 }))
    await flushPromises()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('feature/git-status')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('main')
  })


  it('opens terminal settings and renders the loaded shell state', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        selected: shellSetting({ path: '/usr/bin/zsh', displayName: 'zsh', source: 'manual' }),
        detected: shellSetting({ path: '/usr/bin/bash', displayName: 'bash', source: 'detected' })
      })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(LoadTerminalSettings).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="terminal-settings-current"]').text()).toContain('/usr/bin/zsh')
    expect(wrapper.find('[data-testid="terminal-settings-detected"]').text()).toContain('/usr/bin/bash')
  })

  it('shows an unavailable saved shell and fallback shell in settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        selected: shellSetting({ path: '/old/bin/zsh', displayName: 'zsh', source: 'manual', available: false }),
        fallback: shellSetting({ path: '/bin/sh', displayName: 'sh', source: 'detected' })
      })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="terminal-settings-current"]').text()).toContain('Unavailable')
    expect(wrapper.find('[data-testid="terminal-settings-fallback"]').text()).toContain('/bin/sh')
  })

  it('re-detects a terminal shell from settings', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-redetect"]').trigger('click')
    await flushPromises()

    expect(DetectTerminalShell).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-settings-detected"]').text()).toContain('/usr/bin/bash')
  })

  it('saves the detected terminal shell setting', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ detected: shellSetting({ path: '/usr/bin/bash', displayName: 'bash', source: 'detected' }) })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-detected-option"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalShell).toHaveBeenCalledWith('/usr/bin/bash', 'detected')
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(false)
  })

  it('saves a manual terminal shell path', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-manual-option"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-manual-path"]').setValue('/opt/custom/bin/fish')
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalShell).toHaveBeenCalledWith('/opt/custom/bin/fish', 'manual')
  })

  it('shows terminal settings save errors without closing the dialog', async () => {
    appApiMock.SaveTerminalShell.mockRejectedValue(new Error('terminal shell path is not executable'))
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-manual-option"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-manual-path"]').setValue('/missing/shell')
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-settings-error"]').text()).toContain('terminal shell path is not executable')
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(true)
  })

  it('does not restart the active terminal when terminal settings are saved', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-detected-option"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(StartShell).not.toHaveBeenCalled()
    expect(xtermMock.sessions.has('terminal-a')).toBe(true)
  })
})

async function mountReadyApp() {
  const wrapper = mount(App)
  await flushPromises()
  return wrapper
}

async function openSettings(wrapper) {
  await wrapper.find('[data-testid="settings-toggle"]').trigger('click')
  await flushPromises()
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

function settingsState(overrides = {}) {
  return {
    version: 1,
    selected: shellSetting(),
    ...overrides
  }
}

function shellSetting(overrides = {}) {
  return {
    path: '/usr/bin/zsh',
    displayName: 'zsh',
    source: 'manual',
    available: true,
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

function gitStatus(overrides = {}) {
  return {
    projectId: 'project-a',
    isRepo: true,
    branch: 'main',
    changedCount: 0,
    stagedCount: 0,
    unstagedCount: 0,
    untrackedCount: 0,
    ahead: 0,
    behind: 0,
    ...overrides
  }
}
