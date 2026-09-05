<template>
  <section
    v-if="!hidden"
    ref="root"
    class="card card--list resume-block"
    :aria-busy="loading || refreshing ? 'true' : 'false'"
    :aria-label="archived ? 'Left unfinished' : 'Continue'"
    @keydown.esc="onEscape"
  >
    <div class="list-head resume-head">
      <div class="resume-head-left">
        <h2 class="resume-heading">
          <button
            type="button"
            class="resume-disclosure"
            :aria-expanded="open ? 'true' : 'false'"
            :aria-controls="bodyId"
            @click="toggle"
          >
            <svg class="resume-caret" :class="{ 'is-open': open }" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="9 18 15 12 9 6"/></svg>
            {{ archived ? 'Left unfinished' : 'Continue' }}
          </button>
        </h2>
        <p v-if="archived" class="resume-lead">This project is archived. Unfinished items are saved here for reference.</p>
        <!-- Collapsing must not hide the news. The counts move into the head as
             plain text — the same numbers, without the controls. There is no
             "as of" line on a fresh load: it always read "just now", which told
             a reader nothing and could be mistaken for the research having
             moved just now. The refresh error below dates a stale picture. -->
        <p v-if="!open && ledger.length" class="resume-lead resume-collapsed-counts">
          <span v-for="group in ledger" :key="group.key" class="resume-collapsed-count">
            {{ group.label }} {{ group.total === null ? `${group.returned}+` : group.total }}
          </span>
        </p>
      </div>

      <div class="resume-head-right cluster">
        <span class="sr-only" role="status">{{ copyAnnouncement }}</span>
        <!-- One session: a link, not a control. Several: a picker. -->
        <NuxtLink
          v-if="onlySession"
          :to="sessionPath(researchSlug, onlySession.code || onlySession.id)"
          class="resume-session-chip"
        >
          <ShortCode v-if="onlySession.code" :code="onlySession.code" />
          <span class="resume-session-title">{{ onlySession.title }}</span>
          <StatusBadge :status="onlySession.status" />
        </NuxtLink>
        <!-- With several sessions the picker chooses and the chip beside it
             opens. Removing the sessions grid moved that link here, and a
             picker alone would have left a multi-session research with no
             one-click way into the session at all. -->
        <template v-else-if="sessions.length > 1">
          <select
            class="select resume-session-select"
            :value="summary?.sessions?.selected_id || ''"
            aria-label="Session this summary is about"
            @change="$emit('select-session', ($event.target as HTMLSelectElement).value)"
          >
            <option v-if="selectionRequired" value="" disabled>Choose a session…</option>
            <!-- The status is in the label because this is the choice being
                 made: "2 sessions are active" above a list of six unlabelled
                 options is not a question anybody can answer. -->
            <option v-for="s in sessions" :key="s.id" :value="s.id">
              {{ sessionOptionLabel(s) }}
            </option>
          </select>
          <NuxtLink
            v-if="selectedSession"
            :to="sessionPath(researchSlug, selectedSession.code || selectedSession.id)"
            class="btn btn-icon"
            :title="`Open ${selectedSession.code || selectedSession.title}`"
            :aria-label="`Open session ${selectedSession.code || selectedSession.title}`"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h13"/><path d="m12 5 7 7-7 7"/></svg>
          </NuxtLink>
        </template>

        <!-- The one thing this block owes a person: the words that start a new
             chat about this research. An icon, not the sentence itself — the
             sentence is for pasting, not for reading, and printing it in the
             head spent 200px on text nobody needs on screen.

             What is copied carries the guide's URL, because the reader is not
             the audience for it either: the agent on the other end is, and a
             link it can open beats a link a person is asked to read first. -->
        <button
          v-if="summary && canWrite && !archived"
          type="button"
          class="btn btn-icon resume-handoff"
          :class="{ 'is-copied': copied }"
          :title="handoffTitle"
          :aria-label="handoffTitle"
          @click="copyHandoff"
        >
          <svg v-if="copied" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="20 6 9 17 4 12"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
        </button>
        <button
          type="button"
          class="btn btn-icon"
          title="Refresh the summary"
          aria-label="Refresh the summary"
          :disabled="refreshing"
          @click="$emit('refresh')"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/></svg>
        </button>
      </div>
    </div>

    <div v-show="open" :id="bodyId">
      <!-- First load. The skeletons are row-height so nothing jumps when the
           real rows arrive. -->
      <!-- Sized to the rows they stand in for: an action row is a title over a
           reason line, not a single line, and three skeletons a line short each
           moved the whole page down by 73px when the data landed. The last one
           is the ledger. -->
      <div v-if="loading" class="resume-skeletons">
        <div v-for="i in 3" :key="i" class="skeleton-card resume-skeleton-row"></div>
        <div class="skeleton-card resume-skeleton-ledger"></div>
      </div>

      <EmptyState
        v-else-if="error && !summary"
        icon="&#x26A0;&#xFE0F;"
        title="Could not load the summary"
        description="Nothing was changed. The documents below are unaffected."
        role="alert"
      >
        <button type="button" class="btn btn-sm" @click="$emit('refresh')">Try again</button>
      </EmptyState>

      <template v-else-if="summary">
        <p v-if="selectionRequired" class="resume-note" role="status">
          {{ summary.sessions.active_count }} sessions are active. Choose one to see its questions.
        </p>
        <!-- The last session is shown with its real status rather than an empty
             head. Starting a new one is a write, and stays the agent's move. -->
        <p v-else-if="noSessionRunning" class="resume-note">
          No active session. Ask your AI assistant to start one.
        </p>

        <!-- Next up: the only thing the eye should land on. -->
        <div v-if="actions.length" class="data-rows">
          <ResearchResumeRow
            v-for="(action, i) in actions"
            :key="`${action.reason_code}-${action.target.id || i}`"
            variant="action"
            :code="action.target.code"
            :title="actionTitle(action)"
            :href="actionHref(action)"
            :actor="archived ? undefined : action.actor"
            :reason="action.reason"
          />
        </div>

        <EmptyState
          v-else-if="nothingWaiting"
          icon="&#x2713;"
          title="Nothing waiting"
          :description="emptyDescription"
          :command="canWrite && !archived ? `Continue ${summary.research.code}` : undefined"
        >
          <NuxtLink class="btn btn-sm" :to="`/research/${researchSlug}/sessions`">Past sessions</NuxtLink>
        </EmptyState>

        <!-- Work exists but none of it is a candidate — every open item is
             deferred, or waiting on somebody else. Without this the card was a
             head above a lone counter, with nothing saying why. -->
        <p v-else class="resume-note">
          Nothing is queued for the agent right now. The counters below are what is still open.
        </p>

        <!-- The ledger. A group with nothing in it is not rendered: a row of
             zeroes is furniture, and it hides the counts that are not zero. -->
        <!-- Only the open panel exists in the DOM, so only the open counter may
             name one: an `aria-controls` pointing at a missing id sends a
             screen reader nowhere. -->
        <div v-if="ledger.length" class="data-row resume-ledger">
          <button
            v-for="group in ledger"
            :key="group.key"
            type="button"
            class="resume-chip"
            :class="[`resume-chip--${group.tone}`, { 'is-open': openGroup === group.key }]"
            :ref="el => registerChip(group.key, el)"
            :aria-expanded="openGroup === group.key ? 'true' : 'false'"
            :aria-controls="openGroup === group.key ? panelId : undefined"
            @click="toggleGroup(group.key)"
          >
            {{ group.label }}
            <!-- "n+" when the count is unknown: it is what we can prove, and a
                 total we did not get is not a total of n. -->
            <span class="count-chip">{{ group.total === null ? `${group.returned}+` : group.total }}</span>
          </button>
        </div>

        <!-- One group open at a time. Eight groups of five rows would turn the
             research page into a dashboard and push the documents off screen. -->
        <div v-if="openGroupData" :id="panelId" class="resume-panel">
          <div class="resume-panel-head">
            <span class="resume-panel-title">{{ openGroupData.label }}</span>
            <NuxtLink v-if="openGroupData.seeAll" :to="openGroupData.seeAll" class="btn btn-sm">
              {{ openGroupData.total === null ? 'See all' : `See all ${openGroupData.total}` }}
            </NuxtLink>
          </div>
          <div v-if="openGroupData.rows.length" class="data-rows">
            <ResearchResumeRow
              v-for="row in openGroupData.rows"
              :key="row.key"
              variant="item"
              :code="row.code"
              :title="row.title"
              :href="row.href"
              :status="row.status"
              :priority="row.priority"
              :note="row.note"
              :meta="row.meta"
              :update-kind="row.updateKind"
              :unseen-revisions="row.unseenRevisions"
            />
          </div>
          <p v-else class="list-empty">{{ openGroupData.empty }}</p>
        </div>

        <p v-if="error" class="inline-error inline-error--action resume-refresh-error" role="status">
          <span>Could not refresh — this is the picture from {{ generatedLabel }}.</span>
          <button type="button" class="btn btn-sm" @click="$emit('refresh')">Try again</button>
        </p>
        <!-- The clipboard was refused, so the words are put on the page where
             they can be selected by hand. -->
        <p v-if="copyFailed" class="inline-error inline-error--action resume-refresh-error" role="status">
          <span>Could not copy. Select this and copy it: <code>{{ handoffCommand }}</code></span>
          <button type="button" class="btn btn-sm" @click="dismissCopyFailure">Dismiss</button>
        </p>
        <p v-if="summary.truncated" class="resume-truncated">
          Some details were left out to keep this summary small.
        </p>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useCopyToClipboard } from '~/composables/useCopyToClipboard'

/**
 * "Continue" — what is unfinished in this research and what to do next.
 *
 * It replaces the active-sessions grid when it renders, because two stacked
 * rows of session cards give the page two focal points and neither wins. The
 * head therefore always names the selected session and links to it: that is
 * not decoration, it is the grid's job moved here.
 *
 * Nothing in this block writes. It never marks a document seen, never moves a
 * status, never starts a session — the personal new/changed queue stays a
 * separate request precisely so that reading this cannot acknowledge anything.
 */
import { relativeTime } from '~/composables/useRelativeTime'
import { authorKind } from '~/composables/useAuthorKind'
import { annotationsPath, entryPath, researchPath, sessionPath, taskPath, tasksPath } from '~/composables/useResearchPaths'
import type { EntryUpdate } from '~/composables/useEntryUpdates'
import type { ResumeAction, ResumeSummary } from '~/composables/useResearchResume'

const props = defineProps<{
  summary: ResumeSummary | null
  researchSlug: string
  loading?: boolean
  refreshing?: boolean
  error?: string | null
  /** Archived or completed: the same content, read as history rather than a queue. */
  archived?: boolean
  /** The reader's own unseen-document markers, for the Changed group only. */
  updatesByEntry?: Record<string, EntryUpdate>
  canWrite?: boolean
}>()

defineEmits<{ refresh: []; 'select-session': [string] }>()

const open = ref(true)
const openGroup = ref('')

// Unique per instance: the autodocs page renders every story of this component
// at once, and a hard-coded id there means seventeen elements share it and each
// aria-controls resolves to the first.
const uid = useId()
const bodyId = `resume-body-${uid}`
const panelId = computed(() => `resume-group-${uid}-${openGroup.value}`)

const STORAGE_KEY = 'research_resume_open'

onMounted(() => {
  // One key for the whole product, not one per research: "I do not want this
  // block" is a fact about the person, and per-research keys grow without end.
  try {
    open.value = localStorage.getItem(STORAGE_KEY) !== 'closed'
  } catch {
    open.value = true
  }
  // An archived research opens collapsed without overwriting that preference.
  if (props.archived) open.value = false
})

function toggle() {
  open.value = !open.value
  // An archived research opens collapsed for this render only, so toggling it
  // must not decide how every other research renders tomorrow.
  if (props.archived) return
  try {
    localStorage.setItem(STORAGE_KEY, open.value ? 'open' : 'closed')
  } catch {
    // A browser that refuses storage still gets a working disclosure.
  }
}

function toggleGroup(key: string) {
  openGroup.value = openGroup.value === key ? '' : key
}

const root = ref<HTMLElement | null>(null)
const chips = new Map<string, HTMLElement>()

function registerChip(key: string, el: unknown) {
  if (el instanceof HTMLElement) chips.set(key, el)
  else chips.delete(key)
}

/**
 * Escape closes an open group and puts focus back on the counter that opened
 * it — the panel is where the focus came from, and leaving a keyboard user at
 * the top of the document is the failure this prevents.
 *
 * With nothing open it does nothing on purpose. Collapsing the whole block on a
 * stray Escape is a surprise, and there is no focus trap here to escape from.
 */
function onEscape() {
  if (!openGroup.value) return
  const key = openGroup.value
  openGroup.value = ''
  nextTick(() => chips.get(key)?.focus())
}

const selectionRequired = computed(() => !!props.summary?.sessions?.selection_required)
const sessions = computed(() => props.summary?.sessions?.items ?? [])
/** Sessions exist, but none is open — the head shows the last one as it stands. */
const noSessionRunning = computed(
  () => !!props.summary && sessions.value.length > 0 && props.summary.sessions.active_count === 0,
)

/** The one session, when there is exactly one: a link, not a control. */
const onlySession = computed(() => (sessions.value.length === 1 ? sessions.value[0] : null))
const selectedSession = computed(
  () => sessions.value.find((s) => s.id === props.summary?.sessions?.selected_id) ?? null,
)

/**
 * What to type into a new chat to pick this up.
 *
 * It names the session when one is selected, because that is the choice the
 * server refuses to make on its own — handing over "Continue R1" alone on a
 * research with three open threads just moves the ambiguity into the chat.
 */
const handoffCommand = computed(() => {
  const code = props.summary?.research.code || props.researchSlug
  const session = selectedSession.value?.code
  const what = session ? `Continue ${code}, session ${session}.` : `Continue ${code}.`
  // The instruction and the place it is written down, in one paste.
  return `${what} How to pick it up: ${handoffGuide.value}`
})

const handoffTitle = computed(() => `Copy a prompt to continue this project: ${handoffCommand.value}`)

/**
 * The guide, absolutely. A relative path is meaningless once the sentence has
 * been pasted into a chat somewhere else, and that is the only place it goes.
 * It is served by this binary, so the link always matches the running version.
 */
const handoffGuide = computed(() => {
  const origin = import.meta.client ? window.location.origin : ''
  return `${origin}/llms/conducting-research.md#picking-up-a-research-that-is-already-running`
})

const { copied, failed: copyFailed, announcement: copyAnnouncement, copy, dismiss: dismissCopyFailure } = useCopyToClipboard()

function copyHandoff() {
  return copy(handoffCommand.value, {
    success: `Copied “${handoffCommand.value}” to the clipboard`,
    // A refused clipboard leaves nothing on screen to fall back to now that the
    // button is an icon, so the sentence itself becomes the message — and it
    // stays until dismissed, because the reader has to select it by hand.
    failure: `Could not copy. The sentence is: ${handoffCommand.value}`,
  })
}


function sessionOptionLabel(s: { code?: string; title: string; status: string }) {
  const name = s.code ? `${s.code} — ${s.title}` : s.title
  return s.status === 'active' ? name : `${name} (${s.status})`
}
// Only ever shown to date a picture that failed to refresh. On a fresh load the
// answer is always "just now", which is not information.
const generatedLabel = computed(() =>
  props.summary?.generated_at ? relativeTime(props.summary.generated_at) : '',
)
const actions = computed(() => {
  const all = props.summary?.next_actions ?? []
  // The server proposes "choose a session" as an action because an agent has
  // nowhere else to be told. On screen the picker and the note above already
  // say it, and a row repeating them is the same sentence three times.
  return all.filter((a) => !(selectionRequired.value && a.kind === 'choose_session')).slice(0, 3)
})

// "Nothing waiting" is a claim, and it is only safe to make when every group
// reported a real count. A group whose total is unknown, or one carrying items
// with no total, means the summary does not know — and printing "Nothing
// waiting" over outstanding work is the one wrong answer this block can give.
const nothingWaiting = computed(() => {
  const s = props.summary
  if (!s) return false
  // `recent_entries` is deliberately not in this: a document that changed is
  // news, not a queue, and counting it as outstanding work would mean a
  // research where everything is finished never gets to say so.
  return workGroupsOf(s).every((g) => g.total === 0 && g.returned === 0)
})

// An archived research with nothing left needs no card saying so — unless the
// server still proposes something, in which case hiding the card would drop it.
const hidden = computed(() => props.archived && nothingWaiting.value && actions.value.length === 0)

const emptyDescription = computed(() =>
  props.canWrite
    ? 'No task is in progress, nothing is blocked, and no mark is waiting for you. New work shows up here as your agent does it.'
    : "No task is in progress, nothing is blocked, and no mark is waiting for you. New work shows up here as an editor's agent does it.",
)

/** The groups that represent outstanding work — everything a person could act on. */
function workGroupsOf(s: ResumeSummary) {
  return [
    s.work.in_progress, s.work.blocked, s.work.pending,
    s.questions.open, s.questions.deferred,
    s.annotations.to_work, s.annotations.awaiting_human,
  ]
}

const collapsedSummary = computed(() => {
  const s = props.summary
  if (!s) return ''
  const parts = ledger.value.map((g) => `${g.label} ${g.total === null ? `${g.returned} or more` : g.total}`)
  return parts.length ? `— ${parts.join(', ')}` : '— nothing waiting'
})

interface LedgerGroup {
  key: string
  label: string
  /** Null when the server could not count — rendered as "n+", never as a total. */
  total: number | null
  returned: number
  tone: 'plain' | 'error' | 'warning'
}

function counts(g: { total: number | null; returned: number }) {
  return { total: g.total, returned: g.returned }
}

const ledger = computed<LedgerGroup[]>(() => {
  const s = props.summary
  if (!s) return []
  const groups: LedgerGroup[] = [
    { key: 'in_progress', label: 'In progress', ...counts(s.work.in_progress), tone: 'plain' },
    { key: 'blocked', label: 'Blocked', ...counts(s.work.blocked), tone: 'error' },
    // "Pending", not "Todo": the board's Todo column also holds blocked and
    // deferred tasks, so borrowing its word here put a 5 under a heading the
    // destination shows as 7.
    { key: 'pending', label: 'Pending', ...counts(s.work.pending), tone: 'plain' },
    { key: 'questions', label: 'Questions', ...counts(s.questions.open), tone: 'plain' },
    { key: 'deferred', label: 'Deferred', ...counts(s.questions.deferred), tone: 'plain' },
    { key: 'marks', label: 'Marks', ...counts(s.annotations.to_work), tone: 'plain' },
    { key: 'awaiting', label: 'Awaiting you', ...counts(s.annotations.awaiting_human), tone: 'warning' },
    { key: 'changed', label: 'Changed', ...counts(s.recent_entries), tone: 'plain' },
  ]
  return groups.filter((g) => {
    // Questions belong to a session; with none chosen there is no queue to
    // count, and a counter that opens an empty panel teaches distrust.
    if ((g.key === 'questions' || g.key === 'deferred') && selectionRequired.value) return false
    // An unknown total with items in hand is still a group worth offering: the
    // alternative is hiding work because the count failed.
    return g.total === null ? g.returned > 0 : g.total > 0
  })
})

interface PanelRow {
  key: string
  code?: string
  title: string
  href: string
  status?: string
  priority?: string
  note?: string
  meta?: string
  updateKind?: 'new' | 'changed'
  unseenRevisions?: number
}

const openGroupData = computed(() => {
  const s = props.summary
  if (!s || !openGroup.value) return null
  const slug = props.researchSlug
  const marksHref = annotationsPath(slug)
  const sessionCode = sessions.value.find((x) => x.id === s.sessions.selected_id)?.code || ''

  const taskRows = (items: typeof s.work.pending.items): PanelRow[] =>
    items.map((t) => ({
      key: t.id, code: t.code, title: plainText(t.title), href: taskPath(slug, t.code || t.id),
      status: t.status, priority: t.priority, note: t.note,
    }))
  const questionRows = (items: typeof s.questions.open.items): PanelRow[] =>
    items.map((q) => ({
      key: q.id, code: q.code, title: plainText(q.text), status: q.status, priority: q.priority,
      href: q.session_code
        ? `/research/${slug}/session/${q.session_code}/question/${q.code || q.id}`
        : sessionPath(slug, sessionCode || q.session_id),
    }))
  const markRows = (items: typeof s.annotations.to_work.items): PanelRow[] =>
    items.map((a) => ({
      key: a.id, code: a.code, title: plainText(a.quote || a.entry_title) || 'A mark',
      href: a.entry_code ? entryPath(slug, a.entry_code) : marksHref || researchPath(slug),
      status: a.status, meta: a.entry_code,
    }))

  const map: Record<string, { label: string; total: number | null; seeAll: string; empty: string; rows: PanelRow[] }> = {
    in_progress: {
      label: 'In progress', total: s.work.in_progress.total, seeAll: tasksPath(slug),
      empty: 'Nothing is in progress.', rows: taskRows(s.work.in_progress.items),
    },
    blocked: {
      label: 'Blocked', total: s.work.blocked.total, seeAll: tasksPath(slug),
      empty: 'Nothing is blocked.',
      // A blocker nobody wrote down is a fact worth printing. Left blank it
      // reads as "no reason needed", and the reader clicks through to find
      // an empty field.
      rows: taskRows(s.work.blocked.items).map((row) => ({ ...row, note: row.note || 'no reason recorded' })),
    },
    pending: {
      label: 'Pending', total: s.work.pending.total, seeAll: tasksPath(slug),
      empty: 'The board is clear.', rows: taskRows(s.work.pending.items),
    },
    questions: {
      label: 'Questions', total: s.questions.open.total,
      seeAll: sessionCode ? sessionPath(slug, sessionCode) : `/research/${slug}/sessions`,
      empty: 'Every question in this session has an answer.', rows: questionRows(s.questions.open.items),
    },
    deferred: {
      label: 'Deferred', total: s.questions.deferred.total,
      seeAll: sessionCode ? sessionPath(slug, sessionCode) : `/research/${slug}/sessions`,
      empty: 'Nothing deferred.', rows: questionRows(s.questions.deferred.items),
    },
    marks: {
      label: 'Marks', total: s.annotations.to_work.total, seeAll: marksHref,
      empty: 'No mark is waiting for the agent.', rows: markRows(s.annotations.to_work.items),
    },
    awaiting: {
      label: 'Awaiting you', total: s.annotations.awaiting_human.total,
      seeAll: marksHref ? `${marksHref}?status=answered` : '',
      empty: 'Nothing is waiting on you.', rows: markRows(s.annotations.awaiting_human.items),
    },
    changed: {
      label: 'Changed', total: s.recent_entries.total,
      seeAll: `${researchPath(slug)}?section=__all__`,
      empty: 'No document changed recently.',
      rows: s.recent_entries.items.map((e) => {
        const update = props.updatesByEntry?.[e.id]
        return {
          key: e.id, code: e.code, title: plainText(e.title) || e.code, href: entryPath(slug, e.code || e.id),
          // Who wrote it last, in the product's one vocabulary for authorship —
          // a person's correction is the row an agent must not mistake for its
          // own stale draft, and a fifth phrasing of "a person" is how those
          // four words drift apart.
          note: e.author_kind === 'human' ? `edited by ${authorKind(e.author_kind).word}` : undefined,
          meta: relativeTime(e.updated_at),
          updateKind: update?.kind,
          unseenRevisions: update?.unseen_revisions,
        }
      }),
    },
  }
  return map[openGroup.value] ?? null
})

// The row's title is what the reader is being asked to do. For most actions
// that is the thing itself — a task's title, a question's text. For the session
// choice it is not: the target is the research, and printing the research name
// as the row title would read as "open this research", which is where they
// already are.
function actionTitle(action: ResumeAction): string {
  if (action.kind === 'choose_session') return 'Choose which session you are continuing'
  // Never the raw `kind`: `continue_task` printed as a sentence is a leaked
  // enum. A target with no title still has a code, which the row also shows.
  return plainText(action.target.title) || action.target.code || 'Open it'
}

/**
 * Cross-references, as text.
 *
 * Titles and question text are agent-written and carry `[[E3]]`, which every
 * other surface renders as a link. This row is a single link end to end, so a
 * nested anchor is not available — but printing the brackets is wrong twice
 * over, so the markers are dropped and the code left as the word it is.
 */
function plainText(value: string | undefined): string {
  return (value ?? '').replace(/\[\[([^\]]+)\]\]/g, '$1')
}

// Every href is built here rather than taken from the payload, so links go
// through the same helpers as the rest of the product and keep working under a
// share or a route change.
function actionHref(action: ResumeAction): string {
  const slug = props.researchSlug
  const t = action.target
  switch (t.type) {
    case 'task':
      return taskPath(slug, t.code || t.id || '')
    case 'annotation':
      return t.entry_code ? entryPath(slug, t.entry_code) : annotationsPath(slug) || researchPath(slug)
    case 'question':
      return t.session_code
        ? `/research/${slug}/session/${t.session_code}/question/${t.code || t.id}`
        : `/research/${slug}/sessions`
    case 'entry':
      return entryPath(slug, t.code || t.id || '')
    case 'session':
      return sessionPath(slug, t.code || t.id || '')
    default:
      return `/research/${slug}/sessions`
  }
}
</script>

<style scoped>
.resume-block { margin-bottom: var(--space-4); }
/* The list card rounds its first row's corners, which is right when the rows
   are the top of the card and wrong here: the head is above them, so a hovered
   first row grew two 9px corners in the middle of the card. */
.resume-block :deep(.data-rows:first-child > .data-row:first-child) {
  border-start-start-radius: 0;
  border-start-end-radius: 0;
}

/* Tighter than `.list-head`'s own 24/16. That padding is right for a card whose
   heading introduces a page of rows; this heading is one line above three, and
   the block sits between the research title and the documents — the most
   expensive space on the page. */
.resume-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  padding-bottom: var(--space-3);
}
.resume-head-left { min-width: 0; }
.resume-head-right { flex: none; }

.resume-heading { margin: 0; font-size: var(--type-md); font-weight: var(--weight-semibold); }
.resume-disclosure {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-h);
  padding: 0;
  border: none;
  background: none;
  color: var(--color-text);
  font: inherit;
  cursor: pointer;
}
.resume-caret { transition: transform var(--transition-fast); }
.resume-caret.is-open { transform: rotate(90deg); }

.resume-lead { font-size: var(--type-xs); color: var(--color-text-faint); }
.resume-collapsed-counts { display: flex; flex-wrap: wrap; gap: var(--space-3); margin-top: var(--space-1); }
.resume-collapsed-count { white-space: nowrap; }
.resume-session-open { font-size: var(--type-xs); white-space: nowrap; }
/* The one control in the head that is primary-coloured: it is the thing this
   block is for. The tick after a copy is the only feedback, so it holds the
   colour too. */
.resume-handoff { color: var(--color-primary); }
.resume-handoff.is-copied { color: var(--color-success); border-color: var(--color-success); }
.resume-note { padding: var(--space-3) var(--row-inset); font-size: var(--type-sm); color: var(--color-text-muted); }

.resume-session-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-h);
  max-width: 22rem;
  color: inherit;
  text-decoration: none;
}
.resume-session-chip:hover { text-decoration: none; color: var(--color-primary); }
.resume-session-title { font-size: var(--type-sm); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.resume-session-select { max-width: 22rem; }

.resume-skeletons { padding: var(--space-3) var(--row-inset); }
/* Exactly a row: one line of content in the row's own padding. */
.resume-skeleton-row {
  height: calc(var(--control-h) + var(--space-6));
  margin-bottom: var(--space-2);
}
.resume-skeleton-ledger { height: var(--control-h); }

/* The ledger is a sibling of the rows container, not a sibling row, so the
   `.data-row + .data-row` rule never reaches it and it merged into the last
   action row. */
.resume-ledger { display: flex; flex-wrap: wrap; gap: var(--space-3); border-top: 1px solid var(--color-border); }
.resume-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-h);
  padding: 0 var(--space-2);
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: none;
  color: var(--color-text-muted);
  font: inherit;
  font-size: var(--type-xs);
  cursor: pointer;
}
/* Raised, not hover-coloured: the count pill inside is already
   `--color-surface-hover`, so an open chip painted the same colour swallowed
   its own number. */
.resume-chip:hover,
.resume-chip.is-open { color: var(--color-text); background: var(--color-surface-raised); }
.resume-chip.is-open { border-color: var(--color-border-strong); }
/* The hue is a second signal; the label is the first. It has to survive hover
   and the open state — a warning that goes grey when you point at it is a
   warning that disappears exactly when it is being read. */
.resume-chip--error,
.resume-chip--error:hover,
.resume-chip--error.is-open { color: var(--color-error); }
.resume-chip--warning,
.resume-chip--warning:hover,
.resume-chip--warning.is-open { color: var(--color-warning); }

.resume-panel { border-top: 1px solid var(--color-border); }
.resume-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--row-inset);
}
.resume-panel-title { font-size: var(--type-sm); font-weight: var(--weight-semibold); }

.resume-refresh-error { margin: 0; padding: var(--space-3) var(--row-inset); }
.resume-truncated { padding: var(--space-3) var(--row-inset); font-size: var(--type-xs); color: var(--color-text-faint); }

@media (max-width: 768px) {
  .resume-head-right { width: 100%; }
  /* The select was the one control in the head left at the desktop height,
     beside a 36px refresh button it then sat 3px inside. */
  .resume-session-select { flex: 1; max-width: none; height: var(--control-h-touch); }
  /* Every control in the head grows together. The disclosure and the session
     link are the two a thumb reaches for most, and they were the two left at
     the desktop height beside a 36px refresh button. */
  .resume-chip,
  .resume-disclosure,
  .resume-session-chip { min-height: var(--control-h-touch); }
}
</style>
