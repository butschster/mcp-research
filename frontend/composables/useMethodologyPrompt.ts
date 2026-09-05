import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { useRuntimeConfig } from '#imports'

export interface MethodologyPromptSource {
  id?: string
  name: string
  slug: string
}

export function useMethodologyPrompt(source: MaybeRefOrGetter<MethodologyPromptSource | null | undefined>) {
  const config = useRuntimeConfig()

  return computed(() => {
    const methodology = toValue(source)
    if (!methodology) return ''
    const origin = typeof window === 'undefined' ? '' : window.location.origin
    if (!origin) return ''
    const base = String(config.public.apiBase || '').replace(/\/+$/, '')
    const docsUrl = new URL(`${base}/llms.txt`, origin).href

    // A team copy can share a slug with another methodology. Get accepts an
    // ID in its slug argument, so the pasted prompt selects this exact row.
    const selector = JSON.stringify({ slug: methodology.id || methodology.slug })
    return `Read ${docsUrl} for the Dovod server instructions.\n\n`
      + `Help me start a new project using the ${JSON.stringify(methodology.name)} methodology. `
      + `Load it with template_get using ${selector}, then follow its instructions. `
      + `When creating the project, set template_slug to ${JSON.stringify(methodology.slug)}.`
  })
}
