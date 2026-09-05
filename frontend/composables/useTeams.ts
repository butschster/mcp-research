export type TeamRole = 'viewer' | 'editor' | 'owner'

export interface Team {
  id: string
  name: string
  personal: boolean
  role: TeamRole
  member_count: number
  research_count: number
}

export interface TeamMember {
  team_id: string
  user_id: string
  role: TeamRole
  email: string
  name: string
  created_at: string
}

export interface TeamInvite {
  id: string
  team_id: string
  email: string
  role: TeamRole
  inviter_name?: string
  inviter_email?: string
  expires_at: string
  created_at: string
}

/**
 * The teams the signed-in user belongs to.
 *
 * State is module-scoped because four screens want the same list — the research
 * list's filter, `/teams`, the Settings section and the transfer dialog — and a
 * per-component fetch would mean four requests to render one page.
 *
 * Everything here is inert when auth is off: there is no identity, so there is
 * nobody to have teams. Callers still get an empty list rather than an error,
 * so a component can render `v-if="teams.length > 1"` without branching on the
 * auth mode as well.
 */
const teams = ref<Team[]>([])
const loaded = ref(false)
const loading = ref(false)
const failed = ref(false)
let inFlight: Promise<void> | null = null

export function useTeams() {
  const config = useRuntimeConfig()
  const base = config.public.apiBase || ''
  const { authFetch, authEnabled } = useAuth()

  async function load(force = false) {
    if (!authEnabled.value) return
    if (loaded.value && !force) return
    // Four consumers mount at once on the research list; without this they
    // would each start their own request.
    if (inFlight) return inFlight

    loading.value = true
    failed.value = false
    inFlight = (async () => {
      try {
        const res = await authFetch<{ data: Team[] }>(`${base}/api/teams`)
        teams.value = res.data ?? []
        loaded.value = true
      } catch {
        failed.value = true
      } finally {
        loading.value = false
        inFlight = null
      }
    })()
    return inFlight
  }

  const refresh = () => load(true)

  /**
   * Forgets everything about the signed-in person.
   *
   * Called from `logout()`. Signing out is a client-side route change, so
   * without this the cache — and the team names in it — outlives the session
   * that fetched them.
   */
  function reset() {
    teams.value = []
    loaded.value = false
    loading.value = false
    failed.value = false
    inFlight = null
  }

  async function create(name: string) {
    const res = await authFetch<{ data: Team }>(`${base}/api/teams`, {
      method: 'POST',
      body: { name },
    })
    await refresh()
    return res.data
  }

  async function rename(id: string, name: string) {
    await authFetch(`${base}/api/teams/${id}`, { method: 'PUT', body: { name } })
    const team = teams.value.find((t) => t.id === id)
    if (team) team.name = name
  }

  async function remove(id: string) {
    await authFetch(`${base}/api/teams/${id}`, { method: 'DELETE' })
    await refresh()
  }

  return {
    teams: readonly(teams),
    loading: readonly(loading),
    loaded: readonly(loaded),
    failed: readonly(failed),
    load,
    refresh,
    reset,
    create,
    rename,
    remove,
    /** Teams the user may put a research into, for the transfer picker. */
    ownedTeams: computed(() => teams.value.filter((t) => t.role === 'owner')),
    personalTeam: computed(() => teams.value.find((t) => t.personal) ?? null),
    byId: (id: string) => teams.value.find((t) => t.id === id) ?? null,
  }
}

/** The one-line description of what a role may do, shown wherever one is picked. */
export const ROLE_DESCRIPTIONS: Record<TeamRole, string> = {
  viewer: 'Can read and export.',
  editor: 'Can read, and write documents, sessions, tasks and roadmaps.',
  owner: 'Everything an editor can do, plus managing members and the team.',
}

export const ROLE_LABELS: Record<TeamRole, string> = {
  viewer: 'Viewer',
  editor: 'Editor',
  owner: 'Owner',
}
