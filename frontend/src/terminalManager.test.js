import { describe, expect, it, vi } from 'vitest'
import { TerminalSessionManager } from './terminalManager'

describe('TerminalSessionManager', () => {
  it('preserves one xterm instance per terminal while routing inactive output', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const containerA = document.createElement('div')
    const containerB = document.createElement('div')

    manager.activate('terminal-a', containerA)
    manager.write('terminal-a', 'first')
    manager.activate('terminal-b', containerB)
    manager.write('terminal-a', 'background')
    manager.activate('terminal-a', containerA)

    expect(factory.createdFor).toEqual(['terminal-a', 'terminal-b'])
    expect(factory.sessions.get('terminal-a').terminal.openedIn).toBe(containerA)
    expect(factory.sessions.get('terminal-a').terminal.writes).toEqual(['first', 'background'])
    expect(factory.sessions.get('terminal-a').terminal.openCount).toBe(1)
  })

  it('routes terminal input with the terminal id that owns the xterm instance', () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn()
    })

    manager.activate('terminal-b', document.createElement('div'))
    factory.sessions.get('terminal-b').terminal.emitData('pwd\n')

    expect(sendInput).toHaveBeenCalledWith('terminal-b', 'pwd\n')
  })

  it('copies selected terminal text through the clipboard writer', async () => {
    const writeText = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      clipboard: { writeText }
    })

    manager.activate('terminal-a', document.createElement('div'))
    factory.sessions.get('terminal-a').terminal.selection = 'git status'

    await manager.copySelection('terminal-a')

    expect(writeText).toHaveBeenCalledWith('git status')
  })

  it('pastes non-empty clipboard text into the owning project shell', async () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn(),
      clipboard: { readText: vi.fn().mockResolvedValue('npm test\n') }
    })

    manager.activate('terminal-b', document.createElement('div'))

    await manager.paste('terminal-b')

    expect(sendInput).toHaveBeenCalledWith('terminal-b', 'npm test\n')
  })

  it('ignores empty clipboard text when pasting', async () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn(),
      clipboard: { readText: vi.fn().mockResolvedValue('') }
    })

    manager.activate('terminal-b', document.createElement('div'))

    await manager.paste('terminal-b')

    expect(sendInput).not.toHaveBeenCalled()
  })

  it('reports clipboard errors through the configured error handler', async () => {
    const error = new Error('clipboard unavailable')
    const onError = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      clipboard: { writeText: vi.fn().mockRejectedValue(error) },
      onError
    })

    manager.activate('terminal-a', document.createElement('div'))
    factory.sessions.get('terminal-a').terminal.selection = 'selected'

    await manager.copySelection('terminal-a')

    expect(onError).toHaveBeenCalledWith(error)
  })

  it('fits and reports the active terminal size', () => {
    const resizeTerminal = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal
    })

    manager.activate('terminal-a', document.createElement('div'))
    manager.fitActive()

    const session = factory.sessions.get('terminal-a')
    expect(session.fitAddon.fit).toHaveBeenCalledTimes(2)
    expect(resizeTerminal).toHaveBeenLastCalledWith('terminal-a', 100, 32)
  })

  it('forwards command-state events with the terminal id', () => {
    const onCommandState = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      onCommandState
    })

    manager.activate('terminal-a', document.createElement('div'))
    factory.sessions.get('terminal-a').emitCommandState({ type: 'command-start', command: 'npm test' })

    expect(onCommandState).toHaveBeenCalledWith('terminal-a', {
      type: 'command-start',
      command: 'npm test'
    })
  })

  it('forwards title-change events with the terminal id that owns the session', () => {
    const onTitleChange = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      onTitleChange
    })

    manager.activate('terminal-a', document.createElement('div'))
    manager.activate('terminal-b', document.createElement('div'))
    factory.sessions.get('terminal-a').emitTitleChange('! codex')

    expect(onTitleChange).toHaveBeenCalledWith('terminal-a', '! codex')
  })
})

function createFakeTerminalFactory() {
  const sessions = new Map()
  const createdFor = []

  return {
    sessions,
    createdFor,
    createSession(terminalId, onData, onShortcut, onCommandState, onTitleChange) {
      createdFor.push(terminalId)
      const terminal = {
        cols: 100,
        rows: 32,
        writes: [],
        openedIn: null,
        openCount: 0,
        open(container) {
          this.openedIn = container
          this.openCount += 1
        },
        write(data) {
          this.writes.push(data)
        },
        hasSelection() {
          return Boolean(this.selection)
        },
        getSelection() {
          return this.selection || ''
        },
        onData(handler) {
          this.dataHandler = handler
        },
        emitData(data) {
          this.dataHandler(data)
        }
      }
      terminal.onData(onData)
      const fitAddon = {
        fit: vi.fn()
      }
      const session = {
        terminal,
        fitAddon,
        emitCommandState(event) {
          onCommandState(event)
        },
        emitTitleChange(title) {
          onTitleChange(title)
        }
      }
      sessions.set(terminalId, session)
      return session
    }
  }
}
