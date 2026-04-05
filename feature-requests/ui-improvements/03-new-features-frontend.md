# Track A — New Frontend Features (No Backend Changes)

**Scope:** Новые фичи только на фронтенде. Используют существующий API.  
**Effort:** 4-16 часов на фичу.

---

## NF-1: Client-side Search (без нового API)

**Проблема:** Нет поиска. При 10+ проектах — критический blocker.

**Подход:** Клиентский поиск по уже загруженным данным. Загружаем все researches + все entries при открытии поиска (lazy).

**Новый файл:** `frontend/components/SearchModal.vue`

```vue
<template>
  <!-- Триггер -->
  <button class="search-trigger" @click="open = true">
    <span class="search-icon">⌕</span>
    <span class="search-hint">Search... <kbd>⌘K</kbd></span>
  </button>

  <!-- Modal -->
  <Teleport to="body">
    <div v-if="open" class="search-overlay" @click.self="open = false">
      <div class="search-modal">
        <div class="search-input-wrap">
          <span class="search-icon-lg">⌕</span>
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
          <!-- Research matches -->
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
              <span class="result-icon">🔬</span>
              <div class="result-content">
                <span class="result-title" v-html="highlight(r.name, query)"></span>
                <span class="result-meta">{{ r.goal }}</span>
              </div>
              <StatusBadge :status="r.status" />
            </NuxtLink>
          </div>

          <!-- Empty -->
          <div v-if="!researchResults.length" class="result-empty">
            No results for "{{ query }}"
          </div>
        </div>

        <div v-else class="search-hints">
          <span class="hint-item"><kbd>↑↓</kbd> navigate</span>
          <span class="hint-item"><kbd>↵</kbd> open</span>
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

// Keyboard shortcut Cmd+K / Ctrl+K
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

// Поиск по загруженным researches
const { data } = useApi<{ data: any[] }>('/api/researches')
const researches = computed(() => data.value?.data ?? [])

const researchResults = computed(() => {
  if (query.value.length < 2) return []
  const q = query.value.toLowerCase()
  return researches.value.filter(r =>
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
```

```css
/* Глобально в main.css */
.search-mark { background: rgba(56,189,248,0.25); color: var(--color-primary); border-radius: 2px; }
```

**Интеграция:** Добавить `<SearchModal />` в `nav-right` в `app.vue`.

---

## NF-2: Getting Started Banner (Onboarding)

**Проблема:** Первый запуск — пустой экран. Нет объяснения что это и как начать.

**Новый файл:** `frontend/components/GettingStartedBanner.vue`

```vue
<template>
  <div v-if="show" class="getting-started">
    <button class="gs-close" @click="dismiss">×</button>
    <div class="gs-header">
      <span class="gs-icon">🔬</span>
      <div>
        <h2 class="gs-title">Welcome to MCP Research</h2>
        <p class="gs-subtitle">A read-only view of your AI-driven research sessions</p>
      </div>
    </div>
    <div class="gs-steps">
      <div class="gs-step">
        <span class="gs-step-num">1</span>
        <div class="gs-step-content">
          <strong>Add to Claude</strong>
          <p>Configure mcp-research as an MCP server in your Claude Desktop or Cursor settings</p>
        </div>
      </div>
      <div class="gs-step">
        <span class="gs-step-num">2</span>
        <div class="gs-step-content">
          <strong>Start a research session</strong>
          <p>Type in Claude:</p>
          <div class="gs-command">
            <code>Use the research/initialize prompt</code>
            <button class="copy-btn" @click="copyInit">{{ copiedInit ? '✓' : 'Copy' }}</button>
          </div>
        </div>
      </div>
      <div class="gs-step">
        <span class="gs-step-num">3</span>
        <div class="gs-step-content">
          <strong>Watch it unfold here</strong>
          <p>This UI updates in real-time as Claude populates your research</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// Используем localStorage (или query param) чтобы помнить dismissed state
// Показываем только если researches.length === 0
const props = defineProps<{ hasResearches: boolean }>()

const dismissed = ref(false)

onMounted(() => {
  dismissed.value = localStorage.getItem('gs-dismissed') === '1'
})

const show = computed(() => !props.hasResearches && !dismissed.value)

function dismiss() {
  dismissed.value = true
  localStorage.setItem('gs-dismissed', '1')
}

const copiedInit = ref(false)
function copyInit() {
  navigator.clipboard.writeText('Use the research/initialize prompt')
  copiedInit.value = true
  setTimeout(() => { copiedInit.value = false }, 2000)
}
</script>

<style scoped>
.getting-started {
  position: relative;
  background: linear-gradient(135deg, rgba(56,189,248,0.05), rgba(167,139,250,0.05));
  border: 1px solid rgba(56,189,248,0.2);
  border-radius: var(--radius);
  padding: 1.5rem;
  margin-bottom: 2rem;
}
.gs-close {
  position: absolute; top: 0.75rem; right: 0.75rem;
  background: none; border: none; color: var(--color-text-muted);
  font-size: 1.25rem; cursor: pointer; line-height: 1;
}
.gs-close:hover { color: var(--color-text); }
.gs-header { display: flex; align-items: flex-start; gap: 1rem; margin-bottom: 1.5rem; }
.gs-icon { font-size: 2rem; flex-shrink: 0; }
.gs-title { font-size: 1.25rem; font-weight: 700; margin-bottom: 0.25rem; }
.gs-subtitle { color: var(--color-text-muted); font-size: 0.875rem; }
.gs-steps { display: flex; flex-direction: column; gap: 1rem; }
.gs-step { display: flex; gap: 1rem; }
.gs-step-num {
  width: 1.75rem; height: 1.75rem; border-radius: 50%;
  background: rgba(56,189,248,0.15); color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.8125rem; font-weight: 700; flex-shrink: 0;
}
.gs-step-content strong { display: block; font-weight: 600; margin-bottom: 0.25rem; }
.gs-step-content p { font-size: 0.875rem; color: var(--color-text-muted); margin-bottom: 0.375rem; }
.gs-command {
  display: inline-flex; align-items: center; gap: 0.5rem;
  background: var(--color-surface-hover); border: 1px solid var(--color-border);
  border-radius: 6px; padding: 0.25rem 0.5rem; font-size: 0.8125rem;
}
.gs-command code { color: var(--color-primary); font-family: monospace; }
.copy-btn {
  padding: 0.125rem 0.375rem; background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: 4px; color: var(--color-text-muted); font-size: 0.75rem; cursor: pointer;
}
.copy-btn:hover { color: var(--color-text); }
</style>
```

---

## NF-3: Keyboard Navigation (глобальный)

**Новый файл:** `frontend/composables/useKeyboardNav.ts`

```typescript
export function useKeyboardNav() {
  onMounted(() => {
    window.addEventListener('keydown', handleKey)
  })
  onUnmounted(() => {
    window.removeEventListener('keydown', handleKey)
  })

  function handleKey(e: KeyboardEvent) {
    // Не перехватываем если фокус в input/textarea
    const tag = (e.target as HTMLElement).tagName
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(tag)) return

    const router = useRouter()
    const route = useRoute()

    switch (e.key) {
      case 'g':
        if (e.shiftKey) { e.preventDefault(); router.push('/') } // G — go home
        break
      case '?':
        e.preventDefault()
        // TODO: показать keyboard shortcuts modal
        break
      case 'r':
        if (e.metaKey || e.ctrlKey) break // не перехватываем Cmd+R
        e.preventDefault()
        window.location.reload()
        break
    }
  }
}
```

**Интеграция:** Вызвать `useKeyboardNav()` в `app.vue`.

---

## NF-4: Print/Export Enhancement

**Проблема:** Print CSS уже есть, но нет явной кнопки и нет PDF export guidance.

**Улучшение `frontend/components/PrintButton.vue`:**

```vue
<template>
  <div class="print-wrap">
    <button class="btn" @click="showMenu = !showMenu">
      ⎙ Export
    </button>
    <div v-if="showMenu" class="print-menu">
      <button class="print-item" @click="printPage">
        🖨 Print / Save as PDF
      </button>
      <button class="print-item" @click="copyMarkdown">
        📋 Copy as Markdown
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const showMenu = ref(false)
function printPage() { window.print(); showMenu.value = false }
// copyMarkdown — собирает все entry content в буфер
</script>
```
