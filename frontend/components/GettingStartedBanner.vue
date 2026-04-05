<template>
  <div v-if="show" class="getting-started">
    <button class="gs-close" @click="dismiss">&times;</button>
    <div class="gs-header">
      <span class="gs-icon">&#x1F52C;</span>
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
            <button class="copy-btn" :class="{ copied: copiedInit }" @click="copyInit">{{ copiedInit ? '&#x2713; Copied' : 'Copy' }}</button>
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
</style>
