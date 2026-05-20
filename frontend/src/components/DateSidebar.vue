<script lang="ts" setup>
import { computed, ref, watch, nextTick } from 'vue'
import { useTasksStore } from '@/stores/tasks'
import { today, formatDisplay, prevDate, nextDate } from '@/composables/useDate'

const tasksStore = useTasksStore()

const todayStr = today()
const d = new Date()
d.setDate(d.getDate() - 1)
const yesterdayStr = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`

// Past dates: in datesWithTasks, older than yesterday, descending
const pastDates = computed(() =>
  tasksStore.datesWithTasks.filter((date) => date < yesterdayStr),
)

// Future dates: in datesWithTasks, newer than today, ascending
const futureDates = computed(() =>
  [...tasksStore.datesWithTasks]
    .filter((date) => date > todayStr)
    .sort((a, b) => (a < b ? -1 : 1)),
)

function isSelected(date: string): boolean {
  return tasksStore.selectedDate === date
}

function taskCount(date: string): number {
  return tasksStore.taskCountsByDate[date] ?? 0
}

function select(date: string) {
  tasksStore.loadDate(date)
}

// Keyboard navigation support (Cmd/Ctrl+←/→ is wired in App.vue via useKeyboard)
// Expose navigate for use by parent if needed.
function navigatePrev() {
  select(prevDate(tasksStore.selectedDate))
}
function navigateNext() {
  select(nextDate(tasksStore.selectedDate))
}
defineExpose({ navigatePrev, navigateNext })

// Scroll selected item into view whenever selection changes.
const listRef = ref<HTMLElement | null>(null)
watch(
  () => tasksStore.selectedDate,
  async () => {
    await nextTick()
    listRef.value?.querySelector<HTMLElement>('.date-item.active')?.scrollIntoView({
      block: 'nearest',
      behavior: 'smooth',
    })
  },
)
</script>

<template>
  <nav ref="listRef" class="date-sidebar" aria-label="Date navigation">
    <!-- Always-pinned: Today -->
    <button
      class="date-item pinned"
      :class="{ active: isSelected(todayStr) }"
      @click="select(todayStr)"
    >
      <span class="date-label">Today</span>
      <span v-if="taskCount(todayStr) > 0" class="count-chip">{{ taskCount(todayStr) }}</span>
    </button>

    <!-- Always-pinned: Yesterday -->
    <button
      class="date-item pinned"
      :class="{ active: isSelected(yesterdayStr) }"
      @click="select(yesterdayStr)"
    >
      <span class="date-label">Yesterday</span>
      <span v-if="taskCount(yesterdayStr) > 0" class="count-chip">{{ taskCount(yesterdayStr) }}</span>
    </button>

    <div class="date-divider" />

    <!-- Upcoming future dates -->
    <template v-if="futureDates.length > 0">
      <div class="section-label">Upcoming</div>
      <button
        v-for="date in futureDates"
        :key="date"
        class="date-item"
        :class="{ active: isSelected(date) }"
        @click="select(date)"
      >
        <span class="date-label">{{ formatDisplay(date) }}</span>
        <span v-if="taskCount(date) > 0" class="count-chip">{{ taskCount(date) }}</span>
      </button>
      <div class="date-divider" />
    </template>

    <!-- Scrollable past dates (older than yesterday) -->
    <div class="past-dates">
      <button
        v-for="date in pastDates"
        :key="date"
        class="date-item"
        :class="{ active: isSelected(date) }"
        @click="select(date)"
      >
        <span class="date-label">{{ formatDisplay(date) }}</span>
        <span v-if="taskCount(date) > 0" class="count-chip">{{ taskCount(date) }}</span>
      </button>
      <p v-if="pastDates.length === 0" class="empty-past">No past dates</p>
    </div>
  </nav>
</template>

<style scoped>
.date-sidebar {
  width: 220px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border);
  overflow: hidden;
}

.date-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 14px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  border-radius: 0;
  transition: background-color 150ms ease;
}

.date-item:hover {
  background-color: var(--color-surface);
}

.date-item.active {
  background-color: var(--color-accent);
}

.date-item.active .date-label,
.date-item.active .count-chip {
  color: #fff;
}

.date-item.pinned {
  flex-shrink: 0;
}

.date-label {
  font-size: 13px;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.count-chip {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--color-text-secondary);
  background: var(--color-surface);
  border-radius: 10px;
  padding: 1px 6px;
  min-width: 20px;
  text-align: center;
}

.date-item.active .count-chip {
  background: rgba(255, 255, 255, 0.25);
}

.date-divider {
  height: 1px;
  background: var(--color-border);
  margin: 4px 0;
  flex-shrink: 0;
}

.section-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-secondary);
  padding: 6px 14px 2px;
  flex-shrink: 0;
}

/* Scrollable past dates */
.past-dates {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.past-dates::-webkit-scrollbar {
  width: 4px;
}
.past-dates::-webkit-scrollbar-track {
  background: transparent;
}
.past-dates::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 2px;
}

.empty-past {
  font-size: 12px;
  color: var(--color-text-secondary);
  opacity: 0.6;
  padding: 12px 14px;
  margin: 0;
}
</style>
