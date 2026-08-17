<template>
  <div class="data-rows research-rows">
    <NuxtLink
      v-for="research in researches"
      :key="research.id"
      :to="`/research/${research.code || research.id}`"
      class="data-row research-row"
    >
      <ShortCode v-if="research.code" :code="research.code" />
      <span v-else class="code-placeholder"></span>

      <div class="research-identity">
        <span class="research-name">{{ research.name }}</span>
        <span v-if="research.goal" class="research-goal">{{ research.goal }}</span>
      </div>

      <StatusBadge :status="research.status" />
      <span v-if="research.updated_at" class="research-time">{{ relativeTime(research.updated_at) }}</span>
      <span v-else class="research-time"></span>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { relativeTime } from '~/composables/useRelativeTime'

/**
 * The work a team holds, on the team's own page.
 *
 * `ResearchCard` covers the same nouns but is a 400px grid card built for the
 * home page, where a research is the subject and there is room for tags, a
 * team chip and a viewer notice. Here it is a supporting fact about a team, in
 * a page whose other two sections are rule lists — a grid of cards under a
 * heading between two rule lists is three layouts for three lists of things.
 *
 * The team is not named on the row, because the page is the team.
 */
defineProps<{
  researches: {
    id: string
    code?: string
    name: string
    goal?: string
    status: string
    updated_at?: string
  }[]
}>()
</script>

<style scoped>
.research-row {
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  color: inherit;
  text-decoration: none;
  transition: background var(--transition-fast);
}
.research-row:hover { background: var(--color-surface-hover); text-decoration: none; }
.code-placeholder { width: 2.5rem; }
.research-identity { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.research-name { font-weight: var(--weight-medium); overflow-wrap: anywhere; }
/* One line: the goal is a sentence and this is a list, not a reading surface. */
.research-goal {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.research-time { font-size: var(--type-xs); color: var(--color-text-muted); white-space: nowrap; }

@media (max-width: 768px) {
  .research-row { grid-template-columns: auto minmax(0, 1fr) auto; row-gap: var(--space-2); }
  .research-identity { grid-column: 2; }
  .research-time { grid-column: 2 / -1; grid-row: 2; }
}
</style>
