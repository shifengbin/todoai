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

      attachCustomKeyEventHandler(handler) {
        this.keyHandler = handler
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
