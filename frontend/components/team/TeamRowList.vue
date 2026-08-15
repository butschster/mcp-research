<template>
  <div class="team-rows">
    <component
      :is="team.personal ? 'div' : NuxtLink"
      v-for="team in visible"
      :key="team.id"
      :to="team.personal ? undefined : `/teams/${team.id}`"
      class="team-row"
      :class="{ 'team-row-static': team.personal }"
    >
      <span class="team-name">{{ team.name }}</span>
      <span v-if="team.personal" class="tag personal-chip">Personal</span>
      <span class="team-sub">
        <span class="team-meta">{{ team.member_count }} {{ team.member_count === 1 ? 'member' : 'members' }}</span>
        <span class="team-role">{{ ROLE_LABELS[team.role] }}</span>
      </span>
      <span v-if="!team.personal" class="team-action">
        {{ team.role === 'owner' ? 'Manage' : 'Open' }}
        <span aria-hidden="true">→</span>
      </span>
      <span v-else class="team-action team-action-empty"></span>
    </component>
  </div>
</template>

<script setup lang="ts">
import { ROLE_LABELS, type Team } from '~/composables/useTeams'
import { NuxtLink } from '#components'

/**
 * The list of teams, as rules rather than cards.
 *
 * A team is one line of information — a name, a headcount, your role. Boxing
 * each one produces a column of identical rectangles and a scroll for what is
 * really a short list.
 *
 * Rendered by both `/teams` and the Settings summary. One implementation on
 * purpose: two would grow apart the first time either page is touched.
 */
const props = withDefaults(defineProps<{ teams: Team[]; limit?: number }>(), { limit: 0 })

// The personal team goes last: it is the row that says nothing new.
const visible = computed(() => {
  const sorted = [...props.teams].sort((a, b) => Number(a.personal) - Number(b.personal))
  return props.limit > 0 ? sorted.slice(0, props.limit) : sorted
})
</script>

<style scoped>
/* One grid for the list, not one per row: separate grids sized their columns
   independently, so "3 members" in one row sat under "Personal" in the next. */
.team-rows {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto auto;
  border-top: 1px solid var(--color-border);
}
.team-sub { display: contents; }
.team-row {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: subgrid;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-1);
  border-bottom: 1px solid var(--color-border);
  color: inherit;
  text-decoration: none;
  transition: background var(--transition-fast);
}
.team-row:not(.team-row-static):hover { background: var(--color-surface-hover); }
.team-name { font-weight: var(--weight-medium); overflow-wrap: anywhere; }
.team-meta, .team-role { font-size: var(--type-xs); color: var(--color-text-muted); white-space: nowrap; }
.team-action { font-size: var(--type-xs); color: var(--color-primary); white-space: nowrap; }
.team-action-empty { width: 4rem; }
.personal-chip { color: var(--color-text-muted); }

@media (max-width: 768px) {
  /* Two lines rather than a horizontal scroll: the name, then everything that
     qualifies it. Rows go back to sizing themselves here — there are only two
     columns and nothing to align across. */
  .team-rows { display: flex; flex-direction: column; }
  .team-row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: var(--space-1) var(--space-2);
  }
  .team-name { grid-column: 1; }
  .personal-chip { grid-column: 2; justify-self: end; }
  .team-sub { display: flex; gap: var(--space-2); grid-column: 1; grid-row: 2; }
  .team-meta::after { content: ' ·'; }
  .team-action { grid-column: 2; grid-row: 2; justify-self: end; }
}
</style>
