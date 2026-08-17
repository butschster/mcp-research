<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="research" class="sessions-page">
    <PageHeader
      :crumbs="[
        { label: 'Research', to: '/' },
        { label: research.name, to: `/research/${researchSlug}` },
        { label: 'Sessions' },
      ]"
      :code="research.code"
      title="Sessions"
      :count="sessions.length"
    />

    <p class="lead">
      Every interview this research has held. A session carries the questions that were
      asked and the answers they got, so a finished one is the record of how something
      came to be known — not a thing to be tidied away.
    </p>

    <template v-if="sessions.length">
      <div v-if="active.length" class="card card--list">
        <div class="list-head">
          <h2 class="card-section-title">
            Open <span class="heading-note">{{ active.length }}</span>
          </h2>
          <p class="group-blurb">Still running. An agent picks these up rather than starting again.</p>
        </div>
        <div class="data-rows">
          <NuxtLink
            v-for="sess in active"
            :key="sess.id"
            :to="sessionPath(researchSlug, sess.code || sess.id)"
            class="data-row session-row"
          >
            <span class="session-main">
              <span class="session-line">
                <span v-if="sess.code" class="short-code">{{ sess.code }}</span>
                <span class="session-title">{{ sess.title }}</span>
              </span>
              <span v-if="sess.focus" class="session-focus">{{ sess.focus }}</span>
            </span>
            <StatusBadge :status="sess.status" />
          </NuxtLink>
        </div>
      </div>

      <div v-if="closed.length" class="card card--list">
        <div class="list-head">
          <h2 class="card-section-title">
            Finished <span class="heading-note">{{ closed.length }}</span>
          </h2>
          <p class="group-blurb">Kept, not archived. Open one to read what was asked and answered.</p>
        </div>
        <div class="data-rows">
          <NuxtLink
            v-for="sess in closed"
            :key="sess.id"
            :to="sessionPath(researchSlug, sess.code || sess.id)"
            class="data-row session-row"
          >
            <span class="session-main">
              <span class="session-line">
                <span v-if="sess.code" class="short-code">{{ sess.code }}</span>
                <span class="session-title">{{ sess.title }}</span>
              </span>
              <span v-if="sess.focus" class="session-focus">{{ sess.focus }}</span>
            </span>
            <StatusBadge :status="sess.status" />
          </NuxtLink>
        </div>
      </div>
    </template>

    <EmptyState
      v-else
      icon="&#x1F5E3;"
      title="No sessions yet"
      description="A session is an interview: the agent asks, you answer, and the answers become entries. Ask your AI client to start one on this research."
    >
      <NuxtLink :to="`/research/${researchSlug}`" class="btn">Back to the research</NuxtLink>
    </EmptyState>
  </div>

  <EmptyState
    v-else
    icon="&#x1F50D;"
    title="Research not found"
    description="It may have been deleted, or the address may name a research that never existed."
  >
    <NuxtLink to="/" class="btn btn-primary">Back to all research</NuxtLink>
  </EmptyState>
</template>

<script setup lang="ts">
import { sessionPath } from '~/composables/useResearchPaths'

/*
 * Sessions had no page. The finished ones lived behind a collapsed toggle on
 * the research page, right-aligned in a row of their own with nothing beside
 * them — a whole line of the most valuable screen in the product spent on a
 * disclosure nobody opened. They are the record of how the research came to
 * know what it knows, which is worth a page.
 */
const route = useRoute()
const id = route.params.id as string

const { data: researchData, pending } = await useApi<any>(`/api/researches/${id}`)
const research = computed(() => researchData.value?.data?.research)
const researchSlug = computed(() => research.value?.code || id)

const { data: sessionsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/sessions`)
const sessions = computed<any[]>(() => sessionsData.value?.data ?? [])
const active = computed(() => sessions.value.filter(s => s.status === 'active'))
const closed = computed(() => sessions.value.filter(s => s.status !== 'active'))
</script>

<style scoped>
.lead {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-bottom: var(--space-6);
}
.card + .card { margin-top: var(--space-6); }
.heading-note {
  font-size: var(--type-xs);
  font-weight: var(--weight-normal);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}
.group-blurb {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  max-width: var(--measure-prose);
  margin-top: var(--space-1);
}

.session-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  text-decoration: none;
  color: inherit;
  transition: background var(--transition-fast);
}
.session-row:hover { background: var(--color-surface-hover); text-decoration: none; }
.session-main { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.session-line { display: flex; align-items: baseline; gap: var(--space-2); min-width: 0; }
.session-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  /* A session title is free text and can arrive with no spaces in it. */
  overflow-wrap: anywhere;
}
.session-focus {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  overflow-wrap: anywhere;
}
</style>
