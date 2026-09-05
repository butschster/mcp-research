<template>
  <section class="danger-zone" :aria-labelledby="headingId">
    <!-- The heading and its sentence are the head of this box, and they carry
         the rule. The rows below then read as a series rather than as a stack
         with an orphan line above it. -->
    <div class="danger-zone-head">
      <h2 :id="headingId" class="danger-zone-title">{{ title }}</h2>
      <p v-if="lead" class="danger-zone-lead">{{ lead }}</p>
    </div>
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
/* Framed the way every other list in the product is: the box gives up its
   horizontal padding and hands it to its rows through --row-inset, so the rule
   under the heading and the rules between rows run to the box's edges instead
   of stopping short of them. The rows are a separate component and read the
   property across that boundary. */
.danger-zone {
  margin-top: var(--space-12);
  border: 1px solid rgba(var(--color-error-rgb), 0.25);
  border-radius: var(--radius);
  padding: 0;
  --row-inset: var(--space-5);
}
.danger-zone-head {
  padding: var(--space-4) var(--row-inset);
  /* The box's own red, not the neutral border: this rule belongs to the frame,
     and a grey line across a red box reads as a seam rather than a division. */
  border-bottom: 1px solid rgba(var(--color-error-rgb), 0.25);
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
.danger-zone-rows { display: flex; flex-direction: column; }

@media (max-width: 768px) {
  .danger-zone { --row-inset: var(--space-4); }
}
</style>
