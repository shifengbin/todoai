import { readFileSync } from 'node:fs'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import {
  AddProjectToTodo,
  AddProjectsToTodo,
  AddProjectSelectionsToTodo,
  ChangeTodoStatus,
  ClearRecentWorkspaces,
  CloseWorkspace,
  CompleteTodo,
  CreateTodo,
  CreateTaskTerminal,
  CreateTodoTerminal,
  CreateWorkspaceTerminal,
  CreateProjectFromDialog,
  DeleteCompletedTodos,
  DeleteProject,
  DeleteProjects,
  DeleteTerminal,
  DeleteTodo,
  DetectTerminalShell,
  GetCompletedTodoProjectMergeStatuses,
  GetProjectGitStatus,
  GetTodoGitStatus,
  GetTodoProjectGitStatus,
  ImportProjectsFromParentDirectoryDialog,
  InitializeGitRepositoryAndImportProject,
  InitializeProjectGitRepository,
  ListProjectBranches,
  LoadTerminalSettings,
  LoadTodoInitializationFiles,
  LoadTodoLifecycleScripts,
  LoadTodoProjectUIState,
  OpenTodoFolder,
  OpenRecentWorkspace,
  OpenWorkspaceFromDialog,
  OpenWorkspaceFromPath,
  RemoveTodoProject,
  RetryTodoLifecycleScript,
  SaveTerminalLaunchProfiles,
  SaveTerminalShell,
  SaveTerminalTheme,
  SaveTodoInitializationFiles,
  SaveTodoLifecycleScripts,
  SaveTodoSidebarWidth,
  SaveTodoProjectUIState,
  SelectTerminal,
  SendTerminalInput,
  StartShell,
  StartTaskBackgroundCommand,
  StartTodoProjectBackgroundCommand,
  UpdateTodo,
  WorkspaceState
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText } from '../wailsjs/runtime/runtime'

const appApiMock = vi.hoisted(() => ({
  AddProjectToTodo: vi.fn(),
  AddProjectsToTodo: vi.fn(),
  AddProjectSelectionsToTodo: vi.fn(),
  ChangeTodoStatus: vi.fn(),
  ClearRecentWorkspaces: vi.fn(),
  CloseWorkspace: vi.fn(),
  CompleteTodo: vi.fn(),
  CreateTodo: vi.fn(),
  CreateTaskTerminal: vi.fn(),
  CreateTodoTerminal: vi.fn(),
  CreateWorkspaceTerminal: vi.fn(),
  CreateProjectFromDialog: vi.fn(),
  DeleteCompletedTodos: vi.fn(),
  DeleteProject: vi.fn(),
  DeleteProjects: vi.fn(),
  DeleteTerminal: vi.fn(),
  DeleteTodo: vi.fn(),
  DetectTerminalShell: vi.fn(),
  GetCompletedTodoProjectMergeStatuses: vi.fn(),
  GetProjectGitStatus: vi.fn(),
  GetTodoGitStatus: vi.fn(),
  GetTodoProjectGitStatus: vi.fn(),
  ImportProjectsFromParentDirectoryDialog: vi.fn(),
  InitializeGitRepositoryAndImportProject: vi.fn(),
  InitializeProjectGitRepository: vi.fn(),
  ListProjectBranches: vi.fn(),
  ListProjects: vi.fn(),
  LoadTerminalSettings: vi.fn(),
  LoadTodoInitializationFiles: vi.fn(),
  LoadTodoLifecycleScripts: vi.fn(),
  LoadTodoProjectUIState: vi.fn(),
  OpenTodoFolder: vi.fn(),
  OpenRecentWorkspace: vi.fn(),
  OpenWorkspaceFromDialog: vi.fn(),
  OpenWorkspaceFromPath: vi.fn(),
  RemoveTodoProject: vi.fn(),
  ResizeTerminal: vi.fn(),
  RetryTodoLifecycleScript: vi.fn(),
  SelectProject: vi.fn(),
  SelectTerminal: vi.fn(),
  SelectTodoProject: vi.fn(),
  SaveTerminalLaunchProfiles: vi.fn(),
  SaveTerminalShell: vi.fn(),
  SaveTerminalTheme: vi.fn(),
  SaveTodoInitializationFiles: vi.fn(),
  SaveTodoLifecycleScripts: vi.fn(),
  SaveTodoSidebarWidth: vi.fn(),
  SaveTodoProjectUIState: vi.fn(),
  SendTerminalInput: vi.fn(),
  StartShell: vi.fn(),
  StartTaskBackgroundCommand: vi.fn(),
  StartTodoProjectBackgroundCommand: vi.fn(),
  UpdateTodo: vi.fn(),
  WorkspaceState: vi.fn()
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
        write: vi.fn((data, callback) => callback?.()),
        focus: vi.fn(),
        dispose: vi.fn(() => {
          xtermMock.sessions.delete(terminalId)
        }),
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

const defaultCodexLaunchCommand =
  'codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox'
const defaultClaudeLaunchCommand = 'claude --dangerously-skip-permissions'
const mountedWrappers = []

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
    appApiMock.LoadTodoInitializationFiles.mockResolvedValue([])
    appApiMock.LoadTodoLifecycleScripts.mockResolvedValue([])
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(todoProjectUIStateFile())
    appApiMock.SaveTodoProjectUIState.mockResolvedValue()
    appApiMock.DetectTerminalShell.mockResolvedValue(shellSetting({ path: '/usr/bin/bash', displayName: 'bash' }))
    appApiMock.SaveTerminalShell.mockResolvedValue(
      settingsState({ selected: shellSetting({ path: '/usr/bin/bash', displayName: 'bash', source: 'detected' }) })
    )
    appApiMock.SaveTerminalLaunchProfiles.mockResolvedValue(settingsState())
    appApiMock.SaveTerminalTheme.mockResolvedValue(settingsState())
    appApiMock.SaveTodoInitializationFiles.mockResolvedValue(settingsState())
    appApiMock.SaveTodoLifecycleScripts.mockResolvedValue(settingsState())
    appApiMock.SaveTodoSidebarWidth.mockResolvedValue()
    appApiMock.StartTaskBackgroundCommand.mockResolvedValue()
    appApiMock.StartTodoProjectBackgroundCommand.mockResolvedValue()
    appApiMock.SelectProject.mockResolvedValue(projectState())
    appApiMock.SelectTodoProject.mockResolvedValue(projectState())
    appApiMock.SelectTerminal.mockResolvedValue(projectState())
    appApiMock.CreateTodoTerminal.mockResolvedValue(
      inProgressProjectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          terminal({ id: 'terminal-b', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.CreateWorkspaceTerminal.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          workspaceTerminal({ id: 'global-terminal', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'global-terminal'
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(projectState({ todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })] }))
    appApiMock.AddProjectToTodo.mockResolvedValue(projectState())
    appApiMock.AddProjectsToTodo.mockResolvedValue(projectState())
    appApiMock.AddProjectSelectionsToTodo.mockResolvedValue(projectState())
    appApiMock.CreateTaskTerminal.mockResolvedValue(
      inProgressProjectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          taskTerminal({ id: 'task-terminal-a', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'task-terminal-a'
      })
    )
    appApiMock.OpenTodoFolder.mockResolvedValue()
    appApiMock.ChangeTodoStatus.mockResolvedValue(projectState({ todos: [todo({ status: 'in-progress' })] }))
    appApiMock.CompleteTodo.mockResolvedValue(projectState({ todos: [completedTodo()], todoProjects: [], terminals: [], activeTodoId: '', activeTodoProjectId: '', activeTerminalId: '' }))
    appApiMock.DeleteTodo.mockResolvedValue(projectState({ todos: [], todoProjects: [], terminals: [], activeTodoId: '', activeTodoProjectId: '', activeTerminalId: '' }))
    appApiMock.ImportProjectsFromParentDirectoryDialog.mockResolvedValue(
      projectState({
        importSummary: { parentPath: '/work', addedCount: 2, skippedCount: 1 }
      })
    )
    appApiMock.DeleteProject.mockResolvedValue(
      projectState({
        projects: [],
        todoProjects: [],
        activeProjectId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.DeleteProjects.mockResolvedValue(
      projectState({
        projects: [],
        todoProjects: [],
        activeProjectId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.DeleteCompletedTodos.mockResolvedValue(
      projectState({
        todos: [],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.DeleteTerminal.mockResolvedValue(projectState({ terminals: [], activeTerminalId: '' }))
    appApiMock.RemoveTodoProject.mockResolvedValue(projectState({ todoProjects: [], terminals: [], activeTodoProjectId: '', activeTerminalId: '' }))
    appApiMock.RetryTodoLifecycleScript.mockResolvedValue(inProgressProjectState({ lifecycleScriptStatuses: [lifecycleScriptStatus({ status: 'running' })] }))
    appApiMock.UpdateTodo.mockResolvedValue(projectState())
    appApiMock.WorkspaceState.mockResolvedValue(workspaceState())
    appApiMock.OpenWorkspaceFromDialog.mockResolvedValue(projectState())
    appApiMock.OpenWorkspaceFromPath.mockResolvedValue(projectState())
    appApiMock.OpenRecentWorkspace.mockResolvedValue(projectState())
    appApiMock.CloseWorkspace.mockResolvedValue(noWorkspaceState())
    appApiMock.ClearRecentWorkspaces.mockResolvedValue(workspaceState({ recentWorkspaces: [] }))
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus())
    appApiMock.GetTodoGitStatus.mockResolvedValue(gitStatus({ projectId: '', isRepo: false, branch: '' }))
    appApiMock.GetTodoProjectGitStatus.mockResolvedValue(gitStatus())
    appApiMock.CreateProjectFromDialog.mockResolvedValue(projectImportResult(projectState()))
    appApiMock.InitializeGitRepositoryAndImportProject.mockResolvedValue(projectState())
    appApiMock.GetCompletedTodoProjectMergeStatuses.mockResolvedValue([])
    appApiMock.InitializeProjectGitRepository.mockResolvedValue()
    appApiMock.ListProjectBranches.mockResolvedValue([])
    appApiMock.StartShell.mockResolvedValue({ projectId: 'project-a', terminalId: 'terminal-a', state: 'running' })
    appApiMock.SendTerminalInput.mockResolvedValue()
    runtimeMock.ClipboardGetText.mockResolvedValue('')
    runtimeMock.ClipboardSetText.mockResolvedValue(true)
    vi.stubGlobal('confirm', vi.fn(() => true))
    vi.stubGlobal('prompt', vi.fn(() => ''))
  })

  afterEach(() => {
    for (const wrapper of mountedWrappers.splice(0)) {
      wrapper.unmount()
    }
  })

  it('shows no workspace empty state while keeping global settings available', async () => {
    appApiMock.ListProjects.mockResolvedValue(noWorkspaceState())
    const wrapper = await mountReadyApp()

    expect(LoadTerminalSettings).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-testid="todo-workspace-empty"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="new-todo"]').attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="settings-toggle"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('[data-testid="terminal-surface"]').text()).toContain('Open a project')
    expect(wrapper.find('[data-testid="project-git-status"]').findAll('.status-chip')).toHaveLength(0)
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('No project')

    await openSettings(wrapper)
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(true)

    expect(wrapper.find('[data-testid="workspace-tabs"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sidebar-tab-projects"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="project-library"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="import-parent-directory"]').exists()).toBe(false)
  })

  it('opens a selected recent workspace from the menu event picker', async () => {
    appApiMock.ListProjects.mockResolvedValue(noWorkspaceState())
    appApiMock.OpenRecentWorkspace.mockResolvedValue(
      projectState({ currentWorkspace: workspace({ name: 'Customer A', path: '/work/customer-a' }) })
    )
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['workspace-recent'](
      workspaceState({
        recentWorkspaces: [
          workspace({ name: 'Customer A', path: '/work/customer-a', available: true }),
          workspace({ name: 'Missing', path: '/missing/customer-b', available: false })
        ]
      })
    )
    await nextTick()

    const dialog = wrapper.find('[data-testid="recent-workspace-dialog"]')
    expect(dialog.exists()).toBe(true)
    expect(dialog.text()).toContain('Customer A')
    expect(dialog.text()).toContain('/missing/customer-b')
    expect(dialog.text()).toContain('Unavailable')

    await wrapper.find('[data-testid="recent-workspace-0"]').trigger('click')
    await flushPromises()

    expect(OpenRecentWorkspace).toHaveBeenCalledWith('/work/customer-a')
    expect(LoadTerminalSettings).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-testid="recent-workspace-dialog"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-workspace-empty"]').exists()).toBe(false)
  })

  it('applies workspace-state events by clearing previous terminal and git context', async () => {
    const wrapper = await mountReadyApp()
    const session = xtermMock.sessions.get('terminal-a')
    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      updatedAt: 10
    })
    await nextTick()
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')
    ClipboardGetText.mockRejectedValueOnce(new Error('old workspace error'))
    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-paste"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('.status-error').text()).toContain('old workspace error')

    LoadTerminalSettings.mockClear()
    GetProjectGitStatus.mockClear()
    runtimeMock.handlers['workspace-state'](noWorkspaceState())
    await flushPromises()

    expect(session.terminal.dispose).toHaveBeenCalledTimes(1)
    expect(xtermMock.sessions.has('terminal-a')).toBe(false)
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="terminal-surface"]').text()).toContain('Open a project')
    expect(wrapper.find('[data-testid="project-git-status"]').findAll('.status-chip')).toHaveLength(0)
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('No project')
    expect(wrapper.find('.status-error').exists()).toBe(false)
    expect(LoadTerminalSettings).not.toHaveBeenCalled()
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
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

  it('copies selected Unicode terminal text from the context menu', async () => {
    const wrapper = await mountReadyApp()
    xtermMock.sessions.get('terminal-a').terminal.selection = '中文 ✓ 🔧   '

    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-copy"]').trigger('click')
    await flushPromises()

    expect(ClipboardSetText).toHaveBeenCalledWith('中文 ✓ 🔧   ')
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
  })

  it('pastes Unicode clipboard text into the active shell from the context menu', async () => {
    const text = "printf '中文 ✓ 🔧   \\n'"
    runtimeMock.ClipboardGetText.mockResolvedValue(text)
    const wrapper = await mountReadyApp()

    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-paste"]').trigger('click')
    await flushPromises()

    expect(ClipboardGetText).toHaveBeenCalled()
    expect(SendTerminalInput).toHaveBeenCalledWith('terminal-a', text)
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
    expect(xtermMock.sessions.get('terminal-a').terminal.focus).toHaveBeenCalledTimes(1)
  })

  it('creates and selects workspace global terminals from the terminal surface', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' })],
        activeTerminalId: 'terminal-a'
      })
    )
    appApiMock.CreateWorkspaceTerminal.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          workspaceTerminal({ id: 'global-terminal', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'global-terminal'
      })
    )
    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="global-terminal-group"]').exists()).toBe(false)

    await wrapper.find('[data-testid="create-global-terminal"]').trigger('click')
    await flushPromises()

    expect(CreateWorkspaceTerminal).toHaveBeenCalledWith(100, 32)
    expect(wrapper.find('[data-testid="global-terminal-group"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="global-terminal-global-terminal"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="terminal-pane-global-terminal"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="terminal-global-terminal"]').exists()).toBe(false)

    CreateWorkspaceTerminal.mockClear()
    await wrapper.find('[data-testid="create-global-terminal-from-group"]').trigger('click')
    await flushPromises()
    expect(CreateWorkspaceTerminal).toHaveBeenCalledWith(100, 32)
  })

  it('uses terminal iconography for global terminal creation controls', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          workspaceTerminal({ id: 'global-terminal', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'global-terminal'
      })
    )
    const wrapper = await mountReadyApp()

    const headerCreateButton = wrapper.find('[data-testid="create-global-terminal"]')
    expect(headerCreateButton.find('.lucide-square-terminal').exists()).toBe(true)
    expect(headerCreateButton.find('.lucide-plus').exists()).toBe(false)

    const groupCreateButton = wrapper.find('[data-testid="create-global-terminal-from-group"]')
    expect(groupCreateButton.find('.lucide-square-terminal').exists()).toBe(true)
    expect(groupCreateButton.find('.lucide-plus').exists()).toBe(false)
  })

  it('shows global terminal group only while workspace terminals exist', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          workspaceTerminal({ id: 'global-terminal', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'global-terminal'
      })
    )
    appApiMock.DeleteTerminal.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' })],
        activeTerminalId: 'terminal-a'
      })
    )
    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="global-terminal-group"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="terminal-global-terminal"]').exists()).toBe(false)

    await wrapper.find('[data-testid="delete-global-terminal-global-terminal"]').trigger('click')
    await flushPromises()

    expect(DeleteTerminal).toHaveBeenCalledWith('global-terminal')
    expect(wrapper.find('[data-testid="global-terminal-group"]').exists()).toBe(false)
  })

  it('closes the context menu and restores terminal focus when pasting an empty clipboard', async () => {
    runtimeMock.ClipboardGetText.mockResolvedValue('')
    const wrapper = await mountReadyApp()

    await openTerminalMenu(wrapper)
    await wrapper.find('[data-testid="terminal-menu-paste"]').trigger('click')
    await flushPromises()

    expect(ClipboardGetText).toHaveBeenCalled()
    expect(SendTerminalInput).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-context-menu"]').exists()).toBe(false)
    expect(xtermMock.sessions.get('terminal-a').terminal.focus).toHaveBeenCalledTimes(1)
  })

  it('shows configured terminal launch profiles from loaded settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: true },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]')
    expect(LoadTerminalSettings).toHaveBeenCalled()
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).toContain('codex')
    expect(menu.text()).not.toContain('claude')
  })

  it('shows only Terminal in launch menu when every custom launch profile is disabled', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: false },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]')
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).not.toContain('codex')
    expect(menu.text()).not.toContain('claude')
  })

  it('creates an additional terminal under the active project', async () => {
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(SendTerminalInput).not.toHaveBeenCalled()
    expect(xtermMock.sessions.has('terminal-b')).toBe(true)
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').classes()).toContain('active')
  })

  it('creates a cleared TODO project terminal without showing project path unavailable', async () => {
    const clearedState = inProgressProjectState({
      todoProjects: [
        todoProject({
          id: 'todo-project-a',
          todoId: 'todo-a',
          projectId: 'project-a',
          path: '/work/alpha',
          available: true,
          worktreeStatus: 'cleared',
          worktreePath: '/work/customer-a/tasks/abc123/alpha'
        })
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(clearedState)
    appApiMock.CreateTodoTerminal.mockResolvedValue(
      inProgressProjectState({
        todoProjects: clearedState.todoProjects,
        terminals: [
          terminal({ id: 'terminal-a' }),
          terminal({ id: 'terminal-b', todoProjectId: 'todo-project-a', projectId: 'project-a', state: 'running' })
        ],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ projectId: 'project-a', branch: 'main', changedCount: 0 }))
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('Project path unavailable')
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('main')
  })

  it('creates a terminal for available TODO projects without loaded worktree metadata', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            worktreeStatus: undefined,
            worktreePath: undefined
          })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    const addTerminalButton = wrapper.find('[data-testid="add-terminal-todo-project-a"]')

    expect(addTerminalButton.attributes('disabled')).toBeUndefined()

    await addTerminalButton.trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
  })

  it('preserves current TODO view and sidebar width when creating a terminal', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 380,
        todoProjects: {
          'todo-project-a': { todoView: 'not-started' }
        }
      })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    appApiMock.CreateTodoTerminal.mockResolvedValue(
      inProgressProjectState({
        activeTodoProjectId: 'todo-project-a',
        terminals: [
          terminal({ id: 'terminal-a' }),
          terminal({ id: 'terminal-b', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 380px')

    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 380px')
  })

  it('shows unsupported embedded terminal state without restarting the shell', async () => {
    appApiMock.CreateTodoTerminal.mockResolvedValue(
      inProgressProjectState({
        terminals: [terminal({ id: 'terminal-unsupported', state: 'unsupported' })],
        activeTerminalId: 'terminal-unsupported'
      })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()
    StartShell.mockClear()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()
    window.dispatchEvent(new Event('focus'))
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-surface"]').text()).toContain('not supported on Windows')
    expect(StartShell).not.toHaveBeenCalled()
  })

  it('creates a terminal from a custom launch profile and submits its command', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Codex GPT-5', command: 'codex --model gpt-5', enabled: true }] })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(SendTerminalInput).toHaveBeenCalledWith('terminal-b', 'codex --model gpt-5\r')
    const terminalRow = wrapper.find('[data-testid="terminal-terminal-b"]')
    expect(terminalRow.classes()).toContain('active')
    expect(terminalRow.text()).toContain('codex --model gpt-5')
    expect(terminalRow.attributes('data-activity-state')).toBe('idle')

    xtermMock.sessions.get('terminal-b').onCommandState({ type: 'command-end' })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').text()).toContain('bash')
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').text()).not.toContain('codex --model gpt-5')
  })

  it('starts a project background launch profile without creating a terminal', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Sync Docs', command: 'npm run sync-docs', enabled: true, background: true }] })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()
    CreateTodoTerminal.mockClear()
    SendTerminalInput.mockClear()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')
    await flushPromises()

    expect(StartTodoProjectBackgroundCommand).toHaveBeenCalledWith('todo-project-a', 'npm run sync-docs')
    expect(CreateTodoTerminal).not.toHaveBeenCalled()
    expect(SendTerminalInput).not.toHaveBeenCalled()
    expect(xtermMock.sessions.has('terminal-b')).toBe(false)
    expect(wrapper.find('[data-testid="terminal-pane-terminal-a"]').classes()).toContain('active')
  })

  it('starts a task background launch profile without creating a task terminal', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Prepare Task', command: 'npm run prepare', enabled: true, background: true }] })
    )
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()
    CreateTaskTerminal.mockClear()
    SendTerminalInput.mockClear()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-task-terminal-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-task-todo-a-1"]').trigger('click')
    await flushPromises()

    expect(StartTaskBackgroundCommand).toHaveBeenCalledWith('todo-a', 'npm run prepare')
    expect(CreateTaskTerminal).not.toHaveBeenCalled()
    expect(SendTerminalInput).not.toHaveBeenCalled()
    expect(xtermMock.sessions.has('task-terminal-a')).toBe(false)
    expect(wrapper.find('[data-testid="terminal-pane-terminal-a"]').classes()).toContain('active')
  })

  it('keeps structured agent status when title changes arrive', async () => {
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'needs-input',
      source: 'claude-hook',
      confidence: 'structured',
      reason: 'permission-prompt',
      updatedAt: 10
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('needs-input')

    xtermMock.sessions.get('terminal-a').onTitleChange('claude thinking')
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('needs-input')
  })

  it('clears structured agent status when the shell exits', async () => {
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      updatedAt: 10
    })
    await nextTick()
    runtimeMock.handlers['terminal-status']({ projectId: 'project-a', terminalId: 'terminal-a', state: 'exited' })
    await nextTick()
    xtermMock.sessions.get('terminal-a').onTitleChange('codex thinking')
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('idle')
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
    expect(xtermMock.sessions.get('terminal-b').terminal.focus).toHaveBeenCalled()
  })

  it('preserves current TODO view and sidebar width when selecting a terminal', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 380,
        todoProjects: {
          'todo-project-a': { todoView: 'not-started' }
        }
      })
    )
    const twoTerminalState = inProgressProjectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    appApiMock.SelectTerminal.mockResolvedValue(
      inProgressProjectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b',
        activeTodoProjectId: 'todo-project-a'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 380px')

    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(SelectTerminal).toHaveBeenCalledWith('terminal-b')
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 380px')
  })

  it('does not focus a terminal when selecting it fails', async () => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })]
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    appApiMock.SelectProject.mockResolvedValue(twoTerminalState)
    appApiMock.SelectTerminal.mockResolvedValueOnce(
      projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(xtermMock.sessions.has('terminal-a')).toBe(true)
    expect(xtermMock.sessions.has('terminal-b')).toBe(true)
    xtermMock.sessions.get('terminal-a').terminal.focus.mockClear()
    xtermMock.sessions.get('terminal-b').terminal.focus.mockClear()

    appApiMock.SelectTerminal.mockRejectedValueOnce(new Error('select failed'))
    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
    await flushPromises()

    expect(SelectTerminal).toHaveBeenLastCalledWith('terminal-a')
    expect(xtermMock.sessions.get('terminal-a').terminal.focus).not.toHaveBeenCalled()
    expect(xtermMock.sessions.get('terminal-b').terminal.focus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-pane-terminal-b"]').classes()).toContain('active')
  })

  it('automatically restarts a restored terminal when selected', async () => {
    const restoredTerminal = terminal({
      id: 'terminal-restored',
      shellName: 'zsh',
      state: 'exited',
      output: 'previous codex output\r\n'
    })
    const initialState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), restoredTerminal],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(initialState)
    appApiMock.SelectTerminal.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' }), restoredTerminal],
        activeTerminalId: 'terminal-restored'
      })
    )
    const wrapper = await mountReadyApp()
    StartShell.mockClear()

    await wrapper.find('[data-testid="terminal-terminal-restored"]').trigger('click')
    await flushPromises()

    expect(SelectTerminal).toHaveBeenCalledWith('terminal-restored')
    expect(xtermMock.sessions.get('terminal-restored').terminal.write).toHaveBeenCalledWith(
      'previous codex output\r\n',
      expect.any(Function)
    )
    expect(StartShell).toHaveBeenCalledWith('terminal-restored', 100, 32)
  })

  it('clears global candidate projects from the add-project dialog after confirmation', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            projectId: 'project-a',
            sourceProjectId: 'project-a',
            name: 'alpha',
            path: '/work/alpha',
            available: true
          })
        ]
      })
    )
    appApiMock.DeleteProjects.mockResolvedValue(
      projectState({
        projects: [],
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            projectId: 'project-a',
            sourceProjectId: 'project-a',
            name: 'alpha',
            path: '/work/alpha',
            available: true
          })
        ],
        activeProjectId: '',
        activeTodoProjectId: 'todo-project-a'
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    window.confirm.mockReturnValueOnce(false)
    await wrapper.find('[data-testid="clear-global-project-candidates"]').trigger('click')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('Clear global project candidates'))
    expect(DeleteProject).not.toHaveBeenCalled()
    expect(DeleteProjects).not.toHaveBeenCalled()

    window.confirm.mockReturnValueOnce(true)
    await wrapper.find('[data-testid="clear-global-project-candidates"]').trigger('click')
    await flushPromises()

    expect(DeleteProjects).toHaveBeenCalledWith(['project-a', 'project-b'])
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').text()).toContain('alpha')
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').text()).not.toContain('/work/alpha')
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-a"]').exists()).toBe(false)
  })

  it('does not clear global candidates when clear confirmation is cancelled', async () => {
    window.confirm.mockReturnValueOnce(false)
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="clear-global-project-candidates"]').trigger('click')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalled()
    expect(DeleteProjects).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-b"]').exists()).toBe(true)
  })

  it('clears one global project candidate after confirmation', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        todoProjects: [],
        activeProjectId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.DeleteProject.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-b', name: 'beta', path: '/work/beta', available: true }],
        todoProjects: [],
        activeProjectId: 'project-b',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="clear-project-candidate-project-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-candidate-clear-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="project-candidate-clear-name"]').text()).toContain('alpha')
    expect(wrapper.find('[data-testid="project-candidate-clear-path"]').text()).toContain('/work/alpha')
    expect(DeleteProject).not.toHaveBeenCalled()

    await wrapper.find('[data-testid="project-candidate-clear-confirm"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).toHaveBeenCalledWith('project-a')
    expect(DeleteProjects).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-candidate-clear-dialog"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-b"]').exists()).toBe(true)
  })

  it('does not clear one global project candidate when confirmation is cancelled', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        todoProjects: [],
        activeProjectId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="clear-project-candidate-project-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-candidate-clear-dialog"]').exists()).toBe(true)

    await wrapper.find('[data-testid="project-candidate-clear-cancel"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).not.toHaveBeenCalled()
    expect(DeleteProjects).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-candidate-clear-dialog"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-b"]').exists()).toBe(true)
  })

  it('removes a cleared selected candidate from pending TODO creation', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ]
      })
    )
    appApiMock.DeleteProject.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(true)

    await wrapper.find('[data-testid="clear-project-candidate-project-b"]').trigger('click')
    await wrapper.find('[data-testid="project-candidate-clear-confirm"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).toHaveBeenCalledWith('project-b')
    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: []
    })
  })

  it('keeps TODO project copies and terminals when clearing one candidate', async () => {
    const todoProjectCopy = todoProject({
      id: 'todo-project-a',
      projectId: 'project-a',
      sourceProjectId: 'project-a',
      name: 'alpha-copy',
      path: '/work/alpha-copy',
      available: true
    })
    const runningTerminal = terminal({ id: 'terminal-a', projectId: 'project-a', todoProjectId: 'todo-project-a' })
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        todoProjects: [todoProjectCopy],
        terminals: [runningTerminal],
        activeProjectId: 'project-a',
        activeTodoProjectId: 'todo-project-a',
        activeTerminalId: 'terminal-a'
      })
    )
    appApiMock.DeleteProject.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-b', name: 'beta', path: '/work/beta', available: true }],
        todoProjects: [todoProjectCopy],
        terminals: [runningTerminal],
        activeProjectId: 'project-b',
        activeTodoProjectId: 'todo-project-a',
        activeTerminalId: 'terminal-a'
      })
    )
    const wrapper = await mountReadyApp()
    const session = xtermMock.sessions.get('terminal-a')

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="clear-project-candidate-project-a"]').trigger('click')
    await wrapper.find('[data-testid="project-candidate-clear-confirm"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).toHaveBeenCalledWith('project-a')
    expect(DeleteTerminal).not.toHaveBeenCalled()
    expect(session.terminal.dispose).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').text()).toContain('alpha-copy')
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').text()).not.toContain('/work/alpha-copy')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(true)
  })

  it('creates a TODO with details, priority, and searched optional projects', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ]
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ],
        todos: [todo({ id: 'todo-b', title: 'Write tests', description: 'Cover login flow', priority: 'high' })],
        todoProjects: [
          todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b' }),
          todoProject({ id: 'todo-project-c', todoId: 'todo-b', projectId: 'project-c' })
        ],
        activeTodoId: 'todo-b',
        activeTodoProjectId: 'todo-project-b',
        activeProjectId: 'project-b'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-description-input"]').setValue('Cover login flow')
    await wrapper.find('[data-testid="todo-priority-high"]').setValue(true)
    await wrapper.find('[data-testid="todo-project-filter"]').setValue('api')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-option-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-option-project-c"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-option-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-option-project-c"]').trigger('click')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: 'Cover login flow',
      priority: 'high',
      projects: [
        { projectId: 'project-b', baseBranch: '' },
        { projectId: 'project-c', baseBranch: '' }
      ]
    })
    expect(wrapper.text()).toContain('Write tests')
  })

  it('preselects default initialization files and submits selected snapshots when creating a TODO', async () => {
    appApiMock.LoadTodoInitializationFiles.mockResolvedValue([
      initializationFile({ name: 'Agent Rules', fileName: 'AGENTS.md', content: 'rules', defaultSelected: true }),
      initializationFile({ name: 'Prompt', description: '可选提示词', fileName: 'prompt.md', content: 'prompt', defaultSelected: false })
    ])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')

    expect(wrapper.find('[data-testid="todo-initialization-file-0"]').text()).toContain('Agent Rules')
    expect(wrapper.find('[data-testid="todo-initialization-file-0"]').text()).toContain('AGENTS.md')
    expect(wrapper.find('[data-testid="todo-initialization-file-1"]').text()).toContain('Prompt')
    expect(wrapper.find('[data-testid="todo-initialization-file-selected-0"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="todo-initialization-file-selected-1"]').element.checked).toBe(false)

    await wrapper.find('[data-testid="todo-initialization-file-selected-1"]').setValue(true)
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [],
      initializationFiles: [
        { name: 'Agent Rules', description: '任务执行约束', fileName: 'AGENTS.md', content: 'rules' },
        { name: 'Prompt', description: '可选提示词', fileName: 'prompt.md', content: 'prompt' }
      ]
    })
  })

  it('defaults selected project base branch from the active project git branch when available', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'develop' }))
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        activeProjectId: 'project-a',
        activeTodoProjectId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [
        { projectId: 'project-a', baseBranch: 'develop' },
        { projectId: 'project-b', baseBranch: '' }
      ]
    })
  })

  it('defaults selected project base branches from workspace preferences when creating a TODO', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'main' }))
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        projectBranchPreferences: {
          'project-a': { baseBranch: 'release/2026' },
          'project-b': { baseBranch: 'develop' }
        },
        activeProjectId: 'project-a',
        activeTodoProjectId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-selected-project-branch-project-a"]').element.value).toBe('release/2026')
    expect(wrapper.find('[data-testid="todo-selected-project-branch-project-b"]').element.value).toBe('develop')

    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [
        { projectId: 'project-a', baseBranch: 'release/2026' },
        { projectId: 'project-b', baseBranch: 'develop' }
      ]
    })
  })

  it('keeps empty workspace branch preferences from falling back to git status', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'main' }))
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ],
        projectBranchPreferences: {
          'project-a': { baseBranch: '' }
        },
        activeProjectId: 'project-a',
        activeTodoProjectId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-selected-project-branch-project-a"]').element.value).toBe('')

    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [{ projectId: 'project-a', baseBranch: '' }]
    })
  })

  it('does not update workspace branch preference while a create TODO form is canceled', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ],
        projectBranchPreferences: {
          'project-a': { baseBranch: 'develop' }
        },
        activeProjectId: 'project-a',
        activeTodoProjectId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-selected-project-branch-project-a"]').setValue('release/2026')
    await wrapper.find('[data-testid="todo-create-close"]').trigger('click')
    await nextTick()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-selected-project-branch-project-a"]').element.value).toBe('develop')
    expect(CreateTodo).not.toHaveBeenCalled()
  })

  it('creates TODOs with selected project base branches', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-selected-project-branch-project-a"]').setValue(' develop ')
    await wrapper.find('[data-testid="todo-selected-project-branch-project-b"]').setValue(' ')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [
        { projectId: 'project-a', baseBranch: 'develop' },
        { projectId: 'project-b', baseBranch: '' }
      ]
    })
  })

  it('offers local and remote project branches when creating a TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ]
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(['main', 'origin/main', 'origin/feature/login'])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await flushPromises()

    expect(ListProjectBranches).toHaveBeenCalledWith('project-a')
    const input = wrapper.find('[data-testid="todo-selected-project-branch-project-a"]')
    await input.trigger('focus')
    await nextTick()
    const options = wrapper
      .find('[data-testid="project-branch-picker-options-todo-create-project-a"]')
      .findAll('[data-testid="project-branch-picker-option-todo-create-project-a"]')
      .map((option) => option.text())
    expect(options).toEqual(['main', 'origin/main', 'origin/feature/login'])
  })

  it('filters project branch candidates as the user types', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ]
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(['main', 'develop', 'release'])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await flushPromises()

    const input = wrapper.find('[data-testid="todo-selected-project-branch-project-a"]')
    await input.trigger('focus')
    await nextTick()
    await input.setValue('rel')
    await nextTick()

    const options = wrapper
      .find('[data-testid="project-branch-picker-options-todo-create-project-a"]')
      .findAll('[data-testid="project-branch-picker-option-todo-create-project-a"]')
      .map((option) => option.text())
    expect(options).toEqual(['release'])
  })

  it('uses opaque branch picker backgrounds', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const inputRule = styles.match(/\.todo-branch-input\s*{([^}]*)}/s)?.[1] || ''
    const menuRule = styles.match(/\.project-branch-picker-menu\s*{([^}]*)}/s)?.[1] || ''

    expect(inputRule).toContain('background: var(--surface-bg);')
    expect(menuRule).toContain('background: var(--surface-bg);')
    expect(inputRule).not.toContain('var(--panel)')
    expect(menuRule).not.toContain('var(--panel)')
  })

  it('keeps project candidate delete buttons away from the scrollbar', () => {
    const styles = readFileSync('src/style.css', 'utf8')
    const optionsRule = styles.match(/\.todo-project-options\s*{([^}]*)}/s)?.[1] || ''

    expect(optionsRule).toContain('padding-right: 10px;')
    expect(optionsRule).toContain('scrollbar-gutter: stable;')
  })

  it('limits rendered project branch candidates while creating a TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ]
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(
      Array.from({ length: 75 }, (_, index) => `feature/branch-${String(index + 1).padStart(2, '0')}`)
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await flushPromises()
    const input = wrapper.find('[data-testid="todo-selected-project-branch-project-a"]')
    await input.setValue('feature/')
    await input.trigger('focus')
    await nextTick()

    const options = wrapper
      .find('[data-testid="project-branch-picker-options-todo-create-project-a"]')
      .findAll('[data-testid="project-branch-picker-option-todo-create-project-a"]')
    expect(options).toHaveLength(50)
    expect(wrapper.find('[data-testid="project-branch-picker-status-todo-create-project-a"]').text()).toContain(
      'Keep typing'
    )
  })

  it('reopens project branch candidates without a stale filter', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ]
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(['main', 'develop', 'release'])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-option-project-a"]').trigger('click')
    await flushPromises()
    const input = wrapper.find('[data-testid="todo-selected-project-branch-project-a"]')
    await input.setValue('rel')
    await input.trigger('keydown', { key: 'Escape' })
    await nextTick()
    await input.trigger('focus')
    await nextTick()

    const options = wrapper
      .find('[data-testid="project-branch-picker-options-todo-create-project-a"]')
      .findAll('[data-testid="project-branch-picker-option-todo-create-project-a"]')
      .map((option) => option.text())
    expect(options).toEqual(['main', 'develop', 'release'])
  })

  it('removes selected project tags before creating a TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-project-filter"]').setValue('api')
    await nextTick()

    await wrapper.find('[data-testid="todo-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-option-project-c"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-c"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-selected-project-remove-project-b"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-c"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [{ projectId: 'project-c', baseBranch: '' }]
    })
  })

  it('creates a TODO without selecting optional projects', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [],
        todos: [],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(
      projectState({
        projects: [],
        todos: [todo({ id: 'todo-no-project', title: 'Write docs' })],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-projects-optional"]').text()).toBe('Optional')
    expect(wrapper.find('[data-testid="todo-project-options"]').text()).not.toContain('No matching projects')

    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write docs')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write docs',
      description: '',
      priority: 'medium',
      projects: []
    })
  })

  it('preserves sidebar width after creating a TODO without an active TODO project', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [],
        todos: [],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(
      projectState({
        projects: [],
        todos: [todo({ id: 'todo-no-project', title: 'Write docs' })],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="sidebar-resize-handle"]').trigger('mousedown', { clientX: 280 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 360 }))
    window.dispatchEvent(new MouseEvent('mouseup', { clientX: 360 }))
    await flushPromises()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write docs')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')
  })

  it('preserves current TODO view after creating a TODO without an active TODO project', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [],
        todos: [todo({ id: 'todo-running', title: 'Investigate', status: 'in-progress' })],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(
      projectState({
        projects: [],
        todos: [
          todo({ id: 'todo-running', title: 'Investigate', status: 'in-progress' }),
          todo({ id: 'todo-new', title: 'Write docs', status: 'not-started' })
        ],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write docs')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="todo-todo-running"]').exists()).toBe(true)
  })

  it('does not create a TODO with a blank name', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('   ')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).not.toHaveBeenCalled()
    expect(wrapper.find('.status-error').text()).toContain('TODO title is required')
  })

  it('adds searched projects to an existing TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.AddProjectSelectionsToTodo.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ],
        todoProjects: [
          todoProject({ projectId: 'project-a' }),
          todoProject({ id: 'todo-project-b', projectId: 'project-b' }),
          todoProject({ id: 'todo-project-c', projectId: 'project-c' })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-filter"]').setValue('api')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-picker-option-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-c"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-option-project-c"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: '' },
      { projectId: 'project-c', baseBranch: '' }
    ])
    expect(wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-todo-project-c"]').exists()).toBe(true)
  })

  it('adds selected projects to existing TODOs with base branches', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-branch-project-b"]').setValue('feature/login-fix')
    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: 'feature/login-fix' }
    ])
  })

  it('defaults added project base branch from workspace preferences', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        projectBranchPreferences: {
          'project-b': { baseBranch: 'origin/release' }
        },
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-project-picker-branch-project-b"]').element.value).toBe('origin/release')

    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: 'origin/release' }
    ])
  })

  it('offers local and remote branches when adding projects to a TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(['develop', 'origin/develop', 'origin/release'])
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await flushPromises()

    expect(ListProjectBranches).toHaveBeenCalledWith('project-b')
    const input = wrapper.find('[data-testid="todo-project-picker-branch-project-b"]')
    await input.trigger('focus')
    await nextTick()
    const options = wrapper
      .find('[data-testid="project-branch-picker-options-project-picker-project-b"]')
      .findAll('[data-testid="project-branch-picker-option-project-picker-project-b"]')
      .map((option) => option.text())
    expect(options).toEqual(['develop', 'origin/develop', 'origin/release'])
  })

  it('adds a project to a TODO using a selected branch candidate', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.ListProjectBranches.mockResolvedValue(['develop', 'origin/develop', 'origin/release'])
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await flushPromises()
    const input = wrapper.find('[data-testid="todo-project-picker-branch-project-b"]')
    await input.trigger('focus')
    await nextTick()
    await wrapper
      .findAll('[data-testid="project-branch-picker-option-project-picker-project-b"]')
      .find((option) => option.text() === 'origin/release')
      .trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: 'origin/release' }
    ])
  })

  it('preserves workspace sidebar width after adding projects to a TODO', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 360,
        todoProjects: {
          'todo-project-a': { todoView: 'not-started' },
          'todo-project-b': { todoView: 'not-started' }
        }
      })
    )
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todoProjects: [todoProject({ id: 'todo-project-a', projectId: 'project-a' })],
        activeTodoProjectId: 'todo-project-a',
        activeProjectId: 'project-a',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.AddProjectSelectionsToTodo.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todoProjects: [
          todoProject({ id: 'todo-project-a', projectId: 'project-a' }),
          todoProject({ id: 'todo-project-b', projectId: 'project-b' })
        ],
        activeTodoId: 'todo-a',
        activeTodoProjectId: 'todo-project-b',
        activeProjectId: 'project-b',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-filter"]').setValue('api')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: '' }
    ])
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')
  })

  it('removes selected project tags before adding projects to a TODO', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true },
          { id: 'project-c', name: 'api-worker', path: '/work/api-worker', available: true }
        ],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [],
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-picker-filter"]').setValue('api')
    await nextTick()

    await wrapper.find('[data-testid="todo-project-picker-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-project-picker-option-project-c"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-c"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-project-picker-remove-project-c"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-c"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-project-picker-submit"]').trigger('click')
    await flushPromises()

    expect(AddProjectSelectionsToTodo).toHaveBeenCalledWith('todo-a', [
      { projectId: 'project-b', baseBranch: '' }
    ])
  })

  it('edits TODO details and confirms project removals that close terminals', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todos: [todo({ title: 'Fix login', description: 'Old description', priority: 'medium' })],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [terminal({ todoProjectId: 'todo-project-a' })]
      })
    )
    appApiMock.UpdateTodo.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todos: [todo({ title: 'Fix login redirect', description: '登录后跳回首页', priority: 'high' })],
        todoProjects: [todoProject({ id: 'todo-project-b', projectId: 'project-b' })],
        activeTodoProjectId: 'todo-project-b',
        activeProjectId: 'project-b',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-detail-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-detail-name-input"]').element.value).toBe('Fix login')
    expect(wrapper.find('[data-testid="todo-detail-description-input"]').element.value).toBe('Old description')

    await wrapper.find('[data-testid="todo-detail-name-input"]').setValue('Fix login redirect')
    await wrapper.find('[data-testid="todo-detail-description-input"]').setValue('登录后跳回首页')
    await wrapper.find('[data-testid="todo-detail-priority-high"]').setValue(true)
    await wrapper.find('[data-testid="todo-detail-selected-project-remove-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-detail-project-filter"]').setValue('api')
    await nextTick()
    await wrapper.find('[data-testid="todo-detail-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-detail-submit"]').trigger('click')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('close terminals'))
    expect(UpdateTodo).toHaveBeenCalledWith({
      id: 'todo-a',
      title: 'Fix login redirect',
      description: '登录后跳回首页',
      priority: 'high',
      projects: [{ projectId: 'project-b', baseBranch: '' }]
    })
    expect(wrapper.find('[data-testid="todo-detail-dialog"]').exists()).toBe(false)
  })

  it('updates TODO details with selected project base branches', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todos: [todo({ title: 'Fix login', description: 'Old description', priority: 'medium' })],
        todoProjects: [todoProject({ projectId: 'project-a', baseBranch: 'main' })],
        terminals: []
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()
    await wrapper.find('[data-testid="todo-detail-selected-project-branch-project-a"]').setValue('release')
    await wrapper.find('[data-testid="todo-detail-project-option-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-detail-selected-project-branch-project-b"]').setValue('develop')
    await wrapper.find('[data-testid="todo-detail-submit"]').trigger('click')
    await flushPromises()

    expect(UpdateTodo).toHaveBeenCalledWith({
      id: 'todo-a',
      title: 'Fix login',
      description: 'Old description',
      priority: 'medium',
      projects: [
        { projectId: 'project-a', baseBranch: 'release' },
        { projectId: 'project-b', baseBranch: 'develop' }
      ]
    })
  })

  it('uses workspace preferences only for newly added projects while editing TODO details', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        projectBranchPreferences: {
          'project-a': { baseBranch: 'develop' },
          'project-b': { baseBranch: 'release/2026' }
        },
        todos: [todo({ title: 'Fix login', description: 'Old description', priority: 'medium' })],
        todoProjects: [todoProject({ projectId: 'project-a', baseBranch: 'hotfix/login' })],
        terminals: []
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-detail-selected-project-branch-project-a"]').element.value).toBe('hotfix/login')

    await wrapper.find('[data-testid="todo-detail-project-option-project-b"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-detail-selected-project-branch-project-b"]').element.value).toBe('release/2026')
  })

  it('keeps TODO detail branch input editable when branch loading fails', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }
        ],
        todos: [todo({ title: 'Fix login', description: 'Old description', priority: 'medium' })],
        todoProjects: [todoProject({ projectId: 'project-a', baseBranch: 'main' })],
        terminals: []
      })
    )
    appApiMock.ListProjectBranches.mockRejectedValue(new Error('git branch list timed out'))
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await flushPromises()
    const input = wrapper.find('[data-testid="todo-detail-selected-project-branch-project-a"]')
    await input.trigger('focus')
    await input.setValue('feature/manual-branch')
    await nextTick()

    expect(wrapper.find('[data-testid="project-branch-picker-status-todo-detail-project-a"]').text()).toContain(
      'Suggestions unavailable'
    )

    await wrapper.find('[data-testid="todo-detail-submit"]').trigger('click')
    await flushPromises()

    expect(UpdateTodo).toHaveBeenCalledWith({
      id: 'todo-a',
      title: 'Fix login',
      description: 'Old description',
      priority: 'medium',
      projects: [{ projectId: 'project-a', baseBranch: 'feature/manual-branch' }]
    })
  })

  it('opens completed TODO details in read-only mode with project snapshots', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true },
          { id: 'project-b', name: 'api-service', path: '/work/api-service', available: true }
        ],
        todos: [
          completedTodo({
            title: 'Fix login',
            description: '登录后跳回首页',
            priority: 'high',
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'frontend-app',
                path: '/work/frontend-app',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-button-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-edit-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-detail-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-detail-name-input"]').element.value).toBe('Fix login')
    expect(wrapper.find('[data-testid="todo-detail-name-input"]').attributes('readonly')).toBeDefined()
    expect(wrapper.find('[data-testid="todo-detail-description-input"]').element.value).toBe('登录后跳回首页')
    expect(wrapper.find('[data-testid="todo-detail-description-input"]').attributes('readonly')).toBeDefined()
    expect(wrapper.find('[data-testid="todo-detail-priority-high"]').attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="todo-detail-submit"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-detail-project-filter"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-detail-selected-project-remove-project-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-detail-project-snapshot-project-a"]').text()).toContain('frontend-app')
    expect(wrapper.find('[data-testid="todo-detail-project-snapshot-project-a"]').text()).toContain('feature/login -> main')

    await wrapper.find('[data-testid="todo-detail-name-input"]').setValue('Changed')
    expect(UpdateTodo).not.toHaveBeenCalled()
  })

  it('keeps TODO detail editor open when terminal-closing save is cancelled', async () => {
    window.confirm.mockReturnValue(false)
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-a', name: 'frontend-app', path: '/work/frontend-app', available: true }],
        todos: [todo({ title: 'Fix login', priority: 'medium' })],
        todoProjects: [todoProject({ projectId: 'project-a' })],
        terminals: [terminal({ todoProjectId: 'todo-project-a' })]
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await wrapper.find('[data-testid="todo-detail-selected-project-remove-project-a"]').trigger('click')
    await wrapper.find('[data-testid="todo-detail-submit"]').trigger('click')
    await flushPromises()

    expect(UpdateTodo).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-detail-dialog"]').exists()).toBe(true)
  })

  it('removes a TODO project from the tree through the confirmation popover', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="remove-todo-project-todo-project-a"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="confirm-remove-todo-project-todo-project-a"]').trigger('click')
    await flushPromises()

    expect(RemoveTodoProject).toHaveBeenCalledWith('todo-project-a')
    expect(wrapper.find('[data-testid="todo-project-todo-project-a"]').exists()).toBe(false)
  })

  it('selects a TODO project context without creating a terminal', async () => {
    const wrapper = await mountReadyApp()
    CreateTodoTerminal.mockClear()

    await wrapper.find('[data-testid="todo-project-todo-project-a"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectTodoProject).toHaveBeenCalledWith('todo-project-a')
    expect(CreateTodoTerminal).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-pane-terminal-a"]').classes()).toContain('active')
  })

  it('resizes the sidebar by dragging the divider and refits the active terminal', async () => {
    const wrapper = await mountReadyApp()
    const fitAddon = xtermMock.sessions.get('terminal-a').fitAddon
    fitAddon.fit.mockClear()

    await wrapper.find('[data-testid="sidebar-resize-handle"]').trigger('mousedown', { clientX: 280 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 360 }))
    window.dispatchEvent(new MouseEvent('mouseup', { clientX: 360 }))
    await flushPromises()

    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')
    expect(fitAddon.fit).toHaveBeenCalled()
  })

  it('restores and saves TODO project UI state for the active TODO project', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 360,
        todoProjects: {
          'todo-project-a': { todoView: 'completed' },
          'todo-project-b': { todoView: 'in-progress' }
        }
      })
    )
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          todo({ id: 'todo-a', status: 'not-started' }),
          todo({ id: 'todo-b', title: 'Write tests', status: 'in-progress' }),
          completedTodo({ id: 'todo-completed', title: 'Done' })
        ],
        todoProjects: [
          todoProject({ id: 'todo-project-a', todoId: 'todo-a' }),
          todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-a' })
        ],
        activeTodoProjectId: 'todo-project-a'
      })
    )
    const todoProjectBState = projectState({
      todos: [
        todo({ id: 'todo-a', status: 'not-started' }),
        todo({ id: 'todo-b', title: 'Write tests', status: 'in-progress' }),
        completedTodo({ id: 'todo-completed', title: 'Done' })
      ],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-a' })
      ],
      activeTodoId: 'todo-b',
      activeTodoProjectId: 'todo-project-b'
    })
    const wrapper = await mountReadyApp()

    expect(LoadTodoProjectUIState).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-view-completed"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    runtimeMock.handlers['workspace-state'](todoProjectBState)
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-view-completed"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    await wrapper.find('[data-testid="todo-view-not-started"]').trigger('click')
    await flushPromises()

    expect(SaveTodoProjectUIState).toHaveBeenLastCalledWith('todo-project-b', {
      todoView: 'not-started'
    })
    expect(SaveTodoSidebarWidth).not.toHaveBeenCalled()
  })

  it('keeps current TODO view and workspace sidebar width when the user selects a TODO project item', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 360,
        todoProjects: {
          'todo-project-a': { todoView: 'not-started' },
          'todo-project-b': { todoView: 'in-progress' }
        }
      })
    )
    const selectedTodoProjectState = projectState({
      todos: [
        todo({ id: 'todo-a', status: 'not-started' }),
        todo({ id: 'todo-b', title: 'Write tests', status: 'not-started' })
      ],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-a' })
      ],
      activeTodoId: 'todo-b',
      activeTodoProjectId: 'todo-project-b'
    })
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          todo({ id: 'todo-a', status: 'not-started' }),
          todo({ id: 'todo-b', title: 'Write tests', status: 'not-started' })
        ],
        todoProjects: [
          todoProject({ id: 'todo-project-a', todoId: 'todo-a' }),
          todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-a' })
        ],
        activeTodoProjectId: 'todo-project-a'
      })
    )
    appApiMock.SelectTodoProject.mockResolvedValue(selectedTodoProjectState)
    const wrapper = await mountReadyApp()

    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectTodoProject).toHaveBeenCalledWith('todo-project-b')
    expect(wrapper.find('[data-testid="todo-view-not-started"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')
  })

  it('keeps not-started view when selecting a TODO project item after selecting an in-progress terminal', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 380,
        todoProjects: {
          'todo-project-a': { todoView: 'in-progress' },
          'todo-project-b': { todoView: 'in-progress' }
        }
      })
    )
    const baseState = projectState({
      projects: [
        { id: 'project-a', name: 'frontend', path: '/work/frontend', available: true },
        { id: 'project-b', name: 'api', path: '/work/api', available: true }
      ],
      todos: [
        todo({ id: 'todo-a', status: 'in-progress' }),
        todo({ id: 'todo-b', title: 'Plan tests', status: 'not-started' })
      ],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b' })
      ],
      activeTodoId: 'todo-a',
      activeTodoProjectId: 'todo-project-a',
      activeProjectId: 'project-a',
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
      ],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(baseState)
    appApiMock.SelectTerminal.mockResolvedValue({
      ...baseState,
      activeTodoId: 'todo-a',
      activeTodoProjectId: 'todo-project-a',
      activeProjectId: 'project-a',
      activeTerminalId: 'terminal-a'
    })
    appApiMock.SelectTodoProject.mockResolvedValue({
      ...baseState,
      activeTodoId: 'todo-b',
      activeTodoProjectId: 'todo-project-b',
      activeProjectId: 'project-b'
    })
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')

    await wrapper.find('[data-testid="todo-view-not-started"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="todo-view-not-started"]').classes()).toContain('active')

    await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectTerminal).toHaveBeenCalledWith('terminal-a')
    expect(appApiMock.SelectTodoProject).toHaveBeenCalledWith('todo-project-b')
    expect(wrapper.find('[data-testid="todo-view-not-started"]').classes()).toContain('active')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 380px')
  })

  it('keeps sidebar width shared across projects under the same TODO item', async () => {
    appApiMock.LoadTodoProjectUIState.mockResolvedValue(
      todoProjectUIStateFile({
        sidebarWidth: 360,
        todoProjects: {
          'todo-project-a': { todoView: 'not-started' },
          'todo-project-b': { todoView: 'not-started' }
        }
      })
    )
    const baseState = projectState({
      projects: [
        { id: 'project-a', name: 'frontend', path: '/work/frontend', available: true },
        { id: 'project-b', name: 'api', path: '/work/api', available: true }
      ],
      todos: [todo({ id: 'todo-a', status: 'not-started' })],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-a', projectId: 'project-b' })
      ],
      activeTodoId: 'todo-a',
      activeTodoProjectId: 'todo-project-a',
      activeProjectId: 'project-a'
    })
    appApiMock.ListProjects.mockResolvedValue(baseState)
    appApiMock.SelectTodoProject.mockImplementation((todoProjectId) =>
      Promise.resolve({
        ...baseState,
        activeTodoProjectId: todoProjectId,
        activeProjectId: todoProjectId === 'todo-project-b' ? 'project-b' : 'project-a'
      })
    )
    const wrapper = await mountReadyApp()

    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectTodoProject).toHaveBeenCalledWith('todo-project-b')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')

    await wrapper.find('[data-testid="todo-project-todo-project-a"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectTodoProject).toHaveBeenCalledWith('todo-project-a')
    expect(wrapper.find('.app-shell').attributes('style')).toContain('--sidebar-width: 360px')
  })

  it('saves sidebar width after divider dragging ends for the workspace', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="sidebar-resize-handle"]').trigger('mousedown', { clientX: 280 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 360 }))
    expect(SaveTodoSidebarWidth).not.toHaveBeenCalled()
    window.dispatchEvent(new MouseEvent('mouseup', { clientX: 360 }))
    await flushPromises()

    expect(SaveTodoSidebarWidth).toHaveBeenLastCalledWith(360)
    expect(SaveTodoProjectUIState).not.toHaveBeenCalled()
  })

  it('serializes TODO project view and workspace sidebar width saves independently', async () => {
    let resolveFirstSave
    const firstSave = new Promise((resolve) => {
      resolveFirstSave = resolve
    })
    appApiMock.SaveTodoProjectUIState
      .mockImplementationOnce(() => firstSave)
      .mockResolvedValue()
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await flushPromises()
    expect(SaveTodoProjectUIState).toHaveBeenCalledTimes(1)
    expect(SaveTodoProjectUIState).toHaveBeenLastCalledWith('todo-project-a', {
      todoView: 'in-progress'
    })

    await wrapper.find('[data-testid="sidebar-resize-handle"]').trigger('mousedown', { clientX: 280 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 360 }))
    window.dispatchEvent(new MouseEvent('mouseup', { clientX: 360 }))
    await flushPromises()

    expect(SaveTodoProjectUIState).toHaveBeenCalledTimes(1)
    expect(SaveTodoSidebarWidth).toHaveBeenCalledTimes(1)
    expect(SaveTodoSidebarWidth).toHaveBeenLastCalledWith(360)

    resolveFirstSave()
    await flushPromises()

    expect(SaveTodoProjectUIState).toHaveBeenCalledTimes(1)
  })

  it('completes a TODO and shows its completed snapshot', async () => {
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await nextTick()
    expect(CompleteTodo).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')

    expect(window.confirm).not.toHaveBeenCalled()
    expect(CompleteTodo).toHaveBeenCalledWith('todo-a')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="completed-todos"]').text()).toContain('completed')
    expect(wrapper.find('[data-testid="completed-todos"]').text()).toContain('feature/alpha -> main')
  })

  it('checks completed snapshot merge status asynchronously when the completed view opens', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses.mockResolvedValue([
      { id: 'todo-a::project-a::/work/alpha::0', status: 'merged' }
    ])
    const wrapper = await mountReadyApp()

    expect(GetCompletedTodoProjectMergeStatuses).not.toHaveBeenCalled()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledWith([
      {
        id: 'todo-a::project-a::/work/alpha::0',
        todoId: 'todo-a',
        snapshotIndex: 0,
        path: '/work/alpha',
        worktreeBranch: 'feature/login',
        baseBranch: 'main',
        fingerprint: '/work/alpha::feature/login::main'
      }
    ])
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')
  })

  it('uses persisted completed snapshot merge status without querying again', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main',
                mergeStatus: 'confirmed',
                mergeStatusReason: 'worktree-removed'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')
  })

  it('does not recheck a completed snapshot after it has shown a checkmark', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses.mockResolvedValue([
      { id: 'todo-a::project-a::/work/alpha::0', status: 'merged', reason: 'worktree-removed' }
    ])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')

    await wrapper.find('[data-testid="todo-view-not-started"]').trigger('click')
    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledTimes(1)
  })

  it('rechecks completed snapshot merge status when returning to the completed view', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses
      .mockResolvedValueOnce([{ id: 'todo-a::project-a::/work/alpha::0', status: 'unmerged' }])
      .mockResolvedValueOnce([{ id: 'todo-a::project-a::/work/alpha::0', status: 'merged' }])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('unmerged')

    await wrapper.find('[data-testid="todo-view-not-started"]').trigger('click')
    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')
  })

  it('keeps stale completed merge-status responses from replacing newer results', async () => {
    const staleRequest = deferred()
    const freshRequest = deferred()
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses
      .mockReturnValueOnce(staleRequest.promise)
      .mockReturnValueOnce(freshRequest.promise)
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    runtimeMock.handlers['workspace-state'](
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              },
              {
                projectId: 'project-b',
                name: 'beta',
                path: '/work/beta',
                worktreeBranch: 'feature/beta',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledTimes(2)

    freshRequest.resolve([
      { id: 'todo-a::project-a::/work/alpha::0', status: 'merged' },
      { id: 'todo-a::project-b::/work/beta::1', status: 'unmerged' }
    ])
    await flushPromises()
    staleRequest.resolve([{ id: 'todo-a::project-a::/work/alpha::0', status: 'unmerged' }])
    await flushPromises()

    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-b-1"]').classes()).toContain('unmerged')
  })

  it('reuses cached completed merge statuses and only requests new completed snapshots', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses
      .mockResolvedValueOnce([{ id: 'todo-a::project-a::/work/alpha::0', status: 'merged' }])
      .mockResolvedValueOnce([{ id: 'todo-a::project-b::/work/beta::1', status: 'unmerged' }])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    runtimeMock.handlers['workspace-state'](
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              },
              {
                projectId: 'project-b',
                name: 'beta',
                path: '/work/beta',
                worktreeBranch: 'feature/beta',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledTimes(2)
    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenLastCalledWith([
      {
        id: 'todo-a::project-b::/work/beta::1',
        todoId: 'todo-a',
        snapshotIndex: 1,
        path: '/work/beta',
        worktreeBranch: 'feature/beta',
        baseBranch: 'main',
        fingerprint: '/work/beta::feature/beta::main'
      }
    ])
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('merged')
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-b-1"]').classes()).toContain('unmerged')
  })

  it('rechecks a cached completed snapshot when its branch pair changes', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'main'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    appApiMock.GetCompletedTodoProjectMergeStatuses
      .mockResolvedValueOnce([{ id: 'todo-a::project-a::/work/alpha::0', status: 'merged' }])
      .mockResolvedValueOnce([{ id: 'todo-a::project-a::/work/alpha::0', status: 'unmerged' }])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await flushPromises()

    runtimeMock.handlers['workspace-state'](
      projectState({
        todos: [
          completedTodo({
            projectSnapshots: [
              {
                projectId: 'project-a',
                name: 'alpha',
                path: '/work/alpha',
                worktreeBranch: 'feature/login',
                baseBranch: 'release/2026'
              }
            ]
          })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    await flushPromises()

    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenCalledTimes(2)
    expect(GetCompletedTodoProjectMergeStatuses).toHaveBeenLastCalledWith([
      {
        id: 'todo-a::project-a::/work/alpha::0',
        todoId: 'todo-a',
        snapshotIndex: 0,
        path: '/work/alpha',
        worktreeBranch: 'feature/login',
        baseBranch: 'release/2026',
        fingerprint: '/work/alpha::feature/login::release/2026'
      }
    ])
    expect(wrapper.find('[data-testid="completed-project-merge-status-todo-a-project-a-0"]').classes()).toContain('unmerged')
  })

  it('deletes a completed TODO after completed-view confirmation', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [completedTodo()],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-button-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="completed-todo-menu-delete-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    expect(DeleteTodo).not.toHaveBeenCalled()

    await wrapper.find('[data-testid="confirm-delete-todo-todo-a"]').trigger('click')
    await flushPromises()

    expect(DeleteTodo).toHaveBeenCalledWith('todo-a')
    expect(wrapper.find('[data-testid="completed-todo-todo-a"]').exists()).toBe(false)
  })

  it('bulk deletes selected completed TODOs from the completed view', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [
          completedTodo({ id: 'todo-a', title: '完成 A' }),
          completedTodo({ id: 'todo-b', title: '完成 B', completedAt: '2026-06-10T11:00:00Z' })
        ],
        todoProjects: [],
        terminals: [],
        activeTodoId: '',
        activeTodoProjectId: '',
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-completed"]').trigger('click')
    await wrapper.find('[data-testid="select-completed-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="select-completed-todo-todo-b"]').trigger('click')
    await wrapper.find('[data-testid="bulk-delete-completed-todos"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="bulk-delete-completed-todos-popover"]').exists()).toBe(true)
    expect(DeleteCompletedTodos).not.toHaveBeenCalled()

    await wrapper.find('[data-testid="confirm-bulk-delete-completed-todos"]').trigger('click')
    await flushPromises()

    expect(DeleteCompletedTodos).toHaveBeenCalledWith(['todo-a', 'todo-b'])
    expect(wrapper.find('[data-testid="completed-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="completed-todo-todo-b"]').exists()).toBe(false)
  })

  it('does not delete a TODO when the sidebar confirmation is cancelled', async () => {
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'delete', 'todo-a')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    await wrapper.find('[data-testid="cancel-delete-todo-todo-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(DeleteTodo).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
  })

  it('shows only start and management actions for not-started TODOs', async () => {
    appApiMock.ListProjects.mockResolvedValue(projectState({ todos: [todo({ status: 'not-started' })] }))
    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="mark-todo-in-progress-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="delete-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="edit-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-project-to-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="complete-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mark-todo-not-started-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').exists()).toBe(false)

    await openTodoContextMenu(wrapper, 'todo-a')
    expect(wrapper.find('[data-testid="todo-menu-edit-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-menu-add-project-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-menu-delete-todo-a"]').exists()).toBe(true)
  })

  it('shows complete and terminal actions for in-progress TODOs', async () => {
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.find('[data-testid="complete-todo-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="delete-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="edit-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-project-to-todo-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="mark-todo-not-started-todo-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mark-todo-in-progress-todo-a"]').exists()).toBe(false)

    await openTodoContextMenu(wrapper, 'todo-a')
    expect(wrapper.find('[data-testid="todo-menu-edit-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-menu-add-project-todo-a"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-menu-delete-todo-a"]').exists()).toBe(true)
  })

  it('shows task terminals before project rows and creates task terminals from the TODO tree', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        terminals: [
          taskTerminal({ id: 'task-terminal-a', shellName: 'bash' }),
          terminal({ id: 'terminal-a', shellName: 'zsh' })
        ],
        activeTerminalId: 'task-terminal-a'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await flushPromises()

    const taskTerminalRow = wrapper.find('[data-testid="task-terminal-task-terminal-a"]')
    const projectRow = wrapper.find('[data-testid="todo-project-todo-project-a"]')
    expect(taskTerminalRow.exists()).toBe(true)
    expect(projectRow.exists()).toBe(true)
    expect(taskTerminalRow.element.compareDocumentPosition(projectRow.element) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(wrapper.find('[data-testid="terminal-task-terminal-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="add-task-terminal-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-task-todo-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTaskTerminal).toHaveBeenCalledWith('todo-a', 100, 32)
  })

  it('creates a task terminal from a custom launch profile and submits its command', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Codex GPT-5', command: 'codex --model gpt-5', enabled: true }] })
    )
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.CreateTaskTerminal.mockResolvedValue(
      inProgressProjectState({
        terminals: [taskTerminal({ id: 'task-terminal-b', shellName: 'bash', state: 'running' })],
        activeTerminalId: 'task-terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await wrapper.find('[data-testid="add-task-terminal-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-task-todo-a-1"]').trigger('click')
    await flushPromises()

    expect(CreateTaskTerminal).toHaveBeenCalledWith('todo-a', 80, 24)
    expect(SendTerminalInput).toHaveBeenCalledWith('task-terminal-b', 'codex --model gpt-5\r')
    expect(wrapper.find('[data-testid="task-terminal-task-terminal-b"]').text()).toContain('codex --model gpt-5')
  })

  it('opens task folders from the TODO row menu', async () => {
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState())
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-open-folder-todo-a"]').trigger('click')
    await flushPromises()

    expect(OpenTodoFolder).toHaveBeenCalledWith('todo-a')
  })

  it('copies Unicode TODO title and description from the TODO row menu', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        todos: [
          todo({
            title: '修复登录问题 ✓',
            description: '登录后跳回首页 🔧 ',
            status: 'in-progress'
          })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-copy-title-description-todo-a"]').trigger('click')
    await flushPromises()

    expect(ClipboardSetText).toHaveBeenCalledWith('修复登录问题 ✓\n登录后跳回首页 🔧 ')
    expect(wrapper.find('[data-testid="todo-context-menu-todo-a"]').exists()).toBe(false)
  })

  it('reports TODO clipboard copy failures without changing TODO data', async () => {
    runtimeMock.ClipboardSetText.mockRejectedValueOnce(new Error('clipboard unavailable'))
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        todos: [
          todo({
            title: '修复登录问题',
            description: '登录后跳回首页',
            status: 'in-progress'
          })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    const beforeText = wrapper.find('[data-testid="todo-todo-a"]').text()
    await openTodoContextMenu(wrapper, 'todo-a')
    await wrapper.find('[data-testid="todo-menu-copy-title-description-todo-a"]').trigger('click')
    await flushPromises()

    expect(ClipboardSetText).toHaveBeenCalledWith('修复登录问题\n登录后跳回首页')
    expect(wrapper.find('.status-error').text()).toContain('clipboard unavailable')
    expect(wrapper.find('[data-testid="todo-todo-a"]').text()).toBe(beforeText)
    expect(UpdateTodo).not.toHaveBeenCalled()
    expect(ChangeTodoStatus).not.toHaveBeenCalled()
  })

  it('shows failed worktree status and blocks project terminal creation', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            worktreeStatus: 'failed',
            worktreeError: 'branch is already checked out'
          })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    expect(wrapper.find('[data-testid="todo-project-worktree-error-todo-project-a"]').text()).toContain(
      'branch is already checked out'
    )
    expect(wrapper.find('[data-testid="add-terminal-todo-project-a"]').attributes('disabled')).toBeDefined()

    CreateTodoTerminal.mockClear()
    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).not.toHaveBeenCalled()
  })

  it('changes TODO workflow status from the sidebar', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="mark-todo-in-progress-todo-a"]').trigger('click')
    await flushPromises()

    expect(ChangeTodoStatus).toHaveBeenCalledWith('todo-a', 'in-progress')
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
  })

  it('stays on in-progress view after starting a TODO without an active TODO project', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todos: [todo({ id: 'todo-a', status: 'not-started' })],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    appApiMock.ChangeTodoStatus.mockResolvedValue(
      projectState({
        todos: [todo({ id: 'todo-a', status: 'in-progress' })],
        todoProjects: [],
        activeProjectId: '',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="mark-todo-in-progress-todo-a"]').trigger('click')
    await flushPromises()

    expect(ChangeTodoStatus).toHaveBeenCalledWith('todo-a', 'in-progress')
    expect(wrapper.find('[data-testid="todo-view-in-progress"]').classes()).toContain('active')
    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
  })

  it('imports projects from a parent directory and shows the summary', async () => {
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-global-project-candidates"]').trigger('click')
    await flushPromises()

    expect(ImportProjectsFromParentDirectoryDialog).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="import-summary"]').text()).toContain('2 imported')
    expect(wrapper.find('[data-testid="import-summary"]').text()).toContain('1 skipped')
  })

  it('shows single project import in every project candidate picker', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="import-single-project-candidate"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-create-close"]').trigger('click')
    await nextTick()
    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()
    expect(wrapper.find('[data-testid="import-single-project-candidate"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-detail-close"]').trigger('click')
    await nextTick()
    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    expect(wrapper.find('[data-testid="import-single-project-candidate"]').exists()).toBe(true)
    await wrapper.find('[data-testid="todo-project-picker-close"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="import-single-project-candidate"]').exists()).toBe(false)
  })

  it('imports a single project candidate without refreshing git status immediately', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue(projectImportResult(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ]
      })
    ))
    const wrapper = await mountReadyApp()
    GetProjectGitStatus.mockClear()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()

    expect(CreateProjectFromDialog).toHaveBeenCalled()
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-project-picker-option-project-b"]').exists()).toBe(true)
  })

  it('initializes a non-Git single project after confirmation and selects it', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue({
      requiresGitInitialization: true,
      path: '/work/beta'
    })
    appApiMock.InitializeGitRepositoryAndImportProject.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        importSummary: {
          parentPath: '/work/beta',
          addedCount: 1,
          skippedCount: 0,
          added: [{ id: 'project-b', name: 'beta', path: '/work/beta', available: true }]
        }
      })
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="git-init-confirm-path"]').text()).toBe('/work/beta')
    expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').text()).toContain('暂存当前目录内容')
    expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').text()).toContain('初始提交')
    expect(InitializeGitRepositoryAndImportProject).not.toHaveBeenCalled()

    await wrapper.find('[data-testid="git-init-confirm-submit"]').trigger('click')
    await flushPromises()

    expect(InitializeGitRepositoryAndImportProject).toHaveBeenCalledWith('/work/beta')
    expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-b"]').exists()).toBe(true)
  })

  it('shows an error and keeps the project unselected when non-Git initialization commit fails', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue({
      requiresGitInitialization: true,
      path: '/work/beta'
    })
    appApiMock.InitializeGitRepositoryAndImportProject.mockRejectedValue(
      new Error('git commit failed: Author identity unknown')
    )
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()

    await wrapper.find('[data-testid="git-init-confirm-submit"]').trigger('click')
    await flushPromises()

    expect(InitializeGitRepositoryAndImportProject).toHaveBeenCalledWith('/work/beta')
    expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').exists()).toBe(false)
    expect(wrapper.find('.status-error').text()).toContain('git commit failed')
    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-b"]').exists()).toBe(false)
  })

  it('shows a temporary toast when non-Git single project initialization is declined', async () => {
    vi.useFakeTimers()
    try {
      appApiMock.CreateProjectFromDialog.mockResolvedValue({
        requiresGitInitialization: true,
        path: '/work/beta'
      })
      const wrapper = await mountReadyApp()

      await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
      await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
      await flushPromises()

      expect(window.confirm).not.toHaveBeenCalled()
      expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').exists()).toBe(true)

      await wrapper.find('[data-testid="git-init-confirm-cancel"]').trigger('click')
      await flushPromises()

      expect(InitializeGitRepositoryAndImportProject).not.toHaveBeenCalled()
      expect(wrapper.find('[data-testid="git-init-confirm-dialog"]').exists()).toBe(false)
      expect(wrapper.find('[data-testid="app-toast"]').text()).toContain('只能导入 Git 仓库')

      vi.advanceTimersByTime(1999)
      await nextTick()
      expect(wrapper.find('[data-testid="app-toast"]').exists()).toBe(true)

      vi.advanceTimersByTime(1)
      await nextTick()
      expect(wrapper.find('[data-testid="app-toast"]').exists()).toBe(false)
    } finally {
      vi.useRealTimers()
    }
  })

  it('renders import toast above modal overlays', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue({
      requiresGitInitialization: true,
      path: '/work/beta'
    })
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="git-init-confirm-cancel"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="app-toast"]').exists()).toBe(true)

    const styles = readFileSync('src/style.css', 'utf8')
    const toastRule = styles.match(/\.app-toast\s*{([^}]*)}/s)?.[1] || ''
    const settingsOverlayRule = styles.match(/\.settings-overlay\s*{([^}]*)}/s)?.[1] || ''
    const gitInitOverlayRule = styles.match(/\.git-init-confirm-overlay\s*{([^}]*)}/s)?.[1] || ''
    const zIndexFromRule = (rule) => Number(rule.match(/z-index:\s*(\d+);/)?.[1])

    expect(zIndexFromRule(toastRule)).toBeGreaterThan(zIndexFromRule(settingsOverlayRule))
    expect(zIndexFromRule(toastRule)).toBeGreaterThan(zIndexFromRule(gitInitOverlayRule))
  })

  it('selects a single imported project in the create TODO form only after import', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue(projectImportResult(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        importSummary: {
          parentPath: '/work/beta',
          addedCount: 1,
          skippedCount: 0,
          added: [{ id: 'project-b', name: 'beta', path: '/work/beta', available: true }]
        }
      })
    ))
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-selected-project-remove-project-b"]').trigger('click')
    await wrapper.find('[data-testid="todo-name-input"]').setValue('New task')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'New task',
      description: '',
      priority: 'medium',
      projects: []
    })
  })

  it('accepts Unicode paste text in the create TODO name and description fields', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    const nameInput = wrapper.find('[data-testid="todo-name-input"]')
    const descriptionInput = wrapper.find('[data-testid="todo-description-input"]')

    await nameInput.trigger('paste', {
      clipboardData: { getData: () => '修复登录问题' }
    })
    await nameInput.setValue('修复登录问题')
    await descriptionInput.trigger('paste', {
      clipboardData: { getData: () => '登录后跳回首页 🔧' }
    })
    await descriptionInput.setValue('登录后跳回首页 🔧')

    expect(nameInput.element.value).toBe('修复登录问题')
    expect(descriptionInput.element.value).toBe('登录后跳回首页 🔧')
  })

  it('selects an existing project candidate when single import skips a duplicate', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ]
      })
    )
    appApiMock.CreateProjectFromDialog.mockResolvedValue(projectImportResult(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        importSummary: {
          parentPath: '/work',
          addedCount: 0,
          skippedCount: 1,
          skippedPaths: ['/work/beta']
        }
      })
    ))
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-selected-project-tag-project-b"]').exists()).toBe(true)
  })

  it('selects a single imported project in the edit TODO and add-project pickers', async () => {
    appApiMock.CreateProjectFromDialog.mockResolvedValue(projectImportResult(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        importSummary: {
          parentPath: '/work/beta',
          addedCount: 1,
          skippedCount: 0,
          added: [{ id: 'project-b', name: 'beta', path: '/work/beta', available: true }]
        }
      })
    ))
    const wrapper = await mountReadyApp()

    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="todo-detail-selected-project-tag-project-b"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-detail-close"]').trigger('click')
    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-single-project-candidate"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="todo-project-picker-tag-project-b"]').exists()).toBe(true)
  })

  it('explains that bulk parent import only imports first-level Git repositories', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="import-global-project-candidates"]').attributes('title')).toBe(
      '仅导入一级子目录中的 Git 仓库'
    )

    await wrapper.find('[data-testid="todo-create-close"]').trigger('click')
    await selectTodoMenuAction(wrapper, 'edit', 'todo-a')
    await nextTick()
    expect(wrapper.find('[data-testid="import-global-project-candidates"]').attributes('title')).toBe(
      '仅导入一级子目录中的 Git 仓库'
    )

    await wrapper.find('[data-testid="todo-detail-close"]').trigger('click')
    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await nextTick()
    expect(wrapper.find('[data-testid="import-global-project-candidates"]').attributes('title')).toBe(
      '仅导入一级子目录中的 Git 仓库'
    )
  })

  it('does not refresh git status immediately after importing projects', async () => {
    appApiMock.ImportProjectsFromParentDirectoryDialog.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        activeProjectId: 'project-b',
        activeTodoId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: '',
        importSummary: { parentPath: '/work', addedCount: 2, skippedCount: 0 }
      })
    )
    const wrapper = await mountReadyApp()
    GetProjectGitStatus.mockClear()

    await selectTodoMenuAction(wrapper, 'add-project', 'todo-a')
    await wrapper.find('[data-testid="import-global-project-candidates"]').trigger('click')
    await flushPromises()

    expect(ImportProjectsFromParentDirectoryDialog).toHaveBeenCalled()
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('main')
  })

  it('does not render the removed project tab selection path', async () => {
    const wrapper = await mountReadyApp()
    CreateTodoTerminal.mockClear()

    expect(wrapper.find('[data-testid="sidebar-tab-projects"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(false)
    expect(appApiMock.SelectProject).not.toHaveBeenCalled()
    expect(CreateTodoTerminal).not.toHaveBeenCalled()
  })

  it('deletes a terminal from the project tree without confirmation', async () => {
    const wrapper = await mountReadyApp()
    const session = xtermMock.sessions.get('terminal-a')

    await wrapper.find('[data-testid="delete-terminal-terminal-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(DeleteTerminal).toHaveBeenCalledWith('terminal-a')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(session.terminal.dispose).toHaveBeenCalledTimes(1)
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

  it('updates terminal labels from backend command-state events', async () => {
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-command-state']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      type: 'command-start',
      command: 'codex'
    })
    runtimeMock.handlers['terminal-output']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      data: 'ready\r\n'
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('codex')
    expect(xtermMock.sessions.get('terminal-a').terminal.write).toHaveBeenCalledWith('ready\r\n')
    expect(xtermMock.sessions.get('terminal-a').terminal.write).not.toHaveBeenCalledWith(
      expect.stringContaining('tui-helper')
    )

    runtimeMock.handlers['terminal-command-state']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      type: 'command-end'
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).toContain('zsh')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').text()).not.toContain('codex')
  })

  it('ignores invalid backend command-state events without changing agent status', async () => {
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      label: 'codex',
      updatedAt: 10
    })
    await nextTick()

    runtimeMock.handlers['terminal-command-state']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      type: 'command-start',
      command: ''
    })
    await nextTick()

    const terminalRow = wrapper.find('[data-testid="terminal-terminal-a"]')
    expect(terminalRow.text()).toContain('zsh')
    expect(terminalRow.attributes('data-activity-state')).toBe('busy')
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

  it('marks an interactive agent busy from title changes without replacing the command label', async () => {
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

  it('uses title change activity and returns idle after one second without changes', async () => {
    vi.useFakeTimers()
    try {
      const wrapper = await mountReadyApp()

      xtermMock.sessions.get('terminal-a').onTitleChange('codex')
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')

      vi.advanceTimersByTime(999)
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')

      vi.advanceTimersByTime(1)
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('idle')

      xtermMock.sessions.get('terminal-a').onTitleChange('codex working')
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('busy')

      vi.advanceTimersByTime(1000)
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('idle')
    } finally {
      vi.useRealTimers()
    }
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

  it('marks a background title fallback terminal needs-ack when it returns idle', async () => {
    vi.useFakeTimers()
    try {
      const twoTerminalState = projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-a'
      })
      appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
      appApiMock.SelectTerminal
        .mockResolvedValueOnce(
          projectState({
            terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
            activeTerminalId: 'terminal-b'
          })
        )
        .mockResolvedValueOnce(twoTerminalState)
      const wrapper = await mountReadyApp()

      await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
      await flushPromises()
      await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
      await flushPromises()

      xtermMock.sessions.get('terminal-b').onTitleChange('codex working')
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('busy')

      vi.advanceTimersByTime(1000)
      await nextTick()

      expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('needs-ack')
    } finally {
      vi.useRealTimers()
    }
  })

  it('keeps a collapsed TODO collapsed and marks it needs-ack when a hidden background terminal returns idle', async () => {
    vi.useFakeTimers()
    try {
      const twoTerminalState = projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-a'
      })
      appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
      appApiMock.SelectTerminal
        .mockResolvedValueOnce(
          projectState({
            terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
            activeTerminalId: 'terminal-b'
          })
        )
        .mockResolvedValueOnce(twoTerminalState)
      const wrapper = await mountReadyApp()

      await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
      await flushPromises()
      await wrapper.find('[data-testid="terminal-terminal-a"]').trigger('click')
      await flushPromises()
      await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')
      await nextTick()

      xtermMock.sessions.get('terminal-b').onTitleChange('codex working')
      await nextTick()
      vi.advanceTimersByTime(1000)
      await nextTick()

      expect(wrapper.find('[data-testid="todo-project-list-todo-a"]').exists()).toBe(false)
      expect(wrapper.find('[data-testid="todo-todo-a"]').attributes('data-activity-state')).toBe('needs-ack')
    } finally {
      vi.useRealTimers()
    }
  })

  it.each([
    ['idle', { phase: 'idle', source: 'codex-jsonl', reason: 'turn-idle' }],
    ['done', { phase: 'done', source: 'codex-jsonl', reason: 'turn-completed' }],
    ['failed', { phase: 'failed', source: 'codex-jsonl', reason: 'turn-failed' }],
    ['exited', { phase: 'exited', source: 'shell', reason: 'shell-exited', shellState: 'exited' }]
  ])('marks a background terminal needs-ack when busy becomes %s', async (_label, event) => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-b',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      updatedAt: 10
    })
    await nextTick()

    if (event.shellState) {
      runtimeMock.handlers['terminal-status']({
        projectId: 'project-a',
        terminalId: 'terminal-b',
        state: event.shellState
      })
    } else {
      runtimeMock.handlers['terminal-agent-status']({
        projectId: 'project-a',
        terminalId: 'terminal-b',
        phase: event.phase,
        source: event.source,
        confidence: 'authoritative',
        reason: event.reason,
        updatedAt: 20
      })
    }
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('needs-ack')
  })

  it('does not mark the active terminal needs-ack when busy becomes idle', async () => {
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      updatedAt: 10
    })
    await nextTick()
    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-a',
      phase: 'idle',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-idle',
      updatedAt: 20
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-a"]').attributes('data-activity-state')).toBe('idle')
  })

  it('clears needs-ack when selecting the corresponding terminal', async () => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    appApiMock.SelectTerminal.mockResolvedValue(
      projectState({
        terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-b',
      phase: 'busy',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-started',
      updatedAt: 10
    })
    await nextTick()
    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-b',
      phase: 'done',
      source: 'codex-jsonl',
      confidence: 'authoritative',
      reason: 'turn-completed',
      updatedAt: 20
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('needs-ack')

    await wrapper.find('[data-testid="terminal-terminal-b"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('idle')
  })

  it('does not mark a background terminal needs-ack when needs-input becomes idle', async () => {
    const twoTerminalState = projectState({
      terminals: [terminal({ id: 'terminal-a' }), terminal({ id: 'terminal-b', shellName: 'bash' })],
      activeTerminalId: 'terminal-a'
    })
    appApiMock.ListProjects.mockResolvedValue(twoTerminalState)
    const wrapper = await mountReadyApp()

    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-b',
      phase: 'needs-input',
      source: 'claude-hook',
      confidence: 'structured',
      reason: 'permission-prompt',
      updatedAt: 10
    })
    await nextTick()
    runtimeMock.handlers['terminal-agent-status']({
      projectId: 'project-a',
      terminalId: 'terminal-b',
      phase: 'idle',
      source: 'claude-hook',
      confidence: 'structured',
      reason: 'stop',
      updatedAt: 20
    })
    await nextTick()

    expect(wrapper.find('[data-testid="terminal-terminal-b"]').attributes('data-activity-state')).toBe('idle')
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
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).toHaveBeenCalledWith('project-a')
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('main')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('3 changed')
  })

  it('shows the active TODO project worktree git branch and changed file count', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'main', changedCount: 0 }))
    appApiMock.GetTodoProjectGitStatus.mockResolvedValue(
      gitStatus({ branch: 'todo/fix-login/frontend-app', changedCount: 2 })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('todo/fix-login/frontend-app')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('2 changed')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('main')
    expect(wrapper.find('.heading-path').text()).toContain('/work/customer-a/tasks/abc123/alpha')
  })

  it('uses the TODO workspace git status for task terminals and hides chips when it is not a repo', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'previous-project', changedCount: 4 }))
    appApiMock.GetTodoGitStatus.mockResolvedValue(gitStatus({ projectId: '', isRepo: false, branch: '' }))
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        activeTodoProjectId: '',
        terminals: [taskTerminal({ id: 'task-terminal-a', state: 'running' })],
        activeTerminalId: 'task-terminal-a'
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoGitStatus).toHaveBeenCalledWith('todo-a')
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').findAll('.status-chip')).toHaveLength(0)
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('previous-project')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('Not a git repository')
  })

  it('shows task terminal heading and task workspace path for active task context', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'previous-project', changedCount: 4 }))
    appApiMock.GetTodoGitStatus.mockResolvedValue(gitStatus({ projectId: '', branch: 'todo/root', changedCount: 0 }))
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        activeProjectId: '',
        activeTodoId: 'todo-a',
        activeTodoProjectId: '',
        todos: [todo({ status: 'in-progress', workspaceDirName: 'abc123' })],
        terminals: [taskTerminal({ id: 'task-terminal-a', state: 'running' })],
        activeTerminalId: 'task-terminal-a'
      })
    )

    const wrapper = await mountReadyApp()

    expect(wrapper.find('.heading-name').text()).toBe('Fix login / 任务终端')
    expect(wrapper.find('.heading-path').text()).toBe('/work/customer-a/tasks/abc123')
    expect(wrapper.find('.heading-path').text()).not.toContain('/work/alpha')
  })

  it('shows the TODO workspace git branch and changed file count for task terminals', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'previous-project', changedCount: 4 }))
    appApiMock.GetTodoGitStatus.mockResolvedValue(gitStatus({ projectId: '', branch: 'todo/root', changedCount: 1 }))
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        activeTodoProjectId: '',
        terminals: [taskTerminal({ id: 'task-terminal-a', state: 'running' })],
        activeTerminalId: 'task-terminal-a'
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoGitStatus).toHaveBeenCalledWith('todo-a')
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('todo/root')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('1 changed')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('previous-project')
  })

  it('shows the TODO project sidebar branch from live worktree git status without changing the heading', async () => {
    appApiMock.GetTodoProjectGitStatus.mockResolvedValue(gitStatus({ branch: 'feature/live-worktree' }))
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            todoId: 'todo-a',
            projectId: 'project-a',
            worktreeBranch: 'todo/static-worktree-branch'
          })
        ]
      })
    )

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toContain(
      'alpha(feature/live-worktree)'
    )
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).not.toContain(
      'todo/static-worktree-branch'
    )
    expect(wrapper.find('.heading-name').text()).toBe('Fix login / alpha')
  })

  it('shows the TODO project sidebar cleared label from persisted worktree state', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            todoId: 'todo-a',
            projectId: 'project-a',
            worktreeStatus: 'cleared',
            worktreeBranch: 'todo/static-worktree-branch'
          })
        ]
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toBe(
      'alpha(worktree已清除)'
    )
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).not.toContain(
      'todo/static-worktree-branch'
    )
  })

  it('shows the TODO project sidebar cleared label from git status refresh results', async () => {
    appApiMock.GetTodoProjectGitStatus.mockResolvedValue(
      gitStatus({
        branch: 'todo/static-worktree-branch',
        pathUnavailable: true,
        worktreeCleared: true
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toBe(
      'alpha(worktree已清除)'
    )
    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).not.toContain(
      'todo/static-worktree-branch'
    )
  })

  it('refreshes the TODO project sidebar branch when its worktree terminal command ends', async () => {
    let branch = 'feature/login'
    appApiMock.GetTodoProjectGitStatus.mockImplementation(() => Promise.resolve(gitStatus({ branch })))
    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toContain('alpha(feature/login)')

    branch = 'feature/payments'
    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-end' })
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-project-name-todo-project-a"]').text()).toContain('alpha(feature/payments)')
  })

  it('loads sidebar branches for non-active ready TODO project rows', async () => {
    appApiMock.GetTodoProjectGitStatus.mockImplementation((todoProjectId) =>
      Promise.resolve(
        gitStatus({
          branch: todoProjectId === 'todo-project-b' ? 'feature/beta' : 'feature/alpha'
        })
      )
    )
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [
          { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
          { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
        ],
        todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Upgrade deps', status: 'active' })],
        todoProjects: [
          todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a', worktreePath: '/work/tasks/a/alpha' }),
          todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b', worktreePath: '/work/tasks/b/beta' })
        ],
        terminals: [
          terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
          terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
        ]
      })
    )

    const wrapper = await mountReadyApp()
    if (!wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()) {
      await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
      await flushPromises()
    }

    expect(wrapper.find('[data-testid="todo-project-name-todo-project-b"]').text()).toContain('beta(feature/beta)')
  })

  it('shows detailed git status chips when counts are present', async () => {
    appApiMock.GetTodoProjectGitStatus.mockResolvedValue(
      gitStatus({
        branch: 'feature/status-bar',
        changedCount: 6,
        stagedCount: 1,
        unstagedCount: 2,
        untrackedCount: 3,
        ahead: 4,
        behind: 5
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('feature/status-bar')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('6 changed')
    expect(wrapper.find('[data-testid="status-chip-staged"]').text()).toContain('1 staged')
    expect(wrapper.find('[data-testid="status-chip-unstaged"]').text()).toContain('2 unstaged')
    expect(wrapper.find('[data-testid="status-chip-untracked"]').text()).toContain('3 untracked')
    expect(wrapper.find('[data-testid="status-chip-ahead"]').text()).toContain('4 ahead')
    expect(wrapper.find('[data-testid="status-chip-behind"]').text()).toContain('5 behind')
  })

  it('shows a stable empty git status when no project is selected', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [],
        todoProjects: [],
        activeProjectId: '',
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').findAll('.status-chip')).toHaveLength(0)
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('No project')
  })

  it('shows when the active project is not a git repository', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Not a git repository')
    expect(wrapper.find('[data-testid="initialize-git-repository"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="initialize-git-repository"]').text()).toContain('Commit')
  })

  it('shows when git is not installed', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(
      gitStatus({ isRepo: false, branch: '', changedCount: 0, gitUnavailable: true })
    )
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('未安装 Git')
  })

  it('hides git initialization when git is not installed', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(
      gitStatus({ isRepo: false, branch: '', changedCount: 0, gitUnavailable: true })
    )
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="initialize-git-repository"]').exists()).toBe(false)
  })

  it('initializes a git repository from the status bar and refreshes status', async () => {
    appApiMock.GetProjectGitStatus
      .mockResolvedValueOnce(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))
      .mockResolvedValueOnce(gitStatus({ branch: 'main', changedCount: 0 }))
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))

    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="initialize-git-repository"]').trigger('click')
    await flushPromises()

    expect(InitializeProjectGitRepository).toHaveBeenCalledWith('project-a')
    expect(GetProjectGitStatus).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="initialize-git-repository"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('main')
  })

  it('disables the git initialization action while initialization is pending', async () => {
    let resolveInit
    appApiMock.GetProjectGitStatus
      .mockResolvedValueOnce(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))
      .mockResolvedValueOnce(gitStatus({ branch: 'main', changedCount: 0 }))
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))
    appApiMock.InitializeProjectGitRepository.mockReturnValue(
      new Promise((resolve) => {
        resolveInit = resolve
      })
    )

    const wrapper = await mountReadyApp()
    await wrapper.find('[data-testid="initialize-git-repository"]').trigger('click')
    await nextTick()

    const button = wrapper.find('[data-testid="initialize-git-repository"]')
    expect(button.attributes('disabled')).toBeDefined()
    expect(button.text()).toContain('Initializing')

    resolveInit()
    await flushPromises()
  })

  it('keeps non-repository status visible when git initialization fails', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))
    appApiMock.ListProjects.mockResolvedValue(projectState({ activeTodoProjectId: '' }))
    appApiMock.InitializeProjectGitRepository.mockRejectedValue(new Error('git init failed'))

    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="initialize-git-repository"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Not a git repository')
    expect(wrapper.find('[data-testid="initialize-git-repository"]').exists()).toBe(true)
    expect(wrapper.find('.status-error').text()).toContain('git init failed')
  })

  it('shows when the active project path is unavailable without querying git', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        projects: [{ id: 'project-a', name: 'alpha', path: '/missing/alpha', available: false }],
        todoProjects: [],
        activeTodoProjectId: '',
        terminals: [],
        activeTerminalId: ''
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Project path unavailable')
  })

  it('shows unavailable when the active TODO project worktree is not ready without querying source project git', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      projectState({
        todoProjects: [
          todoProject({
            id: 'todo-project-a',
            projectId: 'project-a',
            worktreeStatus: 'pending',
            worktreePath: ''
          })
        ]
      })
    )

    const wrapper = await mountReadyApp()

    expect(GetTodoProjectGitStatus).not.toHaveBeenCalled()
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Project path unavailable')
  })

  it('shows when git status cannot be loaded', async () => {
    appApiMock.GetTodoProjectGitStatus.mockRejectedValue(new Error('git status failed'))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('Git status unavailable')
  })

  it('refreshes git status when the active TODO project changes', async () => {
    const twoProjectState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ],
      todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a', name: 'alpha', path: '/work/alpha', available: true }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b', name: 'beta', path: '/work/beta', available: true })
      ],
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(twoProjectState)
    appApiMock.SelectTodoProject.mockResolvedValue(
      projectState({
        ...twoProjectState,
        activeProjectId: 'project-b',
        activeTodoId: 'todo-b',
        activeTodoProjectId: 'todo-project-b',
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.GetTodoProjectGitStatus.mockImplementation(async (todoProjectId) => {
      if (todoProjectId === 'todo-project-b') {
        return gitStatus({ projectId: 'project-b', branch: 'feature/git-status', changedCount: 2 })
      }
      return gitStatus({ projectId: 'project-a', branch: 'main' })
    })
    const wrapper = await mountReadyApp()

    if (!wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()) {
      await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
      await flushPromises()
    }
    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-b')
    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('feature/git-status')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('2 changed')
  })

  it('keeps same-source TODO worktree git status requests independent', async () => {
    const sharedProjectState = projectState({
      projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
      todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Upgrade deps' })],
      todoProjects: [
        todoProject({
          id: 'todo-project-a',
          todoId: 'todo-a',
          projectId: 'project-a',
          name: 'alpha',
          path: '/work/alpha',
          worktreePath: '/work/customer-a/tasks/aaa/alpha'
        }),
        todoProject({
          id: 'todo-project-b',
          todoId: 'todo-b',
          projectId: 'project-a',
          name: 'alpha',
          path: '/work/alpha',
          worktreePath: '/work/customer-a/tasks/bbb/alpha'
        })
      ],
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-a' })
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(sharedProjectState)
    appApiMock.SelectTodoProject.mockResolvedValue(
      projectState({
        ...sharedProjectState,
        activeProjectId: 'project-a',
        activeTodoId: 'todo-b',
        activeTodoProjectId: 'todo-project-b',
        activeTerminalId: 'terminal-b'
      })
    )
    const resolvers = {}
    appApiMock.GetTodoProjectGitStatus.mockImplementation(
      (todoProjectId) =>
        new Promise((resolve) => {
          resolvers[todoProjectId] = resolve
        })
    )
    const wrapper = await mountReadyApp()
    if (!wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()) {
      await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
      await flushPromises()
    }

    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await nextTick()
    resolvers['todo-project-b'](gitStatus({ projectId: 'project-a', branch: 'todo/upgrade', changedCount: 5 }))
    await flushPromises()
    resolvers['todo-project-a'](gitStatus({ projectId: 'project-a', branch: 'todo/login', changedCount: 1 }))
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-b')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('todo/upgrade')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).toContain('5 changed')
    expect(wrapper.find('[data-testid="project-git-status"]').text()).not.toContain('todo/login')
  })

  it('refreshes git status when a terminal command ends', async () => {
    const wrapper = await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()

    xtermMock.sessions.get('terminal-a').onCommandState({ type: 'command-end' })
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
  })

  it('refreshes git status when the window receives focus', async () => {
    await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()

    window.dispatchEvent(new Event('focus'))
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
  })

  it('deduplicates focus git refresh while the active project request is pending', async () => {
    const resolvers = []
    await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()
    GetTodoProjectGitStatus.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvers.push(resolve)
        })
    )

    window.dispatchEvent(new Event('focus'))
    await nextTick()
    window.dispatchEvent(new Event('focus'))
    await nextTick()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledTimes(1)
    resolvers.forEach((resolve) => resolve(gitStatus()))
    await flushPromises()
  })

  it('debounces repeated focus git refreshes in the same interval', async () => {
    const nowSpy = vi.spyOn(Date, 'now').mockReturnValue(1000)
    try {
      await mountReadyApp()
      GetTodoProjectGitStatus.mockClear()

      window.dispatchEvent(new Event('focus'))
      await flushPromises()
      window.dispatchEvent(new Event('focus'))
      await flushPromises()

      expect(GetTodoProjectGitStatus).toHaveBeenCalledTimes(1)
    } finally {
      nowSpy.mockRestore()
    }
  })

  it('refreshes git status when the active TODO branch is expanded', async () => {
    const wrapper = await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()

    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')
    await wrapper.find('[data-testid="toggle-todo-todo-a"]').trigger('click')
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
  })

  it('toggles a TODO branch by double clicking the TODO header row', async () => {
    const wrapper = await mountReadyApp()

    expect(wrapper.find('[data-testid="todo-project-list-todo-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('dblclick')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-list-todo-a"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-todo-a"]').trigger('dblclick')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-list-todo-a"]').exists()).toBe(true)
  })

  it('does not toggle a TODO branch when double clicking a header action button', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-menu-button-todo-a"]').trigger('dblclick')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-project-list-todo-a"]').exists()).toBe(true)
  })

  it('does not refresh active git status when an unrelated TODO branch expands', async () => {
    const twoTodoState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ],
      todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b' })
      ],
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(twoTodoState)
    const wrapper = await mountReadyApp()
    GetProjectGitStatus.mockClear()
    GetTodoProjectGitStatus.mockClear()

    await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
    await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
    await flushPromises()

    expect(GetProjectGitStatus).not.toHaveBeenCalled()
    expect(GetTodoProjectGitStatus).toHaveBeenCalledTimes(1)
    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-b')
    expect(GetTodoProjectGitStatus).not.toHaveBeenCalledWith('todo-project-a')
  })

  it('refreshes git status when selecting a TODO project changes the active project', async () => {
    const twoProjectState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ],
      todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a' }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b' })
      ],
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
      ]
    })
    appApiMock.ListProjects.mockResolvedValue(twoProjectState)
    appApiMock.SelectTodoProject.mockResolvedValue(
      projectState({
        ...twoProjectState,
        activeProjectId: 'project-b',
        activeTodoId: 'todo-b',
        activeTodoProjectId: 'todo-project-b',
        activeTerminalId: 'terminal-b'
      })
    )
    const wrapper = await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()

    if (!wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()) {
      await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
      await flushPromises()
    }
    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-b')
  })

  it('refreshes git status when selecting the current TODO project', async () => {
    const wrapper = await mountReadyApp()
    GetTodoProjectGitStatus.mockClear()

    await wrapper.find('[data-testid="todo-project-todo-project-a"]').trigger('click')
    await flushPromises()

    expect(GetTodoProjectGitStatus).toHaveBeenCalledWith('todo-project-a')
  })

  it('ignores stale git status responses from a previous active project', async () => {
    const twoProjectState = projectState({
      projects: [
        { id: 'project-a', name: 'alpha', path: '/work/alpha', available: true },
        { id: 'project-b', name: 'beta', path: '/work/beta', available: true }
      ],
      todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })],
      todoProjects: [
        todoProject({ id: 'todo-project-a', todoId: 'todo-a', projectId: 'project-a', name: 'alpha', path: '/work/alpha', available: true }),
        todoProject({ id: 'todo-project-b', todoId: 'todo-b', projectId: 'project-b', name: 'beta', path: '/work/beta', available: true })
      ],
      terminals: [
        terminal({ id: 'terminal-a', todoId: 'todo-a', todoProjectId: 'todo-project-a', projectId: 'project-a' }),
        terminal({ id: 'terminal-b', todoId: 'todo-b', todoProjectId: 'todo-project-b', projectId: 'project-b' })
      ]
    })
    let resolveProjectA
    let resolveProjectB
    appApiMock.ListProjects.mockResolvedValue(twoProjectState)
    appApiMock.SelectTodoProject.mockResolvedValue(
      projectState({
        ...twoProjectState,
        activeProjectId: 'project-b',
        activeTodoId: 'todo-b',
        activeTodoProjectId: 'todo-project-b',
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.GetTodoProjectGitStatus.mockImplementation((todoProjectId) => {
      return new Promise((resolve) => {
        if (todoProjectId === 'todo-project-a') {
          resolveProjectA = resolve
        }
        if (todoProjectId === 'todo-project-b') {
          resolveProjectB = resolve
        }
      })
    })
    const wrapper = await mountReadyApp()

    if (!wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()) {
      await wrapper.find('[data-testid="toggle-todo-todo-b"]').trigger('click')
      await flushPromises()
    }
    await wrapper.find('[data-testid="todo-project-todo-project-b"]').trigger('click')
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

  it('applies the loaded dark appearance theme to the app shell without changing terminal sessions', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(settingsState({ theme: 'dark' }))

    const wrapper = await mountReadyApp()

    expect(wrapper.find('.app-shell').attributes('data-theme')).toBe('dark')
    expect(xtermMock.sessions.has('terminal-a')).toBe(true)
  })

  it('saves a dark appearance theme from settings without updating terminal colors or restarting terminals', async () => {
    appApiMock.SaveTerminalTheme.mockResolvedValue(settingsState({ theme: 'dark' }))
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="appearance-theme-dark"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalTheme).toHaveBeenCalledWith('dark')
    expect(wrapper.find('.app-shell').attributes('data-theme')).toBe('dark')
    expect(StartShell).not.toHaveBeenCalled()
    expect(xtermMock.sessions.has('terminal-a')).toBe(true)
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(false)
  })

  it('does not apply a changed appearance theme when settings are cancelled', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="appearance-theme-dark"]').setValue(true)
    await wrapper.findAll('.settings-actions button').find((button) => button.text() === 'Cancel').trigger('click')
    await flushPromises()

    expect(SaveTerminalTheme).not.toHaveBeenCalled()
    expect(wrapper.find('.app-shell').attributes('data-theme')).toBe('light')
    expect(xtermMock.sessions.has('terminal-a')).toBe(true)
  })

  it('shows appearance theme save errors without closing settings', async () => {
    appApiMock.SaveTerminalTheme.mockRejectedValue(new Error('unsupported appearance theme'))
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="appearance-theme-dark"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="terminal-settings-error"]').text()).toContain('unsupported appearance theme')
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(true)
    expect(wrapper.find('.app-shell').attributes('data-theme')).toBe('light')
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

  it('renders terminal launch profiles in settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'Codex GPT-5', command: 'codex --model gpt-5', enabled: true, background: true },
          { name: 'Claude Plan', command: 'claude', enabled: false, background: false }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="terminal-settings-built-in-launch-profile"]').text()).toContain('Terminal')
    expect(wrapper.find('[data-testid="terminal-launch-profile-name-0"]').element.value).toBe('Codex GPT-5')
    expect(wrapper.find('[data-testid="terminal-launch-profile-command-0"]').element.value).toBe('codex --model gpt-5')
    expect(wrapper.find('[data-testid="terminal-launch-profile-enabled-0"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="terminal-launch-profile-background-0"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="terminal-launch-profile-name-1"]').element.value).toBe('Claude Plan')
    expect(wrapper.find('[data-testid="terminal-launch-profile-enabled-1"]').element.checked).toBe(false)
    expect(wrapper.find('[data-testid="terminal-launch-profile-background-1"]').element.checked).toBe(false)
  })

  it('uses default terminal launch profile commands when settings omit launch profiles', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(settingsState({ launchProfiles: undefined }))
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="terminal-launch-profile-name-0"]').element.value).toBe('codex')
    expect(wrapper.find('[data-testid="terminal-launch-profile-command-0"]').element.value).toBe(defaultCodexLaunchCommand)
    expect(wrapper.find('[data-testid="terminal-launch-profile-enabled-0"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="terminal-launch-profile-background-0"]').element.checked).toBe(false)
    expect(wrapper.find('[data-testid="terminal-launch-profile-name-1"]').element.value).toBe('claude')
    expect(wrapper.find('[data-testid="terminal-launch-profile-command-1"]').element.value).toBe(defaultClaudeLaunchCommand)
    expect(wrapper.find('[data-testid="terminal-launch-profile-enabled-1"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="terminal-launch-profile-background-1"]').element.checked).toBe(false)
  })

  it('treats launch profiles without enabled state as enabled in settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Legacy Codex', command: 'codex --model gpt-5' }] })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="terminal-launch-profile-enabled-0"]').element.checked).toBe(true)
    expect(wrapper.find('[data-testid="terminal-launch-profile-background-0"]').element.checked).toBe(false)
  })

  it('re-detects a terminal shell from settings', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-settings-redetect"]').trigger('click')
    await flushPromises()

    expect(DetectTerminalShell).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-settings-detected"]').text()).toContain('/usr/bin/bash')
  })

  it('saves edited terminal launch profiles from settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'codex', command: 'codex', enabled: true, background: false }] })
    )
    appApiMock.SaveTerminalLaunchProfiles.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Codex GPT-5', command: 'codex --model gpt-5', enabled: false, background: true }] })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-launch-profile-name-0"]').setValue(' Codex GPT-5 ')
    await wrapper.find('[data-testid="terminal-launch-profile-command-0"]').setValue(' codex --model gpt-5 ')
    await wrapper.find('[data-testid="terminal-launch-profile-enabled-0"]').setValue(false)
    await wrapper.find('[data-testid="terminal-launch-profile-background-0"]').setValue(true)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalShell).toHaveBeenCalled()
    expect(SaveTerminalLaunchProfiles).toHaveBeenCalledWith([
      { name: 'Codex GPT-5', command: 'codex --model gpt-5', enabled: false, background: true }
    ])
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(false)
  })

  it('adds removes and reorders terminal launch profiles from settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'codex', command: 'codex', enabled: true },
          { name: 'claude', command: 'claude', enabled: false }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-launch-profile-down-0"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-profile-remove-1"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-profile-add"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-profile-name-1"]').setValue('Gemini')
    await wrapper.find('[data-testid="terminal-launch-profile-command-1"]').setValue('gemini')
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalLaunchProfiles).toHaveBeenCalledWith([
      { name: 'claude', command: 'claude', enabled: false, background: false },
      { name: 'Gemini', command: 'gemini', enabled: true, background: false }
    ])
  })

  it('keeps TODO initialization file management out of terminal settings', async () => {
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="todo-initialization-files-settings"]').exists()).toBe(false)
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTodoInitializationFiles).not.toHaveBeenCalled()
  })

  it('adds edits reorders and saves TODO initialization files from global file management', async () => {
    appApiMock.LoadTodoInitializationFiles.mockResolvedValue([
      initializationFile({ name: 'Agent Rules', fileName: 'AGENTS.md', content: 'rules', defaultSelected: true }),
      initializationFile({ name: 'Prompt', description: '可选提示词', fileName: 'prompt.md', content: 'prompt', defaultSelected: false })
    ])
    const wrapper = await mountReadyApp()

    await openFileManagement(wrapper)

    expect(wrapper.find('[data-testid="todo-initialization-file-management-dialog"]').exists()).toBe(true)
    expect(LoadTodoInitializationFiles).toHaveBeenCalled()
    await wrapper.find('[data-testid="todo-initialization-file-down-0"]').trigger('click')
    await wrapper.find('[data-testid="todo-initialization-file-remove-1"]').trigger('click')
    await wrapper.find('[data-testid="todo-initialization-file-add"]').trigger('click')
    await wrapper.find('[data-testid="todo-initialization-file-name-1"]').setValue('Prompt')
    await wrapper.find('[data-testid="todo-initialization-file-description-1"]').setValue('记录上下文')
    expect(wrapper.find('[data-testid="todo-initialization-file-name-1"]').attributes('placeholder')).toBe('显示名称')
    expect(wrapper.find('[data-testid="todo-initialization-file-description-1"]').attributes('placeholder')).toBe('描述')
    expect(wrapper.find('[data-testid="todo-initialization-file-filename-1"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-initialization-file-content-1"]').exists()).toBe(false)
    await uploadInitializationFile(wrapper, 1, new File(['notes'], 'notes.md', { type: 'text/markdown' }))
    expect(wrapper.find('[data-testid="todo-initialization-file-uploaded-name-1"]').text()).toContain('notes.md')
    await wrapper.find('[data-testid="todo-initialization-file-default-1"]').setValue(true)
    await wrapper.find('[data-testid="todo-initialization-file-management-save"]').trigger('click')
    await flushPromises()

    expect(SaveTodoInitializationFiles).toHaveBeenCalledWith([
      { name: 'Prompt', description: '可选提示词', fileName: 'prompt.md', content: 'prompt', defaultSelected: false },
      { name: 'Prompt', description: '记录上下文', fileName: 'notes.md', content: 'notes', defaultSelected: true }
    ])
    expect(wrapper.find('[data-testid="todo-initialization-file-management-dialog"]').exists()).toBe(false)
  })

  it('shows TODO initialization file save errors without losing edited files', async () => {
    appApiMock.SaveTodoInitializationFiles.mockRejectedValue(new Error('initialization file filename is duplicated'))
    const wrapper = await mountReadyApp()

    await openFileManagement(wrapper)
    await wrapper.find('[data-testid="todo-initialization-file-add"]').trigger('click')
    await wrapper.find('[data-testid="todo-initialization-file-name-0"]').setValue('Agent Rules')
    await uploadInitializationFile(wrapper, 0, new File(['rules'], 'AGENTS.md', { type: 'text/markdown' }))
    await wrapper.find('[data-testid="todo-initialization-file-management-save"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-initialization-file-management-error"]').text()).toContain('initialization file filename is duplicated')
    expect(wrapper.find('[data-testid="todo-initialization-file-management-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-initialization-file-name-0"]').element.value).toBe('Agent Rules')
  })

  it('adds edits reorders and saves TODO lifecycle scripts from global script management', async () => {
    appApiMock.LoadTodoLifecycleScripts.mockResolvedValue([
      lifecycleScriptTemplate({ name: 'Node setup', description: 'Install deps', initScript: 'npm install', defaultSelected: true }),
      lifecycleScriptTemplate({ name: 'Verify', description: 'Run tests', initScript: '', completeScript: 'npm test', defaultSelected: false })
    ])
    const wrapper = await mountReadyApp()

    await openScriptManagement(wrapper)

    expect(wrapper.find('[data-testid="todo-lifecycle-script-management-dialog"]').exists()).toBe(true)
    expect(LoadTodoLifecycleScripts).toHaveBeenCalled()
    await wrapper.find('[data-testid="todo-lifecycle-script-down-0"]').trigger('click')
    await wrapper.find('[data-testid="todo-lifecycle-script-remove-1"]').trigger('click')
    await wrapper.find('[data-testid="todo-lifecycle-script-add"]').trigger('click')
    await wrapper.find('[data-testid="todo-lifecycle-script-name-1"]').setValue('Release')
    await wrapper.find('[data-testid="todo-lifecycle-script-description-1"]').setValue('Prepare release')
    await wrapper.find('[data-testid="todo-lifecycle-script-init-1"]').setValue('pnpm install')
    await wrapper.find('[data-testid="todo-lifecycle-script-complete-1"]').setValue('pnpm test')
    await wrapper.find('[data-testid="todo-lifecycle-script-default-1"]').setValue(true)
    await wrapper.find('[data-testid="todo-lifecycle-script-management-save"]').trigger('click')
    await flushPromises()

    expect(SaveTodoLifecycleScripts).toHaveBeenCalledWith([
      { name: 'Verify', description: 'Run tests', initScript: '', completeScript: 'npm test', defaultSelected: false },
      { name: 'Release', description: 'Prepare release', initScript: 'pnpm install', completeScript: 'pnpm test', defaultSelected: true }
    ])
    expect(wrapper.find('[data-testid="todo-lifecycle-script-management-dialog"]').exists()).toBe(false)
  })

  it('shows TODO lifecycle script save errors without losing edited scripts', async () => {
    appApiMock.SaveTodoLifecycleScripts.mockRejectedValue(new Error('lifecycle script name is required'))
    const wrapper = await mountReadyApp()

    await openScriptManagement(wrapper)
    await wrapper.find('[data-testid="todo-lifecycle-script-add"]').trigger('click')
    await wrapper.find('[data-testid="todo-lifecycle-script-init-0"]').setValue('npm install')
    await wrapper.find('[data-testid="todo-lifecycle-script-management-save"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="todo-lifecycle-script-management-error"]').text()).toContain('lifecycle script name is required')
    expect(wrapper.find('[data-testid="todo-lifecycle-script-management-dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-init-0"]').element.value).toBe('npm install')
  })

  it('shows lifecycle scripts in one dropdown and submits the selected snapshot when creating a TODO', async () => {
    appApiMock.LoadTodoLifecycleScripts.mockResolvedValue([
      lifecycleScriptTemplate({ name: 'Node setup', description: 'Install deps', initScript: 'npm install', completeScript: 'npm test' }),
      lifecycleScriptTemplate({ name: 'API checks', description: 'Backend verification', initScript: 'go mod download', completeScript: 'go test ./...', defaultSelected: true })
    ])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')

    expect(wrapper.find('[data-testid="todo-lifecycle-script-filter"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-description"]').exists()).toBe(false)
    expect(wrapper.find('select[data-testid="todo-lifecycle-script-select"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-menu"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-lifecycle-script-select"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="todo-lifecycle-script-menu"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-option-0"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-option-0"]').text()).toBe('Node setup - Install deps')
    expect(wrapper.find('[data-testid="todo-lifecycle-script-option-1"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-lifecycle-script-option-1"]').text()).toBe('API checks - Backend verification')

    await wrapper.find('[data-testid="todo-lifecycle-script-option-0"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="todo-lifecycle-script-select"]').text()).toContain('Node setup - Install deps')
    expect(wrapper.find('[data-testid="todo-lifecycle-script-menu"]').exists()).toBe(false)

    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: [],
      lifecycleScript: {
        name: 'Node setup',
        description: 'Install deps',
        initScript: 'npm install',
        completeScript: 'npm test'
      }
    })
  })

  it('supports creating a TODO without selecting a lifecycle script', async () => {
    appApiMock.LoadTodoLifecycleScripts.mockResolvedValue([
      lifecycleScriptTemplate({ name: 'Node setup', description: 'Install deps', initScript: 'npm install', defaultSelected: true })
    ])
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="new-todo"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="todo-name-input"]').setValue('Write tests')
    await wrapper.find('[data-testid="todo-lifecycle-script-select"]').trigger('click')
    await nextTick()
    await wrapper.find('[data-testid="todo-lifecycle-script-option-none"]').trigger('click')
    await wrapper.find('[data-testid="todo-create-submit"]').trigger('click')
    await flushPromises()

    expect(CreateTodo).toHaveBeenCalledWith({
      title: 'Write tests',
      description: '',
      priority: 'medium',
      projects: []
    })
  })

  it('merges lifecycle script status events and hides cleared successful states', async () => {
    appApiMock.ListProjects.mockResolvedValue(inProgressProjectState({ lifecycleScriptStatuses: [] }))
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')
    runtimeMock.handlers['todo-lifecycle-script-status'](
      lifecycleScriptStatus({ phase: 'init', status: 'running', scriptName: 'Node setup' })
    )
    await nextTick()

    expect(wrapper.find('[data-testid="todo-lifecycle-script-status-todo-a-init"]').text()).toContain('初始化脚本执行中')

    runtimeMock.handlers['todo-lifecycle-script-status'](
      lifecycleScriptStatus({ phase: 'init', status: '', scriptName: 'Node setup' })
    )
    await nextTick()

    expect(wrapper.find('[data-testid="todo-lifecycle-script-status-todo-a-init"]').exists()).toBe(false)
  })

  it('shows failed lifecycle scripts with retry actions and replaces them with running state', async () => {
    appApiMock.ListProjects.mockResolvedValue(
      inProgressProjectState({
        lifecycleScriptStatuses: [
          lifecycleScriptStatus({ phase: 'complete', status: 'failed', outputTail: 'lint failed', exitCode: 2 })
        ]
      })
    )
    appApiMock.RetryTodoLifecycleScript.mockResolvedValue(
      inProgressProjectState({
        lifecycleScriptStatuses: [
          lifecycleScriptStatus({ phase: 'complete', status: 'running', outputTail: '', exitCode: 0 })
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="todo-view-in-progress"]').trigger('click')

    const failedStatus = wrapper.find('[data-testid="todo-lifecycle-script-status-todo-a-complete"]')
    expect(failedStatus.text()).toContain('完成脚本失败')
    expect(failedStatus.text()).toContain('lint failed')

    await wrapper.find('[data-testid="retry-todo-lifecycle-script-todo-a-complete"]').trigger('click')
    await flushPromises()

    expect(RetryTodoLifecycleScript).toHaveBeenCalledWith('todo-a', 'complete')
    expect(wrapper.find('[data-testid="todo-lifecycle-script-status-todo-a-complete"]').text()).toContain('完成脚本执行中')
    expect(wrapper.find('[data-testid="retry-todo-lifecycle-script-todo-a-complete"]').exists()).toBe(false)
  })

  it('shows launch profile validation errors without closing settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'codex', command: 'codex' }] })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-launch-profile-command-0"]').setValue(' ')
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalShell).not.toHaveBeenCalled()
    expect(SaveTerminalLaunchProfiles).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="terminal-settings-error"]').text()).toContain('Launch profile command is required')
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(true)
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
  mountedWrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

async function openSettings(wrapper) {
  await wrapper.find('[data-testid="settings-toggle"]').trigger('click')
  await flushPromises()
}

async function openFileManagement(wrapper) {
  await wrapper.find('[data-testid="global-management-toggle"]').trigger('click')
  await nextTick()
  await wrapper.find('[data-testid="global-file-management"]').trigger('click')
  await flushPromises()
}

async function openScriptManagement(wrapper) {
  await wrapper.find('[data-testid="global-management-toggle"]').trigger('click')
  await nextTick()
  await wrapper.find('[data-testid="global-script-management"]').trigger('click')
  await flushPromises()
}

async function uploadInitializationFile(wrapper, index, file) {
  const input = wrapper.find(`[data-testid="todo-initialization-file-upload-${index}"]`)
  Object.defineProperty(input.element, 'files', {
    configurable: true,
    value: [file]
  })
  await input.trigger('change')
  await flushPromises()
}

async function openTerminalMenu(wrapper) {
  await wrapper.find('[data-testid="terminal-pane-terminal-a"]').trigger('contextmenu', {
    clientX: 48,
    clientY: 64
  })
  await nextTick()
}

async function openTodoContextMenu(wrapper, todoId) {
  await wrapper.find(`[data-testid="todo-${todoId}"]`).trigger('contextmenu', {
    clientX: 48,
    clientY: 64
  })
  await nextTick()
}

async function selectTodoMenuAction(wrapper, action, todoId) {
  await openTodoContextMenu(wrapper, todoId)
  const actionMap = {
    edit: `todo-menu-edit-${todoId}`,
    'add-project': `todo-menu-add-project-${todoId}`,
    'copy-description': `todo-menu-copy-description-${todoId}`,
    delete: `todo-menu-delete-${todoId}`
  }
  const testId = actionMap[action]
  if (!testId) {
    throw new Error(`Unknown todo menu action: ${action}`)
  }
  await wrapper.find(`[data-testid="${testId}"]`).trigger('click')
  await nextTick()
}

async function flushPromises() {
  await nextTick()
  await Promise.resolve()
  await nextTick()
  await Promise.resolve()
  await nextTick()
}

function deferred() {
  let resolve
  let reject
  const promise = new Promise((promiseResolve, promiseReject) => {
    resolve = promiseResolve
    reject = promiseReject
  })
  return { promise, resolve, reject }
}

function projectState(overrides = {}) {
  return {
    currentWorkspace: workspace(),
    recentWorkspaces: [workspace()],
    projects: [{ id: 'project-a', name: 'alpha', path: '/work/alpha', available: true }],
    todos: [todo()],
    todoProjects: [todoProject()],
    projectBranchPreferences: {},
    lifecycleScriptStatuses: [],
    activeProjectId: 'project-a',
    activeTodoId: 'todo-a',
    activeTodoProjectId: 'todo-project-a',
    terminals: [terminal({ id: 'terminal-a' })],
    activeTerminalId: 'terminal-a',
    ...overrides
  }
}

function projectImportResult(state, overrides = {}) {
  return {
    state,
    requiresGitInitialization: false,
    path: '',
    ...overrides
  }
}

function noWorkspaceState(overrides = {}) {
  return projectState({
    currentWorkspace: null,
    projects: [],
    todos: [],
    todoProjects: [],
    activeProjectId: '',
    activeTodoId: '',
    activeTodoProjectId: '',
    terminals: [],
    activeTerminalId: '',
    importSummary: null,
    ...overrides
  })
}

function workspace(overrides = {}) {
  return {
    name: 'Customer A',
    path: '/work/customer-a',
    dataPath: '/work/customer-a/.data',
    available: true,
    lastOpenedAt: '2026-06-10T09:00:00Z',
    ...overrides
  }
}

function workspaceState(overrides = {}) {
  return {
    version: 1,
    currentWorkspace: workspace(),
    recentWorkspaces: [workspace()],
    ...overrides
  }
}

function inProgressProjectState(overrides = {}) {
  return projectState({
    todos: [todo({ status: 'in-progress' })],
    ...overrides
  })
}

function todo(overrides = {}) {
  return {
    id: 'todo-a',
    title: 'Fix login',
    status: 'active',
    createdAt: '2026-06-10T09:00:00Z',
    ...overrides
  }
}

function completedTodo(overrides = {}) {
  return {
    ...todo({
      status: 'completed',
      completedAt: '2026-06-10T10:00:00Z',
      projectSnapshots: [
        {
          projectId: 'project-a',
          name: 'alpha',
          path: '/work/alpha',
          worktreeBranch: 'feature/alpha',
          baseBranch: 'main'
        }
      ]
    }),
    ...overrides
  }
}

function todoProject(overrides = {}) {
  return {
    id: 'todo-project-a',
    todoId: 'todo-a',
    projectId: 'project-a',
    worktreeStatus: 'ready',
    worktreePath: '/work/customer-a/tasks/abc123/alpha',
    ...overrides
  }
}

function settingsState(overrides = {}) {
  return {
    version: 1,
    selected: shellSetting(),
    theme: 'light',
    launchProfiles: [
      { name: 'codex', command: defaultCodexLaunchCommand },
      { name: 'claude', command: defaultClaudeLaunchCommand }
    ],
    todoInitializationFiles: [],
    todoLifecycleScripts: [],
    ...overrides
  }
}

function initializationFile(overrides = {}) {
  return {
    name: 'Agent Rules',
    description: '任务执行约束',
    fileName: 'AGENTS.md',
    content: '请先阅读任务说明',
    defaultSelected: true,
    ...overrides
  }
}

function lifecycleScriptTemplate(overrides = {}) {
  return {
    name: 'Node setup',
    description: 'Install dependencies',
    initScript: 'npm install',
    completeScript: 'npm test',
    defaultSelected: false,
    ...overrides
  }
}

function lifecycleScriptStatus(overrides = {}) {
  return {
    todoId: 'todo-a',
    phase: 'init',
    status: 'running',
    scriptName: 'Node setup',
    startedAt: '2026-06-10T09:00:00Z',
    finishedAt: '',
    exitCode: 0,
    outputTail: '',
    message: '',
    ...overrides
  }
}

function todoProjectUIStateFile(overrides = {}) {
  return {
    version: 1,
    sidebarWidth: 280,
    todoProjects: {},
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
    todoId: 'todo-a',
    todoProjectId: 'todo-project-a',
    shellName: 'zsh',
    currentCommand: '',
    state: 'running',
    ...overrides
  }
}

function workspaceTerminal(overrides = {}) {
  return terminal({
    id: 'global-terminal',
    projectId: '',
    todoId: '',
    todoProjectId: '',
    workspaceTerminal: true,
    ...overrides
  })
}

function taskTerminal(overrides = {}) {
  return terminal({
    id: 'task-terminal-a',
    projectId: '',
    todoProjectId: '',
    todoId: 'todo-a',
    taskTerminal: true,
    ...overrides
  })
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
