<template>
  <span :class="['badge', variant]" :title="explanation">{{ text }}</span>
</template>

<script setup lang="ts">
/*
 * Where a methodology comes from.
 *
 * A deliberate near-copy of SkillOriginBadge with a different vocabulary: a
 * template's tier is only ownership. Two templates never apply at once — one is
 * chosen at kickoff and then it is done — so there is no precedence to explain
 * here, which is the whole reason it is not the same component.
 *
 * The rule both keep: the tier is text, never colour alone.
 */
const props = defineProps<{
  tier: 'global' | 'team' | string
  /** Slug of the shipped template this one copies, when it is a fork. */
  forkedFrom?: string
  /** Name of the owning team, when it is known. */
  teamName?: string
}>()

const text = computed(() => {
  if (props.tier !== 'team') return 'Ships with the app'
  return props.forkedFrom ? 'Your team · edited copy' : 'Your team'
})

const variant = computed(() => (props.tier === 'team' ? 'badge-completed' : 'badge-archived'))

const explanation = computed(() => {
  if (props.tier !== 'team') {
    return 'Written by us, shipped in the binary and updated with it. Editing it makes a copy for your team; the original stays as it is for everybody else.'
  }
  const owner = props.teamName ? `${props.teamName}'s` : "your team's"
  return props.forkedFrom
    ? `${owner} copy of the shipped "${props.forkedFrom}", edited. It replaces the original everywhere this team looks.`
    : `Written by ${owner === "your team's" ? 'your team' : props.teamName}, and theirs alone. No other team can see it.`
})
</script>
