<template>
  <!-- The whole row is one link. A row with a link inside it and clickable
       space around it gives a keyboard user two targets for one destination,
       and a mouse user a dead zone between them. -->
  <NuxtLink :to="href" class="data-row resume-row">
    <div class="resume-row-main">
      <div class="resume-row-line">
        <span v-if="actor" :class="['resume-actor', `resume-actor--${actor}`]">{{ actorLabel }}</span>
        <ShortCode v-if="code" :code="code" />
        <span class="resume-row-title">{{ title }}</span>
        <!-- The evidence for the suggestion, on the same line as the thing it
             is about. A summary that says what to do and not why is asking to
             be trusted; on its own line it cost a second line of height per
             row, and this block sits above the documents. -->
        <span v-if="reason || note" class="resume-row-reason">{{ reason || note }}</span>
      </div>
    </div>

    <div class="resume-row-badges">
      <EntryUpdateBadge v-if="updateKind" :kind="updateKind" :unseen-revisions="unseenRevisions" />
      <StatusBadge v-if="status" :status="status" />
      <StatusBadge v-if="priority" :status="priority" />
      <span v-if="meta" class="resume-row-meta">{{ meta }}</span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
/**
 * One line of the Continue block — a proposed next action, or an item inside a
 * group. Both variants are the same anatomy: who acts, the short code, what it
 * is, and the badges that qualify it. Two components would have drifted within
 * a release, so there is one with a variant.
 */
const props = defineProps<{
  variant?: 'action' | 'item'
  code?: string
  title: string
  href: string
  status?: string
  priority?: string
  actor?: 'agent' | 'human'
  /** Why this is being suggested. Rendered verbatim; the server writes it. */
  reason?: string
  /** A recorded note on the item itself, shown when there is no reason. */
  note?: string
  /** A short trailing fact — a relative time, an entry code. */
  meta?: string
  updateKind?: 'new' | 'changed'
  unseenRevisions?: number
}>()

// "You", not "Human": the block is read by the person it is talking about.
const actorLabel = computed(() => (props.actor === 'human' ? 'You' : 'Agent'))
</script>

<style scoped>
/* Centred, like every other `.data-row` in the product. `flex-start` put a
   single-line row 2.4px above its optical centre in a 54px box. A two-line row
   fills the box either way. */
.resume-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  color: inherit;
  text-decoration: none;
}
.resume-row:hover { text-decoration: none; background: var(--color-surface-hover); }

.resume-row-main { min-width: 0; flex: 1; }
.resume-row-line {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  min-width: 0;
}
.resume-row-line > .short-code,
.resume-row-line > .resume-actor { flex: none; align-self: center; }
.resume-row-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  /* One line, then an ellipsis. A four-hundred-character title is a real
     document title here, and the full text is on the page the row links to. */
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
/* Faint rather than muted: this sits on a row that lightens on hover, and muted
   text at this size does not hold its contrast against that background. It is
   the first thing to be cut when the row runs out of width — the title and the
   code are what identify the thing. */
.resume-row-reason {
  flex: 0 1 auto;
  min-width: 0;
  font-size: var(--type-xs);
  color: var(--color-text-faint);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.resume-row-badges {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: none;
}
.resume-row-meta { font-size: var(--type-xs); color: var(--color-text-faint); white-space: nowrap; }

/* Who is expected to act. It carries a word, never a colour alone. */
.resume-actor {
  display: inline-flex;
  align-items: center;
  min-height: var(--control-h-sm);
  padding: 0 0.4rem;
  border-radius: var(--radius-xs);
  font-size: var(--type-3xs);
  font-weight: var(--weight-semibold);
  line-height: 1;
  flex: none;
}
/* The hues follow AuthorBadge rather than inventing a second scheme: the agent
   is the default and the default is quiet, the person is the one to look at.
   The word carries the meaning either way — this is a second signal, not the
   only one. */
.resume-actor--agent { background: var(--color-surface-hover); color: var(--color-text-muted); }
.resume-actor--human { background: var(--color-primary-muted); color: var(--color-primary); }

@media (max-width: 768px) {
  .resume-row { flex-direction: column; align-items: stretch; gap: var(--space-1); }
  .resume-row-badges { flex-wrap: wrap; }
  /* On a phone the row is already stacked, so the reason gets its own line
     rather than competing with the title for a narrow one. */
  .resume-row-line { flex-wrap: wrap; }
}
</style>
