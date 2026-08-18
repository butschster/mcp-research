<template>
  <div class="rm-graph-node">
    <RoadmapNodeCard :data="data" />
    <Handle type="target" :position="targetPosition" />
    <Handle type="source" :position="sourcePosition" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'
import type { RoadmapCardData } from './RoadmapNodeCard.vue'

// Thin Vue Flow wrapper: the card body is shared with the stages board and the
// timeline via RoadmapNodeCard; only the <Handle>s (which require the Vue Flow
// provider) live here.
defineProps<{
  data: RoadmapCardData
  sourcePosition?: Position
  targetPosition?: Position
}>()
</script>

<style scoped>
.rm-graph-node {
  position: relative;
  width: 320px;
}
</style>
