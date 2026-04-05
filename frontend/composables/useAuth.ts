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

  async function register(email: string, password: string, name: string) {
    const res = await $fetch<{ user: User; token: string }>(`${baseURL}/api/auth/register`, {
      method: 'POST',
      body: { email, password, name },
    })
    user.value = res.user
    token.value = res.token
    localStorage.setItem('auth_token', res.token)
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('auth_token')
    navigateTo('/login')
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
  }
}
