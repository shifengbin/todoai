<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RotateCcw } from '@lucide/vue'
import ProjectSidebar from './components/ProjectSidebar.vue'
import { TerminalSessionManager } from './terminalManager'
import { createXtermSession } from './xtermFactory'
import {
  CreateProjectFromDialog,
  CreateTerminal,
  ListProjects,
  ResizeTerminal,
  SelectProject,
  SelectTerminal,
  SendTerminalInput,
  StartShell
} from '../wailsjs/go/main/App'
import { ClipboardGetText, ClipboardSetText, EventsOff, EventsOn } from '../wailsjs/runtime/runtime'

const projects = ref([])
const terminals = ref([])
const activeProjectId = ref('')
const activeTerminalId = ref('')
const shellStatuses = reactive({})
const terminalContainers = new Map()
const errorMessage = ref('')
const terminalMenu = reactive({
  visible: false,
  terminalId: '',
  x: 0,
  y: 0
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
  onError: showError
})

const activeProject = computed(() => {
  return projects.value.find((project) => project.id === activeProjectId.value) || null
})

const activeTerminal = computed(() => {
  return terminals.value.find((terminal) => terminal.id === activeTerminalId.value) || null
})

const activeTerminalState = computed(() => {
  return activeTerminal.value ? terminalState(activeTerminal.value.id) : ''
})

onMounted(async () => {
  EventsOn('terminal-output', (event) => {
    terminalManager.write(event.terminalId, event.data)
  })
  EventsOn('terminal-status', (status) => {
    updateTerminalState(status.terminalId, status.state)
  })
  window.addEventListener('resize', fitActiveTerminal)
  window.addEventListener('click', closeTerminalMenu)

  try {
    applyState(await ListProjects())
    if (activeProject.value?.available) {
      await selectProject(activeProject.value.id)
    } else {
      await activateActiveTerminal()
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
  const previousTerminals = new Map(terminals.value.map((terminal) => [terminal.id, terminal]))
  projects.value = state?.projects || []
  terminals.value = (state?.terminals || []).map((terminal) => ({
    ...terminal,
    currentCommand:
      terminal.state === 'running'
        ? terminal.currentCommand || previousTerminals.get(terminal.id)?.currentCommand || ''
        : ''
  }))
  activeProjectId.value = state?.activeProjectId || ''
  activeTerminalId.value = state?.activeTerminalId || ''
  for (const terminal of terminals.value) {
    if (terminal.state) {
      shellStatuses[terminal.id] = terminal.state
    }
  }
  closeTerminalMenu()
}

async function createProject() {
  try {
    applyState(await CreateProjectFromDialog())
    if (activeProject.value?.available) {
      await selectProject(activeProject.value.id)
    }
  } catch (error) {
    showError(error)
  }
}

async function selectProject(projectId) {
  try {
    applyState(await SelectProject(projectId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function selectTerminal(terminalId) {
  try {
    applyState(await SelectTerminal(terminalId))
    await activateActiveTerminal()
  } catch (error) {
    showError(error)
  }
}

async function createTerminal(projectId) {
  try {
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    applyState(await CreateTerminal(projectId, size.cols || 80, size.rows || 24))
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
}

async function restartActiveShell() {
  const terminal = activeTerminal.value
  if (!terminal || !activeProject.value?.available) {
    return
  }
  const size = terminalManager.size(terminal.id) || { cols: 80, rows: 24 }
  try {
    const status = await StartShell(terminal.id, size.cols || 80, size.rows || 24)
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
}

function hasTerminalSelection(terminalId) {
  return terminalManager.hasSelection(terminalId)
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
    terminal.state = state
    if (state !== 'running') {
      terminal.currentCommand = ''
    }
  }
}

function handleTerminalCommandState(terminalId, event) {
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (!terminal) {
    return
  }
  if (event.type === 'command-start') {
    terminal.currentCommand = sanitizeCommandLabel(event.command)
  }
  if (event.type === 'command-end') {
    terminal.currentCommand = ''
  }
}

function sanitizeCommandLabel(command) {
  return (command || '').replace(/\s+/g, ' ').trim().slice(0, 120)
}

function showError(error) {
  errorMessage.value = error?.message || String(error)
}
</script>

<template>
  <main class="app-shell">
    <ProjectSidebar
      :projects="projects"
      :terminals="terminals"
      :active-project-id="activeProjectId"
      :active-terminal-id="activeTerminalId"
      @create-project="createProject"
      @select-project="selectProject"
      @create-terminal="createTerminal"
      @select-terminal="selectTerminal"
    />

    <section class="workspace">
      <header class="workspace-header">
        <div v-if="activeProject" class="project-heading">
          <span class="heading-name">{{ activeProject.name }}</span>
          <span class="heading-path">{{ activeProject.path }}</span>
        </div>
        <div v-else class="project-heading muted">No project selected</div>
        <button
          v-if="activeProject && activeTerminalState === 'exited'"
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

        <div v-if="!activeProject" class="state-layer">Select a project</div>
        <div v-else-if="!activeProject.available" class="state-layer warning">Project path unavailable</div>
        <div v-else-if="!activeTerminal" class="state-layer">Select a terminal</div>
        <div v-else-if="activeTerminalState === 'exited'" class="state-layer warning">Shell exited</div>
      </div>

      <footer v-if="errorMessage" class="error-bar">{{ errorMessage }}</footer>
    </section>
  </main>
</template>
