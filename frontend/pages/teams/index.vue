<template>
  <div>
    <Breadcrumbs :crumbs="[{ label: 'Projects', to: { name: 'index' } }, { label: 'Teams' }]" />

    <PageHeader title="Teams" lead="Team members can access all projects in their team.">
      <!-- Neutral, like every other page header in the product. A filled button
           lived here and on the team page and nowhere else, so these two pages
           read as imported from somewhere with a different design. -->
      <template #actions><button v-if="showNewButton" class="btn" @click="creating = true">+ New team</button></template>
    </PageHeader>

    <div v-if="!loaded && !failed" class="skeleton-list">
      <div v-for="i in 3" :key="i" class="skeleton-card team-skeleton"></div>
    </div>

    <EmptyState
      v-else-if="failed"
      icon="&#x26A0;"
      title="Couldn't load your teams"
      description="The server didn't answer. Your projects are unaffected."
    >
      <button class="btn" @click="refresh()">Try again</button>
    </EmptyState>

    <!-- One personal team is not a list; it is a user who has not met the
         feature yet. -->
    <EmptyState
      v-else-if="onlyPersonal"
      icon="&#x1F91D;"
      title="You're working on your own"
      description="Create a team, invite a colleague, and move a project into the team to work on it together."
    >
      <!-- The one filled button that survives on this surface: there is nothing
           else on the screen, and it is the action the page exists for. -->
      <button v-if="authEnabled" class="btn btn-primary" @click="creating = true">+ New team</button>
    </EmptyState>

    <!-- The list sits in a frame, like every other list in the product: the
         rules then run to the card's edges and the text is inset by the row
         rather than sitting four pixels off the page edge. -->
    <div v-else class="card card--list">
      <TeamRowList :teams="[...teams]" />
    </div>

    <ModalOverlay :visible="creating" size="sm" labelledby="new-team-title" @close="closeCreate">
      <h3 id="new-team-title" class="modal-title">New team</h3>
      <label class="field-label" for="new-team-name">Name</label>
      <input
        id="new-team-name"
        ref="nameInput"
        v-model="name"
        class="text-input"
        placeholder="Integrations"
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
    // Move first, then invite. The old wording recommended inviting first,
    // which is exactly how a colleague arrives to an empty list and concludes
    // the invitation failed — the researches the owner already had stayed
    // behind in their personal team, and nothing said so.
    success(`${team.name} is ready. Move a project into it, then invite someone.`, 'Team created')
    navigateTo(`/teams/${team.id}`)
  } catch (e: any) {
    error.value = e?.data?.error || 'Could not create the team'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.team-skeleton { height: 56px; }
</style>
