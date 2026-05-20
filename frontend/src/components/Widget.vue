<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useTasksStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import TaskList from './TaskList.vue'
import TaskInput from './TaskInput.vue'
import CountBadge from './CountBadge.vue'
import * as App from '../../wailsjs/go/main/App'

defineEmits<{
  (e: 'openSettings'): void
}>()

const tasksStore = useTasksStore()
const settingsStore = useSettingsStore()

const idle = ref(false)
const taskInputRef = ref<InstanceType<typeof TaskInput> | null>(null)
let idleTimer: ReturnType<typeof setTimeout> | null = null

const widgetOpacity = computed(() => {
  if (!settingsStore.settings.autoHide) return 1
  return idle.value ? settingsStore.settings.opacity : 1
})

function onMouseEnter() {
  idle.value = false
  if (idleTimer !== null) {
    clearTimeout(idleTimer)
    idleTimer = null
  }
}

function onMouseLeave() {
  if (!settingsStore.settings.autoHide) return
  idleTimer = setTimeout(() => {
    idle.value = true
  }, 2000)
}

function focusInput() {
  taskInputRef.value?.focus()
}

defineExpose({ focusInput })
</script>

<template>
  <div
    class="widget"
    :style="{ opacity: widgetOpacity }"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
  >
    <!-- Title bar — full width drag handle -->
    <div
      class="title-bar"
      style="-webkit-app-region: drag"
      @dblclick="App.ToggleExpanded()"
    >
      <span class="app-name">todo</span>
      <CountBadge :done="tasksStore.doneCount" :total="tasksStore.totalCount" />
      <div class="title-actions" style="-webkit-app-region: no-drag">
        <button
          class="icon-btn"
          title="Expand (Ctrl+E)"
          @click="App.ToggleExpanded()"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M9 1h4v4M5 13H1V9M13 1L8 6M1 13l5-5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <button
          class="icon-btn"
          title="Settings (Ctrl+,)"
          @click="$emit('openSettings')"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="2" stroke="currentColor" stroke-width="1.5"/>
            <path d="M7 1v1.5M7 11.5V13M1 7h1.5M11.5 7H13M2.93 2.93l1.06 1.06M10.01 10.01l1.06 1.06M2.93 11.07l1.06-1.06M10.01 3.99l1.06-1.06" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Scrollable task area -->
    <div class="task-area">
      <TaskList :tasks="tasksStore.tasks" :readonly="false" context="widget" />
    </div>

    <!-- Always-visible add input -->
    <TaskInput ref="taskInputRef" />
  </div>
</template>

<style scoped>
.widget {
  width: 280px;
  height: 420px;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
  border-radius: var(--radius-widget);
  box-shadow: var(--shadow-widget);
  overflow: hidden;
  transition: opacity 400ms ease;
}

/* Title bar */
.title-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border);
  cursor: grab;
}

.app-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
  flex-shrink: 0;
}

/* Push badge to center, actions to right */
.title-bar .count-badge {
  flex: 1;
  text-align: center;
}

.title-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.icon-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  padding: 0;
  transition: color 150ms ease, background-color 150ms ease;
}

.icon-btn:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface);
}

/* Task area fills remaining space, scroll handled inside TaskList */
.task-area {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
