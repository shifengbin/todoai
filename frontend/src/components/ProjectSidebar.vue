<script setup>
import { computed } from 'vue'
import { FolderPlus, Plus, TerminalSquare } from '@lucide/vue'

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

defineEmits(['create-project', 'select-project', 'create-terminal', 'select-terminal'])

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

function terminalDisplayName(terminal) {
  return terminal.currentCommand || terminal.shellName || 'shell'
}
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
        @click="$emit('create-project')"
      >
        <FolderPlus :size="18" />
      </button>
    </div>

    <div class="project-list">
      <div v-for="project in projects" :key="project.id" class="project-node">
        <div class="project-header-row">
          <button
            type="button"
            class="project-row"
            :class="{ active: project.id === activeProjectId, unavailable: !project.available }"
            :data-testid="`project-${project.id}`"
            @click="$emit('select-project', project.id)"
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
            @click.stop="$emit('create-terminal', project.id)"
          >
            <Plus :size="14" />
          </button>
        </div>

        <div v-if="projectTerminals(project.id).length" class="terminal-list">
          <button
            v-for="terminal in projectTerminals(project.id)"
            :key="terminal.id"
            type="button"
            class="terminal-row"
            :class="{ active: terminal.id === activeTerminalId, exited: terminal.state === 'exited' }"
            :data-testid="`terminal-${terminal.id}`"
            @click="$emit('select-terminal', terminal.id)"
          >
            <TerminalSquare class="terminal-icon" :size="15" />
            <span class="terminal-name">{{ terminalDisplayName(terminal) }}</span>
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>
