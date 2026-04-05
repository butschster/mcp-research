# MCP Research — UI/UX Improvement Plan

**Дата аудита:** 2026-04-05  
**Аудитор:** Landing Page Architect (multi-expert simulation)  
**Версия:** 1.0

## Scope

Полный аудит Web UI (Nuxt 4 frontend) и REST API (Go backend) с планом улучшений в двух треках:

- **Track A — Frontend-only** — только изменения в `frontend/`, без затрагивания Go. Быстрый цикл.
- **Track B — Full-stack** — новые API endpoints в Go + frontend. Фундаментальные улучшения.

## Файлы спецификаций

| Файл | Содержание |
|------|-----------|
| `01-quick-wins-frontend.md` | Track A: 8 quick wins, только CSS/Vue |
| `02-component-improvements.md` | Track A: улучшение существующих компонентов |
| `03-new-features-frontend.md` | Track A: новые фичи только на фронте |
| `04-backend-api-specs.md` | Track B: новые API endpoints |
| `05-full-stack-features.md` | Track B: фичи с бэкендом |
| `06-implementation-roadmap.md` | Общая дорожная карта с приоритетами |

## Ключевые проблемы (Executive Summary)

### 🔴 Critical (блокируют retention)
1. **Нет global search** — невозможно найти запись при 10+ проектах
2. **Нет onboarding** — пустой экран при первом запуске = abandon
3. **Нет activity indicator** — непонятно, что делает Claude прямо сейчас
4. **Mobile layout broken** — sidebar layout ломается, sticky теряется

### 🟡 High Impact (снижают удовольствие)
5. **Active sidebar item почти невидим** — нет left-border accent
6. **Empty states — не actionable** — нет copyable команд для Claude
7. **Нет "что нужно от меня"** — pending questions не prominent
8. **Card hover — только border** — слабый feedback

### 🟢 Medium (polish)
9. **Tags без семантики** — все одного цвета
10. **Nav — слишком пустой** — нет проекционных breadcrumbs
11. **Entry context пропадает** — нет пути секция → запись
12. **Нет keyboard navigation** — полностью mouse-driven

## Success Metrics

| Метрика | Baseline | Target |
|---------|----------|--------|
| Time to find specific entry | ~45 sec | < 10 sec |
| First-run understand | Low | High (onboarding) |
| Mobile usability score | Poor | Acceptable |
| Keyboard-only navigation | 0% | 80% coverage |
