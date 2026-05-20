<script lang="ts" setup>
withDefaults(
  defineProps<{
    context?: 'widget' | 'expanded'
  }>(),
  { context: 'widget' },
)
</script>

<template>
  <div class="empty-state">
    <template v-if="context === 'widget'">
      <svg width="56" height="56" viewBox="0 0 56 56" fill="none" aria-hidden="true">
        <rect x="10" y="14" width="36" height="38" rx="4" stroke="var(--color-border)" stroke-width="2"/>
        <rect x="21" y="8" width="14" height="10" rx="3" fill="var(--color-surface)" stroke="var(--color-border)" stroke-width="2"/>
        <line x1="17" y1="27" x2="39" y2="27" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
        <line x1="17" y1="34" x2="39" y2="34" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
        <line x1="17" y1="41" x2="31" y2="41" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
      </svg>
      <p class="empty-title">Nothing for today.</p>
      <p class="empty-hint">Press N to add a task.</p>
    </template>
    <template v-else>
      <p class="empty-title">No tasks on this day.</p>
    </template>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  gap: 8px;
}

.empty-title {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.empty-hint {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-secondary);
  opacity: 0.7;
}
</style>
