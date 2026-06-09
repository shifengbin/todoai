import { describe, expect, it, vi } from 'vitest'
import { TerminalSessionManager } from './terminalManager'

describe('TerminalSessionManager', () => {
  it('preserves one terminal instance per project while routing inactive output', () => {
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput: vi.fn(),
      resizeTerminal: vi.fn()
    })
    const containerA = document.createElement('div')
    const containerB = document.createElement('div')

    manager.activate('project-a', containerA)
    manager.write('project-a', 'first')
    manager.activate('project-b', containerB)
    manager.write('project-a', 'background')
    manager.activate('project-a', containerA)

    expect(factory.createdFor).toEqual(['project-a', 'project-b'])
    expect(factory.sessions.get('project-a').terminal.openedIn).toBe(containerA)
    expect(factory.sessions.get('project-a').terminal.writes).toEqual(['first', 'background'])
    expect(factory.sessions.get('project-a').terminal.openCount).toBe(1)
  })

  it('routes terminal input with the project id that owns the terminal', () => {
    const sendInput = vi.fn()
    const factory = createFakeTerminalFactory()
    const manager = new TerminalSessionManager({
      createSession: factory.createSession,
      sendInput,
      resizeTerminal: vi.fn()
    })

    manager.activate('project-b', document.createElement('div'))
    factory.sessions.get('project-b').terminal.emitData('pwd\n')

    expect(sendInput).toHaveBeenCalledWith('project-b', 'pwd\n')
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

    manager.activate('project-a', document.createElement('div'))
    factory.sessions.get('project-a').terminal.selection = 'git status'

    await manager.copySelection('project-a')

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

    manager.activate('project-b', document.createElement('div'))

    await manager.paste('project-b')

    expect(sendInput).toHaveBeenCalledWith('project-b', 'npm test\n')
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

    manager.activate('project-b', document.createElement('div'))

    await manager.paste('project-b')

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

    manager.activate('project-a', document.createElement('div'))
    factory.sessions.get('project-a').terminal.selection = 'selected'

    await manager.copySelection('project-a')

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

    manager.activate('project-a', document.createElement('div'))
    manager.fitActive()

    const session = factory.sessions.get('project-a')
    expect(session.fitAddon.fit).toHaveBeenCalledTimes(2)
    expect(resizeTerminal).toHaveBeenLastCalledWith('project-a', 100, 32)
  })
})

function createFakeTerminalFactory() {
  const sessions = new Map()
  const createdFor = []

  return {
    sessions,
    createdFor,
    createSession(projectId, onData) {
      createdFor.push(projectId)
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
      const session = { terminal, fitAddon }
      sessions.set(projectId, session)
      return session
    }
  }
}
