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

    <div v-if="pending" class="skeleton-groups">
      <!-- One box per group the page is about to draw. A single 320px card for a
           list that lands at four times that height made the whole page jump the
           moment the data arrived. -->
      <div class="skeleton-card skeleton-group"></div>
      <div class="skeleton-card skeleton-group"></div>
    </div>

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
          blurb="Written by a team you belong to, and theirs alone. A copy of a global methodology replaces the original everywhere that team looks."
        />
      </div>

      <div v-if="houseTemplates.length" class="card">
        <TemplateRowList
          :templates="houseTemplates"
          heading="Added on this server"
          :note="houseTemplates.length"
          blurb="Written here by whoever runs this server, and visible to every team on it. Not part of the app, so an upgrade will not change them."
        />
      </div>

      <!-- Hidden while filtering, unlike the other two, which are hidden whenever
           they are empty. Its empty text says the server is at fault, and a filter
           that simply matches nothing here is not a fault — it used to print the
           accusation directly above a healthy list of somebody's team templates. -->
      <div v-if="!filter || shippedTemplates.length" class="card">
        <TemplateRowList
          :templates="shippedTemplates"
          heading="Ships with the app"
          :note="shippedTemplates.length"
          blurb="Written by us and updated with the binary. Editing one makes a copy for your team; the original stays as it is for everybody else."
          empty-text="None loaded. That is a fault on the server rather than a setting — the app ships with several."
        />
      </div>

      <p v-if="filter" class="no-match" aria-live="polite">
        <template v-if="matching.length">{{ matching.length }} of {{ all.length }} shown.</template>
        <template v-else>Nothing matches “{{ filter }}”.</template>
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
        <p class="footnote">
          Your team can write its own, and whoever runs this server can add one that every team
          here sees. Both are written through the API — see <code>/llms/templates.md</code>.
        </p>
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

/* The global tier holds two different things and a reader needs them apart: what
   we ship is rewritten on every upgrade and can only be forked, while what the
   operator added here was written by somebody they can go and ask. Grouping them
   together made the second look like an app feature. */
const globalTemplates = computed(() => matching.value.filter(tp => tp.tier !== 'team'))
const houseTemplates = computed(() => globalTemplates.value.filter(tp => tp.source === 'user'))
const shippedTemplates = computed(() => globalTemplates.value.filter(tp => tp.source !== 'user'))

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

.footnote {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-top: var(--space-4);
}

.skeleton-groups { display: flex; flex-direction: column; gap: var(--space-6); }
.skeleton-group { height: 22rem; }
</style>
