<template>
  <div ref="rootRef" class="action-menu">
    <button
      class="btn btn-icon"
      :title="title"
      :aria-label="title"
      aria-haspopup="true"
      :aria-expanded="open ? 'true' : 'false'"
      @click.stop="open = !open"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>
    </button>

    <Transition name="action-menu">
      <div
        v-if="open"
        class="action-menu-list"
        :class="[`action-menu-list--${props.align}`, `action-menu-list--${props.width}`]"
        @click="onPanelClick"
      >
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
/**
 * The `⋯` that holds a row's or a page's less-used actions.
 *
 * The panel deliberately carries **no** `role="menu"`. It had one, and none of
 * its slotted children carried `role="menuitem"` — which is not a partial
 * implementation but a wrong announcement: a screen reader tells the reader
 * they are in a menu and then finds nothing in it, and the arrow-key navigation
 * the role promises does not exist either. A group of buttons announced as
 * buttons is the truth, and Tab already works on it.
 */
const props = withDefaults(
  defineProps<{
    /** Tooltip on the trigger button. */
    title?: string
    /** Which edge the panel is anchored to. */
    align?: 'left' | 'right'
    /**
     * `wide` is 232px, for a panel whose content does not ellipsise well. It
     * buys 52px: a text column of 180px rather than 128px, which is roughly a
     * 19-character name instead of a 13-character one. Long names still
     * truncate — the caller owns saying so.
     */
    width?: 'default' | 'wide'
  }>(),
  { title: 'More actions', align: 'right', width: 'default' }
)

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

function onClickOutside(e: MouseEvent) {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    open.value = false
    // Return focus to the trigger so keyboard users are not dropped at the top.
    rootRef.value?.querySelector('button')?.focus()
  }
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  document.removeEventListener('keydown', onKeydown)
})

// The panel closes when an item is used, and only then. It used to close on any
// click inside it, which meant the deliberately inert header — the most
// tile-like thing in the list — answered a press by making everything vanish
// with nothing having happened. Inertness that responds is not inertness.
function onPanelClick(e: MouseEvent) {
  const el = e.target as HTMLElement | null
  if (el?.closest('.action-menu-header, .action-menu-divider')) return
  open.value = false
}

function focusTrigger() {
  rootRef.value?.querySelector('button')?.focus()
}

defineExpose({
  close: () => (open.value = false),
  /**
   * Puts focus back on the `⋯`.
   *
   * A menu item unmounts the moment the menu closes, so anything that saved the
   * active element on open — a modal restoring focus, for one — is holding a
   * detached node by the time it tries to return there, and the reader is
   * dropped at the top of the page. The trigger is the only element still
   * standing, so it is where focus belongs.
   */
  focusTrigger,
})
</script>

<style scoped>
.action-menu { position: relative; }

.action-menu-list {
  position: absolute;
  top: calc(100% + var(--space-1));
  min-width: 180px;
  /* Raised, not surface. The token exists for exactly this — "a floating layer
     so it separates from a card behind it instead of matching it" — and was
     used nowhere, so every panel in the product floated over a card of its own
     colour, and everything inside was being judged against a background that
     was not doing its job. */
  background: var(--color-surface-raised);
  /* One column for the glyphs and icons, shared by the header and the items, so
     a name and a label start at the same x. That alignment is most of the
     difference between "designed" and "pasted in". */
  --menu-icon: 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-2);
  /* The header this sits in is raised to --z-in-page, which is below the sticky
     nav on purpose. --z-overlay keeps the panel above in-page content without
     escaping that header. */
  z-index: var(--z-overlay);
  padding: var(--space-1) 0;
}
.action-menu-list--wide { min-width: 232px; }
.action-menu-list--right { right: 0; }
.action-menu-list--left { left: 0; }

/* Slotted content belongs to the parent's scope, so it needs :slotted(). */
.action-menu-list :slotted(.action-menu-item) {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: 0.45rem 0.75rem;
  border: none;
  background: none;
  color: var(--color-text-muted);
  font-size: var(--type-sm);
  text-align: left;
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.action-menu-list :slotted(.action-menu-item:hover) {
  background: var(--color-surface-hover);
  color: var(--color-text);
  text-decoration: none;
}
/* A non-interactive block at the top of a panel: who wrote this, when.
   Inset rather than full-bleed, because the panel's own top padding would
   otherwise leave a 4px strip of raised colour above it and read as a mistake;
   and recessed rather than lightened, because hover in this product goes
   lighter and a lighter tile reads as already-hovered.

   It has no hover and no pointer on purpose. Inertness is carried by the
   absence of feedback, which is what a person actually tests with. */
.action-menu-list :slotted(.action-menu-header) {
  display: block;
  margin: 0 var(--space-1) var(--space-1);
  padding: var(--space-2) calc(0.75rem - var(--space-1));
  background: var(--color-surface);
  border-radius: var(--radius-sm);
  white-space: normal;
  line-height: var(--line-tight);
}

.action-menu-list :slotted(.action-menu-item:disabled) {
  opacity: 0.5;
  cursor: not-allowed;
}
.action-menu-list :slotted(.action-menu-item--danger) {
  color: var(--color-error);
}
/* A rule between two rows is a divider; a rule at the edge of a list is an
   edge, and an edge belongs to whatever frames the list. So this is a border on
   the row below it rather than a floating 1px block with air on both sides —
   three items separated by two full-bleed rules read as more scaffolding than
   content. */
.action-menu-list :slotted(.action-menu-divider) {
  height: 0;
  margin: var(--space-1) 0 0;
  border-top: 1px solid var(--color-border);
}

.action-menu-enter-active,
.action-menu-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.action-menu-enter-from,
.action-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
