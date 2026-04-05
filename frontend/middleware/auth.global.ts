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

  const publicPages = ['/login', '/register']
  if (publicPages.includes(to.path)) {
    // If already logged in, redirect away from auth pages
    if (isAuthenticated.value) {
      return navigateTo('/')
    }
    return
  }

  // Protected page — must be authenticated
  if (!isAuthenticated.value) {
    return navigateTo('/login')
  }
})
