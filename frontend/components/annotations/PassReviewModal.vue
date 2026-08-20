<template>
  <ModalOverlay :visible="visible" size="xl" flush labelledby="pass-review-title" @close="$emit('close')">
    <ModalHeader title="Review the pass" title-id="pass-review-title" @close="$emit('close')" />

    <div class="pass">
      <p class="pass__lead">
        {{ annotations.length }} {{ annotations.length === 1 ? 'mark was' : 'marks were' }} answered and
        {{ annotations.length === 1 ? 'is' : 'are' }} waiting for you. Accept them together, or send back the
        ones that are not settled.
      </p>

      <div class="pass__body">
        <div class="pass__rail">
          <AnnotationsAnnotationList
            :annotations="annotations"
            :research-slug="researchSlug"
            grouped
            selectable
            dense
            :selected-ids="selected"
            empty-variant="research"
            @open="(a) => (activeId = a.id)"
            @toggle-select="toggle"
          />
        </div>

        <div class="pass__detail">
          <template v-if="active">
            <header class="pass__detail-head">
              <ShortCode :code="active.code" />
              <AnnotationsKindChip :kind="active.kind" size="sm" />
              <AnnotationsAnchorBadge
                :state="active.anchor?.state ?? 'anchored'"
                :confidence="active.anchor?.confidence"
                :entry-type="active.entry_type"
              />
            </header>

            <blockquote class="pass__quote">{{ active.quote.exact }}</blockquote>
            <p v-if="active.body" class="pass__note" v-html="renderRefs(active.body, researchSlug)" />

            <h4 class="pass__answer-title">What the agent did</h4>
            <p class="pass__answer" v-html="renderRefs(active.resolution || '—', researchSlug)" />
          </template>

          <EmptyState
            v-else
            icon="◧"
            title="Pick a mark"
            description="Its answer shows here, so you can accept the batch knowing what is in it."
          />
        </div>
      </div>

      <!-- A partial failure is its own outcome. "Twelve of fourteen" has to be
           sayable without asking the server again, which is why the bulk call
           reports per row. -->
      <p v-if="result" class="pass__result" role="status">
        {{ result }}
      </p>

      <footer class="pass__actions">
        <span class="pass__selected">{{ selected.length }} selected</span>
        <button type="button" class="btn btn-sm" :disabled="!annotations.length" @click="selectAll">
          Select all
        </button>
        <span class="pass__spacer" />
        <button type="button" class="btn btn-sm" :disabled="!selected.length || busy" @click="sendBackOpen = true">
          Send back
        </button>
        <button type="button" class="btn btn-sm btn-primary" :disabled="!selected.length || busy" @click="$emit('accept', selected)">
          {{ busy ? 'Working…' : `Accept ${selected.length || ''}` }}
        </button>
      </footer>

      <AnnotationsSendBackModal
        :visible="sendBackOpen"
        :count="selected.length"
        :busy="busy"
        @confirm="onSendBack"
        @cancel="sendBackOpen = false"
      />
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
/**
 * Accepting a pass.
 *
 * The point of this screen is that it is not fifty clicks. A person reviews
 * what the agent did as one batch, sees each answer beside the sentence it
 * addresses, and decides once — which is the only way a queue of any size stays
 * worth keeping.
 *
 * HistoryPanel is the nearest thing in the catalog and does not fit: it is a
 * reader for one document's history, with no selection and no batch action.
 */
import type { Annotation } from '~/composables/useAnnotations'
import { renderRefs } from '~/composables/useCrossRefs'

const props = defineProps<{
  visible: boolean
  researchSlug: string
  annotations: Annotation[]
  busy?: boolean
  result?: string | null
}>()

const emit = defineEmits<{
  accept: [ids: string[]]
  'send-back': [ids: string[], reason: string]
  close: []
}>()

const selected = ref<string[]>([])
const activeId = ref<string | null>(null)

const active = computed(() => props.annotations.find((a) => a.id === activeId.value) ?? null)

function toggle(a: Annotation) {
  selected.value = selected.value.includes(a.id)
    ? selected.value.filter((id) => id !== a.id)
    : [...selected.value, a.id]
}

function selectAll() {
  selected.value = props.annotations.map((a) => a.id)
}

const sendBackOpen = ref(false)

function onSendBack(reason: string) {
  sendBackOpen.value = false
  emit('send-back', selected.value, reason)
}

// A fresh batch is a fresh decision: carrying a selection across would accept
// rows nobody looked at.
watch(() => props.visible, (open) => {
  if (open) {
    selected.value = []
    activeId.value = props.annotations[0]?.id ?? null
  }
})
</script>

<style scoped>
.pass {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-4);
}

.pass__lead {
  margin: 0;
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.pass__body {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: var(--space-4);
  align-items: start;
}

.pass__rail {
  max-height: 50vh;
  overflow-y: auto;
}

.pass__detail {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.pass__detail-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.pass__quote {
  margin: 0;
  padding-left: var(--space-3);
  border-left: 2px solid var(--color-border-strong);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
}

.pass__note {
  margin: 0;
  font-size: var(--type-sm);
}

.pass__answer-title {
  margin: var(--space-2) 0 0;
  font-size: var(--type-2xs);
  font-weight: 600;
  color: var(--color-text-muted);
}

.pass__answer {
  margin: 0;
  font-size: var(--type-sm);
}

.pass__result {
  margin: 0;
  font-size: var(--type-2xs);
  color: var(--color-text-muted);
}

.pass__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.pass__spacer { flex: 1; }

.pass__selected {
  font-size: var(--type-2xs);
  color: var(--color-text-faint);
}

@media (max-width: 768px) {
  .pass__body { grid-template-columns: 1fr; }
  .pass__rail { max-height: 32vh; }
}
</style>
