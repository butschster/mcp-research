<template>
  <div class="tabs content-tabs" role="tablist" :aria-label="label">
    <button
      v-for="tab in tabs"
      :id="`tab-${tab.id}`"
      :key="tab.id"
      type="button"
      role="tab"
      :aria-selected="modelValue === tab.id"
      :aria-controls="`panel-${tab.id}`"
      :tabindex="modelValue === tab.id ? 0 : -1"
      :class="['tab', 'content-tab', { active: modelValue === tab.id }]"
      @click="select(tab.id)"
      @keydown="onKey($event, tab.id)"
    >
      {{ tab.label }}
      <!-- The visible badge reads as two numbers to a screen reader — "Skills
           4 6" — so it is hidden and the button carries the sentence instead. -->
      <span v-if="tab.count !== undefined" class="tab-count" aria-hidden="true">{{ tab.count }}</span>
      <span v-if="tab.srCount" class="sr-only">, {{ tab.srCount }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
/*
 * The ARIA tab strip, in one place.
 *
 * This markup, its roving tabindex and its arrow-key handler existed inline on
 * the session page, with thirty lines of scoped CSS beside them. The settings
 * page needs the same thing, and a second hand-written copy is how the two tab
 * strips already in this product drifted apart — one had `role="tab"`, the
 * other shipped as bare buttons.
 */
export interface TabSpec {
  id: string
  label: string
  /** Shown as a badge. Hidden from screen readers; give `srCount` the sentence. */
  count?: string | number
  /** What the badge means, read aloud: "4 of 6 attached", "23 notes". */
  srCount?: string
}

const props = defineProps<{
  tabs: TabSpec[]
  modelValue: string
  label: string
}>()

const emit = defineEmits<{ 'update:modelValue': [string] }>()

function select(id: string) {
  if (id !== props.modelValue) emit('update:modelValue', id)
}

/* Selection follows focus, and focus moves to the tab that was selected —
   otherwise the arrow keys change the panel while the keyboard stays behind. */
function onKey(e: KeyboardEvent, current: string) {
  const ids = props.tabs.map(t => t.id)
  const i = ids.indexOf(current)
  let next: string | undefined
  if (e.key === 'ArrowRight') next = ids[(i + 1) % ids.length]
  else if (e.key === 'ArrowLeft') next = ids[(i - 1 + ids.length) % ids.length]
  else if (e.key === 'Home') next = ids[0]
  else if (e.key === 'End') next = ids[ids.length - 1]
  if (!next) return
  e.preventDefault()
  select(next)
  const target = next
  nextTick(() => document.getElementById(`tab-${target}`)?.focus())
}
</script>

<style scoped>
.content-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border);
  /* Four labels and their badges do not fit a 375px screen. Scroll the strip
     rather than letting the page body scroll sideways. */
  overflow-x: auto;
  scrollbar-width: none;
}
.content-tabs::-webkit-scrollbar { display: none; }

.content-tab {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  font-family: inherit;
  cursor: pointer;
  transition: color var(--transition-fast), border-color var(--transition-fast);
  margin-bottom: -1px;
}
.content-tab:hover { color: var(--color-text); }
.content-tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

.tab-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: var(--radius-xs);
  padding: 0.15rem 0.4rem;
  font-variant-numeric: tabular-nums;
  min-width: 1.6rem;
  text-align: center;
  white-space: nowrap;
}
.content-tab.active .tab-count { background: var(--color-primary-muted); color: var(--color-primary); }

@media (max-width: 768px) {
  .content-tab { padding: var(--space-2) var(--space-3); font-size: var(--type-xs); }
}
</style>
