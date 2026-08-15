<template>
  <span class="short-code">{{ code }}</span>
</template>

<script setup lang="ts">
/**
 * A research's or entry's short code.
 *
 * The rule lives in main.css, not here. Ten pages inline the same span, so a
 * scoped copy in this component meant two definitions of one look — and the
 * scoped one silently won wherever this component was used, which is the
 * failure mode that makes a badge look subtly different from screen to screen.
 */
defineProps<{ code: string }>()
</script>
