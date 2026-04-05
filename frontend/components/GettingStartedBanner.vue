<template>
  <div v-if="show" class="getting-started">
    <button class="gs-close" @click="dismiss">&times;</button>
    <div class="gs-header">
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
          <div class="empty-command" style="margin: var(--space-2) 0 0; max-width: none;">
            <code class="command-text">Use the research/initialize prompt</code>
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
  background: var(--color-surface);
  border: 1px solid rgba(56,189,248,0.2);
  border-radius: var(--radius);
  padding: var(--space-6);
  margin-bottom: var(--space-8);
}
.gs-close {
  position: absolute; top: var(--space-3); right: var(--space-3);
  background: none; border: none; color: var(--color-text-muted);
  font-size: var(--type-lg); cursor: pointer; line-height: 1;
}
.gs-close:hover { color: var(--color-text); }
.gs-header { margin-bottom: var(--space-6); }
.gs-title { font-size: var(--type-xl); font-weight: 700; margin-bottom: var(--space-1); }
.gs-subtitle { color: var(--color-text-muted); font-size: var(--type-sm); }
.gs-steps { display: flex; flex-direction: column; gap: var(--space-4); }
.gs-step { display: flex; gap: var(--space-4); }
.gs-step-num {
  width: var(--space-8); height: var(--space-8); border-radius: 50%;
  background: rgba(56,189,248,0.12); color: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-size: var(--type-sm); font-weight: 700; flex-shrink: 0;
}
.gs-step-content strong { display: block; font-weight: 600; margin-bottom: var(--space-1); }
.gs-step-content p { font-size: var(--type-sm); color: var(--color-text-muted); }
</style>
