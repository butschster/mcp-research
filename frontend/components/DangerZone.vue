<template>
  <section class="danger-zone" :aria-labelledby="headingId">
    <h2 :id="headingId" class="danger-zone-title">{{ title }}</h2>
    <p v-if="lead" class="danger-zone-lead">{{ lead }}</p>
    <div class="danger-zone-rows">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
/**
 * The box at the bottom of a page holding the things that cannot be undone.
 *
 * It was two classes in the global stylesheet — extracted from the team page,
 * which then went on declaring its own copies of both. Scoped selectors carry a
 * `[data-v]` attribute and win, so one class name meant a red-bordered box in
 * the primitive and a borderless column on the only page using it, and the CSS
 * discipline check counted the page overriding the rule as its consumer.
 *
 * A component cannot be shadowed that way, which is the whole reason it is one.
 */
withDefaults(
  defineProps<{
    title?: string
    /** One sentence about what the whole section is, when the rows need framing. */
    lead?: string
  }>(),
  { title: 'Danger zone' },
)

const headingId = `danger-zone-${useId()}`
</script>

<style scoped>
.danger-zone {
  margin-top: var(--space-12);
  border: 1px solid rgba(239, 107, 107, 0.25);
  border-radius: var(--radius);
  padding: var(--space-4) var(--space-5);
}
.danger-zone-title {
  color: var(--color-danger);
  font-size: var(--type-md);
  font-weight: var(--weight-semibold);
  margin: 0;
}
.danger-zone-lead {
  margin: var(--space-1) 0 0;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.danger-zone-rows { display: flex; flex-direction: column; margin-top: var(--space-2); }
</style>
