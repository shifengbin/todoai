import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export function createXtermSession(terminalId, onData, onShortcut, onCommandState) {
  const terminal = new Terminal({
    cursorBlink: true,
    fontFamily: '"Cascadia Mono", "JetBrains Mono", "SFMono-Regular", monospace',
    fontSize: 13,
    lineHeight: 1.15,
    scrollback: 4000,
    theme: {
      background: '#111418',
      foreground: '#e6edf3',
      cursor: '#f0c674',
      selectionBackground: '#3c5266',
      black: '#111418',
      red: '#ef5b5b',
      green: '#7bc96f',
      yellow: '#f0c674',
      blue: '#5aa7e8',
      magenta: '#c678dd',
      cyan: '#56b6c2',
      white: '#e6edf3',
      brightBlack: '#5c6773',
      brightRed: '#ff7777',
      brightGreen: '#9be282',
      brightYellow: '#ffd479',
      brightBlue: '#7cc5ff',
      brightMagenta: '#d89df0',
      brightCyan: '#74d5df',
      brightWhite: '#ffffff'
    }
  })
  const fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.onData(onData)
  terminal.parser?.registerOscHandler?.(777, (data) => {
    const event = parseCommandStateOsc(data)
    if (!event) {
      return false
    }
    onCommandState?.(event)
    return true
  })
  terminal.attachCustomKeyEventHandler((event) => {
    if (event.ctrlKey && event.shiftKey && !event.altKey && !event.metaKey) {
      const key = event.key.toLowerCase()
      if (key === 'c') {
        onShortcut?.('copy')
        return false
      }
      if (key === 'v') {
        onShortcut?.('paste')
        return false
      }
    }
    return true
  })
  return { terminal, fitAddon }
}

function parseCommandStateOsc(data) {
  const parts = data.split(';')
  if (parts[0] !== 'tui-helper') {
    return null
  }
  if (parts[1] === 'command-end') {
    return { type: 'command-end' }
  }
  if (parts[1] === 'command-start' && parts[2]) {
    return {
      type: 'command-start',
      command: decodeBase64(parts[2])
    }
  }
  return null
}

function decodeBase64(value) {
  try {
    if (typeof atob === 'function') {
      const binary = atob(value)
      if (typeof TextDecoder === 'function') {
        const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
        return new TextDecoder().decode(bytes)
      }
      return binary
    }
  } catch {
    return ''
  }
  return ''
}
