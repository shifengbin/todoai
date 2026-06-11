<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Archive,
  Check,
  ChevronDown,
  ChevronRight,
  CircleAlert,
  FolderInput,
  FolderPlus,
  Eye,
  ListChevronsDownUp,
  ListChevronsUpDown,
  ListTodo,
  LoaderCircle,
  Play,
  Plus,
  RotateCcw,
  TerminalSquare,
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
  'delete-todo',
  'create-terminal',
  'select-terminal',
  'delete-terminal',
  'todo-expanded'
])

const activeTab = ref('todos')
const todoView = ref('not-started')
const activeTodoSortMode = ref('priority')
const collapsedTodoIds = ref(new Set())
const knownTodoIds = ref(new Set(props.todos.map((todo) => todo.id)))
const openLaunchTodoProjectId = ref('')
const confirmRemoveTodoProjectId = ref('')
const todoActionConfirm = ref({ todoId: '', action: '' })
const confirmDeleteProjectId = ref('')
const selectedProjectIds = ref(new Set())
const confirmBulkDeleteProjects = ref(false)
const launchMenuPlacement = ref('down')
const launchMenuMaxHeight = ref('')

const launchMenuBorderHeight = 2
const launchMenuMinimumHeight = 32
const launchMenuOptionHeight = 32
const launchMenuViewportPadding = 8

const terminalLaunchOptions = computed(() => [
  { name: 'Terminal', command: '' },
  ...props.launchProfiles
])

const todoPriorityOrder = {
  high: 0,
  medium: 1,
  low: 2
}

const notStartedTodos = computed(() => sortedOpenTodos('not-started'))
const inProgressTodos = computed(() => sortedOpenTodos('in-progress'))
const completedTodos = computed(() => props.todos.filter((todo) => todoWorkflowStatus(todo) === 'completed'))
const currentOpenTodos = computed(() => (todoView.value === 'in-progress' ? inProgressTodos.value : notStartedTodos.value))
const currentOpenTodoListTestId = computed(() => `${todoView.value}-todos`)
const isOpenTodoView = computed(() => ['not-started', 'in-progress'].includes(todoView.value))
const activeTodos = computed(() => currentOpenTodos.value)
const activeTodoIds = computed(() => currentOpenTodos.value.map((todo) => todo.id))
const hasActiveTodos = computed(() => activeTodoIds.value.length > 0)
const selectedProjectCount = computed(() => selectedProjectIds.value.size)
const selectedProjectIdsList = computed(() => props.projects.filter((project) => selectedProjectIds.value.has(project.id)).map((project) => project.id))

function sortedOpenTodos(status) {
  return props.todos
    .map((todo, index) => ({ todo, index }))
    .filter(({ todo }) => todoWorkflowStatus(todo) === status)
    .sort(compareActiveTodoEntries)
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
  expandTodo(todoProject.todoId)
  emit('select-todo-project', todoProject.id)
}

function toggleTerminalLaunchMenu(todoProject, event) {
  expandTodo(todoProject.todoId)
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
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
  closeTerminalLaunchMenu()
  closeTodoActionPopover()
  closeProjectDeletePopover()
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
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeProjectDeletePopover()
  todoActionConfirm.value = { todoId, action }
}

function closeTodoActionPopover() {
  todoActionConfirm.value = { todoId: '', action: '' }
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
  todoView.value = status
  emit('change-todo-status', todoId, status)
}

function openProjectDeletePopover(projectId) {
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeBulkProjectDeletePopover()
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
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
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

function closeFloatingMenus() {
  closeTerminalLaunchMenu()
  closeTodoProjectRemovePopover()
  closeTodoActionPopover()
  closeProjectDeletePopover()
  closeBulkProjectDeletePopover()
}

function selectTerminalLaunchOption(todoProject, option) {
  expandTodo(todoProject.todoId)
  emit('create-terminal', todoProject.id, option.command ? option : null)
  closeTerminalLaunchMenu()
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
  return terminal.activityState || 'idle'
}

function activityStateLabel(state) {
  if (state === 'busy') {
    return 'Running'
  }
  if (state === 'needs-input') {
    return 'Needs input'
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
    if (state === 'busy') {
      hasBusyTerminal = true
    }
  }
  if (!hasTerminal) {
    return ''
  }
  return hasBusyTerminal ? 'busy' : 'idle'
}

function collapsedTodoActivityState(todo) {
  return isTodoCollapsed(todo.id) ? todoActivityState(todo) : ''
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
  [() => props.activeTerminalId, () => props.terminals],
  ([terminalId]) => {
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
</script>

<template>
  <aside class="project-sidebar">
    <div class="sidebar-header">
      <div class="sidebar-title">Workspace</div>
      <div class="sidebar-actions">
        <button
          v-if="activeTab === 'todos'"
          type="button"
          class="icon-button"
          data-testid="new-todo"
          title="New TODO"
          @click="emit('create-todo')"
        >
          <Plus :size="18" />
        </button>
        <button
          v-else
          type="button"
          class="icon-button"
          data-testid="new-project"
          title="New project"
          @click="emit('create-project')"
        >
          <FolderPlus :size="18" />
        </button>
      </div>
    </div>

    <div class="sidebar-tabs tab-strip" data-testid="workspace-tabs" role="tablist" aria-label="Workspace sections">
      <button
        type="button"
        class="sidebar-tab"
        :class="{ active: activeTab === 'todos' }"
        data-testid="sidebar-tab-todos"
        role="tab"
        :aria-selected="activeTab === 'todos'"
        @click="activeTab = 'todos'"
      >
        <ListTodo :size="15" />
        <span>TODO</span>
      </button>
      <button
        type="button"
        class="sidebar-tab"
        :class="{ active: activeTab === 'projects' }"
        data-testid="sidebar-tab-projects"
        role="tab"
        :aria-selected="activeTab === 'projects'"
        @click="activeTab = 'projects'"
      >
        <TerminalSquare :size="15" />
        <span>项目</span>
      </button>
    </div>

    <div v-if="activeTab === 'todos'" class="project-list" data-testid="todo-workspace">
      <div class="todo-view-tabs" data-testid="todo-workflow-tabs" role="tablist" aria-label="TODO views">
        <button
          type="button"
          class="todo-view-tab"
          :class="{ active: todoView === 'not-started' }"
          data-testid="todo-view-not-started"
          @click="todoView = 'not-started'"
        >
          未执行
        </button>
        <button
          type="button"
          class="todo-view-tab"
          :class="{ active: todoView === 'in-progress' }"
          data-testid="todo-view-in-progress"
          @click="todoView = 'in-progress'"
        >
          执行中
        </button>
        <button
          type="button"
          class="todo-view-tab"
          :class="{ active: todoView === 'completed' }"
          data-testid="todo-view-completed"
          @click="todoView = 'completed'"
        >
          已完成
        </button>
      </div>

      <div
        v-if="isOpenTodoView"
        class="todo-tree-toolbar"
        role="toolbar"
        aria-label="TODO tree controls"
      >
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
            :class="[{ active: todo.id === activeTodoId }, todoPriorityClass(todo)]"
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
            >
              <ListTodo class="project-icon" :size="17" />
              <span class="project-copy">
                <span class="todo-title-line">
                  <span class="project-name">{{ todo.title }}</span>
                  <span
                    v-if="collapsedTodoActivityState(todo) && collapsedTodoActivityState(todo) !== 'idle'"
                    class="terminal-activity todo-activity"
                    :class="collapsedTodoActivityState(todo)"
                    :data-testid="`todo-activity-${todo.id}`"
                    :aria-label="collapsedTodoActivityLabel(todo)"
                    role="img"
                  >
                    <LoaderCircle v-if="collapsedTodoActivityState(todo) === 'busy'" :size="13" aria-hidden="true" />
                    <CircleAlert v-else-if="collapsedTodoActivityState(todo) === 'needs-input'" :size="13" aria-hidden="true" />
                  </span>
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

            <div
              class="todo-actions"
              :data-testid="`todo-actions-${todo.id}`"
              role="group"
              :aria-label="`${todo.title} actions`"
            >
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
              <button
                v-else-if="todoWorkflowStatus(todo) === 'in-progress'"
                type="button"
                class="todo-action-button"
                :data-testid="`mark-todo-not-started-${todo.id}`"
                title="Mark not started"
                aria-label="Mark TODO not started"
                @click.stop="changeTodoStatus(todo.id, 'not-started')"
              >
                <RotateCcw :size="14" />
              </button>
              <button
                type="button"
                class="todo-action-button"
                :data-testid="`edit-todo-${todo.id}`"
                title="View and edit TODO"
                @click.stop="emit('edit-todo', todo.id)"
              >
                <Eye :size="14" />
              </button>
              <button
                type="button"
                class="todo-action-button"
                :data-testid="`add-project-to-todo-${todo.id}`"
                title="Add project"
                @click.stop="emit('add-project-to-todo', todo.id)"
              >
                <FolderPlus :size="14" />
              </button>
              <div class="todo-action-confirm-control">
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
              <div class="todo-action-confirm-control">
                <button
                  type="button"
                  class="delete-project-button todo-action-button"
                  :data-testid="`delete-todo-${todo.id}`"
                  title="Delete TODO"
                  :aria-expanded="isTodoActionPopoverOpen(todo.id, 'delete')"
                  :aria-controls="`delete-todo-popover-${todo.id}`"
                  @click.stop="openTodoActionPopover(todo.id, 'delete')"
                >
                  <Trash2 :size="14" />
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
            </div>
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

                <div v-if="projectForTodoProject(todoProject)?.available" class="terminal-launch-control">
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
                      'activity-busy': terminal.activityState === 'busy',
                      'activity-needs-input': terminal.activityState === 'needs-input'
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
          <div class="archived-todo-title completed-todo-title">
            <Archive :size="15" />
            <span>{{ todo.title }}</span>
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
    </div>

    <div v-else class="project-list" data-testid="project-library">
      <div class="project-library-actions">
        <button
          type="button"
          class="library-action-button"
          data-testid="import-parent-directory"
          @click="emit('import-projects')"
        >
          <FolderInput :size="15" />
          <span>Import parent</span>
        </button>
        <div class="bulk-project-delete-control">
          <button
            type="button"
            class="library-action-button library-action-button-delete"
            data-testid="bulk-delete-projects"
            :disabled="selectedProjectCount === 0"
            :aria-expanded="confirmBulkDeleteProjects"
            aria-controls="bulk-delete-projects-popover"
            @click.stop="openBulkProjectDeletePopover"
          >
            <Trash2 :size="15" />
            <span>Delete selected ({{ selectedProjectCount }})</span>
          </button>
          <div
            v-if="confirmBulkDeleteProjects"
            id="bulk-delete-projects-popover"
            class="bulk-project-delete-popover"
            data-testid="bulk-delete-projects-popover"
            @click.stop
          >
            <span class="project-delete-copy">Delete {{ selectedProjectCount }} projects?</span>
            <div class="project-delete-actions">
              <button
                type="button"
                class="project-delete-cancel"
                data-testid="cancel-bulk-delete-projects"
                @click="closeBulkProjectDeletePopover"
              >
                Cancel
              </button>
              <button
                type="button"
                class="project-delete-confirm"
                data-testid="confirm-bulk-delete-projects"
                @click="confirmBulkProjectDeletion"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="importSummary" class="import-summary" data-testid="import-summary">
        <span>{{ importSummary.addedCount || 0 }} imported</span>
        <span>{{ importSummary.skippedCount || 0 }} skipped</span>
      </div>

      <div v-if="projects.length === 0" class="sidebar-empty">No projects imported</div>

      <div
        v-for="project in projects"
        :key="project.id"
        class="project-node library-project-node"
        :class="{
          'is-active-project': project.id === activeProjectId,
          'is-unavailable': !project.available
        }"
      >
        <div class="project-header-row library-project-header-row">
          <input
            type="checkbox"
            class="project-select-checkbox"
            :checked="isProjectSelected(project.id)"
            :data-testid="`select-project-${project.id}`"
            :aria-label="`Select ${project.name}`"
            @click.stop="toggleProjectSelection(project.id)"
          />
          <button
            type="button"
            class="project-row"
            :class="{ active: project.id === activeProjectId, unavailable: !project.available }"
            :data-testid="`project-${project.id}`"
            @click="emit('select-project', project.id)"
          >
            <TerminalSquare class="project-icon" :size="17" />
            <span class="project-copy">
              <span class="project-name" :data-testid="`project-name-${project.id}`">{{ project.name }}</span>
              <span class="project-path">{{ project.path }}</span>
              <span v-if="!project.available" class="project-status">Unavailable</span>
            </span>
          </button>

          <div class="project-delete-control">
            <button
              type="button"
              class="delete-project-button"
              :data-testid="`delete-project-${project.id}`"
              title="Delete project"
              :aria-expanded="confirmDeleteProjectId === project.id"
              :aria-controls="`delete-project-popover-${project.id}`"
              @click.stop="openProjectDeletePopover(project.id)"
            >
              <Trash2 :size="14" />
            </button>
            <div
              v-if="confirmDeleteProjectId === project.id"
              :id="`delete-project-popover-${project.id}`"
              class="project-delete-popover"
              :data-testid="`delete-project-popover-${project.id}`"
              @click.stop
            >
              <span class="project-delete-copy">Delete project?</span>
              <div class="project-delete-actions">
                <button
                  type="button"
                  class="project-delete-cancel"
                  :data-testid="`cancel-delete-project-${project.id}`"
                  @click="closeProjectDeletePopover"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  class="project-delete-confirm"
                  :data-testid="`confirm-delete-project-${project.id}`"
                  @click="confirmProjectDeletion(project.id)"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>
