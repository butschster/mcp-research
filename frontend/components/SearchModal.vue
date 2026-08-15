<template>
  <button class="search-trigger" @click="open = true">
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
    <span class="search-placeholder">Search...</span>
    <kbd class="search-shortcut">&#x2318;K</kbd>
  </button>

  <Teleport to="body">
    <div v-if="open" class="search-overlay" @click.self="open = false">
      <div class="search-modal">
        <div class="search-input-wrap">
          <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
          <input
            ref="inputRef"
            v-model="query"
            class="search-input"
            placeholder="Search projects, entries, tags..."
            @keydown.escape="open = false"
            @keydown.up.prevent="moveUp"
            @keydown.down.prevent="moveDown"
            @keydown.enter="selectCurrent"
          />
          <kbd class="search-kbd">Esc</kbd>
        </div>

        <div class="search-results" v-if="query.length > 1">
          <!-- Research results -->
          <div v-if="researchResults.length" class="result-group">
            <div class="result-group-label">Projects</div>
            <NuxtLink
              v-for="(r, i) in researchResults"
              :key="'r-' + r.id"
              :to="`/research/${r.code || r.id}`"
              :class="['result-item', { 'result-active': cursor === i }]"
              @click="open = false"
              @mouseenter="cursor = i"
            >
              <div class="result-content">
                <span class="result-title" v-html="highlight(r.name, query)"></span>
                <span class="result-meta">{{ r.goal }}</span>
              </div>
              <StatusBadge :status="r.status" />
            </NuxtLink>
          </div>

          <!-- Entry results -->
          <div v-if="entryResults.length" class="result-group">
            <div class="result-group-label">Entries</div>
            <NuxtLink
              v-for="(e, i) in entryResults"
              :key="'e-' + e.id"
              :to="entryLink(e)"
              :class="['result-item', { 'result-active': cursor === researchResults.length + i }]"
              @click="open = false"
              @mouseenter="cursor = researchResults.length + i"
            >
              <span v-if="e.code" class="result-code">{{ e.code }}</span>
              <div class="result-content">
                <span class="result-title" v-html="highlight(e.title, query)"></span>
                <span class="result-meta" v-html="highlight(e.description || '', query)"></span>
              </div>
              <div v-if="e.tags?.length" class="result-tags">
                <span v-for="tag in e.tags.slice(0, 3)" :key="tag" class="result-tag">{{ tag }}</span>
              </div>
            </NuxtLink>
          </div>

          <div v-if="!researchResults.length && !entryResults.length" class="result-empty">
            No results for "{{ query }}"
          </div>
        </div>

        <div v-else class="search-hints">
          <span class="hint-item"><kbd>&uarr;&darr;</kbd> navigate</span>
          <span class="hint-item"><kbd>&crarr;</kbd> open</span>
          <span class="hint-item"><kbd>Esc</kbd> close</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { escapeHtml } from '~/utils/escapeHtml'
const open = ref(false)
const query = ref('')
const cursor = ref(0)
const inputRef = ref<HTMLInputElement>()

onMounted(() => {
  window.addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      open.value = true
    }
  })
})

watch(open, (val) => {
  if (val) {
    query.value = ''
    cursor.value = 0
    nextTick(() => inputRef.value?.focus())
  }
})

// Research search (client-side)
const { data } = useApi<{ data: any[] }>('/api/researches')
const researches = computed(() => data.value?.data ?? [])

const researchResults = computed(() => {
  if (query.value.length < 2) return []
  const q = query.value.toLowerCase()
  return researches.value.filter((r: any) =>
    r.name.toLowerCase().includes(q) ||
    r.goal?.toLowerCase().includes(q) ||
    r.tags?.some((t: string) => t.toLowerCase().includes(q))
  ).slice(0, 5)
})

// Entry search (server-side full-text)
const searchQuery = computed(() => query.value.length >= 2 ? query.value : '')
const { data: searchData } = useApi<{ entries: any[] }>(
  computed(() => searchQuery.value ? `/api/search?q=${encodeURIComponent(searchQuery.value)}` : '/api/search?q=')
)
const entryResults = computed(() => (searchData.value?.entries ?? []).slice(0, 10))

// Research code lookup for entry links
const researchCodeMap = computed(() => {
  const map = new Map<string, string>()
  for (const r of researches.value) {
    map.set(r.id, r.code || r.id)
  }
  return map
})

function entryLink(e: any): string {
  const rSlug = researchCodeMap.value.get(e.research_id) || e.research_id
  return `/research/${rSlug}/entry/${e.code || e.id}`
}

// Total results for keyboard navigation
const totalResults = computed(() => researchResults.value.length + entryResults.value.length)

/**
 * Wraps the matched run in a `<mark>`, for v-html.
 *
 * Both arguments are attacker-controlled — the text is a research name or an
 * entry title somebody's agent wrote, and the query is whatever was typed — so
 * both are escaped before any markup is introduced. The regex is built from the
 * escaped query, because escaping afterwards would eat the `<mark>` too.
 */
function highlight(text: string, q: string): string {
  if (!text) return ''
  if (!q) return escapeHtml(text)
  // Match against the raw text and escape each piece as it is assembled.
  // Escaping first and matching afterwards looked simpler and was wrong: the
  // escape introduces entity names that were not in the source, so a search for
  // `lt` matched inside the `&lt;` it had just produced and split it — a title
  // reading `a < b` came back as the literal text `a &lt; b`.
  const re = new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
  let out = ''
  let last = 0
  for (const m of text.matchAll(re)) {
    if (!m[0]) break
    out += escapeHtml(text.slice(last, m.index))
      + `<mark class="search-mark">${escapeHtml(m[0])}</mark>`
    last = m.index + m[0].length
  }
  return out + escapeHtml(text.slice(last))
}

function moveUp() { cursor.value = Math.max(0, cursor.value - 1) }
function moveDown() { cursor.value = Math.min(totalResults.value - 1, cursor.value + 1) }
function selectCurrent() {
  const rLen = researchResults.value.length
  if (cursor.value < rLen) {
    const item = researchResults.value[cursor.value]
    if (item) navigateTo(`/research/${item.code || item.id}`)
  } else {
    const item = entryResults.value[cursor.value - rLen]
    if (item) navigateTo(entryLink(item))
  }
  open.value = false
}
</script>

<style scoped>
.search-trigger {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  font-family: inherit;
  font-size: var(--type-sm);
  min-width: 200px;
  transition: all var(--transition-fast);
}
.search-trigger:hover {
  border-color: var(--color-border-strong);
  color: var(--color-text);
}
.search-placeholder {
  flex: 1;
  text-align: left;
}
.search-shortcut {
  font-size: 0.625rem;
  padding: 1px 5px;
  background: var(--color-surface-hover);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  color: var(--color-text-muted);
  font-family: inherit;
  line-height: 1.4;
}

.search-overlay {
  position: fixed; inset: 0; z-index: var(--z-overlay);
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(8px) saturate(0.8);
  -webkit-backdrop-filter: blur(8px) saturate(0.8);
  display: flex; align-items: flex-start; justify-content: center;
  padding-top: 14vh;
  animation: overlay-in 0.15s ease both;
}
@keyframes overlay-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
.search-modal {
  width: 100%; max-width: 600px; margin: 0 var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-lg); overflow: hidden;
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(108, 197, 224, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.04);
  animation: modal-in 0.2s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
@keyframes modal-in {
  from { opacity: 0; transform: scale(0.96) translateY(-8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
.search-input-wrap {
  display: flex; align-items: center; gap: var(--space-3);
  padding: var(--space-4) var(--space-5); border-bottom: 1px solid var(--color-border);
}
.search-icon { color: var(--color-text-muted); flex-shrink: 0; }
.search-input {
  flex: 1; background: none; border: none; outline: none;
  color: var(--color-text); font-size: var(--type-base); font-family: inherit;
}
.search-input::placeholder { color: var(--color-text-muted); }
.search-kbd {
  font-size: 0.625rem; padding: var(--space-1) var(--space-2);
  background: var(--color-surface-hover); border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); color: var(--color-text-muted);
  font-family: inherit;
}
.search-results { max-height: 420px; overflow-y: auto; padding: var(--space-2) 0; }
.result-group + .result-group { border-top: 1px solid var(--color-border); }
.result-group-label {
  font-size: var(--type-xs); font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.04em; color: var(--color-text-muted);
  padding: var(--space-3) var(--space-5) var(--space-1);
}
.result-item {
  display: flex; align-items: center; gap: var(--space-3);
  padding: var(--space-2) var(--space-4); text-decoration: none; color: inherit;
  transition: background var(--transition-fast); cursor: pointer;
  margin: 0 var(--space-2);
  border-radius: var(--radius-sm);
}
.result-item:hover, .result-active { background: var(--color-surface-hover); }
.result-code {
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  flex-shrink: 0;
}
.result-content { flex: 1; min-width: 0; }
.result-title { display: block; font-weight: 500; font-size: var(--type-sm); }
.result-meta { display: block; font-size: var(--type-xs); color: var(--color-text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-top: 1px; }
.result-tags { display: flex; gap: 3px; flex-shrink: 0; }
.result-tag {
  font-size: 0.6rem;
  padding: 1px 4px;
  background: var(--color-surface-hover);
  border-radius: 2px;
  color: var(--color-text-muted);
}
.result-empty { padding: var(--space-10); text-align: center; color: var(--color-text-muted); font-size: var(--type-sm); }
.search-hints { display: flex; gap: var(--space-6); padding: var(--space-3) var(--space-5); border-top: 1px solid var(--color-border); }
.hint-item { font-size: var(--type-xs); color: var(--color-text-muted); display: flex; align-items: center; gap: var(--space-2); }
.hint-item kbd { background: var(--color-surface-hover); border: 1px solid var(--color-border); border-radius: var(--radius-sm); padding: var(--space-1) var(--space-2); font-size: 0.625rem; font-family: inherit; }

/* Responsive */
@media (max-width: 768px) {
  .search-trigger { min-width: 0; padding: var(--space-2); }
  .search-placeholder { display: none; }
  .search-shortcut { display: none; }
  .search-overlay { padding-top: var(--space-4); }
  .search-modal { margin: 0 var(--space-3); max-height: calc(100dvh - var(--space-8)); }
  .search-results { max-height: 60dvh; }
  .search-hints { flex-wrap: wrap; gap: var(--space-3); }
}
</style>
