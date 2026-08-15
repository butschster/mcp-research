interface User {
  id: string
  email: string
  name: string
}

interface AuthInfo {
  auth_enabled: boolean
  allow_registration: boolean
  auto_login_token?: string
}

const user = ref<User | null>(null)
const token = ref<string | null>(null)
const authEnabled = ref<boolean | null>(null)
const allowRegistration = ref(true)
const loading = ref(true)

export function useAuth() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || ''

  async function fetchAuthInfo() {
    try {
      const res = await $fetch<AuthInfo>(`${baseURL}/api/auth/info`)
      authEnabled.value = true
      allowRegistration.value = res.allow_registration

      // Auto-login when default_user is configured (local dev)
      if (res.auto_login_token && !localStorage.getItem('auth_token')) {
        token.value = res.auto_login_token
        localStorage.setItem('auth_token', res.auto_login_token)
      }
    } catch {
      authEnabled.value = false
    }
  }

  async function checkAuth() {
    loading.value = true
    const stored = localStorage.getItem('auth_token')
    if (!stored) {
      loading.value = false
      return
    }
    token.value = stored
    try {
      const res = await $fetch<User>(`${baseURL}/api/auth/me`, {
        headers: { Authorization: `Bearer ${stored}` },
      })
      user.value = res
    } catch {
      token.value = null
      localStorage.removeItem('auth_token')
    }
    loading.value = false
  }

  async function login(email: string, password: string) {
    const res = await $fetch<{ user: User; token: string }>(`${baseURL}/api/auth/login`, {
      method: 'POST',
      body: { email, password },
    })
    user.value = res.user
    token.value = res.token
    localStorage.setItem('auth_token', res.token)
  }

  /**
   * Creates an account. `inviteToken` lets someone who was handed a link
   * register on a server with registration closed, and joins them to the team
   * in the same request — the invitation is the authorization.
   */
  async function register(email: string, password: string, name: string, inviteToken?: string) {
    const res = await $fetch<{ user: User; token: string }>(`${baseURL}/api/auth/register`, {
      method: 'POST',
      body: { email, password, name, invite_token: inviteToken },
    })
    user.value = res.user
    token.value = res.token
    localStorage.setItem('auth_token', res.token)
  }

  function logout(next = '/login') {
    user.value = null
    token.value = null
    localStorage.removeItem('auth_token')
    // Module-scoped caches survive a client-side route change, so anything
    // keyed to the person has to be dropped here by name. Without this the
    // next account in the same tab renders the previous one's team names,
    // member counts and filter options.
    useTeams().reset()
    navigateTo(next)
  }

  // Authenticated $fetch wrapper for use outside useApi composable
  function authFetch<T>(url: string, opts: Record<string, any> = {}) {
    const headers: Record<string, string> = { ...(opts.headers || {}) }
    if (token.value) {
      headers['Authorization'] = `Bearer ${token.value}`
    }
    // Which tab is writing. The event that comes back carries it, which is how
    // this tab knows not to refetch a change already on its screen.
    if (!import.meta.server) {
      headers['X-Client-Id'] = realtimeClientId()
    }
    return $fetch<T>(url, { ...opts, headers })
  }

  return {
    user: readonly(user),
    token: readonly(token),
    authEnabled: readonly(authEnabled),
    allowRegistration: readonly(allowRegistration),
    loading: readonly(loading),
    isAuthenticated: computed(() => !!user.value),
    fetchAuthInfo,
    checkAuth,
    login,
    register,
    logout,
    authFetch,
  }
}
