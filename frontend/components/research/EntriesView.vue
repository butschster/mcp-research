<template>
  <div>
    <!-- Searching inside one research.
         The command palette has always been global, and tags were the only
         in-research filter — which exist only if the agent applied them. In a
         sixty-entry research that left nothing to search with, on the surface
         where finding one paragraph is most of the job. -->
    <div class="entries-search cluster">
      <input
        v-model="query"
        type="search"
        class="text-input entries-search-input"
        placeholder="Search this research…"
        :aria-label="`Search ${researchName || 'this research'}`"
      />
      <span v-if="query.length === 1" class="card-meta">Keep typing…</span>
      <span v-else-if="query.length > 1" class="card-meta" role="status">
        {{ searching ? 'Searching…' : `${found.length} ${found.length === 1 ? 'match' : 'matches'}` }}
      </span>
    </div>

    <!-- Results replace the list while a query stands. -->
    <div v-if="query.length > 1">
      <div v-if="found.length" class="grid grid-2">
        <EntryCard
          v-for="entry in found"
          :key="entry.id"
          :entry="entry"
          :research-slug="researchSlug"
          :missing-required="entry.metadata_status?.missing_required?.length ?? 0"
        />
      </div>
      <EmptyState
        v-else-if="!searching"
        icon="&#x1F50D;"
        :title="`Nothing matches “${query}”`"
        description="Titles, descriptions and the body of every entry in this research were searched."
      />
    </div>

    <template v-else>
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
              :to="entryPath(researchSlug, entry.code || entry.id)"
              class="card entry-card"
            >
              <div class="entry-card-header">
                <div class="entry-title-row">
                  <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
                  <span v-if="entry.entry_type === 'artifact'" class="entry-artifact-badge" title="HTML artifact">artifact</span>
                  <h3 class="card-title">{{ entry.title }}</h3>
                </div>
                <div class="entry-card-flags">
                  <span
                    v-if="missingIn(group.section, entry)"
                    class="badge badge-draft"
                    :title="`${missingIn(group.section, entry)} required ${missingIn(group.section, entry) === 1 ? 'field is' : 'fields are'} unanswered`"
                  >{{ missingIn(group.section, entry) }} missing</span>
                  <StatusBadge :status="entry.status" />
                </div>
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
        :description="emptyDescription('research')"
      />
    </template>

    <!-- Single section view -->
    <template v-else-if="sectionInfo">
      <div class="section-header">
        <!-- Grouped, because `space-between` over three loose children put the
             status badge 292px from the title it describes on a section that
             declares no fields — there was nothing else to absorb the row. -->
        <div class="section-identity">
        <h2 class="section-title">{{ sectionInfo.display_name || sectionInfo.name }}</h2>
        <StatusBadge :status="sectionInfo.status" />
        <!-- Offered only where there is something to tabulate. A toggle over a
             table with one column would be a control that promises more than
             the section declares. -->
        <NuxtLink
          v-if="fieldSpec.length"
          :to="`/research/${researchSlug}/settings?tab=sections#fields-${sectionInfo.code || sectionInfo.id}`"
          class="section-fields-link"
        >{{ fieldSpec.length }} fields</NuxtLink>
        <SegmentedToggle
          v-if="fieldSpec.length"
          :model-value="view"
          :options="viewOptions"
          label="Section view"
          class="section-view-toggle"
          @update:model-value="view = $event as 'cards' | 'table'"
        />
        </div>
        <!-- Removed for a viewer rather than disabled: the house rule, and the
             drop handler still runs for them so a dropped file cannot navigate
             the browser away from the app. -->
        <button v-if="canImport" type="button" class="btn btn-sm section-import" @click="dropZone?.open()">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 15V3M7 8l5-5 5 5M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2"/></svg>
          Import .md
        </button>
      </div>
      <p v-if="sectionInfo.description" class="card-meta mb-4">
        {{ sectionInfo.description }}
      </p>

      <p v-if="importer.refusal.value" class="inline-error import-refusal">
        <span>{{ importer.refusal.value.message }}</span>
        <button
          v-if="importer.refusal.value.retryable && importer.file.value"
          type="button"
          class="btn btn-sm"
          @click="importer.reread()"
        >Try again</button>
        <button type="button" class="btn btn-sm" @click="importer.dismissRefusal()">Dismiss</button>
      </p>
      <!-- Mounted always, text swapped — a `role` on an element that does not
           exist until it should speak is announced unreliably. Same pattern as
           ToastHost. -->
      <p class="sr-only" role="status">{{ importAnnouncement }}</p>

      <ResearchImportDropZone
        ref="dropZone"
        :enabled="canImport"
        :section-name="sectionInfo.display_name || sectionInfo.name"
        :max-bytes="maxBytes"
        :extensions="extensions"
        :reading="importer.phase.value === 'reading'"
        @file="importer.read"
        @refuse="onRefuse"
      >

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

      <!-- The table is the point of declaring fields at all: a blank cell beside
           filled ones is the strongest force there is on whether an optional
           field ever gets answered. -->
      <!-- Gated on the declaration, not only on the toggle: the component
           instance is reused across sections, so a section with no fields
           inherited the table and had no control to leave it with. -->
      <div v-else-if="view === 'table' && fieldSpec.length && filteredEntries.length" class="meta-table-wrap">
        <table class="meta-table">
          <thead>
            <tr>
              <th scope="col">Document</th>
              <th scope="col">Status</th>
              <th v-for="f in fieldSpec" :key="f.key" scope="col" :title="f.help || undefined">
                {{ fieldLabel(f) }}<span v-if="f.required" class="meta-req" aria-label="required">*</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in filteredEntries" :key="entry.id">
              <th scope="row" class="meta-table-doc">
                <NuxtLink :to="entryPath(researchSlug, entry.code || entry.id)">
                  <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
                  {{ entry.title }}
                </NuxtLink>
              </th>
              <td><StatusBadge :status="entry.status" /></td>
              <!-- A blank required cell and a blank optional one used to read
                   the same, on the one screen built to make gaps visible. So
                   did a value outside its vocabulary and a legitimate one. -->
              <td
                v-for="f in fieldSpec"
                :key="f.key"
                :class="{
                  'is-empty': !isFilled(entry.metadata?.[f.key]),
                  'is-missing': f.required && !isFilled(entry.metadata?.[f.key]),
                  'is-invalid': hasIssue(entry, f.key),
                }"
                :title="issueReason(entry, f.key) || undefined"
              >
                <template v-if="isUnknown(entry.metadata?.[f.key])">unknown</template>
                <template v-else-if="isFilled(entry.metadata?.[f.key])">
                  <!-- The same renderer the document's own block uses. Printing
                       the raw string here would show `[[E47]]` in the table and
                       a live link on the page, for one value. -->
                  <span v-html="renderRefs(displayValue(f, entry.metadata?.[f.key]), researchSlug)"></span>
                </template>
                <template v-else>&mdash;</template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else-if="filteredEntries.length" class="grid entries-grid">
        <NuxtLink
          v-for="entry in filteredEntries"
          :key="entry.id"
          :to="entryPath(researchSlug, entry.code || entry.id)"
          class="card entry-card"
        >
          <div class="entry-card-header">
            <div class="entry-title-row">
              <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
              <span v-if="entry.entry_type === 'artifact'" class="entry-artifact-badge" title="HTML artifact">artifact</span>
              <h3 class="card-title">{{ entry.title }}</h3>
            </div>
            <div class="entry-card-flags">
              <span
                v-if="missingCount(entry)"
                class="badge badge-draft"
                :title="`${missingCount(entry)} required ${missingCount(entry) === 1 ? 'field is' : 'fields are'} unanswered`"
              >{{ missingCount(entry) }} missing</span>
              <StatusBadge :status="entry.status" />
            </div>
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
        :title="shareActive() ? 'No entries in this section' : 'No entries yet'"
        :description="emptyDescription('section')"
      >
        <button v-if="canImport" type="button" class="btn btn-sm" @click="dropZone?.open()">
          Import .md
        </button>
      </EmptyState>
      </ResearchImportDropZone>


      <ResearchImportPreviewDialog
        :visible="importer.phase.value === 'previewing' || importer.phase.value === 'committing'"
        :file-bytes="importer.file.value?.size ?? 0"
        :section-name="sectionInfo.display_name || sectionInfo.name"
        :preview="importer.preview.value"
        :committing="importer.phase.value === 'committing'"
        :error="importer.commitError.value"
        :stale-spec="importer.staleSpec.value"
        @commit="onCommit"
        @reread="importer.reread"
        @close="importer.reset"
      />
    </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { tagHue } from '~/composables/useTagHue'
import { renderRefs } from '~/composables/useCrossRefs'
import { entryPath } from '~/composables/useResearchPaths'
import { shareActive } from '~/composables/useShare'
import { displayValue, fieldLabel, isFilled, isUnknown, missingRequired } from '~/composables/useFieldSpec'
import { useSectionImport } from '~/composables/useSectionImport'
import { useMetadataSchema } from '~/composables/useMetadataSchema'

const props = defineProps<{
  entries: any[]
  researchId?: string
  researchName?: string
  sections: any[]
  researchSlug: string
  loading: boolean
  mode: 'all' | 'section'
  sectionInfo?: any
  tags: Array<{ tag: string; count: number }>
}>()

/* What this section declares, and the two things every surface needs from it.
   The predicate lives in the composable rather than here, because a document
   reading "complete" in the table and "1 missing" on its own page is the same
   class of bug as a stored status disagreeing with the prose. */
const fieldSpec = computed<any[]>(() => props.sectionInfo?.field_spec ?? [])
const view = ref<'cards' | 'table'>('cards')

/* Dropping a markdown file into this section.
   The state machine is in the composable: this component already owns the
   search, the tag filter, the view toggle and the grouping. */
const emit = defineEmits<{ imported: [{ id: string; code: string }] }>()

const dropZone = ref<{ open: () => void } | null>(null)
const sectionId = computed(() => props.sectionInfo?.id ?? '')
const importer = useSectionImport(sectionId)
const { maxBytes, extensions, load: loadCaps } = useMetadataSchema()
const { canWrite } = useResearchRole()
const toasts = useToasts()

// Never on a share, never for a viewer, and never in the all-entries mode where
// there is no one section to drop onto.
const canImport = computed(() => props.mode === 'section' && !shareActive() && canWrite.value)

onMounted(() => {
  if (canImport.value) loadCaps()
})

// A file half-read into one section must not survive a move to another.
watch(sectionId, () => importer.reset())

// The section's declaration can change under an open dialog: the page reloads
// the research on any `section` event, and what the preview was validated
// against is then no longer what the commit will be checked by. Watching the
// spec here rather than being told by the page keeps the knowledge where the
// dialog is.
watch(
  () => JSON.stringify(fieldSpec.value),
  (next, prev) => {
    if (next !== prev) importer.markSpecStale()
  },
)

function onRefuse(r: { message: string; retryable: boolean }) {
  importer.refusal.value = r
}

// One live region, mounted for the life of the pane, whose text changes.
const importAnnouncement = computed(() => {
  if (importer.refusal.value) return importer.refusal.value.message
  if (importer.phase.value === 'reading') return `Reading ${importer.file.value?.name ?? 'the file'}…`
  if (importer.phase.value === 'previewing') return 'The file was read. Review it before creating the document.'
  return ''
})

async function onCommit(overrides: { title: string }) {
  const created = await importer.commit(overrides)
  if (!created) return
  toasts.push({
    variant: 'success',
    message: `${created.code} created from the file.`,
    action: {
      label: `Open ${created.code}`,
      onClick: () => {
        navigateTo(entryPath(props.researchSlug, created.code || created.id))
      },
    },
  })
  // The entries list and the sidebar counts belong to the page.
  // Deliberately staying in the section rather than navigating to the entry:
  // multi-file import is out of scope, so importing five documents is five
  // rounds of this, and leaving the page each time makes that loop miserable.
  emit('imported', created)
}

const viewOptions = [
  { value: 'cards', label: 'Cards' },
  { value: 'table', label: 'Table' },
]

function missingCount(entry: any): number {
  return missingIn(props.sectionInfo, entry)
}

/* The all-entries list groups by section, so it holds the declaration for each
   group and can show the same gap. Leaving it out would have hidden the signal
   on the surface people actually browse a research from. */
function hasIssue(entry: any, key: string): boolean {
  return !!issueReason(entry, key)
}

/* Recomputed on the server and carried on the list payload, not derived here:
   enum membership, date parsing and URL schemes are the server's judgement, and
   a second copy in TypeScript is how the two come to disagree. */
function issueReason(entry: any, key: string): string {
  return entry?.metadata_status?.issues?.find((i: any) => i.key === key)?.reason ?? ''
}

function missingIn(section: any, entry: any): number {
  const specs = section?.field_spec ?? []
  if (!specs.length) return 0
  return missingRequired(specs, entry.metadata ?? {}).length
}

/* The search is the component's own: it holds the query, so nothing above it
   has to thread state through, and it asks the server rather than filtering
   `entries`, because the body is where the answer usually is and the body is
   not in the payload this component is handed. */
const query = ref('')
const searching = ref(false)
const found = ref<any[]>([])
const { authFetch } = useAuth()
const base = useRuntimeConfig().public.apiBase || ''

let seq = 0
let timer: ReturnType<typeof setTimeout> | null = null

async function run(q: string) {
  const mine = ++seq
  searching.value = true
  try {
    const res = await authFetch<{ entries: any[] }>(
      `${base}/api/search?q=${encodeURIComponent(q)}&research=${encodeURIComponent(props.researchId ?? '')}`,
    )
    // An earlier request that lands late must not overwrite a later one.
    if (mine === seq) found.value = res?.entries ?? []
  } catch {
    if (mine === seq) found.value = []
  } finally {
    if (mine === seq) searching.value = false
  }
}

/* Debounced by hand rather than by a dependency: a request per keystroke is
   what the global palette still does, and typing "authentication" there fires
   fourteen. */
watch(query, (q) => {
  if (timer) clearTimeout(timer)
  if (q.trim().length < 2) { found.value = []; searching.value = false; return }
  timer = setTimeout(() => run(q.trim()), 200)
})

onUnmounted(() => { if (timer) clearTimeout(timer) })

/**
 * What to say about an empty list.
 *
 * "Claude will populate this section with research entries" is true and useful
 * to the person running the research, and meaningless to a client reading a
 * shared link — it describes a tool they have never heard of doing work they
 * have no part in.
 */
function emptyDescription(scope: 'research' | 'section') {
  if (shareActive()) return "There's nothing here yet."
  if (scope === 'research') return 'Claude will populate this research with entries.'
  return canImport.value
    ? 'Claude will populate this section with research entries. You can also drop a markdown file here.'
    : 'Claude will populate this section with research entries.'
}

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
.entries-search { margin-bottom: var(--space-4); }
.entries-search-input { flex: 1; min-width: 12rem; }

/* Wrapping, because a fourth control overflows the row at the widths this pane
   actually gets. */
.section-header { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: var(--space-3); margin-bottom: var(--space-4); }
/* `.btn-sm` already declares `min-height: var(--control-h-sm)`, so the row
   cannot come out at three heights; this only adds the icon layout. */
.section-import { display: inline-flex; align-items: center; gap: var(--space-2); }

.section-identity { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-3); }
/* Above the list, beside the control that caused it. Below it was two screens
   away on a section of any size. */
.import-refusal {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.import-refusal span { min-width: 0; overflow-wrap: anywhere; }
.import-refusal .btn { flex: none; }
.section-title { font-size: var(--type-xl); font-weight: var(--weight-semibold); letter-spacing: -0.02em; }
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
  font-weight: var(--weight-semibold);
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
  font-weight: var(--weight-semibold);
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: var(--radius-xs);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
.entry-artifact-badge {
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-sm);
  background: var(--color-primary-muted);
  color: var(--color-primary);
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

/* No auto margins here. `.section-header` is space-between with no gap, and an
   auto margin is resolved first — it absorbed all the free space, so the title
   and the status badge packed flush together and a section that declares fields
   looked broken next to one that does not. */
.section-header { gap: var(--space-3); }
.section-view-toggle { margin-left: auto; }
.section-fields-link {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  white-space: nowrap;
}
.section-fields-link:hover { color: var(--color-primary); }

.meta-table-wrap {
  /* Without this the wrapper takes its min-content width from the table and
     blows out the `1fr` grid track it sits in, which scrolls the page body
     rather than the table. */
  min-width: 0;
  max-width: 100%;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  /* The page body must never scroll sideways; a wide table scrolls inside its
     own frame instead. */
  overflow-x: auto;
}

.meta-table { width: 100%; border-collapse: collapse; font-size: var(--type-sm); }
.meta-table th,
.meta-table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  vertical-align: top;
  white-space: nowrap;
  border-bottom: 1px solid var(--color-border);
}
.meta-table thead th {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text-muted);
}
.meta-table tbody tr:last-child th,
.meta-table tbody tr:last-child td { border-bottom: 0; }

.meta-table-doc { font-weight: var(--weight-normal); max-width: 22rem; white-space: normal; }
.meta-table-doc a { color: inherit; text-decoration: none; }
.meta-table-doc a:hover { text-decoration: underline; }

.meta-table td.is-empty { color: var(--color-text-faint); }
/* A required field nobody answered, and a value that is not one of the declared
   options. Both carry a word or a mark, never colour alone. */
.meta-table td.is-missing { color: var(--color-warning); }
.meta-table td.is-invalid { color: var(--color-danger); text-decoration: underline dotted; }
.meta-req { color: var(--color-warning); margin-left: 0.15em; }
/* The chip and the status badge are one group, or the header's space-between
   distributes free space between them and the chip floats mid-row. */
.entry-card-flags { display: flex; align-items: center; gap: var(--space-2); flex-shrink: 0; }

</style>
