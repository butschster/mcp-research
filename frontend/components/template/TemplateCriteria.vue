<template>
  <dl :class="['criteria', { 'criteria--dense': dense }]">
    <div v-if="whenToUse" class="criterion">
      <dt class="criterion-label">Use when</dt>
      <dd class="criterion-text">{{ whenToUse }}</dd>
    </div>
    <div v-else class="criterion">
      <dt class="criterion-label">Use when</dt>
      <dd class="criterion-text criterion-missing">
        — no matching criteria, so an agent will not know when to choose this —
      </dd>
    </div>

    <div v-if="whenNotToUse" class="criterion">
      <dt class="criterion-label">Not when</dt>
      <dd class="criterion-text">{{ whenNotToUse }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
/*
 * The pair of matching criteria, in one place.
 *
 * This is the most important formatting on the surface: the two lines are what
 * an agent chooses on, and what a reader compares across methodologies. The row
 * and the detail page both use it so the comparison looks the same in both.
 *
 * The criteria themselves are --color-text, not muted. `.data-row` has a hover
 * state, and muted text at --type-xs measures 4.36:1 against the hovered
 * surface — so every hovered row would fail contrast. The labels take the
 * demotion instead; they are two fixed words a reader learns once.
 */
withDefaults(defineProps<{
  whenToUse?: string
  whenNotToUse?: string
  /** Tighter spacing, for a list row rather than a page. */
  dense?: boolean
}>(), { whenToUse: '', whenNotToUse: '', dense: false })
</script>

<style scoped>
.criteria {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.criteria--dense { gap: var(--space-2); margin-top: var(--space-2); }

.criterion {
  display: grid;
  grid-template-columns: 5.5rem minmax(0, 1fr);
  gap: var(--space-3);
  align-items: baseline;
}

.criterion-label {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.criterion-text {
  font-size: var(--type-sm);
  color: var(--color-text);
  /* A criterion can be 240 characters of prose with a long product name in it. */
  overflow-wrap: anywhere;
}
.criteria--dense .criterion-text { font-size: var(--type-xs); }

.criterion-missing { color: var(--color-warning); }

@media (max-width: 768px) {
  .criterion { grid-template-columns: 1fr; gap: var(--space-1); }
}
</style>
