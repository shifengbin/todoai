<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Archive,
  Check,
  ChevronDown,
  ChevronRight,
  CircleAlert,
  Copy,
  EllipsisVertical,
  FolderPlus,
  Eye,
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
  'copy-todo-description',
  'delete-todo',
  'delete-completed-todos',
  'create-terminal',
  'select-terminal',
  'delete-terminal',
  'todo-expanded',
  'update:todo-view',
  'todo-view-change'
])

const internalTodoView = ref('not-started')
const activeTodoSortMode = ref('priority')
const collapsedTodoIds = ref(new Set())
const knownTodoIds = ref(new Set(props.todos.map((todo) => todo.id)))
const openLaunchTodoProjectId = ref('')
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
const launchMenuPlacement = ref('down')
const launchMenuMaxHeight = ref('')

const descriptionTooltipLayer = createTodoDescriptionTooltipLayer()
let descriptionTooltipTimer = null
const descriptionTooltipDelayMs = 600
const descriptionTooltipOffset = 12
const launchMenuBorderHeight = 2
const launchMenuMinimumHeight = 32
const launchMenuOptionHeight = 32
const launchMenuViewportPadding = 8
const todoContextMenuViewportPadding = 8
const todoContextMenuWidth = 180
const todoContextMenuHeight = 160

const terminalLaunchOptions = computed(() => [
  { name: 'Terminal', command: '' },
  ...props.launchProfiles.filter((profile) => profile?.enabled !== false)
])

const todoView = computed(() => normalizedTodoView(props.todoView || internalTodoView.value))

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
  return props.todos
    .map((todo, index) => ({ todo, index }))
    .filter(({ todo }) => todoWorkflowStatus(todo) === status)
    .sort(compareActiveTodoEntries)
    .map(({ todo }) => todo)
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

function todoProjectTerminals(todoProjectId) {
  return terminalsByTodoProject.value.get(todoProjectId) || []
}

function hasTodoProjectTerminals(todoProjectId) {
  return todoProjectTerminals(todoProjectId).length > 0
}

function todoHasActiveTerminal(todo) {
  return props.terminals.some((terminal) => terminal.todoId === todo.id && terminal.id === props.activeTerminalId)
}

function todoProjectHasActiveTerminal(todoProject) {
  return todoProjectTerminals(todoProject.id).some((terminal) => terminal.id === props.activeTerminalId)
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
  if (!hasActiveTodos.value) {
    return
  }

  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  for (const todoId of activeTodoIds.value) {
    nextCollapsedTodoIds.add(todoId)
  }
  collapsedTodoIds.value = nextCollapsedTodoIds
}

function expandAllTodos() {
  if (!hasActiveTodos.value) {
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
  hideTodoDescriptionTooltip()
  expandTodo(todoProject.todoId)
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeTodoContextMenu()
  if (openLaunchTodoProjectId.value === todoProject.id) {
    closeTerminalLaunchMenu()
    return
  }

  updateTerminalLaunchMenuPlacement(event?.currentTarget)
  openLaunchTodoProjectId.value = todoProject.id
}

function closeTerminalLaunchMenu() {
  openLaunchTodoProjectId.value = ''
  resetTerminalLaunchMenuPlacement()
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
  hideTodoDescriptionTooltip()
  setTodoView(status)
  emit('change-todo-status', todoId, status)
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
  hideTodoDescriptionTooltip()
  expandTodo(todoProject.todoId)
  emit('create-terminal', todoProject.id, option.command ? option : null)
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

  launchMenuPlacement.value = opensUp ? 'up' : 'down'
  launchMenuMaxHeight.value =
    availableHeight < desiredMenuHeight
      ? `${Math.max(availableHeight, launchMenuMinimumHeight)}px`
      : ''
}

function terminalLaunchMenuClass() {
  return {
    'terminal-launch-menu--up': launchMenuPlacement.value === 'up',
    'terminal-launch-menu--down': launchMenuPlacement.value !== 'up',
    'terminal-launch-menu--constrained': Boolean(launchMenuMaxHeight.value)
  }
}

function terminalLaunchMenuStyle() {
  return launchMenuMaxHeight.value ? { maxHeight: launchMenuMaxHeight.value } : {}
}

function terminalDisplayName(terminal) {
  return terminal.currentCommand || terminal.shellName || 'shell'
}

function compareActiveTodoEntries(left, right) {
  if (activeTodoSortMode.value === 'time') {
    return compareActiveTodosByTime(left, right)
  }
  return compareActiveTodosByPriority(left, right)
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
  if (!['priority', 'time'].includes(mode)) {
    return
  }
  activeTodoSortMode.value = mode
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

onMounted(() => {
  window.addEventListener('click', closeFloatingMenus)
})

onBeforeUnmount(() => {
  hideTodoDescriptionTooltip()
  descriptionTooltipLayer.remove()
  window.removeEventListener('click', closeFloatingMenus)
})

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
              @click="setActiveTodoSortMode('time')"
            >
              Time
            </button>
          </div>
          <button
            type="button"
            class="todo-tree-action"
            data-testid="collapse-all-todos"
            :disabled="!hasActiveTodos"
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
            :disabled="!hasActiveTodos"
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

      <div v-if="isOpenTodoView" class="todo-list" :data-testid="currentOpenTodoListTestId">
        <div v-if="currentOpenTodos.length === 0" class="sidebar-empty">
          {{ todoView === 'in-progress' ? 'No in-progress TODOs' : 'No not-started TODOs' }}
        </div>

        <div
          v-for="todo in currentOpenTodos"
          :key="todo.id"
          class="todo-node"
          :class="{
            active: todo.id === activeTodoId,
            'has-active-terminal': todoHasActiveTerminal(todo),
            'is-collapsed': isTodoCollapsed(todo.id),
            'is-expanded': isTodoExpanded(todo.id)
          }"
        >
          <div
            class="todo-header-row"
            :class="[{ active: todo.id === activeTodoId }, todoPriorityClass(todo), collapsedTodoActivityClass(todo)]"
            :data-activity-state="collapsedTodoFeedbackState(todo) || null"
          >
            <button
              type="button"
              class="branch-toggle"
              :aria-controls="todoProjectListId(todo.id)"
              :aria-expanded="!isTodoCollapsed(todo.id)"
              :aria-label="`${isTodoCollapsed(todo.id) ? 'Expand' : 'Collapse'} ${todo.title}`"
              :data-testid="`toggle-todo-${todo.id}`"
              :title="isTodoCollapsed(todo.id) ? 'Expand TODO' : 'Collapse TODO'"
              @click.stop="toggleTodoBranch(todo.id)"
            >
              <ChevronRight v-if="isTodoCollapsed(todo.id)" :size="16" />
              <ChevronDown v-else :size="16" />
            </button>

            <div
              class="todo-row"
              :class="{ active: todo.id === activeTodoId }"
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
            >
              <div class="todo-action-confirm-control">
                <button
                  type="button"
                  class="todo-action-button"
                  :data-testid="`todo-menu-button-${todo.id}`"
                  :title="`${todo.title} menu`"
                  aria-label="Open TODO menu"
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
                @click.stop="changeTodoStatus(todo.id, 'in-progress')"
              >
                <Play :size="14" />
              </button>
              <div v-if="todoWorkflowStatus(todo) === 'in-progress'" class="todo-action-confirm-control">
                <button
                  type="button"
                  class="todo-action-button"
                  :data-testid="`complete-todo-${todo.id}`"
                  title="Complete TODO"
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
            <div v-if="todoProjectsForTodo(todo.id).length === 0" class="sidebar-empty nested">
              No projects linked
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
                  <TerminalSquare class="project-icon" :size="17" />
                  <span class="project-copy">
                    <span
                      class="project-name"
                      :data-testid="`todo-project-name-${todoProject.id}`"
                    >
                      {{ projectForTodoProject(todoProject)?.name || 'Missing project' }}
                    </span>
                    <span class="project-path">{{ projectForTodoProject(todoProject)?.path || todoProject.projectId }}</span>
                    <span v-if="!projectForTodoProject(todoProject)?.available" class="project-status">Unavailable</span>
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
                    :aria-expanded="openLaunchTodoProjectId === todoProject.id"
                    :aria-controls="`terminal-launch-menu-${todoProject.id}`"
                    @click.stop="toggleTerminalLaunchMenu(todoProject, $event)"
                  >
                    <Plus :size="14" />
                  </button>
                  <div
                    v-if="openLaunchTodoProjectId === todoProject.id"
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
          </div>
          <div v-if="todo.projectSnapshots?.length" class="archived-projects">
            <div
              v-for="snapshot in todo.projectSnapshots"
              :key="`${todo.id}-${snapshot.projectId}`"
              class="archived-project"
            >
              <span class="project-name">{{ snapshot.name }}</span>
              <span class="project-path">{{ snapshot.path }}</span>
            </div>
          </div>
        </div>
      </div>
      </template>
    </div>
  </aside>
</template>
