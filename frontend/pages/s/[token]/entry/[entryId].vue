<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <EmptyState
    v-else-if="!entry && gone"
    title="This entry was removed"
    description="It was deleted while you were reading. The rest of the research is still here."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <EmptyState
    v-else-if="!entry"
    title="Couldn't load this entry"
    description="The server didn't answer. The rest of the research is still here — try again in a moment."
  >
    <button class="btn btn-primary" @click="loadEntry()">Try again</button>
  </EmptyState>

  <div v-else>
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: researchName, to: researchPath(slug) },
        { label: sectionName, to: `${researchPath(slug)}?section=${entry.section_id}` },
        { label: entry.title },
      ]" />

      <PageHeader :code="entry.code" :title="entry.title">
        <template #actions>
          <StatusBadge :status="entry.status" />
        </template>
      </PageHeader>

      <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, slug)"></p>
      <div v-if="entry.tags?.length" class="entry-tags">
        <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
      </div>
    </div>

    <div class="entry-content card">
      <BlocksBlockRenderer
        v-if="isBlocks"
        :blocks="blocks"
        :research-slug="slug"
        :bridge-data="null"
        :entry-id="entry.id"
      />
      <div v-else ref="contentEl" class="markdown-content" v-html="renderedContent"></div>
    </div>

    <EntryCrossReferencesBlock
      :outgoing="outgoingRefs"
      :incoming="incomingRefs"
      :research-slug="slug"
    />

    <EntryExternalLinksBlock :links="externalLinks" />

    <EntryRelatedEntriesBlock
      :entries="relatedEntries"
      :current-tags="entry.tags ?? []"
      :research-slug="slug"
      :research-id="researchId"
    />

    <EntryNavigation
      v-if="siblings.length > 1"
      :prev="prevEntry"
      :next="nextEntry"
      :research-slug="slug"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * One entry, as a visitor sees it.
 *
 * This is a separate page from the owner's rather than a mode of it, and that
 * is the first of the two read-only defences: the editor, the delete
 * confirmation, the status picker and the revision history are not imported
 * here, so no flag can make them appear. Revision history is left out on
 * purpose — who edited what and when is internal working process, of a piece
 * with `instruction` and `memory`.
 */
import { parseMarkdown } from '~/composables/useSafeMarkdown'
import { tagHue } from '~/composables/useTagHue'

const route = useRoute()
const entryCode = computed(() => route.params.entryId as string)

const { shareFetch, research, sections, researchId, researchCode, slug } = useShare()
const researchName = computed(() => research.value?.name || 'Research')

const entry = ref<any | null>(null)
const pending = ref(true)
/** The server said it is not there, as opposed to not answering at all. */
const gone = ref(false)
const outgoingRefs = ref<any[]>([])
const incomingRefs = ref<any[]>([])
const externalLinks = ref<any[]>([])
const relatedEntries = ref<any[]>([])
const siblings = ref<any[]>([])

const sectionName = computed(() => {
  const s = sections.value.find((x: any) => x.id === entry.value?.section_id)
  return s?.display_name || s?.name || 'Section'
})

async function loadEntry() {
  pending.value = true
  // The previous entry's siblings are not this entry's siblings; leaving them
  // in place renders a navigation bar belonging to the section just left.
  siblings.value = []
  try {
    const res = await shareFetch<{ data: any }>(
      `/researches/${researchId.value}/entries/${entryCode.value}`,
    )
    entry.value = res.data
    gone.value = false
  } catch (e: any) {
    const status = e?.response?.status ?? e?.statusCode
    gone.value = status === 404
    entry.value = null
    pending.value = false
    return
  }
  pending.value = false
  await loadContext()
}

/**
 * The blocks around the entry — references, links, siblings.
 *
 * Each is allowed to fail on its own: a missing "related by tags" is a missing
 * panel, and it must not take the entry down with it.
 */
async function loadContext() {
  const id = entry.value?.id
  if (!id) return
  const [refs, links, related, sectionEntries] = await Promise.allSettled([
    shareFetch<{ outgoing: any[]; incoming: any[] }>(`/entries/${id}/crossrefs`),
    shareFetch<{ data: any[] }>(`/entries/${id}/links`),
    shareFetch<{ data: any[] }>(`/entries/${id}/related`),
    shareFetch<{ data: any[] }>(
      `/researches/${researchId.value}/sections/${entry.value.section_id}/entries`,
    ),
  ])
  outgoingRefs.value = refs.status === 'fulfilled' ? (refs.value.outgoing ?? []) : []
  incomingRefs.value = refs.status === 'fulfilled' ? (refs.value.incoming ?? []) : []
  externalLinks.value = links.status === 'fulfilled' ? (links.value.data ?? []) : []
  relatedEntries.value = related.status === 'fulfilled' ? (related.value.data ?? []) : []
  siblings.value = sectionEntries.status === 'fulfilled' ? (sectionEntries.value.data ?? []) : []
}

watch([entryCode, researchId], () => void loadEntry(), { immediate: true })

const currentIndex = computed(() => siblings.value.findIndex((e: any) => e.id === entry.value?.id))
const prevEntry = computed(() => (currentIndex.value > 0 ? siblings.value[currentIndex.value - 1] : null))
const nextEntry = computed(() =>
  currentIndex.value >= 0 && currentIndex.value < siblings.value.length - 1
    ? siblings.value[currentIndex.value + 1]
    : null,
)

const isBlocks = computed(() => entry.value?.entry_type === 'blocks' && blocks.value.length > 0)
const blocks = computed<any[]>(() => {
  if (entry.value?.entry_type !== 'blocks' || !entry.value?.content) return []
  try {
    const doc = JSON.parse(entry.value.content)
    return Array.isArray(doc) ? doc : (doc.blocks ?? [])
  } catch {
    return []
  }
})

const contentEl = ref<HTMLElement | null>(null)
const renderedContent = computed(() => {
  if (!entry.value?.content) return ''
  const html = parseMarkdown(String(entry.value.content).replace(/\\n/g, '\n')) as string
  return linkRefs(html, slug.value)
})

watch(renderedContent, () => {
  nextTick(() => { if (contentEl.value) renderMermaidBlocks(contentEl.value) })
})
onMounted(() => { if (contentEl.value) renderMermaidBlocks(contentEl.value) })

useResearchRealtime(
  () => slug.value,
  (event) => {
    if (event.entity === 'entry' || event.entity === 'crossref') void loadEntry()
  },
  { onResync: () => void loadEntry(), researchId: () => researchId.value },
)
</script>
