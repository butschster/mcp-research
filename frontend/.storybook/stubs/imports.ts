import { ref, computed } from 'vue'

export {
  ref,
  computed,
  reactive,
  watch,
  watchEffect,
  onMounted,
  onUnmounted,
  onBeforeMount,
  onBeforeUnmount,
  nextTick,
  provide,
  inject,
  toRef,
  toRefs,
  unref,
  isRef,
  shallowRef,
  triggerRef,
  defineProps,
  defineEmits,
  defineExpose,
  withDefaults,
} from 'vue'

export { useRoute, useRouter } from 'vue-router'

// Nuxt-specific stubs
export const navigateTo = (path: string) => {
  console.log('[Storybook stub] navigateTo:', path)
}
export const useRuntimeConfig = () => ({
  public: { apiBase: '' },
})
export const useFetch = () => ({ data: ref(null), pending: ref(false), error: ref(null) })
export const useAsyncData = () => ({ data: ref(null), pending: ref(false), error: ref(null) })
export const useCookie = (_name: string) => ref('')
export const useHead = () => {}
export const useSeoMeta = () => {}
export const definePageMeta = () => {}

// Project composable stubs
export const useApi = (_url: any) => ({
  data: ref(null),
  pending: ref(false),
  error: ref(null),
  refresh: () => Promise.resolve(),
})

export const useAuth = () => ({
  user: ref(null),
  token: ref(null),
  authEnabled: ref(false),
  allowRegistration: ref(true),
  loading: ref(false),
  isAuthenticated: computed(() => false),
  fetchAuthInfo: () => Promise.resolve(),
  checkAuth: () => Promise.resolve(),
  login: () => Promise.resolve(),
  register: () => Promise.resolve(),
  logout: () => {},
  authFetch: (_url: string, _opts?: any) => Promise.resolve({} as any),
})

export const renderRefs = (text: string, _slug?: string) => text ?? ''
export const useCrossRefs = () => ({ renderRefs })

export const useRealtimeUpdates = (_handler?: any) => {}
export const useKeyboardNav = () => {}
