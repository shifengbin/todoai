<script setup>
import { FolderPlus, TerminalSquare } from '@lucide/vue'

defineProps({
  projects: {
    type: Array,
    default: () => []
  },
  activeProjectId: {
    type: String,
    default: ''
  }
})

defineEmits(['create-project', 'select-project'])
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
      <button
        v-for="project in projects"
        :key="project.id"
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
    </div>
  </aside>
</template>
