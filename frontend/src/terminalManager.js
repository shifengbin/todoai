export class TerminalSessionManager {
  constructor({ createSession, sendInput, resizeTerminal, clipboard = {}, onError = () => {} }) {
    this.createSession = createSession
    this.sendInput = sendInput
    this.resizeTerminal = resizeTerminal
    this.clipboard = clipboard
    this.onError = onError
    this.sessions = new Map()
    this.activeProjectId = null
  }

  ensure(projectId) {
    if (!this.sessions.has(projectId)) {
      const session = this.createSession(
        projectId,
        (data) => {
          this.sendInput(projectId, data)
        },
        (action) => {
          this.handleShortcut(projectId, action)
        }
      )
      session.opened = false
      this.sessions.set(projectId, session)
    }
    return this.sessions.get(projectId)
  }

  activate(projectId, container) {
    if (!projectId || !container) {
      return null
    }

    const session = this.ensure(projectId)
    this.activeProjectId = projectId
    if (!session.opened) {
      session.terminal.open(container)
      session.opened = true
    }
    this.fit(projectId, false)
    return session
  }

  write(projectId, data) {
    const session = this.sessions.get(projectId)
    if (session) {
      session.terminal.write(data)
    }
  }

  hasSelection(projectId) {
    const session = this.sessions.get(projectId)
    return Boolean(session?.terminal.hasSelection?.())
  }

  async copySelection(projectId) {
    return this.withClipboardError(async () => {
      const session = this.sessions.get(projectId)
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

  async paste(projectId) {
    return this.withClipboardError(async () => {
      const text = (await this.clipboard.readText?.()) || ''
      if (!text) {
        return false
      }

      this.sendInput(projectId, text)
      return true
    })
  }

  handleShortcut(projectId, action) {
    if (action === 'copy') {
      void this.copySelection(projectId)
    }
    if (action === 'paste') {
      void this.paste(projectId)
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
    if (this.activeProjectId) {
      this.fit(this.activeProjectId, true)
    }
  }

  fit(projectId, reportResize = true) {
    const session = this.sessions.get(projectId)
    if (!session) {
      return
    }

    session.fitAddon.fit()
    if (reportResize && this.resizeTerminal) {
      this.resizeTerminal(projectId, session.terminal.cols, session.terminal.rows)
    }
  }
}
