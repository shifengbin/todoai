export class TerminalSessionManager {
  constructor({
    createSession,
    sendInput,
    resizeTerminal,
    clipboard = {},
    onError = () => {},
    onCommandState = () => {},
    onTitleChange = () => {}
  }) {
    this.createSession = createSession
    this.sendInput = sendInput
    this.resizeTerminal = resizeTerminal
    this.clipboard = clipboard
    this.onError = onError
    this.onCommandState = onCommandState
    this.onTitleChange = onTitleChange
    this.sessions = new Map()
    this.activeTerminalId = null
    this.replayedTerminals = new Set()
    this.replayingTerminals = new Set()
  }

  ensure(terminalId) {
    if (!this.sessions.has(terminalId)) {
      const session = this.createSession(
        terminalId,
        (data) => {
          if (this.replayingTerminals.has(terminalId)) {
            return
          }
          this.sendInput(terminalId, data)
        },
        (action) => {
          this.handleShortcut(terminalId, action)
        },
        (event) => {
          this.onCommandState(terminalId, event)
        },
        (title) => {
          this.onTitleChange(terminalId, title)
        }
      )
      session.opened = false
      this.sessions.set(terminalId, session)
    }
    return this.sessions.get(terminalId)
  }

  activate(terminalId, container) {
    if (!terminalId || !container) {
      return null
    }

    const session = this.ensure(terminalId)
    this.activeTerminalId = terminalId
    if (!session.opened) {
      session.terminal.open(container)
      session.opened = true
    }
    this.fit(terminalId, false)
    return session
  }

  write(terminalId, data) {
    const session = this.sessions.get(terminalId)
    if (session) {
      session.terminal.write(data)
    }
  }

  replayHistory(terminalId, output) {
    if (!output || this.replayedTerminals.has(terminalId)) {
      return
    }
    const session = this.sessions.get(terminalId)
    if (session) {
      this.replayedTerminals.add(terminalId)
      this.replayingTerminals.add(terminalId)
      session.terminal.write(output, () => {
        this.replayingTerminals.delete(terminalId)
      })
    }
  }

  dispose(terminalId) {
    const session = this.sessions.get(terminalId)
    if (!session) {
      return
    }

    session.terminal.dispose?.()
    this.sessions.delete(terminalId)
    this.replayedTerminals.delete(terminalId)
    this.replayingTerminals.delete(terminalId)
    if (this.activeTerminalId === terminalId) {
      this.activeTerminalId = null
    }
  }

  hasSelection(terminalId) {
    const session = this.sessions.get(terminalId)
    return Boolean(session?.terminal.hasSelection?.())
  }

  focus(terminalId) {
    const session = this.sessions.get(terminalId)
    session?.terminal.focus?.()
  }

  async copySelection(terminalId) {
    return this.withClipboardError(async () => {
      const session = this.sessions.get(terminalId)
      if (!session?.terminal.hasSelection?.()) {
        return false
      }

      const selectedText = session.terminal.getSelection?.() || ''
      if (!selectedText) {
        return false
      }

      await this.clipboard.writeText?.(selectedText)
      return true
    })
  }

  async paste(terminalId) {
    return this.withClipboardError(async () => {
      const text = (await this.clipboard.readText?.()) || ''
      if (!text) {
        return false
      }

      this.sendInput(terminalId, text)
      return true
    })
  }

  handleShortcut(terminalId, action) {
    if (action === 'copy') {
      void this.copySelection(terminalId)
    }
    if (action === 'paste') {
      void this.paste(terminalId)
    }
  }

  async withClipboardError(action) {
    try {
      return await action()
    } catch (error) {
      this.onError(error)
      return false
    }
  }

  fitActive() {
    if (this.activeTerminalId) {
      this.fit(this.activeTerminalId, true)
    }
  }

  fit(terminalId, reportResize = true) {
    const session = this.sessions.get(terminalId)
    if (!session) {
      return
    }

    session.fitAddon.fit()
    if (reportResize && this.resizeTerminal) {
      this.resizeTerminal(terminalId, session.terminal.cols, session.terminal.rows)
    }
  }

  size(terminalId = this.activeTerminalId) {
    const session = this.sessions.get(terminalId)
    if (!session) {
      return null
    }
    return {
      cols: session.terminal.cols,
      rows: session.terminal.rows
    }
  }
}
