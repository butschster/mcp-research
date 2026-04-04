<template>
  <div v-if="isInMemory" class="warning-banner">
    ⚠ In-memory database — all data will be lost on restart. Use <code>--db ./research.db</code> for persistence.
  </div>
</template>

<script setup lang="ts">
const { data } = useApi<{ in_memory: boolean }>('/api/health')
const isInMemory = computed(() => data.value?.in_memory ?? false)
</script>
