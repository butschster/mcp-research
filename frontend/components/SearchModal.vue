<template>
  <button class="btn btn-sm search-trigger" @click="open = true">
    <span>&#x2315;</span>
    <span class="search-hint">Search <kbd>&#x2318;K</kbd></span>
  </button>

  <Teleport to="body">
    <div v-if="open" class="search-overlay" @click.self="open = false">
      <div class="search-modal">
        <div class="search-input-wrap">
          <span class="search-icon">&#x2315;</span>
          <input
            ref="inputRef"
            v-model="query"
            class="search-input"
            placeholder="Search projects, entries..."
            @keydown.escape="open = false"
            @keydown.up.prevent="moveUp"
            @keydown.down.prevent="moveDown"
            @keydown.enter="selectCurrent"
          />
          <kbd class="search-kbd">Esc</kbd>
        </div>

        <div class="search-results" v-if="query.length > 1">
          <div v-if="researchResults.length" class="result-group">
            <div class="result-group-label">Projects</div>
            <NuxtLink
              v-for="(r, i) in researchResults"
              :key="r.id"
              :to="`/research/${r.id}`"
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

          <div v-if="!researchResults.length" class="result-empty">
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

const { data } = useApi<{ data: any[] }>('/api/researches')
const researches = computed(() => data.value?.data ?? [])

const researchResults = computed(() => {
  if (query.value.length < 2) return []
  const q = query.value.toLowerCase()
  return researches.value.filter((r: any) =>
    r.name.toLowerCase().includes(q) ||
    r.goal?.toLowerCase().includes(q) ||
    r.tags?.some((t: string) => t.toLowerCase().includes(q))
  ).slice(0, 8)
})

function highlight(text: string, q: string): string {
  const re = new RegExp(`(${q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return text.replace(re, '<mark class="search-mark">$1</mark>')
}

function moveUp() { cursor.value = Math.max(0, cursor.value - 1) }
function moveDown() { cursor.value = Math.min(researchResults.value.length - 1, cursor.value + 1) }
function selectCurrent() {
  const item = researchResults.value[cursor.value]
  if (item) navigateTo(`/research/${item.id}`)
  open.value = false
}
</script>

<style scoped>
.search-trigger { gap: var(--space-2); }
.search-hint kbd { font-size: 0.625rem; opacity: 0.6; margin-left: var(--space-1); }

.search-overlay {
  position: fixed; inset: 0; z-index: 100;
  background: rgba(0,0,0,0.6); backdrop-filter: blur(4px);
  display: flex; align-items: flex-start; justify-content: center;
  padding-top: 12vh;
}
.search-modal {
  width: 100%; max-width: 560px; margin: 0 var(--space-4);
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius); overflow: hidden;
  box-shadow: 0 16px 48px rgba(0,0,0,0.4);
}
.search-input-wrap {
  display: flex; align-items: center; gap: var(--space-3);
  padding: var(--space-4); border-bottom: 1px solid var(--color-border);
}
.search-icon { font-size: var(--type-lg); color: var(--color-text-muted); }
.search-input {
  flex: 1; background: none; border: none; outline: none;
  color: var(--color-text); font-size: var(--type-base); font-family: inherit;
}
.search-input::placeholder { color: var(--color-text-muted); }
.search-kbd {
  font-size: 0.625rem; padding: var(--space-1) var(--space-2);
  background: var(--color-surface-hover); border: 1px solid var(--color-border);
  border-radius: var(--radius); color: var(--color-text-muted);
  font-family: inherit;
}
.search-results { max-height: 400px; overflow-y: auto; }
.result-group-label {
  font-size: var(--type-xs); font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--color-text-muted);
  padding: var(--space-3) var(--space-4) var(--space-1);
}
.result-item {
  display: flex; align-items: center; gap: var(--space-3);
  padding: var(--space-3) var(--space-4); text-decoration: none; color: inherit;
  transition: background var(--transition-fast); cursor: pointer;
}
.result-item:hover, .result-active { background: var(--color-surface-hover); }
.result-content { flex: 1; min-width: 0; }
.result-title { display: block; font-weight: 500; font-size: var(--type-sm); }
.result-meta { display: block; font-size: var(--type-xs); color: var(--color-text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.result-empty { padding: var(--space-8); text-align: center; color: var(--color-text-muted); font-size: var(--type-sm); }
.search-hints { display: flex; gap: var(--space-6); padding: var(--space-3) var(--space-4); }
.hint-item { font-size: var(--type-xs); color: var(--color-text-muted); display: flex; align-items: center; gap: var(--space-2); }
.hint-item kbd { background: var(--color-surface-hover); border: 1px solid var(--color-border); border-radius: var(--radius); padding: var(--space-1) var(--space-2); font-size: 0.625rem; font-family: inherit; }
</style>
