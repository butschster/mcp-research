<template>
  <div class="data-rows team-rows">
    <NuxtLink
      v-for="team in visible"
      :key="team.id"
      :to="`/teams/${team.id}`"
      class="data-row team-row"
    >
      <span class="user-avatar team-avatar" aria-hidden="true">{{ initial(team) }}</span>

      <div class="team-identity">
        <span class="team-line">
          <span class="team-name">{{ team.name }}</span>
          <span v-if="team.personal" class="tag personal-chip">Personal</span>
        </span>
        <span class="team-sub">
          <span>{{ team.member_count }} {{ team.member_count === 1 ? 'member' : 'members' }}</span>
          <span>{{ team.research_count }} {{ team.research_count === 1 ? 'research' : 'researches' }}</span>
        </span>
      </div>

      <span class="badge badge-quiet team-role">{{ ROLE_LABELS[team.role] }}</span>
      <span class="team-arrow" aria-hidden="true">→</span>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { ROLE_LABELS, type Team } from '~/composables/useTeams'

/**
 * The list of teams, as rules rather than cards.
 *
 * A team is one line of information — a name, a headcount, what work is in it,
 * your role. Boxing each one produces a column of identical rectangles and a
 * scroll for what is really a short list.
 *
 * This is the list a reader meets first, and it was the one list on the teams
 * surface that did not use `.data-rows`/`.data-row`: it kept a private subgrid
 * through the split that named it as a consumer, which is how one page ended up
 * with 56px rows above 72px rows above 48px rows. It shares the frame now, and
 * the row is the same four zones as the member and invite lists — who, what,
 * standing, way in.
 *
 * The research count is new and is the point of the row: a team is a set of
 * people around some work, and the work was the one thing this list never said.
 *
 * Rendered by both `/teams` and the Settings summary. One implementation on
 * purpose: two would grow apart the first time either page is touched.
 */
const props = withDefaults(defineProps<{ teams: Team[]; limit?: number }>(), { limit: 0 })

// The personal team goes last: it is the row that says nothing new. It is a
// link like the others, though — it holds researches, and its page is the one
// place that lists them.
const visible = computed(() => {
  const sorted = [...props.teams].sort((a, b) => Number(a.personal) - Number(b.personal))
  return props.limit > 0 ? sorted.slice(0, props.limit) : sorted
})

function initial(team: Team) {
  return (team.name || '?').trim().charAt(0).toUpperCase()
}
</script>

<style scoped>
.team-row {
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  color: inherit;
  text-decoration: none;
  transition: background var(--transition-fast);
}
.team-row:hover { background: var(--color-surface-hover); text-decoration: none; }
.team-row .user-avatar { --avatar-size: 28px; }
.team-avatar { border-radius: var(--radius-xs); }
.team-identity { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.team-line { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.team-name { font-weight: var(--weight-medium); overflow-wrap: anywhere; }
.team-sub { display: flex; flex-wrap: wrap; gap: var(--space-1); font-size: var(--type-xs); color: var(--color-text-muted); }
.team-sub > * + *::before { content: '· '; }
.personal-chip { color: var(--color-text-muted); }
.badge-quiet { background: var(--color-surface-hover); color: var(--color-text-muted); }
/* A floor rather than a width: three roles of nearly equal length, and rows
   whose badges start at different places read as a broken column. */
.team-role { min-width: 4.5rem; justify-content: center; }
.team-arrow { color: var(--color-primary); font-size: var(--type-sm); }

@media (max-width: 768px) {
  .team-row { grid-template-columns: auto minmax(0, 1fr) auto; row-gap: var(--space-2); }
  .team-identity { grid-column: 2; }
  .team-arrow { grid-column: 3; grid-row: 1; }
  .team-role { grid-column: 2 / -1; grid-row: 2; justify-self: start; min-width: 0; }
}
</style>
