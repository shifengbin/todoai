import { describe, expect, it, vi } from 'vitest'
import { createXtermSession } from './xtermFactory'

const terminalMock = vi.hoisted(() => ({ lastTerminal: null }))

vi.mock('@xterm/xterm', () => {
  return {
    Terminal: class FakeTerminal {
      constructor(options = {}) {
        this.options = options
        terminalMock.lastTerminal = this
      }

      loadAddon() {}

      onData(handler) {
        this.dataHandler = handler
      }

      onTitleChange(handler) {
        this.titleHandler = handler
        return { dispose() {} }
      }

      attachCustomKeyEventHandler(handler) {
        this.keyHandler = handler
      }

      get parser() {
        return {
          registerOscHandler: (ident, handler) => {
            this.oscHandlers ||= new Map()
            this.oscHandlers.set(ident, handler)
            return { dispose() {} }
          }
        }
      }
    }
  }
})

vi.mock('@xterm/addon-fit', () => {
  return {
    FitAddon: class FakeFitAddon {
      fit() {}
    }
  }
})

describe('createXtermSession', () => {
  it('keeps the fixed terminal theme even when an appearance theme is passed', () => {
    createXtermSession('terminal-a', vi.fn(), vi.fn(), vi.fn(), vi.fn(), 'light')

    expect(terminalMock.lastTerminal.options.theme.background).toBe('#111418')
    expect(terminalMock.lastTerminal.options.theme.foreground).toBe('#e6edf3')
    expect(terminalMock.lastTerminal.options.theme.cursor).toBe('#f0c674')
  })

  it('does not enable client-side EOL conversion for PTY output', () => {
    createXtermSession('project-a', vi.fn(), vi.fn())

    expect(terminalMock.lastTerminal.options.convertEol).toBeUndefined()
  })

  it('intercepts Ctrl+Shift+C for copy', () => {
    const onShortcut = vi.fn()

    createXtermSession('project-a', vi.fn(), onShortcut)
    const result = terminalMock.lastTerminal.keyHandler(keyEvent({ key: 'C', ctrlKey: true, shiftKey: true }))

    expect(result).toBe(false)
    expect(onShortcut).toHaveBeenCalledWith('copy')
  })

  it('intercepts Ctrl+Shift+V for paste', () => {
    const onShortcut = vi.fn()

    createXtermSession('project-a', vi.fn(), onShortcut)
    const result = terminalMock.lastTerminal.keyHandler(keyEvent({ key: 'v', ctrlKey: true, shiftKey: true }))

    expect(result).toBe(false)
    expect(onShortcut).toHaveBeenCalledWith('paste')
  })

  it('does not consume plain Ctrl+C', () => {
    const onShortcut = vi.fn()

    createXtermSession('project-a', vi.fn(), onShortcut)
    const result = terminalMock.lastTerminal.keyHandler(keyEvent({ key: 'c', ctrlKey: true }))

    expect(result).toBe(true)
    expect(onShortcut).not.toHaveBeenCalled()
  })

  it('emits command state from TodoAI OSC messages', () => {
	const onCommandState = vi.fn()

	createXtermSession('terminal-a', vi.fn(), vi.fn(), onCommandState)
	terminalMock.lastTerminal.oscHandlers.get(777)('todoai;command-start;bnBtIHRlc3Q=')
	terminalMock.lastTerminal.oscHandlers.get(777)('todoai;command-end')

    expect(onCommandState).toHaveBeenNthCalledWith(1, {
      type: 'command-start',
      command: 'npm test'
    })
	expect(onCommandState).toHaveBeenNthCalledWith(2, {
	  type: 'command-end'
	})
  })

  it('emits command state from legacy TUI Helper OSC messages', () => {
	const onCommandState = vi.fn()

	createXtermSession('terminal-a', vi.fn(), vi.fn(), onCommandState)
	terminalMock.lastTerminal.oscHandlers.get(777)('tui-helper;command-start;Y29kZXg=')

	expect(onCommandState).toHaveBeenCalledWith({
	  type: 'command-start',
	  command: 'codex'
	})
  })

  it('emits terminal title changes from xterm title events', () => {
    const onTitleChange = vi.fn()

    createXtermSession('terminal-a', vi.fn(), vi.fn(), vi.fn(), onTitleChange)
    terminalMock.lastTerminal.titleHandler('! codex')

    expect(onTitleChange).toHaveBeenCalledWith('! codex')
  })
})

function keyEvent(overrides = {}) {
  return {
    key: '',
    ctrlKey: false,
    shiftKey: false,
    metaKey: false,
    altKey: false,
    ...overrides
  }
}
