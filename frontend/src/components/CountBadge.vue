<script lang="ts" setup>
const props = defineProps<{
  done: number
  total: number
}>()
</script>

<template>
  <span v-if="props.total > 0" class="count-badge" :class="{ 'all-done': props.done === props.total }">
    <template v-if="props.done === props.total">All done ✓</template>
    <template v-else>{{ props.done }} / {{ props.total }} done</template>
  </span>
</template>

<style scoped>
.count-badge {
  font-size: 11px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.all-done {
  color: var(--color-accent);
  font-weight: 600;
}
</style>
