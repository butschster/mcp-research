export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  ssr: false, // SPA mode for static embedding

  runtimeConfig: {
    apiBase: 'http://localhost:8088',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '',
    },
  },

  css: ['~/assets/css/main.css'],

  app: {
    head: {
      title: 'MCP Research',
      meta: [
        { name: 'description', content: 'AI-driven structured research sessions' },
        { property: 'og:title', content: 'MCP Research' },
        { property: 'og:description', content: 'AI-driven structured research sessions' },
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap' },
      ],
    },
  },
})
