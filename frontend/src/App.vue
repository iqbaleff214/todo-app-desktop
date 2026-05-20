<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useTasksStore } from '@/stores/tasks'
import { useKeyboard } from '@/composables/useKeyboard'
import Widget from '@/components/Widget.vue'
import * as App from '../wailsjs/go/main/App'

const settingsStore = useSettingsStore()
const tasksStore = useTasksStore()
const widgetRef = ref<InstanceType<typeof Widget> | null>(null)

const showSettings = ref(false)

onMounted(async () => {
  await settingsStore.load()
  await tasksStore.loadToday()
  await tasksStore.loadDatesWithTasks()
})

useKeyboard({
  onAddTask: () => widgetRef.value?.focusInput(),
  onToggleExpand: () => App.ToggleExpanded(),
  onOpenSettings: () => { showSettings.value = true },
  onToday: () => tasksStore.loadToday(),
  onPrevDate: () => { /* Phase 6 */ },
  onNextDate: () => { /* Phase 6 */ },
  onQuit: () => { /* Wails handles Cmd+Q natively */ },
})
</script>

<template>
  <div id="widget-root">
    <Widget ref="widgetRef" @open-settings="showSettings = true" />
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
