<template>
  <div class="share-banner">
    <div class="container share-banner-inner">
      <button
        class="share-banner-main"
        :aria-expanded="open"
        aria-controls="share-banner-detail"
        @click="open = !open"
      >
        <span class="share-dot" aria-hidden="true"></span>
        <span class="share-banner-text">
          <strong>Read-only shared view</strong>
          <span v-if="ownerName" class="share-by">&mdash; shared by {{ ownerName }}</span>
          <span class="share-contents">&middot; {{ contents }}</span>
        </span>
        <svg
          class="share-chevron" :class="{ 'share-chevron--open': open }"
          width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      <ActivityIndicator :active="live" label="Updating" />
    </div>

    <div v-if="open" id="share-banner-detail" class="container share-banner-detail">
      <p class="card-meta">
        You are reading a link someone shared with you. Nothing here can be edited, and the link
        reaches this one research and nothing else.
      </p>
      <p class="card-meta">
        <strong>Included:</strong> {{ contents }}.
        <template v-if="expiresAt"> This link stops working {{ expiryPhrase }}.</template>
        <template v-else> This link has no end date; the person who shared it can turn it off at any time.</template>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The permanent strip across a shared view.
 *
 * It exists so nobody mistakes this page for their own workspace — the page
 * below it is, deliberately, the same page the owner sees. The disclosure holds
 * the part that only matters once: what the link includes and when it lapses.
 */
const props = defineProps<{
  ownerName?: string
  include: ShareInclude
  expiresAt?: string | null
  /** Something changed in the last few seconds. */
  live?: boolean
}>()

const open = ref(false)

// Both come from useShare, so the words a visitor reads in this banner and the
// words the owner reads on the row are the same words.
const contents = computed(() => shareContents(props.include, 'prose'))
const expiryPhrase = computed(() => shareExpiryPhrase(props.expiresAt))
</script>

<style scoped>
.share-banner {
  position: sticky;
  top: 0;
  z-index: var(--z-elevated);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}
.share-banner-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  min-height: 36px;
}
.share-banner-main {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: 1;
  min-width: 0;
  background: none;
  border: none;
  padding: var(--space-2) 0;
  color: var(--color-text-muted);
  font-family: inherit;
  font-size: var(--type-xs);
  text-align: left;
  cursor: pointer;
}
.share-banner-main:hover { color: var(--color-text); }
.share-banner-text { min-width: 0; overflow-wrap: anywhere; }
.share-banner-text strong { color: var(--color-text); font-weight: 600; }
.share-dot {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 50%;
  background: var(--color-primary);
}
.share-chevron { flex: none; transition: transform var(--transition-fast); }
.share-chevron--open { transform: rotate(180deg); }
.share-banner-detail {
  padding-bottom: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.share-banner-detail .card-meta { font-size: var(--type-xs); overflow-wrap: anywhere; }

@media (max-width: 768px) {
  .share-contents { display: none; }
}
@media print {
  .share-banner { display: none; }
}
</style>
