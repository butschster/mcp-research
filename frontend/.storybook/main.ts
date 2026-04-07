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
          '@vue-flow/core': path.resolve(__dirname, './stubs/vue-flow.ts'),
        },
      },
      plugins: [
        vue(),
        AutoImport({
          imports: ['vue', 'vue-router'],
          dts: false,
        }),
      ],
    })
  },
}
export default config
