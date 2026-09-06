<script setup lang="ts">
import { useCopyToClipboard } from '~/composables/useCopyToClipboard'
import BrandLogo from '~/components/BrandLogo.vue'
import ThemeToggle from '~/components/ThemeToggle.vue'

const props = defineProps<{
  /** Where this instance answers. Shown so a reader knows what they are calling. */
  baseUrl: string
  specUrlYaml: string
  specUrlJson: string
  /** API keys are meaningless without accounts, and /settings is a dead end then. */
  showApiKeysLink: boolean
  /** Signed out, the same link is a trip through the sign-in page, and saying so
   *  beforehand is the difference between a detour and a surprise. */
  signedIn?: boolean
}>()

const { copied, failed, announcement, copy, dismiss } = useCopyToClipboard()

// The URL, not the file. The reader's next move is
// `openapi-generator generate -i <url>`, and a clipboard holding four thousand
// lines of YAML is not what they came for — the YAML button already opens the
// file in a tab if that is what they wanted.
function copyUrl() {
  return copy(props.specUrlYaml, {
    success: 'Copied the specification URL to the clipboard',
    failure: 'Could not copy. The address is shown below — select it and copy it yourself.',
  })
}
</script>

<template>
  <header class="api-docs-bar">
    <div class="api-docs-bar-row">
    <div class="api-docs-bar-left">
      <NuxtLink :to="{ name: 'index' }" class="api-docs-wordmark" aria-label="Dovod home"><BrandLogo /></NuxtLink>
      <!-- Deliberately small. Scalar renders the document's own title as a large
           heading directly below; a page title above it would give the screen
           two competing headlines and push the document's first line under the
           fold. The bar is chrome, the document is the subject. -->
      <h1 class="api-docs-title">API reference</h1>
      <code class="api-docs-base" :title="baseUrl">{{ baseUrl }}</code>
    </div>

    <div class="api-docs-bar-right">
      <ThemeToggle />
      <NuxtLink
        v-if="showApiKeysLink"
        :to="signedIn ? '/settings' : '/login?next=%2Fsettings'"
        class="api-docs-keys"
      >{{ signedIn ? 'API keys' : 'Sign in for a key' }} &rarr;</NuxtLink>
      <a class="btn" :href="specUrlYaml" target="_blank" rel="noopener">YAML</a>
      <a class="btn" :href="specUrlJson" target="_blank" rel="noopener">JSON</a>
      <button
        type="button"
        class="btn-icon"
        :title="copied ? 'Copied' : 'Copy the specification URL'"
        :aria-label="copied ? 'Copied' : 'Copy the specification URL'"
        @click="copyUrl"
      >
        <svg v-if="copied" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path d="M20 6 9 17l-5-5" />
        </svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <rect x="9" y="9" width="12" height="12" rx="2" />
          <path d="M5 15V5a2 2 0 0 1 2-2h10" />
        </svg>
      </button>
      <span class="sr-only" role="status">{{ announcement }}</span>
    </div>
    </div>

    <!-- Visible as well as announced: a 30px icon button that does nothing
         teaches a sighted reader nothing, and outside a secure context the
         clipboard API is simply absent. The address itself is here rather than
         a pointer at the bar, because the bar hides it below 769px — and it is
         a different string from the one the button copies. -->
    <p v-if="failed" class="api-docs-copy-failed">
      <span>Could not copy. Select this and copy it yourself:</span>
      <code class="api-docs-copy-url">{{ specUrlYaml }}</code>
      <button type="button" class="btn btn-sm" @click="dismiss">Dismiss</button>
    </p>
  </header>
</template>

<style scoped>
.api-docs-copy-failed {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  margin: var(--space-2) 0 0;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}
.api-docs-copy-url {
  font-family: 'JetBrains Mono', monospace;
  overflow-wrap: anywhere;
  user-select: all;
}

.api-docs-bar {
  /* Sticky, because the page below it scrolls sixteen thousand pixels and
     everything that leads back to the product — the wordmark, both raw-file
     links, the copy button — is in here. */
  position: sticky;
  top: 0;
  z-index: var(--z-elevated);
  padding: var(--space-2) var(--space-6);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.api-docs-bar-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  min-height: var(--control-h-touch);
}

.api-docs-bar-left,
.api-docs-bar-right {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.api-docs-bar-right { gap: var(--space-2); }

.api-docs-wordmark {
  display: block;
  width: 6.5rem;
  flex-shrink: 0;
  font-size: var(--type-sm);
  font-weight: var(--weight-bold);
  color: var(--color-text);
  text-decoration: none;
  transition: color var(--transition-fast);
}
.api-docs-wordmark:hover { color: var(--color-primary); text-decoration: none; }

.api-docs-title {
  margin: 0;
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
}

.api-docs-base {
  max-width: 28rem;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: 'JetBrains Mono', monospace;
  font-size: var(--type-xs);
  color: var(--color-text-muted);
}

.api-docs-keys {
  font-size: var(--type-xs);
  color: var(--color-primary);
  text-decoration: none;
  white-space: nowrap;
}
.api-docs-keys:hover { text-decoration: underline; }

/* Two rows, and the two things that are only useful beside a wide try-it panel
   are dropped rather than wrapped into a third. */
@media (max-width: 768px) {
  .api-docs-bar { padding-inline: var(--space-4); }
  .api-docs-bar-row {
    flex-wrap: wrap;
    row-gap: var(--space-2);
  }
  .api-docs-bar-left { flex: 1 1 100%; }
  /* Wrapping here matters below 360px: signed out the link reads "Sign in for a
     key", 46px wider than "API keys", and the row went over the edge. */
  .api-docs-bar-right { flex: 1 1 100%; flex-wrap: wrap; justify-content: flex-end; }
  /* The address is only actionable beside the try-it panel, which is barely
     usable at this width. The keys link stays: on a phone it is the only route
     from this page to a credential. */
  .api-docs-base { display: none; }
}
</style>
