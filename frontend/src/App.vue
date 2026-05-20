<script lang="ts" setup>
import { onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useTasksStore } from '@/stores/tasks'
import { useKeyboard } from '@/composables/useKeyboard'
import * as App from '../wailsjs/go/main/App'

const settingsStore = useSettingsStore()
const tasksStore = useTasksStore()

onMounted(async () => {
  await settingsStore.load()
  await tasksStore.loadToday()
  await tasksStore.loadDatesWithTasks()
})

useKeyboard({
  onToggleExpand: () => App.ToggleExpanded(),
  onToday: () => tasksStore.loadToday(),
  onPrevDate: () => {
    // Phase 6: implemented in DateSidebar
  },
  onNextDate: () => {
    // Phase 6: implemented in DateSidebar
  },
  onQuit: () => {
    // Wails runtime handles Cmd+Q natively; this is a fallback.
  },
})
</script>

<template>
  <div id="widget-root">
    <!-- Phase 5: Widget.vue and expanded view rendered here -->
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
