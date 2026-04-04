<template>
  <div v-if="isInMemory" class="warning-banner">
    ⚠ <strong>In-memory mode</strong> — data will be lost on restart.
    Run with <code>--db ./research.db</code> to persist data.
  </div>
</template>

<script setup lang="ts">
const { data } = useApi<{ in_memory: boolean }>('/api/health')
const isInMemory = computed(() => data.value?.in_memory ?? false)
</script>
