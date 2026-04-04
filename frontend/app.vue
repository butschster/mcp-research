<template>
  <div>
    <nav>
      <div class="container nav-inner">
        <NuxtLink to="/" class="logo">MCP Research</NuxtLink>
        <div class="nav-right">
          <ConnectionStatus />
        </div>
      </div>
    </nav>

    <main class="container" style="padding-top: 2rem; padding-bottom: 4rem;">
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
        >GitHub ↗</a>
      </div>
    </footer>
  </div>
</template>

<style>
.nav-inner  { display: flex; align-items: center; justify-content: space-between; }
.nav-right  { display: flex; align-items: center; gap: 1rem; }
.app-footer { border-top: 1px solid var(--color-border); padding: 0.875rem 0; margin-top: 2rem; }
.footer-inner { display: flex; align-items: center; justify-content: space-between; }
.footer-link  { text-decoration: none; transition: color 0.15s; }
.footer-link:hover { color: var(--color-primary); }
</style>
