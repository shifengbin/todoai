import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'
import {
  AddProjectToTodo,
  AddProjectsToTodo,
  CompleteTodo,
  CreateTodo,
  CreateTodoTerminal,
  DeleteProject,
  DeleteTerminal,
  DeleteTodo,
  DetectTerminalShell,
  GetProjectGitStatus,
  ImportProjectsFromParentDirectoryDialog,
  InitializeProjectGitRepository,
  LoadTerminalSettings,
  RemoveTodoProject,
  SaveTerminalLaunchProfiles,
  SaveTerminalShell,
  SaveTerminalTheme,
  SelectTerminal,
  SendTerminalInput,
  StartShell,
  UpdateTodo
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText } from '../wailsjs/runtime/runtime'

const appApiMock = vi.hoisted(() => ({
  AddProjectToTodo: vi.fn(),
  AddProjectsToTodo: vi.fn(),
  CompleteTodo: vi.fn(),
  CreateTodo: vi.fn(),
  CreateTodoTerminal: vi.fn(),
  CreateProjectFromDialog: vi.fn(),
  DeleteProject: vi.fn(),
  DeleteTerminal: vi.fn(),
  DeleteTodo: vi.fn(),
  DetectTerminalShell: vi.fn(),
  GetProjectGitStatus: vi.fn(),
  ImportProjectsFromParentDirectoryDialog: vi.fn(),
  InitializeProjectGitRepository: vi.fn(),
  ListProjects: vi.fn(),
  LoadTerminalSettings: vi.fn(),
  RemoveTodoProject: vi.fn(),
  ResizeTerminal: vi.fn(),
  SelectProject: vi.fn(),
  SelectTerminal: vi.fn(),
  SelectTodoProject: vi.fn(),
  SaveTerminalLaunchProfiles: vi.fn(),
  SaveTerminalShell: vi.fn(),
  SaveTerminalTheme: vi.fn(),
  SendTerminalInput: vi.fn(),
  StartShell: vi.fn(),
  UpdateTodo: vi.fn()
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
        dispose: vi.fn(),
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
    appApiMock.SaveTerminalLaunchProfiles.mockResolvedValue(settingsState())
    appApiMock.SaveTerminalTheme.mockResolvedValue(settingsState())
    appApiMock.SelectProject.mockResolvedValue(projectState())
    appApiMock.SelectTodoProject.mockResolvedValue(projectState())
    appApiMock.SelectTerminal.mockResolvedValue(projectState())
    appApiMock.CreateTodoTerminal.mockResolvedValue(
      projectState({
        terminals: [
          terminal({ id: 'terminal-a' }),
          terminal({ id: 'terminal-b', shellName: 'bash', state: 'running' })
        ],
        activeTerminalId: 'terminal-b'
      })
    )
    appApiMock.CreateTodo.mockResolvedValue(projectState({ todos: [todo({ id: 'todo-a' }), todo({ id: 'todo-b', title: 'Write tests' })] }))
    appApiMock.AddProjectToTodo.mockResolvedValue(projectState())
    appApiMock.AddProjectsToTodo.mockResolvedValue(projectState())
    appApiMock.CompleteTodo.mockResolvedValue(projectState({ todos: [archivedTodo()], todoProjects: [], terminals: [], activeTodoId: '', activeTodoProjectId: '', activeTerminalId: '' }))
    appApiMock.DeleteTodo.mockResolvedValue(projectState({ todos: [archivedTodo({ archivedReason: 'deleted' })], todoProjects: [], terminals: [], activeTodoId: '', activeTodoProjectId: '', activeTerminalId: '' }))
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
    appApiMock.DeleteTerminal.mockResolvedValue(projectState({ terminals: [], activeTerminalId: '' }))
    appApiMock.RemoveTodoProject.mockResolvedValue(projectState({ todoProjects: [], terminals: [], activeTodoProjectId: '', activeTerminalId: '' }))
    appApiMock.UpdateTodo.mockResolvedValue(projectState())
    appApiMock.GetProjectGitStatus.mockResolvedValue(gitStatus())
    appApiMock.InitializeProjectGitRepository.mockResolvedValue()
    appApiMock.StartShell.mockResolvedValue({ projectId: 'project-a', terminalId: 'terminal-a', state: 'running' })
    appApiMock.SendTerminalInput.mockResolvedValue()
    runtimeMock.ClipboardGetText.mockResolvedValue('')
    runtimeMock.ClipboardSetText.mockResolvedValue(true)
    vi.stubGlobal('confirm', vi.fn(() => true))
    vi.stubGlobal('prompt', vi.fn(() => ''))
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

  it('shows configured terminal launch profiles from loaded settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'codex', command: 'codex' },
          { name: 'claude', command: 'claude' }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await nextTick()

    const menu = wrapper.find('[data-testid="terminal-launch-menu-todo-project-a"]')
    expect(LoadTerminalSettings).toHaveBeenCalled()
    expect(menu.exists()).toBe(true)
    expect(menu.text()).toContain('Terminal')
    expect(menu.text()).toContain('codex')
    expect(menu.text()).toContain('claude')
  })

  it('creates an additional terminal under the active project', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-0"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(xtermMock.sessions.has('terminal-b')).toBe(true)
    expect(wrapper.find('[data-testid="terminal-terminal-b"]').classes()).toContain('active')
  })

  it('creates a terminal from a custom launch profile and submits its command', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Codex GPT-5', command: 'codex --model gpt-5' }] })
    )
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="add-terminal-todo-project-a"]').trigger('click')
    await wrapper.find('[data-testid="terminal-launch-option-todo-project-a-1"]').trigger('click')
    await flushPromises()

    expect(CreateTodoTerminal).toHaveBeenCalledWith('todo-project-a', 100, 32)
    expect(SendTerminalInput).toHaveBeenCalledWith('terminal-b', 'codex --model gpt-5\n')
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

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('alpha'))
    expect(DeleteProject).toHaveBeenCalledWith('project-a')
    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(false)
  })

  it('does not delete a project when confirmation is cancelled', async () => {
    window.confirm.mockReturnValue(false)
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
    await wrapper.find('[data-testid="delete-project-project-a"]').trigger('click')
    await flushPromises()

    expect(DeleteProject).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="project-project-a"]').exists()).toBe(true)
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
      projectIds: ['project-b', 'project-c']
    })
    expect(wrapper.text()).toContain('Write tests')
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
      projectIds: ['project-c']
    })
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
    appApiMock.AddProjectsToTodo.mockResolvedValue(
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

    await wrapper.find('[data-testid="add-project-to-todo-todo-a"]').trigger('click')
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

    expect(AddProjectsToTodo).toHaveBeenCalledWith('todo-a', ['project-b', 'project-c'])
    expect(wrapper.find('[data-testid="todo-project-todo-project-b"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="todo-project-todo-project-c"]').exists()).toBe(true)
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

    await wrapper.find('[data-testid="add-project-to-todo-todo-a"]').trigger('click')
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

    expect(AddProjectsToTodo).toHaveBeenCalledWith('todo-a', ['project-b'])
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

    await wrapper.find('[data-testid="edit-todo-todo-a"]').trigger('click')
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
      projectIds: ['project-b']
    })
    expect(wrapper.find('[data-testid="todo-detail-dialog"]').exists()).toBe(false)
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

    await wrapper.find('[data-testid="edit-todo-todo-a"]').trigger('click')
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

  it('completes a TODO and shows its archived snapshot', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="complete-todo-todo-a"]').trigger('click')
    await nextTick()
    expect(CompleteTodo).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="complete-todo-popover-todo-a"]').exists()).toBe(true)

    await wrapper.find('[data-testid="confirm-complete-todo-todo-a"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="todo-view-archived"]').trigger('click')

    expect(window.confirm).not.toHaveBeenCalled()
    expect(CompleteTodo).toHaveBeenCalledWith('todo-a')
    expect(wrapper.find('[data-testid="terminal-terminal-a"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="archived-todos"]').text()).toContain('completed')
    expect(wrapper.find('[data-testid="archived-todos"]').text()).toContain('/work/alpha')
  })

  it('does not delete a TODO when the sidebar confirmation is cancelled', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="delete-todo-todo-a"]').trigger('click')
    await nextTick()

    expect(wrapper.find('[data-testid="delete-todo-popover-todo-a"]').exists()).toBe(true)
    await wrapper.find('[data-testid="cancel-delete-todo-todo-a"]').trigger('click')
    await flushPromises()

    expect(window.confirm).not.toHaveBeenCalled()
    expect(DeleteTodo).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="todo-todo-a"]').exists()).toBe(true)
  })

  it('imports projects from a parent directory and shows the summary', async () => {
    const wrapper = await mountReadyApp()

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
    await wrapper.find('[data-testid="import-parent-directory"]').trigger('click')
    await flushPromises()

    expect(ImportProjectsFromParentDirectoryDialog).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="import-summary"]').text()).toContain('2 imported')
    expect(wrapper.find('[data-testid="import-summary"]').text()).toContain('1 skipped')
  })

  it('selects a project from the project tab without creating a terminal', async () => {
    const wrapper = await mountReadyApp()
    CreateTodoTerminal.mockClear()

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
    await wrapper.find('[data-testid="project-project-a"]').trigger('click')
    await flushPromises()

    expect(appApiMock.SelectProject).toHaveBeenCalledWith('project-a')
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
    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('main')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('3 changed')
  })

  it('shows detailed git status chips when counts are present', async () => {
    appApiMock.GetProjectGitStatus.mockResolvedValue(
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

    expect(wrapper.find('[data-testid="status-chip-branch"]').text()).toContain('feature/status-bar')
    expect(wrapper.find('[data-testid="status-chip-changed"]').text()).toContain('6 changed')
    expect(wrapper.find('[data-testid="status-chip-staged"]').text()).toContain('1 staged')
    expect(wrapper.find('[data-testid="status-chip-unstaged"]').text()).toContain('2 unstaged')
    expect(wrapper.find('[data-testid="status-chip-untracked"]').text()).toContain('3 untracked')
    expect(wrapper.find('[data-testid="status-chip-ahead"]').text()).toContain('4 ahead')
    expect(wrapper.find('[data-testid="status-chip-behind"]').text()).toContain('5 behind')
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
    expect(wrapper.find('[data-testid="initialize-git-repository"]').exists()).toBe(true)
  })

  it('initializes a git repository from the status bar and refreshes status', async () => {
    appApiMock.GetProjectGitStatus
      .mockResolvedValueOnce(gitStatus({ isRepo: false, branch: '', changedCount: 0 }))
      .mockResolvedValueOnce(gitStatus({ branch: 'main', changedCount: 0 }))

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

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
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

    await wrapper.find('[data-testid="sidebar-tab-projects"]').trigger('click')
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
          { name: 'Codex GPT-5', command: 'codex --model gpt-5' },
          { name: 'Claude Plan', command: 'claude' }
        ]
      })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)

    expect(wrapper.find('[data-testid="terminal-settings-built-in-launch-profile"]').text()).toContain('Terminal')
    expect(wrapper.find('[data-testid="terminal-launch-profile-name-0"]').element.value).toBe('Codex GPT-5')
    expect(wrapper.find('[data-testid="terminal-launch-profile-command-0"]').element.value).toBe('codex --model gpt-5')
    expect(wrapper.find('[data-testid="terminal-launch-profile-name-1"]').element.value).toBe('Claude Plan')
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
      settingsState({ launchProfiles: [{ name: 'codex', command: 'codex' }] })
    )
    appApiMock.SaveTerminalLaunchProfiles.mockResolvedValue(
      settingsState({ launchProfiles: [{ name: 'Codex GPT-5', command: 'codex --model gpt-5' }] })
    )
    const wrapper = await mountReadyApp()

    await openSettings(wrapper)
    await wrapper.find('[data-testid="terminal-launch-profile-name-0"]').setValue(' Codex GPT-5 ')
    await wrapper.find('[data-testid="terminal-launch-profile-command-0"]').setValue(' codex --model gpt-5 ')
    await wrapper.find('[data-testid="terminal-settings-save"]').trigger('click')
    await flushPromises()

    expect(SaveTerminalShell).toHaveBeenCalled()
    expect(SaveTerminalLaunchProfiles).toHaveBeenCalledWith([
      { name: 'Codex GPT-5', command: 'codex --model gpt-5' }
    ])
    expect(wrapper.find('[data-testid="terminal-settings-dialog"]').exists()).toBe(false)
  })

  it('adds removes and reorders terminal launch profiles from settings', async () => {
    appApiMock.LoadTerminalSettings.mockResolvedValue(
      settingsState({
        launchProfiles: [
          { name: 'codex', command: 'codex' },
          { name: 'claude', command: 'claude' }
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
      { name: 'claude', command: 'claude' },
      { name: 'Gemini', command: 'gemini' }
    ])
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
    todos: [todo()],
    todoProjects: [todoProject()],
    activeProjectId: 'project-a',
    activeTodoId: 'todo-a',
    activeTodoProjectId: 'todo-project-a',
    terminals: [terminal({ id: 'terminal-a' })],
    activeTerminalId: 'terminal-a',
    ...overrides
  }
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

function archivedTodo(overrides = {}) {
  return {
    ...todo({
      status: 'archived',
      archivedReason: 'completed',
      archivedAt: '2026-06-10T10:00:00Z',
      projectSnapshots: [{ projectId: 'project-a', name: 'alpha', path: '/work/alpha' }]
    }),
    ...overrides
  }
}

function todoProject(overrides = {}) {
  return {
    id: 'todo-project-a',
    todoId: 'todo-a',
    projectId: 'project-a',
    ...overrides
  }
}

function settingsState(overrides = {}) {
  return {
    version: 1,
    selected: shellSetting(),
    theme: 'light',
    launchProfiles: [
      { name: 'codex', command: 'codex' },
      { name: 'claude', command: 'claude' }
    ],
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
