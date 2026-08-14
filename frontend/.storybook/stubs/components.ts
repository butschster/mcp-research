import { defineComponent, h } from 'vue'

/**
 * Stand-in for Nuxt's `#components` virtual module.
 *
 * Most components resolve `NuxtLink` by name from their own template, which the
 * global registration in `preview.ts` covers. A component that binds it to a
 * dynamic `:is` needs the component object itself and imports it — there is no
 * virtual module behind that import in Storybook, so it resolves here.
 *
 * Same anchor markup as the global stub, so a row renders identically whether
 * it was resolved by name or by import.
 */
export const NuxtLink = defineComponent({
  name: 'NuxtLink',
  props: {
    to: { type: [String, Object], default: undefined },
    href: { type: String, default: undefined },
  },
  setup(props, { slots }) {
    return () => h('a', { href: (props.to as string) || props.href }, slots.default?.())
  },
})
