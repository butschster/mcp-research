---
name: storybook-setup
description: Sets up Storybook for the Nuxt 4 / Vue 3 frontend, creates stories for all components, and documents the component library. Use when the user asks to add Storybook, create component stories, build a component library/catalog, document UI components, set up visual testing, preview components in isolation, create a design system showcase, or mentions "storybook" in any context. Also trigger on requests to catalog components, create a component playground, audit the component library visually, or bring the component library to order.
---

# Storybook Setup for mcp-research

You are setting up Storybook 8 for the Nuxt 4 frontend at `frontend/` and writing stories for every component. The goal
is a comprehensive, well-documented component library with stories covering all states and edge cases.

## Tech context

- **Nuxt ^4.0.0** SPA mode (`ssr: false`), **Vue ^3.5.0** (Composition API, `<script setup>`)
- **Pure custom CSS** — CSS custom properties in `assets/css/main.css` (no Tailwind, no utility framework)
- **Design tokens**: `--color-*`, `--space-*`, `--type-*`, `--radius-*`, `--transition-*`
- **Fonts**: Outfit (display/body, 300-800), JetBrains Mono (code/mono, 400-500) — Google Fonts CDN
- **No icon library** — inline SVGs throughout
- **No PrimeVue** — all components are custom-built
- **Composables**: `useApi`, `useAuth`, `useCrossRefs` (`renderRefs`), `useRealtimeUpdates`, `useKeyboardNav`, `useResearchMindmap`, `useTagHue` (`tagHue`)

## Core Principle — Document All States

Every component story must cover:
1. **Default** state with typical data
2. **Empty** state (no data, empty arrays, null values)
3. **Many items** state (long lists, many tags, long text)
4. **Edge cases** (missing optional props, boundary values)
5. **All Variants** story showing every variant side by side (when applicable)

---

## Phase 1 — Install & Configure Storybook

### 1.1 Install dependencies

```bash
cd frontend
npx storybook@latest init --type vue3 --builder @storybook/builder-vite
```

If the interactive installer doesn't work, install manually:

```bash
npm install -D @storybook/vue3 @storybook/vue3-vite @storybook/addon-essentials @storybook/addon-links @storybook/addon-interactions @storybook/blocks storybook unplugin-auto-import
```

### 1.2 Configure `.storybook/main.ts`

Read `references/storybook-config.md` for the full configuration file. Key points:

- Vite alias `~` and `@` → `..` (project root, NOT `../app`)
- `unplugin-auto-import` for Vue/vue-router auto-imports
- Stubs for `#imports` and `#app`
- No Tailwind PostCSS — pure CSS
- No PrimeVue configuration

### 1.3 Configure `.storybook/preview.ts`

Read `references/storybook-config.md` for full file. Key points:

- Import `../assets/css/main.css` for design tokens and global styles
- `NuxtLink` → stub `<a>`, `ClientOnly` → passthrough, `Teleport` → inline render
- Vue Router with memory history
- Dark theme decorator using `--color-bg` and `--color-text`
- Font loading via `preview-head.html`

### 1.4 Add scripts and manager config

See `references/storybook-config.md` for `package.json` scripts, `.storybook/manager.ts` dark theme.

---

## Phase 2 — Component Audit

**Run `bash .claude/skills/storybook-setup/scripts/audit-components.sh` from project root first.**

Then manually scan all components. Organize by tier:

### Tier 1 — Base / shared (write first)

| Component              | Key behavior                              |
|------------------------|-------------------------------------------|
| `ShortCode`            | Code badge (R1, E5, S3, T42)              |
| `StatusBadge`          | 14 status variants with icons             |
| `ProgressBar`          | Value/total with color states             |
| `Breadcrumbs`          | Navigation path                           |
| `EmptyState`           | Icon, title, description, command          |
| `ActivityIndicator`    | Pulsing dot when active                   |
| `ConnectionStatus`     | WebSocket state (connected/disconnected)  |
| `WarningBanner`        | In-memory database warning                |
| `GettingStartedBanner` | Onboarding with dismiss                   |
| `PrintButton`          | Simple print trigger                      |
| `TagList`              | Tags with hue colors, clickable mode      |
| `ModalOverlay`         | Teleport + overlay + card chrome          |
| `EntryCard`            | Entry link card with code/status/tags     |

### Tier 2 — Research domain

| Component                          | Notes                                      |
|------------------------------------|--------------------------------------------|
| `ResearchCard`                     | Project card with archive, tags, timestamp  |
| `research/ResearchDetailsPanel`    | Collapsible details with inline editing     |
| `research/ActiveSessionsGrid`      | Active session cards                        |
| `research/PastSessionsList`        | Collapsible closed sessions                 |
| `research/ResearchSidebar`         | Section nav with progress bars              |
| `research/EntriesView`             | Entry grid with tag filtering               |
| `research/ExternalLinksView`       | Links grouped by domain                     |

### Tier 3 — Task domain

| Component                    | Notes                                |
|------------------------------|--------------------------------------|
| `tasks/KanbanBoard`         | 4-column board with drag-drop        |
| `tasks/KanbanCard`          | Draggable task card                  |
| `tasks/TaskDetailModal`     | Full task detail with inline editing |
| `tasks/StatusChangeModal`   | Status change confirmation           |
| `tasks/CreateTaskModal`     | New task form                        |

### Tier 4 — Entry domain

| Component                     | Notes                                |
|-------------------------------|--------------------------------------|
| `entry/CrossReferencesBlock`  | Outgoing + incoming refs             |
| `entry/ExternalLinksBlock`    | Entry-level links                    |
| `entry/RelatedEntriesBlock`   | Related by shared tags               |
| `entry/EntryNavigation`       | Prev/next sibling navigation         |

### Tier 5 — Complex components

| Component       | Notes                                |
|-----------------|--------------------------------------|
| `SearchModal`   | Global search with keyboard nav      |
| `QuestionList`  | Grouped questions with filters       |
| `ResearchCard`  | Project card with archive toggle     |

### Tier 6 — Mindmap nodes

| Component                    | Notes                              |
|------------------------------|------------------------------------|
| `mindmap/RootNode`           | Research root with stats           |
| `mindmap/SectionNode`        | Section with progress              |
| `mindmap/EntryNode`          | Entry with tags                    |
| `mindmap/QuestionNode`       | Question with answer preview       |
| `mindmap/AnswerNode`         | Answer display                     |
| `mindmap/TaskNode`           | Task with priority                 |
| `mindmap/GroupLabelNode`     | Collapsible group header           |

---

## Phase 3 — Writing Stories

### 3.1 Story file conventions

- Co-locate: `ComponentName.stories.ts` next to `ComponentName.vue`
- Use CSF3 (Component Story Format 3)
- Provide `argTypes` for all props with controls
- Include at least: Default, each variant, edge cases (empty/null/long text)
- For components with multiple states: ALWAYS include an "All Variants" story
- For components that depend on composables/stores, mock at story level

### 3.2 Story template — simple component

```ts
import type { Meta, StoryObj } from '@storybook/vue3'
import StatusBadge from './StatusBadge.vue'

const meta: Meta<typeof StatusBadge> = {
  title: 'Base/StatusBadge',
  component: StatusBadge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: ['active', 'completed', 'archived', 'draft', 'pending', 'answered',
                'in_progress', 'deferred', 'skipped', 'blocked', 'failed', 'high', 'medium', 'low'],
    },
  },
}
export default meta
type Story = StoryObj<typeof StatusBadge>

export const Default: Story = { args: { status: 'active' } }
export const AllStatuses: Story = {
  render: () => ({
    components: { StatusBadge },
    template: `
      <div style="display: flex; flex-wrap: wrap; gap: 8px;">
        <StatusBadge v-for="s in ['active','completed','archived','draft','pending','answered','in_progress','deferred','skipped','blocked','failed','high','medium','low']" :key="s" :status="s" />
      </div>
    `,
  }),
}
```

### 3.3 Story template — component with slots

```ts
import type { Meta, StoryObj } from '@storybook/vue3'
import ModalOverlay from './ModalOverlay.vue'

const meta: Meta<typeof ModalOverlay> = {
  title: 'Base/ModalOverlay',
  component: ModalOverlay,
  tags: ['autodocs'],
  argTypes: {
    visible: { control: 'boolean' },
    size: { control: 'select', options: ['sm', 'md', 'lg'] },
  },
}
export default meta
type Story = StoryObj<typeof ModalOverlay>

export const Default: Story = {
  args: { visible: true },
  render: (args) => ({
    components: { ModalOverlay },
    setup: () => ({ args }),
    template: `
      <ModalOverlay v-bind="args">
        <h3 style="font-size: 1.25rem; font-weight: 600; margin-bottom: 1rem;">Modal Title</h3>
        <p style="color: var(--color-text-muted);">Modal content goes here.</p>
      </ModalOverlay>
    `,
  }),
}
```

### 3.4 Story template — component needing composable mock

```ts
export const Default: Story = {
  render: () => ({
    components: { EntryCard },
    setup() {
      const entry = {
        id: 'ent_001',
        code: 'E1',
        title: 'Component Architecture Patterns',
        status: 'completed',
        description: 'Analysis of Vue 3 component design',
        tags: ['vue', 'architecture'],
      }
      return { entry }
    },
    template: `<EntryCard :entry="entry" research-slug="R1" />`,
  }),
}
```

### 3.5 NuxtLink, ClientOnly, Teleport stubs

Handled globally in `.storybook/preview.ts` — see `references/storybook-config.md`.

---

## Phase 4 — Story Hierarchy

```
Design System/
  Tokens          ← color swatches, spacing, typography specimens
  Buttons         ← all btn-* variants
  Cards           ← card, card styles
  Badges          ← badge variants, status colors
  Typography      ← type scale specimens
Base/
  ShortCode
  StatusBadge
  ProgressBar
  Breadcrumbs
  EmptyState
  ActivityIndicator
  ConnectionStatus
  TagList
  ModalOverlay
  EntryCard
Research/
  ResearchCard
  ResearchDetailsPanel
  ActiveSessionsGrid
  PastSessionsList
  ResearchSidebar
  EntriesView
  ExternalLinksView
Tasks/
  KanbanBoard
  KanbanCard
  TaskDetailModal
  StatusChangeModal
  CreateTaskModal
Entry/
  CrossReferencesBlock
  ExternalLinksBlock
  RelatedEntriesBlock
  EntryNavigation
Search/
  SearchModal
Session/
  QuestionList
Mindmap/
  RootNode
  SectionNode
  EntryNode
  QuestionNode
  AnswerNode
  TaskNode
  GroupLabelNode
```

## Phase 5 — Design System Token Stories

Create `components/__docs__/DesignTokens.stories.ts` — renders all CSS custom properties as visual swatches plus all
semantic component classes. Read `references/storybook-config.md` for the full file content.

---

## Rules

### Story rules

- MUST co-locate story files next to components: `ComponentName.stories.ts`
- MUST use CSF3 format
- MUST add `tags: ['autodocs']` to every meta
- MUST mock composables/stores at the story level — never import real API clients
- MUST NOT import from `#imports` or `#app` — these are Nuxt auto-import aliases unavailable in Storybook
- MUST import Vue utilities (`ref`, `computed`, etc.) explicitly in story files
- When a component has multiple states/variants, ALWAYS create an "All Variants" / "All States" story
- MUST provide realistic mock data (not "Lorem ipsum")
- MUST test empty states and edge cases

### Nuxt/platform rules

- MUST register `NuxtLink`, `ClientOnly`, and `Teleport` stubs in preview.ts
- MUST wrap every story in the dark theme decorator (global in preview.ts)
- MUST handle `<Teleport to="body">` — Teleport is stubbed to render inline
- MUST keep explicit imports for non-auto-imported modules (e.g., `marked`)

---

## Execution Order

1. Install Storybook + configure (Phase 1)
2. Run component audit (Phase 2, `audit-components.sh` + manual scan)
3. Create Design System token stories (Phase 5)
4. Write Tier 1 base component stories
5. Write Tier 2–6 stories
6. Verify: `npm run storybook` launches with all stories visible

---

## Nuxt auto-import workaround

Components use Nuxt auto-imports (`computed`, `ref`, `watch`, etc.) without explicit imports. Storybook does NOT have
Nuxt's auto-import.

**Solution:** `unplugin-auto-import` in Storybook's Vite config + stubs for `#imports`/`#app`.

See `references/storybook-config.md` for full setup.

---

## Bundled references (read during execution)

| File                                | When to read | Purpose                                           |
|-------------------------------------|--------------|---------------------------------------------------|
| `references/storybook-config.md`    | Phase 1      | Full .storybook/ config files, stubs, workarounds |
| `references/decomposition-map.md`   | Phase 2      | Component decomposition specs                     |
| `references/mock-data-templates.md` | Phase 3      | Shared mock data for stories                      |
| `scripts/audit-components.sh`       | Phase 2      | Component coverage audit                          |
