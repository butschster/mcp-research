import { addons } from '@storybook/manager-api'
import { themes } from '@storybook/theming'

addons.setConfig({
  theme: {
    ...themes.light,
    brandTitle: 'Dovod — Component Library',
    brandImage: '/brand/dovod.svg',
    appBg: '#f6f3ec',
    appContentBg: '#fffdf8',
    appBorderColor: '#d9ded0',
    barBg: '#fffdf8',
    textColor: '#243d32',
    textMutedColor: '#606c5a',
  },
})
