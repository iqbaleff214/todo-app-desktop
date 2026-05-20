<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useTasksStore } from '@/stores/tasks'
import { useKeyboard } from '@/composables/useKeyboard'
import { prevDate, nextDate } from '@/composables/useDate'
import Widget from '@/components/Widget.vue'
import ExpandedView from '@/components/ExpandedView.vue'
import { EventsOn } from '../wailsjs/runtime/runtime'
import * as App from '../wailsjs/go/main/App'

const settingsStore = useSettingsStore()
const tasksStore = useTasksStore()
const widgetRef = ref<InstanceType<typeof Widget> | null>(null)
const showSettings = ref(false)
const isExpanded = ref(false)

let offModeEvent: (() => void) | null = null

onMounted(async () => {
  await settingsStore.load()
  await tasksStore.loadToday()
  await tasksStore.loadDatesWithTasks()
  await tasksStore.loadTaskCountsByDate()

  // Sync initial expanded state from Go.
  isExpanded.value = await App.IsExpanded()

  // Listen for subsequent mode changes (e.g. tray-triggered).
  offModeEvent = EventsOn('window:mode', (expanded: boolean) => {
    isExpanded.value = expanded
  })
})

onUnmounted(() => {
  offModeEvent?.()
})

useKeyboard({
  onAddTask: () => widgetRef.value?.focusInput(),
  onToggleExpand: () => App.ToggleExpanded(),
  onOpenSettings: () => { showSettings.value = true },
  onToday: () => tasksStore.loadToday(),
  onPrevDate: () => tasksStore.loadDate(prevDate(tasksStore.selectedDate)),
  onNextDate: () => tasksStore.loadDate(nextDate(tasksStore.selectedDate)),
  onQuit: () => { /* Wails handles Cmd+Q natively */ },
})
</script>

<template>
  <div id="widget-root">
    <ExpandedView v-if="isExpanded" @open-settings="showSettings = true" />
    <Widget v-else ref="widgetRef" @open-settings="showSettings = true" />
  </div>
</template>

<style>
#widget-root {
  width: 100%;
  height: 100%;
  background: var(--color-bg);
  border-radius: var(--radius-widget);
  overflow: hidden;
}
</style>
