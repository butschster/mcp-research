<template>
  <div>
    <h3 v-if="heading" class="card-section-title">
      {{ heading }}
      <span v-if="note !== undefined" class="heading-note">{{ note }}</span>
    </h3>
    <p v-if="blurb" class="group-blurb">{{ blurb }}</p>

    <div v-if="templates.length" class="data-rows">
      <div v-for="tp in templates" :key="tp.id" class="data-row template-row">
        <div class="row-head">
          <NuxtLink :to="`/templates/${tp.id}`" class="template-name">{{ tp.name }}</NuxtLink>
          <TemplateOriginBadge :tier="tp.tier" :forked-from="tp.forked_from" :team-name="teamName?.(tp.team_id)" />
        </div>

        <TemplateCriteria :when-to-use="tp.when_to_use" :when-not-to-use="tp.when_not_to_use" dense />

        <p class="row-meta">
          <span>{{ tp.body_words }} words</span>
          <span v-if="tp.skills?.length"> · attaches {{ tp.skills.length }} skill{{ tp.skills.length === 1 ? '' : 's' }}</span>
          <span v-if="tp.research_count"> · started {{ tp.research_count }} research{{ tp.research_count === 1 ? '' : 'es' }}</span>
        </p>
      </div>
    </div>

    <p v-else class="group-empty">{{ emptyText }}</p>
  </div>
</template>

<script setup lang="ts">
/*
 * A list of methodologies, deliberately not SkillRowList.
 *
 * The two lists are built on the same `.data-row` skeleton and share their type
 * and colour rhythm, so they look like one product — but nothing about the
 * payload matches. A skill row shows one trigger line and carries Read, Attach
 * and Detach; a template row shows a *pair* of criteria and carries no buttons
 * at all, because there is nothing to do to a template from a list. Serving both
 * from one component would mean a `kind` prop and a branch in every block.
 */
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
.card-section-title {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}
.heading-note {
  font-size: var(--type-xs);
  font-weight: var(--weight-normal);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
.group-blurb {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-bottom: var(--space-3);
}
.group-empty {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  padding: var(--space-4) 0;
}

.template-row { display: block; }

.row-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.template-name {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  color: var(--color-text);
  overflow-wrap: anywhere;
}
.template-name:hover { color: var(--color-primary); }

.row-meta {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  margin-top: var(--space-2);
  font-variant-numeric: tabular-nums;
}
</style>
