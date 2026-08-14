<template>
  <span class="readonly-badge" :title="explanation">
    Read-only
    <span class="sr-only">. {{ explanation }}</span>
  </span>
</template>

<script setup lang="ts">
/**
 * The one thing a viewer is told, instead of twelve disabled buttons.
 *
 * Edit controls are removed from the page rather than dimmed — a disabled
 * button still says "you could have done this" and still needs a tooltip
 * nobody reads. But absence is undetectable by a screen reader, so the missing
 * controls are explained once, at the top, next to the status badge where the
 * eye already goes.
 */
const props = defineProps<{ teamName?: string }>()

const explanation = computed(() =>
  props.teamName
    ? `You have read-only access to this research through the team ${props.teamName}.`
    : 'You have read-only access to this research.',
)
</script>
