<template>
  <article class="thread" :class="busy && 'is-busy'">
    <header class="thread__head">
      <ShortCode :code="annotation.code" />
      <AnnotationsKindChip :kind="annotation.kind" size="sm" />
      <StatusBadge :status="annotation.status" />
      <AnnotationsAnchorBadge
        :state="annotation.anchor?.state ?? 'anchored'"
        :confidence="annotation.anchor?.confidence"
        :entry-type="annotation.entry_type"
      />
      <span class="thread__spacer" />
      <button type="button" class="btn btn-sm btn-icon" aria-label="Close thread" @click="$emit('close')">×</button>
    </header>

    <blockquote class="thread__quote">{{ annotation.quote.exact }}</blockquote>

    <!-- The one state a person has to act on rather than read past. The diff is
         the only honest way to answer "was my doubt addressed or buried". -->
    <p v-if="isLost" class="thread__lost">
      {{ lostMessage }}
      <button
        v-if="annotation.anchored_revision"
        type="button"
        class="thread__lost-link"
        @click="$emit('show-diff', annotation.anchored_revision)"
      >
        See what changed
      </button>
    </p>

    <!-- The slot is how a field says its value is not plain text: a note
         reading "contradicts [[E7]]" is a reference like any other. -->
    <EditableField
      label="Note"
      :value="annotation.body || ''"
      :editable="canWrite"
      multiline
      placeholder="What is wrong with it?"
      empty-text="No note"
      @save="(v: string) => $emit('update-body', v)"
    >
      <template #default="{ value }">
        <span v-if="value" v-html="renderRefs(value, researchSlug)" />
        <span v-else>No note</span>
      </template>
    </EditableField>

    <section v-if="annotation.resolution" class="thread__answer">
      <h4 class="thread__answer-title">
        Answer
        <span v-if="annotation.resolved_revision" class="thread__rev">rev {{ annotation.resolved_revision }}</span>
      </h4>
      <p class="thread__answer-body" v-html="renderRefs(annotation.resolution, researchSlug)" />
    </section>

    <!-- Kept in full and shown in full: an agent repeating a rejected answer is
         the commonest way a pass wastes itself, and these are the reasons it is
         required to read. -->
    <section v-if="annotation.rejections?.length" class="thread__rejections">
      <h4 class="thread__answer-title">Sent back {{ annotation.rejections.length }}×</h4>
      <p
        v-for="(r, i) in annotation.rejections"
        :key="i"
        class="thread__rejection"
        v-html="r.reason ? renderRefs(r.reason, researchSlug) : 'No reason recorded'"
      />
    </section>

    <footer v-if="canWrite" class="thread__actions">
      <template v-if="annotation.status === 'answered'">
        <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="$emit('accept')">Accept</button>
        <button type="button" class="btn btn-sm" :disabled="busy" @click="sendBackOpen = true">Send back</button>
      </template>
      <template v-else-if="annotation.status === 'open'">
        <button type="button" class="btn btn-sm" :disabled="busy" @click="$emit('dismiss')">Dismiss</button>
      </template>
      <template v-else>
        <button type="button" class="btn btn-sm" :disabled="busy" @click="sendBackOpen = true">Reopen</button>
      </template>
      <button type="button" class="btn btn-sm btn-danger" :disabled="busy" @click="confirmDelete = true">
        Delete
      </button>
    </footer>

    <AnnotationsSendBackModal
      :visible="sendBackOpen"
      :busy="busy"
      @confirm="onSendBack"
      @cancel="sendBackOpen = false"
    />

    <!-- A mis-drag — half a word, a stray line — is the one thing only the
         person who made it can judge. Dismissing it would leave a row saying
         "somebody considered this and let it go", which is not what happened. -->
    <ConfirmModal
      :visible="confirmDelete"
      title="Delete this mark?"
      :message="`${annotation.code} and everything recorded on it goes. Dismiss it instead if the doubt was real and is now settled.`"
      confirm-label="Delete"
      variant="danger"
      @confirm="onDelete"
      @cancel="confirmDelete = false"
    />
  </article>
</template>

<script setup lang="ts">
/**
 * One mark, opened.
 *
 * Rendered inline under the block it belongs to rather than in a modal. A modal
 * would cover the sentence the whole conversation is about, which is the one
 * thing the reader needs on screen while deciding.
 */
import type { Annotation } from '~/composables/useAnnotations'
import { renderRefs } from '~/composables/useCrossRefs'

const props = defineProps<{
  annotation: Annotation
  researchSlug: string
  canWrite?: boolean
  busy?: boolean
}>()

const emit = defineEmits<{
  'update-body': [value: string]
  accept: []
  dismiss: []
  reopen: [reason: string]
  delete: []
  /** The revision this mark was anchored to, so the page can open the history
   *  comparing it against now. A link to the page's own route did nothing:
   *  this card is rendered ON that page, so navigating there re-rendered
   *  nothing and the query string was read by no one. */
  'show-diff': [revision: number]
  close: []
}>()

const isLost = computed(() => {
  const state = props.annotation.anchor?.state
  return state === 'drifted' || state === 'orphaned'
})

const lostMessage = computed(() =>
  props.annotation.anchor?.state === 'orphaned'
    ? 'The text this marks is gone from the document.'
    : 'The text under this mark changed after it was made.')

const sendBackOpen = ref(false)
const confirmDelete = ref(false)

function onSendBack(reason: string) {
  sendBackOpen.value = false
  emit('reopen', reason)
}

function onDelete() {
  confirmDelete.value = false
  emit('delete')
}
</script>

<style scoped>
.thread {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-3);
  margin: var(--space-3) 0;
  background: var(--color-surface-raised);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.thread.is-busy { opacity: 0.6; pointer-events: none; }

.thread__head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.thread__spacer { flex: 1; }

.thread__quote {
  margin: 0;
  padding-left: var(--space-3);
  border-left: 2px solid var(--color-border-strong);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.thread__lost {
  margin: 0;
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  background: rgba(var(--color-warning-rgb), 0.10);
  color: var(--color-warning);
  font-size: var(--type-2xs);
}

.thread__lost-link {
  margin-left: var(--space-2);
  padding: 0;
  border: 0;
  background: none;
  color: inherit;
  font: inherit;
  text-decoration: underline;
  cursor: pointer;
}

.thread__answer-title {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  margin: 0 0 var(--space-1);
  font-size: var(--type-2xs);
  font-weight: 600;
  color: var(--color-text-muted);
}

.thread__rev { font-weight: 400; color: var(--color-text-faint); }

.thread__answer-body,
.thread__rejection {
  margin: 0 0 var(--space-1);
  font-size: var(--type-sm);
}

.thread__rejection { color: var(--color-text-muted); }

.thread__actions {
  display: flex;
  gap: var(--space-2);
}
</style>
