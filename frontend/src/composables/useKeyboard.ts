import { useEventListener } from '@vueuse/core'

interface KeyboardHandlers {
  onAddTask?: () => void
  onToggleExpand?: () => void
  onOpenSettings?: () => void
  onToday?: () => void
  onPrevDate?: () => void
  onNextDate?: () => void
  onQuit?: () => void
}

/**
 * Registers global keyboard shortcuts for the app.
 * Call once in App.vue.
 */
export function useKeyboard(handlers: KeyboardHandlers): void {
  useEventListener(document, 'keydown', (e: KeyboardEvent) => {
    const meta = e.metaKey || e.ctrlKey
    const target = e.target as HTMLElement
    const inInput =
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.isContentEditable

    // Cmd/Ctrl+Q — quit
    if (meta && e.key === 'q') {
      e.preventDefault()
      handlers.onQuit?.()
      return
    }

    // Cmd/Ctrl+E — toggle expanded
    if (meta && e.key === 'e') {
      e.preventDefault()
      handlers.onToggleExpand?.()
      return
    }

    // Cmd/Ctrl+, — open settings
    if (meta && e.key === ',') {
      e.preventDefault()
      handlers.onOpenSettings?.()
      return
    }

    // Cmd/Ctrl+T — jump to today
    if (meta && e.key === 't') {
      e.preventDefault()
      handlers.onToday?.()
      return
    }

    // Cmd/Ctrl+ArrowLeft — previous date
    if (meta && e.key === 'ArrowLeft') {
      e.preventDefault()
      handlers.onPrevDate?.()
      return
    }

    // Cmd/Ctrl+ArrowRight — next date
    if (meta && e.key === 'ArrowRight') {
      e.preventDefault()
      handlers.onNextDate?.()
      return
    }

    // N — add new task (not when inside an input)
    if (!meta && e.key === 'n' && !inInput) {
      e.preventDefault()
      handlers.onAddTask?.()
      return
    }
  })
}
