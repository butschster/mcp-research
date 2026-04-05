# Implementation Roadmap

**Дата:** 2026-04-05  
**Версия:** 1.0

## Принципы приоритизации

- **Impact** — насколько улучшает experience для всех пользователей
- **Effort** — сложность реализации (часы работы)
- **Risk** — вероятность breaking changes или регрессий
- **Dependencies** — нужны ли другие изменения первыми

---

## SPRINT 1 — Quick Wins (1-2 дня, только frontend)

Эти изменения можно сделать за один PR без рисков.

| # | Задача | Файл | Effort | Impact |
|---|--------|------|--------|--------|
| QW-1 | Active sidebar left-border accent | `main.css` | 30 min | 🔴 Critical |
| QW-2 | Card hover shadow + tint | `main.css` | 30 min | 🟡 High |
| QW-3 | "Read-only view" badge в nav | `app.vue` + `main.css` | 1h | 🟡 High |
| QW-5 | Sidebar section — mini progress + StatusBadge | `research/[id]/index.vue` | 2h | 🟡 High |
| QW-8 | Mobile sidebar — pill horizontal scroll | `main.css` | 2h | 🔴 Critical |
| CI-6 | StatusBadge — иконки к статусам | `StatusBadge.vue` | 1h | 🟢 Medium |

**Total Sprint 1:** ~7 hours

---

## SPRINT 2 — Component Improvements (2-3 дня, только frontend)

| # | Задача | Файл | Effort | Impact |
|---|--------|------|--------|--------|
| QW-4 | EmptyState — copyable command prop | `EmptyState.vue` + usages | 3h | 🔴 Critical |
| QW-6 | Tag color-coding | `main.css` + `ResearchCard.vue` | 2h | 🟢 Medium |
| QW-7 | Entry breadcrumb — section в пути | `[entryId].vue` | 1h | 🟡 High |
| CI-1 | ActivityIndicator компонент | `ActivityIndicator.vue` + `app.vue` | 3h | 🔴 Critical |
| CI-2 | ResearchCard — stats row | `ResearchCard.vue` | 2h | 🟡 High |
| CI-3 | ConnectionStatus — last event | `ConnectionStatus.vue` | 2h | 🟡 High |
| CI-4 | QuestionList — count badge в closed group | `QuestionList.vue` | 1h | 🟢 Medium |
| CI-5 | ProgressBar — label + color classes | `ProgressBar.vue` | 1h | 🟢 Medium |

**Total Sprint 2:** ~15 hours

---

## SPRINT 3 — New Frontend Features (3-4 дня, только frontend)

| # | Задача | Файл | Effort | Impact | Notes |
|---|--------|------|--------|--------|-------|
| NF-2 | GettingStartedBanner — onboarding | `GettingStartedBanner.vue` | 4h | 🔴 Critical | |
| NF-1 | SearchModal — client-side search | `SearchModal.vue` | 8h | 🔴 Critical | Cmd+K |
| NF-3 | Keyboard navigation (useKeyboardNav) | `useKeyboardNav.ts` | 3h | 🟡 High | |
| NF-4 | Print/Export enhancement | `PrintButton.vue` | 2h | 🟢 Medium | |
| FS-4 | Attention Required widget (frontend part) | `research/[id]/index.vue` | 3h | 🔴 Critical | Из существующих данных tasks+session |

**Total Sprint 3:** ~20 hours

---

## SPRINT 4 — Backend API (3-4 дня, Go + frontend)

| # | Задача | Effort | Impact | Risk |
|---|--------|--------|--------|------|
| BS-1 | Research list — enriched metadata (SQL aggregation) | 4h | 🔴 Critical | Low — read-only SQL change |
| BS-4 | SSE events — extended payload | 2h | 🟡 High | Low |
| BS-5 | Tags autocomplete endpoint | 2h | 🟢 Medium | Low |

**Total Sprint 4:** ~8 hours Go + 4h frontend integration

---

## SPRINT 5 — Full-Stack Features (5-7 дней)

| # | Задача | Effort | Impact | Risk |
|---|--------|--------|--------|------|
| BS-2 | Activity Log table + endpoint | 8h Go | 🔴 Critical | Medium — new DB migration |
| BS-3 | Full-text Search (SQLite FTS5) | 12h Go | 🔴 Critical | Medium — new index |
| BS-6 | Research Stats endpoint | 4h Go | 🟡 High | Low |
| FS-1 | Activity Feed widget (frontend) | 4h Vue | 🔴 Critical | — |
| FS-2 | Stats Dashboard widget (frontend) | 4h Vue | 🟡 High | — |
| FS-3 | Full-text search integration in SearchModal | 3h Vue | 🔴 Critical | — |

**Total Sprint 5:** ~35 hours

---

## Summary: Total Effort by Track

| Track | Sprints | Effort | Risk | First Deliverable |
|-------|---------|--------|------|-------------------|
| Track A Frontend-only | 1+2+3 | ~42h | Low | Sprint 1 (~7h) |
| Track B Full-stack | 4+5 | ~43h | Medium | Sprint 4 (~12h) |
| **Total** | | **~85h** | | |

---

## Dependency Graph

```
QW-1,2,3,5,8,CI-6 (Sprint 1, независимые)
       ↓
QW-4,6,7,CI-1,2,3,4,5 (Sprint 2, независимые)
       ↓
NF-1,2,3,4, FS-4 (Sprint 3, NF-1 зависит от CI-1)
       ↓
BS-1,4,5 (Sprint 4, Go backend)
       ↓ BS-1 → CI-2 (ResearchCard stats)
BS-2 → FS-1 (Activity Log → Activity Feed)
BS-3 → FS-3 (FTS5 → SearchModal full-text)
BS-6 → FS-2 (Stats endpoint → Stats widget)
```

---

## A/B Testing Roadmap (Post-Launch)

| Test | Hypothesis | Variable | Expected Impact |
|------|-----------|----------|----------------|
| T-1 | Active left-border improves navigation clarity | border width: 2px vs 3px vs accent circle | -20% mis-clicks на sidebar |
| T-2 | GettingStartedBanner increases setup completion | Banner always-on vs dismiss-after-first-project | +30% пользователей доходят до первого research |
| T-3 | Search trigger placement | Nav center vs nav right | +15% search usage |
| T-4 | Stats widget placement | Above session widget vs below | CTR на session detail |
| T-5 | ActivityIndicator label | "Claude is working" vs "Updating..." vs no label | User perceived reliability |

---

## Known Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| BS-2 activity log — DB migration может сломать существующие installs | Предусмотреть `migrate up/down`, проверить на in-memory DB |
| BS-3 FTS5 — тяжёлый для больших БД | Добавить индексирование только при create/update, не rebuild-on-read |
| NF-1 SearchModal — localStorage usage запрещён в artifacts | Только в реальном app, не в Claude artifact |
| CI-2 ResearchCard stats — требует BS-1 | Показывать stats gracefully когда поля отсутствуют (optional chaining) |
