import { describe, expect, it, vi } from 'vitest'
import { TerminalSessionManager } from './terminalManager'

describe('TerminalSessionManager', () => {
  it('replays restored output into the xterm session once', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', 'restored output line 1\nrestored output line 2\n')

    const session = factory.sessions.get('terminal-a')
    expect(session.terminal.writes).toEqual(['restored output line 1\nrestored output line 2\n'])
  })

  it('does not replay history a second time for the same terminal', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', 'first replay\n')
    manager.replayHistory('terminal-a', 'second replay\n')

    const session = factory.sessions.get('terminal-a')
    expect(session.terminal.writes).toEqual(['first replay\n'])
  })

  it('does not send terminal responses generated while replaying restored history', () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    factory.sessions.get('terminal-a').terminal.responseForNextWrite = '\x1b[?1;2c'

    manager.replayHistory('terminal-a', '\x1b[c')
    factory.sessions.get('terminal-a').terminal.emitData('typed command\n')

    expect(sendInput).toHaveBeenCalledTimes(1)
    expect(sendInput).toHaveBeenCalledWith('terminal-a', 'typed command\n')
  })

  it('does not replay empty history', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', '')

    const session = factory.sessions.get('terminal-a')
    expect(session.terminal.writes).toEqual([])
  })

  it('clears replay flag when terminal is disposed', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', 'first\n')
    manager.dispose('terminal-a')
    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', 'second\n')

    const session = factory.sessions.get('terminal-a')
    expect(session.terminal.writes).toEqual(['second\n'])
  })

  it('does not interfere with live output after replay', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const container = document.createElement('div')

    manager.activate('terminal-a', container)
    manager.replayHistory('terminal-a', 'restored\n')
    manager.write('terminal-a', 'live output\n')

    const session = factory.sessions.get('terminal-a')
    expect(session.terminal.writes).toEqual(['restored\n', 'live output\n'])
  })

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

  it('copies selected Unicode terminal text through the clipboard writer', async () => {
    const writeText = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      clipboard: { writeText }
    })

    manager.activate('terminal-a', document.createElement('div'))
    factory.sessions.get('terminal-a').terminal.selection = '中文 ✓ 🔧   '

    await manager.copySelection('terminal-a')

    expect(writeText).toHaveBeenCalledWith('中文 ✓ 🔧   ')
  })

  it('pastes non-empty Unicode clipboard text into the owning project shell', async () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const text = "printf '中文 ✓ 🔧   \\n'"
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn(),
      clipboard: { readText: vi.fn().mockResolvedValue(text) }
    })

    manager.activate('terminal-b', document.createElement('div'))

    await manager.paste('terminal-b')

    expect(sendInput).toHaveBeenCalledWith('terminal-b', text)
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

  it('focuses the terminal session for the provided terminal id', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })

    manager.activate('terminal-a', document.createElement('div'))
    manager.activate('terminal-b', document.createElement('div'))

    manager.focus('terminal-a')

    expect(factory.sessions.get('terminal-a').terminal.focus).toHaveBeenCalledTimes(1)
    expect(factory.sessions.get('terminal-b').terminal.focus).not.toHaveBeenCalled()
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

  it('does not pass application appearance themes into terminal sessions', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn(),
      theme: 'light'
    })
    const containerA = document.createElement('div')

    manager.activate('terminal-a', containerA)

    expect(factory.sessions.get('terminal-a').theme).toBeUndefined()
    expect(factory.sessions.get('terminal-a').terminal.openCount).toBe(1)
  })

  it('disposes a terminal session and recreates it on the next activation', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const containerA = document.createElement('div')

    manager.activate('terminal-a', containerA)
    const firstSession = factory.sessions.get('terminal-a')

    manager.dispose('terminal-a')
    manager.write('terminal-a', 'ignored')
    manager.activate('terminal-a', containerA)

    expect(firstSession.terminal.dispose).toHaveBeenCalledTimes(1)
    expect(firstSession.terminal.writes).toEqual([])
    expect(factory.createdFor).toEqual(['terminal-a', 'terminal-a'])
    expect(factory.sessions.get('terminal-a')).not.toBe(firstSession)
  })
})

function createFakeTerminalFactory() {
  const sessions = new Map()
  const createdFor = []

  return {
    sessions,
    createdFor,
    createSession(terminalId, onData, onShortcut, onCommandState, onTitleChange, theme) {
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
        write(data, callback) {
          this.writes.push(data)
          if (this.responseForNextWrite) {
            this.emitData(this.responseForNextWrite)
            this.responseForNextWrite = ''
          }
          callback?.()
        },
        focus: vi.fn(),
        dispose: vi.fn(),
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
        theme,
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
