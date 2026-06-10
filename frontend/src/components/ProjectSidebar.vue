<script setup>
import { computed, ref, watch } from 'vue'
import { ChevronDown, ChevronRight, FolderPlus, Plus, TerminalSquare, Trash2 } from '@lucide/vue'

const props = defineProps({
  projects: {
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
  activeTerminalId: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'create-project',
  'select-project',
  'create-terminal',
  'select-terminal',
  'delete-project',
  'delete-terminal'
])

const collapsedProjectIds = ref(new Set())

const terminalsByProject = computed(() => {
  const groups = new Map()
  for (const terminal of props.terminals) {
    if (!groups.has(terminal.projectId)) {
      groups.set(terminal.projectId, [])
    }
    groups.get(terminal.projectId).push(terminal)
  }
  return groups
})

function projectTerminals(projectId) {
  return terminalsByProject.value.get(projectId) || []
}

function hasProjectTerminals(projectId) {
  return projectTerminals(projectId).length > 0
}

function projectHasActiveTerminal(projectId) {
  return projectTerminals(projectId).some((terminal) => terminal.id === props.activeTerminalId)
}

function isProjectCollapsed(projectId) {
  return collapsedProjectIds.value.has(projectId)
}

function isProjectExpanded(projectId) {
  return hasProjectTerminals(projectId) && !isProjectCollapsed(projectId)
}

function terminalListId(projectId) {
  return `terminal-list-${projectId}`
}

function toggleProjectBranch(projectId) {
  const nextCollapsedProjectIds = new Set(collapsedProjectIds.value)
  if (nextCollapsedProjectIds.has(projectId)) {
    nextCollapsedProjectIds.delete(projectId)
  } else {
    nextCollapsedProjectIds.add(projectId)
  }
  collapsedProjectIds.value = nextCollapsedProjectIds
}

function expandProject(projectId) {
  if (!projectId || !collapsedProjectIds.value.has(projectId)) {
    return
  }

  const nextCollapsedProjectIds = new Set(collapsedProjectIds.value)
  nextCollapsedProjectIds.delete(projectId)
  collapsedProjectIds.value = nextCollapsedProjectIds
}

function selectProject(projectId) {
  expandProject(projectId)
  emit('select-project', projectId)
}

function createTerminal(projectId) {
  expandProject(projectId)
  emit('create-terminal', projectId)
}

function terminalDisplayName(terminal) {
  return terminal.currentCommand || terminal.shellName || 'shell'
}

watch(
  () => props.activeProjectId,
  (projectId) => {
    expandProject(projectId)
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
    if (terminal) {
      expandProject(terminal.projectId)
    }
  },
  { immediate: true }
)
</script>

<template>
  <aside class="project-sidebar">
    <div class="sidebar-header">
      <div class="sidebar-title">Projects</div>
      <button
        type="button"
        class="icon-button"
        data-testid="new-project"
        title="New project"
        @click="emit('create-project')"
      >
        <FolderPlus :size="18" />
      </button>
    </div>

    <div class="project-list">
      <div
        v-for="project in projects"
        :key="project.id"
        class="project-node"
        :class="{
          'has-terminals': hasProjectTerminals(project.id),
          'has-active-terminal': projectHasActiveTerminal(project.id),
          'is-expanded': isProjectExpanded(project.id),
          'is-collapsed': hasProjectTerminals(project.id) && isProjectCollapsed(project.id),
          'is-active-project': project.id === activeProjectId,
          'is-unavailable': !project.available
        }"
      >
        <div class="project-header-row">
          <button
            v-if="hasProjectTerminals(project.id)"
            type="button"
            class="branch-toggle"
            :aria-controls="terminalListId(project.id)"
            :aria-expanded="!isProjectCollapsed(project.id)"
            :aria-label="`${isProjectCollapsed(project.id) ? 'Expand' : 'Collapse'} terminals for ${project.name}`"
            :data-testid="`toggle-project-${project.id}`"
            :title="isProjectCollapsed(project.id) ? 'Expand terminals' : 'Collapse terminals'"
            @click.stop="toggleProjectBranch(project.id)"
          >
            <ChevronRight v-if="isProjectCollapsed(project.id)" :size="16" />
            <ChevronDown v-else :size="16" />
          </button>
          <span v-else class="branch-toggle-placeholder" aria-hidden="true"></span>

          <button
            type="button"
            class="project-row"
            :class="{ active: project.id === activeProjectId, unavailable: !project.available }"
            :data-testid="`project-${project.id}`"
            @click="selectProject(project.id)"
          >
            <TerminalSquare class="project-icon" :size="17" />
            <span class="project-copy">
              <span class="project-name">{{ project.name }}</span>
              <span class="project-path">{{ project.path }}</span>
              <span v-if="!project.available" class="project-status">Unavailable</span>
            </span>
          </button>

          <button
            v-if="project.available"
            type="button"
            class="add-terminal-button"
            :data-testid="`add-terminal-${project.id}`"
            title="New terminal"
            @click.stop="createTerminal(project.id)"
          >
            <Plus :size="14" />
          </button>
          <span v-else class="add-terminal-placeholder" aria-hidden="true"></span>

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

        <div
          v-if="hasProjectTerminals(project.id) && !isProjectCollapsed(project.id)"
          :id="terminalListId(project.id)"
          class="terminal-list"
          role="group"
          :aria-label="`Terminals for ${project.name}`"
          :data-testid="`terminal-list-${project.id}`"
        >
          <div v-for="terminal in projectTerminals(project.id)" :key="terminal.id" class="terminal-entry">
            <button
              type="button"
              class="terminal-row"
              :class="{ active: terminal.id === activeTerminalId, exited: terminal.state === 'exited' }"
              :data-testid="`terminal-${terminal.id}`"
              @click="emit('select-terminal', terminal.id)"
            >
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
  </aside>
</template>
