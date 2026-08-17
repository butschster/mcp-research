<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="research" class="settings-page">
    <PageHeader
      :crumbs="[
        { label: 'Research', to: '/' },
        { label: research.name, to: `/research/${researchSlug}` },
        { label: 'Settings' },
      ]"
      :code="research.code"
      title="Settings"
    >
      <template #actions>
        <TeamViewerNotice v-if="isViewer" :team-name="research?.team_name" />
      </template>
    </PageHeader>

    <TabBar v-model="activeTab" :tabs="tabs" label="Research settings" />

    <!-- Overview -->
    <div v-if="activeTab === 'overview'" id="panel-overview" role="tabpanel" aria-labelledby="tab-overview" tabindex="0" class="panel">
      <div class="card">
        <h3 class="card-section-title">Ownership</h3>
        <p class="lead">
          <template v-if="research.team_is_personal">
            In your personal space. Nobody else has a role on it.
          </template>
          <template v-else>
            Owned by <NuxtLink :to="`/teams/${research.team_id}`" class="team-link">{{ research.team_name }}</NuxtLink>,
            where your role is <strong>{{ research.role }}</strong>.
          </template>
          Share links are managed from the research page.
        </p>
      </div>

      <div class="card">
        <div class="field-group">
          <EditableField
            label="Goal"
            :value="research.goal"
            :editable="canWrite"
            placeholder="What is this research trying to achieve?"
            :empty-text="canWrite ? 'Click the pencil to set a goal' : 'Not set'"
            @save="v => save('goal', v)"
          />
          <EditableField
            label="Description"
            :value="research.description"
            :editable="canWrite"
            multiline
            placeholder="Describe what this research covers..."
            :empty-text="canWrite ? 'Click the pencil to add a description' : 'Not set'"
            @save="v => save('description', v)"
          />
          <EditableField
            label="Tags"
            :value="(research.tags ?? []).join(', ')"
            :editable="canWrite"
            placeholder="tag1, tag2, tag3"
            empty-text="No tags yet"
            @save="v => save('tags', v.split(',').map((t: string) => t.trim()).filter(Boolean))"
          >
            <template #default>
              <TagList v-if="research.tags?.length" :tags="research.tags" />
              <span v-else class="field-empty">No tags yet</span>
            </template>
          </EditableField>
          <EditableField
            label="AI Instruction"
            :value="research.instruction"
            :editable="canWrite"
            multiline
            :rows="6"
            placeholder="How should the agent work on this research?"
            empty-text="No instruction set"
            @save="v => save('instruction', v)"
          />
        </div>
      </div>
    </div>

    <!-- Skills -->
    <div v-else-if="activeTab === 'skills'" id="panel-skills" role="tabpanel" aria-labelledby="tab-skills" tabindex="0" class="panel">
      <p class="lead">
        What the agent follows on this research. It sees every name and trigger line below on
        each call, and opens a skill's text only when the line matches what it is about to do.
        A skill belonging to this research overrides a team one, which overrides a built-in.
      </p>

      <div v-if="skillsPending" class="skeleton-card" style="height: 200px"></div>

      <template v-else>
        <div class="card card--list">
          <ResearchSettingsSkillRowList
            :skills="chosen"
            :can-write="canWrite"
            :busy-slug="busySlug"
            heading="Chosen for this research"
            :note="`${chosenCount} of ${cap}`"
            blurb="Methodology somebody picked. These are what the six-skill budget counts."
            empty-text="Nothing chosen yet — the agent works from the built-in skills below alone."
            @open="openSkill"
            @detach="detach"
          />
        </div>

        <div class="card card--list">
          <ResearchSettingsSkillRowList
            :skills="ambient"
            :can-write="canWrite"
            heading="Always on"
            blurb="Skills about the product itself — how to manage a research, how to write an entry. They ship with the app, are on everywhere, and are outside the budget."
            empty-text="None. The server has no built-in skills loaded, which is a fault rather than a setting."
            @open="openSkill"
          />
        </div>

        <div v-if="canWrite" class="card card--list">
          <ResearchSettingsSkillRowList
            :skills="library"
            :actions="false"
            :can-write="canWrite"
            :busy-slug="busySlug"
            heading="Available to attach"
            blurb="Built-in methodology and anything your team has written. Attaching one spends a slot in the budget above."
            empty-text="Nothing left to attach — everything available is already on."
            @open="openSkill"
            @attach="attach"
          />
        </div>
      </template>
    </div>

    <!-- Memory -->
    <div v-else-if="activeTab === 'memory'" id="panel-memory" role="tabpanel" aria-labelledby="tab-memory" tabindex="0" class="panel">
      <p class="lead">
        Notes the agent wrote for its future self. All of them are sent at the start of every
        session, so the list is a cost as well as a record.
      </p>
      <div class="card card--list">
        <div v-if="research.memory?.length" class="data-rows">
          <div v-for="(item, i) in (research.memory as string[])" :key="i" class="data-row memory-row">
            <span class="memory-index">{{ i + 1 }}</span>
            <span class="memory-text">{{ item }}</span>
          </div>
        </div>
        <p v-else class="list-empty">
          Nothing remembered yet. The agent adds a note when it learns something the next
          session should start with.
        </p>
      </div>
      <!-- Said plainly rather than implied by absent buttons: a screen with no
           controls reads as broken unless it says why. -->
      <p class="lead footnote">
        Memory cannot be edited here yet. A note has no id, date or author in the API today,
        so there is nothing to address — curating it needs that groundwork first.
      </p>
    </div>

    <ResearchSettingsSkillDetail
      :visible="!!openSlug"
      :skill="openedSkill"
      :loading="skillLoading"
      :error="skillError"
      @close="closeSkill"
    />
  </div>

  <EmptyState
    v-else
    icon="&#x1F50D;"
    title="Research not found"
    description="It may have been deleted, or the address may name a research that never existed."
  >
    <NuxtLink to="/" class="btn btn-primary">Back to all research</NuxtLink>
  </EmptyState>
</template>

<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const { data: researchData, pending } = await useApi<any>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchSlug = computed(() => research.value?.code || id)

const { canWrite, isViewer, setFromResearch } = useResearchRole()
watch(research, r => setFromResearch(r), { immediate: true })

/* The tab lives in the query string so a link can point at one, and it is
   replaced rather than pushed: Back should leave the page, not walk the tabs. */
const TABS = ['overview', 'skills', 'memory'] as const
type Tab = (typeof TABS)[number]
const activeTab = computed<Tab>({
  get: () => (TABS.includes(route.query.tab as Tab) ? (route.query.tab as Tab) : 'overview'),
  set: (tab) => router.replace({ query: { ...route.query, tab: tab === 'overview' ? undefined : tab } }),
})

const { authFetch } = useAuth()
const base = useRuntimeConfig().public.apiBase || ''

// --- Skills ---
const skillsData = ref<any>(null)
const libraryData = ref<any>(null)
const skillsPending = ref(true)

async function loadSkills() {
  skillsPending.value = true
  try {
    skillsData.value = await authFetch<any>(`${base}/api/researches/${id}/skills`)
    if (canWrite.value) {
      libraryData.value = await authFetch<any>(`${base}/api/researches/${id}/skills/library`)
    }
  } catch {
    skillsData.value = { data: [], cap: 6, chosen: 0 }
  } finally {
    skillsPending.value = false
  }
}
onMounted(loadSkills)

const allSkills = computed<any[]>(() => skillsData.value?.data ?? [])
const chosen = computed(() => allSkills.value.filter(s => !s.ambient))
const ambient = computed(() => allSkills.value.filter(s => s.ambient))
const chosenCount = computed(() => skillsData.value?.chosen ?? chosen.value.length)
const cap = computed(() => skillsData.value?.cap ?? 6)
const library = computed<any[]>(() => (libraryData.value?.data ?? []).filter((s: any) => !s.attached))

const tabs = computed(() => [
  { id: 'overview', label: 'Overview' },
  {
    id: 'skills',
    label: 'Skills',
    count: `${chosenCount.value}/${cap.value}`,
    srCount: `${chosenCount.value} of ${cap.value} chosen`,
  },
  {
    id: 'memory',
    label: 'Memory',
    count: research.value?.memory?.length ?? 0,
    srCount: `${research.value?.memory?.length ?? 0} notes`,
  },
])

const busySlug = ref<string | null>(null)
const toasts = useToasts()

/* Attach and detach wait for the server rather than moving the row first. The
   cap is a server rule, and a seventh row that appears and then vanishes is a
   worse answer than half a second of nothing. */
async function attach(sk: any) {
  busySlug.value = sk.slug
  try {
    await authFetch(`${base}/api/researches/${id}/skills`, { method: 'POST', body: { slug: sk.slug } })
    await loadSkills()
  } catch (e: any) {
    const code = e?.data?.code
    toasts.push({
      variant: 'error',
      title: code === 'skill_cap_reached' ? 'No room' : "Couldn't attach",
      message: code === 'skill_cap_reached'
        ? `${cap.value} skills are already chosen. Detach one to make room.`
        : e?.data?.error || 'The server refused that.',
    })
  } finally {
    busySlug.value = null
  }
}

async function detach(sk: any) {
  busySlug.value = sk.slug
  try {
    await authFetch(`${base}/api/researches/${id}/skills/${sk.slug}`, { method: 'DELETE' })
    await loadSkills()
  } catch (e: any) {
    toasts.push({ variant: 'error', title: "Couldn't detach", message: e?.data?.error || 'The server refused that.' })
  } finally {
    busySlug.value = null
  }
}

// --- Reading one skill ---
const openSlug = ref<string | null>(null)
const openedSkill = ref<any>(null)
const skillLoading = ref(false)
const skillError = ref(false)

async function openSkill(slug: string) {
  openSlug.value = slug
  openedSkill.value = allSkills.value.find(s => s.slug === slug)
    ?? library.value.find((s: any) => s.slug === slug)
    ?? null
  skillLoading.value = true
  skillError.value = false
  try {
    const res = await authFetch<any>(`${base}/api/researches/${id}/skills/${slug}`)
    openedSkill.value = res.data
  } catch {
    skillError.value = true
  } finally {
    skillLoading.value = false
  }
}

function closeSkill() {
  openSlug.value = null
  openedSkill.value = null
  skillError.value = false
}

// --- Overview writes ---
async function save(field: string, value: any) {
  try {
    await authFetch(`${base}/api/researches/${id}`, { method: 'PUT', body: { [field]: value } })
    researchData.value = await authFetch<any>(`${base}/api/researches/${id}`)
  } catch (e: any) {
    toasts.push({
      variant: 'error',
      title: 'Not saved',
      message: e?.data?.error || 'The server refused that change.',
    })
  }
}
</script>

<style scoped>
.panel { margin-top: var(--space-6); }
.card + .card { margin-top: var(--space-6); }

.lead {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-bottom: var(--space-5);
}
.footnote { margin-top: var(--space-5); margin-bottom: 0; font-size: var(--type-xs); }

.card-section-title { margin-bottom: var(--space-2); }
.team-link { color: var(--color-primary); }

.memory-row { display: flex; gap: var(--space-3); align-items: baseline; }
.memory-index {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
  min-width: 1.5rem;
}
.memory-text {
  font-size: var(--type-sm);
  /* A memory note can be a pasted log line with no spaces in it. */
  overflow-wrap: anywhere;
}
</style>
