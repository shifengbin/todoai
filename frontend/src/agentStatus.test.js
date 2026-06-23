import { describe, expect, it } from 'vitest'
import {
  AGENT_CONFIDENCE,
  AGENT_PHASE,
  AGENT_SOURCE,
  applyAgentStatusEvent,
  codexAppServerEventToAgentStatus,
  codexJsonlEventToAgentStatus,
  createAgentStatus,
  detectAgentStatusCapabilities,
  claudeAgentsJsonToAgentStatus,
  claudeHookEventToAgentStatus
} from './agentStatus'

describe('agent status reducer', () => {
  it('creates an idle shell status for a running terminal', () => {
    const terminal = applyAgentStatusEvent(baseTerminal(), {
      type: 'shell-status',
      state: 'running',
      at: 10
    })

    expect(terminal.agentStatus).toEqual({
      phase: AGENT_PHASE.IDLE,
      source: AGENT_SOURCE.SHELL,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'shell-running',
      label: '',
      updatedAt: 10
    })
    expect(terminal.activityState).toBe('idle')
  })

  it('keeps shell exit above later title events', () => {
    const busy = applyAgentStatusEvent(baseTerminal({ currentCommand: 'codex' }), {
      type: 'agent-status',
      phase: AGENT_PHASE.BUSY,
      source: AGENT_SOURCE.CODEX_JSONL,
      confidence: AGENT_CONFIDENCE.AUTHORITATIVE,
      reason: 'turn-started',
      at: 10
    })
    const exited = applyAgentStatusEvent(busy, {
      type: 'shell-status',
      state: 'exited',
      at: 11
    })
    const afterTitle = applyAgentStatusEvent(exited, {
      type: 'title',
      title: 'codex thinking',
      at: 12
    })

    expect(afterTitle.agentStatus.phase).toBe(AGENT_PHASE.EXITED)
    expect(afterTitle.agentStatus.source).toBe(AGENT_SOURCE.SHELL)
    expect(afterTitle.activityState).toBe('idle')
  })

  it('does not clear a structured busy state with a stable title event', () => {
    const busy = applyAgentStatusEvent(baseTerminal({ currentCommand: 'codex' }), {
      type: 'agent-status',
      phase: AGENT_PHASE.BUSY,
      source: AGENT_SOURCE.CODEX_JSONL,
      confidence: AGENT_CONFIDENCE.AUTHORITATIVE,
      reason: 'turn-started',
      at: 10
    })
    const afterTitle = applyAgentStatusEvent(busy, {
      type: 'title',
      title: 'codex',
      at: 11
    })

    expect(afterTitle.agentStatus.phase).toBe(AGENT_PHASE.BUSY)
    expect(afterTitle.agentStatus.source).toBe(AGENT_SOURCE.CODEX_JSONL)
    expect(afterTitle.activityState).toBe('busy')
  })

  it('updates runtimeTitle without changing agent status from title events', () => {
    const withoutStatus = applyAgentStatusEvent(baseTerminal({ currentCommand: 'codex' }), {
      type: 'title',
      title: '  codex   thinking  ',
      at: 9
    })
    expect(withoutStatus.runtimeTitle).toBe('codex thinking')
    expect(withoutStatus.agentStatus).toBeUndefined()
    expect(withoutStatus.activityState).toBe('idle')

    const structured = applyAgentStatusEvent(baseTerminal({ currentCommand: 'codex' }), {
      type: 'agent-status',
      phase: AGENT_PHASE.BUSY,
      source: AGENT_SOURCE.CODEX_JSONL,
      confidence: AGENT_CONFIDENCE.AUTHORITATIVE,
      reason: 'turn-started',
      at: 10
    })
    const afterTitle = applyAgentStatusEvent(structured, {
      type: 'title',
      title: '  codex   thinking  ',
      at: 11
    })

    expect(afterTitle.runtimeTitle).toBe('codex thinking')
    expect(afterTitle.agentStatus).toEqual(structured.agentStatus)
    expect(afterTitle.activityState).toBe('busy')
  })

  it('lets structured needs-input apply after a title event', () => {
    const busy = applyAgentStatusEvent(baseTerminal({ currentCommand: 'claude', activityState: 'busy' }), {
      type: 'title',
      title: 'claude thinking',
      at: 10
    })
    const needsInput = applyAgentStatusEvent(busy, {
      type: 'agent-status',
      phase: AGENT_PHASE.NEEDS_INPUT,
      source: AGENT_SOURCE.CLAUDE_HOOK,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'permission-prompt',
      at: 11
    })

    expect(needsInput.agentStatus.phase).toBe(AGENT_PHASE.NEEDS_INPUT)
    expect(needsInput.agentStatus.source).toBe(AGENT_SOURCE.CLAUDE_HOOK)
    expect(needsInput.activityState).toBe('needs-input')
  })

  it('sets a launch profile command label without marking the terminal busy', () => {
    const terminal = applyAgentStatusEvent(baseTerminal({ shellName: 'pwsh' }), {
      type: 'launch-profile-label',
      command: 'claude --dangerously-skip-permissions',
      at: 10
    })

    expect(terminal.currentCommand).toBe('claude --dangerously-skip-permissions')
    expect(terminal.agentStatus.phase).toBe(AGENT_PHASE.IDLE)
    expect(terminal.agentStatus.source).toBe(AGENT_SOURCE.SHELL)
    expect(terminal.activityState).toBe('idle')
  })

  it('keeps launch profile label across an unpaired command end', () => {
    const labeled = applyAgentStatusEvent(baseTerminal({ shellName: 'zsh' }), {
      type: 'launch-profile-label',
      command: 'codex',
      at: 10
    })
    const idleEnd = applyAgentStatusEvent(labeled, {
      type: 'command-state',
      commandType: 'command-end',
      at: 11
    })

    expect(idleEnd.currentCommand).toBe('codex')
    expect(idleEnd.pendingLaunchProfileCommand).toBe('')

    const started = applyAgentStatusEvent(idleEnd, {
      type: 'command-state',
      commandType: 'command-start',
      command: 'codex',
      at: 12
    })
    const realEnd = applyAgentStatusEvent(started, {
      type: 'command-state',
      commandType: 'command-end',
      at: 13
    })

    expect(realEnd.currentCommand).toBe('')
  })

  it('keeps Claude launch profile stable titles idle', () => {
    const started = applyAgentStatusEvent(baseTerminal(), {
      type: 'command-state',
      commandType: 'command-start',
      command: 'claude --dangerously-skip-permissions',
      at: 10
    })
    const initialTitle = applyAgentStatusEvent(started, {
      type: 'title',
      title: 'Claude',
      at: 11
    })
    const stableTitle = applyAgentStatusEvent(initialTitle, {
      type: 'title',
      title: 'Claude Code',
      at: 12
    })

    expect(stableTitle.agentStatus.phase).toBe(AGENT_PHASE.IDLE)
    expect(stableTitle.activityState).toBe('idle')
  })

  it('keeps Claude launch profile activity idle through title changes', () => {
    const started = applyAgentStatusEvent(baseTerminal(), {
      type: 'command-state',
      commandType: 'command-start',
      command: 'claude --dangerously-skip-permissions',
      at: 10
    })
    const stableTitle = applyAgentStatusEvent(started, {
      type: 'title',
      title: 'Claude Code',
      at: 11
    })
    const busyTitle = applyAgentStatusEvent(stableTitle, {
      type: 'title',
      title: 'Claude thinking',
      at: 12
    })
    const backToStableTitle = applyAgentStatusEvent(busyTitle, {
      type: 'title',
      title: 'Claude Code',
      at: 13
    })

    expect(busyTitle.runtimeTitle).toBe('Claude thinking')
    expect(busyTitle.agentStatus).toEqual(stableTitle.agentStatus)
    expect(busyTitle.activityState).toBe('idle')
    expect(backToStableTitle.runtimeTitle).toBe('Claude Code')
    expect(backToStableTitle.agentStatus.phase).toBe(AGENT_PHASE.IDLE)
    expect(backToStableTitle.activityState).toBe('idle')
  })

  it('creates default status objects with explicit fields', () => {
    expect(createAgentStatus({ updatedAt: 42 })).toEqual({
      phase: AGENT_PHASE.IDLE,
      source: AGENT_SOURCE.SHELL,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'shell-running',
      label: '',
      updatedAt: 42
    })
  })

  it('maps Claude hook notifications to needs-input status', () => {
    expect(claudeHookEventToAgentStatus({
      hook_event_name: 'Notification',
      notification_type: 'permission_prompt'
    })).toMatchObject({
      type: 'agent-status',
      phase: AGENT_PHASE.NEEDS_INPUT,
      source: AGENT_SOURCE.CLAUDE_HOOK,
      confidence: AGENT_CONFIDENCE.STRUCTURED,
      reason: 'permission-prompt'
    })
  })

  it('maps Claude agents JSON waiting state to needs-input status', () => {
    expect(claudeAgentsJsonToAgentStatus({
      state: 'blocked',
      status: 'waiting',
      waitingFor: 'permission prompt'
    })).toMatchObject({
      type: 'agent-status',
      phase: AGENT_PHASE.NEEDS_INPUT,
      source: AGENT_SOURCE.CLAUDE_AGENTS_JSON,
      confidence: AGENT_CONFIDENCE.AUTHORITATIVE,
      reason: 'permission-prompt'
    })
  })

  it('maps Codex JSONL turn events to busy done and failed statuses', () => {
    expect(codexJsonlEventToAgentStatus({ type: 'turn.started' })).toMatchObject({
      phase: AGENT_PHASE.BUSY,
      source: AGENT_SOURCE.CODEX_JSONL,
      reason: 'turn-started'
    })
    expect(codexJsonlEventToAgentStatus({ type: 'turn.completed' })).toMatchObject({
      phase: AGENT_PHASE.DONE,
      source: AGENT_SOURCE.CODEX_JSONL,
      reason: 'turn-completed'
    })
    expect(codexJsonlEventToAgentStatus({ type: 'error' })).toMatchObject({
      phase: AGENT_PHASE.FAILED,
      source: AGENT_SOURCE.CODEX_JSONL,
      reason: 'error'
    })
  })

  it('normalizes structured ISO timestamps before status ordering', () => {
    const older = codexJsonlEventToAgentStatus({
      type: 'turn.started',
      timestamp: '2026-06-15T09:00:00Z'
    })
    expect(older.at).toBe(Date.parse('2026-06-15T09:00:00Z'))

    const terminal = applyAgentStatusEvent(baseTerminal(), older)
    const newer = applyAgentStatusEvent(terminal, codexJsonlEventToAgentStatus({
      type: 'turn.completed',
      timestamp: '2026-06-15T09:01:00Z'
    }))

    expect(newer.agentStatus.phase).toBe(AGENT_PHASE.DONE)
    expect(newer.agentStatus.updatedAt).toBe(Date.parse('2026-06-15T09:01:00Z'))
  })

  it('maps Codex app-server item notifications to busy status', () => {
    expect(codexAppServerEventToAgentStatus({
      method: 'item/started',
      params: { item: { type: 'command_execution' } }
    })).toMatchObject({
      type: 'agent-status',
      phase: AGENT_PHASE.BUSY,
      source: AGENT_SOURCE.CODEX_APP_SERVER,
      confidence: AGENT_CONFIDENCE.AUTHORITATIVE,
      reason: 'item-command-execution-started'
    })
  })

  it('detects status capabilities from available sources', () => {
    expect(detectAgentStatusCapabilities({
      claudeAgentsJson: true,
      claudeHooks: false,
      codexJsonl: true,
      codexAppServer: false,
      codexHooks: true
    })).toEqual({
      claudeAgentsJson: true,
      claudeHooks: false,
      codexJsonl: true,
      codexAppServer: false,
      codexHooks: true,
      structured: true
    })
  })
})

function baseTerminal(overrides = {}) {
  return {
    id: 'terminal-a',
    shellName: 'zsh',
    currentCommand: '',
    state: 'running',
    runtimeTitle: '',
    activityState: 'idle',
    ...overrides
  }
}
