<template>
  <div>
    <!-- All entries view -->
    <template v-if="mode === 'all'">
      <div class="section-header">
        <h2 class="section-title">All entries</h2>
      </div>

      <!-- Global tags with counters -->
      <div v-if="tags.length" class="tags-panel mb-4">
        <span
          v-for="tc in tags"
          :key="tc.tag"
          :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tc.tag)}`, { 'tag-active': activeTag === tc.tag }]"
          @click="activeTag = activeTag === tc.tag ? '' : tc.tag"
        >{{ tc.tag }}<span v-if="tc.count > 1" class="tag-count">{{ tc.count }}</span></span>
      </div>

      <!-- Loading -->
      <div v-if="loading">
        <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
      </div>

      <template v-else-if="filteredEntries.length">
        <!-- Group by section -->
        <template v-for="group in groupedEntries" :key="group.section.id">
          <h3 class="group-section-title">{{ group.section.display_name || group.section.name }}</h3>
          <div class="grid entries-grid mb-4">
            <NuxtLink
              v-for="entry in group.entries"
              :key="entry.id"
              :to="`/research/${researchSlug}/entry/${entry.code || entry.id}`"
              class="card entry-card"
            >
              <div class="entry-card-header">
                <div class="entry-title-row">
                  <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
                  <h3 class="card-title">{{ entry.title }}</h3>
                </div>
                <StatusBadge :status="entry.status" />
              </div>
              <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
              <div v-if="entry.tags?.length" class="entry-tags">
                <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
              </div>
            </NuxtLink>
          </div>
        </template>
      </template>

      <EmptyState
        v-else
        icon="&#x1F4C4;"
        title="No entries yet"
        description="Claude will populate this research with entries."
      />
    </template>

    <!-- Single section view -->
    <template v-else-if="sectionInfo">
      <div class="section-header">
        <h2 class="section-title">{{ sectionInfo.display_name || sectionInfo.name }}</h2>
        <StatusBadge :status="sectionInfo.status" />
      </div>
      <p v-if="sectionInfo.description" class="card-meta mb-4">
        {{ sectionInfo.description }}
      </p>

      <!-- Tag filter for entries -->
      <div v-if="sectionTags.length" class="tags-panel mb-4">
        <span
          v-for="tc in sectionTags"
          :key="tc.tag"
          :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tc.tag)}`, { 'tag-active': activeTag === tc.tag }]"
          @click="activeTag = activeTag === tc.tag ? '' : tc.tag"
        >{{ tc.tag }}<span v-if="tc.count > 1" class="tag-count">{{ tc.count }}</span></span>
      </div>

      <!-- Entries loading -->
      <div v-if="loading">
        <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
      </div>

      <div v-else-if="filteredEntries.length" class="grid entries-grid">
        <NuxtLink
          v-for="entry in filteredEntries"
          :key="entry.id"
          :to="`/research/${researchSlug}/entry/${entry.code || entry.id}`"
          class="card entry-card"
        >
          <div class="entry-card-header">
            <div class="entry-title-row">
              <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
              <h3 class="card-title">{{ entry.title }}</h3>
            </div>
            <StatusBadge :status="entry.status" />
          </div>
          <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
          <div v-if="entry.tags?.length" class="entry-tags">
            <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
          </div>
        </NuxtLink>
      </div>

      <EmptyState
        v-else
        icon="&#x1F4C4;"
        title="No entries yet"
        description="Claude will populate this section with research entries."
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { tagHue } from '~/composables/useTagHue'

const props = defineProps<{
  entries: any[]
  sections: any[]
  researchSlug: string
  loading: boolean
  mode: 'all' | 'section'
  sectionInfo?: any
  tags: Array<{ tag: string; count: number }>
}>()

const activeTag = ref('')

// Reset tag filter when mode or section changes
watch(() => [props.mode, props.sectionInfo?.id], () => {
  activeTag.value = ''
})

// Compute section-level tag counts from entries
const sectionTags = computed(() => {
  if (props.mode !== 'section') return []
  const map = new Map<string, number>()
  for (const e of props.entries) {
    for (const t of (e.tags ?? [])) {
      map.set(t, (map.get(t) || 0) + 1)
    }
  }
  return [...map.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([tag, count]) => ({ tag, count }))
})

const filteredEntries = computed(() =>
  activeTag.value ? props.entries.filter((e: any) => e.tags?.includes(activeTag.value)) : props.entries
)

// Group entries by section for the all-entries view
const groupedEntries = computed(() => {
  if (props.mode !== 'all') return []
  const groups: { section: any; entries: any[] }[] = []
  const sectionMap = new Map<string, any[]>()

  for (const entry of filteredEntries.value) {
    const list = sectionMap.get(entry.section_id) ?? []
    list.push(entry)
    sectionMap.set(entry.section_id, list)
  }

  for (const section of props.sections) {
    const sectionEntries = sectionMap.get(section.id)
    if (sectionEntries?.length) {
      groups.push({ section, entries: sectionEntries })
    }
  }

  return groups
})
</script>

<style scoped>
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-4); }
.section-title { font-size: var(--type-xl); font-weight: 600; letter-spacing: -0.02em; }
.tags-panel { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.tag-active { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-count {
  font-size: 0.75em;
  opacity: 0.7;
  margin-left: 0.15em;
}
.group-section-title {
  font-size: var(--type-base);
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}
.entries-grid { grid-template-columns: 1fr; }
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
.entry-tags { display: flex; gap: var(--space-2); flex-wrap: wrap; margin-top: var(--space-3); }
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
</style>
