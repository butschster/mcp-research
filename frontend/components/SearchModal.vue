<template>
  <button class="search-trigger" @click="open = true">
    <span class="search-icon">&#x2315;</span>
    <span class="search-hint">Search... <kbd>&#x2318;K</kbd></span>
  </button>

  <Teleport to="body">
    <div v-if="open" class="search-overlay" @click.self="open = false">
      <div class="search-modal">
        <div class="search-input-wrap">
          <span class="search-icon-lg">&#x2315;</span>
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
          <kbd class="search-esc">Esc</kbd>
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
              <span class="result-icon">&#x1F52C;</span>
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
.search-trigger {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  background: var(--color-surface-hover); border: 1px solid var(--color-border);
  border-radius: var(--radius); color: var(--color-text-muted);
  font-size: 0.8125rem; cursor: pointer; transition: all 0.15s;
}
.search-trigger:hover { border-color: var(--color-primary); color: var(--color-text); }
.search-hint kbd { font-size: 0.6875rem; opacity: 0.7; }

.search-overlay {
  position: fixed; inset: 0; z-index: 100;
  background: rgba(0,0,0,0.6); backdrop-filter: blur(4px);
  display: flex; align-items: flex-start; justify-content: center;
  padding-top: 10vh;
}
.search-modal {
  width: 100%; max-width: 560px; margin: 0 1rem;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: 12px; overflow: hidden;
  box-shadow: 0 24px 80px rgba(0,0,0,0.5);
}
.search-input-wrap {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.875rem 1rem; border-bottom: 1px solid var(--color-border);
}
.search-icon-lg { font-size: 1.125rem; color: var(--color-text-muted); }
.search-input {
  flex: 1; background: none; border: none; outline: none;
  color: var(--color-text); font-size: 1rem; font-family: inherit;
}
.search-input::placeholder { color: var(--color-text-muted); }
.search-esc {
  font-size: 0.6875rem; padding: 0.25rem 0.5rem;
  background: var(--color-surface-hover); border: 1px solid var(--color-border);
  border-radius: 4px; color: var(--color-text-muted);
}
.search-results { max-height: 400px; overflow-y: auto; }
.result-group-label {
  font-size: 0.6875rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--color-text-muted);
  padding: 0.75rem 1rem 0.25rem;
}
.result-item {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.625rem 1rem; text-decoration: none; color: inherit;
  transition: background 0.1s; cursor: pointer;
}
.result-item:hover, .result-active { background: var(--color-surface-hover); }
.result-icon { font-size: 1rem; flex-shrink: 0; }
.result-content { flex: 1; min-width: 0; }
.result-title { display: block; font-weight: 500; font-size: 0.875rem; }
.result-meta { display: block; font-size: 0.75rem; color: var(--color-text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.result-empty { padding: 2rem; text-align: center; color: var(--color-text-muted); font-size: 0.875rem; }
.search-hints { display: flex; gap: 1.5rem; padding: 0.75rem 1rem; }
.hint-item { font-size: 0.75rem; color: var(--color-text-muted); display: flex; align-items: center; gap: 0.375rem; }
.hint-item kbd { background: var(--color-surface-hover); border: 1px solid var(--color-border); border-radius: 4px; padding: 0.125rem 0.375rem; font-size: 0.6875rem; }
</style>
