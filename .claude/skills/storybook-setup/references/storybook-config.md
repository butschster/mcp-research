# Storybook Configuration Reference

Complete configuration files for Storybook 8 + Vue 3 + Nuxt 4 integration.

## .storybook/main.ts — Full version

```ts
import type { StorybookConfig } from '@storybook/vue3-vite'
import { mergeConfig } from 'vite'
import path from 'path'

const config: StorybookConfig = {
  stories: ['../components/**/*.stories.ts'],
  addons: [
    '@storybook/addon-essentials',
    '@storybook/addon-links',
    '@storybook/addon-interactions',
  ],
  framework: {
    name: '@storybook/vue3-vite',
    options: {},
  },
  docs: {
    autodocs: 'tag',
  },
  staticDirs: ['../public'],
  viteFinal: async (config) => {
    const AutoImport = (await import('unplugin-auto-import/vite')).default

    return mergeConfig(config, {
      resolve: {
        alias: {
          '~': path.resolve(__dirname, '..'),
          '@': path.resolve(__dirname, '..'),
          '#imports': path.resolve(__dirname, './stubs/imports.ts'),
          '#app': path.resolve(__dirname, './stubs/app.ts'),
        },
      },
      plugins: [
        AutoImport({
          imports: ['vue', 'vue-router'],
          dts: false,
        }),
      ],
    })
  },
}
export default config
```

## .storybook/stubs/imports.ts

Stub for `#imports` used by Nuxt auto-imports in components:

```ts
export {
  ref,
  computed,
  reactive,
  watch,
  watchEffect,
  onMounted,
  onUnmounted,
  onBeforeMount,
  onBeforeUnmount,
  nextTick,
  provide,
  inject,
  toRef,
  toRefs,
  unref,
  isRef,
  shallowRef,
  triggerRef,
  defineProps,
  defineEmits,
  defineExpose,
  withDefaults,
} from 'vue'

export { useRoute, useRouter } from 'vue-router'

// Nuxt-specific stubs
export const navigateTo = (path: string) => {
  console.log('[Storybook stub] navigateTo:', path)
}
export const useRuntimeConfig = () => ({
  public: { apiBase: '' },
})
export const useFetch = () => ({ data: ref(null), pending: ref(false), error: ref(null) })
export const useAsyncData = () => ({ data: ref(null), pending: ref(false), error: ref(null) })
export const useCookie = (name: string) => ref('')
export const useHead = () => {}
export const useSeoMeta = () => {}
export const definePageMeta = () => {}

// Project composable stubs
export const useApi = (url: any) => ({
  data: ref(null),
  pending: ref(false),
  error: ref(null),
  refresh: () => Promise.resolve(),
})

export const useAuth = () => ({
  user: ref(null),
  token: ref(null),
  authEnabled: ref(false),
  allowRegistration: ref(true),
  loading: ref(false),
  isAuthenticated: computed(() => false),
  fetchAuthInfo: () => Promise.resolve(),
  checkAuth: () => Promise.resolve(),
  login: () => Promise.resolve(),
  register: () => Promise.resolve(),
  logout: () => {},
  authFetch: (url: string, opts?: any) => Promise.resolve({} as any),
})

export const renderRefs = (text: string, slug?: string) => text ?? ''
export const useCrossRefs = () => ({ renderRefs })

export const useRealtimeUpdates = (handler?: any) => {}
export const useKeyboardNav = () => {}
```

## .storybook/stubs/app.ts

Stub for `#app`:

```ts
export { navigateTo, useRuntimeConfig, useFetch, useAsyncData, useCookie } from './imports'
```

## .storybook/preview.ts — Full version

```ts
import type { Preview } from '@storybook/vue3'
import { setup } from '@storybook/vue3'
import { createMemoryHistory, createRouter } from 'vue-router'
import '../assets/css/main.css'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div>Home</div>' } },
    { path: '/:catchAll(.*)', component: { template: '<div>Page</div>' } },
  ],
})

setup((app) => {
  app.use(router)

  // Nuxt component stubs
  app.component('NuxtLink', {
    props: ['to', 'href'],
    template: '<a :href="to || href"><slot /></a>',
  })
  app.component('ClientOnly', {
    template: '<slot />',
  })
  app.component('NuxtPage', {
    template: '<div>[NuxtPage stub]</div>',
  })
  // Render Teleport content inline (for modal stories)
  app.component('Teleport', {
    template: '<slot />',
  })
})

const preview: Preview = {
  parameters: {
    backgrounds: { disable: true },
    layout: 'padded',
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    docs: {
      toc: true,
    },
  },
  decorators: [
    (story) => ({
      components: { story },
      template: `
        <div style="background: var(--color-bg); color: var(--color-text); min-height: 100vh; padding: 1.5rem; font-family: 'Outfit', system-ui, sans-serif;">
          <story />
        </div>
      `,
    }),
  ],
}
export default preview
```

## .storybook/preview-head.html

```html
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet" />
```

## .storybook/manager.ts

```ts
import { addons } from '@storybook/manager-api'
import { themes } from '@storybook/theming'

addons.setConfig({
  theme: {
    ...themes.dark,
    brandTitle: 'MCP Research — Component Library',
    appBg: '#0c1220',
    appContentBg: '#151d2e',
    appBorderColor: 'rgba(148, 163, 184, 0.12)',
    barBg: '#151d2e',
    textColor: '#e2e8f0',
    textMutedColor: '#7f8ea3',
  },
})
```

## package.json additions

```json
{
  "scripts": {
    "storybook": "storybook dev -p 6006",
    "build-storybook": "storybook build -o storybook-static"
  },
  "devDependencies": {
    "@storybook/vue3": "^8.0.0",
    "@storybook/vue3-vite": "^8.0.0",
    "@storybook/addon-essentials": "^8.0.0",
    "@storybook/addon-links": "^8.0.0",
    "@storybook/addon-interactions": "^8.0.0",
    "@storybook/blocks": "^8.0.0",
    "@storybook/manager-api": "^8.0.0",
    "@storybook/theming": "^8.0.0",
    "storybook": "^8.0.0",
    "unplugin-auto-import": "^0.18.0"
  }
}
```

## .gitignore additions

```
storybook-static/
```

## Known issues & workarounds

### Teleport components (modals)
Modals use `<Teleport to="body">`. The global Teleport stub in preview.ts renders content inline, so modals appear within the story canvas.

### useRoute / useRouter
Already handled by the router setup in preview.ts. If a component reads route params:
```ts
export const WithRouteParam: Story = {
  decorators: [
    () => ({
      setup() {
        const router = useRouter()
        router.push('/research/R1/entry/E5')
      },
      template: '<story />',
    }),
  ],
}
```

### marked library
Components using `marked` need it explicitly imported. It works in Storybook without issues since it's a pure JS library.

### renderRefs
The `renderRefs` function from `useCrossRefs.ts` is stubbed as a pass-through in the imports stub. Cross-reference links will render as plain text in stories, which is fine for visual testing.

### Vue Flow (mindmap)
Mindmap node components use `@vue-flow/core`. These may need the VueFlow provider in story decorators:
```ts
import { VueFlow } from '@vue-flow/core'

decorators: [
  () => ({
    components: { VueFlow },
    template: '<VueFlow><story /></VueFlow>',
  }),
]
```
Or they can be rendered standalone if they don't strictly need the flow context for visual display.

## Design System Token Stories

Create `components/__docs__/DesignTokens.stories.ts`:

```ts
import type { Meta, StoryObj } from '@storybook/vue3'

const meta: Meta = {
  title: 'Design System/Tokens',
  tags: ['autodocs'],
}
export default meta

export const Colors: StoryObj = {
  render: () => ({
    template: `
      <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 16px;">
        <div v-for="c in colors" :key="c.name" style="display: flex; align-items: center; gap: 12px;">
          <div :style="{ width: '40px', height: '40px', borderRadius: '6px', background: c.value, border: '1px solid rgba(148,163,184,0.12)' }"></div>
          <div>
            <div style="font-size: 0.875rem; font-weight: 500;">{{ c.name }}</div>
            <div style="font-size: 0.75rem; color: var(--color-text-muted); font-family: 'JetBrains Mono', monospace;">{{ c.value }}</div>
          </div>
        </div>
      </div>
    `,
    setup() {
      const colors = [
        { name: '--color-bg', value: '#0c1220' },
        { name: '--color-surface', value: '#151d2e' },
        { name: '--color-surface-hover', value: '#1e2940' },
        { name: '--color-text', value: '#e2e8f0' },
        { name: '--color-text-muted', value: '#7f8ea3' },
        { name: '--color-primary', value: '#6cc5e0' },
        { name: '--color-success', value: '#34d399' },
        { name: '--color-warning', value: '#f0b849' },
        { name: '--color-error', value: '#ef6b6b' },
        { name: '--color-info', value: '#6b9df0' },
        { name: '--color-border', value: 'rgba(148,163,184,0.12)' },
        { name: '--color-border-strong', value: 'rgba(148,163,184,0.22)' },
        { name: '--color-primary-muted', value: 'rgba(108,197,224,0.10)' },
      ]
      return { colors }
    },
  }),
}

export const Typography: StoryObj = {
  render: () => ({
    template: `
      <div style="display: flex; flex-direction: column; gap: 24px;">
        <div v-for="t in types" :key="t.name" style="display: flex; align-items: baseline; gap: 16px;">
          <code style="font-size: 0.75rem; color: var(--color-text-muted); min-width: 120px; font-family: 'JetBrains Mono', monospace;">{{ t.name }}</code>
          <span :style="{ fontSize: t.value, fontFamily: t.mono ? '\\'JetBrains Mono\\', monospace' : '\\'Outfit\\', sans-serif' }">
            {{ t.mono ? 'Code sample ABC 123' : 'The quick brown fox jumps' }}
          </span>
        </div>
      </div>
    `,
    setup() {
      const types = [
        { name: '--type-xs', value: '0.875rem', mono: false },
        { name: '--type-sm', value: '0.9375rem', mono: false },
        { name: '--type-base', value: '1.0625rem', mono: false },
        { name: '--type-lg', value: '1.25rem', mono: false },
        { name: '--type-xl', value: '1.5rem', mono: false },
        { name: '--type-2xl', value: '2rem', mono: false },
        { name: 'JetBrains Mono 400', value: '0.875rem', mono: true },
        { name: 'JetBrains Mono 500', value: '0.875rem', mono: true },
      ]
      return { types }
    },
  }),
}

export const Spacing: StoryObj = {
  render: () => ({
    template: `
      <div style="display: flex; flex-direction: column; gap: 8px;">
        <div v-for="s in spaces" :key="s.name" style="display: flex; align-items: center; gap: 16px;">
          <code style="font-size: 0.75rem; color: var(--color-text-muted); min-width: 100px; font-family: 'JetBrains Mono', monospace;">{{ s.name }}</code>
          <div :style="{ width: s.value, height: '16px', background: 'var(--color-primary)', borderRadius: '2px', opacity: 0.7 }"></div>
          <span style="font-size: 0.75rem; color: var(--color-text-muted);">{{ s.value }}</span>
        </div>
      </div>
    `,
    setup() {
      const spaces = [
        { name: '--space-1', value: '0.25rem' },
        { name: '--space-2', value: '0.5rem' },
        { name: '--space-3', value: '0.75rem' },
        { name: '--space-4', value: '1rem' },
        { name: '--space-5', value: '1.25rem' },
        { name: '--space-6', value: '1.5rem' },
        { name: '--space-8', value: '2rem' },
        { name: '--space-10', value: '2.5rem' },
        { name: '--space-12', value: '3rem' },
        { name: '--space-16', value: '4rem' },
      ]
      return { spaces }
    },
  }),
}

export const Badges: StoryObj = {
  render: () => ({
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 8px;">
        <span v-for="s in statuses" :key="s" :class="'badge badge-' + s">{{ s }}</span>
      </div>
    `,
    setup() {
      const statuses = ['active', 'completed', 'archived', 'draft', 'pending', 'answered', 'in_progress', 'deferred', 'skipped', 'blocked', 'failed', 'high', 'medium', 'low']
      return { statuses }
    },
  }),
}

export const Buttons: StoryObj = {
  render: () => ({
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 12px; align-items: center;">
        <button class="btn">Default</button>
        <button class="btn btn-primary">Primary</button>
        <button class="btn btn-danger">Danger</button>
        <button class="btn btn-sm">Small</button>
        <button class="btn btn-sm btn-primary">Small Primary</button>
      </div>
    `,
  }),
}

export const Cards: StoryObj = {
  render: () => ({
    template: `
      <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px;">
        <div class="card" style="padding: var(--space-5);">
          <h3 style="font-weight: 600; margin-bottom: var(--space-2);">Default Card</h3>
          <p style="color: var(--color-text-muted); font-size: var(--type-sm);">Card content with standard padding and border.</p>
        </div>
        <div class="card" style="padding: var(--space-5); border-color: var(--color-primary);">
          <h3 style="font-weight: 600; margin-bottom: var(--space-2);">Accent Card</h3>
          <p style="color: var(--color-text-muted); font-size: var(--type-sm);">Card with primary border accent.</p>
        </div>
      </div>
    `,
  }),
}

export const Tags: StoryObj = {
  render: () => ({
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 8px;">
        <span v-for="i in 6" :key="i" :class="'tag tag-hue-' + (i - 1)">tag-hue-{{ i - 1 }}</span>
      </div>
    `,
  }),
}
```
