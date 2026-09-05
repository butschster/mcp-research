<template>
  <!-- role="alert" because this arrives unasked: the page the reader was on has
       just been replaced under them. Without it the content simply becomes
       different content, announced to nobody, with the virtual cursor's anchor
       gone. tabindex="-1" so focus can be moved here from whatever unmounted —
       otherwise it falls to <body> and the next Tab restarts at the top of the
       document, several stops from the only action left. -->
  <div ref="root" class="access-revoked" role="alert" tabindex="-1">
    <EmptyState icon="&#x26D4;" :title="copy.title" :description="copy.long">
      <NuxtLink :to="backTo" class="btn btn-primary">{{ backLabel }}</NuxtLink>
    </EmptyState>
  </div>
</template>

<script setup lang="ts">
import { revocationCopy } from '~/composables/useAccessRevoked'
import type { Revocation } from '~/composables/useAccessRevoked'

/**
 * Stands in for a page whose subject the reader has just lost access to.
 *
 * It replaces the body rather than redirecting: the URL still explains where
 * they were, and nothing they were looking at is taken off the screen without
 * them being told why.
 */
const props = defineProps<{ revocation: Revocation }>()

// The wording lives with the type, so the toast raised for the same event and
// this notice cannot say different things about it.
const copy = computed(() => revocationCopy(props.revocation))

// A team page's way out is the team list, not the research list it was never on.
const backTo = computed(() => (props.revocation.scope === 'team' ? '/teams' : { name: 'index' }))
const backLabel = computed(() => (props.revocation.scope === 'team' ? 'Back to teams' : 'Back to projects'))

const root = ref<HTMLElement | null>(null)
onMounted(() => root.value?.focus())
</script>

<style scoped>
.access-revoked:focus { outline: none; }
.access-revoked:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 4px;
}
</style>
