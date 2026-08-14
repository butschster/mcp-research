<template>
  <div>
    <Breadcrumbs :crumbs="[{ label: 'Research', to: '/' }, { label: 'Teams' }]" />

    <div class="page-header">
      <div class="page-header-row">
        <h1 class="page-title">Teams</h1>
        <button v-if="showNewButton" class="btn btn-sm btn-primary" @click="creating = true">+ New team</button>
      </div>
      <p class="page-lead">Teams own researches. Everyone in a team sees its researches.</p>
    </div>

    <div v-if="!loaded && !failed" class="skeleton-list">
      <div v-for="i in 3" :key="i" class="skeleton-card team-skeleton"></div>
    </div>

    <EmptyState
      v-else-if="failed"
      icon="&#x26A0;"
      title="Couldn't load your teams"
      description="The server didn't answer. Your researches are unaffected."
    >
      <button class="btn btn-sm" @click="refresh()">Try again</button>
    </EmptyState>

    <!-- One personal team is not a list; it is a user who has not met the
         feature yet. -->
    <EmptyState
      v-else-if="onlyPersonal"
      icon="&#x1F91D;"
      title="You're working on your own"
      description="Teams let other people into your researches. Create one, invite a colleague with a link, and move a research into it."
    >
      <button v-if="authEnabled" class="btn btn-sm btn-primary" @click="creating = true">+ New team</button>
    </EmptyState>

    <TeamRowList v-else :teams="[...teams]" />

    <ModalOverlay :visible="creating" size="sm" labelledby="new-team-title" @close="closeCreate">
      <h3 id="new-team-title" class="modal-title">New team</h3>
      <label class="field-label" for="new-team-name">Name</label>
      <input
        id="new-team-name"
        ref="nameInput"
        v-model="name"
        class="text-input"
        placeholder="Отдел интеграций"
        @keydown.enter="submit"
      />
      <p v-if="error" class="inline-error" role="alert">{{ error }}</p>
      <div class="modal-actions">
        <button class="btn btn-sm" :disabled="busy" @click="closeCreate">Cancel</button>
        <button class="btn btn-sm btn-primary" :disabled="busy || !name.trim()" @click="submit">
          {{ busy ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </ModalOverlay>
  </div>
</template>

<script setup lang="ts">
const { authEnabled } = useAuth()
const { teams, loading, loaded, failed, load, refresh, create } = useTeams()
const { success } = useToasts()

const creating = ref(false)
const name = ref('')
const busy = ref(false)
const error = ref('')
const nameInput = ref<HTMLInputElement | null>(null)

// With auth off there is nobody to be in a team with. The route stays
// reachable so a bookmark does not break, and says nothing.
const onlyPersonal = computed(() => teams.value.every((t) => t.personal))
const showNewButton = computed(() => !!authEnabled.value && !onlyPersonal.value)

onMounted(() => load())

watch(creating, async (open) => {
  if (!open) return
  error.value = ''
  name.value = ''
  await nextTick()
  nameInput.value?.focus()
})

function closeCreate() {
  if (busy.value) return
  creating.value = false
}

async function submit() {
  if (!name.value.trim() || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const team = await create(name.value.trim())
    creating.value = false
    success(`${team.name} is ready. Invite someone with a link.`, 'Team created')
    navigateTo(`/teams/${team.id}`)
  } catch (e: any) {
    error.value = e?.data?.error || 'Could not create the team'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.page-lead { font-size: var(--type-sm); color: var(--color-text-muted); max-width: 65ch; margin-top: var(--space-2); }
.team-skeleton { height: 56px; }
.modal-title {
  font-size: var(--type-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  margin: 0 0 var(--space-4);
}
.field-label { display: block; font-size: var(--type-xs); color: var(--color-text-muted); margin-bottom: var(--space-1); }
.text-input {
  width: 100%;
  padding: 0.5rem 0.7rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font: inherit;
  font-size: var(--type-sm);
}
.text-input:focus { outline: none; border-color: var(--color-primary); }
.inline-error { font-size: var(--type-xs); color: var(--color-error); margin: var(--space-3) 0 0; }
.modal-actions { display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-5); }
</style>
