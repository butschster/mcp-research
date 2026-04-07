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
