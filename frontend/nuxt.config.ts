import { readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { projectsPath } from './config/projects-path.mjs'

const projectListPath = projectsPath(process.env.NUXT_PROJECTS_PATH)

export default defineNuxtConfig({
  hooks: {
    async 'nitro:build:public-assets'(nitro) {
      const file = join(nitro.options.output.publicDir, 'site.webmanifest')
      const manifest = JSON.parse(await readFile(file, 'utf8'))
      manifest.start_url = projectListPath
      manifest.scope = '/'
      await writeFile(file, JSON.stringify(manifest, null, 2) + '\n')
    },
    'pages:extend'(pages) {
      const index = pages.find(page => page.name === 'index')
      if (!index) throw new Error('Projects index route was not found')
      index.path = projectListPath
    },
  },

  compatibilityDate: '2025-01-01',
  ssr: false, // SPA mode for static embedding

  runtimeConfig: {
    apiBase: 'http://localhost:8088',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '',
    },
  },

  // Order is load-bearing. tokens before anything that uses them; system before
  // markdown, whose `.markdown-content pre` the mermaid error state builds on;
  // mermaid last for the same reason.
  css: [
    '~/assets/css/tokens.css',
    '~/assets/css/base.css',
    '~/assets/css/brand.css',
    '~/assets/css/system.css',
    '~/assets/css/markdown.css',
    '~/assets/css/mermaid.css',
  ],

  app: {
    head: {
      title: 'Dovod',
      // Apply the saved preference before CSS paints, including auth pages.
      script: [{ innerHTML: "try{document.documentElement.dataset.theme=localStorage.getItem('dovod-theme')==='dark'?'dark':'light'}catch{}" }],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' },
        { name: 'description', content: 'Work through questions with AI. Keep sources, reasoning, and next steps in one place.' },
        { property: 'og:title', content: 'Dovod' },
        { property: 'og:site_name', content: 'Dovod' },
        { property: 'og:image', content: '/brand/social-card.png' },
        { property: 'og:type', content: 'website' },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'theme-color', content: '#f6f3ec' },
        { name: 'application-name', content: 'Dovod' },
        { property: 'og:description', content: 'Work through questions with AI. Keep sources, reasoning, and next steps in one place.' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap' },
      ],
    },
  },
})
