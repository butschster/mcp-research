import type { Preview } from '@storybook/vue3'
import { setup } from '@storybook/vue3'
import { createMemoryHistory, createRouter } from 'vue-router'
import { INITIAL_VIEWPORTS } from '@storybook/addon-viewport'
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
  // Render Teleport content inline for modal stories
  app.component('Teleport', {
    template: '<slot />',
  })
})

const preview: Preview = {
  parameters: {
    backgrounds: { disable: true },
    layout: 'padded',
    viewport: {
      viewports: {
        mobile: {
          name: 'Mobile',
          styles: { width: '375px', height: '812px' },
        },
        mobileLarge: {
          name: 'Mobile Large',
          styles: { width: '414px', height: '896px' },
        },
        tablet: {
          name: 'Tablet',
          styles: { width: '768px', height: '1024px' },
        },
        desktop: {
          name: 'Desktop',
          styles: { width: '1280px', height: '800px' },
        },
        ...INITIAL_VIEWPORTS,
      },
    },
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
        <div style="background: var(--color-bg); color: var(--color-text); padding: 1.5rem; font-family: 'Outfit', system-ui, sans-serif;">
          <story />
        </div>
      `,
    }),
  ],
}
export default preview
