<template>
  <span :class="['badge', variant]" :title="explanation">{{ text }}</span>
</template>

<script setup lang="ts">
/*
 * Where a skill comes from, said in words.
 *
 * Colour alone would not carry it: the three tiers are also the precedence
 * order, and somebody reading the list needs to know which of two skills wins
 * — not merely that they are different. So the badge is text, and the title
 * says what the tier means rather than repeating its name.
 */
const props = defineProps<{
  tier: 'builtin' | 'team' | 'private' | string
  /** A product skill: on by definition, outside the budget, not detachable. */
  ambient?: boolean
  /** Slug of the built-in this one was forked from, when it is a fork. */
  forkedFrom?: string
  /** Attached by a template rather than chosen by a person. */
  viaTemplate?: boolean
}>()

const text = computed(() => {
  if (props.ambient) return 'Built in'
  if (props.tier === 'private') return 'This research'
  if (props.tier === 'team') return props.forkedFrom ? 'Team · forked' : 'Team'
  return 'Built in'
})

const variant = computed(() => {
  if (props.tier === 'private') return 'badge-active'
  if (props.tier === 'team') return 'badge-completed'
  return 'badge-archived'
})

const explanation = computed(() => {
  if (props.ambient) return 'Ships with the app and is always on. It does not count against the six, and it cannot be switched off.'
  if (props.tier === 'private') return 'Belongs to this research alone. It overrides a team skill and a built-in of the same name.'
  if (props.tier === 'team') {
    return props.forkedFrom
      ? `Your team's copy of the built-in "${props.forkedFrom}". The original is untouched and still updates with the app.`
      : 'Written by your team and reusable across its researches. It overrides a built-in of the same name.'
  }
  return 'Ships with the app and updates with it. Editing it makes a copy for your team.'
})
</script>
