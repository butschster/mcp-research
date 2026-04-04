<template>
  <span :class="['badge', `badge-${status}`]">{{ status }}</span>
</template>

<script setup lang="ts">
defineProps<{ status: string }>()
</script>
