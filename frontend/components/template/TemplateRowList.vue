<template>
  <div>
    <div v-if="heading || blurb" class="list-head">
      <h3 v-if="heading" class="card-section-title">
        {{ heading }}
        <span v-if="note !== undefined" class="heading-note">{{ note }}</span>
      </h3>
      <p v-if="blurb" class="group-blurb">{{ blurb }}</p>
    </div>

    <div v-if="templates.length" class="data-rows">
      <div v-for="tp in templates" :key="tp.id" class="data-row template-row">
        <div class="template-overview">
          <div class="row-head">
            <NuxtLink :to="`/templates/${tp.id}`" class="template-name">{{ tp.name }}</NuxtLink>
            <span v-if="tp.tier === 'team'" class="template-team">
              {{ teamName?.(tp.team_id) || 'Your team' }}{{ tp.forked_from ? ' · edited copy' : '' }}
            </span>
          </div>
          <p class="template-description">{{ tp.description || tp.when_to_use || 'Open this methodology to read its instructions.' }}</p>
        </div>

        <TemplateStartButton :methodology="tp" class="row-start" />

        <details class="methodology-details">
          <summary>When to choose this</summary>
          <TemplateCriteria :when-to-use="tp.when_to_use" :when-not-to-use="tp.when_not_to_use" dense />
        </details>
      </div>
    </div>

    <p v-else class="list-empty">{{ emptyText }}</p>
  </div>
</template>

<script setup lang="ts">
import TemplateCriteria from './TemplateCriteria.vue'
import TemplateStartButton from './TemplateStartButton.vue'

// Compare the short descriptions first; open the criteria or full guide when
// needed. Each row owns its copy feedback and disclosure state.
withDefaults(defineProps<{
  templates: any[]
  heading?: string
  note?: string | number
  blurb?: string
  emptyText?: string
  /** Resolves a team id to a name, for the badge's explanation. */
  teamName?: (id: string) => string | undefined
}>(), { emptyText: 'Nothing here.' })
</script>

<style scoped>
.heading-note {
  font-size: var(--type-xs);
  font-weight: var(--weight-normal);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
.group-blurb {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}

.template-row {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  column-gap: var(--space-6);
  row-gap: var(--space-2);
  padding-block: var(--space-5);
}
.template-overview { min-width: 0; }
.template-description { margin-top: var(--space-1); font-size: var(--type-sm); color: var(--color-text-muted); overflow-wrap: anywhere; }
.template-team { font-size: var(--type-2xs); color: var(--color-text-muted); }
.row-start { justify-self: end; }
.methodology-details { grid-column: 1 / -1; min-width: 0; }
.methodology-details summary { width: fit-content; font-size: var(--type-2xs); color: var(--color-text-muted); cursor: pointer; }
.methodology-details summary:hover { color: var(--color-text); }
.methodology-details[open] summary { margin-bottom: var(--space-3); }

.row-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.template-name {
  font-size: var(--type-base);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  overflow-wrap: anywhere;
}
.template-name:hover { color: var(--color-primary); }

@media (max-width: 600px) {
  .template-row { grid-template-columns: minmax(0, 1fr); row-gap: var(--space-3); }
  .row-start { justify-self: start; grid-row: 3; }
}
</style>
