<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useTasksStore } from '@/stores/tasks'

const props = withDefaults(
  defineProps<{
    readonly?: boolean
  }>(),
  { readonly: false },
)

const tasksStore = useTasksStore()
const text = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

const showCount = computed(() => text.value.length > 400)
const isDisabled = computed(() => props.readonly || tasksStore.loading)

function focus() {
  inputRef.value?.focus()
}
defineExpose({ focus })

async function submit() {
  const trimmed = text.value.trim()
  if (!trimmed || isDisabled.value) return
  await tasksStore.addTask(trimmed)
  text.value = ''
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    submit()
  } else if (e.key === 'Escape') {
    text.value = ''
    inputRef.value?.blur()
  }
}
</script>

<template>
  <div class="task-input-wrap">
    <input
      ref="inputRef"
      v-model="text"
      class="task-input"
      :class="{ loading: tasksStore.loading }"
      type="text"
      placeholder="Add a task…"
      maxlength="500"
      :disabled="isDisabled"
      @keydown="handleKeydown"
    />
    <span v-if="showCount" class="char-count">{{ text.length }}/500</span>
  </div>
</template>

<style scoped>
.task-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.task-input {
  width: 100%;
  background: transparent;
  border: none;
  border-top: 1px solid var(--color-border);
  outline: none;
  font-size: 13px;
  font-family: inherit;
  color: var(--color-text-primary);
  padding: 10px 12px;
  transition: border-color 150ms ease;
}

.task-input::placeholder {
  color: var(--color-text-secondary);
  opacity: 0.7;
}

.task-input:focus {
  border-top-color: var(--color-accent);
}

.task-input:disabled {
  opacity: 0.5;
}

.task-input.loading {
  cursor: wait;
}

.char-count {
  position: absolute;
  right: 10px;
  font-size: 10px;
  color: var(--color-text-secondary);
  pointer-events: none;
}
</style>
