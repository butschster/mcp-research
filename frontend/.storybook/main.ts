import type { StorybookConfig } from '@storybook/vue3-vite'
import { mergeConfig } from 'vite'
import path from 'path'

const config: StorybookConfig = {
  stories: ['../components/**/*.stories.ts'],
  addons: [
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
  ],
  framework: {
    name: '@storybook/vue3-vite',
    options: {},
  },
  docs: {
    autodocs: 'tag',
  },
  // staticDirs: ['../public'],
  viteFinal: async (config) => {
    const AutoImport = (await import('unplugin-auto-import/vite')).default
    const vue = (await import('@vitejs/plugin-vue')).default

    return mergeConfig(config, {
      resolve: {
        alias: {
          '~': path.resolve(__dirname, '..'),
          '@': path.resolve(__dirname, '..'),
          '#imports': path.resolve(__dirname, './stubs/imports.ts'),
          '#app': path.resolve(__dirname, './stubs/app.ts'),
          '#components': path.resolve(__dirname, './stubs/components.ts'),
        },
      },
      plugins: [
        vue(),
        AutoImport({
          imports: [
            'vue',
            'vue-router',
            {
              // `renderRefs` escapes its input; `linkRefs` is the variant for a
              // caller that already holds HTML. A story reaching for the wrong
              // one is exactly the mistake worth catching in the catalogue.
              [path.resolve(__dirname, '../composables/useCrossRefs')]: ['renderRefs', 'linkRefs'],
              [path.resolve(__dirname, '../utils/escapeHtml')]: ['escapeHtml'],
              [path.resolve(__dirname, '../composables/useTagHue')]: ['tagHue'],
              // Pure formatting over a string — the real one, so a card in the
              // catalogue reads the same "2h ago" the product does.
              [path.resolve(__dirname, '../composables/useRelativeTime')]: [
                'relativeTime',
                'absoluteTime',
                'parseTimestamp',
              ],
              // Module-scoped role state. The stories set it through the real
              // composable so a viewer story renders the viewer's card.
              [path.resolve(__dirname, '../composables/useResearchRole')]: ['useResearchRole'],
              // Where a link inside a research points. Real, not stubbed: half
              // the catalogue is components whose only job is to link, and a
              // stub would let a wrong href through the one place that checks.
              [path.resolve(__dirname, '../composables/useResearchPaths')]: [
                'researchPath',
                'entryPath',
                'sessionPath',
                'roadmapPath',
                'roadmapsPath',
                'tasksPath',
                'exportPath',
                'foreignResearchPath',
              ],
              // Module-scoped share state, read by the path helpers above, by
              // `renderRefs`, and by three components that render a target
              // outside the share as inert text. Stories set it through
              // `withShare()` in __mocks__/share.ts.
              [path.resolve(__dirname, '../composables/useShare')]: [
                'useShare',
                'shareActive',
                'shareToken',
                'shareInclude',
                'shareResearchCode',
              ],
              // Real composable, not a stub: it is module-scoped state plus
              // vue refs, so a component that raises a toast raises a real one.
              [path.resolve(__dirname, '../composables/useToasts')]: ['useToasts'],
              [path.resolve(__dirname, './stubs/imports')]: [
                'useApi',
                'useAuth',
                'useRuntimeConfig',
                'useRealtimeUpdates',
                'useKeyboardNav',
                'navigateTo',
                'useFetch',
                'useAsyncData',
                'useCookie',
                'useHead',
                'definePageMeta',
              ],
            },
          ],
          dts: false,
        }),
      ],
    })
  },
}
export default config
