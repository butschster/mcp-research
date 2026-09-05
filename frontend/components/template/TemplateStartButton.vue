<script setup lang="ts">
import { useId } from 'vue'
import ModalOverlay from '~/components/ModalOverlay.vue'
import ModalHeader from '~/components/ModalHeader.vue'
import { useCopyToClipboard } from '~/composables/useCopyToClipboard'
import { useMethodologyPrompt, type MethodologyPromptSource } from '~/composables/useMethodologyPrompt'

const props = defineProps<{ methodology: MethodologyPromptSource }>()
const titleId = `methodology-prompt-${useId()}`
const prompt = useMethodologyPrompt(() => props.methodology)
const { copied, failed, announcement, copy, dismiss } = useCopyToClipboard()

function copyPrompt() {
  return copy(prompt.value, {
    success: `Start prompt for ${props.methodology.name} copied`,
    failure: 'Could not copy. The prompt is available to select manually.',
  })
}
</script>

<template>
  <div class="methodology-start">
    <button
      type="button"
      class="btn btn-sm start-prompt-button"
      :class="{ 'is-copied': copied }"
      :aria-label="copied ? `Prompt for ${methodology.name} copied` : `Copy prompt to start a project with ${methodology.name}`"
      :title="`Copy a start prompt for ${methodology.name}`"
      @click="copyPrompt"
    >
      <svg v-if="copied" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M20 6 9 17l-5-5" /></svg>
      <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><rect x="9" y="9" width="12" height="12" rx="2" /><path d="M5 15V5a2 2 0 0 1 2-2h10" /></svg>
      {{ copied ? 'Copied' : 'Copy prompt' }}
    </button>
    <span class="sr-only" role="status">{{ announcement }}</span>
    <ModalOverlay :visible="failed" :labelledby="titleId" size="sm" @close="dismiss">
      <ModalHeader title="Copy start prompt" :title-id="titleId" @close="dismiss" />
      <div class="prompt-fallback">
      <p>Could not copy. Select this prompt and copy it manually:</p>
      <pre tabindex="0">{{ prompt }}</pre>
      <button type="button" class="btn btn-sm" @click="dismiss">Dismiss</button>
      </div>
    </ModalOverlay>
  </div>
</template>

<style scoped>
.methodology-start { min-width: 0; }
.start-prompt-button { min-width: 8rem; }
.is-copied { color: var(--color-success); }
.prompt-fallback { margin-top: var(--space-3); color: var(--color-text-muted); font-size: var(--type-xs); }
.prompt-fallback pre { margin-block: var(--space-2); padding: var(--space-3); background: var(--color-bg); border: 1px solid var(--color-border); border-radius: var(--radius-sm); color: var(--color-text); white-space: pre-wrap; overflow-wrap: anywhere; user-select: text; font-family: 'JetBrains Mono', monospace; }
</style>
