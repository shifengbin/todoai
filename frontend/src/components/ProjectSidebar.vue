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
  ListTodo,
  LoaderCircle,
  Plus,
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
  'create-todo',
  'add-project-to-todo',
  'select-todo-project',
  'complete-todo',
  'delete-todo',
  'create-terminal',
  'select-terminal',
  'delete-terminal'
])

const activeTab = ref('todos')
const todoView = ref('active')
const collapsedTodoIds = ref(new Set())
const openLaunchTodoProjectId = ref('')
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

const activeTodos = computed(() => props.todos.filter((todo) => todo.status === 'active'))
const archivedTodos = computed(() => props.todos.filter((todo) => todo.status !== 'active'))

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
  } else {
    nextCollapsedTodoIds.add(todoId)
  }
  collapsedTodoIds.value = nextCollapsedTodoIds
}

function expandTodo(todoId) {
  if (!todoId || !collapsedTodoIds.value.has(todoId)) {
    return
  }

  const nextCollapsedTodoIds = new Set(collapsedTodoIds.value)
  nextCollapsedTodoIds.delete(todoId)
  collapsedTodoIds.value = nextCollapsedTodoIds
}

function selectTodoProject(todoProject) {
  expandTodo(todoProject.todoId)
  emit('select-todo-project', todoProject.id)
}

function toggleTerminalLaunchMenu(todoProject, event) {
  expandTodo(todoProject.todoId)
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

function terminalActivityState(terminal) {
  return terminal.activityState || 'idle'
}

function terminalActivityLabel(terminal) {
  const state = terminalActivityState(terminal)
  if (state === 'busy') {
    return 'Running'
  }
  if (state === 'needs-input') {
    return 'Needs input'
  }
  return 'Idle'
}

function terminalRowLabel(terminal) {
  const activityLabel = terminalActivityLabel(terminal)
  const displayName = terminalDisplayName(terminal)
  return activityLabel === 'Idle' ? displayName : `${displayName} - ${activityLabel}`
}

function todoPriority(todo) {
  return ['high', 'medium', 'low'].includes(todo?.priority) ? todo.priority : 'medium'
}

function todoPriorityLabel(todo) {
  const priority = todoPriority(todo)
  if (priority === 'high') {
    return '高'
  }
  if (priority === 'low') {
    return '低'
  }
  return '中'
}

function todoPriorityClass(todo) {
  return `todo-row-priority-${todoPriority(todo)}`
}

function archivedAtLabel(todo) {
  return todo.archivedAt || 'No archive time'
}

onMounted(() => {
  window.addEventListener('click', closeTerminalLaunchMenu)
})

onBeforeUnmount(() => {
  window.removeEventListener('click', closeTerminalLaunchMenu)
})

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
      <div class="todo-view-tabs" role="tablist" aria-label="TODO views">
        <button
          type="button"
          class="todo-view-tab"
          :class="{ active: todoView === 'active' }"
          data-testid="todo-view-active"
          @click="todoView = 'active'"
        >
          Active
        </button>
        <button
          type="button"
          class="todo-view-tab"
          :class="{ active: todoView === 'archived' }"
          data-testid="todo-view-archived"
          @click="todoView = 'archived'"
        >
          Archived
        </button>
      </div>

      <div v-if="todoView === 'active'" class="todo-list" data-testid="active-todos">
        <div v-if="activeTodos.length === 0" class="sidebar-empty">No active TODOs</div>

        <div
          v-for="todo in activeTodos"
          :key="todo.id"
          class="todo-node"
          :class="{
            active: todo.id === activeTodoId,
            'has-active-terminal': todoHasActiveTerminal(todo),
            'is-collapsed': isTodoCollapsed(todo.id),
            'is-expanded': isTodoExpanded(todo.id)
          }"
        >
          <div class="todo-header-row">
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
              :class="[{ active: todo.id === activeTodoId }, todoPriorityClass(todo)]"
              :data-testid="`todo-${todo.id}`"
            >
              <ListTodo class="project-icon" :size="17" />
              <span class="project-copy">
                <span class="todo-title-line">
                  <span class="project-name">{{ todo.title }}</span>
                  <span
                    class="todo-priority-badge"
                    :class="`todo-priority-badge-${todoPriority(todo)}`"
                    :data-testid="`todo-priority-${todo.id}`"
                  >
                    {{ todoPriorityLabel(todo) }}
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

            <button
              type="button"
              class="todo-action-button"
              :data-testid="`add-project-to-todo-${todo.id}`"
              title="Add project"
              @click.stop="emit('add-project-to-todo', todo.id)"
            >
              <FolderPlus :size="14" />
            </button>
            <button
              type="button"
              class="todo-action-button"
              :data-testid="`complete-todo-${todo.id}`"
              title="Complete TODO"
              @click.stop="emit('complete-todo', todo.id)"
            >
              <Check :size="14" />
            </button>
            <button
              type="button"
              class="delete-project-button"
              :data-testid="`delete-todo-${todo.id}`"
              title="Delete TODO"
              @click.stop="emit('delete-todo', todo.id)"
            >
              <Trash2 :size="14" />
            </button>
          </div>

          <div
            v-if="!isTodoCollapsed(todo.id)"
            :id="todoProjectListId(todo.id)"
            class="todo-project-list"
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

      <div v-else class="archived-todos" data-testid="archived-todos">
        <div v-if="archivedTodos.length === 0" class="sidebar-empty">No archived TODOs</div>

        <div
          v-for="todo in archivedTodos"
          :key="todo.id"
          class="archived-todo"
          :data-testid="`archived-todo-${todo.id}`"
        >
          <div class="archived-todo-title">
            <Archive :size="15" />
            <span>{{ todo.title }}</span>
          </div>
          <div class="archived-todo-meta">
            <span>{{ todo.archivedReason || todo.status }}</span>
            <span>{{ archivedAtLabel(todo) }}</span>
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

          <button
            type="button"
            class="delete-project-button"
            :data-testid="`delete-project-${project.id}`"
            title="Delete project"
            @click.stop="emit('delete-project', project.id)"
          >
            <Trash2 :size="14" />
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>
