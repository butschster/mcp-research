import { defineComponent, h } from 'vue'

export const Handle = defineComponent({
  name: 'Handle',
  props: ['type', 'position', 'id'],
  setup(_, { slots }) {
    return () => h('div', { style: 'display:none' }, slots.default?.())
  },
})

export enum Position {
  Left = 'left',
  Top = 'top',
  Right = 'right',
  Bottom = 'bottom',
}
