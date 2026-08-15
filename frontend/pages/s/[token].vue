<template>
  <!-- Password first: a locked link has not been resolved, so there is nothing
       to put a banner on top of. -->
  <SharePasswordGate
    v-if="screen === 'locked'"
    :busy="unlocking"
    :error="unlockError"
    @submit="tryUnlock"
  />

  <div v-else-if="screen === 'dead'" class="share-standalone">
    <EmptyState
      icon="&#x1F517;"
      title="This link isn't available"
      description="It may have been turned off, it may have expired, or the address may be incomplete — these links are long and easy to cut short. Ask the person who sent it for a new one."
    />
  </div>

  <div v-else-if="screen === 'unreachable'" class="share-standalone">
    <EmptyState
      title="Couldn't open this link"
      description="The server didn't answer. The link is probably fine — try again in a moment."
    >
      <button class="btn btn-primary" @click="load()">Try again</button>
    </EmptyState>
  </div>

  <div v-else class="share-shell">
    <ShareBanner
      v-if="screen !== 'loading'"
      :owner-name="ownerName"
      :include="include"
      :expires-at="expiresAt"
      :live="hasRecentUpdate"
    />

    <main class="container share-main">
      <div v-if="screen === 'loading'" class="skeleton-page">
        <div class="skeleton-card skeleton-header"></div>
        <div class="layout-sidebar">
          <div class="skeleton-card skeleton-sidebar"></div>
          <div>
            <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
          </div>
        </div>
      </div>
      <NuxtPage v-else />
    </main>

    <footer class="app-footer">
      <div class="container footer-inner">
        <span class="card-meta">Shared research</span>
        <div class="footer-right">
          <ConnectionStatus
            :state="connection"
            :reason="connectionReason"
            :last-synced-at="lastSyncedAt"
            @retry="retryNow"
          />
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
/**
 * The shared view's shell, and the only place in it that decides anything.
 *
 * It owns the single fetch of the share payload, the password gate, the two
 * failure screens, the realtime subscription and the read-only lock. Children
 * render content; a child never decides whether the visitor is allowed in,
 * because a decision made in seven places is a decision six of them will
 * eventually get wrong.
 */
const route = useRoute()
const token = route.params.token as string

const {
  claim, setPayload, setUnlock, release, shareFetch,
  include, ownerName, expiresAt, researchId, researchCode, research, unlock: unlockValue,
} = useShare()

// Claimed before anything is fetched. The read-only lock and the path helpers
// have to be right while the payload is still in flight — a page that turns
// read-only a moment after rendering shows edit controls to a stranger first.
claim(token)
const { lockForShare } = useResearchRole()
lockForShare()

type Screen = 'loading' | 'ready' | 'locked' | 'dead' | 'unreachable'
const screen = ref<Screen>('loading')
const unlocking = ref(false)
const unlockError = ref<'' | 'wrong' | 'throttled'>('')

async function load() {
  screen.value = 'loading'
  try {
    // `visit=1` is what makes this count as a visit. The refetch below asks the
    // same route without it, so a tab left open does not inflate the number the
    // owner reads as "how many people opened this".
    const res = await shareFetch<{ data: any }>('?visit=1')
    setPayload(res.data)
    screen.value = 'ready'
  } catch (e: any) {
    const status = e?.response?.status ?? e?.statusCode
    if (status === 401) {
      screen.value = 'locked'
      return
    }
    // 404 is the one answer for revoked, expired and never-existed, and the UI
    // must not speculate about which of the three it was.
    screen.value = status === 404 ? 'dead' : 'unreachable'
  }
}

async function tryUnlock(password: string) {
  unlocking.value = true
  unlockError.value = ''
  try {
    const res = await shareFetch<{ data: { unlock: string } }>('/unlock', {
      method: 'POST',
      body: { password },
    })
    setUnlock(res.data.unlock)
    await load()
    // The socket was opened before the password was known, and the server
    // refused it — a locked link resolves to nothing. Reopen it now, or the
    // page a visitor just unlocked never receives an update and settles on a
    // permanently offline indicator.
    connectSocket()
  } catch (e: any) {
    const status = e?.response?.status ?? e?.statusCode
    if (status === 429) unlockError.value = 'throttled'
    else if (status === 401) unlockError.value = 'wrong'
    else screen.value = 'unreachable'
  } finally {
    unlocking.value = false
  }
}

// Deliberately not awaited.
//
// A top-level `await` here makes the shell a Suspense boundary, and on a cold
// load there is no previous view to hold — so the visitor saw a blank dark page
// for the length of the request and the skeleton below was unreachable. Kicking
// the fetch off without awaiting it lets the shell render its loading state
// immediately, which is the whole point of having one.
void load()

// --- Realtime ---
//
// Watching a research change while you read it is the single most compelling
// thing this product does, and a share that were a frozen snapshot would throw
// it away. The socket carries the share token, and the server scopes it to this
// one research.
const hasRecentUpdate = ref(false)
let blipTimer: ReturnType<typeof setTimeout>

function connectSocket() {
  if (import.meta.server) return
  useShareSocket(token, unlockValue.value)
}

// Not for a link still behind its password: the handshake would be refused and
// the client would spend its retry budget before the visitor has typed a
// character.
if (screen.value !== 'locked') connectSocket()

const {
  display: connection,
  reason: connectionReason,
  lastSyncedAt,
  retryNow,
} = useConnectionBanner()

useResearchRealtime(
  () => researchCode.value || researchId.value,
  () => {
    hasRecentUpdate.value = true
    clearTimeout(blipTimer)
    blipTimer = setTimeout(() => { hasRecentUpdate.value = false }, 5000)
    // The counts live on the share payload and the entry lists do not, so
    // refetching one and not the other leaves the page half true.
    void refreshPayload()
  },
  { onResync: () => void refreshPayload(), researchId: () => researchId.value },
)

/** A refetch that does not reset the screen — the page must not blink. */
async function refreshPayload() {
  try {
    const res = await shareFetch<{ data: any }>('')
    setPayload(res.data)
  } catch (e: any) {
    const status = e?.response?.status ?? e?.statusCode
    // Revocation is not advisory. The owner pressing Revoke has to mean the
    // content leaves the screen, and there is no unsaved work here to protect.
    if (status === 404) screen.value = 'dead'
    else if (status === 401) screen.value = 'locked'
  }
}

// The server closes a share socket with 4401 when the link stops being valid.
// The close code alone does not say whether that was a revocation or a hiccup,
// so ask once — the same question the page would ask on its next navigation.
watch(connectionReason, (r) => {
  if (r === 'auth') void refreshPayload()
})

onUnmounted(() => {
  clearTimeout(blipTimer)
  releaseShareSocket()
  // Cleared here, unlike the research role: the pages that follow a shared view
  // are never share pages, so nothing downstream can be racing to set it — and
  // a token left behind would keep building `/s/…` links on an ordinary page.
  release()
})

// Every share tab was called "Shared research", so a client sent three links
// had three indistinguishable tabs and the sender's message was the only thing
// carrying the name.
useHead({ title: () => (research.value?.name ? `${research.value.name} — shared` : 'Shared research') })
</script>

<style scoped>
.share-shell { min-height: 100dvh; display: flex; flex-direction: column; }
.share-main { flex: 1; padding-top: var(--space-4); padding-bottom: var(--space-4); }
.share-standalone {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}
</style>
