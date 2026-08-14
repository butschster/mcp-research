export default defineNuxtRouteMiddleware(async (to) => {
  const { authEnabled, isAuthenticated, fetchAuthInfo, checkAuth, loading } = useAuth()

  // Fetch auth info once
  if (authEnabled.value === null) {
    await fetchAuthInfo()
  }

  // Not an auth-enabled server — skip all checks
  if (!authEnabled.value) {
    return
  }

  // Check token validity if we haven't yet
  if (loading.value) {
    await checkAuth()
  }

  // An invitation is read before there is an account to read it with, so this
  // page renders either way: signed in it offers Join, signed out it offers a
  // way to get an account. It asks for a session only when Join is pressed.
  if (to.path.startsWith('/invite/')) {
    return
  }

  if (to.path === '/login' || to.path === '/register') {
    // Already signed in: go where they were headed, or home.
    if (isAuthenticated.value) {
      return navigateTo(safeNext(to.query.next) ?? '/')
    }
    return
  }

  // Protected page — must be authenticated. The destination rides along so
  // signing in returns here instead of dropping the reader on the research
  // list, which is how an invitation link gets lost between two clicks.
  if (!isAuthenticated.value) {
    return navigateTo(`/login?next=${encodeURIComponent(to.fullPath)}`)
  }
})

