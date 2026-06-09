<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RotateCcw } from '@lucide/vue'
import ProjectSidebar from './components/ProjectSidebar.vue'
import { TerminalSessionManager } from './terminalManager'
import { createXtermSession } from './xtermFactory'
import {
  CreateProjectFromDialog,
  ListProjects,
  ResizeTerminal,
  SelectProject,
  SendTerminalInput,
  StartShell
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText, EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

const projects = ref([])
const activeProjectId = ref('')
const shellStatuses = reactive({})
const terminalContainers = new Map()
const errorMessage = ref('')
const terminalMenu = reactive({
  visible: false,
  projectId: '',
  x: 0,
  y: 0
})
const terminalManager = new TerminalSessionManager({
  createSession: createXtermSession,
  sendInput: (projectId, data) => SendTerminalInput(projectId, data),
  resizeTerminal: (projectId, cols, rows) => {
    if (shellStatuses[projectId] === 'running') {
      ResizeTerminal(projectId, cols, rows)
    }
  },
  clipboard: {
    readText: ClipboardGetText,
    writeText: ClipboardSetText
  },
  onError: showError
})

const activeProject = computed(() => {
  return projects.value.find((project) => project.id === activeProjectId.value) || null
})

onMounted(async () => {
  EventsOn('terminal-output', (event) => {
    terminalManager.write(event.projectId, event.data)
  })
  EventsOn('terminal-status', (status) => {
    shellStatuses[status.projectId] = status.state
  })
  window.addEventListener('resize', fitActiveTerminal)
  window.addEventListener('click', closeTerminalMenu)

  try {
    applyState(await ListProjects())
    if (activeProject.value?.available) {
      await activateProject(activeProject.value.id)
    }
  } catch (error) {
    showError(error)
  }
})

onBeforeUnmount(() => {
  EventsOff('terminal-output')
  EventsOff('terminal-status')
  window.removeEventListener('resize', fitActiveTerminal)
  window.removeEventListener('click', closeTerminalMenu)
})

function applyState(state) {
  projects.value = state?.projects || []
  activeProjectId.value = state?.activeProjectId || ''
  closeTerminalMenu()
}

async function createProject() {
  try {
    applyState(await CreateProjectFromDialog())
    if (activeProject.value?.available) {
      await activateProject(activeProject.value.id)
    }
  } catch (error) {
    showError(error)
  }
}

async function selectProject(projectId) {
  try {
    applyState(await SelectProject(projectId))
    if (activeProject.value?.available) {
      await activateProject(projectId)
    }
  } catch (error) {
    showError(error)
  }
}

async function activateProject(projectId) {
  await nextTick()
  const container = terminalContainers.get(projectId)
  if (!container) {
    return
  }

  const session = terminalManager.activate(projectId, container)
  const status = await StartShell(projectId, session.terminal.cols || 80, session.terminal.rows || 24)
  shellStatuses[projectId] = status.state
  terminalManager.fitActive()
}

async function restartActiveShell() {
  if (activeProject.value?.available) {
    await activateProject(activeProject.value.id)
  }
}

function setTerminalContainer(projectId, element) {
  if (element) {
    terminalContainers.set(projectId, element)
    if (projectId === activeProjectId.value) {
      nextTick(() => activateProject(projectId))
    }
  } else {
    terminalContainers.delete(projectId)
  }
}

function openTerminalMenu(projectId, event) {
  if (projectId !== activeProjectId.value) {
    return
  }
  terminalMenu.projectId = projectId
  terminalMenu.x = event.clientX
  terminalMenu.y = event.clientY
  terminalMenu.visible = true
}

function closeTerminalMenu() {
  terminalMenu.visible = false
  terminalMenu.projectId = ''
}

async function copyFromTerminalMenu() {
  const projectId = terminalMenu.projectId
  await terminalManager.copySelection(projectId)
  closeTerminalMenu()
}

async function pasteFromTerminalMenu() {
  const projectId = terminalMenu.projectId
  await terminalManager.paste(projectId)
  closeTerminalMenu()
}

function hasTerminalSelection(projectId) {
  return terminalManager.hasSelection(projectId)
}

function fitActiveTerminal() {
  terminalManager.fitActive()
}

function showError(error) {
  errorMessage.value = error?.message || String(error)
}
</script>

<template>
  <main class="app-shell">
    <ProjectSidebar
      :projects="projects"
      :active-project-id="activeProjectId"
      @create-project="createProject"
      @select-project="selectProject"
    />

    <section class="workspace">
      <header class="workspace-header">
        <div v-if="activeProject" class="project-heading">
          <span class="heading-name">{{ activeProject.name }}</span>
          <span class="heading-path">{{ activeProject.path }}</span>
        </div>
        <div v-else class="project-heading muted">No project selected</div>
        <button
          v-if="activeProject && shellStatuses[activeProject.id] === 'exited'"
          type="button"
          class="toolbar-button"
          title="Restart shell"
          @click="restartActiveShell"
        >
          <RotateCcw :size="16" />
          <span>Restart</span>
        </button>
      </header>

      <div class="terminal-surface">
        <div
          v-for="project in projects"
          :key="project.id"
          class="terminal-pane"
          :class="{ active: project.id === activeProjectId }"
          :data-testid="`terminal-pane-${project.id}`"
          :ref="(element) => setTerminalContainer(project.id, element)"
          @contextmenu.prevent="openTerminalMenu(project.id, $event)"
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
            :disabled="!hasTerminalSelection(terminalMenu.projectId)"
            @click="copyFromTerminalMenu"
          >
            Copy
          </button>
          <button type="button" data-testid="terminal-menu-paste" @click="pasteFromTerminalMenu">
            Paste
          </button>
        </div>

        <div v-if="!activeProject" class="state-layer">Select a project</div>
        <div v-else-if="!activeProject.available" class="state-layer warning">Project path unavailable</div>
        <div v-else-if="shellStatuses[activeProject.id] === 'exited'" class="state-layer warning">
          Shell exited
        </div>
      </div>

      <footer v-if="errorMessage" class="error-bar">{{ errorMessage }}</footer>
    </section>
  </main>
</template>
