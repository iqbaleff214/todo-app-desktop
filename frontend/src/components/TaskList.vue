<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useTasksStore } from '@/stores/tasks'
import type { models } from '../../wailsjs/go/models'
import TaskItem from './TaskItem.vue'
import EmptyState from './EmptyState.vue'

const props = defineProps<{
  tasks: models.Task[]
  readonly: boolean
  context?: 'widget' | 'expanded'
}>()

const tasksStore = useTasksStore()

const showDone = ref(true)

const pendingTasks = computed(() =>
  props.tasks.filter((t) => !t.Done).sort((a, b) => a.Position - b.Position),
)

const doneTasks = computed(() =>
  props.tasks.filter((t) => t.Done).sort((a, b) => a.Position - b.Position),
)

// --- Drag-and-drop state ---
const draggingId = ref<string | null>(null)
const dragOverId = ref<string | null>(null)

function onDragStart(e: DragEvent, id: string) {
  if (props.readonly) return
  draggingId.value = id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', id)
  }
}

function onDragOver(e: DragEvent, id: string) {
  if (props.readonly || !draggingId.value || draggingId.value === id) return
  e.preventDefault()
  dragOverId.value = id
}

function onDragLeave(e: DragEvent, id: string) {
  if (dragOverId.value === id) {
    const related = e.relatedTarget as HTMLElement | null
    const row = (e.currentTarget as HTMLElement)
    if (!row.contains(related)) {
      dragOverId.value = null
    }
  }
}

function onDrop(e: DragEvent, targetId: string) {
  e.preventDefault()
  if (!draggingId.value || draggingId.value === targetId) {
    cleanup()
    return
  }

  const ids = pendingTasks.value.map((t) => t.ID)
  const fromIdx = ids.indexOf(draggingId.value)
  const toIdx = ids.indexOf(targetId)

  if (fromIdx === -1 || toIdx === -1) { cleanup(); return }

  ids.splice(fromIdx, 1)
  ids.splice(toIdx, 0, draggingId.value)
  tasksStore.reorderTasks(ids)
  cleanup()
}

function onDragEnd() {
  cleanup()
}

function cleanup() {
  draggingId.value = null
  dragOverId.value = null
}

// --- Arrow-key navigation ---
const listRef = ref<HTMLElement | null>(null)

function handleArrowNav(e: KeyboardEvent) {
  if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return
  const rows = Array.from(
    listRef.value?.querySelectorAll<HTMLElement>('.task-item[tabindex="0"]') ?? [],
  )
  const focused = document.activeElement as HTMLElement
  const idx = rows.indexOf(focused)
  if (idx === -1) {
    if (rows[0]) rows[0].focus()
    return
  }
  e.preventDefault()
  if (e.key === 'ArrowDown' && idx < rows.length - 1) rows[idx + 1].focus()
  if (e.key === 'ArrowUp' && idx > 0) rows[idx - 1].focus()
}

// --- Task event handlers ---
function handleToggle(id: string) {
  tasksStore.toggleDone(id)
}

function handleEdit(id: string, text: string) {
  tasksStore.updateText(id, text)
}

function handleDelete(id: string) {
  tasksStore.deleteTask(id)
}
</script>

<template>
  <div ref="listRef" class="task-list" @keydown="handleArrowNav">
    <!-- Pending tasks -->
    <TransitionGroup name="task" tag="div" class="task-section">
      <template v-if="pendingTasks.length === 0 && doneTasks.length === 0">
        <EmptyState key="empty" :context="context ?? 'widget'" />
      </template>

      <div
        v-for="task in pendingTasks"
        :key="task.ID"
        class="task-row"
        :class="{
          dragging: draggingId === task.ID,
          'drag-over': dragOverId === task.ID,
        }"
        :draggable="!readonly"
        @dragstart="onDragStart($event, task.ID)"
        @dragover="onDragOver($event, task.ID)"
        @dragleave="onDragLeave($event, task.ID)"
        @drop="onDrop($event, task.ID)"
        @dragend="onDragEnd"
      >
        <TaskItem
          :task="task"
          :readonly="readonly"
          @toggle="handleToggle"
          @edit="handleEdit"
          @delete="handleDelete"
        />
      </div>
    </TransitionGroup>

    <!-- Done tasks -->
    <div v-if="doneTasks.length > 0" class="done-section">
      <button class="done-header" @click="showDone = !showDone">
        <span class="done-label">Completed ({{ doneTasks.length }})</span>
        <span class="done-chevron" :class="{ open: showDone }">▾</span>
      </button>

      <TransitionGroup v-if="showDone" name="task" tag="div" class="task-section">
        <div v-for="task in doneTasks" :key="task.ID" class="task-row">
          <TaskItem
            :task="task"
            :readonly="readonly"
            @toggle="handleToggle"
            @edit="handleEdit"
            @delete="handleDelete"
          />
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<style scoped>
.task-list {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  outline: none;
}

/* Hide scrollbar visually but keep functionality */
.task-list::-webkit-scrollbar {
  width: 4px;
}
.task-list::-webkit-scrollbar-track {
  background: transparent;
}
.task-list::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 2px;
}

.task-section {
  position: relative;
}

/* Drag visuals */
.task-row.dragging {
  opacity: 0.4;
}

.task-row.drag-over {
  border-top: 2px solid var(--color-accent);
}

/* Done section */
.done-section {
  margin-top: 4px;
}

.done-header {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 100%;
  background: none;
  border: none;
  padding: 6px 12px;
  cursor: pointer;
  text-align: left;
}

.done-label {
  font-size: 11px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.done-chevron {
  font-size: 10px;
  color: var(--color-text-secondary);
  transition: transform 150ms ease;
}

.done-chevron.open {
  transform: rotate(0deg);
}

.done-chevron:not(.open) {
  transform: rotate(-90deg);
}

/* TransitionGroup animations */
.task-enter-active,
.task-leave-active {
  transition: opacity 200ms ease, transform 200ms ease;
}

.task-enter-from {
  opacity: 0;
  transform: translateY(-6px);
}

.task-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.task-leave-active {
  position: absolute;
  width: 100%;
}

.task-move {
  transition: transform 200ms ease;
}
</style>
