<template>
  <span
    :class="['entry-update-badge', `entry-update-badge--${kind}`]"
    :aria-label="accessibleLabel"
  >
    <span class="entry-update-dot" aria-hidden="true"></span>
    {{ label }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  kind: 'new' | 'changed'
  unseenRevisions?: number
}>()

const count = computed(() => Math.max(1, props.unseenRevisions ?? 1))
const label = computed(() => props.kind === 'new'
  ? 'New'
  : count.value > 1 ? `Changed · ${count.value}` : 'Changed')
const accessibleLabel = computed(() => props.kind === 'new'
  ? 'New document'
  : `Changed, ${count.value} unseen ${count.value === 1 ? 'revision' : 'revisions'}`)
</script>

<style scoped>
.entry-update-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  /* The border contributes two pixels; subtract one per side so this aligns
     with the borderless ShortCode chip it commonly sits beside. */
  padding: calc(0.15rem - 1px) 0.45rem;
  border: 1px solid currentColor;
  border-radius: var(--radius-xs);
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  line-height: 1;
  white-space: nowrap;
}
.entry-update-badge--new {
  color: var(--color-primary);
  background: var(--color-primary-muted);
}
.entry-update-badge--changed {
  color: var(--color-warning);
  background: var(--color-surface);
}
.entry-update-dot {
  width: 0.35rem;
  height: 0.35rem;
  border-radius: 50%;
  background: currentColor;
}
</style>
