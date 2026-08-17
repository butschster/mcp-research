<template>
  <ModalOverlay :labelledby="titleId" :visible="visible" size="lg" flush @close="emit('close')">
    <ModalHeader :title="skill?.name || 'Skill'" :title-id="titleId" @close="emit('close')" />

    <div class="modal-body">
      <div v-if="loading" class="skeleton-card" style="height: 320px"></div>

      <div v-else-if="error" class="inline-error" role="alert">
        Couldn't load this skill. It is still attached and the agent can still open it —
        this page just failed to fetch the text.
      </div>

      <template v-else-if="skill">
        <div class="skill-head">
          <ResearchSettingsSkillOriginBadge
            :tier="skill.tier"
            :ambient="skill.ambient"
            :forked-from="skill.forked_from"
          />
          <span class="skill-tokens">{{ skill.body_tokens }} tokens</span>
        </div>

        <!-- The trigger line is set apart because it is the only part always in
             the agent's context. Everything below it is read on demand. -->
        <section class="trigger-block">
          <h4 class="block-label">When the agent uses this</h4>
          <p class="trigger-text">{{ skill.description }}</p>
        </section>

        <section>
          <h4 class="block-label">What it says</h4>
          <div class="markdown-content skill-body" v-html="rendered"></div>
        </section>
      </template>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
import { parseMarkdown } from '~/composables/useSafeMarkdown'

const props = defineProps<{
  visible: boolean
  skill: any | null
  loading?: boolean
  error?: boolean
}>()

const emit = defineEmits<{ close: [] }>()

const titleId = useId()

/* parseMarkdown, and deliberately no cross-reference pass: a skill is not
   research content, so `[[E3]]` inside one refers to nothing and linking it
   would invent a relationship that does not exist. */
const rendered = computed(() => parseMarkdown(props.skill?.body ?? ''))
</script>

<style scoped>
/* Flush, so the header's rule runs to the dialog's edges and the body's
   padding is not stacked on top of the card's own. */
.modal-body { padding: var(--space-5) var(--space-6); overflow-y: auto; }

.skill-head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
}
.skill-tokens {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  font-variant-numeric: tabular-nums;
}

.block-label {
  font-size: var(--type-xs);
  font-weight: var(--weight-medium);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: var(--space-2);
}

.trigger-block {
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.trigger-text {
  font-size: var(--type-sm);
  color: var(--color-text);
  max-width: var(--measure-prose);
  overflow-wrap: anywhere;
}

.skill-body { max-width: var(--measure-prose); }
/* A skill body carries tables and fenced code. Neither may push the modal
   sideways; they scroll inside themselves instead. */
.skill-body :deep(pre),
.skill-body :deep(table) { overflow-x: auto; display: block; max-width: 100%; }
</style>
