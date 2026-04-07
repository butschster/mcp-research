<template>
  <div v-if="node" class="rm-popover" :style="popoverStyle" @click.stop>
    <div class="rm-popover-header">
      <span class="rm-popover-type">{{ node.nodeType }}</span>
      <button class="rm-popover-close" @click="$emit('close')">&times;</button>
    </div>
    <h4 class="rm-popover-title">{{ node.title }}</h4>
    <p v-if="node.description" class="rm-popover-desc">{{ node.description }}</p>
    <div v-if="statuses.length" class="rm-popover-statuses">
      <span class="rm-popover-label">Status</span>
      <div class="rm-popover-chips">
        <button
          v-for="s in statuses"
          :key="s"
          :class="['rm-chip', { active: s === node.status }]"
          @click="$emit('update-status', node.id, s)"
        >{{ s }}</button>
        <button
          v-if="node.status"
          class="rm-chip rm-chip-clear"
          @click="$emit('update-status', node.id, '')"
        >clear</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  node: {
    id: string
    title: string
    description: string
    nodeType: string
    status: string
  } | null
  statuses: string[]
  position: { x: number; y: number }
}>()

defineEmits<{
  'update-status': [nodeId: string, status: string]
  'close': []
}>()

const popoverStyle = computed(() => ({
  left: `${props.position.x}px`,
  top: `${props.position.y}px`,
}))

function onClickOutside(e: MouseEvent) {
  // Handled by parent
}

onMounted(() => {
  document.addEventListener('keydown', onEscape)
})
onUnmounted(() => {
  document.removeEventListener('keydown', onEscape)
})

function onEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    // emit close via parent
  }
}
</script>

<style scoped>
.rm-popover {
  position: fixed;
  z-index: 1000;
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius);
  padding: var(--space-4);
  min-width: 240px;
  max-width: 340px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  pointer-events: auto;
}
.rm-popover-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}
.rm-popover-type {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
}
.rm-popover-close {
  background: none;
  border: none;
  color: var(--color-text-muted);
  font-size: var(--type-lg);
  cursor: pointer;
  padding: 0 var(--space-1);
  line-height: 1;
}
.rm-popover-close:hover {
  color: var(--color-text);
}
.rm-popover-title {
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 var(--space-1);
  line-height: 1.3;
}
.rm-popover-desc {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  line-height: 1.4;
  margin-bottom: var(--space-3);
}
.rm-popover-label {
  font-size: 0.5625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-1);
  display: block;
}
.rm-popover-statuses {
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-3);
}
.rm-popover-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
}
.rm-chip {
  font-size: 0.625rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  background: var(--color-surface-hover);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  line-height: 1.2;
}
.rm-chip:hover {
  border-color: var(--color-primary);
  color: var(--color-text);
}
.rm-chip.active {
  background: var(--color-primary-muted);
  border-color: rgba(108, 197, 224, 0.4);
  color: var(--color-primary);
  font-weight: 600;
}
.rm-chip-clear {
  color: var(--color-text-muted);
  opacity: 0.6;
  font-style: italic;
}
.rm-chip-clear:hover {
  opacity: 1;
  border-color: rgba(239, 107, 107, 0.4);
  color: rgba(239, 107, 107, 1);
}
</style>
