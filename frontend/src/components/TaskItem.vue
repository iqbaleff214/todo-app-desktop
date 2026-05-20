<script lang="ts" setup>
import { ref, nextTick } from 'vue'
import type { models } from '../../wailsjs/go/models'

const props = defineProps<{
  task: models.Task
  readonly: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle', id: string): void
  (e: 'edit', id: string, text: string): void
  (e: 'delete', id: string): void
}>()

const editing = ref(false)
const editText = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

async function startEdit() {
  if (props.readonly || props.task.Done) return
  editing.value = true
  editText.value = props.task.Text
  await nextTick()
  inputRef.value?.focus()
  inputRef.value?.select()
}

function commitEdit() {
  if (!editing.value) return
  editing.value = false
  const trimmed = editText.value.trim()
  if (trimmed && trimmed !== props.task.Text) {
    emit('edit', props.task.ID, trimmed)
  }
}

function cancelEdit() {
  editing.value = false
}

function handleRowKeydown(e: KeyboardEvent) {
  if (editing.value) return
  if (e.key === ' ') {
    e.preventDefault()
    if (!props.readonly) emit('toggle', props.task.ID)
  } else if (e.key === 'Delete' || e.key === 'Backspace') {
    e.preventDefault()
    if (!props.readonly) emit('delete', props.task.ID)
  }
}

function handleEditKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    commitEdit()
  } else if (e.key === 'Escape') {
    e.preventDefault()
    cancelEdit()
  }
}
</script>

<template>
  <div
    class="task-item"
    :class="{ done: task.Done, editing }"
    tabindex="0"
    @keydown="handleRowKeydown"
  >
    <!-- Checkbox -->
    <button
      class="checkbox"
      :class="{ checked: task.Done }"
      :disabled="readonly"
      tabindex="-1"
      @click.stop="emit('toggle', task.ID)"
      aria-label="Toggle task done"
    >
      <svg v-if="task.Done" width="10" height="8" viewBox="0 0 10 8" fill="none">
        <path d="M1 4L3.5 6.5L9 1" stroke="white" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </button>

    <!-- Text / Inline edit -->
    <div class="task-text-wrap" @click="startEdit">
      <input
        v-if="editing"
        ref="inputRef"
        v-model="editText"
        class="task-edit-input"
        maxlength="500"
        @blur="commitEdit"
        @keydown="handleEditKeydown"
        @click.stop
      />
      <span v-else class="task-text">{{ task.Text }}</span>
    </div>

    <!-- Delete button (hover only) -->
    <button
      v-if="!readonly"
      class="delete-btn"
      tabindex="-1"
      @click.stop="emit('delete', task.ID)"
      aria-label="Delete task"
    >
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
        <path d="M1 1L11 11M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
      </svg>
    </button>
  </div>
</template>

<style scoped>
.task-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: default;
  outline: none;
  position: relative;
  transition: background-color 150ms ease;
}

.task-item:focus-visible {
  background-color: var(--color-surface);
}

.task-item:hover {
  background-color: var(--color-surface);
}

/* Checkbox */
.checkbox {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 2px solid var(--color-border);
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: background-color 150ms ease, border-color 150ms ease, transform 150ms ease;
}

.checkbox:not(:disabled):hover {
  border-color: var(--color-accent);
  transform: scale(1.05);
}

.checkbox.checked {
  background-color: var(--color-accent);
  border-color: var(--color-accent);
  transform: scale(1);
}

.checkbox:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Task text */
.task-text-wrap {
  flex: 1;
  min-width: 0;
  cursor: text;
}

.task-text {
  display: block;
  font-size: 13px;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.task-item.done .task-text {
  text-decoration: line-through;
  color: var(--color-done);
}

.task-edit-input {
  width: 100%;
  font-size: 13px;
  font-family: inherit;
  color: var(--color-text-primary);
  background: transparent;
  border: none;
  outline: none;
  padding: 0;
  line-height: 1.4;
}

/* Delete button */
.delete-btn {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  opacity: 0;
  transition: opacity 150ms ease, color 150ms ease;
}

.task-item:hover .delete-btn,
.task-item:focus-visible .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #ef4444;
}
</style>
