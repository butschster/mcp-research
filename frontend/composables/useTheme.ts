import { readonly, ref } from 'vue'

export type Theme = 'light' | 'dark'
export const THEME_STORAGE_KEY = 'dovod-theme'
const theme = ref<Theme>(typeof document !== 'undefined' && document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light')

/** Apply the palette before notifying canvas/diagram renderers. */
export function setTheme(value: Theme, persist = true) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.theme = value
  theme.value = value
  const background = getComputedStyle(document.documentElement).getPropertyValue('--color-bg').trim()
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', background)
  if (persist) {
    try { localStorage.setItem(THEME_STORAGE_KEY, value) } catch { /* Private browsing can refuse storage. */ }
  }
  window.dispatchEvent(new Event('dovod:theme-change'))
}

if (typeof window !== 'undefined') {
  window.addEventListener('storage', (event) => {
    if (event.key === THEME_STORAGE_KEY || event.key === null) {
      setTheme(event.newValue === 'dark' ? 'dark' : 'light', false)
    }
  })
}

export function useTheme() {
  return {
    theme: readonly(theme),
    setTheme,
    toggleTheme: () => setTheme(theme.value === 'light' ? 'dark' : 'light'),
  }
}
