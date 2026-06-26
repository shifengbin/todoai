<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ChevronDown, ChevronUp, FileText, FolderInput, FolderPlus, GitBranch, Plus, RotateCcw, Settings, Trash2, X } from '@lucide/vue'
import ProjectSidebar from './components/ProjectSidebar.vue'
import {
  AGENT_CONFIDENCE,
  AGENT_PHASE,
  AGENT_SOURCE,
  applyAgentStatusEvent,
  createAgentStatus
} from './agentStatus'
import { TerminalSessionManager } from './terminalManager'
import { createXtermSession } from './xtermFactory'
import {
  AddProjectSelectionsToTodo,
  ChangeTodoStatus,
  ClaudeStatusHookState,
  CompleteTodo,
  CreateProjectFromDialog,
  CreateTodo,
  CreateTaskTerminal,
  CreateTodoTerminal,
  CreateWorkspaceTerminal,
  DeleteCompletedTodos,
  DeleteProject,
  DeleteProjects,
  DeleteTerminal,
  DeleteTodo,
  DetectTerminalShell,
  EnsureClaudeStatusHook,
  GetCompletedTodoProjectMergeStatuses,
  GetProjectGitStatus,
  GetTodoGitStatus,
  GetTodoProjectGitStatus,
  ImportProjectsFromParentDirectoryDialog,
  InitializeGitRepositoryAndImportProject,
  InitializeProjectGitRepository,
  ListProjectBranches,
  ListProjects,
  LoadTerminalSettings,
  LoadTodoInitializationFiles,
  LoadTodoLifecycleScripts,
  LoadTodoProjectUIState,
  OpenTodoFolder,
  OpenRecentWorkspace,
  RemoveClaudeStatusHook,
  RemoveTodoProject,
  ResizeTerminal,
  RetryTodoLifecycleScript,
  SelectProject,
  SelectTerminal,
  SelectTodoProject,
  SaveTerminalLaunchProfiles,
  SaveTerminalShell,
  SaveTerminalTheme,
  SaveTodoInitializationFiles,
  SaveTodoLifecycleScripts,
  SaveTodoSidebarWidth,
  SaveTodoProjectUIState,
  SendTerminalInput,
  StartShell,
  StartTaskBackgroundCommand,
  StartTodoProjectBackgroundCommand,
  UpdateTodo
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText, EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

const projects = ref([])
const todos = ref([])
const todoProjects = ref([])
const projectBranchPreferences = ref({})
const lifecycleScriptStatuses = ref([])
const terminals = ref([])
const currentWorkspace = ref(null)
const recentWorkspaces = ref([])
const activeProjectId = ref('')
const activeTodoId = ref('')
const activeTodoProjectId = ref('')
const activeTerminalId = ref('')
const importSummary = ref(null)
const shellStatuses = reactive({})
const toastMessage = ref('')
const terminalContainers = new Map()
const titleActivityTimers = new Map()
const autoRestartedTerminals = new Set()
const terminalAckIds = reactive(new Set())
const errorMessage = ref('')
let gitStatusRequestId = 0
let gitStatusInFlightContextKey = ''
let gitStatusInFlightRequestId = 0
const todoProjectGitBranches = reactive({})
const todoProjectGitStatusInFlight = new Map()
let todoProjectGitStatusRequestId = 0
let completedMergeStatusRequestGeneration = 0
let lastFocusGitRefreshContextKey = ''
let lastFocusGitRefreshAt = 0
const focusGitRefreshDedupeMs = 500
const gitOnlyImportToastText = '只能导入 Git 仓库'
const bulkGitImportTooltip = '仅导入一级子目录中的 Git 仓库'
const toastDurationMs = 2000
let toastTimer = null
const gitInitializationPrompt = reactive({
  visible: false,
  path: ''
})
let gitInitializationPromptResolve = null
const terminalMenu = reactive({
  visible: false,
  terminalId: '',
  x: 0,
  y: 0
})
const globalManagementMenu = reactive({
  visible: false
})
const sidebarWidth = ref(280)
const defaultSidebarWidth = 280
const sidebarResize = reactive({
  active: false,
  startX: 0,
  startWidth: defaultSidebarWidth
})
const sidebarMinWidth = 220
const sidebarMaxWidth = 520
const currentTodoView = ref('not-started')
const todoProjectUIStates = ref({})
const todoSidebarWidthState = ref(0)
const completedMergeStatuses = ref({})
const todoProjectUIStateSaveQueues = new Map()
const todoSidebarWidthSaveQueue = {
  saving: false,
  pending: null
}
const defaultTerminalLaunchProfiles = [
  { name: 'codex', command: 'codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox', enabled: true, background: false },
  { name: 'claude', command: 'claude --dangerously-skip-permissions', enabled: true, background: false }
]
const todoPriorities = [
  { value: 'high', label: '高' },
  { value: 'medium', label: '中' },
  { value: 'low', label: '低' }
]
const appearanceThemes = ['light', 'dark']
const currentTheme = ref('light')
const terminalSettings = ref(null)
const detectedTerminalShell = ref(null)
const gitStatus = ref(null)
const gitStatusLoading = ref(false)
const gitStatusError = ref('')
const gitInitLoading = ref(false)
const projectBranchOptions = reactive({})
const projectBranchLoadStates = reactive({})
const projectBranchPickerQueries = reactive({})
const openProjectBranchPickerKey = ref('')
const projectBranchCandidateLimit = 50
const settingsPanel = reactive({
  visible: false,
  loading: false,
  detecting: false,
  saving: false,
  mode: 'detected',
  manualPath: '',
  launchProfiles: [],
  theme: 'light',
  error: '',
  claudeStatus: {
    installed: false,
    stale: false,
    checking: false,
    command: '',
    eventsCovered: 0
  }
})
const initializationFileManagement = reactive({
  visible: false,
  loading: false,
  saving: false,
  files: [],
  error: ''
})
const lifecycleScriptManagement = reactive({
  visible: false,
  loading: false,
  saving: false,
  scripts: [],
  error: ''
})
const todoForm = reactive({
  visible: false,
  title: '',
  description: '',
  priority: 'medium',
  projectSelections: [],
  projectSearch: '',
  initializationFiles: [],
  lifecycleScripts: [],
  selectedLifecycleScriptIndex: '',
  lifecycleScriptMenuOpen: false,
  saving: false
})
const todoDetail = reactive({
  visible: false,
  todoId: '',
  title: '',
  description: '',
  priority: 'medium',
  projectSelections: [],
  projectSnapshots: [],
  projectSearch: '',
  readOnly: false,
  saving: false
})
const projectPicker = reactive({
  visible: false,
  todoId: '',
  query: '',
  projectSelections: [],
  saving: false
})
const projectCandidateClearPrompt = reactive({
  visible: false,
  project: null,
  clearing: false
})
const recentWorkspacePicker = reactive({
  visible: false,
  openingPath: '',
  error: ''
})
const terminalManager = new TerminalSessionManager({
  createSession: createXtermSession,
  sendInput: (terminalId, data) => SendTerminalInput(terminalId, data),
  resizeTerminal: (terminalId, cols, rows) => {
    if (terminalState(terminalId) === 'running') {
      ResizeTerminal(terminalId, cols, rows)
    }
  },
  clipboard: {
    readText: ClipboardGetText,
    writeText: ClipboardSetText
  },
  onCommandState: handleTerminalCommandState,
  onTitleChange: handleTerminalTitleChange,
  onError: showError
})

const activeProject = computed(() => {
  const todoProject = todoProjects.value.find((candidate) => candidate.id === activeTodoProjectId.value)
  return todoProjectDisplayProject(todoProject) || projects.value.find((project) => project.id === activeProjectId.value) || null
})

const activeProjectPath = computed(() => activeProject.value?.path || '')

const hasWorkspace = computed(() => Boolean(currentWorkspace.value?.path))

const activeTodo = computed(() => {
  return todos.value.find((todo) => todo.id === activeTodoId.value) || null
})

const activeTodoProject = computed(() => {
  return todoProjects.value.find((todoProject) => todoProject.id === activeTodoProjectId.value) || null
})

const activeTodoProjectProject = computed(() => {
  const todoProject = activeTodoProject.value
  if (!todoProject) {
    return null
  }
  return todoProjectDisplayProject(todoProject)
})

const activeTerminal = computed(() => {
  return terminals.value.find((terminal) => terminal.id === activeTerminalId.value) || null
})

const activeTerminalIsTaskTerminal = computed(() => {
  const terminal = activeTerminal.value
  return Boolean(terminal && !terminal.workspaceTerminal && terminal.todoId && !terminal.todoProjectId)
})

const activeTerminalState = computed(() => {
  return activeTerminal.value ? terminalState(activeTerminal.value.id) : ''
})

const canRestartActiveShell = computed(() => {
  return activeTerminalState.value === 'exited' && terminalCanRestart(activeTerminal.value)
})

const terminalStateLayer = computed(() => {
  if (!hasWorkspace.value) {
    return { text: 'Open a project', testId: 'workspace-empty-state', warning: false }
  }

  const terminal = activeTerminal.value
  if (terminal?.workspaceTerminal) {
    if (activeTerminalState.value === 'unsupported') {
      return { text: 'Embedded terminal is not supported on Windows', warning: true }
    }
    if (activeTerminalState.value === 'exited') {
      return { text: 'Shell exited', warning: true }
    }
    return null
  }

  if (activeTerminalIsTaskTerminal.value) {
    if (activeTerminalState.value === 'unsupported') {
      return { text: 'Embedded terminal is not supported on Windows', warning: true }
    }
    if (activeTerminalState.value === 'exited') {
      return { text: 'Shell exited', warning: true }
    }
    return null
  }

  if (!activeTodoProject.value || !activeTodoProjectProject.value) {
    return { text: 'Select a TODO project', warning: false }
  }
  if (!activeTodoProjectProject.value.available) {
    return { text: 'Project path unavailable', warning: true }
  }
  if (!terminal) {
    return { text: 'Select a terminal', warning: false }
  }
  if (activeTerminalState.value === 'unsupported') {
    return { text: 'Embedded terminal is not supported on Windows', warning: true }
  }
  if (activeTerminalState.value === 'exited') {
    return { text: 'Shell exited', warning: true }
  }
  return null
})

const selectedTerminalShell = computed(() => terminalSettings.value?.selected || null)
const terminalLaunchProfiles = computed(() => {
  if (Array.isArray(terminalSettings.value?.launchProfiles)) {
    return terminalSettings.value.launchProfiles
  }
  return defaultTerminalLaunchProfiles
})

const currentGitStatus = computed(() => {
  const context = activeGitStatusContext.value
  if (!context || gitStatus.value?.contextKey !== context.key) {
    return null
  }
  return gitStatus.value
})

const displayedGitStatus = computed(() => {
  const context = activeGitStatusContext.value
  if (!context || !gitStatus.value) {
    return null
  }
  if (gitStatusLoading.value && gitStatus.value.contextKey !== context.key) {
    return null
  }
  return gitStatus.value
})

const projectGitStatusChips = computed(() => {
  const context = activeGitStatusContext.value
  if (!context) {
    return []
  }
  const status = displayedGitStatus.value
  if (context.type === 'todo' && status && !status.isRepo) {
    return []
  }
  if ((context.type === 'project' && !activeProject.value?.available) || status?.pathUnavailable) {
    return [statusChip('warning', 'warning', 'Project path unavailable')]
  }
  if (status?.gitUnavailable) {
    return [statusChip('git-unavailable', 'warning', '未安装 Git')]
  }
  if (gitStatusLoading.value) {
    return [statusChip('neutral', 'neutral', 'Loading git status')]
  }
  if (gitStatusError.value) {
    return [statusChip('error', 'error', 'Git status unavailable')]
  }
  if (status && !status.isRepo) {
    return [statusChip('warning', 'warning', 'Not a git repository')]
  }
  if (status?.isRepo) {
    const branch = status.branch || '(detached)'
    const chips = [
      statusChip('branch', 'branch', branch),
      statusChip('changed', 'changed', `${status.changedCount || 0} changed`)
    ]
    addCountChip(chips, 'staged', status.stagedCount, 'staged')
    addCountChip(chips, 'unstaged', status.unstagedCount, 'unstaged')
    addCountChip(chips, 'untracked', status.untrackedCount, 'untracked')
    addCountChip(chips, 'ahead', status.ahead, 'ahead')
    addCountChip(chips, 'behind', status.behind, 'behind')
    return chips
  }
  return [statusChip('error', 'error', 'Git status unavailable')]
})

const showInitializeGitRepository = computed(() => {
  const status = currentGitStatus.value
  return Boolean(
    activeProject.value &&
    activeGitStatusContext.value?.type === 'project' &&
    activeProject.value.available &&
    status &&
    !status.pathUnavailable &&
    !status.gitUnavailable &&
    !status.isRepo &&
    !gitStatusLoading.value &&
    !gitStatusError.value
  )
})

const gitInitializeButtonText = computed(() => {
  return gitInitLoading.value ? 'Initializing Git Repository' : 'Initialize Git Repository'
})

const terminalSettingsDetected = computed(() => {
  return detectedTerminalShell.value || terminalSettings.value?.detected || terminalSettings.value?.fallback || null
})

const terminalSettingsFallback = computed(() => terminalSettings.value?.fallback || null)

const selectedTodoFormProjects = computed(() => {
  const selectedProjectIds = new Set(todoForm.projectSelections.map((selection) => selection.projectId))
  return projects.value.filter((project) => selectedProjectIds.has(project.id))
})

const todoFormProjectOptions = computed(() => {
  return filteredProjects(projects.value, todoForm.projectSearch)
})

const todoFormLifecycleScriptOptions = computed(() => {
  return todoForm.lifecycleScripts.map((script, index) => ({ ...script, index }))
})

const selectedTodoFormLifecycleScript = computed(() => {
  if (todoForm.selectedLifecycleScriptIndex === '') {
    return null
  }
  const index = Number(todoForm.selectedLifecycleScriptIndex)
  if (!Number.isInteger(index) || index < 0 || index >= todoForm.lifecycleScripts.length) {
    return null
  }
  return todoForm.lifecycleScripts[index] || null
})

const selectedTodoFormLifecycleScriptLabel = computed(() => {
  return lifecycleScriptOptionLabel(selectedTodoFormLifecycleScript.value)
})

const selectedTodoDetailProjects = computed(() => {
  if (todoDetail.readOnly) {
    return []
  }
  const selectedProjectIds = new Set(todoDetail.projectSelections.map((selection) => selection.projectId))
  const selected = []
  const seenProjectIds = new Set()
  for (const project of projects.value) {
    if (selectedProjectIds.has(project.id)) {
      selected.push(project)
      seenProjectIds.add(project.id)
    }
  }
  for (const todoProject of todoProjectsForTodo(todoDetail.todoId)) {
    if (selectedProjectIds.has(todoProject.projectId) && !seenProjectIds.has(todoProject.projectId)) {
      const project = todoProjectDisplayProject(todoProject)
      if (project) {
        selected.push(project)
        seenProjectIds.add(project.id)
      }
    }
  }
  return selected
})

const todoDetailProjectOptions = computed(() => {
  if (todoDetail.readOnly) {
    return []
  }
  return filteredProjects(projects.value, todoDetail.projectSearch)
})

const removedTodoDetailProjectsWithTerminals = computed(() => {
  if (!todoDetail.visible || !todoDetail.todoId) {
    return []
  }

  const selectedProjectIds = new Set(todoDetail.projectSelections.map((selection) => selection.projectId))
  return todoProjects.value.filter((todoProject) => {
    return (
      todoProject.todoId === todoDetail.todoId &&
      !selectedProjectIds.has(todoProject.projectId) &&
      terminals.value.some((terminal) => terminal.todoProjectId === todoProject.id)
    )
  })
})

const projectPickerOptions = computed(() => {
  const linkedProjectPaths = new Set(
    todoProjects.value
      .filter((todoProject) => todoProject.todoId === projectPicker.todoId)
      .map((todoProject) => normalizeProjectPath(todoProject.path))
      .filter(Boolean)
  )
  const linkedProjectIds = new Set(
    todoProjects.value
      .filter((todoProject) => todoProject.todoId === projectPicker.todoId && !todoProject.path)
      .map((todoProject) => todoProject.projectId)
  )
  return filteredProjects(
    projects.value.filter((project) => {
      const path = normalizeProjectPath(project.path)
      return path ? !linkedProjectPaths.has(path) : !linkedProjectIds.has(project.id)
    }),
    projectPicker.query
  )
})

const selectedProjectPickerProjects = computed(() => {
  const selectedProjectIds = new Set(projectPicker.projectSelections.map((selection) => selection.projectId))
  return projects.value.filter((project) => selectedProjectIds.has(project.id))
})

const workspaceTerminals = computed(() =>
  terminals.value.filter((terminal) => terminal.workspaceTerminal).map(terminalWithAttention)
)

const sidebarTerminals = computed(() =>
  terminals.value.filter((terminal) => !terminal.workspaceTerminal).map(terminalWithAttention)
)

onMounted(async () => {
  EventsOn('terminal-output', (event) => {
    terminalManager.write(event.terminalId, event.data)
  })
  EventsOn('terminal-command-state', (event) => {
    terminalManager.onCommandState(event.terminalId, event)
  })
  EventsOn('terminal-status', (status) => {
    updateTerminalState(status.terminalId, status.state)
  })
  EventsOn('terminal-agent-status', (event) => {
    handleTerminalAgentStatus(event)
  })
  EventsOn('todo-lifecycle-script-status', (status) => {
    handleTodoLifecycleScriptStatus(status)
  })
  EventsOn('workspace-state', (state) => {
    void applyWorkspaceProjectState(state)
  })
  EventsOn('workspace-recent', (state) => {
    showRecentWorkspacePicker(state)
  })
  window.addEventListener('resize', fitActiveTerminal)
  window.addEventListener('focus', refreshProjectGitStatusOnFocus)
  window.addEventListener('click', closeTerminalMenu)
  window.addEventListener('click', closeGlobalManagementMenu)

  try {
    applyState(await ListProjects())
    await loadTodoProjectUIStateForCurrentWorkspace({ restoreTodoProjectUIState: true })
    await loadTerminalSettingsForCurrentWorkspace()
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
})

onBeforeUnmount(() => {
  EventsOff('terminal-output')
  EventsOff('terminal-command-state')
  EventsOff('terminal-status')
  EventsOff('terminal-agent-status')
  EventsOff('todo-lifecycle-script-status')
  EventsOff('workspace-state')
  EventsOff('workspace-recent')
  window.removeEventListener('resize', fitActiveTerminal)
  window.removeEventListener('focus', refreshProjectGitStatusOnFocus)
  window.removeEventListener('click', closeTerminalMenu)
  window.removeEventListener('click', closeGlobalManagementMenu)
  window.removeEventListener('mousemove', resizeSidebar)
  window.removeEventListener('mouseup', stopSidebarResize)
  clearAllTitleActivityTimers()
  clearToastTimer()
  resolveGitInitializationPrompt(false)
})

function applyState(state, options = {}) {
  const previousGitStatusContextKey = activeGitStatusContextKey()
  const previousTerminals = new Map(terminals.value.map((terminal) => [terminal.id, terminal]))
  currentWorkspace.value = state?.currentWorkspace || null
  recentWorkspaces.value = state?.recentWorkspaces || []
  projects.value = state?.projects || []
  todos.value = state?.todos || []
  todoProjects.value = state?.todoProjects || []
  projectBranchPreferences.value = state?.projectBranchPreferences || {}
  lifecycleScriptStatuses.value = state?.lifecycleScriptStatuses || []
  importSummary.value = state?.importSummary || null
  const nextTerminals = (state?.terminals || []).map((terminal) => {
    const previous = previousTerminals.get(terminal.id)
    const running = terminal.state === 'running'
    return {
      ...terminal,
      currentCommand: running ? terminal.currentCommand || previous?.currentCommand || '' : '',
      runtimeTitle: running ? previous?.runtimeTitle || '' : '',
      agentStatus: running ? previous?.agentStatus || createAgentStatus() : exitedAgentStatus(),
      activityState: running ? previous?.activityState || 'idle' : 'idle'
    }
  })
  const nextTerminalIds = new Set(nextTerminals.map((terminal) => terminal.id))
  for (const terminalId of previousTerminals.keys()) {
    if (!nextTerminalIds.has(terminalId)) {
      terminalManager.dispose(terminalId)
      terminalContainers.delete(terminalId)
      clearTitleActivityTimer(terminalId)
      autoRestartedTerminals.delete(terminalId)
      terminalAckIds.delete(terminalId)
      delete shellStatuses[terminalId]
    }
  }
  terminals.value = nextTerminals
  activeProjectId.value = state?.activeProjectId || ''
  activeTodoId.value = state?.activeTodoId || ''
  activeTodoProjectId.value = state?.activeTodoProjectId || ''
  activeTerminalId.value = state?.activeTerminalId || ''
  pruneTodoProjectGitBranches()
  for (const terminal of terminals.value) {
    if (terminal.state) {
      shellStatuses[terminal.id] = terminal.state
    }
  }
  if (!hasWorkspace.value) {
    closeWorkspaceScopedPanels()
  }
  if (options.restoreTodoProjectUIState === true) {
    applyTodoProjectUIState(activeTodoProjectId.value)
  }
  closeTerminalMenu()
  syncGitStatusForActiveProject(previousGitStatusContextKey, {
    refresh: options.refreshGitStatus !== false,
    dedupePending: options.dedupeGitStatus === true,
    force: options.forceGitStatusRefresh === true
  })
  refreshTodoProjectBranchStatuses({ dedupePending: true })
  scheduleCompletedMergeStatusRefresh()
}

function handleTodoLifecycleScriptStatus(status) {
  if (!status?.todoId || !status?.phase) {
    return
  }
  const nextStatuses = lifecycleScriptStatuses.value.filter(
    (candidate) => candidate.todoId !== status.todoId || candidate.phase !== status.phase
  )
  if (status.status) {
    nextStatuses.push(status)
  }
  lifecycleScriptStatuses.value = nextStatuses
}

async function applyWorkspaceProjectState(state, options = {}) {
  const restoreTodoProjectUIState =
    options.restoreTodoProjectUIState === true ||
    (state?.currentWorkspace?.path || '') !== (currentWorkspace.value?.path || '')
  applyState(state, { forceGitStatusRefresh: true, restoreTodoProjectUIState })
  await loadTodoProjectUIStateForCurrentWorkspace({ restoreTodoProjectUIState })
  errorMessage.value = ''
  await activateActiveTerminal()
}

function showRecentWorkspacePicker(state) {
  recentWorkspaces.value = state?.recentWorkspaces || []
  recentWorkspacePicker.visible = true
  recentWorkspacePicker.openingPath = ''
  recentWorkspacePicker.error = ''
}

function closeRecentWorkspacePicker() {
  recentWorkspacePicker.visible = false
  recentWorkspacePicker.openingPath = ''
  recentWorkspacePicker.error = ''
}

async function openRecentWorkspace(workspace) {
  if (!workspace?.path || recentWorkspacePicker.openingPath) {
    return
  }
  recentWorkspacePicker.openingPath = workspace.path
  recentWorkspacePicker.error = ''
  try {
    await applyWorkspaceProjectState(await OpenRecentWorkspace(workspace.path), { restoreTodoProjectUIState: true })
    closeRecentWorkspacePicker()
  } catch (error) {
    recentWorkspacePicker.error = errorMessageFrom(error)
  } finally {
    recentWorkspacePicker.openingPath = ''
  }
}

async function loadTerminalSettingsForCurrentWorkspace() {
  try {
    applyTerminalSettings(await LoadTerminalSettings())
  } catch (error) {
    showError(error)
  }
}

function closeWorkspaceScopedPanels() {
  closeTodoForm()
  closeTodoDetail()
  closeProjectPicker()
}

async function loadTodoProjectUIStateForCurrentWorkspace(options = {}) {
  if (!hasWorkspace.value) {
    todoProjectUIStates.value = {}
    todoSidebarWidthState.value = 0
    if (options.restoreTodoProjectUIState === true) {
      applyTodoWorkspaceUIState('')
    }
    return
  }
  try {
    const state = await LoadTodoProjectUIState()
    todoProjectUIStates.value = state?.todoProjects || {}
    todoSidebarWidthState.value = Number(state?.sidebarWidth) || 0
    if (options.restoreTodoProjectUIState === true) {
      applyTodoWorkspaceUIState(activeTodoProjectId.value)
    }
  } catch (error) {
    todoProjectUIStates.value = {}
    todoSidebarWidthState.value = 0
    if (options.restoreTodoProjectUIState === true) {
      applyTodoWorkspaceUIState(activeTodoProjectId.value)
    }
    showError(error)
  }
}

async function createProject() {
  const previousProjectIds = new Set(projects.value.map((project) => project.id).filter(Boolean))
  try {
    const imported = await resolveProjectImportResult(await CreateProjectFromDialog(), previousProjectIds)
    if (imported?.state) {
      applyState(imported.state)
    }
  } catch (error) {
    showError(error)
  }
}

async function importSingleProjectCandidate() {
  const previousProjectIds = new Set(projects.value.map((project) => project.id).filter(Boolean))
  try {
    const imported = await resolveProjectImportResult(await CreateProjectFromDialog(), previousProjectIds)
    if (!imported?.state) {
      return
    }
    applyState(imported.state, { refreshGitStatus: false })
    selectImportedProjectCandidate(imported.projectId)
  } catch (error) {
    showError(error)
  }
}

async function resolveProjectImportResult(result, previousProjectIds) {
  if (result?.requiresGitInitialization) {
    const path = result.path || ''
    if (!await requestGitInitializationConfirmation(path)) {
      showToast(gitOnlyImportToastText)
      return null
    }
    const state = await InitializeGitRepositoryAndImportProject(path)
    return {
      state,
      projectId: importedSingleProjectId(state, previousProjectIds)
    }
  }

  const state = result?.state || result
  if (!state) {
    return null
  }
  return {
    state,
    projectId: importedSingleProjectId(state, previousProjectIds)
  }
}

function requestGitInitializationConfirmation(path) {
  if (gitInitializationPromptResolve) {
    gitInitializationPromptResolve(false)
  }
  gitInitializationPrompt.visible = true
  gitInitializationPrompt.path = path
  return new Promise((resolve) => {
    gitInitializationPromptResolve = resolve
  })
}

function resolveGitInitializationPrompt(confirmed) {
  const resolve = gitInitializationPromptResolve
  gitInitializationPromptResolve = null
  gitInitializationPrompt.visible = false
  gitInitializationPrompt.path = ''
  if (resolve) {
    resolve(confirmed)
  }
}

async function importProjectsFromParentDirectory() {
  try {
    applyState(await ImportProjectsFromParentDirectoryDialog(), { refreshGitStatus: false })
  } catch (error) {
    showError(error)
  }
}

async function selectProject(projectId) {
  try {
    applyState(await SelectProject(projectId))
  } catch (error) {
    showError(error)
  }
}

async function createTodo() {
  todoForm.visible = true
  todoForm.title = ''
  todoForm.description = ''
  todoForm.priority = 'medium'
  todoForm.projectSelections = []
  todoForm.projectSearch = ''
  todoForm.initializationFiles = []
  todoForm.lifecycleScripts = []
  todoForm.selectedLifecycleScriptIndex = ''
  todoForm.lifecycleScriptMenuOpen = false
  todoForm.saving = false
  errorMessage.value = ''
  try {
    const [templates, lifecycleScripts] = await Promise.all([
      LoadTodoInitializationFiles(),
      LoadTodoLifecycleScripts()
    ])
    todoForm.initializationFiles = cloneTodoInitializationFiles(templates).map((file) => ({
      ...file,
      selected: file.defaultSelected === true
    }))
    todoForm.lifecycleScripts = cloneTodoLifecycleScripts(lifecycleScripts)
    const defaultIndex = todoForm.lifecycleScripts.findIndex((script) => script.defaultSelected === true)
    todoForm.selectedLifecycleScriptIndex = defaultIndex >= 0 ? String(defaultIndex) : ''
  } catch (error) {
    showError(error)
  }
}

function closeTodoForm() {
  todoForm.visible = false
  todoForm.projectSelections = []
  todoForm.initializationFiles = []
  todoForm.lifecycleScripts = []
  todoForm.selectedLifecycleScriptIndex = ''
  todoForm.lifecycleScriptMenuOpen = false
  todoForm.saving = false
  closeProjectBranchPicker()
}

async function submitTodoForm() {
  const title = todoForm.title.trim()
  if (!title) {
    showError('TODO title is required')
    return
  }

  todoForm.saving = true
  try {
    const request = {
      title,
      description: todoForm.description.trim(),
      priority: normalizedTodoPriority(todoForm.priority),
      projects: todoProjectSelectionPayload(todoForm.projectSelections)
    }
    const initializationFiles = selectedTodoInitializationFileSnapshots(todoForm.initializationFiles)
    if (initializationFiles.length > 0) {
      request.initializationFiles = initializationFiles
    }
    const lifecycleScript = selectedTodoLifecycleScriptSnapshot(selectedTodoFormLifecycleScript.value)
    if (lifecycleScript) {
      request.lifecycleScript = lifecycleScript
    }
    applyState(await CreateTodo(request))
    closeTodoForm()
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  } finally {
    todoForm.saving = false
  }
}

function editTodo(todoId) {
  const todo = todos.value.find((candidate) => candidate.id === todoId)
  if (!todo) {
    return
  }

  const linkedTodoProjects = todoProjects.value.filter((todoProject) => todoProject.todoId === todo.id)

  todoDetail.visible = true
  todoDetail.todoId = todo.id
  todoDetail.title = todo.title || ''
  todoDetail.description = todo.description || ''
  todoDetail.priority = normalizedTodoPriority(todo.priority)
  todoDetail.projectSelections = linkedTodoProjects.map((todoProject) => ({
    projectId: todoProject.projectId,
    baseBranch: todoProject.baseBranch || defaultBaseBranch(todoProject.projectId)
  }))
  for (const selection of todoDetail.projectSelections) {
    void ensureProjectBranchesLoaded(selection.projectId)
  }
  todoDetail.projectSnapshots = Array.isArray(todo.projectSnapshots) ? todo.projectSnapshots : []
  todoDetail.projectSearch = ''
  todoDetail.readOnly = todo.status === 'completed'
  todoDetail.saving = false
  errorMessage.value = ''
}

function closeTodoDetail() {
  todoDetail.visible = false
  todoDetail.todoId = ''
  todoDetail.title = ''
  todoDetail.description = ''
  todoDetail.priority = 'medium'
  todoDetail.projectSelections = []
  todoDetail.projectSnapshots = []
  todoDetail.projectSearch = ''
  todoDetail.readOnly = false
  todoDetail.saving = false
  closeProjectBranchPicker()
}

async function toggleTodoDetailProject(project) {
  if (todoDetail.readOnly || !project?.id) {
    return
  }
  todoDetail.projectSelections = toggleProjectSelection(todoDetail.projectSelections, project.id)
  if (hasProjectSelection(todoDetail.projectSelections, project.id)) {
    await ensureProjectBranchesLoaded(project.id)
  }
}

function removeTodoDetailProject(projectId) {
  if (todoDetail.readOnly) {
    return
  }
  todoDetail.projectSelections = removeProjectSelection(todoDetail.projectSelections, projectId)
}

async function submitTodoDetail() {
  if (todoDetail.readOnly) {
    return
  }
  const title = todoDetail.title.trim()
  if (!title) {
    showError('TODO title is required')
    return
  }

  if (
    removedTodoDetailProjectsWithTerminals.value.length > 0 &&
    !window.confirm('Saving will close terminals for removed projects. Continue?')
  ) {
    return
  }

  todoDetail.saving = true
  try {
    applyState(
      await UpdateTodo({
        id: todoDetail.todoId,
        title,
        description: todoDetail.description.trim(),
        priority: normalizedTodoPriority(todoDetail.priority),
        projects: todoProjectSelectionPayload(todoDetail.projectSelections)
      })
    )
    closeTodoDetail()
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  } finally {
    todoDetail.saving = false
  }
}

async function addProjectToTodo(todoId) {
  projectPicker.todoId = todoId
  projectPicker.query = ''
  projectPicker.projectSelections = []
  projectPicker.saving = false
  projectPicker.visible = true
  errorMessage.value = ''
}

function closeProjectPicker() {
  projectPicker.visible = false
  projectPicker.todoId = ''
  projectPicker.query = ''
  projectPicker.projectSelections = []
  projectPicker.saving = false
  closeProjectBranchPicker()
}

async function toggleProjectForTodo(project) {
  if (!project?.id) {
    return
  }
  projectPicker.projectSelections = toggleProjectSelection(projectPicker.projectSelections, project.id)
  if (hasProjectSelection(projectPicker.projectSelections, project.id)) {
    await ensureProjectBranchesLoaded(project.id)
  }
}

function removeProjectPickerProject(projectId) {
  projectPicker.projectSelections = removeProjectSelection(projectPicker.projectSelections, projectId)
}

async function submitProjectPicker() {
  if (!projectPicker.todoId || projectPicker.projectSelections.length === 0) {
    return
  }
  projectPicker.saving = true
  try {
    applyState(await AddProjectSelectionsToTodo(projectPicker.todoId, todoProjectSelectionPayload(projectPicker.projectSelections)))
    closeProjectPicker()
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  } finally {
    projectPicker.saving = false
  }
}

async function toggleTodoFormProject(project) {
  if (!project?.id) {
    return
  }
  todoForm.projectSelections = toggleProjectSelection(todoForm.projectSelections, project.id)
  if (hasProjectSelection(todoForm.projectSelections, project.id)) {
    await ensureProjectBranchesLoaded(project.id)
  }
}

function removeTodoFormProject(projectId) {
  todoForm.projectSelections = removeProjectSelection(todoForm.projectSelections, projectId)
}

async function selectTodoProject(todoProjectId) {
  try {
    applyState(await SelectTodoProject(todoProjectId), {
      dedupeGitStatus: true,
      forceGitStatusRefresh: true
    })
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function removeTodoProject(todoProjectId) {
  try {
    applyState(await RemoveTodoProject(todoProjectId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function selectTerminal(terminalId) {
  try {
    terminalAckIds.delete(terminalId)
    applyState(await SelectTerminal(terminalId))
    await activateActiveTerminal()
    terminalManager.focus(terminalId)
    await autoRestartIfExited(terminalId)
  } catch (error) {
    showError(error)
  }
}

async function autoRestartIfExited(terminalId) {
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (!terminal || terminal.state !== 'exited') {
    return
  }
  if (autoRestartedTerminals.has(terminalId)) {
    return
  }
  if (!terminalCanRestart(terminal)) {
    return
  }
  autoRestartedTerminals.add(terminalId)
  const size = terminalManager.size(terminalId) || { cols: 80, rows: 24 }
  try {
    const status = await StartShell(terminalId, size.cols || 80, size.rows || 24)
    updateTerminalState(status.terminalId, status.state)
  } catch (error) {
    autoRestartedTerminals.delete(terminalId)
    showError(error)
  }
}

async function createTerminal(todoProjectId, launchProfile = null) {
  try {
    if (launchProfile?.background === true) {
      await StartTodoProjectBackgroundCommand(todoProjectId, launchProfile.command)
      return
    }
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    const state = await CreateTodoTerminal(todoProjectId, size.cols || 80, size.rows || 24)
    applyState(state)
    await activateActiveTerminal()
    await runLaunchProfileCommand(state, launchProfile)
  } catch (error) {
    showError(error)
  }
}

async function createTaskTerminal(todoId, launchProfile = null) {
  try {
    if (launchProfile?.background === true) {
      await StartTaskBackgroundCommand(todoId, launchProfile.command)
      return
    }
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    const state = await CreateTaskTerminal(todoId, size.cols || 80, size.rows || 24)
    applyState(state)
    await activateActiveTerminal()
    await runLaunchProfileCommand(state, launchProfile)
  } catch (error) {
    showError(error)
  }
}

async function runLaunchProfileCommand(state, launchProfile) {
  if (!launchProfile?.command || !state?.activeTerminalId) {
    return
  }
  const terminal = terminals.value.find((candidate) => candidate.id === state.activeTerminalId)
  if (terminal) {
    applyTerminalAgentEvent(terminal, {
      type: 'launch-profile-label',
      command: launchProfile.command,
      at: Date.now()
    })
  }
  await SendTerminalInput(state.activeTerminalId, `${launchProfile.command}\r`)
}

async function createWorkspaceTerminal() {
  if (!hasWorkspace.value) {
    return
  }
  try {
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    const state = await CreateWorkspaceTerminal(size.cols || 80, size.rows || 24)
    applyState(state)
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function completeTodo(todoId) {
  try {
    applyState(await CompleteTodo(todoId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function retryTodoLifecycleScript(todoId, phase) {
  try {
    applyState(await RetryTodoLifecycleScript(todoId, phase))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function deleteTodo(todoId) {
  try {
    applyState(await DeleteTodo(todoId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function openTodoFolder(todoId) {
  try {
    await OpenTodoFolder(todoId)
  } catch (error) {
    showError(error)
  }
}

async function deleteCompletedTodos(todoIds) {
  if (!Array.isArray(todoIds) || todoIds.length === 0) {
    return
  }
  try {
    applyState(await DeleteCompletedTodos(todoIds))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

function copyTodoDescription(todoId) {
  const todo = todos.value.find((candidate) => candidate.id === todoId)
  if (!todo) {
    return
  }
  const title = todo.title || ''
  const description = todo.description || ''
  const text = description ? `${title}\n${description}` : title
  ClipboardSetText(text)
}

async function changeTodoStatus(todoId, status) {
  try {
    applyState(await ChangeTodoStatus(todoId, status))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function deleteProject(projectId) {
  try {
    applyState(await DeleteProject(projectId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function deleteProjects(projectIds) {
  if (!Array.isArray(projectIds) || projectIds.length === 0) {
    return
  }
  try {
    applyState(await DeleteProjects(projectIds))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function clearGlobalProjectCandidates() {
  const projectIds = projects.value.map((project) => project.id).filter(Boolean)
  if (projectIds.length === 0) {
    return
  }
  if (!window.confirm('Clear global project candidates?')) {
    return
  }
  try {
    applyState(await DeleteProjects(projectIds), { refreshGitStatus: false })
  } catch (error) {
    showError(error)
  }
}

async function clearProjectCandidate(project) {
  if (!project?.id) {
    return
  }
  projectCandidateClearPrompt.project = { ...project }
  projectCandidateClearPrompt.visible = true
  projectCandidateClearPrompt.clearing = false
}

function closeProjectCandidateClearPrompt(force = false) {
  if (projectCandidateClearPrompt.clearing && !force) {
    return
  }
  projectCandidateClearPrompt.visible = false
  projectCandidateClearPrompt.project = null
  projectCandidateClearPrompt.clearing = false
}

async function confirmProjectCandidateClear() {
  const project = projectCandidateClearPrompt.project
  if (!project?.id) {
    closeProjectCandidateClearPrompt(true)
    return
  }
  projectCandidateClearPrompt.clearing = true
  try {
    applyState(await DeleteProject(project.id), { refreshGitStatus: false })
    removePendingProjectSelection(project.id)
    closeProjectCandidateClearPrompt(true)
  } catch (error) {
    projectCandidateClearPrompt.clearing = false
    showError(error)
  }
}

function projectCandidateClearLabel() {
  const project = projectCandidateClearPrompt.project
  return project?.name || project?.path || project?.id || ''
}

function projectCandidateClearPath() {
  return projectCandidateClearPrompt.project?.path || ''
}

async function deleteTerminal(terminalId) {
  try {
    applyState(await DeleteTerminal(terminalId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function activateActiveTerminal() {
  await nextTick()
  const terminal = activeTerminal.value
  if (!terminal) {
    return
  }
  const container = terminalContainers.get(terminal.id)
  if (!container) {
    return
  }

  terminalManager.activate(terminal.id, container)
  terminalManager.fitActive()
  if (terminal.state !== 'running' && terminal.output) {
    await nextTick()
    terminalManager.replayHistory(terminal.id, terminal.output)
  }
}

async function restartActiveShell() {
  const terminal = activeTerminal.value
  if (!terminal || !terminalCanRestart(terminal)) {
    return
  }
  const size = terminalManager.size(terminal.id) || { cols: 80, rows: 24 }
  try {
    const status = await StartShell(terminal.id, size.cols || 80, size.rows || 24)
    autoRestartedTerminals.add(terminal.id)
    updateTerminalState(status.terminalId, status.state)
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

function setTerminalContainer(terminalId, element) {
  if (element) {
    terminalContainers.set(terminalId, element)
    if (terminalId === activeTerminalId.value) {
      nextTick(() => activateActiveTerminal())
    }
  } else {
    terminalContainers.delete(terminalId)
  }
}

function openTerminalMenu(terminalId, event) {
  if (terminalId !== activeTerminalId.value) {
    return
  }
  terminalMenu.terminalId = terminalId
  terminalMenu.x = event.clientX
  terminalMenu.y = event.clientY
  terminalMenu.visible = true
}

function closeTerminalMenu() {
  terminalMenu.visible = false
  terminalMenu.terminalId = ''
}

function toggleGlobalManagementMenu() {
  closeTerminalMenu()
  globalManagementMenu.visible = !globalManagementMenu.visible
}

function closeGlobalManagementMenu() {
  globalManagementMenu.visible = false
}

async function copyFromTerminalMenu() {
  const terminalId = terminalMenu.terminalId
  await terminalManager.copySelection(terminalId)
  closeTerminalMenu()
}

async function pasteFromTerminalMenu() {
  const terminalId = terminalMenu.terminalId
  await terminalManager.paste(terminalId)
  closeTerminalMenu()
  await nextTick()
  terminalManager.focus(terminalId)
}

function hasTerminalSelection(terminalId) {
  return terminalManager.hasSelection(terminalId)
}

async function openTerminalSettings() {
  closeTerminalMenu()
  closeGlobalManagementMenu()
  settingsPanel.visible = true
  settingsPanel.loading = true
  settingsPanel.error = ''
  try {
    const state = await LoadTerminalSettings()
    applyTerminalSettings(state)
    if (!detectedTerminalShell.value) {
      detectedTerminalShell.value = await DetectTerminalShell()
    }
    settingsPanel.mode = detectedTerminalShell.value ? 'detected' : 'manual'
  } catch (error) {
    settingsPanel.error = errorMessageFrom(error)
  } finally {
    settingsPanel.loading = false
  }
  // Load Claude status hook state independently so a failure here does not
  // block the terminal shell settings above.
  await loadClaudeStatusHookState()
}

function closeTerminalSettings() {
  settingsPanel.visible = false
  settingsPanel.error = ''
}

async function loadClaudeStatusHookState() {
  const projectPath = activeProjectPath.value
  if (!projectPath) {
    settingsPanel.claudeStatus = {
      installed: false,
      stale: false,
      checking: false,
      command: '',
      eventsCovered: 0
    }
    return
  }
  settingsPanel.claudeStatus.checking = true
  try {
    const state = await ClaudeStatusHookState(projectPath)
    settingsPanel.claudeStatus = {
      installed: Boolean(state?.installed),
      stale: Boolean(state?.stale),
      checking: false,
      command: state?.command || '',
      eventsCovered: state?.eventsCovered || 0
    }
  } catch (error) {
    settingsPanel.claudeStatus.checking = false
    showError(error)
  }
}

async function installClaudeStatusHook() {
  const projectPath = activeProjectPath.value
  if (!projectPath) {
    return
  }
  try {
    await EnsureClaudeStatusHook(projectPath)
    showToast('已安装 Claude 状态监控')
  } catch (error) {
    showError(error)
  }
  await loadClaudeStatusHookState()
}

async function uninstallClaudeStatusHook() {
  const projectPath = activeProjectPath.value
  if (!projectPath) {
    return
  }
  try {
    await RemoveClaudeStatusHook(projectPath)
    showToast('已卸载 Claude 状态监控')
  } catch (error) {
    showError(error)
  }
  await loadClaudeStatusHookState()
}

function applyTerminalSettings(state) {
  terminalSettings.value = state || null
  detectedTerminalShell.value = state?.detected || state?.fallback || null
  settingsPanel.manualPath = state?.selected?.path || settingsPanel.manualPath || ''
  settingsPanel.launchProfiles = cloneLaunchProfiles(launchProfilesFromState(state))
  applyAppearanceTheme(state?.theme)
  settingsPanel.theme = currentTheme.value
}

function applyAppearanceTheme(theme) {
  currentTheme.value = normalizeAppearanceTheme(theme)
}

function normalizeAppearanceTheme(theme) {
  return appearanceThemes.includes(theme) ? theme : 'light'
}

function launchProfilesFromState(state) {
  if (Array.isArray(state?.launchProfiles)) {
    return state.launchProfiles
  }
  return defaultTerminalLaunchProfiles
}

function cloneLaunchProfiles(profiles) {
  return profiles.map((profile) => ({
    name: profile.name || '',
    command: profile.command || '',
    enabled: profile.enabled !== false,
    background: profile.background === true
  }))
}

function cloneTodoInitializationFiles(files) {
  return files.map((file) => ({
    name: file.name || '',
    description: file.description || '',
    fileName: file.fileName || '',
    content: file.content || '',
    defaultSelected: file.defaultSelected === true
  }))
}

function cloneTodoLifecycleScripts(scripts = []) {
  return scripts.map((script) => ({
    name: script.name || '',
    description: script.description || '',
    initScript: script.initScript || '',
    completeScript: script.completeScript || '',
    defaultSelected: script.defaultSelected === true
  }))
}

function addTodoInitializationFile() {
  initializationFileManagement.files.push({
    name: '',
    description: '',
    fileName: '',
    content: '',
    defaultSelected: false
  })
}

function removeTodoInitializationFile(index) {
  initializationFileManagement.files.splice(index, 1)
}

function moveTodoInitializationFile(index, direction) {
  const nextIndex = index + direction
  if (nextIndex < 0 || nextIndex >= initializationFileManagement.files.length) {
    return
  }
  const [file] = initializationFileManagement.files.splice(index, 1)
  initializationFileManagement.files.splice(nextIndex, 0, file)
}

function normalizedTodoInitializationFiles() {
  return initializationFileManagement.files.map((file) => ({
    name: (file.name || '').trim(),
    description: (file.description || '').trim(),
    fileName: (file.fileName || '').trim(),
    content: file.content || '',
    defaultSelected: file.defaultSelected === true
  }))
}

function selectedTodoInitializationFileSnapshots(files) {
  return files
    .filter((file) => file.selected === true)
    .map((file) => ({
      name: (file.name || '').trim(),
      description: (file.description || '').trim(),
      fileName: (file.fileName || '').trim(),
      content: file.content || ''
    }))
}

function selectedTodoLifecycleScriptSnapshot(script) {
  if (!script) {
    return null
  }
  const snapshot = {
    name: (script.name || '').trim(),
    description: (script.description || '').trim(),
    initScript: (script.initScript || '').trim(),
    completeScript: (script.completeScript || '').trim()
  }
  if (!snapshot.name || (!snapshot.initScript && !snapshot.completeScript)) {
    return null
  }
  return snapshot
}

function lifecycleScriptOptionLabel(script) {
  if (!script) {
    return '不选择脚本'
  }
  const name = (script.name || '').trim() || '未命名脚本'
  const description = (script.description || '').trim()
  return description ? `${name} - ${description}` : name
}

function toggleTodoFormLifecycleScriptMenu() {
  todoForm.lifecycleScriptMenuOpen = !todoForm.lifecycleScriptMenuOpen
}

function closeTodoFormLifecycleScriptMenu() {
  todoForm.lifecycleScriptMenuOpen = false
}

function selectTodoFormLifecycleScript(index) {
  todoForm.selectedLifecycleScriptIndex = index === '' ? '' : String(index)
  closeTodoFormLifecycleScriptMenu()
}

async function uploadTodoInitializationFile(index, event) {
  const file = event?.target?.files?.[0]
  if (!file || !initializationFileManagement.files[index]) {
    return
  }
  initializationFileManagement.error = ''
  try {
    initializationFileManagement.files[index].fileName = file.name
    initializationFileManagement.files[index].content = await file.text()
  } catch (error) {
    initializationFileManagement.error = errorMessageFrom(error)
  } finally {
    if (event?.target) {
      event.target.value = ''
    }
  }
}

async function openTodoInitializationFileManagement() {
  closeGlobalManagementMenu()
  closeTerminalMenu()
  closeTerminalSettings()
  initializationFileManagement.visible = true
  initializationFileManagement.loading = true
  initializationFileManagement.error = ''
  initializationFileManagement.files = []
  try {
    initializationFileManagement.files = cloneTodoInitializationFiles(await LoadTodoInitializationFiles())
  } catch (error) {
    initializationFileManagement.error = errorMessageFrom(error)
  } finally {
    initializationFileManagement.loading = false
  }
}

function closeTodoInitializationFileManagement() {
  initializationFileManagement.visible = false
  initializationFileManagement.error = ''
}

async function saveTodoInitializationFileManagement() {
  initializationFileManagement.saving = true
  initializationFileManagement.error = ''
  try {
    await SaveTodoInitializationFiles(normalizedTodoInitializationFiles())
    closeTodoInitializationFileManagement()
  } catch (error) {
    initializationFileManagement.error = errorMessageFrom(error)
  } finally {
    initializationFileManagement.saving = false
  }
}

function addTodoLifecycleScript() {
  lifecycleScriptManagement.scripts.push({
    name: '',
    description: '',
    initScript: '',
    completeScript: '',
    defaultSelected: false
  })
}

function removeTodoLifecycleScript(index) {
  lifecycleScriptManagement.scripts.splice(index, 1)
}

function moveTodoLifecycleScript(index, direction) {
  const nextIndex = index + direction
  if (nextIndex < 0 || nextIndex >= lifecycleScriptManagement.scripts.length) {
    return
  }
  const [script] = lifecycleScriptManagement.scripts.splice(index, 1)
  lifecycleScriptManagement.scripts.splice(nextIndex, 0, script)
}

function setTodoLifecycleScriptDefault(index, selected) {
  if (!lifecycleScriptManagement.scripts[index]) {
    return
  }
  if (!selected) {
    lifecycleScriptManagement.scripts[index].defaultSelected = false
    return
  }
  lifecycleScriptManagement.scripts.forEach((script, scriptIndex) => {
    script.defaultSelected = scriptIndex === index
  })
}

function normalizedTodoLifecycleScripts() {
  return lifecycleScriptManagement.scripts.map((script) => ({
    name: (script.name || '').trim(),
    description: (script.description || '').trim(),
    initScript: (script.initScript || '').trim(),
    completeScript: (script.completeScript || '').trim(),
    defaultSelected: script.defaultSelected === true
  }))
}

async function openTodoLifecycleScriptManagement() {
  closeGlobalManagementMenu()
  closeTerminalMenu()
  closeTerminalSettings()
  lifecycleScriptManagement.visible = true
  lifecycleScriptManagement.loading = true
  lifecycleScriptManagement.error = ''
  lifecycleScriptManagement.scripts = []
  try {
    lifecycleScriptManagement.scripts = cloneTodoLifecycleScripts(await LoadTodoLifecycleScripts())
  } catch (error) {
    lifecycleScriptManagement.error = errorMessageFrom(error)
  } finally {
    lifecycleScriptManagement.loading = false
  }
}

function closeTodoLifecycleScriptManagement() {
  lifecycleScriptManagement.visible = false
  lifecycleScriptManagement.error = ''
}

async function saveTodoLifecycleScriptManagement() {
  lifecycleScriptManagement.saving = true
  lifecycleScriptManagement.error = ''
  try {
    await SaveTodoLifecycleScripts(normalizedTodoLifecycleScripts())
    closeTodoLifecycleScriptManagement()
  } catch (error) {
    lifecycleScriptManagement.error = errorMessageFrom(error)
  } finally {
    lifecycleScriptManagement.saving = false
  }
}

function addTerminalLaunchProfile() {
  settingsPanel.launchProfiles.push({ name: '', command: '', enabled: true, background: false })
}

function removeTerminalLaunchProfile(index) {
  settingsPanel.launchProfiles.splice(index, 1)
}

function moveTerminalLaunchProfile(index, direction) {
  const nextIndex = index + direction
  if (nextIndex < 0 || nextIndex >= settingsPanel.launchProfiles.length) {
    return
  }
  const [profile] = settingsPanel.launchProfiles.splice(index, 1)
  settingsPanel.launchProfiles.splice(nextIndex, 0, profile)
}

function normalizedLaunchProfiles() {
  return settingsPanel.launchProfiles.map((profile) => ({
    name: (profile.name || '').trim(),
    command: (profile.command || '').trim(),
    enabled: profile.enabled !== false,
    background: profile.background === true
  }))
}

function validateLaunchProfiles(profiles) {
  const names = new Set()
  for (const profile of profiles) {
    if (!profile.name) {
      return 'Launch profile name is required'
    }
    if (!profile.command) {
      return `Launch profile command is required for ${profile.name}`
    }
    const key = profile.name.toLowerCase()
    if (key === 'terminal') {
      return 'Terminal is reserved'
    }
    if (names.has(key)) {
      return `Launch profile name is duplicated: ${profile.name}`
    }
    names.add(key)
  }
  return ''
}

async function redetectTerminalShell() {
  settingsPanel.detecting = true
  settingsPanel.error = ''
  try {
    detectedTerminalShell.value = await DetectTerminalShell()
    settingsPanel.mode = 'detected'
  } catch (error) {
    settingsPanel.error = errorMessageFrom(error)
  } finally {
    settingsPanel.detecting = false
  }
}

async function saveTerminalSettings() {
  const source = settingsPanel.mode === 'manual' ? 'manual' : 'detected'
  const path = source === 'manual' ? settingsPanel.manualPath.trim() : terminalSettingsDetected.value?.path || ''
  if (!path) {
    settingsPanel.error = 'Choose a terminal shell path'
    return
  }
  const launchProfiles = normalizedLaunchProfiles()
  const launchProfileError = validateLaunchProfiles(launchProfiles)
  if (launchProfileError) {
    settingsPanel.error = launchProfileError
    return
  }
  settingsPanel.saving = true
  settingsPanel.error = ''
  try {
    await SaveTerminalShell(path, source)
    const profileState = await SaveTerminalLaunchProfiles(launchProfiles)
    applyTerminalSettings(await SaveTerminalTheme(settingsPanel.theme) || profileState)
    closeTerminalSettings()
  } catch (error) {
    settingsPanel.error = errorMessageFrom(error)
  } finally {
    settingsPanel.saving = false
  }
}

function shellDisplay(shell) {
  if (!shell) {
    return ''
  }
  return `${shell.displayName || 'shell'} ${shell.path || ''}`.trim()
}

function syncGitStatusForActiveProject(previousActiveProjectId = '', options = {}) {
  const context = activeGitStatusContext.value
  if (!context) {
    gitStatusRequestId += 1
    gitStatusInFlightContextKey = ''
    gitStatusInFlightRequestId = 0
    gitStatus.value = null
    gitStatusLoading.value = false
    gitStatusError.value = ''
    gitInitLoading.value = false
    return
  }
  if (!context.available) {
    gitStatusRequestId += 1
    gitStatusInFlightContextKey = ''
    gitStatusInFlightRequestId = 0
    gitStatus.value = withGitStatusContext({ projectId: context.projectId, isRepo: false, pathUnavailable: true }, context)
    gitStatusLoading.value = false
    gitStatusError.value = ''
    gitInitLoading.value = false
    if (context.type === 'todo-project') {
      clearTodoProjectGitBranch(context.todoProjectId)
    }
    return
  }
  if (options.refresh !== false && (options.force === true || context.key !== previousActiveProjectId)) {
    refreshProjectGitStatus({ dedupePending: options.dedupePending === true })
  }
}

async function refreshProjectGitStatus(options = {}) {
  const context = activeGitStatusContext.value
  if (!context || !context.available) {
    syncGitStatusForActiveProject(activeGitStatusContextKey())
    return
  }
  if (options.dedupePending === true && gitStatusInFlightContextKey === context.key) {
    return
  }
  const requestId = gitStatusRequestId + 1
  gitStatusRequestId = requestId
  gitStatusInFlightContextKey = context.key
  gitStatusInFlightRequestId = requestId
  gitStatusLoading.value = true
  gitStatusError.value = ''
  try {
    const status =
      context.type === 'todo-project'
        ? await GetTodoProjectGitStatus(context.todoProjectId)
        : context.type === 'todo'
          ? await GetTodoGitStatus(context.todoId)
          : await GetProjectGitStatus(context.projectId)
    if (requestId !== gitStatusRequestId || activeGitStatusContext.value?.key !== context.key) {
      return
    }
    gitStatus.value = withGitStatusContext(status, context)
    if (context.type === 'todo-project') {
      setTodoProjectGitBranch(context.todoProjectId, status)
    }
  } catch (error) {
    if (requestId !== gitStatusRequestId || activeGitStatusContext.value?.key !== context.key) {
      return
    }
    gitStatus.value = null
    gitStatusError.value = errorMessageFrom(error)
    if (context.type === 'todo-project') {
      clearTodoProjectGitBranch(context.todoProjectId)
    }
  } finally {
    if (gitStatusInFlightContextKey === context.key && gitStatusInFlightRequestId === requestId) {
      gitStatusInFlightContextKey = ''
      gitStatusInFlightRequestId = 0
    }
    if (requestId === gitStatusRequestId) {
      gitStatusLoading.value = false
    }
  }
}

function todoProjectCanLoadGitStatus(todoProject) {
  return Boolean(
    todoProject &&
      todoProject.available !== false &&
      todoProject.worktreeStatus === 'ready' &&
      normalizeProjectPath(todoProject.worktreePath)
  )
}

function branchFromGitStatus(status) {
  if (!status?.isRepo || status.pathUnavailable || status.gitUnavailable) {
    return ''
  }
  return (status.branch || '').trim()
}

function setTodoProjectGitBranch(todoProjectId, status) {
  if (!todoProjectId) {
    return
  }
  const branch = branchFromGitStatus(status)
  if (branch) {
    todoProjectGitBranches[todoProjectId] = branch
  } else {
    clearTodoProjectGitBranch(todoProjectId)
  }
}

function clearTodoProjectGitBranch(todoProjectId) {
  if (todoProjectId) {
    delete todoProjectGitBranches[todoProjectId]
  }
}

function pruneTodoProjectGitBranches() {
  const availableTodoProjectIds = new Set(
    todoProjects.value.filter((todoProject) => todoProjectCanLoadGitStatus(todoProject)).map((todoProject) => todoProject.id)
  )
  for (const todoProjectId of Object.keys(todoProjectGitBranches)) {
    if (!availableTodoProjectIds.has(todoProjectId)) {
      delete todoProjectGitBranches[todoProjectId]
    }
  }
}

async function refreshTodoProjectBranchStatus(todoProjectId, options = {}) {
  const todoProject = todoProjects.value.find((candidate) => candidate.id === todoProjectId)
  if (!todoProjectCanLoadGitStatus(todoProject)) {
    clearTodoProjectGitBranch(todoProjectId)
    return
  }
  if (options.dedupePending === true && todoProjectGitStatusInFlight.has(todoProjectId)) {
    return
  }
  const requestId = todoProjectGitStatusRequestId + 1
  todoProjectGitStatusRequestId = requestId
  todoProjectGitStatusInFlight.set(todoProjectId, requestId)
  try {
    const status = await GetTodoProjectGitStatus(todoProjectId)
    if (todoProjectGitStatusInFlight.get(todoProjectId) !== requestId) {
      return
    }
    const currentTodoProject = todoProjects.value.find((candidate) => candidate.id === todoProjectId)
    if (!todoProjectCanLoadGitStatus(currentTodoProject)) {
      clearTodoProjectGitBranch(todoProjectId)
      return
    }
    setTodoProjectGitBranch(todoProjectId, status)
  } catch (_error) {
    if (todoProjectGitStatusInFlight.get(todoProjectId) === requestId) {
      clearTodoProjectGitBranch(todoProjectId)
    }
  } finally {
    if (todoProjectGitStatusInFlight.get(todoProjectId) === requestId) {
      todoProjectGitStatusInFlight.delete(todoProjectId)
    }
  }
}

function refreshTodoProjectBranchStatuses(options = {}) {
  const activeTodoProjectBranchId =
    activeGitStatusContext.value?.type === 'todo-project' ? activeGitStatusContext.value.todoProjectId : ''
  for (const todoProject of todoProjects.value) {
    if (!todoProjectCanLoadGitStatus(todoProject)) {
      clearTodoProjectGitBranch(todoProject.id)
      continue
    }
    if (todoProject.id === activeTodoProjectBranchId) {
      continue
    }
    refreshTodoProjectBranchStatus(todoProject.id, { dedupePending: options.dedupePending === true })
  }
}

function refreshProjectGitStatusOnFocus() {
  if (!hasWorkspace.value) {
    return
  }
  const context = activeGitStatusContext.value
  const now = Date.now()
  if (
    context &&
    lastFocusGitRefreshContextKey === context.key &&
    now - lastFocusGitRefreshAt < focusGitRefreshDedupeMs
  ) {
    return
  }
  lastFocusGitRefreshContextKey = context?.key || ''
  lastFocusGitRefreshAt = now
  refreshProjectGitStatus({ dedupePending: true })
  refreshTodoProjectBranchStatuses({ dedupePending: true })
}

function handleTodoExpanded(todoId) {
  const todoProject = activeTodoProject.value
  if (!todoProject || todoProject.todoId !== todoId) {
    for (const candidate of todoProjectsForTodo(todoId)) {
      refreshTodoProjectBranchStatus(candidate.id, { dedupePending: true })
    }
    return
  }
  refreshProjectGitStatus({ dedupePending: true })
  for (const candidate of todoProjectsForTodo(todoId)) {
    if (candidate.id !== todoProject.id) {
      refreshTodoProjectBranchStatus(candidate.id, { dedupePending: true })
    }
  }
}

async function initializeActiveProjectGitRepository() {
  if (!showInitializeGitRepository.value || gitInitLoading.value) {
    return
  }
  const projectId = activeProject.value.id
  gitInitLoading.value = true
  errorMessage.value = ''
  try {
    await InitializeProjectGitRepository(projectId)
    if (activeProject.value?.id === projectId) {
      await refreshProjectGitStatus()
    }
  } catch (error) {
    showError(error)
  } finally {
    if (activeProject.value?.id === projectId) {
      gitInitLoading.value = false
    }
  }
}

function statusChip(id, tone, text) {
  return { id, tone, text }
}

function filteredProjects(projectList, query) {
  const normalizedQuery = normalizeSearch(query)
  if (!normalizedQuery) {
    return projectList
  }
  return projectList.filter((project) => {
    return [project.name, project.path].some((value) => normalizeSearch(value).includes(normalizedQuery))
  })
}

function todoProjectsForTodo(todoId) {
  if (!todoId) {
    return []
  }
  return todoProjects.value.filter((todoProject) => todoProject.todoId === todoId)
}

function todoProjectDisplayProject(todoProject) {
  if (!todoProject) {
    return null
  }
  const sourceProject = projects.value.find((project) => project.id === todoProject.projectId) || null
  if (todoProject.name || todoProject.path || todoProject.worktreePath || sourceProject) {
    const hasReadyWorktree = todoProject.worktreeStatus === 'ready' && normalizeProjectPath(todoProject.worktreePath)
    return {
      id: todoProject.projectId,
      name: todoProject.name || sourceProject?.name || 'Missing project',
      path: hasReadyWorktree ? todoProject.worktreePath : todoProject.path || sourceProject?.path || todoProject.projectId,
      available: todoProject.available !== false && (!todoProject.worktreeStatus || Boolean(hasReadyWorktree))
    }
  }
  return sourceProject
}

const activeGitStatusContext = computed(() => {
  const terminal = activeTerminal.value
  if (terminal && !terminal.workspaceTerminal && terminal.todoId && !terminal.todoProjectId) {
    return {
      type: 'todo',
      key: `todo:${terminal.todoId}`,
      todoId: terminal.todoId,
      projectId: '',
      available: true
    }
  }
  const todoProject = activeTodoProject.value
  if (todoProject) {
    const hasReadyWorktree =
      todoProject.worktreeStatus === 'ready' && normalizeProjectPath(todoProject.worktreePath)
    return {
      type: 'todo-project',
      key: `todo-project:${todoProject.id}`,
      projectId: todoProject.projectId,
      todoProjectId: todoProject.id,
      available: todoProject.available !== false && Boolean(hasReadyWorktree)
    }
  }
  const project = activeProject.value
  if (!project) {
    return null
  }
  return {
    type: 'project',
    key: `project:${project.id}`,
    projectId: project.id,
    available: project.available
  }
})

function activeGitStatusContextKey() {
  return activeGitStatusContext.value?.key || ''
}

function withGitStatusContext(status, context) {
  return {
    ...status,
    projectId: status?.projectId || context.projectId,
    contextKey: context.key
  }
}

function normalizeProjectPath(path) {
  return (path || '').trim()
}

function normalizeSearch(value) {
  return (value || '').trim().toLowerCase()
}

function normalizedTodoPriority(priority) {
  return todoPriorities.some((option) => option.value === priority) ? priority : 'medium'
}

function normalizedProjectBranches(branches, projectIds) {
  if (!branches || typeof branches !== 'object') {
    return {}
  }
  const allowed = new Set(projectIds || [])
  const result = {}
  for (const projectId of Object.keys(branches)) {
    if (!allowed.has(projectId)) {
      continue
    }
    const value = (branches[projectId] || '').trim()
    if (value) {
      result[projectId] = value
    }
  }
  return result
}

async function ensureProjectBranchesLoaded(projectId) {
  if (!projectId) {
    return
  }
  const loadState = projectBranchLoadStates[projectId]
  if (loadState === 'loading' || loadState === 'loaded' || loadState === 'failed') {
    return
  }
  projectBranchLoadStates[projectId] = 'loading'
  projectBranchOptions[projectId] = []
  try {
    const branches = await ListProjectBranches(projectId)
    projectBranchOptions[projectId] = Array.isArray(branches) ? branches : []
    projectBranchLoadStates[projectId] = 'loaded'
  } catch {
    projectBranchOptions[projectId] = []
    projectBranchLoadStates[projectId] = 'failed'
  }
}

function projectBranchPickerKey(scope, projectId) {
  return `${scope}:${projectId}`
}

function openProjectBranchPicker(scope, projectId, { resetQuery = true } = {}) {
  if (!projectId) {
    return
  }
  const key = projectBranchPickerKey(scope, projectId)
  if (resetQuery) {
    projectBranchPickerQueries[key] = ''
  }
  openProjectBranchPickerKey.value = key
  void ensureProjectBranchesLoaded(projectId)
}

function closeProjectBranchPicker() {
  openProjectBranchPickerKey.value = ''
}

function isProjectBranchPickerOpen(scope, projectId) {
  return openProjectBranchPickerKey.value === projectBranchPickerKey(scope, projectId)
}

function branchesForProject(projectId) {
  return projectBranchOptions[projectId] || []
}

function projectBranchPickerQuery(scope, projectId) {
  return projectBranchPickerQueries[projectBranchPickerKey(scope, projectId)] || ''
}

function matchingProjectBranches(scope, projectId) {
  const query = normalizeSearch(projectBranchPickerQuery(scope, projectId))
  const branches = branchesForProject(projectId)
  if (!query) {
    return branches
  }
  return branches.filter((branch) => normalizeSearch(branch).includes(query))
}

function visibleProjectBranches(scope, projectId) {
  return matchingProjectBranches(scope, projectId).slice(0, projectBranchCandidateLimit)
}

function projectBranchPickerStatus(scope, projectId) {
  const loadState = projectBranchLoadStates[projectId]
  if (loadState === 'loading') {
    return 'Loading suggestions...'
  }
  if (loadState === 'failed') {
    return 'Suggestions unavailable'
  }
  const matchCount = matchingProjectBranches(scope, projectId).length
  if (matchCount > projectBranchCandidateLimit) {
    return 'Keep typing to narrow branch suggestions'
  }
  if (loadState === 'loaded' && matchCount === 0) {
    return 'No matching branches'
  }
  return ''
}

function terminalWithAttention(terminal) {
  return {
    ...terminal,
    attentionState: terminalAckIds.has(terminal.id) ? 'needs-ack' : ''
  }
}

function terminalCanRestart(terminal) {
  if (!terminal) {
    return false
  }
  if (terminal.workspaceTerminal) {
    return hasWorkspace.value
  }
  if (terminal.todoId && !terminal.todoProjectId) {
    const todo = todos.value.find((candidate) => candidate.id === terminal.todoId)
    return todo?.status === 'in-progress'
  }
  return Boolean(activeTodoProjectProject.value?.available)
}

function importedSingleProjectId(state, previousProjectIds) {
  const addedProjects = Array.isArray(state?.importSummary?.added)
    ? state.importSummary.added.filter((project) => project?.id)
    : []
  if (addedProjects.length === 1) {
    return addedProjects[0].id
  }

  const skippedPaths = Array.isArray(state?.importSummary?.skippedPaths)
    ? state.importSummary.skippedPaths.map(normalizeProjectPath).filter(Boolean)
    : []
  if (skippedPaths.length === 1) {
    const skippedPath = skippedPaths[0]
    const existingProject = (state?.projects || []).find((project) => normalizeProjectPath(project.path) === skippedPath)
    if (existingProject?.id) {
      return existingProject.id
    }
  }

  const nextProjectIds = (state?.projects || []).map((project) => project.id).filter(Boolean)
  const newProjectIds = nextProjectIds.filter((projectId) => !previousProjectIds.has(projectId))
  return newProjectIds.length === 1 ? newProjectIds[0] : ''
}

function selectImportedProjectCandidate(projectId) {
  if (!projectId) {
    return
  }
  if (todoForm.visible) {
    todoForm.projectSelections = appendProjectSelection(todoForm.projectSelections, projectId)
  }
  if (todoDetail.visible && !todoDetail.readOnly) {
    todoDetail.projectSelections = appendProjectSelection(todoDetail.projectSelections, projectId)
  }
  if (projectPicker.visible) {
    projectPicker.projectSelections = appendProjectSelection(projectPicker.projectSelections, projectId)
  }
}

function defaultBaseBranch(projectId = '') {
  if (projectId && Object.prototype.hasOwnProperty.call(projectBranchPreferences.value, projectId)) {
    return projectBranchPreferences.value[projectId]?.baseBranch ?? ''
  }
  if (projectId && gitStatus.value?.projectId === projectId && gitStatus.value?.isRepo && gitStatus.value?.branch) {
    return gitStatus.value.branch === '(detached)' ? '' : gitStatus.value.branch
  }
  return ''
}

function appendProjectSelection(projectSelections, projectId) {
  if (projectSelections.some((selection) => selection.projectId === projectId)) {
    return projectSelections
  }
  return [...projectSelections, { projectId, baseBranch: defaultBaseBranch(projectId) }]
}

function hasProjectSelection(projectSelections, projectId) {
  return projectSelections.some((selection) => selection.projectId === projectId)
}

function toggleProjectSelection(projectSelections, projectId) {
  if (projectSelections.some((selection) => selection.projectId === projectId)) {
    return removeProjectSelection(projectSelections, projectId)
  }
  return appendProjectSelection(projectSelections, projectId)
}

function removeProjectSelection(projectSelections, projectId) {
  return projectSelections.filter((selection) => selection.projectId !== projectId)
}

function removePendingProjectSelection(projectId) {
  todoForm.projectSelections = removeProjectSelection(todoForm.projectSelections, projectId)
  todoDetail.projectSelections = removeProjectSelection(todoDetail.projectSelections, projectId)
  projectPicker.projectSelections = removeProjectSelection(projectPicker.projectSelections, projectId)
}

function selectedProjectBaseBranch(projectSelections, projectId) {
  return projectSelections.find((selection) => selection.projectId === projectId)?.baseBranch || ''
}

function updateProjectBaseBranch(projectSelections, projectId, baseBranch) {
  return projectSelections.map((selection) =>
    selection.projectId === projectId ? { ...selection, baseBranch } : selection
  )
}

function setTodoFormProjectBaseBranch(projectId, baseBranch) {
  todoForm.projectSelections = updateProjectBaseBranch(todoForm.projectSelections, projectId, baseBranch)
}

function setTodoDetailProjectBaseBranch(projectId, baseBranch) {
  todoDetail.projectSelections = updateProjectBaseBranch(todoDetail.projectSelections, projectId, baseBranch)
}

function setProjectPickerBaseBranch(projectId, baseBranch) {
  projectPicker.projectSelections = updateProjectBaseBranch(projectPicker.projectSelections, projectId, baseBranch)
}

function setProjectBaseBranchForScope(scope, projectId, baseBranch) {
  if (scope === 'todo-create') {
    setTodoFormProjectBaseBranch(projectId, baseBranch)
    return
  }
  if (scope === 'todo-detail') {
    setTodoDetailProjectBaseBranch(projectId, baseBranch)
    return
  }
  if (scope === 'project-picker') {
    setProjectPickerBaseBranch(projectId, baseBranch)
  }
}

function updateProjectBranchInput(scope, projectId, baseBranch) {
  projectBranchPickerQueries[projectBranchPickerKey(scope, projectId)] = baseBranch
  setProjectBaseBranchForScope(scope, projectId, baseBranch)
  openProjectBranchPicker(scope, projectId, { resetQuery: false })
}

function selectProjectBranchCandidate(scope, projectId, branch) {
  projectBranchPickerQueries[projectBranchPickerKey(scope, projectId)] = branch
  setProjectBaseBranchForScope(scope, projectId, branch)
  closeProjectBranchPicker()
}

function todoProjectSelectionPayload(projectSelections) {
  return projectSelections.map((selection) => ({
    projectId: selection.projectId,
    baseBranch: (selection.baseBranch || '').trim()
  }))
}

function completedTodosWithSnapshots() {
  return todos.value.filter((todo) => normalizeTodoStatus(todo) === 'completed' && Array.isArray(todo.projectSnapshots))
}

function normalizeTodoStatus(todo) {
  if (todo?.status === 'active') {
    return 'not-started'
  }
  return todo?.status || 'not-started'
}

function completedSnapshotKey(todo, snapshot, index) {
  return [todo?.id || '', snapshot?.projectId || '', snapshot?.path || '', index].join('::')
}

function completedSnapshotBranchLabel(snapshot) {
  const worktreeBranch = (snapshot?.worktreeBranch || '').trim() || 'Unknown branch'
  const baseBranch = (snapshot?.baseBranch || '').trim() || 'Unknown base'
  return `${worktreeBranch} -> ${baseBranch}`
}

function completedMergeStatusRequestFor(todo, snapshot, index) {
  return {
    id: completedSnapshotKey(todo, snapshot, index),
    path: (snapshot?.path || '').trim(),
    worktreeBranch: (snapshot?.worktreeBranch || '').trim(),
    baseBranch: (snapshot?.baseBranch || '').trim()
  }
}

function completedMergeStatusFingerprint(request) {
  return [request.path, request.worktreeBranch, request.baseBranch].join('::')
}

function completedMergeStatusEntries() {
  const entries = []
  for (const todo of completedTodosWithSnapshots()) {
    for (const [index, snapshot] of todo.projectSnapshots.entries()) {
      const request = completedMergeStatusRequestFor(todo, snapshot, index)
      entries.push({
        key: completedSnapshotKey(todo, snapshot, index),
        request,
        fingerprint: completedMergeStatusFingerprint(request)
      })
    }
  }
  return entries
}

function scheduleCompletedMergeStatusRefresh(options = {}) {
  nextTick(() => {
    void refreshCompletedMergeStatuses(options)
  })
}

async function refreshCompletedMergeStatuses(options = {}) {
  if (options.force === true) {
    completedMergeStatuses.value = {}
  }
  const entries = completedMergeStatusEntries()
  const entryKeys = new Set(entries.map((entry) => entry.key))
  const nextStatuses = {}
  const requests = []

  for (const entry of entries) {
    const existing = completedMergeStatuses.value[entry.key]
    if (existing?.fingerprint === entry.fingerprint) {
      nextStatuses[entry.key] = existing
      continue
    }
    if (!entry.request.path || !entry.request.worktreeBranch || !entry.request.baseBranch) {
      nextStatuses[entry.key] = {
        id: entry.key,
        status: 'unknown',
        reason: 'missing snapshot branch information',
        fingerprint: entry.fingerprint
      }
      continue
    }
    requests.push(entry.request)
  }

  for (const [key, value] of Object.entries(completedMergeStatuses.value)) {
    if (entryKeys.has(key) && nextStatuses[key] === undefined) {
      nextStatuses[key] = value
    }
  }

  completedMergeStatuses.value = nextStatuses

  if (!hasWorkspace.value || currentTodoView.value !== 'completed') {
    completedMergeStatusRequestGeneration += 1
    return
  }
  if (requests.length === 0) {
    return
  }

  const generation = ++completedMergeStatusRequestGeneration
  try {
    const statuses = await GetCompletedTodoProjectMergeStatuses(requests)
    if (generation !== completedMergeStatusRequestGeneration) {
      return
    }
    const latestEntries = completedMergeStatusEntries()
    const latestFingerprintsByKey = new Map(latestEntries.map((entry) => [entry.key, entry.fingerprint]))
    const mergedStatuses = {}
    for (const [key, value] of Object.entries(completedMergeStatuses.value)) {
      if (latestFingerprintsByKey.get(key) === value?.fingerprint) {
        mergedStatuses[key] = value
      }
    }
    for (const status of statuses || []) {
      const fingerprint = latestFingerprintsByKey.get(status?.id)
      if (status?.id && fingerprint) {
        mergedStatuses[status.id] = { ...status, fingerprint }
      }
    }
    completedMergeStatuses.value = mergedStatuses
  } catch (error) {
    if (generation === completedMergeStatusRequestGeneration) {
      showError(error)
    }
  }
}

function addCountChip(chips, id, count, label) {
  if (!count) {
    return
  }
  chips.push(statusChip(id, id === 'ahead' || id === 'behind' ? 'sync' : id, `${count} ${label}`))
}

function errorMessageFrom(error) {
  return error?.message || String(error)
}

function startSidebarResize(event) {
  sidebarResize.active = true
  sidebarResize.startX = event.clientX
  sidebarResize.startWidth = sidebarWidth.value
  window.addEventListener('mousemove', resizeSidebar)
  window.addEventListener('mouseup', stopSidebarResize)
  event.preventDefault()
}

function resizeSidebar(event) {
  if (!sidebarResize.active) {
    return
  }
  const nextWidth = sidebarResize.startWidth + event.clientX - sidebarResize.startX
  sidebarWidth.value = clampNumber(nextWidth, sidebarMinWidth, sidebarMaxWidth)
  scheduleFitActiveTerminal()
}

function stopSidebarResize() {
  if (!sidebarResize.active) {
    return
  }
  sidebarResize.active = false
  window.removeEventListener('mousemove', resizeSidebar)
  window.removeEventListener('mouseup', stopSidebarResize)
  persistTodoSidebarWidth()
  scheduleFitActiveTerminal()
}

function handleTodoViewChange(view) {
  const previousView = currentTodoView.value
  currentTodoView.value = normalizeTodoView(view)
  persistActiveTodoProjectUIState()
  scheduleCompletedMergeStatusRefresh({ force: currentTodoView.value === 'completed' && previousView !== 'completed' })
}

function applyTodoProjectUIState(todoProjectId) {
  const state = todoProjectId ? todoProjectUIStates.value[todoProjectId] : null
  currentTodoView.value = normalizeTodoView(state?.todoView)
  scheduleCompletedMergeStatusRefresh()
}

function applyTodoWorkspaceUIState(todoProjectId) {
  applyTodoProjectUIState(todoProjectId)
  sidebarWidth.value = clampNumber(todoSidebarWidthState.value || defaultSidebarWidth, sidebarMinWidth, sidebarMaxWidth)
  scheduleFitActiveTerminal()
}

function persistActiveTodoProjectUIState() {
  const todoProjectId = activeTodoProjectId.value
  if (!hasWorkspace.value || !todoProjectId) {
    return
  }
  const state = {
    todoView: normalizeTodoView(currentTodoView.value)
  }
  todoProjectUIStates.value = {
    ...todoProjectUIStates.value,
    [todoProjectId]: state
  }
  queueTodoProjectUIStateSave(todoProjectId, state)
}

function persistTodoSidebarWidth() {
  if (!hasWorkspace.value) {
    return
  }
  const width = clampNumber(sidebarWidth.value, sidebarMinWidth, sidebarMaxWidth)
  todoSidebarWidthState.value = width
  queueTodoSidebarWidthSave(width)
}

function queueTodoProjectUIStateSave(todoProjectId, state) {
  const queue = todoProjectUIStateSaveQueues.get(todoProjectId) || {
    saving: false,
    pending: null
  }
  queue.pending = state
  todoProjectUIStateSaveQueues.set(todoProjectId, queue)
  if (!queue.saving) {
    drainTodoProjectUIStateSaveQueue(todoProjectId, queue)
  }
}

async function drainTodoProjectUIStateSaveQueue(todoProjectId, queue) {
  queue.saving = true
  while (queue.pending) {
    const nextState = queue.pending
    queue.pending = null
    try {
      await SaveTodoProjectUIState(todoProjectId, nextState)
    } catch (error) {
      showError(error)
    }
  }
  queue.saving = false
  if (!queue.pending) {
    todoProjectUIStateSaveQueues.delete(todoProjectId)
  }
}

function queueTodoSidebarWidthSave(sidebarWidth) {
  todoSidebarWidthSaveQueue.pending = sidebarWidth
  if (!todoSidebarWidthSaveQueue.saving) {
    drainTodoSidebarWidthSaveQueue()
  }
}

async function drainTodoSidebarWidthSaveQueue() {
  todoSidebarWidthSaveQueue.saving = true
  while (todoSidebarWidthSaveQueue.pending !== null) {
    const nextSidebarWidth = todoSidebarWidthSaveQueue.pending
    todoSidebarWidthSaveQueue.pending = null
    try {
      await SaveTodoSidebarWidth(nextSidebarWidth)
    } catch (error) {
      showError(error)
    }
  }
  todoSidebarWidthSaveQueue.saving = false
}

function normalizeTodoView(view) {
  return ['not-started', 'in-progress', 'completed'].includes(view) ? view : 'not-started'
}

function clampNumber(value, min, max) {
  return Math.min(max, Math.max(min, Math.round(value)))
}

function scheduleFitActiveTerminal() {
  nextTick(() => fitActiveTerminal())
}

function fitActiveTerminal() {
  terminalManager.fitActive()
}

function terminalState(terminalId) {
  return shellStatuses[terminalId] || terminals.value.find((terminal) => terminal.id === terminalId)?.state || ''
}

function updateTerminalState(terminalId, state) {
  if (!terminalId) {
    return
  }
  shellStatuses[terminalId] = state
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (terminal) {
    applyTerminalAgentEvent(terminal, {
      type: 'shell-status',
      state,
      at: Date.now()
    })
    if (state !== 'running') {
      clearTitleActivityTimer(terminalId)
    }
  }
}

function handleTerminalCommandState(terminalId, event) {
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (!terminal) {
    return
  }
  if (event.type === 'command-start') {
    clearTitleActivityTimer(terminalId)
    applyTerminalAgentEvent(terminal, {
      type: 'command-state',
      commandType: 'command-start',
      command: event.command,
      at: Date.now()
    })
  }
  if (event.type === 'command-end') {
    clearTitleActivityTimer(terminalId)
    applyTerminalAgentEvent(terminal, {
      type: 'command-state',
      commandType: 'command-end',
      at: Date.now()
    })
    if (terminal.id === activeTerminalId.value && !terminal.workspaceTerminal) {
      refreshProjectGitStatus()
    } else if (terminal.todoProjectId) {
      refreshTodoProjectBranchStatus(terminal.todoProjectId, { dedupePending: true })
    }
  }
}

function handleTerminalTitleChange(terminalId, title) {
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (!terminal) {
    return
  }
  const at = Date.now()
  applyTerminalAgentEvent(terminal, {
    type: 'title',
    title,
    at
  })
  markTerminalTitleActivity(terminal, at)
}

function handleTerminalAgentStatus(event) {
  const terminal = terminals.value.find((candidate) => candidate.id === event?.terminalId)
  if (!terminal) {
    return
  }
  applyTerminalAgentEvent(terminal, {
    type: 'agent-status',
    phase: event.phase,
    source: event.source,
    confidence: event.confidence,
    reason: event.reason,
    label: event.label,
    at: event.updatedAt || Date.now()
  })
}

function applyTerminalAgentEvent(terminal, event) {
  const previousActivityState = visibleTerminalActivityState(terminal)
  Object.assign(terminal, applyAgentStatusEvent(terminal, event))
  updateTerminalAckState(terminal, previousActivityState)
}

function visibleTerminalActivityState(terminal) {
  const state = terminal?.activityState || terminal?.agentStatus?.phase || AGENT_PHASE.IDLE
  return [AGENT_PHASE.BUSY, AGENT_PHASE.NEEDS_INPUT].includes(state) ? state : AGENT_PHASE.IDLE
}

function updateTerminalAckState(terminal, previousActivityState) {
  const nextActivityState = visibleTerminalActivityState(terminal)
  if (nextActivityState === AGENT_PHASE.BUSY) {
    terminalAckIds.delete(terminal.id)
    return
  }
  if (
    previousActivityState === AGENT_PHASE.BUSY &&
    terminal.id !== activeTerminalId.value
  ) {
    terminalAckIds.add(terminal.id)
  }
}

function markTerminalTitleActivity(terminal, at = Date.now()) {
  if (!terminal?.id || terminal.state !== 'running') {
    return
  }
  applyTerminalAgentEvent(terminal, {
    type: 'agent-status',
    phase: AGENT_PHASE.BUSY,
    source: AGENT_SOURCE.TITLE_FALLBACK,
    confidence: AGENT_CONFIDENCE.HEURISTIC,
    reason: 'title-changed',
    label: terminal.runtimeTitle,
    at
  })
  restartTitleActivityTimer(terminal.id)
}

function restartTitleActivityTimer(terminalId) {
  clearTitleActivityTimer(terminalId)
  titleActivityTimers.set(terminalId, setTimeout(() => {
    const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
    if (!terminal || terminal.state !== 'running') {
      titleActivityTimers.delete(terminalId)
      return
    }
    applyTerminalAgentEvent(terminal, {
      type: 'agent-status',
      phase: AGENT_PHASE.IDLE,
      source: AGENT_SOURCE.TITLE_FALLBACK,
      confidence: AGENT_CONFIDENCE.HEURISTIC,
      reason: 'title-unchanged',
      label: terminal.runtimeTitle,
      at: Date.now()
    })
    titleActivityTimers.delete(terminalId)
  }, 1000))
}

function clearTitleActivityTimer(terminalId) {
  const timer = titleActivityTimers.get(terminalId)
  if (timer) {
    clearTimeout(timer)
    titleActivityTimers.delete(terminalId)
  }
}

function clearAllTitleActivityTimers() {
  for (const terminalId of titleActivityTimers.keys()) {
    clearTitleActivityTimer(terminalId)
  }
}

function sanitizeCommandLabel(command) {
  return (command || '').replace(/\s+/g, ' ').trim().slice(0, 120)
}

function globalTerminalLabel(terminal) {
  return sanitizeCommandLabel(terminal.currentCommand) || terminal.runtimeTitle || terminal.shellName || 'Terminal'
}

function exitedAgentStatus() {
  return createAgentStatus({
    phase: AGENT_PHASE.EXITED,
    source: AGENT_SOURCE.SHELL,
    confidence: AGENT_CONFIDENCE.STRUCTURED,
    reason: 'shell-exited'
  })
}

function showError(error) {
  errorMessage.value = errorMessageFrom(error)
}

function showToast(message) {
  toastMessage.value = message
  clearToastTimer()
  toastTimer = window.setTimeout(() => {
    toastMessage.value = ''
    toastTimer = null
  }, toastDurationMs)
}

function clearToastTimer() {
  if (!toastTimer) {
    return
  }
  window.clearTimeout(toastTimer)
  toastTimer = null
}
</script>

<template>
  <main class="app-shell" :data-theme="currentTheme" :style="{ '--sidebar-width': `${sidebarWidth}px` }">
    <ProjectSidebar
      :projects="projects"
      :todos="todos"
      :todo-projects="todoProjects"
      :todo-project-branches="todoProjectGitBranches"
      :terminals="sidebarTerminals"
      :active-project-id="activeProjectId"
      :active-todo-id="activeTodoId"
      :active-todo-project-id="activeTodoProjectId"
      :active-terminal-id="activeTerminalId"
      :launch-profiles="terminalLaunchProfiles"
      :import-summary="importSummary"
      :has-workspace="hasWorkspace"
      :todo-view="currentTodoView"
      :completed-merge-statuses="completedMergeStatuses"
      :lifecycle-script-statuses="lifecycleScriptStatuses"
      @create-project="createProject"
      @import-projects="importProjectsFromParentDirectory"
      @select-project="selectProject"
      @create-todo="createTodo"
      @edit-todo="editTodo"
      @add-project-to-todo="addProjectToTodo"
      @select-todo-project="selectTodoProject"
      @todo-expanded="handleTodoExpanded"
      @remove-todo-project="removeTodoProject"
      @change-todo-status="changeTodoStatus"
      @complete-todo="completeTodo"
      @retry-todo-lifecycle-script="retryTodoLifecycleScript"
      @copy-todo-description="copyTodoDescription"
      @delete-todo="deleteTodo"
      @delete-completed-todos="deleteCompletedTodos"
      @create-task-terminal="createTaskTerminal"
      @create-terminal="createTerminal"
      @select-terminal="selectTerminal"
      @delete-project="deleteProject"
      @delete-projects="deleteProjects"
      @delete-terminal="deleteTerminal"
      @open-todo-folder="openTodoFolder"
      @todo-view-change="handleTodoViewChange"
    />

    <div
      class="sidebar-resize-handle"
      data-testid="sidebar-resize-handle"
      role="separator"
      aria-orientation="vertical"
      title="Resize sidebar"
      @mousedown="startSidebarResize"
    ></div>

    <div v-if="toastMessage" class="app-toast" data-testid="app-toast" role="status">
      {{ toastMessage }}
    </div>

    <div
      v-if="gitInitializationPrompt.visible"
      class="git-init-confirm-overlay"
      data-testid="git-init-confirm-overlay"
      @click="resolveGitInitializationPrompt(false)"
    >
      <section
        class="git-init-confirm-dialog"
        data-testid="git-init-confirm-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="git-init-confirm-title"
        @click.stop
      >
        <header class="git-init-confirm-header">
          <h2 id="git-init-confirm-title">初始化 Git 仓库</h2>
          <p>所选目录不是 Git 仓库。初始化后再导入项目？</p>
        </header>
        <div class="git-init-confirm-path" data-testid="git-init-confirm-path">
          {{ gitInitializationPrompt.path }}
        </div>
        <footer class="git-init-confirm-actions">
          <button
            type="button"
            class="toolbar-button"
            data-testid="git-init-confirm-cancel"
            @click="resolveGitInitializationPrompt(false)"
          >
            取消
          </button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="git-init-confirm-submit"
            @click="resolveGitInitializationPrompt(true)"
          >
            初始化并导入
          </button>
        </footer>
      </section>
    </div>

    <div
      v-if="projectCandidateClearPrompt.visible"
      class="git-init-confirm-overlay project-candidate-clear-overlay"
      data-testid="project-candidate-clear-overlay"
      @click="closeProjectCandidateClearPrompt"
    >
      <section
        class="git-init-confirm-dialog project-candidate-clear-dialog"
        data-testid="project-candidate-clear-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="project-candidate-clear-title"
        @click.stop
      >
        <header class="git-init-confirm-header">
          <h2 id="project-candidate-clear-title">Clear Project Candidate</h2>
          <p>This removes the candidate record only. Existing TODO projects and terminals stay unchanged.</p>
        </header>
        <div class="git-init-confirm-path project-candidate-clear-target">
          <strong data-testid="project-candidate-clear-name">{{ projectCandidateClearLabel() }}</strong>
          <span v-if="projectCandidateClearPath()" data-testid="project-candidate-clear-path">
            {{ projectCandidateClearPath() }}
          </span>
        </div>
        <footer class="git-init-confirm-actions">
          <button
            type="button"
            class="toolbar-button"
            data-testid="project-candidate-clear-cancel"
            :disabled="projectCandidateClearPrompt.clearing"
            @click="closeProjectCandidateClearPrompt"
          >
            Cancel
          </button>
          <button
            type="button"
            class="toolbar-button danger"
            data-testid="project-candidate-clear-confirm"
            :disabled="projectCandidateClearPrompt.clearing"
            @click="confirmProjectCandidateClear"
          >
            <Trash2 :size="14" />
            <span>Clear</span>
          </button>
        </footer>
      </section>
    </div>

    <section class="workspace">
      <header class="workspace-header">
        <div v-if="!hasWorkspace" class="project-heading muted">Open a project</div>
        <div v-else-if="activeTodoProject && activeTodoProjectProject" class="project-heading">
          <span class="heading-name">{{ activeTodo?.title || 'TODO' }} / {{ activeTodoProjectProject.name }}</span>
          <span class="heading-path">{{ activeTodoProjectProject.path }}</span>
        </div>
        <div v-else-if="activeProject" class="project-heading">
          <span class="heading-name">{{ activeProject.name }}</span>
          <span class="heading-path">{{ activeProject.path }}</span>
        </div>
        <div v-else class="project-heading muted">No TODO project selected</div>
        <div class="workspace-actions">
          <button
            type="button"
            class="toolbar-button"
            data-testid="create-global-terminal"
            title="New global terminal"
            :disabled="!hasWorkspace"
            @click="createWorkspaceTerminal"
          >
            <Plus :size="16" />
            <span>Global terminal</span>
          </button>
          <div class="global-management-control" @click.stop>
            <button
              type="button"
              class="toolbar-button"
              data-testid="global-management-toggle"
              title="Global management"
              @click="toggleGlobalManagementMenu"
            >
              <FileText :size="16" />
              <span>全局管理</span>
              <ChevronDown :size="14" />
            </button>
            <div
              v-if="globalManagementMenu.visible"
              class="global-management-menu"
              data-testid="global-management-menu"
            >
              <button
                type="button"
                data-testid="global-file-management"
                @click="openTodoInitializationFileManagement"
              >
                <FileText :size="14" />
                <span>文件管理</span>
              </button>
              <button
                type="button"
                data-testid="global-script-management"
                @click="openTodoLifecycleScriptManagement"
              >
                <FileText :size="14" />
                <span>脚本管理</span>
              </button>
            </div>
          </div>
          <button
            type="button"
            class="toolbar-button"
            data-testid="settings-toggle"
            title="Settings"
            @click="openTerminalSettings"
          >
            <Settings :size="16" />
            <span>Settings</span>
          </button>
          <button
            v-if="canRestartActiveShell"
            type="button"
            class="toolbar-button"
            title="Restart shell"
            @click="restartActiveShell"
          >
            <RotateCcw :size="16" />
            <span>Restart</span>
          </button>
        </div>
      </header>

      <div
        class="terminal-surface"
        :class="{ 'has-global-terminals': workspaceTerminals.length > 0 }"
        data-testid="terminal-surface"
      >
        <div
          v-if="workspaceTerminals.length"
          class="global-terminal-group"
          data-testid="global-terminal-group"
        >
          <span class="global-terminal-group-label">Global</span>
          <div class="global-terminal-tabs">
            <div
              v-for="terminal in workspaceTerminals"
              :key="terminal.id"
              class="global-terminal-tab"
              :class="{ active: terminal.id === activeTerminalId }"
              :data-testid="`global-terminal-${terminal.id}`"
              :data-activity-state="terminal.attentionState || terminal.activityState || null"
              role="button"
              tabindex="0"
              @click="selectTerminal(terminal.id)"
              @keydown.enter.prevent="selectTerminal(terminal.id)"
              @keydown.space.prevent="selectTerminal(terminal.id)"
            >
              <span>{{ globalTerminalLabel(terminal) }}</span>
              <button
                type="button"
                class="global-terminal-delete"
                :data-testid="`delete-global-terminal-${terminal.id}`"
                :title="`Delete ${globalTerminalLabel(terminal)}`"
                @click.stop="deleteTerminal(terminal.id)"
              >
                <X :size="12" />
              </button>
            </div>
            <button
              type="button"
              class="global-terminal-create"
              data-testid="create-global-terminal-from-group"
              title="New global terminal"
              @click="createWorkspaceTerminal"
            >
              <Plus :size="14" />
            </button>
          </div>
        </div>

        <div
          v-for="terminal in terminals"
          :key="terminal.id"
          class="terminal-pane"
          :class="{ active: terminal.id === activeTerminalId }"
          :data-testid="`terminal-pane-${terminal.id}`"
          :ref="(element) => setTerminalContainer(terminal.id, element)"
          @contextmenu.prevent="openTerminalMenu(terminal.id, $event)"
        />

        <div
          v-if="terminalMenu.visible"
          class="terminal-context-menu"
          data-testid="terminal-context-menu"
          :style="{ left: `${terminalMenu.x}px`, top: `${terminalMenu.y}px` }"
          @click.stop
        >
          <button
            type="button"
            data-testid="terminal-menu-copy"
            :disabled="!hasTerminalSelection(terminalMenu.terminalId)"
            @click="copyFromTerminalMenu"
          >
            Copy
          </button>
          <button type="button" data-testid="terminal-menu-paste" @click="pasteFromTerminalMenu">
            Paste
          </button>
        </div>

        <div
          v-if="terminalStateLayer"
          class="state-layer"
          :class="{ warning: terminalStateLayer.warning }"
          :data-testid="terminalStateLayer.testId || null"
        >
          {{ terminalStateLayer.text }}
        </div>
      </div>

      <footer class="status-bar">
        <div class="status-item status-cluster" data-testid="project-git-status">
          <span
            v-for="chip in projectGitStatusChips"
            :key="chip.id"
            class="status-chip"
            :class="`status-chip-${chip.tone}`"
            :data-testid="`status-chip-${chip.id}`"
          >
            {{ chip.text }}
          </span>
          <button
            v-if="showInitializeGitRepository"
            type="button"
            class="status-action"
            data-testid="initialize-git-repository"
            :disabled="gitInitLoading"
            @click="initializeActiveProjectGitRepository"
          >
            <GitBranch :size="14" />
            <span>{{ gitInitializeButtonText }}</span>
          </button>
        </div>
        <div v-if="errorMessage" class="status-error">{{ errorMessage }}</div>
      </footer>
    </section>

    <div v-if="todoForm.visible" class="settings-overlay">
      <section class="settings-dialog todo-dialog" data-testid="todo-create-dialog" @click.stop>
        <header class="settings-header">
          <div>
            <h2>New TODO</h2>
            <p>Task context</p>
          </div>
          <button
            type="button"
            class="icon-button"
            data-testid="todo-create-close"
            title="Close TODO form"
            @click="closeTodoForm"
          >
            <X :size="16" />
          </button>
        </header>

        <div class="settings-body todo-form-body">
          <label class="settings-field">
            <span class="settings-label">Name</span>
            <input
              v-model="todoForm.title"
              type="text"
              class="todo-text-input"
              data-testid="todo-name-input"
            />
          </label>

          <label class="settings-field">
            <span class="settings-label">Description</span>
            <textarea
              v-model="todoForm.description"
              class="todo-textarea"
              rows="3"
              data-testid="todo-description-input"
            ></textarea>
          </label>

          <div class="settings-field">
            <span class="settings-label">Priority</span>
            <div class="todo-priority-options">
              <label
                v-for="priority in todoPriorities"
                :key="priority.value"
                class="todo-priority-option"
                :class="`todo-priority-option-${priority.value}`"
              >
                <input
                  v-model="todoForm.priority"
                  type="radio"
                  :value="priority.value"
                  :data-testid="`todo-priority-${priority.value}`"
                />
                <span>{{ priority.label }}</span>
              </label>
            </div>
          </div>

          <div v-if="todoForm.initializationFiles.length" class="settings-field">
            <span class="settings-label">Initialization files</span>
            <div class="todo-initialization-file-options" data-testid="todo-initialization-files">
              <label
                v-for="(file, index) in todoForm.initializationFiles"
                :key="`${file.fileName}-${index}`"
                class="settings-option todo-initialization-file-option"
                :data-testid="`todo-initialization-file-${index}`"
              >
                <input
                  v-model="file.selected"
                  type="checkbox"
                  :data-testid="`todo-initialization-file-selected-${index}`"
                />
                <span>
                  <span class="settings-label">{{ file.name }}</span>
                  <strong>{{ file.fileName }}</strong>
                  <small>{{ file.description }}</small>
                </span>
              </label>
            </div>
          </div>

          <div v-if="todoForm.lifecycleScripts.length" class="settings-field">
            <span class="settings-label">
              Lifecycle script
              <span class="field-optional">Optional</span>
            </span>
            <div class="todo-lifecycle-script-dropdown" @click.stop>
              <button
                type="button"
                class="todo-text-input todo-lifecycle-script-select"
                data-testid="todo-lifecycle-script-select"
                aria-haspopup="listbox"
                :aria-expanded="todoForm.lifecycleScriptMenuOpen"
                @click="toggleTodoFormLifecycleScriptMenu"
                @keydown.escape.stop.prevent="closeTodoFormLifecycleScriptMenu"
              >
                <span>{{ selectedTodoFormLifecycleScriptLabel }}</span>
                <ChevronDown :size="14" />
              </button>
              <div
                v-if="todoForm.lifecycleScriptMenuOpen"
                class="todo-lifecycle-script-menu"
                data-testid="todo-lifecycle-script-menu"
                role="listbox"
              >
                <button
                  type="button"
                  class="todo-lifecycle-script-option"
                  data-testid="todo-lifecycle-script-option-none"
                  role="option"
                  :aria-selected="todoForm.selectedLifecycleScriptIndex === ''"
                  @click="selectTodoFormLifecycleScript('')"
                >
                  不选择脚本
                </button>
                <button
                  v-for="script in todoFormLifecycleScriptOptions"
                  :key="script.index"
                  type="button"
                  class="todo-lifecycle-script-option"
                  :class="{ selected: todoForm.selectedLifecycleScriptIndex === String(script.index) }"
                  :data-testid="`todo-lifecycle-script-option-${script.index}`"
                  role="option"
                  :aria-selected="todoForm.selectedLifecycleScriptIndex === String(script.index)"
                  @click="selectTodoFormLifecycleScript(script.index)"
                >
                  {{ lifecycleScriptOptionLabel(script) }}
                </button>
              </div>
            </div>
          </div>

          <div class="settings-field">
            <span class="settings-label">
              Projects
              <span class="field-optional" data-testid="todo-projects-optional">Optional</span>
            </span>
            <div class="candidate-management-toolbar" data-testid="todo-project-candidate-tools">
              <button
                type="button"
                class="toolbar-button compact"
                data-testid="import-single-project-candidate"
                :disabled="todoForm.saving"
                @click="importSingleProjectCandidate"
              >
                <FolderPlus :size="14" />
                <span>Import project</span>
              </button>
              <button
                type="button"
                class="toolbar-button compact"
                data-testid="import-global-project-candidates"
                :title="bulkGitImportTooltip"
                :disabled="todoForm.saving"
                @click="importProjectsFromParentDirectory"
              >
                <FolderInput :size="14" />
                <span>Import parent</span>
              </button>
              <button
                type="button"
                class="toolbar-button compact danger"
                data-testid="clear-global-project-candidates"
                :disabled="todoForm.saving || projects.length === 0"
                @click="clearGlobalProjectCandidates"
              >
                <Trash2 :size="14" />
                <span>Clear candidates</span>
              </button>
            </div>
            <div v-if="importSummary" class="import-summary" data-testid="import-summary">
              <span>{{ importSummary.addedCount || 0 }} imported</span>
              <span>{{ importSummary.skippedCount || 0 }} skipped</span>
            </div>
            <div
              v-if="selectedTodoFormProjects.length"
              class="todo-selected-project-tags"
              data-testid="todo-selected-project"
            >
              <span
                v-for="project in selectedTodoFormProjects"
                :key="project.id"
                class="todo-selected-project-tag"
                :data-testid="`todo-selected-project-tag-${project.id}`"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="todo-branch-picker">
                  <input
                    :value="selectedProjectBaseBranch(todoForm.projectSelections, project.id)"
                    type="text"
                    class="todo-branch-input"
                    placeholder="base 分支 (留空使用默认)"
                    :data-testid="`todo-selected-project-branch-${project.id}`"
                    aria-label="Base branch"
                    @focus="openProjectBranchPicker('todo-create', project.id)"
                    @input="updateProjectBranchInput('todo-create', project.id, $event.target.value)"
                    @keydown.escape.stop.prevent="closeProjectBranchPicker"
                  />
                  <div
                    v-if="isProjectBranchPickerOpen('todo-create', project.id)"
                    class="project-branch-picker-menu"
                    :data-testid="`project-branch-picker-options-todo-create-${project.id}`"
                    role="listbox"
                  >
                    <button
                      v-for="branch in visibleProjectBranches('todo-create', project.id)"
                      :key="branch"
                      type="button"
                      class="project-branch-picker-option"
                      :data-testid="`project-branch-picker-option-todo-create-${project.id}`"
                      role="option"
                      @mousedown.prevent
                      @click="selectProjectBranchCandidate('todo-create', project.id, branch)"
                    >
                      {{ branch }}
                    </button>
                    <div
                      v-if="projectBranchPickerStatus('todo-create', project.id)"
                      class="project-branch-picker-status"
                      :data-testid="`project-branch-picker-status-todo-create-${project.id}`"
                    >
                      {{ projectBranchPickerStatus('todo-create', project.id) }}
                    </div>
                  </div>
                </span>
                <button
                  type="button"
                  class="todo-selected-project-remove"
                  :title="`Remove ${project.name}`"
                  :aria-label="`Remove ${project.name}`"
                  :data-testid="`todo-selected-project-remove-${project.id}`"
                  @click="removeTodoFormProject(project.id)"
                >
                  <X :size="12" />
                </button>
              </span>
            </div>
            <input
              v-model="todoForm.projectSearch"
              type="text"
              class="todo-text-input"
              data-testid="todo-project-filter"
            />
            <div class="todo-project-options" data-testid="todo-project-options">
              <div
                v-for="project in todoFormProjectOptions"
                :key="project.id"
                class="todo-project-option-row"
              >
                <button
                  type="button"
                  class="todo-project-option"
                  :class="{ selected: todoForm.projectSelections.some((selection) => selection.projectId === project.id) }"
                  :data-testid="`todo-project-option-${project.id}`"
                  :aria-pressed="todoForm.projectSelections.some((selection) => selection.projectId === project.id)"
                  @click="toggleTodoFormProject(project)"
                >
                  <span class="project-name">{{ project.name }}</span>
                  <span class="project-path">{{ project.path }}</span>
                </button>
                <button
                  type="button"
                  class="todo-project-option-clear"
                  :title="`Clear project candidate ${project.name}`"
                  :aria-label="`Clear project candidate ${project.name}`"
                  :data-testid="`clear-project-candidate-${project.id}`"
                  :disabled="todoForm.saving"
                  @click.stop="clearProjectCandidate(project)"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
              <span v-if="todoFormProjectOptions.length === 0" class="sidebar-empty">No projects selected</span>
            </div>
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeTodoForm">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="todo-create-submit"
            :disabled="todoForm.saving"
            @click="submitTodoForm"
          >
            Create
          </button>
        </footer>
      </section>
    </div>

    <div v-if="todoDetail.visible" class="settings-overlay">
      <section class="settings-dialog todo-dialog" data-testid="todo-detail-dialog" @click.stop>
        <header class="settings-header">
          <div>
            <h2>TODO Detail</h2>
            <p>Task context</p>
          </div>
          <button
            type="button"
            class="icon-button"
            data-testid="todo-detail-close"
            title="Close TODO detail"
            @click="closeTodoDetail"
          >
            <X :size="16" />
          </button>
        </header>

        <div class="settings-body todo-form-body">
          <label class="settings-field">
            <span class="settings-label">Name</span>
            <input
              v-model="todoDetail.title"
              type="text"
              class="todo-text-input"
              :readonly="todoDetail.readOnly"
              data-testid="todo-detail-name-input"
            />
          </label>

          <label class="settings-field">
            <span class="settings-label">Description</span>
            <textarea
              v-model="todoDetail.description"
              class="todo-textarea"
              rows="3"
              :readonly="todoDetail.readOnly"
              data-testid="todo-detail-description-input"
            ></textarea>
          </label>

          <div class="settings-field">
            <span class="settings-label">Priority</span>
            <div class="todo-priority-options">
              <label
                v-for="priority in todoPriorities"
                :key="priority.value"
                class="todo-priority-option"
                :class="`todo-priority-option-${priority.value}`"
              >
                <input
                  v-model="todoDetail.priority"
                  type="radio"
                  :value="priority.value"
                  :disabled="todoDetail.readOnly"
                  :data-testid="`todo-detail-priority-${priority.value}`"
                />
                <span>{{ priority.label }}</span>
              </label>
            </div>
          </div>

          <div class="settings-field">
            <span class="settings-label">Projects</span>
            <div
              v-if="todoDetail.readOnly && todoDetail.projectSnapshots.length"
              class="todo-selected-project-tags"
              data-testid="todo-detail-project-snapshots"
            >
              <span
                v-for="snapshot in todoDetail.projectSnapshots"
                :key="`${snapshot.projectId}-${snapshot.path}`"
                class="todo-selected-project-tag todo-project-snapshot"
                :data-testid="`todo-detail-project-snapshot-${snapshot.projectId}`"
              >
                <span class="project-name">{{ snapshot.name }}</span>
                <span class="project-path">{{ completedSnapshotBranchLabel(snapshot) }}</span>
              </span>
            </div>
            <span
              v-else-if="todoDetail.readOnly"
              class="sidebar-empty"
              data-testid="todo-detail-project-snapshots-empty"
            >
              No completed project snapshots
            </span>
            <div
              v-if="!todoDetail.readOnly"
              class="candidate-management-toolbar"
              data-testid="todo-detail-project-candidate-tools"
            >
              <button
                type="button"
                class="toolbar-button compact"
                data-testid="import-single-project-candidate"
                :disabled="todoDetail.saving"
                @click="importSingleProjectCandidate"
              >
                <FolderPlus :size="14" />
                <span>Import project</span>
              </button>
              <button
                type="button"
                class="toolbar-button compact"
                data-testid="import-global-project-candidates"
                :title="bulkGitImportTooltip"
                :disabled="todoDetail.saving"
                @click="importProjectsFromParentDirectory"
              >
                <FolderInput :size="14" />
                <span>Import parent</span>
              </button>
              <button
                type="button"
                class="toolbar-button compact danger"
                data-testid="clear-global-project-candidates"
                :disabled="todoDetail.saving || projects.length === 0"
                @click="clearGlobalProjectCandidates"
              >
                <Trash2 :size="14" />
                <span>Clear candidates</span>
              </button>
            </div>
            <div v-if="!todoDetail.readOnly && importSummary" class="import-summary" data-testid="import-summary">
              <span>{{ importSummary.addedCount || 0 }} imported</span>
              <span>{{ importSummary.skippedCount || 0 }} skipped</span>
            </div>
            <div
              v-if="!todoDetail.readOnly && selectedTodoDetailProjects.length"
              class="todo-selected-project-tags"
              data-testid="todo-detail-selected-projects"
            >
              <span
                v-for="project in selectedTodoDetailProjects"
                :key="project.id"
                class="todo-selected-project-tag"
                :data-testid="`todo-detail-selected-project-tag-${project.id}`"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="todo-branch-picker">
                  <input
                    :value="selectedProjectBaseBranch(todoDetail.projectSelections, project.id)"
                    type="text"
                    class="todo-branch-input"
                    placeholder="base 分支 (留空使用默认)"
                    :disabled="todoDetail.saving"
                    :data-testid="`todo-detail-selected-project-branch-${project.id}`"
                    aria-label="Base branch"
                    @focus="openProjectBranchPicker('todo-detail', project.id)"
                    @input="updateProjectBranchInput('todo-detail', project.id, $event.target.value)"
                    @keydown.escape.stop.prevent="closeProjectBranchPicker"
                  />
                  <div
                    v-if="isProjectBranchPickerOpen('todo-detail', project.id)"
                    class="project-branch-picker-menu"
                    :data-testid="`project-branch-picker-options-todo-detail-${project.id}`"
                    role="listbox"
                  >
                    <button
                      v-for="branch in visibleProjectBranches('todo-detail', project.id)"
                      :key="branch"
                      type="button"
                      class="project-branch-picker-option"
                      :data-testid="`project-branch-picker-option-todo-detail-${project.id}`"
                      role="option"
                      @mousedown.prevent
                      @click="selectProjectBranchCandidate('todo-detail', project.id, branch)"
                    >
                      {{ branch }}
                    </button>
                    <div
                      v-if="projectBranchPickerStatus('todo-detail', project.id)"
                      class="project-branch-picker-status"
                      :data-testid="`project-branch-picker-status-todo-detail-${project.id}`"
                    >
                      {{ projectBranchPickerStatus('todo-detail', project.id) }}
                    </div>
                  </div>
                </span>
                <button
                  type="button"
                  class="todo-selected-project-remove"
                  :title="`Remove ${project.name}`"
                  :aria-label="`Remove ${project.name}`"
                  :data-testid="`todo-detail-selected-project-remove-${project.id}`"
                  :disabled="todoDetail.saving"
                  @click="removeTodoDetailProject(project.id)"
                >
                  <X :size="12" />
                </button>
              </span>
            </div>
            <input
              v-if="!todoDetail.readOnly"
              v-model="todoDetail.projectSearch"
              type="text"
              class="todo-text-input"
              data-testid="todo-detail-project-filter"
            />
            <div v-if="!todoDetail.readOnly" class="todo-project-options" data-testid="todo-detail-project-options">
              <div
                v-for="project in todoDetailProjectOptions"
                :key="project.id"
                class="todo-project-option-row"
              >
                <button
                  type="button"
                  class="todo-project-option"
                  :class="{ selected: todoDetail.projectSelections.some((selection) => selection.projectId === project.id) }"
                  :data-testid="`todo-detail-project-option-${project.id}`"
                  :aria-pressed="todoDetail.projectSelections.some((selection) => selection.projectId === project.id)"
                  :disabled="todoDetail.saving"
                  @click="toggleTodoDetailProject(project)"
                >
                  <span class="project-name">{{ project.name }}</span>
                  <span class="project-path">{{ project.path }}</span>
                </button>
                <button
                  type="button"
                  class="todo-project-option-clear"
                  :title="`Clear project candidate ${project.name}`"
                  :aria-label="`Clear project candidate ${project.name}`"
                  :data-testid="`clear-project-candidate-${project.id}`"
                  :disabled="todoDetail.saving"
                  @click.stop="clearProjectCandidate(project)"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
              <span v-if="todoDetailProjectOptions.length === 0" class="sidebar-empty">No matching projects</span>
            </div>
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeTodoDetail">Cancel</button>
          <button
            v-if="!todoDetail.readOnly"
            type="button"
            class="toolbar-button primary"
            data-testid="todo-detail-submit"
            :disabled="todoDetail.saving"
            @click="submitTodoDetail"
          >
            Save
          </button>
        </footer>
      </section>
    </div>

    <div v-if="projectPicker.visible" class="settings-overlay">
      <section class="settings-dialog todo-dialog" data-testid="todo-project-picker-dialog" @click.stop>
        <header class="settings-header">
          <div>
            <h2>Add Project</h2>
            <p>TODO context</p>
          </div>
          <button
            type="button"
            class="icon-button"
            data-testid="todo-project-picker-close"
            title="Close project picker"
            @click="closeProjectPicker"
          >
            <X :size="16" />
          </button>
        </header>

        <div class="settings-body todo-form-body">
          <div class="candidate-management-toolbar" data-testid="todo-project-picker-candidate-tools">
            <button
              type="button"
              class="toolbar-button compact"
              data-testid="import-single-project-candidate"
              :disabled="projectPicker.saving"
              @click="importSingleProjectCandidate"
            >
              <FolderPlus :size="14" />
              <span>Import project</span>
            </button>
            <button
              type="button"
              class="toolbar-button compact"
              data-testid="import-global-project-candidates"
              :title="bulkGitImportTooltip"
              :disabled="projectPicker.saving"
              @click="importProjectsFromParentDirectory"
            >
              <FolderInput :size="14" />
              <span>Import parent</span>
            </button>
            <button
              type="button"
              class="toolbar-button compact danger"
              data-testid="clear-global-project-candidates"
              :disabled="projectPicker.saving || projects.length === 0"
              @click="clearGlobalProjectCandidates"
            >
              <Trash2 :size="14" />
              <span>Clear candidates</span>
            </button>
          </div>
          <div v-if="importSummary" class="import-summary" data-testid="import-summary">
            <span>{{ importSummary.addedCount || 0 }} imported</span>
            <span>{{ importSummary.skippedCount || 0 }} skipped</span>
          </div>
          <div
            v-if="selectedProjectPickerProjects.length"
            class="todo-selected-project-tags"
            data-testid="todo-project-picker-tags"
          >
            <span
              v-for="project in selectedProjectPickerProjects"
              :key="project.id"
              class="todo-selected-project-tag"
              :data-testid="`todo-project-picker-tag-${project.id}`"
            >
              <span class="project-name">{{ project.name }}</span>
              <span class="todo-branch-picker">
                <input
                  :value="selectedProjectBaseBranch(projectPicker.projectSelections, project.id)"
                  type="text"
                  class="todo-branch-input"
                  placeholder="base 分支 (留空使用默认)"
                  :disabled="projectPicker.saving"
                  :data-testid="`todo-project-picker-branch-${project.id}`"
                  aria-label="Base branch"
                  @focus="openProjectBranchPicker('project-picker', project.id)"
                  @input="updateProjectBranchInput('project-picker', project.id, $event.target.value)"
                  @keydown.escape.stop.prevent="closeProjectBranchPicker"
                />
                <div
                  v-if="isProjectBranchPickerOpen('project-picker', project.id)"
                  class="project-branch-picker-menu"
                  :data-testid="`project-branch-picker-options-project-picker-${project.id}`"
                  role="listbox"
                >
                  <button
                    v-for="branch in visibleProjectBranches('project-picker', project.id)"
                    :key="branch"
                    type="button"
                    class="project-branch-picker-option"
                    :data-testid="`project-branch-picker-option-project-picker-${project.id}`"
                    role="option"
                    @mousedown.prevent
                    @click="selectProjectBranchCandidate('project-picker', project.id, branch)"
                  >
                    {{ branch }}
                  </button>
                  <div
                    v-if="projectBranchPickerStatus('project-picker', project.id)"
                    class="project-branch-picker-status"
                    :data-testid="`project-branch-picker-status-project-picker-${project.id}`"
                  >
                    {{ projectBranchPickerStatus('project-picker', project.id) }}
                  </div>
                </div>
              </span>
              <button
                type="button"
                class="todo-selected-project-remove"
                :title="`Remove ${project.name}`"
                :aria-label="`Remove ${project.name}`"
                :data-testid="`todo-project-picker-remove-${project.id}`"
                :disabled="projectPicker.saving"
                @click="removeProjectPickerProject(project.id)"
              >
                <X :size="12" />
              </button>
            </span>
          </div>
          <label class="settings-field">
            <span class="settings-label">Search</span>
            <input
              v-model="projectPicker.query"
              type="text"
              class="todo-text-input"
              data-testid="todo-project-picker-filter"
            />
          </label>
          <div class="todo-project-options" data-testid="todo-project-picker-options">
            <div
              v-for="project in projectPickerOptions"
              :key="project.id"
              class="todo-project-option-row"
            >
              <button
                type="button"
                class="todo-project-option"
                :class="{ selected: projectPicker.projectSelections.some((selection) => selection.projectId === project.id) }"
                :data-testid="`todo-project-picker-option-${project.id}`"
                :aria-pressed="projectPicker.projectSelections.some((selection) => selection.projectId === project.id)"
                :disabled="projectPicker.saving"
                @click="toggleProjectForTodo(project)"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="project-path">{{ project.path }}</span>
              </button>
              <button
                type="button"
                class="todo-project-option-clear"
                :title="`Clear project candidate ${project.name}`"
                :aria-label="`Clear project candidate ${project.name}`"
                :data-testid="`clear-project-candidate-${project.id}`"
                :disabled="projectPicker.saving"
                @click.stop="clearProjectCandidate(project)"
              >
                <Trash2 :size="14" />
              </button>
            </div>
            <span v-if="projectPickerOptions.length === 0" class="sidebar-empty">No matching projects</span>
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeProjectPicker">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="todo-project-picker-submit"
            :disabled="projectPicker.saving || projectPicker.projectSelections.length === 0"
            @click="submitProjectPicker"
          >
            Add
          </button>
        </footer>
      </section>
    </div>

    <div v-if="recentWorkspacePicker.visible" class="settings-overlay" @click="closeRecentWorkspacePicker">
      <section class="settings-dialog recent-workspace-dialog" data-testid="recent-workspace-dialog" @click.stop>
        <header class="settings-header">
          <div>
            <h2>Recent Projects</h2>
            <p>Workspaces</p>
          </div>
          <button type="button" class="icon-button" title="Close recent projects" @click="closeRecentWorkspacePicker">
            <X :size="16" />
          </button>
        </header>

        <div class="settings-body recent-workspace-list">
          <div
            v-if="recentWorkspacePicker.error"
            class="settings-error"
            data-testid="recent-workspace-error"
          >
            {{ recentWorkspacePicker.error }}
          </div>
          <div v-if="recentWorkspaces.length === 0" class="sidebar-empty" data-testid="recent-workspace-empty">
            No recent projects
          </div>
          <button
            v-for="(workspace, index) in recentWorkspaces"
            :key="workspace.path"
            type="button"
            class="recent-workspace-item"
            :data-testid="`recent-workspace-${index}`"
            :disabled="recentWorkspacePicker.openingPath === workspace.path"
            @click="openRecentWorkspace(workspace)"
          >
            <span class="recent-workspace-name">{{ workspace.name || workspace.path }}</span>
            <span class="recent-workspace-path">{{ workspace.path }}</span>
            <span v-if="workspace.available === false" class="project-status">Unavailable</span>
          </button>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeRecentWorkspacePicker">Cancel</button>
        </footer>
      </section>
    </div>

    <div v-if="settingsPanel.visible" class="settings-overlay" @click="closeTerminalSettings">
      <section class="settings-dialog" data-testid="terminal-settings-dialog" @click.stop>
        <header class="settings-header">
          <div>
            <h2>Terminal Settings</h2>
            <p>Embedded shell</p>
          </div>
          <button type="button" class="icon-button" title="Close settings" @click="closeTerminalSettings">
            <X :size="16" />
          </button>
        </header>

        <div v-if="settingsPanel.loading" class="settings-loading">Loading</div>
        <div v-else class="settings-body">
          <div class="settings-field" data-testid="terminal-settings-current">
            <span class="settings-label">Current</span>
            <strong>{{ shellDisplay(selectedTerminalShell) }}</strong>
            <span v-if="selectedTerminalShell && !selectedTerminalShell.available" class="settings-warning">Unavailable</span>
          </div>

          <div v-if="terminalSettingsFallback" class="settings-field" data-testid="terminal-settings-fallback">
            <span class="settings-label">Fallback</span>
            <strong>{{ shellDisplay(terminalSettingsFallback) }}</strong>
          </div>

          <div class="settings-field" data-testid="appearance-theme-setting">
            <span class="settings-label">Appearance</span>
            <div class="theme-options">
              <label class="theme-option">
                <input
                  v-model="settingsPanel.theme"
                  type="radio"
                  value="light"
                  data-testid="appearance-theme-light"
                />
                <strong>Light</strong>
              </label>
              <label class="theme-option">
                <input
                  v-model="settingsPanel.theme"
                  type="radio"
                  value="dark"
                  data-testid="appearance-theme-dark"
                />
                <strong>Dark</strong>
              </label>
            </div>
          </div>

          <label v-if="terminalSettingsDetected" class="settings-option" data-testid="terminal-settings-detected">
            <input
              v-model="settingsPanel.mode"
              type="radio"
              value="detected"
              data-testid="terminal-settings-detected-option"
            />
            <span>
              <span class="settings-label">Detected</span>
              <strong>{{ shellDisplay(terminalSettingsDetected) }}</strong>
            </span>
          </label>

          <label class="settings-option">
            <input
              v-model="settingsPanel.mode"
              type="radio"
              value="manual"
              data-testid="terminal-settings-manual-option"
            />
            <span class="manual-shell-entry">
              <span class="settings-label">Manual path</span>
              <input
                v-model="settingsPanel.manualPath"
                type="text"
                data-testid="terminal-settings-manual-path"
                placeholder="/usr/bin/zsh"
                @focus="settingsPanel.mode = 'manual'"
              />
            </span>
          </label>

          <div class="settings-field" data-testid="terminal-settings-built-in-launch-profile">
            <span class="settings-label">Built-in launch</span>
            <strong>Terminal</strong>
          </div>

          <div class="launch-profile-settings" data-testid="terminal-launch-profiles">
            <div class="launch-profile-header">
              <span class="settings-label">Launch profiles</span>
              <button
                type="button"
                class="icon-button"
                data-testid="terminal-launch-profile-add"
                title="Add launch profile"
                @click="addTerminalLaunchProfile"
              >
                <Plus :size="14" />
              </button>
            </div>
            <div
              v-for="(profile, index) in settingsPanel.launchProfiles"
              :key="index"
              class="launch-profile-row"
              :data-testid="`terminal-launch-profile-${index}`"
            >
              <label class="launch-profile-enabled" :title="profile.enabled ? 'Enabled' : 'Disabled'">
                <input
                  v-model="profile.enabled"
                  type="checkbox"
                  :data-testid="`terminal-launch-profile-enabled-${index}`"
                />
                <span class="visually-hidden">Enabled</span>
              </label>
              <input
                v-model="profile.name"
                type="text"
                :data-testid="`terminal-launch-profile-name-${index}`"
                placeholder="codex"
              />
              <input
                v-model="profile.command"
                type="text"
                :data-testid="`terminal-launch-profile-command-${index}`"
                placeholder="codex"
              />
              <label class="launch-profile-enabled" :title="profile.background ? 'Background' : 'Foreground'">
                <input
                  v-model="profile.background"
                  type="checkbox"
                  :data-testid="`terminal-launch-profile-background-${index}`"
                />
                <span class="visually-hidden">Background</span>
              </label>
              <button
                type="button"
                class="icon-button"
                :data-testid="`terminal-launch-profile-up-${index}`"
                title="Move up"
                :disabled="index === 0"
                @click="moveTerminalLaunchProfile(index, -1)"
              >
                <ChevronUp :size="14" />
              </button>
              <button
                type="button"
                class="icon-button"
                :data-testid="`terminal-launch-profile-down-${index}`"
                title="Move down"
                :disabled="index === settingsPanel.launchProfiles.length - 1"
                @click="moveTerminalLaunchProfile(index, 1)"
              >
                <ChevronDown :size="14" />
              </button>
              <button
                type="button"
                class="icon-button"
                :data-testid="`terminal-launch-profile-remove-${index}`"
                title="Remove launch profile"
                @click="removeTerminalLaunchProfile(index)"
              >
                <Trash2 :size="14" />
              </button>
            </div>
          </div>

          <div class="settings-field" data-testid="claude-status-setting">
            <span class="settings-label">Claude 状态监控</span>
            <div class="claude-status-row">
              <span v-if="settingsPanel.claudeStatus.checking" class="settings-hint">检测中…</span>
              <template v-else>
                <strong v-if="settingsPanel.claudeStatus.installed && settingsPanel.claudeStatus.stale" class="settings-warning">需更新</strong>
                <strong v-else-if="settingsPanel.claudeStatus.installed">已安装</strong>
                <strong v-else>未安装</strong>
                <button
                  v-if="settingsPanel.claudeStatus.installed && !settingsPanel.claudeStatus.stale"
                  type="button"
                  class="toolbar-button"
                  data-testid="claude-status-uninstall"
                  :disabled="!activeProjectPath"
                  @click="uninstallClaudeStatusHook"
                >
                  卸载
                </button>
                <button
                  v-else
                  type="button"
                  class="toolbar-button primary"
                  data-testid="claude-status-install"
                  :disabled="!activeProjectPath"
                  @click="installClaudeStatusHook"
                >
                  {{ settingsPanel.claudeStatus.installed ? '重装' : '安装' }}
                </button>
              </template>
            </div>
            <p v-if="!activeProjectPath" class="settings-hint">选择一个项目后再操作</p>
          </div>

          <div v-if="settingsPanel.error" class="settings-error" data-testid="terminal-settings-error">
            {{ settingsPanel.error }}
          </div>
        </div>

        <footer class="settings-actions">
          <button
            type="button"
            class="toolbar-button"
            data-testid="terminal-settings-redetect"
            :disabled="settingsPanel.detecting || settingsPanel.loading"
            @click="redetectTerminalShell"
          >
            Detect
          </button>
          <button type="button" class="toolbar-button" @click="closeTerminalSettings">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="terminal-settings-save"
            :disabled="settingsPanel.saving || settingsPanel.loading"
            @click="saveTerminalSettings"
          >
            Save
          </button>
        </footer>
      </section>
    </div>

    <div
      v-if="initializationFileManagement.visible"
      class="settings-overlay"
      @click="closeTodoInitializationFileManagement"
    >
      <section
        class="settings-dialog initialization-file-management-dialog"
        data-testid="todo-initialization-file-management-dialog"
        @click.stop
      >
        <header class="settings-header">
          <div>
            <h2>文件管理</h2>
            <p>TODO initialization files</p>
          </div>
          <button
            type="button"
            class="icon-button"
            title="Close file management"
            @click="closeTodoInitializationFileManagement"
          >
            <X :size="16" />
          </button>
        </header>

        <div v-if="initializationFileManagement.loading" class="settings-loading">Loading</div>
        <div v-else class="settings-body initialization-file-management-body">
          <div class="launch-profile-settings" data-testid="todo-initialization-files-settings">
            <div class="launch-profile-header">
              <span class="settings-label">Files</span>
              <button
                type="button"
                class="icon-button"
                data-testid="todo-initialization-file-add"
                title="Add initialization file"
                @click="addTodoInitializationFile"
              >
                <Plus :size="14" />
              </button>
            </div>
            <div
              v-if="!initializationFileManagement.files.length"
              class="initialization-file-empty"
              data-testid="todo-initialization-file-empty"
            >
              No files
            </div>
            <div
              v-for="(file, index) in initializationFileManagement.files"
              :key="index"
              class="initialization-file-row"
              :data-testid="`todo-initialization-file-setting-${index}`"
            >
              <label
                class="initialization-file-default"
                :title="file.defaultSelected ? 'Selected by default' : 'Optional'"
              >
                <input
                  v-model="file.defaultSelected"
                  type="checkbox"
                  :data-testid="`todo-initialization-file-default-${index}`"
                />
                <span>默认</span>
              </label>
              <input
                v-model="file.name"
                class="initialization-file-name-input"
                type="text"
                :data-testid="`todo-initialization-file-name-${index}`"
                placeholder="显示名称"
              />
              <input
                v-model="file.description"
                class="initialization-file-description-input"
                type="text"
                :data-testid="`todo-initialization-file-description-${index}`"
                placeholder="描述"
              />
              <div class="initialization-file-upload-cell">
                <label class="toolbar-button compact">
                  <FileText :size="14" />
                  <span>{{ file.fileName ? '更换文件' : '上传文件' }}</span>
                  <input
                    class="visually-hidden"
                    type="file"
                    :data-testid="`todo-initialization-file-upload-${index}`"
                    @change="uploadTodoInitializationFile(index, $event)"
                  />
                </label>
                <span
                  class="initialization-file-upload-name"
                  :data-testid="`todo-initialization-file-uploaded-name-${index}`"
                >
                  {{ file.fileName || '未上传文件' }}
                </span>
              </div>
              <button
                type="button"
                class="icon-button initialization-file-move-up"
                :data-testid="`todo-initialization-file-up-${index}`"
                title="Move up"
                :disabled="index === 0"
                @click="moveTodoInitializationFile(index, -1)"
              >
                <ChevronUp :size="14" />
              </button>
              <button
                type="button"
                class="icon-button initialization-file-move-down"
                :data-testid="`todo-initialization-file-down-${index}`"
                title="Move down"
                :disabled="index === initializationFileManagement.files.length - 1"
                @click="moveTodoInitializationFile(index, 1)"
              >
                <ChevronDown :size="14" />
              </button>
              <button
                type="button"
                class="icon-button initialization-file-remove"
                :data-testid="`todo-initialization-file-remove-${index}`"
                title="Remove initialization file"
                @click="removeTodoInitializationFile(index)"
              >
                <Trash2 :size="14" />
              </button>
            </div>
          </div>

          <div
            v-if="initializationFileManagement.error"
            class="settings-error"
            data-testid="todo-initialization-file-management-error"
          >
            {{ initializationFileManagement.error }}
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeTodoInitializationFileManagement">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="todo-initialization-file-management-save"
            :disabled="initializationFileManagement.saving || initializationFileManagement.loading"
            @click="saveTodoInitializationFileManagement"
          >
            Save
          </button>
        </footer>
      </section>
    </div>

    <div
      v-if="lifecycleScriptManagement.visible"
      class="settings-overlay"
      @click="closeTodoLifecycleScriptManagement"
    >
      <section
        class="settings-dialog lifecycle-script-management-dialog"
        data-testid="todo-lifecycle-script-management-dialog"
        @click.stop
      >
        <header class="settings-header">
          <div>
            <h2>脚本管理</h2>
            <p>TODO lifecycle scripts</p>
          </div>
          <button
            type="button"
            class="icon-button"
            title="Close script management"
            @click="closeTodoLifecycleScriptManagement"
          >
            <X :size="16" />
          </button>
        </header>

        <div v-if="lifecycleScriptManagement.loading" class="settings-loading">Loading</div>
        <div v-else class="settings-body lifecycle-script-management-body">
          <div class="launch-profile-settings" data-testid="todo-lifecycle-scripts-settings">
            <div class="launch-profile-header">
              <span class="settings-label">Scripts</span>
              <button
                type="button"
                class="icon-button"
                data-testid="todo-lifecycle-script-add"
                title="Add lifecycle script"
                @click="addTodoLifecycleScript"
              >
                <Plus :size="14" />
              </button>
            </div>
            <div
              v-if="!lifecycleScriptManagement.scripts.length"
              class="initialization-file-empty"
              data-testid="todo-lifecycle-script-empty"
            >
              No scripts
            </div>
            <div
              v-for="(script, index) in lifecycleScriptManagement.scripts"
              :key="index"
              class="lifecycle-script-row"
              :data-testid="`todo-lifecycle-script-setting-${index}`"
            >
              <label
                class="initialization-file-default"
                :title="script.defaultSelected ? 'Selected by default' : 'Optional'"
              >
                <input
                  :checked="script.defaultSelected"
                  type="checkbox"
                  :data-testid="`todo-lifecycle-script-default-${index}`"
                  @change="setTodoLifecycleScriptDefault(index, $event.target.checked)"
                />
                <span>默认</span>
              </label>
              <input
                v-model="script.name"
                class="initialization-file-name-input"
                type="text"
                :data-testid="`todo-lifecycle-script-name-${index}`"
                placeholder="显示名称"
              />
              <input
                v-model="script.description"
                class="initialization-file-description-input"
                type="text"
                :data-testid="`todo-lifecycle-script-description-${index}`"
                placeholder="描述"
              />
              <textarea
                v-model="script.initScript"
                class="lifecycle-script-textarea"
                rows="3"
                :data-testid="`todo-lifecycle-script-init-${index}`"
                placeholder="初始化脚本"
              ></textarea>
              <textarea
                v-model="script.completeScript"
                class="lifecycle-script-textarea"
                rows="3"
                :data-testid="`todo-lifecycle-script-complete-${index}`"
                placeholder="完成脚本"
              ></textarea>
              <button
                type="button"
                class="icon-button initialization-file-move-up"
                :data-testid="`todo-lifecycle-script-up-${index}`"
                title="Move up"
                :disabled="index === 0"
                @click="moveTodoLifecycleScript(index, -1)"
              >
                <ChevronUp :size="14" />
              </button>
              <button
                type="button"
                class="icon-button initialization-file-move-down"
                :data-testid="`todo-lifecycle-script-down-${index}`"
                title="Move down"
                :disabled="index === lifecycleScriptManagement.scripts.length - 1"
                @click="moveTodoLifecycleScript(index, 1)"
              >
                <ChevronDown :size="14" />
              </button>
              <button
                type="button"
                class="icon-button initialization-file-remove"
                :data-testid="`todo-lifecycle-script-remove-${index}`"
                title="Remove lifecycle script"
                @click="removeTodoLifecycleScript(index)"
              >
                <Trash2 :size="14" />
              </button>
            </div>
          </div>

          <div
            v-if="lifecycleScriptManagement.error"
            class="settings-error"
            data-testid="todo-lifecycle-script-management-error"
          >
            {{ lifecycleScriptManagement.error }}
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeTodoLifecycleScriptManagement">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="todo-lifecycle-script-management-save"
            :disabled="lifecycleScriptManagement.saving || lifecycleScriptManagement.loading"
            @click="saveTodoLifecycleScriptManagement"
          >
            Save
          </button>
        </footer>
      </section>
    </div>
  </main>
</template>
