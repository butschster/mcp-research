<template>
  <div :class="['updates-list', { 'updates-list--refreshing': refreshing }]" :aria-busy="loading || refreshing">
    <div v-if="loading" class="updates-skeletons">
      <div v-for="i in 5" :key="i" class="skeleton-card update-skeleton"></div>
    </div>

    <EmptyState
      v-else-if="error"
      icon="&#x26A0;"
      title="Could not load updates"
      description="Your read state was not changed."
    >
      <button type="button" class="btn btn-sm" @click="$emit('retry')">Try again</button>
    </EmptyState>

    <EmptyState
      v-else-if="!updates.length"
      icon="&#x2713;"
      title="You’re up to date"
      description="There are no new or changed documents in this project."
    >
      <NuxtLink :to="researchPath(researchSlug)" class="btn btn-sm">Browse all documents</NuxtLink>
    </EmptyState>

    <template v-else>
      <section v-if="newEntries.length" class="updates-group" aria-labelledby="new-updates-title">
        <h2 id="new-updates-title" class="card-section-title">
          New <span class="count-chip">{{ newEntries.length }}</span>
        </h2>
        <div class="card card--list">
          <div class="data-rows">
            <ResearchUpdatesRow
              v-for="update in visibleNew"
              :key="update.entry_id"
              :update="update"
              :research-slug="researchSlug"
              :section-name="sectionName(update.section_id)"
            />
          </div>
        </div>
        <button v-if="visibleNew.length < newEntries.length" type="button" class="btn btn-sm updates-more" @click="newLimit += PAGE_SIZE">
          Show {{ Math.min(PAGE_SIZE, newEntries.length - visibleNew.length) }} more
        </button>
      </section>

      <section v-if="changedEntries.length" class="updates-group" aria-labelledby="changed-updates-title">
        <h2 id="changed-updates-title" class="card-section-title">
          Changed <span class="count-chip">{{ changedEntries.length }}</span>
        </h2>
        <div class="card card--list">
          <div class="data-rows">
            <ResearchUpdatesRow
              v-for="update in visibleChanged"
              :key="update.entry_id"
              :update="update"
              :research-slug="researchSlug"
              :section-name="sectionName(update.section_id)"
            />
          </div>
        </div>
        <button v-if="visibleChanged.length < changedEntries.length" type="button" class="btn btn-sm updates-more" @click="changedLimit += PAGE_SIZE">
          Show {{ Math.min(PAGE_SIZE, changedEntries.length - visibleChanged.length) }} more
        </button>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { EntryUpdate } from '~/composables/useEntryUpdates'
import { researchPath } from '~/composables/useResearchPaths'

const props = defineProps<{
  updates: EntryUpdate[]
  sections: any[]
  researchSlug: string
  loading?: boolean
  refreshing?: boolean
  error?: string | null
}>()

defineEmits<{ retry: [] }>()

const PAGE_SIZE = 50
const newLimit = ref(PAGE_SIZE)
const changedLimit = ref(PAGE_SIZE)
const newEntries = computed(() => props.updates.filter((update) => update.kind === 'new'))
const changedEntries = computed(() => props.updates.filter((update) => update.kind === 'changed'))
const visibleNew = computed(() => newEntries.value.slice(0, newLimit.value))
const visibleChanged = computed(() => changedEntries.value.slice(0, changedLimit.value))

function sectionName(id: string): string {
  const section = props.sections.find((candidate) => candidate.id === id)
  return section?.display_name || section?.name || 'Section'
}

</script>

<style scoped>
.updates-list--refreshing { color: var(--color-text-muted); }
.updates-skeletons { display: flex; flex-direction: column; gap: var(--space-3); }
.update-skeleton { height: 5.25rem; }
.updates-group + .updates-group { margin-top: var(--space-6); }
.updates-more { margin-top: var(--space-3); }
</style>
