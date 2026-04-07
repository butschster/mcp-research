# Component Decomposition Map

Specifications for decomposing monolithic pages into focused, reusable components.

---

## Source: `pages/research/[id]/index.vue` (918 lines -> ~250 lines)

### Extract: `components/research/ResearchDetailsPanel.vue`
- **Source lines**: 54-132 (template), 527-559 (script), 623-701 (CSS)
- **Props**: `research: any`, `open: boolean`
- **Emits**: `save(field: string, value: any)`, `update:open`
- **Internal state**: `editingField`, `editValue`, `editInput` ref
- **Behavior**: Collapsible card with inline editing for goal, description, instruction, tags. Read-only memory list. Double-click to edit, Save/Cancel actions. Tags edited as comma-separated string.

### Extract: `components/research/ActiveSessionsGrid.vue`
- **Source lines**: 135-150 (template), 734-766 (CSS)
- **Props**: `sessions: any[]`, `researchSlug: string`
- **Behavior**: Grid of active session cards as NuxtLinks. Each shows code badge + status + title.

### Extract: `components/research/PastSessionsList.vue`
- **Source lines**: 153-171 (template), 789-844 (CSS)
- **Props**: `sessions: any[]`, `researchSlug: string`
- **Internal state**: `showPast` toggle
- **Behavior**: Collapsible list of completed sessions. Toggle button with count + chevron.

### Extract: `components/research/ResearchSidebar.vue`
- **Source lines**: 176-220 (template), 702-732 (CSS)
- **Props**: `sections: any[]`, `activeSection: string`, `totalEntryCount: number`, `linksTotal: number`
- **Emits**: `update:activeSection`
- **Behavior**: Sidebar nav with "All entries", per-section items (with progress bars), "External links". Uses v-model pattern for activeSection.

### Extract: `components/research/EntriesView.vue`
- **Source lines**: 224-332 (template), 846-871 (CSS)
- **Props**: `entries: any[]`, `sections: any[]`, `researchSlug: string`, `loading: boolean`, `mode: 'all' | 'section'`, `sectionInfo?: any`, `tags: Array<{tag: string, count: number}>`
- **Internal state**: `activeTag` (tag filter)
- **Behavior**: Two modes: "all entries" grouped by section, or single section view. Tag filter panel. Uses EntryCard for rendering. Empty state fallback.

### Extract: `components/research/ExternalLinksView.vue`
- **Source lines**: 335-370 (template), 873-912 (CSS)
- **Props**: `groups: any[]`, `loading: boolean`
- **Behavior**: Links grouped by domain. Each link shows title + source entry code. Loading skeletons.

### Parent page retains
- Data fetching (research, tasks, sessions, entries, tags, links)
- Route management (activeSection from query param)
- Real-time update handler
- Page header with breadcrumbs and actions
- Archive toggle API call
- Computed properties for section/entries filtering

---

## Source: `pages/research/[id]/tasks.vue` (860 lines -> ~200 lines)

### Extract: `components/tasks/KanbanBoard.vue`
- **Source lines**: 33-72 (template), 502-614 (CSS)
- **Props**: `columns: Array<{status: string, label: string}>`, `tasks: any[]`, `researchSlug: string`
- **Emits**: `taskClick(task)`, `taskDrop(task, targetStatus)`
- **Internal state**: `draggedTask` ref
- **Behavior**: 4-column grid with drag-drop. Each column has header (dot + title + count), scrollable card body. Drag-over highlight effect. Uses KanbanCard internally.

### Extract: `components/tasks/KanbanCard.vue`
- **Source lines**: 51-65 (template), 571-603 (CSS)
- **Props**: `task: any`, `researchSlug: string`
- **Behavior**: Draggable card showing code badge + priority badge (if high) + title. Title renders cross-refs via renderRefs.

### Extract: `components/tasks/TaskDetailModal.vue`
- **Source lines**: 74-188 (template), 293-337 (script), 616-755 (CSS)
- **Props**: `task: any`, `researchSlug: string`
- **Emits**: `close`, `save(field: string, value: string)`, `updatePriority(priority: string)`
- **Internal state**: `editing`, `editValues`, input refs
- **Behavior**: Full-screen detail modal. Inline editing of title (input), description (textarea), result (textarea). Priority selector with 3 chips. Shows dates. Markdown rendering for description/result.

### Extract: `components/tasks/StatusChangeModal.vue`
- **Source lines**: 190-215 (template), 758-833 (CSS)
- **Props**: `visible: boolean`, `task: any`, `targetStatus: string`, `statusLabel: string`
- **Emits**: `confirm(comment: string)`, `cancel`
- **Behavior**: Confirmation modal for drag-drop status changes. Shows task code + title. Optional comment textarea.

### Extract: `components/tasks/CreateTaskModal.vue`
- **Source lines**: 217-252 (template), 429-458 (script), 758-833 (CSS)
- **Props**: `visible: boolean`
- **Emits**: `create(data: {title: string, description: string, priority: string})`, `cancel`
- **Internal state**: `newTask` form values
- **Behavior**: Form with title input, description textarea, priority button selector. Create/Cancel actions.

### Parent page retains
- Data fetching (research, tasks)
- Column definitions
- API calls (saveField, updatePriority, confirmStatusChange, createTask)
- Real-time update handler
- Page header with breadcrumbs

---

## Source: `pages/research/[id]/entry/[entryId].vue` (504 lines -> ~200 lines)

### Extract: `components/entry/CrossReferencesBlock.vue`
- **Source lines**: 59-95 (template), 364-456 (CSS)
- **Props**: `outgoing: any[]`, `incoming: any[]`, `researchSlug: string`
- **Internal**: `refLink()` function for building links
- **Behavior**: Shows outgoing refs ("References from this entry") and incoming refs ("Referenced by"). Each ref is a NuxtLink with code badge + title. Cross-research refs show research name. Unresolved refs show warning.

### Extract: `components/entry/ExternalLinksBlock.vue`
- **Source lines**: 98-117 (template), 458-474 (CSS)
- **Props**: `links: any[]`
- **Behavior**: List of external links. Each shows domain badge + title + external link icon.

### Extract: `components/entry/RelatedEntriesBlock.vue`
- **Source lines**: 120-144 (template), 432-450 (CSS)
- **Props**: `entries: any[]`, `currentTags: string[]`, `researchSlug: string`, `researchId: string`
- **Behavior**: Related entries by shared tags. Each entry shows code + title + colored tag dots.

### Extract: `components/entry/EntryNavigation.vue`
- **Source lines**: 147-155 (template), 476-493 (CSS)
- **Props**: `prev?: {code?: string, id: string, title: string}`, `next?: {code?: string, id: string, title: string}`, `researchSlug: string`
- **Behavior**: Prev/next navigation buttons for sibling entries.

### Parent page retains
- Data fetching (research, entry, sessions, crossrefs, links, related, siblings)
- View toggle (rendered/source)
- Copy markdown function
- Entry header (title, code, status, tags, linked session)
- Markdown rendering
