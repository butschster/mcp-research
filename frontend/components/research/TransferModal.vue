<template>
  <ModalOverlay :visible="visible" size="sm" :labelledby="titleId" @close="emit('close')">
    <h3 :id="titleId" class="modal-title">Move project to another team</h3>

    <!-- Nothing to move it to. Explaining beats a disabled menu item that
         never says why. -->
    <template v-if="!targets.length">
      <p class="modal-help">
        You own only {{ currentTeamName || 'this team' }}. Create a team first, then move the project into it.
      </p>
      <div class="modal-actions">
        <button class="btn btn-sm" @click="emit('close')">Cancel</button>
        <NuxtLink class="btn btn-sm btn-primary" to="/teams">Create a team</NuxtLink>
      </div>
    </template>

    <template v-else>
      <p class="modal-help">
        <strong>{{ research.code }}</strong> {{ research.name }} is in {{ currentTeamName }}.
      </p>

      <label class="field-label" :for="selectId">Move to</label>
      <select :id="selectId" v-model="target" class="team-picker">
        <option v-for="team in targets" :key="team.id" :value="team.id">{{ team.name }}</option>
      </select>

      <p class="consequence">
        Everyone in {{ currentTeamName }} loses access. Everyone in {{ targetName }} gains it.
      </p>

      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>

      <div class="modal-actions">
        <button class="btn btn-sm" :disabled="busy" @click="emit('close')">Cancel</button>
        <button class="btn btn-sm btn-primary" :disabled="busy || !target" @click="emit('transfer', target)">
          {{ busy ? 'Moving…' : 'Move' }}
        </button>
      </div>
    </template>
  </ModalOverlay>
</template>

<script setup lang="ts">
import type { Team } from '~/composables/useTeams'

/**
 * Moves a research between teams — the only operation that changes who can see
 * it, so it states the consequence in both directions before it runs.
 *
 * The short code does not change: `/research/R7` links have to survive a move,
 * and codes are global rather than per team.
 */
const props = defineProps<{
  visible: boolean
  research: { code?: string; name: string }
  currentTeamId: string
  currentTeamName: string
  /** Teams the caller owns; the current one is filtered out here. */
  teams: Team[]
  busy?: boolean
  error?: string
}>()

const emit = defineEmits<{ transfer: [teamId: string]; close: [] }>()

const uid = useId()
const titleId = `transfer-title-${uid}`
const selectId = `transfer-target-${uid}`

const targets = computed(() => props.teams.filter((t) => t.id !== props.currentTeamId))
const target = ref('')
const targetName = computed(() => targets.value.find((t) => t.id === target.value)?.name ?? 'the other team')

// Both the open and the arrival of the teams: whichever happens second is the
// one that can pick a default, and Move stays disabled until one does.
watch(
  [() => props.visible, targets],
  ([open]) => {
    if (open && !targets.value.some((t) => t.id === target.value)) {
      target.value = targets.value[0]?.id ?? ''
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.team-picker {
  width: 100%;
  padding: 0.5rem 0.7rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font: inherit;
  font-size: var(--type-sm);
}
.consequence { font-size: var(--type-xs); color: var(--color-warning); margin: var(--space-3) 0 0; }
</style>
