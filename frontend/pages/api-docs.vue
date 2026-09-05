<script setup lang="ts">
import { useTheme } from '~/composables/useTheme'
const { theme } = useTheme()

import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'

// Aliases, not redirects: the address a person guessed stays in the bar. See
// utils/apiDocs.ts for why these particular ones.
definePageMeta({ alias: [...API_DOCS_ALIASES] })

// Keep the reference recognizable when it is open beside the workspace.
useHead({ title: 'API reference' })

const { authEnabled, isAuthenticated } = useAuth()

// Not bare window.location.origin: `make frontend-dev` serves the SPA on :3000
// and the API on :8088, and the try-it panel would miss every request. This is
// the rule useApi() and useServerInfo() already follow.
const config = useRuntimeConfig()
const base = computed(() => config.public.apiBase || (import.meta.client ? window.location.origin : ''))
const specUrlJson = computed(() => `${base.value}/api/openapi.json`)
const specUrlYaml = computed(() => `${base.value}/api/openapi.yaml`)

type State = 'loading' | 'ready' | 'fetch-failed' | 'viewer-failed' | 'empty'

const state = ref<State>('loading')
const document_ = shallowRef<Record<string, unknown> | null>(null)
const failureAction = ref<HTMLElement | null>(null)
const mount = ref<HTMLElement | null>(null)
const bar = ref<HTMLElement | null>(null)
// Scalar's stylesheet subtracts this from its full height and offsets its
// sticky sidebar by it. Measured rather than written down: the first version
// hard-coded 47px and the bar was 53px by the end of the same afternoon, which
// left the sidebar's footer six pixels past the bottom of the window.
const headerHeight = ref(0)

// Loaded here rather than imported at the top of the module on purpose. A
// static import puts Scalar in this route's chunk graph, and Nuxt then writes a
// <link rel="prefetch"> for it into index.html — so every visitor to any page in
// the product downloads about three megabytes in the background for a reference
// manual they may never open.
const ApiReference = defineAsyncComponent({
  loader: async () => {
    await import('@scalar/api-reference/style.css')
    const module = await import('@scalar/api-reference')
    return module.ApiReference
  },
  // The frontend is baked into the binary, so after a redeploy a tab that has
  // been open holds an index.html naming _nuxt/ hashes the server no longer
  // has. Without this the page sits on its skeleton forever.
  onError(_error, _retry, fail) {
    state.value = 'viewer-failed'
    focusFailure()
    fail()
  },
})

const configuration = computed(() => ({
  content: document_.value,
  layout: 'modern' as const,
  darkMode: theme.value === 'dark',
  forceDarkModeState: theme.value,
  hideDarkModeToggle: true,
  // Outfit and JetBrains Mono are already loaded by the app shell.
  withDefaultFonts: false,
  // Scalar's own download serialises the object in memory — which is the
  // document after we rewrote `servers`. Handing a codegen user a file that
  // differs from what the server publishes is a subtle, expensive kind of
  // wrong, so the raw-file affordance lives in our bar and points at the URLs.
  documentDownloadType: 'none' as const,
  // Scalar persists auth under its own key, which this app's logout() does not
  // clear — a shared machine would keep a live bearer token after sign-out.
  persistAuth: false,
  // Defaults to true. This product is self-hosted, often on an internal
  // network, and its API reference reporting back to a third party is not a
  // thing an operator asked for or would find out about.
  telemetry: false,
  // Both default to showing on localhost, which is where this product mostly
  // runs. They are Scalar's editing and deployment surface, not ours — a
  // "Deploy" button inside our documentation belongs to somebody else's
  // product and means nothing here.
  showToolbar: 'never' as const,
  showDeveloperTools: 'never' as const,
  // "Ask AI" and "Generate MCP", same reasoning. Both render when the page is
  // on a local URL, so they are invisible on a real deployment and appear for
  // exactly the person developing against this product — who is the last person
  // who should be offered a third party's AI chat from inside our own
  // documentation. This product has its own MCP server; a button that generates
  // somebody else's is worse than noise.
  agent: { disabled: true },
  mcp: { disabled: true },
  // And the third one, which only became visible once "Generate MCP" was
  // turned off — Scalar renders one or the other in that slot. It opens their
  // API client; this product is not in the business of advertising it.
  hideClientButton: true,
}))

async function focusFailure() {
  await nextTick()
  failureAction.value?.focus()
}

async function load() {
  // Captured before the state is reset, because it is what tells a first load
  // from a retry — and only a retry should move focus.
  const recovering = state.value !== 'loading'
  state.value = 'loading'
  let doc: Record<string, unknown>
  try {
    const response = await fetch(specUrlJson.value, { headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(`status ${response.status}`)
    doc = await response.json()
  } catch {
    state.value = 'fetch-failed'
    focusFailure()
    return
  }

  const paths = doc.paths as Record<string, unknown> | undefined
  if (!paths || Object.keys(paths).length === 0) {
    state.value = 'empty'
    focusFailure()
    return
  }

  // The document's own `servers` is `/` unless the operator configured
  // `base_url`. Relative is correct for a file a client fetched itself, but
  // Scalar needs somewhere concrete to aim a test request, and it must be the
  // instance this page came from. Patching the document rather than passing
  // `configuration.servers` keeps one mechanism, so what Scalar displays is
  // what it will call.
  doc.servers = [{ url: base.value, description: 'This instance' }]

  document_.value = doc
  state.value = 'ready'
  await nextTick()
  // Only after a retry. Focusing on first load looks helpful and is not: it
  // drops the caret inside Scalar, so forward Tab never reaches the bar, and
  // the skip link — whose whole purpose is not tabbing through 126 operations
  // — ends up eleven Shift+Tabs behind the reader.
  if (recovering) mount.value?.focus?.()
  await verifyStylesheet()
}

// A missing stylesheet does not reject the dynamic import: Vite injects a
// <link>, and a 404 fails the element rather than the promise. The result is an
// unstyled document nineteen thousand pixels wide with no error at all — the
// same redeploy this page has a state for, arriving through the half that was
// not covered.
//
// This checks the symptom, not the library. A first attempt looked for one of
// Scalar's CSS variables on `.scalar-app` and was wrong about where they are
// declared, which broke the page in exactly the way it was meant to detect.
// Width is version-independent: styled, this layout never overflows at any
// viewport; unstyled, it is thirteen times too wide.
async function verifyStylesheet() {
  await nextTick()
  await new Promise(requestAnimationFrame)
  await new Promise(requestAnimationFrame)
  if (state.value !== 'ready') return
  const root = document.documentElement
  if (root.scrollWidth > root.clientWidth * 1.5) {
    state.value = 'viewer-failed'
    focusFailure()
  }
}

// Focus without navigating: an `href` jump replaces Scalar's own routing hash,
// and the address a keyboard reader then copies no longer names the operation
// they are on.
function skipToReference() {
  mount.value?.focus?.()
  mount.value?.scrollIntoView?.({ block: 'start' })
}

function reloadPage() {
  // Not a retry: a second import() of the same dead chunk URL fails the same
  // way, and offering a button that cannot work is worse than saying nothing.
  window.location.reload()
}

let barObserver: ResizeObserver | undefined

onMounted(() => {
  const el = (bar.value as unknown as { $el?: HTMLElement })?.$el ?? bar.value
  if (el && typeof ResizeObserver !== 'undefined') {
    barObserver = new ResizeObserver(() => {
      headerHeight.value = Math.round(el.getBoundingClientRect().height)
    })
    barObserver.observe(el)
  }
  return load()
})

onBeforeUnmount(() => barObserver?.disconnect())
</script>

<template>
  <div class="api-docs-page">
    <a href="#api-reference" class="skip-link" @click.prevent="skipToReference">Skip to the API reference</a>

    <ApiDocsHeader
      ref="bar"
      :base-url="base"
      :spec-url-yaml="specUrlYaml"
      :spec-url-json="specUrlJson"
      :show-api-keys-link="Boolean(authEnabled)"
      :signed-in="Boolean(isAuthenticated)"
    />

    <div v-if="state === 'loading'" class="api-docs-skeleton" role="status" aria-busy="true">
      <span class="sr-only">Loading the API reference</span>
      <div class="skeleton-card api-docs-skeleton-side" />
      <div class="skeleton-card api-docs-skeleton-main" />
    </div>

    <div v-else-if="state !== 'ready'" class="api-docs-failure" role="alert">
      <EmptyState
        v-if="state === 'fetch-failed'"
        icon="⚠"
        title="Couldn't load the API reference"
        description="The server didn't return its specification. It is published at /api/openapi.json — if that address doesn't answer either, the server is down or still starting."
      >
        <button ref="failureAction" type="button" class="btn btn-primary" tabindex="-1" @click="load">Try again</button>
        <a class="btn" :href="specUrlYaml" target="_blank" rel="noopener">Open the raw specification</a>
      </EmptyState>

      <EmptyState
        v-else-if="state === 'viewer-failed'"
        icon="⚠"
        title="Couldn't load the reference viewer"
        description="This page's viewer failed to download. If the server was updated while this tab was open, reloading will pick up the new version."
      >
        <button ref="failureAction" type="button" class="btn btn-primary" tabindex="-1" @click="reloadPage">Reload the page</button>
        <a class="btn" :href="specUrlYaml" target="_blank" rel="noopener">Open the raw specification</a>
      </EmptyState>

      <EmptyState
        v-else
        icon="⚠"
        title="The server's specification came back empty"
        description="It answered, but described no routes. That's a problem on the server, not in this browser."
      >
        <button ref="failureAction" type="button" class="btn btn-primary" tabindex="-1" @click="load">Try again</button>
        <a class="btn" :href="specUrlYaml" target="_blank" rel="noopener">Open the raw specification</a>
      </EmptyState>
    </div>

    <!-- Scalar manages its own sidebar and content scrolling. A second scroll
         container here is the double scrollbar every embedded docs viewer
         eventually grows. -->
    <div
      v-else
      id="api-reference"
      ref="mount"
      class="api-docs-mount"
      tabindex="-1"
      :style="{ '--scalar-custom-header-height': `${headerHeight}px` }"
    >
      <ApiReference :configuration="configuration" />
    </div>
  </div>
</template>

<style scoped>
.api-docs-page {
  display: flex;
  flex-direction: column;
  min-height: 100dvh;
  background: var(--color-bg);
}

.api-docs-skeleton {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: var(--space-4);
  padding: var(--space-6);
}
.api-docs-skeleton-side,
.api-docs-skeleton-main { height: 100%; }

.api-docs-failure {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}

.api-docs-mount {
  flex: 1;
  min-height: 0;
}

/* Scalar creates nested color-mode roots, including one in its sidebar.
   Map each root to the app palette so it follows the same preference. */
.api-docs-mount,
.api-docs-mount :deep(.light-mode),
.api-docs-mount :deep(.dark-mode) {
  --scalar-background-1: var(--color-bg);
  --scalar-background-2: var(--color-surface);
  --scalar-background-3: var(--color-surface-hover);
  --scalar-color-1: var(--color-text);
  --scalar-color-2: var(--color-text-muted);
  --scalar-color-3: var(--color-text-muted);
  --scalar-border-color: var(--color-border);
  --scalar-color-accent: var(--color-primary);
  --scalar-background-accent: var(--color-primary-muted);
  --scalar-button-1: var(--color-primary);
  --scalar-button-1-hover: var(--color-primary-hover);
  --scalar-button-1-color: var(--color-on-primary);
  --scalar-sidebar-background-1: var(--color-surface);
  --scalar-sidebar-color-1: var(--color-text);
  --scalar-sidebar-color-2: var(--color-text-muted);
  --scalar-sidebar-color-active: var(--color-primary);
  --scalar-sidebar-border-color: var(--color-border);
  --scalar-sidebar-item-hover-background: var(--color-surface-hover);
  --scalar-sidebar-item-hover-color: var(--color-text);
  --scalar-sidebar-item-active-background: var(--color-primary-muted);
  --scalar-sidebar-search-background: var(--color-bg);
  --scalar-sidebar-search-color: var(--color-text-muted);
  --scalar-sidebar-search-border-color: var(--color-border);
  --scalar-font: 'Outfit', sans-serif;
  --scalar-font-code: 'JetBrains Mono', monospace;
}
.api-docs-mount:focus { outline: none; }
</style>
