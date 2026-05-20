import { reactive, watch } from 'vue'
import { defineStore } from 'pinia'
import * as App from '../../wailsjs/go/main/App'
import type { models } from '../../wailsjs/go/models'

type Settings = models.Settings

const defaults: Settings = {
  theme: 'system',
  opacity: 0.9,
  autoHide: false,
  launchOnLogin: false,
  allowEditPast: false,
  windowX: -1,
  windowY: -1,
  widgetWidth: 280,
  widgetHeight: 420,
}

function applyTheme(theme: string) {
  const root = document.documentElement
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.dataset.theme = prefersDark ? 'dark' : 'light'
  } else {
    root.dataset.theme = theme
  }
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = reactive<Settings>({ ...defaults })

  // Apply theme whenever it changes; immediate so the first render is correct.
  watch(
    () => settings.theme,
    (theme) => applyTheme(theme),
    { immediate: true },
  )

  // Live-preview opacity via CSS variable.
  watch(
    () => settings.opacity,
    (val) => {
      document.documentElement.style.setProperty('--widget-opacity', String(val))
    },
    { immediate: true },
  )

  // Keep theme in sync when the OS preference changes (only relevant when
  // theme === 'system').
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', () => {
    if (settings.theme === 'system') applyTheme('system')
  })

  async function load(): Promise<void> {
    const loaded = await App.GetSettings()
    Object.assign(settings, loaded)
  }

  async function save(patch: Partial<Settings>): Promise<void> {
    const updated: Settings = { ...settings, ...patch }
    await App.SaveSettings(updated)
    Object.assign(settings, patch)
  }

  async function resetWindowPosition(): Promise<void> {
    await App.ResetWindowPosition()
    settings.windowX = -1
    settings.windowY = -1
  }

  return { settings, load, save, resetWindowPosition }
})
