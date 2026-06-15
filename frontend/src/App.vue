<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ChevronDown, ChevronUp, GitBranch, Plus, RotateCcw, Settings, Trash2, X } from '@lucide/vue'
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
  AddProjectsToTodo,
  ChangeTodoStatus,
  CompleteTodo,
  CreateProjectFromDialog,
  CreateTodo,
  CreateTodoTerminal,
  DeleteProject,
  DeleteProjects,
  DeleteTerminal,
  DeleteTodo,
  DetectTerminalShell,
  GetProjectGitStatus,
  ImportProjectsFromParentDirectoryDialog,
  InitializeProjectGitRepository,
  ListProjects,
  LoadTerminalSettings,
  RemoveTodoProject,
  ResizeTerminal,
  SelectProject,
  SelectTerminal,
  SelectTodoProject,
  SaveTerminalLaunchProfiles,
  SaveTerminalShell,
  SaveTerminalTheme,
  SendTerminalInput,
  StartShell,
  UpdateTodo
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText, EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

const projects = ref([])
const todos = ref([])
const todoProjects = ref([])
const terminals = ref([])
const activeProjectId = ref('')
const activeTodoId = ref('')
const activeTodoProjectId = ref('')
const activeTerminalId = ref('')
const importSummary = ref(null)
const shellStatuses = reactive({})
const terminalContainers = new Map()
const titleActivityTimers = new Map()
const autoRestartedTerminals = new Set()
const errorMessage = ref('')
let gitStatusRequestId = 0
let gitStatusInFlightProjectId = ''
let gitStatusInFlightRequestId = 0
let lastFocusGitRefreshProjectId = ''
let lastFocusGitRefreshAt = 0
const focusGitRefreshDedupeMs = 500
const terminalMenu = reactive({
  visible: false,
  terminalId: '',
  x: 0,
  y: 0
})
const sidebarWidth = ref(280)
const sidebarResize = reactive({
  active: false,
  startX: 0,
  startWidth: 280
})
const sidebarMinWidth = 220
const sidebarMaxWidth = 520
const defaultTerminalLaunchProfiles = [
  { name: 'codex', command: 'codex --dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox', enabled: true },
  { name: 'claude', command: 'claude --dangerously-skip-permissions', enabled: true }
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
const settingsPanel = reactive({
  visible: false,
  loading: false,
  detecting: false,
  saving: false,
  mode: 'detected',
  manualPath: '',
  launchProfiles: [],
  theme: 'light',
  error: ''
})
const todoForm = reactive({
  visible: false,
  title: '',
  description: '',
  priority: 'medium',
  projectIds: [],
  projectSearch: '',
  saving: false
})
const todoDetail = reactive({
  visible: false,
  todoId: '',
  title: '',
  description: '',
  priority: 'medium',
  projectIds: [],
  projectSearch: '',
  saving: false
})
const projectPicker = reactive({
  visible: false,
  todoId: '',
  query: '',
  projectIds: [],
  saving: false
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
  return projects.value.find((project) => project.id === activeProjectId.value) || null
})

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
  return projects.value.find((project) => project.id === todoProject.projectId) || null
})

const activeTerminal = computed(() => {
  return terminals.value.find((terminal) => terminal.id === activeTerminalId.value) || null
})

const activeTerminalState = computed(() => {
  return activeTerminal.value ? terminalState(activeTerminal.value.id) : ''
})

const selectedTerminalShell = computed(() => terminalSettings.value?.selected || null)
const terminalLaunchProfiles = computed(() => {
  if (Array.isArray(terminalSettings.value?.launchProfiles)) {
    return terminalSettings.value.launchProfiles
  }
  return defaultTerminalLaunchProfiles
})

const currentGitStatus = computed(() => {
  if (!activeProject.value || gitStatus.value?.projectId !== activeProject.value.id) {
    return null
  }
  return gitStatus.value
})

const displayedGitStatus = computed(() => {
  if (!activeProject.value || !gitStatus.value) {
    return null
  }
  if (gitStatusLoading.value && gitStatus.value.projectId !== activeProject.value.id) {
    return null
  }
  return gitStatus.value
})

const projectGitStatusChips = computed(() => {
  if (!activeProject.value) {
    return [statusChip('neutral', 'neutral', 'No project')]
  }
  const status = displayedGitStatus.value
  if (!activeProject.value.available || status?.pathUnavailable) {
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
  const selectedProjectIds = new Set(todoForm.projectIds)
  return projects.value.filter((project) => selectedProjectIds.has(project.id))
})

const todoFormProjectOptions = computed(() => {
  return filteredProjects(projects.value, todoForm.projectSearch)
})

const selectedTodoDetailProjects = computed(() => {
  const selectedProjectIds = new Set(todoDetail.projectIds)
  return projects.value.filter((project) => selectedProjectIds.has(project.id))
})

const todoDetailProjectOptions = computed(() => {
  return filteredProjects(projects.value, todoDetail.projectSearch)
})

const removedTodoDetailProjectsWithTerminals = computed(() => {
  if (!todoDetail.visible || !todoDetail.todoId) {
    return []
  }

  const selectedProjectIds = new Set(todoDetail.projectIds)
  return todoProjects.value.filter((todoProject) => {
    return (
      todoProject.todoId === todoDetail.todoId &&
      !selectedProjectIds.has(todoProject.projectId) &&
      terminals.value.some((terminal) => terminal.todoProjectId === todoProject.id)
    )
  })
})

const projectPickerOptions = computed(() => {
  const linkedProjectIds = new Set(
    todoProjects.value
      .filter((todoProject) => todoProject.todoId === projectPicker.todoId)
      .map((todoProject) => todoProject.projectId)
  )
  return filteredProjects(
    projects.value.filter((project) => !linkedProjectIds.has(project.id)),
    projectPicker.query
  )
})

const selectedProjectPickerProjects = computed(() => {
  const selectedProjectIds = new Set(projectPicker.projectIds)
  return projects.value.filter((project) => selectedProjectIds.has(project.id))
})

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
  window.addEventListener('resize', fitActiveTerminal)
  window.addEventListener('focus', refreshProjectGitStatusOnFocus)
  window.addEventListener('click', closeTerminalMenu)

  try {
    applyTerminalSettings(await LoadTerminalSettings())
  } catch (error) {
    showError(error)
  }

  try {
    applyState(await ListProjects())
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
  window.removeEventListener('resize', fitActiveTerminal)
  window.removeEventListener('focus', refreshProjectGitStatusOnFocus)
  window.removeEventListener('click', closeTerminalMenu)
  window.removeEventListener('mousemove', resizeSidebar)
  window.removeEventListener('mouseup', stopSidebarResize)
  clearAllTitleActivityTimers()
})

function applyState(state, options = {}) {
  const previousActiveProjectId = activeProjectId.value
  const previousTerminals = new Map(terminals.value.map((terminal) => [terminal.id, terminal]))
  projects.value = state?.projects || []
  todos.value = state?.todos || []
  todoProjects.value = state?.todoProjects || []
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
      delete shellStatuses[terminalId]
    }
  }
  terminals.value = nextTerminals
  activeProjectId.value = state?.activeProjectId || ''
  activeTodoId.value = state?.activeTodoId || ''
  activeTodoProjectId.value = state?.activeTodoProjectId || ''
  activeTerminalId.value = state?.activeTerminalId || ''
  for (const terminal of terminals.value) {
    if (terminal.state) {
      shellStatuses[terminal.id] = terminal.state
    }
  }
  closeTerminalMenu()
  syncGitStatusForActiveProject(previousActiveProjectId, {
    refresh: options.refreshGitStatus !== false,
    dedupePending: options.dedupeGitStatus === true,
    force: options.forceGitStatusRefresh === true
  })
}

async function createProject() {
  try {
    applyState(await CreateProjectFromDialog())
  } catch (error) {
    showError(error)
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

function createTodo() {
  todoForm.visible = true
  todoForm.title = ''
  todoForm.description = ''
  todoForm.priority = 'medium'
  todoForm.projectIds = []
  todoForm.projectSearch = ''
  todoForm.saving = false
  errorMessage.value = ''
}

function closeTodoForm() {
  todoForm.visible = false
  todoForm.saving = false
}

async function submitTodoForm() {
  const title = todoForm.title.trim()
  if (!title) {
    showError('TODO title is required')
    return
  }

  todoForm.saving = true
  try {
    applyState(
      await CreateTodo({
        title,
        description: todoForm.description.trim(),
        priority: normalizedTodoPriority(todoForm.priority),
        projectIds: [...todoForm.projectIds]
      })
    )
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

  todoDetail.visible = true
  todoDetail.todoId = todo.id
  todoDetail.title = todo.title || ''
  todoDetail.description = todo.description || ''
  todoDetail.priority = normalizedTodoPriority(todo.priority)
  todoDetail.projectIds = todoProjects.value
    .filter((todoProject) => todoProject.todoId === todo.id)
    .map((todoProject) => todoProject.projectId)
  todoDetail.projectSearch = ''
  todoDetail.saving = false
  errorMessage.value = ''
}

function closeTodoDetail() {
  todoDetail.visible = false
  todoDetail.todoId = ''
  todoDetail.title = ''
  todoDetail.description = ''
  todoDetail.priority = 'medium'
  todoDetail.projectIds = []
  todoDetail.projectSearch = ''
  todoDetail.saving = false
}

function toggleTodoDetailProject(project) {
  if (!project?.id) {
    return
  }
  todoDetail.projectIds = toggleProjectId(todoDetail.projectIds, project.id)
}

function removeTodoDetailProject(projectId) {
  todoDetail.projectIds = todoDetail.projectIds.filter((selectedProjectId) => selectedProjectId !== projectId)
}

async function submitTodoDetail() {
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
        projectIds: [...todoDetail.projectIds]
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
  projectPicker.projectIds = []
  projectPicker.saving = false
  if (projectPickerOptions.value.length === 0) {
    showError('No available projects to add')
    return
  }
  projectPicker.visible = true
  errorMessage.value = ''
}

function closeProjectPicker() {
  projectPicker.visible = false
  projectPicker.todoId = ''
  projectPicker.query = ''
  projectPicker.projectIds = []
  projectPicker.saving = false
}

function toggleProjectForTodo(project) {
  if (!project?.id) {
    return
  }
  projectPicker.projectIds = toggleProjectId(projectPicker.projectIds, project.id)
}

function removeProjectPickerProject(projectId) {
  projectPicker.projectIds = projectPicker.projectIds.filter((selectedProjectId) => selectedProjectId !== projectId)
}

async function submitProjectPicker() {
  if (!projectPicker.todoId || projectPicker.projectIds.length === 0) {
    return
  }
  projectPicker.saving = true
  try {
    applyState(await AddProjectsToTodo(projectPicker.todoId, [...projectPicker.projectIds]))
    closeProjectPicker()
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  } finally {
    projectPicker.saving = false
  }
}

function toggleTodoFormProject(project) {
  if (!project?.id) {
    return
  }
  todoForm.projectIds = toggleProjectId(todoForm.projectIds, project.id)
}

function removeTodoFormProject(projectId) {
  todoForm.projectIds = todoForm.projectIds.filter((selectedProjectId) => selectedProjectId !== projectId)
}

async function selectTodoProject(todoProjectId) {
  try {
    applyState(await SelectTodoProject(todoProjectId), { dedupeGitStatus: true, forceGitStatusRefresh: true })
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
  if (!activeTodoProjectProject.value?.available) {
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
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    const state = await CreateTodoTerminal(todoProjectId, size.cols || 80, size.rows || 24)
    applyState(state)
    await activateActiveTerminal()
    if (launchProfile?.command && state?.activeTerminalId) {
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

async function deleteTodo(todoId) {
  try {
    applyState(await DeleteTodo(todoId))
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
  if (!terminal || !activeTodoProjectProject.value?.available) {
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
}

function closeTerminalSettings() {
  settingsPanel.visible = false
  settingsPanel.error = ''
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
    enabled: profile.enabled !== false
  }))
}

function addTerminalLaunchProfile() {
  settingsPanel.launchProfiles.push({ name: '', command: '', enabled: true })
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
    enabled: profile.enabled !== false
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
  if (!activeProject.value) {
    gitStatusRequestId += 1
    gitStatusInFlightProjectId = ''
    gitStatusInFlightRequestId = 0
    gitStatus.value = null
    gitStatusLoading.value = false
    gitStatusError.value = ''
    gitInitLoading.value = false
    return
  }
  if (!activeProject.value.available) {
    gitStatusRequestId += 1
    gitStatusInFlightProjectId = ''
    gitStatusInFlightRequestId = 0
    gitStatus.value = { projectId: activeProject.value.id, isRepo: false, pathUnavailable: true }
    gitStatusLoading.value = false
    gitStatusError.value = ''
    gitInitLoading.value = false
    return
  }
  if (options.refresh !== false && (options.force === true || activeProject.value.id !== previousActiveProjectId)) {
    refreshProjectGitStatus({ dedupePending: options.dedupePending === true })
  }
}

async function refreshProjectGitStatus(options = {}) {
  const project = activeProject.value
  if (!project || !project.available) {
    syncGitStatusForActiveProject(activeProjectId.value)
    return
  }
  if (options.dedupePending === true && gitStatusInFlightProjectId === project.id) {
    return
  }
  const requestId = gitStatusRequestId + 1
  gitStatusRequestId = requestId
  const projectId = project.id
  gitStatusInFlightProjectId = projectId
  gitStatusInFlightRequestId = requestId
  gitStatusLoading.value = true
  gitStatusError.value = ''
  try {
    const status = await GetProjectGitStatus(projectId)
    if (requestId !== gitStatusRequestId || activeProjectId.value !== projectId) {
      return
    }
    gitStatus.value = status
  } catch (error) {
    if (requestId !== gitStatusRequestId || activeProjectId.value !== projectId) {
      return
    }
    gitStatus.value = null
    gitStatusError.value = errorMessageFrom(error)
  } finally {
    if (gitStatusInFlightProjectId === projectId && gitStatusInFlightRequestId === requestId) {
      gitStatusInFlightProjectId = ''
      gitStatusInFlightRequestId = 0
    }
    if (requestId === gitStatusRequestId) {
      gitStatusLoading.value = false
    }
  }
}

function refreshProjectGitStatusOnFocus() {
  const project = activeProject.value
  const now = Date.now()
  if (
    project &&
    lastFocusGitRefreshProjectId === project.id &&
    now - lastFocusGitRefreshAt < focusGitRefreshDedupeMs
  ) {
    return
  }
  lastFocusGitRefreshProjectId = project?.id || ''
  lastFocusGitRefreshAt = now
  refreshProjectGitStatus({ dedupePending: true })
}

function handleTodoExpanded(todoId) {
  const todoProject = activeTodoProject.value
  if (!todoProject || todoProject.todoId !== todoId) {
    return
  }
  refreshProjectGitStatus({ dedupePending: true })
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
    if (activeProjectId.value === projectId) {
      await refreshProjectGitStatus()
    }
  } catch (error) {
    showError(error)
  } finally {
    if (activeProjectId.value === projectId) {
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

function normalizeSearch(value) {
  return (value || '').trim().toLowerCase()
}

function normalizedTodoPriority(priority) {
  return todoPriorities.some((option) => option.value === priority) ? priority : 'medium'
}

function toggleProjectId(projectIds, projectId) {
  if (projectIds.includes(projectId)) {
    return projectIds.filter((candidate) => candidate !== projectId)
  }
  return [...projectIds, projectId]
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
  scheduleFitActiveTerminal()
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
    if (terminal.id === activeTerminalId.value) {
      refreshProjectGitStatus()
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
  Object.assign(terminal, applyAgentStatusEvent(terminal, event))
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
</script>

<template>
  <main class="app-shell" :data-theme="currentTheme" :style="{ '--sidebar-width': `${sidebarWidth}px` }">
    <ProjectSidebar
      :projects="projects"
      :todos="todos"
      :todo-projects="todoProjects"
      :terminals="terminals"
      :active-project-id="activeProjectId"
      :active-todo-id="activeTodoId"
      :active-todo-project-id="activeTodoProjectId"
      :active-terminal-id="activeTerminalId"
      :launch-profiles="terminalLaunchProfiles"
      :import-summary="importSummary"
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
      @copy-todo-description="copyTodoDescription"
      @delete-todo="deleteTodo"
      @create-terminal="createTerminal"
      @select-terminal="selectTerminal"
      @delete-project="deleteProject"
      @delete-projects="deleteProjects"
      @delete-terminal="deleteTerminal"
    />

    <div
      class="sidebar-resize-handle"
      data-testid="sidebar-resize-handle"
      role="separator"
      aria-orientation="vertical"
      title="Resize sidebar"
      @mousedown="startSidebarResize"
    ></div>

    <section class="workspace">
      <header class="workspace-header">
        <div v-if="activeTodoProject && activeTodoProjectProject" class="project-heading">
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
            data-testid="settings-toggle"
            title="Settings"
            @click="openTerminalSettings"
          >
            <Settings :size="16" />
            <span>Settings</span>
          </button>
          <button
            v-if="activeTodoProjectProject && activeTerminalState === 'exited'"
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

      <div class="terminal-surface" data-testid="terminal-surface">
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

        <div v-if="!activeTodoProject || !activeTodoProjectProject" class="state-layer">Select a TODO project</div>
        <div v-else-if="!activeTodoProjectProject.available" class="state-layer warning">Project path unavailable</div>
        <div v-else-if="!activeTerminal" class="state-layer">Select a terminal</div>
        <div v-else-if="activeTerminalState === 'unsupported'" class="state-layer warning">
          Embedded terminal is not supported on Windows
        </div>
        <div v-else-if="activeTerminalState === 'exited'" class="state-layer warning">Shell exited</div>
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
          <button type="button" class="icon-button" title="Close TODO form" @click="closeTodoForm">
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

          <div class="settings-field">
            <span class="settings-label">Projects</span>
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
              <button
                v-for="project in todoFormProjectOptions"
                :key="project.id"
                type="button"
                class="todo-project-option"
                :class="{ selected: todoForm.projectIds.includes(project.id) }"
                :data-testid="`todo-project-option-${project.id}`"
                :aria-pressed="todoForm.projectIds.includes(project.id)"
                @click="toggleTodoFormProject(project)"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="project-path">{{ project.path }}</span>
              </button>
              <span v-if="todoFormProjectOptions.length === 0" class="sidebar-empty">No matching projects</span>
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
          <button type="button" class="icon-button" title="Close TODO detail" @click="closeTodoDetail">
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
              data-testid="todo-detail-name-input"
            />
          </label>

          <label class="settings-field">
            <span class="settings-label">Description</span>
            <textarea
              v-model="todoDetail.description"
              class="todo-textarea"
              rows="3"
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
                  :data-testid="`todo-detail-priority-${priority.value}`"
                />
                <span>{{ priority.label }}</span>
              </label>
            </div>
          </div>

          <div class="settings-field">
            <span class="settings-label">Projects</span>
            <div
              v-if="selectedTodoDetailProjects.length"
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
              v-model="todoDetail.projectSearch"
              type="text"
              class="todo-text-input"
              data-testid="todo-detail-project-filter"
            />
            <div class="todo-project-options" data-testid="todo-detail-project-options">
              <button
                v-for="project in todoDetailProjectOptions"
                :key="project.id"
                type="button"
                class="todo-project-option"
                :class="{ selected: todoDetail.projectIds.includes(project.id) }"
                :data-testid="`todo-detail-project-option-${project.id}`"
                :aria-pressed="todoDetail.projectIds.includes(project.id)"
                :disabled="todoDetail.saving"
                @click="toggleTodoDetailProject(project)"
              >
                <span class="project-name">{{ project.name }}</span>
                <span class="project-path">{{ project.path }}</span>
              </button>
              <span v-if="todoDetailProjectOptions.length === 0" class="sidebar-empty">No matching projects</span>
            </div>
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeTodoDetail">Cancel</button>
          <button
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
          <button type="button" class="icon-button" title="Close project picker" @click="closeProjectPicker">
            <X :size="16" />
          </button>
        </header>

        <div class="settings-body todo-form-body">
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
            <button
              v-for="project in projectPickerOptions"
              :key="project.id"
              type="button"
              class="todo-project-option"
              :class="{ selected: projectPicker.projectIds.includes(project.id) }"
              :data-testid="`todo-project-picker-option-${project.id}`"
              :aria-pressed="projectPicker.projectIds.includes(project.id)"
              :disabled="projectPicker.saving"
              @click="toggleProjectForTodo(project)"
            >
              <span class="project-name">{{ project.name }}</span>
              <span class="project-path">{{ project.path }}</span>
            </button>
            <span v-if="projectPickerOptions.length === 0" class="sidebar-empty">No matching projects</span>
          </div>
        </div>

        <footer class="settings-actions">
          <button type="button" class="toolbar-button" @click="closeProjectPicker">Cancel</button>
          <button
            type="button"
            class="toolbar-button primary"
            data-testid="todo-project-picker-submit"
            :disabled="projectPicker.saving || projectPicker.projectIds.length === 0"
            @click="submitProjectPicker"
          >
            Add
          </button>
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
  </main>
</template>
