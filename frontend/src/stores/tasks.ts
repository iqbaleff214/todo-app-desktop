import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import * as App from '../../wailsjs/go/main/App'
import type { models } from '../../wailsjs/go/models'
import { today } from '@/composables/useDate'

function emitToastError(message: string) {
  window.dispatchEvent(new CustomEvent('toast:error', { detail: message }))
}

export const useTasksStore = defineStore('tasks', () => {
  const tasks = ref<models.Task[]>([])
  const selectedDate = ref<string>(today())
  const datesWithTasks = ref<string[]>([])
  const taskCountsByDate = ref<Record<string, number>>({})
  const loading = ref(false)
  const error = ref<string | null>(null)

  // --- Getters ---

  const pendingTasks = computed(() =>
    tasks.value
      .filter((t) => !t.Done)
      .sort((a, b) => a.Position - b.Position),
  )

  const doneTasks = computed(() =>
    tasks.value
      .filter((t) => t.Done)
      .sort((a, b) => a.Position - b.Position),
  )

  const doneCount = computed(() => doneTasks.value.length)
  const totalCount = computed(() => tasks.value.length)
  const isToday = computed(() => selectedDate.value === today())

  // --- Actions ---

  async function loadDate(date: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      selectedDate.value = date
      tasks.value = await App.GetTasksForDate(date)
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    } finally {
      loading.value = false
    }
  }

  async function loadToday(): Promise<void> {
    return loadDate(today())
  }

  async function loadDatesWithTasks(): Promise<void> {
    try {
      datesWithTasks.value = await App.GetDatesWithTasks()
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    }
  }

  async function loadTaskCountsByDate(): Promise<void> {
    try {
      taskCountsByDate.value = await App.GetTaskCountsByDate()
    } catch {
      // counts are non-critical; sidebar will show without them
    }
  }

  async function addTask(text: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const newTask = await App.AddTask(selectedDate.value, text)
      tasks.value.push(newTask)
      const date = selectedDate.value
      if (!datesWithTasks.value.includes(date)) {
        datesWithTasks.value = [date, ...datesWithTasks.value]
      }
      taskCountsByDate.value = { ...taskCountsByDate.value, [date]: (taskCountsByDate.value[date] ?? 0) + 1 }
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    } finally {
      loading.value = false
    }
  }

  async function toggleDone(id: string): Promise<void> {
    try {
      const updated = await App.ToggleDone(id)
      const idx = tasks.value.findIndex((t) => t.ID === id)
      if (idx !== -1) tasks.value[idx] = updated
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    }
  }

  async function updateText(id: string, text: string): Promise<void> {
    try {
      const updated = await App.UpdateText(id, text)
      const idx = tasks.value.findIndex((t) => t.ID === id)
      if (idx !== -1) tasks.value[idx] = updated
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    }
  }

  async function deleteTask(id: string): Promise<void> {
    try {
      const date = tasks.value.find((t) => t.ID === id)?.Date ?? selectedDate.value
      await App.DeleteTask(id)
      tasks.value = tasks.value.filter((t) => t.ID !== id)
      const newCount = Math.max(0, (taskCountsByDate.value[date] ?? 1) - 1)
      taskCountsByDate.value = { ...taskCountsByDate.value, [date]: newCount }
      if (newCount === 0) {
        datesWithTasks.value = datesWithTasks.value.filter((d) => d !== date)
      }
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    }
  }

  async function reorderTasks(ids: string[]): Promise<void> {
    try {
      await App.ReorderTasks(ids)
      ids.forEach((id, i) => {
        const t = tasks.value.find((t) => t.ID === id)
        if (t) t.Position = i
      })
    } catch (e) {
      const msg = String(e)
      error.value = msg
      emitToastError(msg)
    }
  }

  return {
    // State
    tasks,
    selectedDate,
    datesWithTasks,
    taskCountsByDate,
    loading,
    error,
    // Getters
    pendingTasks,
    doneTasks,
    doneCount,
    totalCount,
    isToday,
    // Actions
    loadDate,
    loadToday,
    loadDatesWithTasks,
    loadTaskCountsByDate,
    addTask,
    toggleDone,
    updateText,
    deleteTask,
    reorderTasks,
  }
})
