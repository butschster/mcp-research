<template>
  <span class="tag team-chip" :title="`Team: ${name}`">
    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
    <span class="sr-only">Team: </span>{{ name }}
  </span>
</template>

<script setup lang="ts">
/**
 * Names the team a research belongs to.
 *
 * Deliberately not a `TagList` tag: tags are coloured by a hash of their text,
 * and running teams through that would paint them six arbitrary colours and
 * imply a taxonomy that does not exist. A team is one neutral chip.
 */
defineProps<{ name: string }>()
</script>

<style scoped>
.team-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3em;
  /* Long Cyrillic team names must wrap inside the card rather than push the
     timestamp off the end of the row. */
  overflow-wrap: anywhere;
}
.team-chip svg { flex-shrink: 0; opacity: 0.7; }
</style>
