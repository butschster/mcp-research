export function useKeyboardNav() {
  if (import.meta.server) return

  function handleKey(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement).tagName
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(tag)) return

    const router = useRouter()

    switch (e.key) {
      case 'G':
        if (e.shiftKey) {
          // "Go home" means the research list, which a share visitor has no
          // account to see. One keystroke to the login wall is not a shortcut.
          if (shareActive()) return
          e.preventDefault()
          router.push('/')
        }
        break
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKey)
  })
  onUnmounted(() => {
    window.removeEventListener('keydown', handleKey)
  })
}
