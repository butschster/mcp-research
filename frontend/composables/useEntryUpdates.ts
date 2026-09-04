export interface EntryViewState {
  kind: 'new' | 'changed' | 'seen'
  current_revision: number
  seen_revision: number
  unseen_revisions: number
}

export interface EntryUpdate {
  entry_id: string
  entry_code?: string
  research_id: string
  section_id: string
  title: string
  description?: string
  entry_type: string
  status: string
  current_revision: number
  seen_revision: number
  unseen_revisions: number
  kind: 'new' | 'changed'
  updated_at: string
}

export interface EntryUpdatesEnvelope {
  entries: EntryUpdate[]
  new: number
  changed: number
  count: number
}

const emptyEntryUpdates: EntryUpdatesEnvelope = {
  entries: [],
  new: 0,
  changed: 0,
  count: 0,
}

/** Convert the personal queue to the lookup used by every entry-list surface. */
export function indexEntryUpdates(entries: EntryUpdate[]): Record<string, EntryUpdate> {
  return Object.fromEntries(entries.map((entry) => [entry.entry_id, entry]))
}

/** One owner for the personal queue request, its empty state and card lookup. */
export async function useEntryUpdates(researchId: string) {
  const request = await useApi<{ data: EntryUpdatesEnvelope }>(`/api/researches/${researchId}/updates`)
  const updates = computed<EntryUpdatesEnvelope>(() => request.data.value?.data ?? emptyEntryUpdates)
  const byEntry = computed(() => indexEntryUpdates(updates.value.entries))
  return { ...request, updates, byEntry }
}
