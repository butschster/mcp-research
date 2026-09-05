<template>
  <div v-if="show" class="getting-started">
    <button class="btn-ghost gs-close" aria-label="Dismiss getting started" @click="dismiss">&times;</button>
    <div class="gs-header">
      <div>
        <h2 class="gs-title">Welcome to Dovod</h2>
        <p class="gs-subtitle">Work through questions with your AI. Review the evidence, refine the conclusions, and decide what comes next.</p>
      </div>
    </div>
    <div class="gs-steps">
      <div class="gs-step">
        <span class="gs-step-num">1</span>
        <div class="gs-step-content">
          <strong>Connect your AI assistant</strong>
          <p>
            Add this MCP server to your AI assistant, such as Claude, Cursor or ChatGPT:
          </p>
          <div class="empty-command gs-command">
            <code class="command-text">{{ mcpUrl }}</code>
            <button class="copy-btn" :class="{ copied: copied === 'mcp' }" @click="copy('mcp', mcpUrl)">
              {{ copied === 'mcp' ? '&#x2713; Copied' : 'Copy' }}
            </button>
          </div>
          <p class="gs-note" v-if="authEnabled">
            If your client asks for an API key,
            <NuxtLink to="/settings">create one in Settings &rarr;</NuxtLink>
          </p>
        </div>
      </div>
      <div class="gs-step">
        <span class="gs-step-num">2</span>
        <div class="gs-step-content">
          <strong>Start your first project</strong>
          <p>Send this to your connected AI assistant:</p>
          <div class="empty-command gs-command">
            <code class="command-text">{{ INIT }}</code>
            <button class="copy-btn" :class="{ copied: copied === 'init' }" @click="copy('init', INIT)">
              {{ copied === 'init' ? '&#x2713; Copied' : 'Copy' }}
            </button>
          </div>
        </div>
      </div>
      <div class="gs-step">
        <span class="gs-step-num">3</span>
        <div class="gs-step-content">
          <strong>Review your progress</strong>
          <p>Documents, questions and next steps appear here as your AI assistant works.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
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


/* The address the agent is pointed at.
   Taken from the browser rather than from config, because this page is served
   by the very server the agent has to reach — so whatever the reader typed to
   get here is the answer, including a reverse proxy and a port. */
const mcpUrl = computed(() => (import.meta.client ? `${window.location.origin}/mcp` : '/mcp'))
const INIT = 'Use the research/initialize prompt to start a new project in Dovod.'
const { authEnabled } = useAuth()

const copied = ref<'mcp' | 'init' | null>(null)
async function copy(which: 'mcp' | 'init', text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = which
    setTimeout(() => { if (copied.value === which) copied.value = null }, 2000)
  } catch {
    // A clipboard the browser refuses is not worth a dialog: the text is on
    // screen and selectable, which is what the fallback has always been.
  }
}
</script>

<style scoped>
.gs-command { margin: var(--space-2) 0 0; max-width: none; }
.gs-note { margin-top: var(--space-2); font-size: var(--type-sm); color: var(--color-text-muted); }

.getting-started {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
  margin-bottom: var(--space-8);
  overflow: hidden;
}
.gs-close {
  position: absolute; top: var(--space-4); right: var(--space-4);
  color: var(--color-text-muted);
  font-size: var(--type-lg); line-height: 1;
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.gs-close:hover { color: var(--color-text); background: var(--color-surface-hover); }
.gs-header { margin-bottom: var(--space-6); }
.gs-title { font-family: var(--font-display); font-size: var(--type-xl); font-weight: 800; margin-bottom: var(--space-1); letter-spacing: -0.02em; }
.gs-subtitle { color: var(--color-text-muted); font-size: var(--type-sm); }
.gs-steps { display: flex; flex-direction: column; gap: var(--space-5); }
.gs-step { display: flex; gap: var(--space-4); align-items: flex-start; }
.gs-step-num {
  width: 32px; height: 32px; border-radius: var(--radius-sm);
  background: var(--color-primary-muted); color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-size: var(--type-sm); font-weight: var(--weight-bold); flex-shrink: 0;
}
.gs-step-content strong { display: block; font-weight: var(--weight-semibold); margin-bottom: var(--space-1); font-size: var(--type-base); }
.gs-step-content p { font-size: var(--type-sm); color: var(--color-text-muted); line-height: 1.5; }

/* Responsive */
@media (max-width: 768px) {
  .getting-started { padding: var(--space-5); margin-bottom: var(--space-5); }
  .gs-title { font-size: var(--type-lg); }
}
</style>
