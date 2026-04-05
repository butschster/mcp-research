<script setup lang="ts">
import { useKeyboardNav } from '~/composables/useKeyboardNav'

const hasRecentUpdate = ref(false)
let activityTimer: ReturnType<typeof setTimeout>

useRealtimeUpdates(() => {
  hasRecentUpdate.value = true
  clearTimeout(activityTimer)
  activityTimer = setTimeout(() => {
    hasRecentUpdate.value = false
  }, 5000)
})

useKeyboardNav()

const { user, isAuthenticated, authEnabled, logout } = useAuth()
</script>

<template>
  <div>
    <nav class="app-nav">
      <div class="container nav-inner">
        <NuxtLink to="/" class="logo">MCP Research</NuxtLink>
        <div class="nav-right">
          <SearchModal />
          <ActivityIndicator :active="hasRecentUpdate" label="Updating" />
          <template v-if="authEnabled && isAuthenticated">
            <NuxtLink to="/settings" class="nav-user">{{ user?.name || user?.email }}</NuxtLink>
            <button class="nav-logout" @click="logout">Sign out</button>
          </template>
          <span v-else-if="!authEnabled" class="readonly-badge">Read-only</span>
          <ConnectionStatus />
        </div>
      </div>
    </nav>

    <main class="container main-content">
      <WarningBanner />
      <NuxtPage />
    </main>

    <footer class="app-footer">
      <div class="container footer-inner">
        <span class="card-meta">MCP Research</span>
        <a
          href="https://github.com/butschster/mcp-research"
          target="_blank"
          rel="noopener"
          class="card-meta footer-link"
        >GitHub &#x2197;</a>
      </div>
    </footer>
  </div>
</template>

<style>
.nav-inner  { display: flex; align-items: center; justify-content: space-between; }
.nav-right  { display: flex; align-items: center; gap: var(--space-3); }
.main-content { padding-top: var(--space-10); padding-bottom: var(--space-16); min-height: calc(100dvh - 200px); }
.app-footer {
  border-top: 1px solid var(--color-border);
  padding: var(--space-4) 0;
  margin-top: var(--space-12);
}
.footer-inner { display: flex; align-items: center; justify-content: space-between; }
.footer-link  { text-decoration: none; transition: color var(--transition-fast); }
.footer-link:hover { color: var(--color-primary); text-decoration: none; }
.nav-user {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  text-decoration: none;
}
.nav-user:hover { color: var(--color-primary); }
.nav-logout {
  font-size: var(--text-sm);
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-family: inherit;
  padding: 0;
}
.nav-logout:hover { color: var(--color-danger, #dc2626); }
</style>
