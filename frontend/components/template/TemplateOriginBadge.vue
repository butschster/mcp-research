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
  tier: 'global' | 'team'
  /**
   * Which kind of global this is. `builtin` ships in the binary; `user` was
   * added on this server by whoever runs it. Both are visible to every team, so
   * the tier alone cannot tell a reader who wrote what they are about to follow
   * — or whether an upgrade will change it under them.
   */
  source?: 'builtin' | 'user'
  /** Slug of the global template this one copies, when it is a fork. */
  forkedFrom?: string
  /** Name of the owning team, when it is known. */
  teamName?: string
}>()

const houseWritten = computed(() => props.tier !== 'team' && props.source === 'user')

const text = computed(() => {
  if (houseWritten.value) return 'Added on this server'
  if (props.tier !== 'team') return 'Ships with the app'
  return props.forkedFrom ? 'Your team · edited copy' : 'Your team'
})

/* Both globals share a variant, and the text is what separates them.
 *
 * The green was wrong twice over. It is the product's status colour — StatusBadge
 * uses it for an active research, and SkillOriginBadge one screen away uses it
 * for the *narrowest* scope, `private` — so borrowing it here made "added on this
 * server" read as a state rather than an origin, and made the loudest badge on
 * the page the group with the fewest and least-vetted items. */
const variant = computed(() => (props.tier === 'team' ? 'badge-completed' : 'badge-archived'))

const explanation = computed(() => {
  if (houseWritten.value) {
    return 'Added on this server by whoever runs it, and visible to every team here. It is not part of the app and an upgrade will not change it.'
  }
  if (props.tier !== 'team') {
    return 'Written by us, shipped in the binary and updated with it. Editing it makes a copy for your team; the original stays as it is for everybody else.'
  }
  const owner = props.teamName ? `${props.teamName}'s` : "your team's"
  return props.forkedFrom
    // Deliberately not "the shipped X". A fork copies whatever is in the global
    // tier, and since the operator can write there too, the parent may never
    // have been in the binary at all — which is the exact fact the source split
    // exists to keep straight. This badge holds the copy, not the parent, so it
    // cannot know which; saying neither is the honest option.
    ? `${owner} copy of "${props.forkedFrom}", edited. It replaces the original everywhere this team looks.`
    : `Written by ${owner === "your team's" ? 'your team' : props.teamName}, and theirs alone. No other team can see it.`
})
</script>
