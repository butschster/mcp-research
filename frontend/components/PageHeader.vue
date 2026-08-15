<template>
  <div class="page-header">
    <Breadcrumbs v-if="crumbs?.length" :crumbs="crumbs" />

    <div class="page-bar">
      <div class="title-with-code">
        <ShortCode v-if="code" :code="code" />
        <h1 class="page-title">{{ title }}</h1>
        <span v-if="count !== undefined" class="task-counter">{{ count }}</span>
      </div>

      <div v-if="$slots.actions" class="cluster page-header-actions">
        <slot name="actions" />
      </div>
    </div>

    <p v-if="lead || $slots.lead" class="page-lead">
      <slot name="lead">{{ lead }}</slot>
    </p>

    <slot name="below" />
  </div>
</template>

<script setup lang="ts">
/**
 * The title of a page, beside whatever that page lets you do to it.
 *
 * This layout existed under seven names — `.page-header-row` (declared twice),
 * `.research-header`, `.entry-header`, `.session-header`, `.tasks-header`,
 * `.roadmaps-header` and `.export-toolbar` — across fourteen pages. Three of
 * them had been dragged into the global stylesheet and back out again twice,
 * the second time breaking `/s/{token}`, because a shared view deliberately
 * renders the owner's markup and a page-scoped copy does not reach it.
 *
 * So the fix is not a renamed class. The pages lose a header block and gain a
 * tag, which is why `.page-bar` — written for this and shipped unadopted — is
 * introduced here rather than by a find-and-replace across fourteen files, four
 * of which are owner/share twins that have to stay identical.
 *
 * `.entry-header` retires with the rest, which also settles a name collision:
 * `components/mindmap/EntryNode.vue` declares its own `.entry-header` for a
 * graph node's header, an unrelated thing under the same name. It is scoped and
 * wins locally, so nothing is visibly broken today — which is exactly how the
 * three-way `.btn-icon` collision this codebase already documents got there.
 */
defineProps<{
  /** Passed straight to `Breadcrumbs`; omit on a page that is its own root. */
  crumbs?: { label: string; to?: string }[]
  /** The research or entry short code, when the page names one. */
  code?: string
  title: string
  /** A count beside the title — tasks, roadmaps. `0` is meaningful, so it is
   *  `undefined` rather than falsy that hides it. */
  count?: number
  /** A sentence under the title. Use the `lead` slot instead when it needs markup. */
  lead?: string
}>()
</script>

<style scoped>
/* `.page-header`, `.page-bar`, `.title-with-code`, `.page-title` and
   `.page-header-actions` stay global: the `/s/{token}` pages render this
   component's markup too, and the entry pages still write the row by hand
   until their header moves here as well. What is scoped is only what this
   component alone decides. */
.page-lead {
  margin-top: var(--space-2);
  max-width: var(--measure-prose);
  color: var(--color-text-muted);
  font-size: var(--type-sm);
}
</style>
