<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="template" class="template-page">
    <PageHeader
      :crumbs="[
        { label: 'Projects', to: { name: 'index' } },
        { label: 'Methodologies', to: '/templates' },
        { label: template.name },
      ]"
      :title="template.name"
    />

    <p v-if="template.description" class="lead">{{ template.description }}</p>

    <div class="card">
      <div class="head-row">
        <TemplateOriginBadge
          :tier="template.tier"
          :source="template.source"
          :forked-from="template.forked_from"
          :team-name="teamName(template.team_id)"
        />
        <span class="meta">
          {{ template.body_words }} words
          <template v-if="template.research_count"> · used in {{ template.research_count }} project{{ template.research_count === 1 ? '' : 's' }}</template>
        </span>
      </div>

      <!-- The same sentence the badge carries in its title attribute. On the list
           page each group's blurb says it; here nothing did, so a keyboard or
           screen-reader user got the label and never the fact behind it. -->
      <p class="origin-note">{{ originNote }}</p>

      <TemplateCriteria :when-to-use="template.when_to_use" :when-not-to-use="template.when_not_to_use" />
    </div>

    <div v-if="template.skills_resolved?.length" class="card">
      <h3 class="card-section-title">Skills it attaches</h3>
      <p class="blurb">
        These skills guide your AI assistant as it works on the project.
      </p>
      <div class="data-rows">
        <div v-for="sk in template.skills_resolved" :key="sk.slug" class="data-row skill-row">
          <div>
            <span class="skill-name">{{ sk.name || sk.slug }}</span>
            <span v-if="sk.missing" class="badge badge-pending" title="This skill is unavailable. Projects started with this methodology will not include it.">Missing</span>
          </div>
          <p class="skill-trigger">{{ sk.description || '— this server has no skill by that name —' }}</p>
        </div>
      </div>
    </div>

    <div class="card">
      <h3 class="card-section-title">What it tells the agent</h3>
      <div class="markdown-content body" v-html="rendered"></div>
    </div>

    <div class="card">
      <h3 class="card-section-title">Using it</h3>
      <p class="blurb">
        Ask your connected AI assistant to start a project with this methodology:
      </p>
      <CopyLine
        :text="startPrompt"
        label="Paste this into your AI client"
      />
    </div>
  </div>

  <EmptyState
    v-else
    icon="&#x1F50D;"
    title="Methodology not found"
    description="It may have been deleted, or it may belong to a team you are not in."
  >
    <NuxtLink to="/templates" class="btn btn-primary">Back to all methodologies</NuxtLink>
  </EmptyState>
</template>

<script setup lang="ts">
import { useMethodologyPrompt } from '~/composables/useMethodologyPrompt'
import { parseMarkdown } from '~/composables/useSafeMarkdown'

const route = useRoute()
const id = route.params.id as string

const { data, pending } = await useApi<{ data: any }>(`/api/templates/${id}`)
const template = computed(() => data.value?.data)
const startPrompt = useMethodologyPrompt(template)

/* parseMarkdown and deliberately no cross-reference pass: a methodology is read
   before any research exists, so `[[E3]]` inside one refers to nothing and
   linking it would invent a relationship. Same reasoning as SkillDetail. */
const rendered = computed(() => parseMarkdown(template.value?.body ?? ''))

/* Where this methodology came from, in a sentence — which for a global one is
   also the answer to "will an upgrade change it under me". */
const originNote = computed(() => {
  const tp = template.value
  if (!tp) return ''
  if (tp.tier !== 'team') {
    return tp.source === 'user'
      ? 'Added on this server by whoever runs it, and visible to every team here. It is not part of the app, so an upgrade will not change it.'
      : 'Ships with the app and is refreshed from it on every upgrade. Editing it makes a copy for your team; the original stays as it is for everybody else.'
  }
  const owner = teamName(tp.team_id) || 'your team'
  return tp.forked_from
    ? `${owner}'s copy of “${tp.forked_from}”, edited. It replaces the original everywhere this team looks.`
    : `Written by ${owner}, and theirs alone. No other team can see it.`
})

const { teams } = useTeams()
function teamName(teamId: string) {
  return teams.value?.find((t: any) => t.id === teamId)?.name
}
</script>

<style scoped>
.origin-note {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-top: var(--space-2);
}

.lead {
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--space-6);
}
.blurb {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-bottom: var(--space-4);
}
.card + .card { margin-top: var(--space-6); }
.card-section-title { margin-bottom: var(--space-2); }

.head-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
  flex-wrap: wrap;
}
.meta {
  font-size: var(--type-3xs);
  color: var(--color-text-faint);
  font-variant-numeric: tabular-nums;
}

.skill-row { display: block; }
.skill-name {
  font-size: var(--type-sm);
  font-weight: var(--weight-medium);
  margin-right: var(--space-2);
}
.skill-trigger {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  margin-top: var(--space-1);
  overflow-wrap: anywhere;
}

.body { max-width: var(--measure-prose); }
/* A methodology carries tables and fenced blocks; neither may push the page. */
.body :deep(pre), .body :deep(table) { overflow-x: auto; display: block; max-width: 100%; }
</style>
