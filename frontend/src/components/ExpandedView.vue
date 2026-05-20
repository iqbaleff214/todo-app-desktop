<script lang="ts" setup>
import { computed } from 'vue'
import { useTasksStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import DateSidebar from './DateSidebar.vue'
import TaskList from './TaskList.vue'
import TaskInput from './TaskInput.vue'
import { isPast } from '@/composables/useDate'
import * as App from '../../wailsjs/go/main/App'

defineEmits<{
  (e: 'openSettings'): void
}>()

const tasksStore = useTasksStore()
const settingsStore = useSettingsStore()

const isReadonly = computed(
  () => isPast(tasksStore.selectedDate) && !settingsStore.settings.allowEditPast,
)
</script>

<template>
  <div class="expanded-view">
    <!-- Header bar — drag handle for the window -->
    <div class="expanded-header" style="-webkit-app-region: drag">
      <span class="app-name">todo</span>
      <div class="header-actions" style="-webkit-app-region: no-drag">
        <button
          class="icon-btn"
          title="Back to widget (Ctrl+E)"
          @click="App.ToggleExpanded()"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M5 9L1 13M1 13h4M1 13v-4M9 5L13 1M13 1H9M13 1v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
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

    <!-- Two-column body -->
    <div class="expanded-body">
      <DateSidebar />

      <div class="main-pane">
        <TaskList
          :tasks="tasksStore.tasks"
          :readonly="isReadonly"
          context="expanded"
        />

        <div
          class="input-wrap"
          :title="isReadonly ? 'Enable editing past dates in Settings' : undefined"
        >
          <TaskInput :readonly="isReadonly" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.expanded-view {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

/* Header */
.expanded-header {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border);
  cursor: grab;
  gap: 8px;
}

.app-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
  flex: 1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 2px;
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

/* Two-column layout */
.expanded-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Right-hand task pane */
.main-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Input wrapper — carries the readonly tooltip */
.input-wrap {
  flex-shrink: 0;
}
</style>
