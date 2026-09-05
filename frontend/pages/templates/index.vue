<template>
  <div class="templates-page">
    <PageHeader
      :crumbs="[{ label: 'Projects', to: { name: 'index' } }, { label: 'Methodologies' }]"
      title="Methodologies"
    />

    <p class="lead">
      Choose a methodology, copy its prompt, and paste it into your connected AI assistant to start a project.
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
      description="Your projects are still available. Try loading the methodologies again."
    >
      <button class="btn btn-primary" @click="reload">Try again</button>
    </EmptyState>

    <template v-else>
      <div class="filter-row">
        <input
          v-model="filter"
          type="search"
          class="text-input"
          placeholder="Find a methodology…"
          aria-label="Filter methodologies"
        />
      </div>

      <div v-if="teamTemplates.length" class="card card--list">
        <TemplateRowList
          :templates="teamTemplates"
          :team-name="teamName"
          heading="Team methodologies"
          :note="teamTemplates.length"
          blurb="Guides created or adapted by your teams."
        />
      </div>

      <div v-if="houseTemplates.length" class="card card--list">
        <TemplateRowList
          :templates="houseTemplates"
          heading="Shared methodologies"
          :note="houseTemplates.length"
          blurb="Custom guides available to every team on this server."
        />
      </div>

      <!-- Hidden while filtering, unlike the other two, which are hidden whenever
           they are empty. Its empty text says the server is at fault, and a filter
           that simply matches nothing here is not a fault — it used to print the
           accusation directly above a healthy list of somebody's team templates. -->
      <div v-if="!filter || shippedTemplates.length" class="card card--list">
        <TemplateRowList
          :templates="shippedTemplates"
          heading="Included with Dovod"
          :note="shippedTemplates.length"
          blurb="Starting points for common questions and decisions."
          empty-text="Built-in methodologies are unavailable. Contact the server administrator."
        />
      </div>

      <p v-if="filter" class="no-match" aria-live="polite">
        <template v-if="matching.length">{{ matching.length }} of {{ all.length }} shown.</template>
        <template v-else>Nothing matches “{{ filter }}”.</template>
      </p>


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
  margin-bottom: var(--space-6);
}

.filter-row { margin-bottom: var(--space-5); }
.filter-row .text-input { max-width: 26rem; }

.card + .card { margin-top: var(--space-6); }

.no-match {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin-top: var(--space-5);
}



.skeleton-groups { display: flex; flex-direction: column; gap: var(--space-6); }
.skeleton-group { height: 22rem; }
</style>
