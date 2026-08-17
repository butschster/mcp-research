<template>
  <div class="templates-page">
    <PageHeader
      :crumbs="[{ label: 'Research', to: '/' }, { label: 'Methodologies' }]"
      title="Methodologies"
    />

    <p class="lead">
      A methodology is what an agent reads before it starts a research: what to ask you first,
      what structure to propose, and when the work is finished. It creates nothing by itself —
      the agent picks one at kickoff and follows it.
    </p>

    <div v-if="pending" class="skeleton-card" style="height: 320px"></div>

    <EmptyState
      v-else-if="error"
      icon="&#x26A0;"
      title="Couldn't load the methodologies"
      description="The server didn't answer. Nothing is wrong with your research."
    >
      <button class="btn btn-primary" @click="reload">Try again</button>
    </EmptyState>

    <template v-else>
      <div class="filter-row">
        <input
          v-model="filter"
          type="search"
          class="text-input"
          placeholder="Filter by name or by when it is used…"
          aria-label="Filter methodologies"
        />
      </div>

      <div v-if="teamTemplates.length" class="card">
        <TemplateRowList
          :templates="teamTemplates"
          :team-name="teamName"
          heading="Your teams"
          :note="teamTemplates.length"
          blurb="Written by a team you belong to, and theirs alone. A copy of a shipped methodology replaces the original everywhere that team looks."
        />
      </div>

      <div class="card">
        <TemplateRowList
          :templates="globalTemplates"
          heading="Ships with the app"
          :note="globalTemplates.length"
          blurb="Written by us and updated with the binary. Editing one makes a copy for your team; the original stays as it is for everybody else."
          empty-text="None loaded. That is a fault on the server rather than a setting — the app ships with four."
        />
      </div>

      <p v-if="filter && !globalTemplates.length && !teamTemplates.length" class="no-match">
        Nothing matches “{{ filter }}”.
      </p>

      <div class="card">
        <h3 class="card-section-title">Using one</h3>
        <p class="lead lead--tight">
          There is no button here that starts a research: the web interface cannot create one yet.
          Ask your AI assistant instead — it reads the same list you are reading.
        </p>
        <CopyLine
          text="Start a new research. Check template_list first and follow the methodology that fits."
          label="Paste this into your AI client"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
const { data, pending, error, refresh } = await useApi<{ data: any[] }>('/api/templates')

const all = computed<any[]>(() => data.value?.data ?? [])
const filter = ref('')

/* Name and the two criteria, because the criteria are what somebody is actually
   scanning for — "which of these is for the thing I am about to do". */
const matching = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return all.value
  return all.value.filter(tp =>
    [tp.name, tp.description, tp.when_to_use, tp.when_not_to_use]
      .some((f: string) => (f || '').toLowerCase().includes(q)),
  )
})

const teamTemplates = computed(() => matching.value.filter(tp => tp.tier === 'team'))
const globalTemplates = computed(() => matching.value.filter(tp => tp.tier !== 'team'))

const { teams } = useTeams()
function teamName(id: string) {
  return teams.value?.find((t: any) => t.id === id)?.name
}

function reload() {
  refresh()
}
</script>

<style scoped>
.lead {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-bottom: var(--space-6);
}
.lead--tight { margin-bottom: var(--space-4); }

.filter-row { margin-bottom: var(--space-5); }
.filter-row .text-input { max-width: 26rem; }

.card + .card { margin-top: var(--space-6); }
.card-section-title { margin-bottom: var(--space-2); }

.no-match {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin-top: var(--space-5);
}
</style>
