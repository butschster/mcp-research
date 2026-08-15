<template>
  <!-- `lg`, because the body is a list of names and `md` is 460px: half the
       titles wrapped and the team's own name wrapped in the header. The header
       says only "Move researches here" for the same reason — `.modal-title` is
       uppercase, and a Cyrillic team name shouted across two lines is not a
       heading anybody reads. The team is named in the sentence below it. -->
  <ModalOverlay :visible="visible" size="lg" :labelledby="titleId" @close="emit('close')">
    <ModalHeader :title-id="titleId" title="Move researches here" @close="emit('close')" />

    <div class="dialog-body">
      <!-- Nothing to move. Said plainly, rather than as an empty list with a
           disabled button that never explains itself. -->
      <template v-if="!candidates.length">
        <p class="dialog-help">
          Every research you own is already in {{ teamName }}. Anything an agent
          creates from now on lands in your personal team, and you can move it
          across from here or from the research itself.
        </p>
      </template>

      <template v-else>
        <p class="dialog-help">
          Everyone in {{ teamName }} can read what you move here. It stops being
          visible to whoever could see it before, unless they are in this team too.
        </p>

        <div v-if="candidates.length > 8" class="filter-row">
          <input
            v-model="filter"
            class="text-input"
            type="search"
            :placeholder="`Filter ${candidates.length} researches`"
            :aria-label="`Filter ${candidates.length} researches`"
          />
        </div>

        <ul class="candidate-list">
          <li v-for="research in filtered" :key="research.id">
            <label class="check-row">
              <input v-model="picked" type="checkbox" :value="research.id" />
              <span class="candidate-text">
                <span class="candidate-name">
                  <span v-if="research.code" class="candidate-code">{{ research.code }}</span>
                  {{ research.name }}
                </span>
                <span class="candidate-meta">
                  {{ research.team_name || 'Personal' }}
                  <template v-if="research.status !== 'active'"> · {{ research.status }}</template>
                </span>
              </span>
            </label>
          </li>
        </ul>
        <p v-if="filter && !filtered.length" class="dialog-help">Nothing matches “{{ filter }}”.</p>
      </template>

      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>
    </div>

    <div class="modal-actions">
      <button class="btn btn-sm" :disabled="busy" @click="emit('close')">Cancel</button>
      <button
        v-if="candidates.length"
        class="btn btn-sm btn-primary"
        :disabled="busy || !picked.length"
        @click="emit('move', [...picked])"
      >
        {{ busy ? 'Moving…' : moveLabel }}
      </button>
    </div>
  </ModalOverlay>
</template>

<script setup lang="ts">
/**
 * Pulls researches into a team.
 *
 * `research/TransferModal` asks the same question from the other end: standing
 * on one research, where should it go. This one stands on a team and asks what
 * belongs in it — which is the direction a brand-new team needs, and the one
 * the product had no control for at all. An owner who created a team and
 * invited a colleague had to go back to each research in turn and find "Move to
 * team…" in its `⋯` menu.
 *
 * It sends one request for the whole selection. Twelve requests would be twelve
 * ways to half-succeed, and no single place to say so.
 */
const props = withDefaults(
  defineProps<{
    visible: boolean
    teamName: string
    /** Researches the caller owns that are not in this team already. */
    candidates: { id: string; code?: string; name: string; status: string; team_name?: string }[]
    busy?: boolean
    error?: string
  }>(),
  { busy: false, error: '' },
)

const emit = defineEmits<{ move: [ids: string[]]; close: [] }>()

const titleId = `add-research-${useId()}`
const picked = ref<string[]>([])
const filter = ref('')

const filtered = computed(() => {
  const needle = filter.value.trim().toLowerCase()
  if (!needle) return props.candidates
  return props.candidates.filter(
    (r) => r.name.toLowerCase().includes(needle) || (r.code || '').toLowerCase().includes(needle),
  )
})

const moveLabel = computed(() =>
  picked.value.length <= 1 ? 'Move' : `Move ${picked.value.length} researches`,
)

// Every opening starts empty, including one that follows a cancelled dialog:
// a selection that survives is a selection nobody remembers making.
watch(
  () => props.visible,
  (open) => {
    if (open) {
      picked.value = []
      filter.value = ''
    }
  },
)
</script>

<style scoped>
.dialog-body { padding: var(--space-4) var(--space-6); }
.dialog-help { font-size: var(--type-sm); color: var(--color-text-muted); margin: 0 0 var(--space-3); }
.filter-row { margin-bottom: var(--space-2); }
.candidate-list {
  list-style: none;
  margin: 0;
  padding: 0;
  /* A list of forty is a scroll inside the dialog, not a dialog taller than
     the window with its buttons below the fold. */
  max-height: 22rem;
  overflow-y: auto;
}
.candidate-text { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.candidate-name { font-size: var(--type-sm); overflow-wrap: anywhere; }
.candidate-code { color: var(--color-text-muted); font-variant-numeric: tabular-nums; margin-right: var(--space-1); }
.candidate-meta { font-size: var(--type-xs); color: var(--color-text-muted); }
</style>
