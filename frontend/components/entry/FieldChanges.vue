<template>
  <div v-if="fields.length" class="fields">
    <h4 class="fields-title">Properties</h4>
    <dl class="fields-list">
      <template v-for="field in fields" :key="field.field">
        <dt class="field-name">{{ field.field }}</dt>
        <dd class="field-change">
          <span class="sr-only">changed from {{ field.before || 'empty' }} to {{ field.after || 'empty' }}</span>
          <span class="field-side field-before" aria-hidden="true">
            <StatusBadge v-if="field.field === 'status' && field.before" :status="field.before" />
            <TagList v-else-if="field.field === 'tags' && field.before" :tags="splitTags(field.before)" />
            <span v-else-if="field.before" class="field-text">{{ field.before }}</span>
            <span v-else class="field-empty">—</span>
          </span>
          <span class="field-arrow" aria-hidden="true">→</span>
          <span class="field-side field-after" aria-hidden="true">
            <StatusBadge v-if="field.field === 'status' && field.after" :status="field.after" />
            <TagList v-else-if="field.field === 'tags' && field.after" :tags="splitTags(field.after)" />
            <span v-else-if="field.after" class="field-text">{{ field.after }}</span>
            <span v-else class="field-empty">—</span>
          </span>
        </dd>
      </template>
    </dl>
  </div>
</template>

<script setup lang="ts">
/**
 * What a revision changed outside the document body — a status flip, a retitle,
 * an edited tag list.
 *
 * This exists because a revision can be real and its content diff empty: every
 * status change made while triaging produces one. The panel used to render
 * those as "Nothing changed", which reads as a broken diff rather than as an
 * accurate account of the write.
 *
 * The arrow is decorative; each row carries the before→after in words for a
 * screen reader, because the visual form is two adjacent values whose order is
 * the only thing saying which is which.
 */
defineProps<{
  fields: { field: string; before: string; after: string }[]
}>()

// The API joins tags with ', ' — see changedFields in revision_service.go.
function splitTags(value: string): string[] {
  return value.split(',').map((t) => t.trim()).filter(Boolean)
}
</script>

<style scoped>
.fields {
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}
.fields-title {
  margin: 0 0 var(--space-3);
  font-size: var(--type-xs);
  font-weight: var(--weight-medium);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.fields-list {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: var(--space-2) var(--space-4);
  margin: 0;
  align-items: baseline;
}
.field-name {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.field-change {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin: 0;
  font-size: var(--type-sm);
  min-width: 0;
}
.field-side {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  min-width: 0;
}
.field-before { color: var(--color-text-muted); }
.field-after { color: var(--color-text); }
.field-text { overflow-wrap: anywhere; }
.field-empty { color: var(--color-text-muted); }
.field-arrow {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .fields-list {
    grid-template-columns: 1fr;
    gap: var(--space-1) 0;
  }
  .field-name { margin-top: var(--space-2); }
}
</style>
