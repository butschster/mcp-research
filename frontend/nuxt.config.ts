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
      ],
    },
  },
})
