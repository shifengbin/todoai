<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ChevronDown, ChevronUp, Plus, RotateCcw, Settings, Trash2, X } from '@lucide/vue'
import ProjectSidebar from './components/ProjectSidebar.vue'
import { TerminalSessionManager } from './terminalManager'
import { createXtermSession } from './xtermFactory'
import {
  CreateProjectFromDialog,
  CreateTerminal,
  DeleteProject,
  DeleteTerminal,
  DetectTerminalShell,
  GetProjectGitStatus,
  ListProjects,
  LoadTerminalSettings,
  ResizeTerminal,
  SelectProject,
  SelectTerminal,
  SaveTerminalLaunchProfiles,
  SaveTerminalShell,
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
let gitStatusRequestId = 0
const terminalMenu = reactive({
  visible: false,
  terminalId: '',
  x: 0,
  y: 0
})
const defaultTerminalLaunchProfiles = [
  { name: 'codex', command: 'codex' },
  { name: 'claude', command: 'claude' }
]
const terminalSettings = ref(null)
const detectedTerminalShell = ref(null)
const gitStatus = ref(null)
const gitStatusLoading = ref(false)
const gitStatusError = ref('')
const settingsPanel = reactive({
  visible: false,
  loading: false,
  detecting: false,
  saving: false,
  mode: 'detected',
  manualPath: '',
  launchProfiles: [],
  error: ''
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

const projectGitStatusText = computed(() => {
  if (!activeProject.value) {
    return 'No project'
  }
  if (!activeProject.value.available || gitStatus.value?.pathUnavailable) {
    return 'Project path unavailable'
  }
  if (gitStatusLoading.value) {
    return 'Loading git status'
  }
  if (gitStatusError.value) {
    return 'Git status unavailable'
  }
  if (gitStatus.value && !gitStatus.value.isRepo) {
    return 'Not a git repository'
  }
  if (gitStatus.value?.isRepo) {
    const branch = gitStatus.value.branch || '(detached)'
    const changedCount = gitStatus.value.changedCount || 0
    return `${branch} · ${changedCount} changed`
  }
  return 'Git status unavailable'
})

const terminalSettingsDetected = computed(() => {
  return detectedTerminalShell.value || terminalSettings.value?.detected || terminalSettings.value?.fallback || null
})

const terminalSettingsFallback = computed(() => terminalSettings.value?.fallback || null)

onMounted(async () => {
  EventsOn('terminal-output', (event) => {
    terminalManager.write(event.terminalId, event.data)
  })
  EventsOn('terminal-status', (status) => {
    updateTerminalState(status.terminalId, status.state)
  })
  window.addEventListener('resize', fitActiveTerminal)
  window.addEventListener('focus', refreshProjectGitStatus)
  window.addEventListener('click', closeTerminalMenu)

  try {
    applyTerminalSettings(await LoadTerminalSettings())
  } catch (error) {
    showError(error)
  }

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
  window.removeEventListener('focus', refreshProjectGitStatus)
  window.removeEventListener('click', closeTerminalMenu)
})

function applyState(state) {
  const previousActiveProjectId = activeProjectId.value
  const previousTerminals = new Map(terminals.value.map((terminal) => [terminal.id, terminal]))
  projects.value = state?.projects || []
  terminals.value = (state?.terminals || []).map((terminal) => ({
    ...terminal,
    currentCommand:
      terminal.state === 'running'
        ? terminal.currentCommand || previousTerminals.get(terminal.id)?.currentCommand || ''
        : '',
    runtimeTitle: terminal.state === 'running' ? previousTerminals.get(terminal.id)?.runtimeTitle || '' : '',
    idleTitle: terminal.state === 'running' ? previousTerminals.get(terminal.id)?.idleTitle || '' : '',
    activityState: terminal.state === 'running' ? previousTerminals.get(terminal.id)?.activityState || 'idle' : 'idle'
  }))
  activeProjectId.value = state?.activeProjectId || ''
  activeTerminalId.value = state?.activeTerminalId || ''
  for (const terminal of terminals.value) {
    if (terminal.state) {
      shellStatuses[terminal.id] = terminal.state
    }
  }
  closeTerminalMenu()
  syncGitStatusForActiveProject(previousActiveProjectId)
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

async function createTerminal(projectId, launchProfile = null) {
  try {
    const size = terminalManager.size() || { cols: 80, rows: 24 }
    const state = await CreateTerminal(projectId, size.cols || 80, size.rows || 24)
    applyState(state)
    await activateActiveTerminal()
    if (launchProfile?.command && state?.activeTerminalId) {
      await SendTerminalInput(state.activeTerminalId, `${launchProfile.command}\n`)
    }
  } catch (error) {
    showError(error)
  }
}

async function deleteProject(projectId) {
  const project = projects.value.find((candidate) => candidate.id === projectId)
  const projectName = project?.name || 'this project'
  if (!window.confirm(`Delete project "${projectName}" from this app? Files on disk will not be deleted.`)) {
    return
  }
  try {
    applyState(await DeleteProject(projectId))
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
    command: profile.command || ''
  }))
}

function addTerminalLaunchProfile() {
  settingsPanel.launchProfiles.push({ name: '', command: '' })
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
    command: (profile.command || '').trim()
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
    applyTerminalSettings(await SaveTerminalLaunchProfiles(launchProfiles))
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

function syncGitStatusForActiveProject(previousActiveProjectId = '') {
  if (!activeProject.value) {
    gitStatusRequestId += 1
    gitStatus.value = null
    gitStatusLoading.value = false
    gitStatusError.value = ''
    return
  }
  if (!activeProject.value.available) {
    gitStatusRequestId += 1
    gitStatus.value = { projectId: activeProject.value.id, isRepo: false, pathUnavailable: true }
    gitStatusLoading.value = false
    gitStatusError.value = ''
    return
  }
  if (activeProject.value.id !== previousActiveProjectId) {
    refreshProjectGitStatus()
  }
}

async function refreshProjectGitStatus() {
  const project = activeProject.value
  if (!project || !project.available) {
    syncGitStatusForActiveProject(activeProjectId.value)
    return
  }
  const requestId = gitStatusRequestId + 1
  gitStatusRequestId = requestId
  const projectId = project.id
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
    if (requestId === gitStatusRequestId) {
      gitStatusLoading.value = false
    }
  }
}

function errorMessageFrom(error) {
  return error?.message || String(error)
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
      resetTerminalActivity(terminal)
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
    resetTerminalActivity(terminal)
  }
  if (event.type === 'command-end') {
    terminal.currentCommand = ''
    resetTerminalActivity(terminal)
    if (terminal.projectId === activeProjectId.value) {
      refreshProjectGitStatus()
    }
  }
}

function handleTerminalTitleChange(terminalId, title) {
  const terminal = terminals.value.find((candidate) => candidate.id === terminalId)
  if (!terminal) {
    return
  }
  terminal.runtimeTitle = sanitizeCommandLabel(title)
  terminal.activityState = classifyTerminalActivity(terminal, terminal.runtimeTitle)
  if (terminal.activityState === 'idle' && terminal.runtimeTitle && !terminal.idleTitle) {
    terminal.idleTitle = terminal.runtimeTitle
  }
}

function classifyTerminalActivity(terminal, title) {
  const normalizedTitle = normalizeActivityText(title)
  if (!normalizedTitle) {
    return 'idle'
  }
  if (normalizedTitle.includes('!')) {
    return 'needs-input'
  }
  const stableLabel = normalizeActivityText(terminal.currentCommand || terminal.shellName || 'shell')
  const idleTitle = normalizeActivityText(terminal.idleTitle)
  if (!stableLabel || normalizedTitle === stableLabel || normalizedTitle === idleTitle) {
    return 'idle'
  }
  if (hasBusyTitleSignal(normalizedTitle)) {
    return 'busy'
  }
  if (titleLooksLikeStableProgramTitle(normalizedTitle, stableLabel)) {
    return 'idle'
  }
  if (!idleTitle) {
    return 'idle'
  }
  return 'busy'
}

function hasBusyTitleSignal(title) {
  return /[|/\\⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏◐◓◑◒⣾⣽⣻⢿⡿⣟⣯⣷]/.test(title) ||
    /\b(working|thinking|running|processing|executing|busy)\b/.test(title)
}

function titleLooksLikeStableProgramTitle(title, stableLabel) {
  if (!stableLabel) {
    return false
  }
  return title.startsWith(`${stableLabel} `) ||
    title.startsWith(`${stableLabel} -`) ||
    title.startsWith(`${stableLabel}:`) ||
    title.startsWith(`${stableLabel}/`)
}

function normalizeActivityText(value) {
  return (value || '').replace(/\s+/g, ' ').trim().toLowerCase()
}

function resetTerminalActivity(terminal) {
  terminal.runtimeTitle = ''
  terminal.idleTitle = ''
  terminal.activityState = 'idle'
}

function sanitizeCommandLabel(command) {
  return (command || '').replace(/\s+/g, ' ').trim().slice(0, 120)
}

function showError(error) {
  errorMessage.value = errorMessageFrom(error)
}
</script>

<template>
  <main class="app-shell">
    <ProjectSidebar
      :projects="projects"
      :terminals="terminals"
      :active-project-id="activeProjectId"
      :active-terminal-id="activeTerminalId"
      :launch-profiles="terminalLaunchProfiles"
      @create-project="createProject"
      @select-project="selectProject"
      @create-terminal="createTerminal"
      @select-terminal="selectTerminal"
      @delete-project="deleteProject"
      @delete-terminal="deleteTerminal"
    />

    <section class="workspace">
      <header class="workspace-header">
        <div v-if="activeProject" class="project-heading">
          <span class="heading-name">{{ activeProject.name }}</span>
          <span class="heading-path">{{ activeProject.path }}</span>
        </div>
        <div v-else class="project-heading muted">No project selected</div>
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
            v-if="activeProject && activeTerminalState === 'exited'"
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

      <footer class="status-bar">
        <div class="status-item" data-testid="project-git-status">{{ projectGitStatusText }}</div>
        <div v-if="errorMessage" class="status-error">{{ errorMessage }}</div>
      </footer>
    </section>

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
