<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Archive,
  Check,
  ChevronDown,
  ChevronRight,
  CircleAlert,
  Copy,
  EllipsisVertical,
  Eye,
  FolderGit2,
  FolderPlus,
  GripVertical,
  ListChevronsDownUp,
  ListChevronsUpDown,
  ListTodo,
  LoaderCircle,
  Play,
  Plus,
  TerminalSquare,
  TriangleAlert,
  Trash2
} from '@lucide/vue'
import { createTodoSortable } from '../todoSortable'

const props = defineProps({
  projects: {
    type: Array,
    default: () => []
  },
  todos: {
    type: Array,
    default: () => []
  },
  todoProjects: {
    type: Array,
    default: () => []
  },
  todoProjectBranches: {
    type: Object,
    default: () => ({})
  },
  terminals: {
    type: Array,
    default: () => []
  },
  activeProjectId: {
    type: String,
    default: ''
  },
  activeTodoId: {
    type: String,
    default: ''
  },
  activeTodoProjectId: {
    type: String,
    default: ''
  },
  activeTerminalId: {
    type: String,
    default: ''
  },
  launchProfiles: {
    type: Array,
    default: () => []
  },
  importSummary: {
    type: Object,
    default: null
  },
  hasWorkspace: {
    type: Boolean,
    default: true
  },
  todoView: {
    type: String,
    default: ''
  },
  todoSortMode: {
    type: String,
    default: ''
  },
  todoOrders: {
    type: Object,
    default: () => ({ notStarted: [], inProgress: [] })
  },
  todoOrdersInitialized: {
    type: Boolean,
    default: false
  },
  todoOrderSaving: {
    type: Boolean,
    default: false
  },
  completedMergeStatuses: {
    type: Object,
    default: () => ({})
  },
  lifecycleScriptStatuses: {
    type: Array,
    default: () => []
  },
  lifecycleScriptErrorOutputs: {
    type: Object,
    default: () => ({})
  },
  lifecycleScriptErrorScope: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'create-project',
  'import-projects',
  'select-project',
  'delete-project',
  'delete-projects',
  'create-todo',
  'edit-todo',
  'add-project-to-todo',
  'select-todo-project',
  'remove-todo-project',
  'change-todo-status',
  'complete-todo',
  'retry-todo-lifecycle-script',
  'request-todo-lifecycle-script-error',
  'copy-todo-lifecycle-script-error',
  'copy-todo-description',
  'delete-todo',
  'delete-completed-todos',
  'create-task-terminal',
  'create-terminal',
  'select-terminal',
  'delete-terminal',
  'open-todo-folder',
  'todo-expanded',
  'update:todo-view',
  'todo-view-change',
  'todo-sort-mode-change',
  'todo-order-change'
])

const internalTodoView = ref('not-started')
const internalTodoSortMode = ref('priority')
const todoListElement = ref(null)
const todoWorkspaceScrollElement = ref(null)
const isTodoReordering = ref(false)
const collapsedTodoIds = ref(new Set())
const knownTodoIds = ref(new Set(props.todos.map((todo) => todo.id)))
const openLaunchTarget = ref({ kind: '', id: '' })
const confirmRemoveTodoProjectId = ref('')
const todoActionConfirm = ref({ todoId: '', action: '' })
const confirmDeleteProjectId = ref('')
const selectedProjectIds = ref(new Set())
const confirmBulkDeleteProjects = ref(false)
const selectedCompletedTodoIds = ref(new Set())
const confirmBulkDeleteCompletedTodos = ref(false)
const todoContextMenu = ref({ todoId: '', x: 0, y: 0 })
const completedTodoMenuId = ref('')
const hoveredTodoId = ref('')
const visibleDescriptionTooltipTodoId = ref('')
const descriptionTooltipPosition = ref({ x: 0, y: 0 })
const hoveredLifecycleErrorKey = ref('')
const visibleLifecycleErrorKey = ref('')
const lifecycleErrorTooltipPosition = ref(null)
const launchMenuPlacement = ref('down')
const launchMenuMaxHeight = ref('')
const launchMenuFixedStyle = ref({})

const descriptionTooltipLayer = createTodoDescriptionTooltipLayer()
let descriptionTooltipTimer = null
let lifecycleErrorTooltipTimer = null
let lifecycleErrorTooltipHideTimer = null
let lifecycleErrorTooltipTrigger = null
let todoSortable = null
let todoReorderPreviousOrder = []
let todoReorderSession = 0
const descriptionTooltipDelayMs = 600
const descriptionTooltipOffset = 12
const lifecycleErrorTooltipDelayMs = 600
const lifecycleErrorTooltipOffset = 12
const launchMenuBorderHeight = 2
const launchMenuGap = 4
const launchMenuMinimumHeight = 32
const launchMenuMinimumWidth = 132
const launchMenuOptionHeight = 32
const launchMenuViewportPadding = 8
const todoContextMenuViewportPadding = 8
const todoContextMenuWidth = 180
const todoContextMenuHeight = 160
const clearedWorktreeLabel = 'worktree已清除'

const terminalLaunchOptions = computed(() => [
  { name: 'Terminal', command: '' },
  ...props.launchProfiles.filter((profile) => profile?.enabled !== false)
])

const todoView = computed(() => normalizedTodoView(props.todoView || internalTodoView.value))
const activeTodoSortMode = computed(() => normalizedTodoSortMode(props.todoSortMode || internalTodoSortMode.value))
const todoInteractionsLocked = computed(() => props.todoOrderSaving || isTodoReordering.value)

const todoPriorityOrder = {
  high: 0,
  medium: 1,
  low: 2
}

const notStartedTodos = computed(() => sortedOpenTodos('not-started'))
const inProgressTodos = computed(() => sortedOpenTodos('in-progress'))
const completedTodos = computed(() => sortedCompletedTodos())
const currentOpenTodos = computed(() => (todoView.value === 'in-progress' ? inProgressTodos.value : notStartedTodos.value))
const currentOpenTodoListTestId = computed(() => `${todoView.value}-todos`)
const isOpenTodoView = computed(() => ['not-started', 'in-progress'].includes(todoView.value))
const activeTodos = computed(() => currentOpenTodos.value)
const activeTodoIds = computed(() => currentOpenTodos.value.map((todo) => todo.id))
const hasActiveTodos = computed(() => activeTodoIds.value.length > 0)
const selectedProjectCount = computed(() => selectedProjectIds.value.size)
const selectedProjectIdsList = computed(() => props.projects.filter((project) => selectedProjectIds.value.has(project.id)).map((project) => project.id))
const selectedCompletedTodoCount = computed(() => selectedCompletedTodoIds.value.size)
const selectedCompletedTodoIdsList = computed(() => {
  const completedTodoIds = new Set(completedTodos.value.map((todo) => todo.id))
  return [...selectedCompletedTodoIds.value].filter((todoId) => completedTodoIds.has(todoId))
})

function sortedOpenTodos(status) {
  return sortedOpenTodosForMode(status, activeTodoSortMode.value)
}

function sortedOpenTodosForMode(status, mode) {
  const entries = props.todos
    .map((todo, index) => ({ todo, index }))
    .filter(({ todo }) => todoWorkflowStatus(todo) === status)
  if (mode === 'manual') {
    const ranks = new Map(manualTodoOrder(status).map((todoId, index) => [todoId, index]))
    entries.sort((left, right) => compareManualTodoEntries(left, right, ranks))
  } else {
    entries.sort((left, right) => compareActiveTodoEntries(left, right, mode))
  }
  return entries.map(({ todo }) => todo)
}

function sortedCompletedTodos() {
  return props.todos
    .map((todo, index) => ({ todo, index }))
    .filter(({ todo }) => todoWorkflowStatus(todo) === 'completed')
    .sort(compareCompletedTodoEntries)
    .map(({ todo }) => todo)
}

function todoWorkflowStatus(todo) {
  if (todo?.status === 'active') {
    return 'not-started'
  }
  return todo?.status || 'not-started'
}

const projectsById = computed(() => {
  const projects = new Map()
  for (const project of props.projects) {
    projects.set(project.id, project)
  }
  return projects
})

const todoProjectsByTodo = computed(() => {
  const groups = new Map()
  for (const todoProject of props.todoProjects) {
    if (!groups.has(todoProject.todoId)) {
      groups.set(todoProject.todoId, [])
    }
    groups.get(todoProject.todoId).push(todoProject)
  }
  return groups
})

const terminalsByTodoProject = computed(() => {
  const groups = new Map()
  for (const terminal of props.terminals) {
    const todoProjectId = terminal.todoProjectId || ''
    if (!todoProjectId) {
      continue
    }
    if (!groups.has(todoProjectId)) {
      groups.set(todoProjectId, [])
    }
    groups.get(todoProjectId).push(terminal)
  }
  return groups
})

const taskTerminalsByTodo = computed(() => {
  const groups = new Map()
  for (const terminal of props.terminals) {
    if (!terminal.todoId || terminal.todoProjectId) {
      continue
    }
    if (!groups.has(terminal.todoId)) {
      groups.set(terminal.todoId, [])
    }
    groups.get(terminal.todoId).push(terminal)
  }
  return groups
})

const lifecycleStatusesByTodo = computed(() => {
  const groups = new Map()
  for (const status of props.lifecycleScriptStatuses) {
    if (!status?.todoId || !status.status) {
      continue
    }
    if (!groups.has(status.todoId)) {
      groups.set(status.todoId, [])
    }
    groups.get(status.todoId).push(status)
  }
  return groups
})

function todoProjectsForTodo(todoId) {
  return todoProjectsByTodo.value.get(todoId) || []
}

function projectForTodoProject(todoProject) {
  if (!todoProject) {
    return null
  }
  if (todoProject.name || todoProject.path) {
    return {
      id: todoProject.projectId,
      name: todoProject.name || 'Missing project',
      path: todoProject.path || todoProject.projectId,
      available: todoProject.available !== false
    }
  }
  return projectsById.value.get(todoProject.projectId) || null
}

function todoProjectDisplayName(todoProject) {
  const projectName = projectForTodoProject(todoProject)?.name || 'Missing project'
  if (todoProjectWorktreeCleared(todoProject)) {
    return `${projectName}(${clearedWorktreeLabel})`
  }
  const branch = (props.todoProjectBranches?.[todoProject.id] || '').trim()
  return branch ? `${projectName}(${branch})` : projectName
}

function todoProjectTerminals(todoProjectId) {
  return terminalsByTodoProject.value.get(todoProjectId) || []
}

function taskTerminalsForTodo(todoId) {
  return taskTerminalsByTodo.value.get(todoId) || []
}

function hasTaskTerminals(todoId) {
  return taskTerminalsForTodo(todoId).length > 0
}

function lifecycleStatusesForTodo(todo) {
  return lifecycleStatusesByTodo.value.get(todo.id) || []
}

function lifecycleScriptPhaseLabel(phase) {
  return phase === 'complete' ? '完成脚本' : '初始化脚本'
}

function lifecycleScriptStatusLabel(status) {
  const phaseLabel = lifecycleScriptPhaseLabel(status.phase)
  if (status.status === 'failed') {
    return `${phaseLabel}失败`
  }
  return `${phaseLabel}执行中`
}

function lifecycleScriptStatusMessage(status) {
  return status.outputTail || status.message || ''
}

function retryLifecycleScript(todoId, phase) {
  hideLifecycleErrorTooltip()
  emit('retry-todo-lifecycle-script', todoId, phase)
}

function lifecycleScriptErrorKey(status) {
  const runIdentity = status.runId || `${status.startedAt || ''}:${status.finishedAt || ''}`
  return `${status.scopeEpoch || 0}:${status.todoId}:${status.phase}:${runIdentity}`
}

function lifecycleScriptErrorOutput(status) {
  return props.lifecycleScriptErrorOutputs[lifecycleScriptErrorKey(status)] || ''
}

function clearLifecycleErrorTooltipTimer() {
  if (lifecycleErrorTooltipTimer !== null) {
    clearTimeout(lifecycleErrorTooltipTimer)
    lifecycleErrorTooltipTimer = null
  }
}

function clearLifecycleErrorTooltipHideTimer() {
  if (lifecycleErrorTooltipHideTimer !== null) {
    clearTimeout(lifecycleErrorTooltipHideTimer)
    lifecycleErrorTooltipHideTimer = null
  }
}

function lifecycleErrorTooltipStyle() {
  if (!lifecycleErrorTooltipPosition.value) {
    return { visibility: 'hidden' }
  }
  return {
    left: `${lifecycleErrorTooltipPosition.value.x}px`,
    top: `${lifecycleErrorTooltipPosition.value.y}px`
  }
}

function calculateLifecycleErrorTooltipPosition(triggerRect, tooltipRect) {
  const viewportWidth = window.innerWidth || 1024
  const viewportHeight = window.innerHeight || 768
  const tooltipWidth = tooltipRect.width
  const tooltipHeight = tooltipRect.height
  const maxLeft = Math.max(lifecycleErrorTooltipOffset, viewportWidth - tooltipWidth - lifecycleErrorTooltipOffset)
  const maxTop = Math.max(lifecycleErrorTooltipOffset, viewportHeight - tooltipHeight - lifecycleErrorTooltipOffset)
  const belowTop = triggerRect.bottom + lifecycleErrorTooltipOffset
  const aboveTop = triggerRect.top - tooltipHeight - lifecycleErrorTooltipOffset
  const desiredTop = belowTop + tooltipHeight <= viewportHeight - lifecycleErrorTooltipOffset
    ? belowTop
    : aboveTop
  return {
    x: clampNumber(triggerRect.left, lifecycleErrorTooltipOffset, maxLeft),
    y: clampNumber(desiredTop, lifecycleErrorTooltipOffset, maxTop)
  }
}

async function updateLifecycleErrorTooltipPosition(key) {
  await nextTick()
  if (visibleLifecycleErrorKey.value !== key || !lifecycleErrorTooltipTrigger) {
    return
  }
  const tooltip = descriptionTooltipLayer.querySelector('.todo-lifecycle-error-tooltip')
  if (!tooltip) {
    return
  }
  lifecycleErrorTooltipPosition.value = calculateLifecycleErrorTooltipPosition(
    lifecycleErrorTooltipTrigger.getBoundingClientRect(),
    tooltip.getBoundingClientRect()
  )
}

function startLifecycleErrorTooltip(status, event) {
  hideLifecycleErrorTooltip()
  if (status.status !== 'failed') {
    return
  }
  const key = lifecycleScriptErrorKey(status)
  hoveredLifecycleErrorKey.value = key
  lifecycleErrorTooltipTrigger = event?.currentTarget || null
  lifecycleErrorTooltipPosition.value = null
  lifecycleErrorTooltipTimer = setTimeout(() => {
    lifecycleErrorTooltipTimer = null
    if (hoveredLifecycleErrorKey.value !== key) {
      return
    }
    visibleLifecycleErrorKey.value = key
    emit('request-todo-lifecycle-script-error', status.todoId, status.phase)
  }, lifecycleErrorTooltipDelayMs)
}

function scheduleHideLifecycleErrorTooltip() {
  clearLifecycleErrorTooltipTimer()
  clearLifecycleErrorTooltipHideTimer()
  lifecycleErrorTooltipHideTimer = setTimeout(hideLifecycleErrorTooltip, 100)
}

function keepLifecycleErrorTooltipVisible() {
  clearLifecycleErrorTooltipHideTimer()
}

function hideLifecycleErrorTooltip() {
  clearLifecycleErrorTooltipTimer()
  clearLifecycleErrorTooltipHideTimer()
  hoveredLifecycleErrorKey.value = ''
  visibleLifecycleErrorKey.value = ''
  lifecycleErrorTooltipTrigger = null
  lifecycleErrorTooltipPosition.value = null
}

function isLifecycleErrorTooltipVisible(status) {
  return (
    status.status === 'failed' &&
    visibleLifecycleErrorKey.value === lifecycleScriptErrorKey(status) &&
    Boolean(lifecycleScriptErrorOutput(status))
  )
}

function hasTodoProjectTerminals(todoProjectId) {
  return todoProjectTerminals(todoProjectId).length > 0
}

function activeTaskTerminalTodoId() {
  const terminal = props.terminals.find((candidate) => (
    candidate.id === props.activeTerminalId &&
    !candidate.todoProjectId &&
    !candidate.workspaceTerminal
  ))
  return terminal?.todoId || ''
}

function isTodoActive(todo) {
  const taskTodoId = activeTaskTerminalTodoId()
  if (taskTodoId) {
    return todo.id === taskTodoId
  }
  return todo.id === props.activeTodoId
}

function todoHasActiveTerminal(todo) {
  return props.terminals.some((terminal) => (
    terminal.id === props.activeTerminalId &&
    terminal.todoId === todo.id
  ))
}

function todoProjectHasActiveTerminal(todoProject) {
  return todoProjectTerminals(todoProject.id).some((terminal) => terminal.id === props.activeTerminalId)
}

function todoProjectWorktreeFailed(todoProject) {
  return todoProject?.worktreeStatus === 'failed'
}

function todoProjectWorktreeCleared(todoProject) {
  return todoProject?.worktreeStatus === 'cleared'
}

function todoProjectCanCreateTerminal(todoProject) {
  if (todoProjectWorktreeFailed(todoProject)) {
    return false
  }
  const worktreeStatus = (todoProject?.worktreeStatus || '').trim()
  return worktreeStatus === '' || worktreeStatus === 'ready' || worktreeStatus === 'cleared'
}

function isTodoCollapsed(todoId) {
  return collapsedTodoIds.value.has(todoId)
}

function isTodoExpanded(todoId) {
  return !isTodoCollapsed(todoId)
}

function todoProjectListId(todoId) {
  return `todo-project-list-${todoId}`
}

function terminalListId(todoProjectId) {
  return `terminal-list-${todoProjectId}`
}

function toggleTodoBranch(todoId) {
  if (todoInteractionsLocked.value) {
    return
  }
  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  if (nextCollapsedTodoIds.has(todoId)) {
    nextCollapsedTodoIds.delete(todoId)
    emit('todo-expanded', todoId)
  } else {
    nextCollapsedTodoIds.add(todoId)
  }
  collapsedTodoIds.value = nextCollapsedTodoIds
}

function collapseAllTodos() {
  if (!hasActiveTodos.value || todoInteractionsLocked.value) {
    return
  }

  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  for (const todoId of activeTodoIds.value) {
    nextCollapsedTodoIds.add(todoId)
  }
  collapsedTodoIds.value = nextCollapsedTodoIds
}

function expandAllTodos() {
  if (!hasActiveTodos.value || todoInteractionsLocked.value) {
    return
  }

  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  const expandedTodoIds = []
  for (const todoId of activeTodoIds.value) {
    if (nextCollapsedTodoIds.delete(todoId)) {
      expandedTodoIds.push(todoId)
    }
  }
  collapsedTodoIds.value = nextCollapsedTodoIds
  for (const todoId of expandedTodoIds) {
    emit('todo-expanded', todoId)
  }
}

function expandTodo(todoId) {
  if (!todoId || !collapsedTodoIds.value.has(todoId)) {
    return
  }

  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  nextCollapsedTodoIds.delete(todoId)
  collapsedTodoIds.value = nextCollapsedTodoIds
  emit('todo-expanded', todoId)
}

function selectTodoProject(todoProject) {
  hideTodoDescriptionTooltip()
  expandTodo(todoProject.todoId)
  emit('select-todo-project', todoProject.id)
}

function toggleTerminalLaunchMenu(todoProject, event) {
  if (!todoProjectCanCreateTerminal(todoProject)) {
    return
  }
  hideTodoDescriptionTooltip()
  expandTodo(todoProject.todoId)
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeTodoContextMenu()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  if (isTerminalLaunchMenuOpen('project', todoProject.id)) {
    closeTerminalLaunchMenu()
    return
  }

  updateTerminalLaunchMenuPlacement(event?.currentTarget)
  openLaunchTarget.value = { kind: 'project', id: todoProject.id }
}

function toggleTaskTerminalLaunchMenu(todoId, event) {
  hideTodoDescriptionTooltip()
  expandTodo(todoId)
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeTodoContextMenu()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  if (isTerminalLaunchMenuOpen('task', todoId)) {
    closeTerminalLaunchMenu()
    return
  }

  updateTerminalLaunchMenuPlacement(event?.currentTarget)
  openLaunchTarget.value = { kind: 'task', id: todoId }
}

function closeTerminalLaunchMenu() {
  openLaunchTarget.value = { kind: '', id: '' }
  resetTerminalLaunchMenuPlacement()
}

function isTerminalLaunchMenuOpen(kind, id) {
  return openLaunchTarget.value.kind === kind && openLaunchTarget.value.id === id
}

function openTodoProjectRemovePopover(todoProjectId) {
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeTodoContextMenu()
  confirmRemoveTodoProjectId.value = todoProjectId
}

function closeTodoProjectRemovePopover() {
  confirmRemoveTodoProjectId.value = ''
}

function confirmTodoProjectRemoval(todoProjectId) {
  emit('remove-todo-project', todoProjectId)
  closeTodoProjectRemovePopover()
}

function openTodoActionPopover(todoId, action) {
  if (todoInteractionsLocked.value) {
    return
  }
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeProjectDeletePopover()
  closeTodoContextMenu()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  todoActionConfirm.value = { todoId, action }
}

function closeTodoActionPopover() {
  todoActionConfirm.value = { todoId: '', action: '' }
}

function openTodoContextMenu(todoId, event) {
  if (todoInteractionsLocked.value) {
    return
  }
  event.preventDefault()
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  todoContextMenu.value = { todoId, ...todoContextMenuPlacement(event.clientX, event.clientY) }
}

function openTodoContextMenuFromButton(todoId, event) {
  if (todoInteractionsLocked.value) {
    return
  }
  event.stopPropagation()
  const rect = event.currentTarget.getBoundingClientRect()
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  todoContextMenu.value = { todoId, ...todoContextMenuPlacement(rect.left, rect.bottom) }
}

function closeTodoContextMenu() {
  todoContextMenu.value = { todoId: '', x: 0, y: 0 }
}

function openCompletedTodoMenu(todoId, event) {
  event.stopPropagation()
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeBulkCompletedTodoDeletePopover()
  closeTodoContextMenu()
  completedTodoMenuId.value = todoId
}

function closeCompletedTodoMenu() {
  completedTodoMenuId.value = ''
}

function isCompletedTodoMenuOpen(todoId) {
  return completedTodoMenuId.value === todoId
}

function isTodoContextMenuOpen(todoId) {
  return todoContextMenu.value.todoId === todoId
}

function todoContextMenuStyle() {
  return {
    left: `${todoContextMenu.value.x}px`,
    top: `${todoContextMenu.value.y}px`
  }
}

function todoContextMenuPlacement(x, y) {
  return {
    x: clampToViewport(x, todoContextMenuWidth, 'innerWidth'),
    y: clampToViewport(y, todoContextMenuHeight, 'innerHeight')
  }
}

function clampToViewport(value, size, dimension) {
  const viewportSize = typeof window !== 'undefined' ? Number(window[dimension]) || 0 : 0
  const max = Math.max(todoContextMenuViewportPadding, viewportSize - size - todoContextMenuViewportPadding)
  return Math.min(Math.max(Number(value) || 0, todoContextMenuViewportPadding), max)
}

function isTodoActionPopoverOpen(todoId, action) {
  return todoActionConfirm.value.todoId === todoId && todoActionConfirm.value.action === action
}

function confirmTodoAction(todoId, action) {
  if (action === 'complete') {
    emit('complete-todo', todoId)
  } else if (action === 'delete') {
    emit('delete-todo', todoId)
  }
  closeTodoActionPopover()
}

function changeTodoStatus(todoId, status) {
  if (todoInteractionsLocked.value) {
    return
  }
  hideTodoDescriptionTooltip()
  setTodoView(status)
  emit('change-todo-status', todoId, status)
}

function createTaskTerminal(todoId, option = null) {
  hideTodoDescriptionTooltip()
  expandTodo(todoId)
  emit('create-task-terminal', todoId, option?.command ? option : null)
}

function setTodoView(view) {
  const nextView = normalizedTodoView(view)
  internalTodoView.value = nextView
  emit('update:todo-view', nextView)
  emit('todo-view-change', nextView)
}

function normalizedTodoView(view) {
  return ['not-started', 'in-progress', 'completed'].includes(view) ? view : 'not-started'
}

function openProjectDeletePopover(projectId) {
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeBulkProjectDeletePopover()
  closeTodoContextMenu()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  confirmDeleteProjectId.value = projectId
}

function closeProjectDeletePopover() {
  confirmDeleteProjectId.value = ''
}

function confirmProjectDeletion(projectId) {
  emit('delete-project', projectId)
  closeProjectDeletePopover()
}

function isProjectSelected(projectId) {
  return selectedProjectIds.value.has(projectId)
}

function toggleProjectSelection(projectId) {
  const nextSelectedProjectIds = new Set(selectedProjectIds.value)
  if (nextSelectedProjectIds.has(projectId)) {
    nextSelectedProjectIds.delete(projectId)
  } else {
    nextSelectedProjectIds.add(projectId)
  }
  selectedProjectIds.value = nextSelectedProjectIds
  if (nextSelectedProjectIds.size === 0) {
    closeBulkProjectDeletePopover()
  }
}

function openBulkProjectDeletePopover() {
  if (selectedProjectCount.value === 0) {
    return
  }
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeTodoContextMenu()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
  confirmBulkDeleteProjects.value = true
}

function closeBulkProjectDeletePopover() {
  confirmBulkDeleteProjects.value = false
}

function clearProjectSelection() {
  selectedProjectIds.value = new Set()
  closeBulkProjectDeletePopover()
}

function confirmBulkProjectDeletion() {
  const projectIds = selectedProjectIdsList.value
  if (projectIds.length === 0) {
    closeBulkProjectDeletePopover()
    return
  }
  emit('delete-projects', projectIds)
  clearProjectSelection()
}

function isCompletedTodoSelected(todoId) {
  return selectedCompletedTodoIds.value.has(todoId)
}

function toggleCompletedTodoSelection(todoId) {
  const nextSelectedTodoIds = new Set(selectedCompletedTodoIds.value)
  if (nextSelectedTodoIds.has(todoId)) {
    nextSelectedTodoIds.delete(todoId)
  } else {
    nextSelectedTodoIds.add(todoId)
  }
  selectedCompletedTodoIds.value = nextSelectedTodoIds
  if (nextSelectedTodoIds.size === 0) {
    closeBulkCompletedTodoDeletePopover()
  }
}

function openBulkCompletedTodoDeletePopover() {
  if (selectedCompletedTodoCount.value === 0) {
    return
  }
  hideTodoDescriptionTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeTodoContextMenu()
  closeCompletedTodoMenu()
  confirmBulkDeleteCompletedTodos.value = true
}

function closeBulkCompletedTodoDeletePopover() {
  confirmBulkDeleteCompletedTodos.value = false
}

function clearCompletedTodoSelection() {
  selectedCompletedTodoIds.value = new Set()
  closeBulkCompletedTodoDeletePopover()
}

function confirmBulkCompletedTodoDeletion() {
  const todoIds = selectedCompletedTodoIdsList.value
  if (todoIds.length === 0) {
    closeBulkCompletedTodoDeletePopover()
    return
  }
  emit('delete-completed-todos', todoIds)
  clearCompletedTodoSelection()
}

function closeFloatingMenus() {
  hideTodoDescriptionTooltip()
  hideLifecycleErrorTooltip()
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
  closeTodoContextMenu()
  closeBulkCompletedTodoDeletePopover()
  closeCompletedTodoMenu()
}

function selectTerminalLaunchOption(todoProject, option) {
  if (!todoProjectCanCreateTerminal(todoProject)) {
    closeTerminalLaunchMenu()
    return
  }
  hideTodoDescriptionTooltip()
  expandTodo(todoProject.todoId)
  emit('create-terminal', todoProject.id, option.command ? option : null)
  closeTerminalLaunchMenu()
}

function selectTaskTerminalLaunchOption(todoId, option) {
  hideTodoDescriptionTooltip()
  expandTodo(todoId)
  emit('create-task-terminal', todoId, option.command ? option : null)
  closeTerminalLaunchMenu()
}

function todoDescription(todo) {
  return (todo?.description || '').trim()
}

function clearTodoDescriptionTooltipTimer() {
  if (descriptionTooltipTimer !== null) {
    clearTimeout(descriptionTooltipTimer)
    descriptionTooltipTimer = null
  }
}

function showTodoDescriptionTooltip(todo) {
  const description = todoDescription(todo)
  if (!description || hoveredTodoId.value !== todo.id) {
    return
  }
  visibleDescriptionTooltipTodoId.value = todo.id
}

function descriptionTooltipStyle() {
  return {
    left: `${descriptionTooltipPosition.value.x}px`,
    top: `${descriptionTooltipPosition.value.y}px`
  }
}

function todoDescriptionTooltipPosition(trigger) {
  const rect = trigger?.getBoundingClientRect?.()
  const viewportWidth = window.innerWidth || 1024
  const tooltipWidth = Math.min(520, viewportWidth * 0.72)
  const maxLeft = Math.max(descriptionTooltipOffset, viewportWidth - tooltipWidth - descriptionTooltipOffset)
  const desiredLeft = (rect?.left || 0) + 34
  return {
    x: Math.min(Math.max(desiredLeft, descriptionTooltipOffset), maxLeft),
    y: (rect?.bottom || 0) + descriptionTooltipOffset
  }
}

function startTodoDescriptionTooltip(todo, event) {
  hideTodoDescriptionTooltip()
  const description = todoDescription(todo)
  if (!description) {
    return
  }

  hoveredTodoId.value = todo.id
  descriptionTooltipPosition.value = todoDescriptionTooltipPosition(event?.currentTarget)
  descriptionTooltipTimer = setTimeout(() => {
    descriptionTooltipTimer = null
    showTodoDescriptionTooltip(todo)
  }, descriptionTooltipDelayMs)
}

function hideTodoDescriptionTooltip() {
  clearTodoDescriptionTooltipTimer()
  hoveredTodoId.value = ''
  visibleDescriptionTooltipTodoId.value = ''
  descriptionTooltipPosition.value = { x: 0, y: 0 }
}

function isTodoDescriptionTooltipVisible(todo) {
  return visibleDescriptionTooltipTodoId.value === todo.id && Boolean(todoDescription(todo))
}

function createTodoDescriptionTooltipLayer() {
  const layer = document.createElement('div')
  layer.className = 'todo-description-tooltip-layer'
  document.body.appendChild(layer)
  return layer
}

function resetTerminalLaunchMenuPlacement() {
  launchMenuPlacement.value = 'down'
  launchMenuMaxHeight.value = ''
  launchMenuFixedStyle.value = {}
}

function updateTerminalLaunchMenuPlacement(trigger) {
  const projectList = trigger?.closest?.('.project-list')
  const triggerRect = trigger?.getBoundingClientRect?.()
  const listRect = projectList?.getBoundingClientRect?.()
  if (!triggerRect || !listRect || listRect.height <= 0) {
    resetTerminalLaunchMenuPlacement()
    return
  }

  const desiredMenuHeight = terminalLaunchOptions.value.length * launchMenuOptionHeight + launchMenuBorderHeight
  const spaceBelow = Math.max(0, listRect.bottom - triggerRect.bottom - launchMenuViewportPadding)
  const spaceAbove = Math.max(0, triggerRect.top - listRect.top - launchMenuViewportPadding)
  const opensUp = spaceBelow < desiredMenuHeight && spaceAbove > spaceBelow
  const availableHeight = opensUp ? spaceAbove : spaceBelow
  const menuHeight = Math.min(desiredMenuHeight, Math.max(availableHeight, launchMenuMinimumHeight))
  const minLeft = listRect.left + launchMenuViewportPadding
  const maxLeft = Math.max(minLeft, listRect.right - launchMenuMinimumWidth - launchMenuViewportPadding)
  const left = clampNumber(triggerRect.right - launchMenuMinimumWidth, minLeft, maxLeft)
  const rawTop = opensUp
    ? triggerRect.top - menuHeight - launchMenuGap
    : triggerRect.bottom + launchMenuGap
  const minTop = listRect.top + launchMenuViewportPadding
  const maxTop = Math.max(minTop, listRect.bottom - menuHeight - launchMenuViewportPadding)
  const top = clampNumber(rawTop, minTop, maxTop)

  launchMenuPlacement.value = opensUp ? 'up' : 'down'
  launchMenuMaxHeight.value =
    availableHeight < desiredMenuHeight
      ? `${Math.max(availableHeight, launchMenuMinimumHeight)}px`
      : ''
  launchMenuFixedStyle.value = {
    position: 'fixed',
    left: `${Math.round(left)}px`,
    top: `${Math.round(top)}px`,
    right: 'auto',
    bottom: 'auto'
  }
}

function clampNumber(value, min, max) {
  return Math.min(Math.max(value, min), max)
}

function terminalLaunchMenuClass() {
  return {
    'terminal-launch-menu--up': launchMenuPlacement.value === 'up',
    'terminal-launch-menu--down': launchMenuPlacement.value !== 'up',
    'terminal-launch-menu--constrained': Boolean(launchMenuMaxHeight.value)
  }
}

function terminalLaunchMenuStyle() {
  return {
    ...launchMenuFixedStyle.value,
    ...(launchMenuMaxHeight.value ? { maxHeight: launchMenuMaxHeight.value } : {})
  }
}

function terminalDisplayName(terminal) {
  return terminal.currentCommand || terminal.shellName || 'shell'
}

function compareActiveTodoEntries(left, right, mode = activeTodoSortMode.value) {
  if (mode === 'time') {
    return compareActiveTodosByTime(left, right)
  }
  return compareActiveTodosByPriority(left, right)
}

function compareManualTodoEntries(left, right, ranks) {
  const leftRank = ranks.has(left.todo.id) ? ranks.get(left.todo.id) : Number.MAX_SAFE_INTEGER
  const rightRank = ranks.has(right.todo.id) ? ranks.get(right.todo.id) : Number.MAX_SAFE_INTEGER
  if (leftRank !== rightRank) {
    return leftRank - rightRank
  }
  return left.index - right.index
}

function compareActiveTodosByPriority(left, right) {
  const priorityDiff = todoPriorityRank(left.todo) - todoPriorityRank(right.todo)
  if (priorityDiff !== 0) {
    return priorityDiff
  }

  return compareActiveTodosByTime(left, right)
}

function compareActiveTodosByTime(left, right) {
  const createdAtDiff = compareTodoCreatedAt(left.todo, right.todo)
  if (createdAtDiff !== 0) {
    return createdAtDiff
  }

  return left.index - right.index
}

function compareCompletedTodoEntries(left, right) {
  const leftTimestamp = completedTodoTimestamp(left.todo)
  const rightTimestamp = completedTodoTimestamp(right.todo)
  if (leftTimestamp > rightTimestamp) {
    return -1
  }
  if (leftTimestamp < rightTimestamp) {
    return 1
  }

  return left.index - right.index
}

function setActiveTodoSortMode(mode) {
  if (!['priority', 'time', 'manual'].includes(mode) || todoInteractionsLocked.value) {
    return
  }
  const previousMode = activeTodoSortMode.value
  internalTodoSortMode.value = mode
  const change = { mode }
  if (mode === 'manual' && !props.todoOrdersInitialized) {
    change.todoOrders = {
      notStarted: sortedOpenTodosForMode('not-started', previousMode).map((todo) => todo.id),
      inProgress: sortedOpenTodosForMode('in-progress', previousMode).map((todo) => todo.id)
    }
  }
  emit('todo-sort-mode-change', change)
}

function moveTodoByKeyboard(todoId, direction) {
  if (todoInteractionsLocked.value || activeTodoSortMode.value !== 'manual' || !isOpenTodoView.value) {
    return
  }
  const previousOrder = currentOpenTodos.value.map((todo) => todo.id)
  const currentIndex = previousOrder.indexOf(todoId)
  const nextIndex = currentIndex + direction
  if (currentIndex < 0 || nextIndex < 0 || nextIndex >= previousOrder.length) {
    return
  }
  const order = [...previousOrder]
  const [movedTodoId] = order.splice(currentIndex, 1)
  order.splice(nextIndex, 0, movedTodoId)
  emit('todo-order-change', {
    status: todoView.value,
    previousOrder,
    order
  })
}

function normalizedTodoSortMode(mode) {
  return ['priority', 'time', 'manual'].includes(mode) ? mode : 'priority'
}

function manualTodoOrder(status) {
  return status === 'in-progress' ? props.todoOrders?.inProgress || [] : props.todoOrders?.notStarted || []
}

function todoPriorityRank(todo) {
  return todoPriorityOrder[todoPriority(todo)]
}

function todoCreatedAtTimestamp(todo) {
  const timestamp = Date.parse(todo?.createdAt || '')
  return Number.isNaN(timestamp) ? Number.POSITIVE_INFINITY : timestamp
}

function completedTodoTimestamp(todo) {
  for (const value of [todo?.completedAt, todo?.archivedAt]) {
    const timestamp = Date.parse(value || '')
    if (!Number.isNaN(timestamp)) {
      return timestamp
    }
  }
  return Number.NEGATIVE_INFINITY
}

function compareTodoCreatedAt(left, right) {
  const leftTimestamp = todoCreatedAtTimestamp(left)
  const rightTimestamp = todoCreatedAtTimestamp(right)
  if (leftTimestamp < rightTimestamp) {
    return -1
  }
  if (leftTimestamp > rightTimestamp) {
    return 1
  }
  return 0
}

function terminalActivityState(terminal) {
  if (terminal.attentionState === 'needs-ack') {
    return 'needs-ack'
  }
  const state = terminal.activityState || terminal.agentStatus?.phase || 'idle'
  return ['busy', 'needs-input', 'needs-ack'].includes(state) ? state : 'idle'
}

function activityStateLabel(state) {
  if (state === 'busy') {
    return 'Running'
  }
  if (state === 'needs-input') {
    return 'Needs input'
  }
  if (state === 'needs-ack') {
    return 'Review needed'
  }
  return 'Idle'
}

function terminalActivityLabel(terminal) {
  return activityStateLabel(terminalActivityState(terminal))
}

function terminalRowLabel(terminal) {
  const activityLabel = terminalActivityLabel(terminal)
  const displayName = terminalDisplayName(terminal)
  return activityLabel === 'Idle' ? displayName : `${displayName} - ${activityLabel}`
}

function todoActivityState(todo) {
  let hasAckTerminal = false
  let hasBusyTerminal = false
  let hasTerminal = false
  for (const terminal of props.terminals) {
    if (terminal.todoId !== todo.id) {
      continue
    }
    hasTerminal = true
    const state = terminalActivityState(terminal)
    if (state === 'needs-input') {
      return 'needs-input'
    }
    if (state === 'needs-ack') {
      hasAckTerminal = true
    }
    if (state === 'busy') {
      hasBusyTerminal = true
    }
  }
  if (!hasTerminal) {
    return ''
  }
  if (hasAckTerminal) {
    return 'needs-ack'
  }
  return hasBusyTerminal ? 'busy' : 'idle'
}

function collapsedTodoActivityState(todo) {
  return isTodoCollapsed(todo.id) ? todoActivityState(todo) : ''
}

function collapsedTodoFeedbackState(todo) {
  const state = collapsedTodoActivityState(todo)
  return ['busy', 'needs-input', 'needs-ack'].includes(state) ? state : ''
}

function collapsedTodoActivityClass(todo) {
  const state = collapsedTodoFeedbackState(todo)
  return state ? `todo-activity-${state}` : ''
}

function collapsedTodoActivityLabel(todo) {
  return activityStateLabel(collapsedTodoActivityState(todo))
}

function todoPriority(todo) {
  return ['high', 'medium', 'low'].includes(todo?.priority) ? todo.priority : 'medium'
}

function todoPriorityClass(todo) {
  return `todo-header-row-priority-${todoPriority(todo)}`
}

function completedAtLabel(todo) {
  return todo.completedAt || todo.archivedAt || 'No completion time'
}

function completedDurationLabel(todo) {
  const startedAt = Date.parse(todo?.startedAt || '')
  const completedAt = Date.parse(todo?.completedAt || '')
  if (Number.isNaN(startedAt) || Number.isNaN(completedAt) || completedAt < startedAt) {
    return ''
  }
  return `Duration ${formatDuration(completedAt - startedAt)}`
}

function formatDuration(durationMs) {
  const totalSeconds = Math.floor(durationMs / 1000)
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (days > 0) {
    return `${days}d ${hours}h`
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}

function completedSnapshotKey(todo, snapshot, index) {
  return [todo?.id || '', snapshot?.projectId || '', snapshot?.path || '', index].join('::')
}

function completedSnapshotTestId(todo, snapshot, index) {
  return `${todo?.id || 'todo'}-${snapshot?.projectId || 'project'}-${index}`
}

function completedSnapshotBranchLabel(snapshot) {
  const worktreeBranch = (snapshot?.worktreeBranch || '').trim() || 'Unknown branch'
  const baseBranch = (snapshot?.baseBranch || '').trim() || 'Unknown base'
  return `${worktreeBranch} -> ${baseBranch}`
}

function completedMergeStatus(todo, snapshot, index) {
  const key = completedSnapshotKey(todo, snapshot, index)
  const status = props.completedMergeStatuses[key]?.status || ''
  if (status === 'confirmed') {
    return 'merged'
  }
  if (['merged', 'unmerged', 'unknown'].includes(status)) {
    return status
  }
  if (!snapshot?.worktreeBranch || !snapshot?.baseBranch) {
    return 'unknown'
  }
  return 'checking'
}

function completedMergeStatusTitle(status) {
  if (status === 'merged') {
    return 'Merged'
  }
  if (status === 'unmerged') {
    return 'Not merged'
  }
  if (status === 'checking') {
    return 'Checking merge status'
  }
  return 'Merge status unknown'
}

function syncTodoSortable() {
  destroyTodoSortable()
  if (!isOpenTodoView.value || activeTodoSortMode.value !== 'manual' || !todoListElement.value) {
    return
  }
  todoSortable = createTodoSortable(todoListElement.value, {
    group: false,
    handle: '.todo-drag-handle',
    draggable: '.todo-node',
    dataIdAttr: 'data-id',
    forceFallback: true,
    fallbackOnBody: false,
    fallbackTolerance: 4,
    animation: 150,
    ghostClass: 'todo-sortable-ghost',
    chosenClass: 'todo-sortable-chosen',
    dragClass: 'todo-sortable-drag',
    disabled: props.todoOrderSaving,
    scroll: todoWorkspaceScrollElement.value,
    scrollSensitivity: 36,
    scrollSpeed: 12,
    onChoose: beginTodoReordering,
    onStart: captureTodoReorderStart,
    onEnd: finishTodoReordering,
    onUnchoose: cancelTodoReordering
  })
}

function destroyTodoSortable() {
  cleanupTodoReordering()
  todoSortable?.destroy()
  todoSortable = null
}

function beginTodoReordering() {
  todoReorderSession += 1
  closeFloatingMenus()
  isTodoReordering.value = true
}

function captureTodoReorderStart() {
  todoReorderPreviousOrder = currentOpenTodos.value.map((todo) => todo.id)
}

function finishTodoReordering(event = {}) {
  const previousOrder = [...todoReorderPreviousOrder]
  if (isTodoReorderCancelled(event)) {
    restoreTodoOrderInDOM(previousOrder)
    cleanupTodoReordering()
    return
  }
  const order = todoOrderFromDOM()
  const status = todoView.value
  const changed = order.length === previousOrder.length && order.some((todoId, index) => todoId !== previousOrder[index])
  cleanupTodoReordering()
  if (!changed) {
    return
  }
  emit('todo-order-change', {
    status,
    previousOrder,
    order
  })
}

function isTodoReorderCancelled(event) {
  return ['pointercancel', 'touchcancel', 'dragend'].includes(event?.originalEvent?.type)
}

function restoreTodoOrderInDOM(order) {
  if (!todoListElement.value) {
    return
  }
  const elementsByTodoId = new Map(
    Array.from(todoListElement.value.children).map((element) => [element.dataset.id || '', element])
  )
  for (const todoId of order) {
    const element = elementsByTodoId.get(todoId)
    if (element) {
      todoListElement.value.appendChild(element)
    }
  }
}

function cancelTodoReordering() {
  const session = todoReorderSession
  isTodoReordering.value = false
  queueMicrotask(() => {
    if (session === todoReorderSession) {
      cleanupTodoReordering()
    }
  })
}

function cleanupTodoReordering() {
  todoReorderSession += 1
  isTodoReordering.value = false
  todoReorderPreviousOrder = []
}

function todoOrderFromDOM() {
  if (!todoListElement.value) {
    return []
  }
  return Array.from(todoListElement.value.children)
    .filter((element) => element.classList.contains('todo-node'))
    .map((element) => element.dataset.id || '')
    .filter(Boolean)
}

onMounted(() => {
  window.addEventListener('click', closeFloatingMenus)
  window.addEventListener('resize', closeTerminalLaunchMenu)
  window.addEventListener('scroll', closeTerminalLaunchMenu, true)
  void nextTick(syncTodoSortable)
})

onBeforeUnmount(() => {
  hideTodoDescriptionTooltip()
  hideLifecycleErrorTooltip()
  descriptionTooltipLayer.remove()
  destroyTodoSortable()
  window.removeEventListener('click', closeFloatingMenus)
  window.removeEventListener('resize', closeTerminalLaunchMenu)
  window.removeEventListener('scroll', closeTerminalLaunchMenu, true)
})

watch(
  [todoView, activeTodoSortMode, () => props.todoOrderSaving],
  () => void nextTick(syncTodoSortable)
)

watch(
  () => props.todos,
  (todos) => {
    const nextKnownTodoIds = new Set(knownTodoIds.value)
    const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
    let changed = false
    for (const todo of todos) {
      if (nextKnownTodoIds.has(todo.id)) {
        continue
      }
      nextKnownTodoIds.add(todo.id)
      if (isOpenTodoStatus(todoWorkflowStatus(todo))) {
        nextCollapsedTodoIds.add(todo.id)
        changed = true
      }
    }
    knownTodoIds.value = nextKnownTodoIds
    if (changed) {
      collapsedTodoIds.value = nextCollapsedTodoIds
    }
  },
  { deep: true }
)

watch(
  () => props.lifecycleScriptStatuses,
  (statuses) => {
    const activeKey = visibleLifecycleErrorKey.value || hoveredLifecycleErrorKey.value
    if (!activeKey) {
      return
    }
    const remainsFailed = statuses.some((status) => (
      status.status === 'failed' && lifecycleScriptErrorKey(status) === activeKey
    ))
    if (!remainsFailed) {
      hideLifecycleErrorTooltip()
    }
  }
)

watch(
  () => props.lifecycleScriptErrorScope,
  () => hideLifecycleErrorTooltip()
)

watch(
  () => props.lifecycleScriptErrorOutputs[visibleLifecycleErrorKey.value] || '',
  (output) => {
    const key = visibleLifecycleErrorKey.value
    lifecycleErrorTooltipPosition.value = null
    if (key && output) {
      void updateLifecycleErrorTooltipPosition(key)
    }
  }
)

function isOpenTodoStatus(status) {
  return status === 'not-started' || status === 'in-progress'
}

watch(
  [() => props.activeTodoId, () => props.activeTodoProjectId],
  ([todoId, todoProjectId]) => {
    if (todoId) {
      expandTodo(todoId)
      return
    }

    const todoProject = props.todoProjects.find((candidate) => candidate.id === todoProjectId)
    if (todoProject) {
      expandTodo(todoProject.todoId)
    }
  },
  { immediate: true }
)

watch(
  () => props.activeTerminalId,
  (terminalId) => {
    if (!terminalId) {
      return
    }

    const terminal = props.terminals.find((candidate) => candidate.id === terminalId)
    if (terminal?.todoId) {
      expandTodo(terminal.todoId)
      return
    }

    const todoProject = props.todoProjects.find((candidate) => candidate.id === terminal?.todoProjectId)
    if (todoProject) {
      expandTodo(todoProject.todoId)
    }
  },
  { immediate: true }
)

watch(
  () => props.projects,
  (projects) => {
    const projectIds = new Set(projects.map((project) => project.id))
    const nextSelectedProjectIds = new Set(
      [...selectedProjectIds.value].filter((projectId) => projectIds.has(projectId))
    )
    if (nextSelectedProjectIds.size !== selectedProjectIds.value.size) {
      selectedProjectIds.value = nextSelectedProjectIds
    }
    if (nextSelectedProjectIds.size === 0) {
      closeBulkProjectDeletePopover()
    }
  }
)

watch(
  completedTodos,
  (todos) => {
    const todoIds = new Set(todos.map((todo) => todo.id))
    const nextSelectedTodoIds = new Set(
      [...selectedCompletedTodoIds.value].filter((todoId) => todoIds.has(todoId))
    )
    if (nextSelectedTodoIds.size !== selectedCompletedTodoIds.value.size) {
      selectedCompletedTodoIds.value = nextSelectedTodoIds
    }
    if (nextSelectedTodoIds.size === 0) {
      closeBulkCompletedTodoDeletePopover()
    }
  },
  { deep: true }
)
</script>

<template>
  <aside class="project-sidebar">
    <div class="sidebar-header">
      <div class="sidebar-title">Workspace</div>
      <div class="sidebar-actions">
        <button
          type="button"
          class="icon-button"
          data-testid="new-todo"
          title="New TODO"
          :disabled="!hasWorkspace"
          @click="emit('create-todo')"
        >
          <Plus :size="18" />
        </button>
      </div>
    </div>

    <div class="project-list" data-testid="todo-workspace">
      <div v-if="!hasWorkspace" class="sidebar-empty workspace-empty" data-testid="todo-workspace-empty">
        Open a project
      </div>
      <template v-else>
        <div class="todo-view-tabs" data-testid="todo-workflow-tabs" role="tablist" aria-label="TODO views">
          <button
            type="button"
            class="todo-view-tab"
            :class="{ active: todoView === 'not-started' }"
            data-testid="todo-view-not-started"
            @click="setTodoView('not-started')"
          >
            未执行
          </button>
          <button
            type="button"
            class="todo-view-tab"
            :class="{ active: todoView === 'in-progress' }"
            data-testid="todo-view-in-progress"
            @click="setTodoView('in-progress')"
          >
            执行中
          </button>
          <button
            type="button"
            class="todo-view-tab"
            :class="{ active: todoView === 'completed' }"
            data-testid="todo-view-completed"
            @click="setTodoView('completed')"
          >
            已完成
          </button>
        </div>

        <div ref="todoWorkspaceScrollElement" class="todo-workspace-scroll" data-testid="todo-workspace-scroll">
          <div
            class="todo-tree-toolbar"
            data-testid="todo-tree-toolbar"
            role="toolbar"
            :aria-label="isOpenTodoView ? 'TODO tree controls' : 'Completed TODO controls'"
          >
        <template v-if="isOpenTodoView">
          <div class="todo-sort-toggle" role="group" aria-label="TODO sort">
            <button
              type="button"
              class="todo-sort-option"
              :class="{ active: activeTodoSortMode === 'priority' }"
              data-testid="sort-active-todos-priority"
              :aria-pressed="activeTodoSortMode === 'priority'"
              :disabled="todoInteractionsLocked"
              @click="setActiveTodoSortMode('priority')"
            >
              Priority
            </button>
            <button
              type="button"
              class="todo-sort-option"
              :class="{ active: activeTodoSortMode === 'time' }"
              data-testid="sort-active-todos-time"
              :aria-pressed="activeTodoSortMode === 'time'"
              :disabled="todoInteractionsLocked"
              @click="setActiveTodoSortMode('time')"
            >
              Time
            </button>
            <button
              type="button"
              class="todo-sort-option"
              :class="{ active: activeTodoSortMode === 'manual' }"
              data-testid="sort-active-todos-manual"
              :aria-pressed="activeTodoSortMode === 'manual'"
              :disabled="todoInteractionsLocked"
              @click="setActiveTodoSortMode('manual')"
            >
              Manual
            </button>
          </div>
          <button
            type="button"
            class="todo-tree-action"
            data-testid="collapse-all-todos"
            :disabled="!hasActiveTodos || todoInteractionsLocked"
            aria-label="Collapse all TODOs"
            title="Collapse all TODOs"
            @click="collapseAllTodos"
          >
            <ListChevronsDownUp :size="15" />
          </button>
          <button
            type="button"
            class="todo-tree-action"
            data-testid="expand-all-todos"
            :disabled="!hasActiveTodos || todoInteractionsLocked"
            aria-label="Expand all TODOs"
            title="Expand all TODOs"
            @click="expandAllTodos"
          >
            <ListChevronsUpDown :size="15" />
          </button>
        </template>
        <template v-else>
          <div class="completed-todo-selection-copy">{{ selectedCompletedTodoCount }} selected</div>
          <div class="bulk-completed-todo-delete-control">
            <button
              type="button"
              class="todo-tree-action todo-tree-action-danger"
              data-testid="bulk-delete-completed-todos"
              :disabled="selectedCompletedTodoCount === 0"
              :aria-expanded="confirmBulkDeleteCompletedTodos"
              aria-controls="bulk-delete-completed-todos-popover"
              title="Delete selected completed TODOs"
              @click.stop="openBulkCompletedTodoDeletePopover"
            >
              <Trash2 :size="15" />
              <span>Delete selected ({{ selectedCompletedTodoCount }})</span>
            </button>
            <div
              v-if="confirmBulkDeleteCompletedTodos"
              id="bulk-delete-completed-todos-popover"
              class="todo-action-popover bulk-completed-todo-delete-popover"
              data-testid="bulk-delete-completed-todos-popover"
              @click.stop
            >
              <span class="todo-action-confirm-copy">Delete {{ selectedCompletedTodoCount }} completed TODOs?</span>
              <div class="todo-action-confirm-actions">
                <button
                  type="button"
                  class="todo-action-confirm-cancel"
                  data-testid="cancel-bulk-delete-completed-todos"
                  @click="closeBulkCompletedTodoDeletePopover"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  class="todo-action-confirm-button todo-action-confirm-button-delete"
                  data-testid="confirm-bulk-delete-completed-todos"
                  @click="confirmBulkCompletedTodoDeletion"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        </template>
          </div>

          <div
            v-if="isOpenTodoView"
            ref="todoListElement"
            class="todo-list"
            :class="{
              'is-reordering': isTodoReordering,
              'is-interaction-locked': todoInteractionsLocked
            }"
            :data-testid="currentOpenTodoListTestId"
          >
        <div v-if="currentOpenTodos.length === 0" class="sidebar-empty">
          {{ todoView === 'in-progress' ? 'No in-progress TODOs' : 'No not-started TODOs' }}
        </div>

        <div
          v-for="todo in currentOpenTodos"
          :key="todo.id"
          :data-id="todo.id"
          class="todo-node"
          :class="{
            active: isTodoActive(todo),
            'has-active-terminal': todoHasActiveTerminal(todo),
            'is-collapsed': isTodoCollapsed(todo.id),
            'is-expanded': isTodoExpanded(todo.id)
          }"
        >
          <div
            class="todo-header-row"
            :class="[
              {
                active: isTodoActive(todo),
                'has-drag-handle': activeTodoSortMode === 'manual'
              },
              todoPriorityClass(todo),
              collapsedTodoActivityClass(todo)
            ]"
            :data-activity-state="collapsedTodoFeedbackState(todo) || null"
            @dblclick="toggleTodoBranch(todo.id)"
          >
            <button
              type="button"
              class="branch-toggle"
              :aria-controls="todoProjectListId(todo.id)"
              :aria-expanded="!isTodoCollapsed(todo.id)"
              :aria-label="`${isTodoCollapsed(todo.id) ? 'Expand' : 'Collapse'} ${todo.title}`"
              :data-testid="`toggle-todo-${todo.id}`"
              :title="isTodoCollapsed(todo.id) ? 'Expand TODO' : 'Collapse TODO'"
              :disabled="todoInteractionsLocked"
              @click.stop="toggleTodoBranch(todo.id)"
              @dblclick.stop
            >
              <ChevronRight v-if="isTodoCollapsed(todo.id)" :size="16" />
              <ChevronDown v-else :size="16" />
            </button>

            <button
              v-if="activeTodoSortMode === 'manual'"
              type="button"
              class="todo-drag-handle"
              :data-testid="`drag-todo-${todo.id}`"
              :aria-label="`Drag ${todo.title} to reorder`"
              aria-keyshortcuts="ArrowUp ArrowDown"
              :aria-disabled="todoInteractionsLocked"
              title="Drag to reorder"
              @click.stop
              @dblclick.stop
              @keydown.up.prevent.stop="moveTodoByKeyboard(todo.id, -1)"
              @keydown.down.prevent.stop="moveTodoByKeyboard(todo.id, 1)"
            >
              <GripVertical :size="15" />
            </button>

            <div
              class="todo-row"
              :class="{ active: isTodoActive(todo) }"
              :data-activity-state="collapsedTodoActivityState(todo) || null"
              :data-testid="`todo-${todo.id}`"
              :title="collapsedTodoActivityState(todo) ? collapsedTodoActivityLabel(todo) : null"
              @contextmenu.prevent.stop="openTodoContextMenu(todo.id, $event)"
              @mouseenter="startTodoDescriptionTooltip(todo, $event)"
              @mouseleave="hideTodoDescriptionTooltip"
            >
              <ListTodo class="project-icon" :size="17" />
              <span class="project-copy">
                <span class="todo-title-line">
                  <span class="project-name">{{ todo.title }}</span>
                </span>
                <span
                  v-if="todo.description"
                  class="todo-description"
                  :data-testid="`todo-description-${todo.id}`"
                >
                  {{ todo.description }}
                </span>
                <span class="project-path">{{ todoProjectsForTodo(todo.id).length }} projects</span>
              </span>
            </div>

            <Teleport :to="descriptionTooltipLayer">
              <span
                v-if="isTodoDescriptionTooltipVisible(todo)"
                class="todo-description-tooltip"
                :style="descriptionTooltipStyle()"
                :data-testid="`todo-description-tooltip-${todo.id}`"
                role="tooltip"
              >
                {{ todoDescription(todo) }}
              </span>
            </Teleport>

            <div
              class="todo-actions"
              :data-testid="`todo-actions-${todo.id}`"
              role="group"
              :aria-label="`${todo.title} actions`"
              @dblclick.stop
            >
              <div class="todo-action-confirm-control">
                <button
                  type="button"
                  class="todo-action-button"
                  :data-testid="`todo-menu-button-${todo.id}`"
                  :title="`${todo.title} menu`"
                  aria-label="Open TODO menu"
                  :disabled="todoInteractionsLocked"
                  @click.stop="openTodoContextMenuFromButton(todo.id, $event)"
                >
                  <EllipsisVertical :size="14" />
                </button>
                <div
                  v-if="isTodoActionPopoverOpen(todo.id, 'delete')"
                  :id="`delete-todo-popover-${todo.id}`"
                  class="todo-action-popover"
                  :data-testid="`delete-todo-popover-${todo.id}`"
                  @click.stop
                >
                  <span class="todo-action-confirm-copy">Delete TODO?</span>
                  <div class="todo-action-confirm-actions">
                    <button
                      type="button"
                      class="todo-action-confirm-cancel"
                      :data-testid="`cancel-delete-todo-${todo.id}`"
                      aria-label="Cancel deleting TODO"
                      @click="closeTodoActionPopover"
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      class="todo-action-confirm-button todo-action-confirm-button-delete"
                      :data-testid="`confirm-delete-todo-${todo.id}`"
                      aria-label="Confirm deleting TODO"
                      @click="confirmTodoAction(todo.id, 'delete')"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
              <button
                v-if="todoWorkflowStatus(todo) === 'not-started'"
                type="button"
                class="todo-action-button"
                :data-testid="`mark-todo-in-progress-${todo.id}`"
                title="Mark in progress"
                aria-label="Mark TODO in progress"
                :disabled="todoInteractionsLocked"
                @click.stop="changeTodoStatus(todo.id, 'in-progress')"
              >
                <Play :size="14" />
              </button>
              <button
                v-if="todoWorkflowStatus(todo) === 'in-progress'"
                type="button"
                class="todo-action-button"
                :data-testid="`add-task-terminal-${todo.id}`"
                title="New task terminal"
                aria-label="New task terminal"
                :disabled="todoInteractionsLocked"
                :aria-expanded="isTerminalLaunchMenuOpen('task', todo.id)"
                :aria-controls="`terminal-launch-menu-task-${todo.id}`"
                @click.stop="toggleTaskTerminalLaunchMenu(todo.id, $event)"
              >
                <TerminalSquare :size="14" />
              </button>
              <div
                v-if="isTerminalLaunchMenuOpen('task', todo.id)"
                :id="`terminal-launch-menu-task-${todo.id}`"
                class="terminal-launch-menu"
                :class="terminalLaunchMenuClass()"
                :style="terminalLaunchMenuStyle()"
                :data-testid="`terminal-launch-menu-task-${todo.id}`"
                @click.stop
              >
                <button
                  v-for="(option, index) in terminalLaunchOptions"
                  :key="`${option.name}-${index}`"
                  type="button"
                  class="terminal-launch-option"
                  :data-testid="`terminal-launch-option-task-${todo.id}-${index}`"
                  @click="selectTaskTerminalLaunchOption(todo.id, option)"
                >
                  {{ option.name }}
                </button>
              </div>
              <div v-if="todoWorkflowStatus(todo) === 'in-progress'" class="todo-action-confirm-control">
                <button
                  type="button"
                  class="todo-action-button"
                  :data-testid="`complete-todo-${todo.id}`"
                  title="Complete TODO"
                  :disabled="todoInteractionsLocked"
                  :aria-expanded="isTodoActionPopoverOpen(todo.id, 'complete')"
                  :aria-controls="`complete-todo-popover-${todo.id}`"
                  @click.stop="openTodoActionPopover(todo.id, 'complete')"
                >
                  <Check :size="14" />
                </button>
                <div
                  v-if="isTodoActionPopoverOpen(todo.id, 'complete')"
                  :id="`complete-todo-popover-${todo.id}`"
                  class="todo-action-popover"
                  :data-testid="`complete-todo-popover-${todo.id}`"
                  @click.stop
                >
                  <span class="todo-action-confirm-copy">Complete TODO?</span>
                  <div class="todo-action-confirm-actions">
                    <button
                      type="button"
                      class="todo-action-confirm-cancel"
                      :data-testid="`cancel-complete-todo-${todo.id}`"
                      aria-label="Cancel completing TODO"
                      @click="closeTodoActionPopover"
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      class="todo-action-confirm-button todo-action-confirm-button-complete"
                      :data-testid="`confirm-complete-todo-${todo.id}`"
                      aria-label="Confirm completing TODO"
                      @click="confirmTodoAction(todo.id, 'complete')"
                    >
                      Complete
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="lifecycleStatusesForTodo(todo).length"
            class="todo-lifecycle-statuses"
            :data-testid="`todo-lifecycle-script-statuses-${todo.id}`"
          >
            <div
              v-for="status in lifecycleStatusesForTodo(todo)"
              :key="lifecycleScriptErrorKey(status)"
              class="todo-lifecycle-status"
              :class="`todo-lifecycle-status-${status.status}`"
              :data-testid="`todo-lifecycle-script-status-${todo.id}-${status.phase}`"
              @mouseenter="startLifecycleErrorTooltip(status, $event)"
              @mouseleave="scheduleHideLifecycleErrorTooltip"
            >
              <span class="terminal-activity" :class="status.status === 'failed' ? 'needs-ack' : 'busy'">
                <TriangleAlert v-if="status.status === 'failed'" :size="13" aria-hidden="true" />
                <LoaderCircle v-else :size="13" aria-hidden="true" />
              </span>
              <span class="todo-lifecycle-status-copy">
                <strong>{{ lifecycleScriptStatusLabel(status) }}</strong>
                <small v-if="lifecycleScriptStatusMessage(status)">{{ lifecycleScriptStatusMessage(status) }}</small>
              </span>
              <button
                v-if="status.status === 'failed'"
                type="button"
                class="todo-action-confirm-button"
                :data-testid="`retry-todo-lifecycle-script-${todo.id}-${status.phase}`"
                @click.stop="retryLifecycleScript(todo.id, status.phase)"
              >
                Retry
              </button>
              <button
                v-if="status.status === 'failed'"
                type="button"
                class="todo-lifecycle-copy-button"
                :data-testid="`copy-todo-lifecycle-script-error-${todo.id}-${status.phase}`"
                aria-label="Copy lifecycle script error"
                title="Copy lifecycle script error"
                @click.stop="emit('copy-todo-lifecycle-script-error', todo.id, status.phase)"
              >
                <Copy :size="14" aria-hidden="true" />
              </button>
              <Teleport :to="descriptionTooltipLayer">
                <pre
                  v-if="isLifecycleErrorTooltipVisible(status)"
                  class="todo-lifecycle-error-tooltip"
                  :style="lifecycleErrorTooltipStyle()"
                  :data-testid="`todo-lifecycle-script-error-tooltip-${todo.id}-${status.phase}`"
                  role="tooltip"
                  @mouseenter="keepLifecycleErrorTooltipVisible"
                  @mouseleave="scheduleHideLifecycleErrorTooltip"
                >{{ lifecycleScriptErrorOutput(status) }}</pre>
              </Teleport>
            </div>
          </div>

          <div
            v-if="isTodoContextMenuOpen(todo.id)"
            class="todo-context-menu"
            :style="todoContextMenuStyle()"
            :data-testid="`todo-context-menu-${todo.id}`"
            @click.stop
          >
            <button
              type="button"
              class="todo-context-menu-item"
              :data-testid="`todo-menu-edit-${todo.id}`"
              @click="emit('edit-todo', todo.id); closeTodoContextMenu()"
            >
              <Eye :size="14" />
              <span>View details</span>
            </button>
            <button
              type="button"
              class="todo-context-menu-item"
              :data-testid="`todo-menu-add-project-${todo.id}`"
              @click="emit('add-project-to-todo', todo.id); closeTodoContextMenu()"
            >
              <FolderPlus :size="14" />
              <span>Add project</span>
            </button>
            <button
              type="button"
              class="todo-context-menu-item"
              :data-testid="`todo-menu-open-folder-${todo.id}`"
              @click="emit('open-todo-folder', todo.id); closeTodoContextMenu()"
            >
              <FolderPlus :size="14" />
              <span>打开任务文件夹</span>
            </button>
            <button
              type="button"
              class="todo-context-menu-item"
              :data-testid="`todo-menu-copy-title-description-${todo.id}`"
              @click="emit('copy-todo-description', todo.id); closeTodoContextMenu()"
            >
              <Copy :size="14" />
              <span>Copy title and description</span>
            </button>
            <div class="todo-context-menu-separator"></div>
            <button
              type="button"
              class="todo-context-menu-item todo-context-menu-item-danger"
              :data-testid="`todo-menu-delete-${todo.id}`"
              @click="openTodoActionPopover(todo.id, 'delete'); closeTodoContextMenu()"
            >
              <Trash2 :size="14" />
              <span>Delete TODO</span>
            </button>
          </div>

          <div
            v-if="!isTodoCollapsed(todo.id)"
            :id="todoProjectListId(todo.id)"
            class="todo-project-list"
            :data-testid="`todo-project-list-${todo.id}`"
            role="group"
            :aria-label="`Projects for ${todo.title}`"
          >
            <div v-if="todoProjectsForTodo(todo.id).length === 0 && !hasTaskTerminals(todo.id)" class="sidebar-empty nested">
              No projects linked
            </div>

            <div
              v-if="hasTaskTerminals(todo.id)"
              class="task-terminal-group"
              :data-testid="`task-terminal-list-${todo.id}`"
            >
              <div
                v-for="terminal in taskTerminalsForTodo(todo.id)"
                :key="terminal.id"
                class="terminal-entry task-terminal-entry"
              >
                <button
                  type="button"
                  class="terminal-row task-terminal-row"
                  :class="{
                    active: terminal.id === activeTerminalId,
                    exited: terminal.state === 'exited',
                    'activity-busy': terminalActivityState(terminal) === 'busy',
                    'activity-needs-input': terminalActivityState(terminal) === 'needs-input',
                    'activity-needs-ack': terminalActivityState(terminal) === 'needs-ack'
                  }"
                  :aria-label="terminalRowLabel(terminal)"
                  :title="terminalRowLabel(terminal)"
                  :data-activity-state="terminalActivityState(terminal)"
                  :data-testid="`task-terminal-${terminal.id}`"
                  @click="emit('select-terminal', terminal.id)"
                >
                  <span
                    class="terminal-activity"
                    :class="terminalActivityState(terminal)"
                    :data-testid="`task-terminal-activity-${terminal.id}`"
                    :aria-label="terminalActivityLabel(terminal)"
                    role="img"
                  >
                    <LoaderCircle v-if="terminalActivityState(terminal) === 'busy'" :size="13" aria-hidden="true" />
                    <CircleAlert v-else-if="terminalActivityState(terminal) === 'needs-input'" :size="13" aria-hidden="true" />
                    <TriangleAlert v-else-if="terminalActivityState(terminal) === 'needs-ack'" :size="13" aria-hidden="true" />
                  </span>
                  <TerminalSquare class="terminal-icon" :size="15" />
                  <span class="terminal-name">{{ terminalDisplayName(terminal) }}</span>
                </button>
                <button
                  type="button"
                  class="delete-terminal-button"
                  :data-testid="`delete-terminal-${terminal.id}`"
                  title="Delete terminal"
                  @click.stop="emit('delete-terminal', terminal.id)"
                >
                  <Trash2 :size="13" />
                </button>
              </div>
            </div>

            <div
              v-for="todoProject in todoProjectsForTodo(todo.id)"
              :key="todoProject.id"
              class="project-node todo-project-node"
              :class="{
                'has-terminals': hasTodoProjectTerminals(todoProject.id),
                'has-active-terminal': todoProjectHasActiveTerminal(todoProject),
                'is-active-project': todoProject.id === activeTodoProjectId,
                'is-unavailable': !projectForTodoProject(todoProject)?.available
              }"
            >
              <div class="project-header-row todo-project-header-row">
                <button
                  type="button"
                  class="project-row"
                  :class="{
                    active: todoProject.id === activeTodoProjectId,
                    unavailable: !projectForTodoProject(todoProject)?.available
                  }"
                  :data-testid="`todo-project-${todoProject.id}`"
                  @click="selectTodoProject(todoProject)"
                >
                  <FolderGit2 class="project-icon" :size="17" />
                  <span class="project-copy">
                    <span
                      class="project-name"
                      :data-testid="`todo-project-name-${todoProject.id}`"
                    >
                      {{ todoProjectDisplayName(todoProject) }}
                    </span>
                    <span v-if="!projectForTodoProject(todoProject)?.available" class="project-status">Unavailable</span>
                    <span
                      v-if="todoProjectWorktreeFailed(todoProject)"
                      class="project-status project-status-error"
                      :data-testid="`todo-project-worktree-error-${todoProject.id}`"
                    >
                      {{ todoProject.worktreeError || 'Worktree preparation failed' }}
                    </span>
                  </span>
                </button>

                <div
                  v-if="todoWorkflowStatus(todo) === 'in-progress' && projectForTodoProject(todoProject)?.available"
                  class="terminal-launch-control"
                >
                  <button
                    type="button"
                    class="add-terminal-button"
                    :data-testid="`add-terminal-${todoProject.id}`"
                    title="New terminal"
                    :disabled="!todoProjectCanCreateTerminal(todoProject)"
                    :aria-expanded="isTerminalLaunchMenuOpen('project', todoProject.id)"
                    :aria-controls="`terminal-launch-menu-${todoProject.id}`"
                    @click.stop="toggleTerminalLaunchMenu(todoProject, $event)"
                  >
                    <TerminalSquare :size="14" />
                  </button>
                  <div
                    v-if="isTerminalLaunchMenuOpen('project', todoProject.id)"
                    :id="`terminal-launch-menu-${todoProject.id}`"
                    class="terminal-launch-menu"
                    :class="terminalLaunchMenuClass()"
                    :style="terminalLaunchMenuStyle()"
                    :data-testid="`terminal-launch-menu-${todoProject.id}`"
                    @click.stop
                  >
                    <button
                      v-for="(option, index) in terminalLaunchOptions"
                      :key="`${option.name}-${index}`"
                      type="button"
                      class="terminal-launch-option"
                      :data-testid="`terminal-launch-option-${todoProject.id}-${index}`"
                      @click="selectTerminalLaunchOption(todoProject, option)"
                    >
                      {{ option.name }}
                    </button>
                  </div>
                </div>
                <span v-else class="add-terminal-placeholder" aria-hidden="true"></span>

                <div class="todo-project-remove-control">
                  <button
                    type="button"
                    class="delete-project-button"
                    :data-testid="`remove-todo-project-${todoProject.id}`"
                    title="Remove project from TODO"
                    :aria-expanded="confirmRemoveTodoProjectId === todoProject.id"
                    :aria-controls="`remove-todo-project-popover-${todoProject.id}`"
                    @click.stop="openTodoProjectRemovePopover(todoProject.id)"
                  >
                    <Trash2 :size="14" />
                  </button>
                  <div
                    v-if="confirmRemoveTodoProjectId === todoProject.id"
                    :id="`remove-todo-project-popover-${todoProject.id}`"
                    class="todo-project-remove-popover"
                    :data-testid="`remove-todo-project-popover-${todoProject.id}`"
                    @click.stop
                  >
                    <span class="todo-project-remove-copy">Remove from TODO?</span>
                    <div class="todo-project-remove-actions">
                      <button
                        type="button"
                        class="todo-project-remove-cancel"
                        :data-testid="`cancel-remove-todo-project-${todoProject.id}`"
                        @click="closeTodoProjectRemovePopover"
                      >
                        Cancel
                      </button>
                      <button
                        type="button"
                        class="todo-project-remove-confirm"
                        :data-testid="`confirm-remove-todo-project-${todoProject.id}`"
                        @click="confirmTodoProjectRemoval(todoProject.id)"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div
                v-if="hasTodoProjectTerminals(todoProject.id)"
                :id="terminalListId(todoProject.id)"
                class="terminal-list"
                role="group"
                :aria-label="`Terminals for ${projectForTodoProject(todoProject)?.name || 'project'}`"
                :data-testid="`terminal-list-${todoProject.id}`"
              >
                <div
                  v-for="terminal in todoProjectTerminals(todoProject.id)"
                  :key="terminal.id"
                  class="terminal-entry"
                >
                  <button
                    type="button"
                    class="terminal-row"
                    :class="{
                      active: terminal.id === activeTerminalId,
                      exited: terminal.state === 'exited',
                      'activity-busy': terminalActivityState(terminal) === 'busy',
                      'activity-needs-input': terminalActivityState(terminal) === 'needs-input',
                      'activity-needs-ack': terminalActivityState(terminal) === 'needs-ack'
                    }"
                    :aria-label="terminalRowLabel(terminal)"
                    :title="terminalRowLabel(terminal)"
                    :data-activity-state="terminalActivityState(terminal)"
                    :data-testid="`terminal-${terminal.id}`"
                    @click="emit('select-terminal', terminal.id)"
                  >
                    <span
                      class="terminal-activity"
                      :class="terminalActivityState(terminal)"
                      :data-testid="`terminal-activity-${terminal.id}`"
                      :aria-label="terminalActivityLabel(terminal)"
                      role="img"
                    >
                      <LoaderCircle v-if="terminalActivityState(terminal) === 'busy'" :size="13" aria-hidden="true" />
                      <CircleAlert v-else-if="terminalActivityState(terminal) === 'needs-input'" :size="13" aria-hidden="true" />
                      <TriangleAlert v-else-if="terminalActivityState(terminal) === 'needs-ack'" :size="13" aria-hidden="true" />
                    </span>
                    <TerminalSquare class="terminal-icon" :size="15" />
                    <span class="terminal-name">{{ terminalDisplayName(terminal) }}</span>
                  </button>
                  <button
                    type="button"
                    class="delete-terminal-button"
                    :data-testid="`delete-terminal-${terminal.id}`"
                    title="Delete terminal"
                    @click.stop="emit('delete-terminal', terminal.id)"
                  >
                    <Trash2 :size="13" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
          </div>

          <div v-else class="archived-todos completed-todos" data-testid="completed-todos">
        <div v-if="completedTodos.length === 0" class="sidebar-empty">No completed TODOs</div>

        <div
          v-for="todo in completedTodos"
          :key="todo.id"
          class="archived-todo completed-todo"
          :data-testid="`completed-todo-${todo.id}`"
        >
          <div class="completed-todo-header">
            <label class="completed-todo-select">
              <input
                type="checkbox"
                :checked="isCompletedTodoSelected(todo.id)"
                :data-testid="`select-completed-todo-${todo.id}`"
                :aria-label="`Select completed TODO ${todo.title}`"
                @click.stop="toggleCompletedTodoSelection(todo.id)"
              />
            </label>
            <div class="archived-todo-title completed-todo-title">
              <Archive :size="15" />
              <span>{{ todo.title }}</span>
            </div>
            <div class="todo-action-confirm-control completed-todo-menu-control">
              <button
                type="button"
                class="todo-action-button"
                :data-testid="`completed-todo-menu-button-${todo.id}`"
                :title="`${todo.title} menu`"
                aria-label="Open completed TODO menu"
                @click.stop="openCompletedTodoMenu(todo.id, $event)"
              >
                <EllipsisVertical :size="14" />
              </button>
              <div
                v-if="isCompletedTodoMenuOpen(todo.id)"
                class="todo-context-menu completed-todo-menu"
                :data-testid="`completed-todo-menu-${todo.id}`"
                @click.stop
              >
                <button
                  type="button"
                  class="todo-context-menu-item"
                  :data-testid="`completed-todo-menu-edit-${todo.id}`"
                  @click="emit('edit-todo', todo.id); closeCompletedTodoMenu()"
                >
                  <Eye :size="14" />
                  <span>View details</span>
                </button>
                <div class="todo-context-menu-separator"></div>
                <button
                  type="button"
                  class="todo-context-menu-item todo-context-menu-item-danger"
                  :data-testid="`completed-todo-menu-delete-${todo.id}`"
                  @click="openTodoActionPopover(todo.id, 'delete')"
                >
                  <Trash2 :size="14" />
                  <span>Delete TODO</span>
                </button>
              </div>
              <div
                v-if="isTodoActionPopoverOpen(todo.id, 'delete')"
                :id="`delete-todo-popover-${todo.id}`"
                class="todo-action-popover"
                :data-testid="`delete-todo-popover-${todo.id}`"
                @click.stop
              >
                <span class="todo-action-confirm-copy">Delete TODO?</span>
                <div class="todo-action-confirm-actions">
                  <button
                    type="button"
                    class="todo-action-confirm-cancel"
                    :data-testid="`cancel-delete-todo-${todo.id}`"
                    aria-label="Cancel deleting TODO"
                    @click="closeTodoActionPopover"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    class="todo-action-confirm-button todo-action-confirm-button-delete"
                    :data-testid="`confirm-delete-todo-${todo.id}`"
                    aria-label="Confirm deleting TODO"
                    @click="confirmTodoAction(todo.id, 'delete')"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div class="archived-todo-meta">
            <span>completed</span>
            <span>{{ completedAtLabel(todo) }}</span>
            <span v-if="completedDurationLabel(todo)">{{ completedDurationLabel(todo) }}</span>
          </div>
          <div v-if="todo.projectSnapshots?.length" class="archived-projects">
            <div
              v-for="(snapshot, snapshotIndex) in todo.projectSnapshots"
              :key="completedSnapshotKey(todo, snapshot, snapshotIndex)"
              class="archived-project"
            >
              <span
                class="completed-project-merge-status"
                :class="completedMergeStatus(todo, snapshot, snapshotIndex)"
                :data-testid="`completed-project-merge-status-${completedSnapshotTestId(todo, snapshot, snapshotIndex)}`"
                :title="completedMergeStatusTitle(completedMergeStatus(todo, snapshot, snapshotIndex))"
                aria-hidden="true"
              >
                <Check v-if="completedMergeStatus(todo, snapshot, snapshotIndex) === 'merged'" :size="14" />
                <LoaderCircle v-else-if="completedMergeStatus(todo, snapshot, snapshotIndex) === 'checking'" :size="14" />
                <TriangleAlert v-else :size="14" />
              </span>
              <span class="project-name">{{ snapshot.name }}</span>
              <span class="project-path">{{ completedSnapshotBranchLabel(snapshot) }}</span>
            </div>
          </div>
        </div>
          </div>
        </div>
      </template>
    </div>
  </aside>
</template>
