export const AGENT_PHASE = {
  IDLE: 'idle',
  BUSY: 'busy',
  NEEDS_INPUT: 'needs-input',
  DONE: 'done',
  FAILED: 'failed',
  EXITED: 'exited'
}

export const AGENT_SOURCE = {
  SHELL: 'shell',
  COMMAND_STATE: 'command-state',
  CLAUDE_AGENTS_JSON: 'claude-agents-json',
  CLAUDE_HOOK: 'claude-hook',
  CODEX_JSONL: 'codex-jsonl',
  CODEX_APP_SERVER: 'codex-app-server',
  CODEX_HOOK: 'codex-hook',
  TITLE_FALLBACK: 'title-fallback'
}

export const AGENT_CONFIDENCE = {
  AUTHORITATIVE: 'authoritative',
  STRUCTURED: 'structured',
  HEURISTIC: 'heuristic'
}

const sourcePriority = {
  [AGENT_SOURCE.SHELL]: 20,
  [AGENT_SOURCE.CLAUDE_AGENTS_JSON]: 90,
  [AGENT_SOURCE.CODEX_JSONL]: 90,
  [AGENT_SOURCE.CODEX_APP_SERVER]: 90,
  [AGENT_SOURCE.CLAUDE_HOOK]: 80,
  [AGENT_SOURCE.CODEX_HOOK]: 80,
  [AGENT_SOURCE.COMMAND_STATE]: 50,
  [AGENT_SOURCE.TITLE_FALLBACK]: 30
}

export function createAgentStatus(overrides = {}) {
  return {
    phase: AGENT_PHASE.IDLE,
    source: AGENT_SOURCE.SHELL,
    confidence: AGENT_CONFIDENCE.STRUCTURED,
    reason: 'shell-running',
    label: '',
    updatedAt: 0,
    ...overrides
  }
}

export function applyAgentStatusEvent(terminal, event) {
  const next = {
    ...terminal
  }

  if (event.type === 'shell-status') {
    applyShellStatus(next, event)
    return withActivityState(next)
  }

  if (event.type === 'launch-profile-label') {
    next.currentCommand = sanitizeCommandLabel(event.command)
    next.pendingLaunchProfileCommand = next.currentCommand
    ensureAgentStatus(next, event.at)
    return withActivityState(next)
  }

  if (event.type === 'command-state') {
    applyCommandState(next, event)
    return withActivityState(next)
  }

  if (event.type === 'agent-status') {
    applyStatus(next, createAgentStatus({
      phase: event.phase,
      source: event.source,
      confidence: event.confidence || confidenceForSource(event.source),
      reason: event.reason || event.phase,
      label: sanitizeCommandLabel(event.label),
      updatedAt: event.at || 0
    }))
    return withActivityState(next)
  }

  if (event.type === 'title') {
    next.runtimeTitle = sanitizeCommandLabel(event.title)
    return next
  }

  ensureAgentStatus(next, event.at)
  return withActivityState(next)
}

export function activityStateFromAgentStatus(status) {
  if (status?.phase === AGENT_PHASE.BUSY || status?.phase === AGENT_PHASE.NEEDS_INPUT) {
    return status.phase
  }
  return AGENT_PHASE.IDLE
}

export function claudeHookEventToAgentStatus(event = {}) {
  const hookName = event.hook_event_name || event.hookEventName || event.event || event.type
  const notificationType = normalizeReason(event.notification_type || event.notificationType || event.matcher)

  if (hookName === 'Notification') {
    if (['permission-prompt', 'idle-prompt', 'elicitation-dialog'].includes(notificationType)) {
      return agentEvent(AGENT_PHASE.NEEDS_INPUT, AGENT_SOURCE.CLAUDE_HOOK, 'permission-prompt', event)
    }
    return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CLAUDE_HOOK, `notification-${notificationType || 'received'}`, event)
  }

  if (hookName === 'UserPromptSubmit' || hookName === 'PreToolUse' || hookName === 'PostToolBatch') {
    return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CLAUDE_HOOK, normalizeReason(hookName), event)
  }

  if (hookName === 'PostToolUse' || hookName === 'SubagentStop' || hookName === 'Stop') {
    return agentEvent(AGENT_PHASE.IDLE, AGENT_SOURCE.CLAUDE_HOOK, normalizeReason(hookName), event)
  }

  if (hookName === 'StopFailure') {
    return agentEvent(AGENT_PHASE.FAILED, AGENT_SOURCE.CLAUDE_HOOK, 'stop-failure', event)
  }

  if (hookName === 'SessionEnd') {
    return agentEvent(AGENT_PHASE.DONE, AGENT_SOURCE.CLAUDE_HOOK, 'session-end', event)
  }

  return null
}

export function claudeAgentsJsonToAgentStatus(session = {}) {
  const state = normalizeReason(session.state)
  const status = normalizeReason(session.status)
  const waitingFor = normalizeReason(session.waitingFor || session.waiting_for)

  if (state === 'blocked' || status === 'waiting') {
    const reason = waitingFor || 'blocked'
    return agentEvent(
      waitingFor.includes('permission') || waitingFor.includes('input') ? AGENT_PHASE.NEEDS_INPUT : AGENT_PHASE.BUSY,
      AGENT_SOURCE.CLAUDE_AGENTS_JSON,
      reason,
      session,
      AGENT_CONFIDENCE.AUTHORITATIVE
    )
  }
  if (state === 'working' || status === 'running') {
    return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CLAUDE_AGENTS_JSON, state || status, session, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  if (state === 'done' || state === 'completed') {
    return agentEvent(AGENT_PHASE.DONE, AGENT_SOURCE.CLAUDE_AGENTS_JSON, state, session, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  if (state === 'failed') {
    return agentEvent(AGENT_PHASE.FAILED, AGENT_SOURCE.CLAUDE_AGENTS_JSON, state, session, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  if (state === 'stopped') {
    return agentEvent(AGENT_PHASE.EXITED, AGENT_SOURCE.CLAUDE_AGENTS_JSON, state, session, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  return null
}

export function codexJsonlEventToAgentStatus(event = {}) {
  switch (event.type) {
    case 'turn.started':
      return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CODEX_JSONL, 'turn-started', event, AGENT_CONFIDENCE.AUTHORITATIVE)
    case 'turn.completed':
      return agentEvent(AGENT_PHASE.DONE, AGENT_SOURCE.CODEX_JSONL, 'turn-completed', event, AGENT_CONFIDENCE.AUTHORITATIVE)
    case 'turn.failed':
    case 'error':
      return agentEvent(AGENT_PHASE.FAILED, AGENT_SOURCE.CODEX_JSONL, normalizeReason(event.type), event, AGENT_CONFIDENCE.AUTHORITATIVE)
    case 'item.started':
      return codexItemEventToAgentStatus(event.item, event, AGENT_SOURCE.CODEX_JSONL)
    default:
      return null
  }
}

export function codexAppServerEventToAgentStatus(message = {}) {
  if (message.method === 'turn/started') {
    return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CODEX_APP_SERVER, 'turn-started', message, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  if (message.method === 'turn/completed') {
    return agentEvent(AGENT_PHASE.DONE, AGENT_SOURCE.CODEX_APP_SERVER, 'turn-completed', message, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  if (message.method === 'item/started') {
    return codexItemEventToAgentStatus(message.params?.item, message, AGENT_SOURCE.CODEX_APP_SERVER)
  }
  if (message.method === 'error') {
    return agentEvent(AGENT_PHASE.FAILED, AGENT_SOURCE.CODEX_APP_SERVER, 'error', message, AGENT_CONFIDENCE.AUTHORITATIVE)
  }
  return null
}

export function codexHookEventToAgentStatus(event = {}) {
  const hookName = event.hook_event_name || event.hookEventName || event.event || event.type
  if (hookName === 'UserPromptSubmit' || hookName === 'PreToolUse' || hookName === 'SubagentStart') {
    return agentEvent(AGENT_PHASE.BUSY, AGENT_SOURCE.CODEX_HOOK, normalizeReason(hookName), event)
  }
  if (hookName === 'PostToolUse' || hookName === 'SubagentStop' || hookName === 'Stop') {
    return agentEvent(AGENT_PHASE.IDLE, AGENT_SOURCE.CODEX_HOOK, normalizeReason(hookName), event)
  }
  return null
}

export function detectAgentStatusCapabilities(sources = {}) {
  const capabilities = {
    claudeAgentsJson: Boolean(sources.claudeAgentsJson),
    claudeHooks: Boolean(sources.claudeHooks),
    codexJsonl: Boolean(sources.codexJsonl),
    codexAppServer: Boolean(sources.codexAppServer),
    codexHooks: Boolean(sources.codexHooks)
  }
  return {
    ...capabilities,
    structured: Object.values(capabilities).some(Boolean)
  }
}

function applyShellStatus(terminal, event) {
  terminal.state = event.state
  if (event.state === 'running') {
    applyStatus(terminal, createAgentStatus({ updatedAt: event.at || 0 }))
    return
  }
  terminal.currentCommand = ''
  terminal.pendingLaunchProfileCommand = ''
  resetTitleFallback(terminal)
  applyStatus(terminal, createAgentStatus({
    phase: AGENT_PHASE.EXITED,
    source: AGENT_SOURCE.SHELL,
    confidence: AGENT_CONFIDENCE.STRUCTURED,
    reason: 'shell-exited',
    updatedAt: event.at || 0
  }))
}

function applyCommandState(terminal, event) {
  if (event.commandType === 'command-start' || event.commandType === 'command-started') {
    terminal.currentCommand = sanitizeCommandLabel(event.command)
    terminal.pendingLaunchProfileCommand = ''
    resetTitleFallback(terminal)
    applyStatus(terminal, createAgentStatus({
      phase: AGENT_PHASE.IDLE,
      source: AGENT_SOURCE.COMMAND_STATE,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'command-start',
      label: terminal.currentCommand,
      updatedAt: event.at || 0
    }))
    return
  }

  if (event.commandType === 'command-end' || event.commandType === 'command-ended') {
    if (terminal.pendingLaunchProfileCommand) {
      terminal.pendingLaunchProfileCommand = ''
      return
    }
    terminal.currentCommand = ''
    resetTitleFallback(terminal)
    applyStatus(terminal, createAgentStatus({
      phase: AGENT_PHASE.IDLE,
      source: AGENT_SOURCE.COMMAND_STATE,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'command-end',
      updatedAt: event.at || 0
    }))
  }
}

function applyStatus(terminal, candidate) {
  const current = terminal.agentStatus
  if (!current || shouldReplaceStatus(current, candidate)) {
    terminal.agentStatus = candidate
  }
}

function shouldReplaceStatus(current, candidate) {
  if (current.phase === AGENT_PHASE.EXITED && candidate.source !== AGENT_SOURCE.SHELL) {
    return false
  }
  if (candidate.phase === AGENT_PHASE.EXITED && candidate.source === AGENT_SOURCE.SHELL) {
    return true
  }
  if (
    current.source === AGENT_SOURCE.COMMAND_STATE &&
    current.phase === AGENT_PHASE.IDLE &&
    candidate.source === AGENT_SOURCE.TITLE_FALLBACK
  ) {
    return true
  }
  if (priorityOf(candidate.source) > priorityOf(current.source)) {
    return true
  }
  if (priorityOf(candidate.source) < priorityOf(current.source)) {
    return false
  }
  return (candidate.updatedAt || 0) >= (current.updatedAt || 0)
}

function ensureAgentStatus(terminal, updatedAt = 0) {
  if (!terminal.agentStatus) {
    terminal.agentStatus = createAgentStatus({ updatedAt: updatedAt || 0 })
  }
}

function withActivityState(terminal) {
  ensureAgentStatus(terminal)
  terminal.activityState = activityStateFromAgentStatus(terminal.agentStatus)
  return terminal
}

function confidenceForSource(source) {
  if (
    source === AGENT_SOURCE.CLAUDE_AGENTS_JSON ||
    source === AGENT_SOURCE.CODEX_JSONL ||
    source === AGENT_SOURCE.CODEX_APP_SERVER
  ) {
    return AGENT_CONFIDENCE.AUTHORITATIVE
  }
  if (source === AGENT_SOURCE.TITLE_FALLBACK) {
    return AGENT_CONFIDENCE.HEURISTIC
  }
  return AGENT_CONFIDENCE.STRUCTURED
}

function codexItemEventToAgentStatus(item, rawEvent, source) {
  const itemType = item?.type || ''
  if (['command_execution', 'tool_call', 'mcp_tool_call', 'web_search'].includes(itemType)) {
    return agentEvent(
      AGENT_PHASE.BUSY,
      source,
      `item-${normalizeReason(itemType)}-started`,
      rawEvent,
      AGENT_CONFIDENCE.AUTHORITATIVE
    )
  }
  return null
}

function agentEvent(phase, source, reason, rawEvent, confidence = confidenceForSource(source)) {
  return {
    type: 'agent-status',
    phase,
    source,
    confidence,
    reason,
    label: rawEvent?.label || rawEvent?.name || '',
    at: normalizeTimestamp(rawEvent?.updatedAt ?? rawEvent?.at ?? rawEvent?.timestamp)
  }
}

function normalizeTimestamp(value) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string') {
    const numeric = Number(value)
    if (Number.isFinite(numeric) && value.trim() !== '') {
      return numeric
    }
    const parsed = Date.parse(value)
    if (!Number.isNaN(parsed)) {
      return parsed
    }
  }
  return 0
}

function normalizeReason(value) {
  return String(value || '')
    .replace(/_/g, '-')
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .replace(/\s+/g, '-')
    .toLowerCase()
}

function priorityOf(source) {
  return sourcePriority[source] || 0
}

function resetTitleFallback(terminal) {
  terminal.runtimeTitle = ''
}

function sanitizeCommandLabel(command) {
  return (command || '').replace(/\s+/g, ' ').trim().slice(0, 120)
}
