# Track B — Backend API Specs (New Go Endpoints)

**Scope:** Новые endpoints в Go backend.  
**Files to modify:** `internal/api/`, `internal/service/`, `internal/storage/`

---

## BS-1: Research List — Расширенные метаданные

**Проблема:** `GET /api/researches` возвращает только базовые поля. Нет счётчиков entries, pending questions.

**Текущий response:**
```json
{
  "data": [
    { "id": "...", "name": "...", "goal": "...", "status": "...", "tags": [...] }
  ]
}
```

**Желаемый response:**
```json
{
  "data": [
    {
      "id": "...",
      "name": "...",
      "goal": "...",
      "status": "...",
      "tags": [...],
      "sections_count": 6,
      "entries_count": 23,
      "entries_today": 5,
      "pending_questions": 3,
      "active_session_id": "uuid-or-null",
      "updated_at": "2026-04-05T12:00:00Z"
    }
  ]
}
```

**Реализация (SQL aggregation):**
```sql
-- В storage/research.go, расширить ListResearches query
SELECT
  r.*,
  COUNT(DISTINCT s.id) as sections_count,
  COUNT(DISTINCT e.id) as entries_count,
  COUNT(DISTINCT CASE WHEN e.created_at > datetime('now', '-1 day') THEN e.id END) as entries_today,
  COUNT(DISTINCT CASE WHEN q.status = 'pending' THEN q.id END) as pending_questions,
  sess.id as active_session_id
FROM researches r
LEFT JOIN sections s ON s.research_id = r.id
LEFT JOIN entries e ON e.section_id = s.id
LEFT JOIN questions q ON q.session_id = sess.id
LEFT JOIN sessions sess ON sess.research_id = r.id AND sess.status = 'active'
WHERE (r.status = ? OR ? = '')
GROUP BY r.id
ORDER BY r.updated_at DESC
```

**Domain struct изменение (`internal/domain/research.go`):**
```go
type ResearchSummary struct {
    Research
    SectionsCount   int    `json:"sections_count"`
    EntriesCount    int    `json:"entries_count"`
    EntriesToday    int    `json:"entries_today"`
    PendingQuestions int   `json:"pending_questions"`
    ActiveSessionID string `json:"active_session_id,omitempty"`
}
```

---

## BS-2: Activity Feed Endpoint

**Endpoint:** `GET /api/researches/{id}/activity?limit=20`

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "research_id": "uuid",
      "entity": "entry",
      "entity_id": "uuid",
      "action": "created",
      "title": "Entry: Market Analysis added to Competitive Landscape",
      "timestamp": "2026-04-05T12:34:56Z"
    },
    {
      "entity": "question",
      "action": "answered",
      "title": "Question answered: What is the target market size?",
      "timestamp": "2026-04-05T12:30:00Z"
    }
  ]
}
```

**Реализация:**

Добавить таблицу `activity_log`:
```sql
CREATE TABLE IF NOT EXISTS activity_log (
    id          TEXT PRIMARY KEY,
    research_id TEXT NOT NULL,
    entity      TEXT NOT NULL,  -- research, section, entry, question, task, session
    entity_id   TEXT NOT NULL,
    action      TEXT NOT NULL,  -- created, updated, completed, answered
    title       TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (research_id) REFERENCES researches(id) ON DELETE CASCADE
);
```

Триггеры записи: вызывать `ActivityLog.Record(...)` в каждом service методе при create/update.

**Frontend использование:**
```vue
<!-- В research/[id]/index.vue — виджет activity feed -->
<div class="card activity-widget">
  <h3>Recent Activity</h3>
  <div v-for="event in activity" :key="event.id" class="activity-item">
    <span class="activity-time">{{ relativeTime(event.timestamp) }}</span>
    <span class="activity-title">{{ event.title }}</span>
  </div>
</div>
```

---

## BS-3: Global Full-text Search Endpoint

**Endpoint:** `GET /api/search?q=query&limit=20`

**Response:**
```json
{
  "data": {
    "researches": [
      { "id": "...", "name": "...", "highlight": "...query match..." }
    ],
    "entries": [
      {
        "id": "...", "title": "...", "research_id": "...", "research_name": "...",
        "section_name": "...", "highlight": "...matched content..."
      }
    ]
  }
}
```

**SQLite FTS5:**
```sql
-- При создании entry — также вставлять в FTS индекс
CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
    entry_id UNINDEXED,
    title,
    content,
    tokenize='porter unicode61'
);

-- Search query
SELECT e.id, e.title, e.section_id, s.display_name as section_name,
       r.id as research_id, r.name as research_name,
       snippet(entries_fts, 2, '<mark>', '</mark>', '...', 30) as highlight
FROM entries_fts
JOIN entries e ON e.id = entries_fts.entry_id
JOIN sections s ON s.id = e.section_id
JOIN researches r ON r.id = s.research_id
WHERE entries_fts MATCH ?
LIMIT ?
```

**Endpoint handler:**
```go
// internal/api/search.go
func (a *API) handleSearch(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    if len(q) < 2 {
        a.respond(w, http.StatusOK, map[string]any{"data": map[string]any{}})
        return
    }
    results, err := a.searchService.Search(r.Context(), q, 20)
    // ...
}
```

---

## BS-4: SSE Events — Расширение payload

**Проблема:** Текущие SSE события содержат минимальный payload. Frontend не знает что именно изменилось.

**Текущий SSE payload:**
```json
{ "entity": "entry", "research_id": "uuid" }
```

**Расширенный payload:**
```json
{
  "entity": "entry",
  "action": "created",
  "research_id": "uuid",
  "entity_id": "uuid",
  "title": "New entry: Market Analysis",
  "section_id": "uuid"
}
```

**Изменение в `internal/api/sse.go`:**
```go
type SSEEvent struct {
    Entity     string `json:"entity"`
    Action     string `json:"action"`      // created, updated, deleted
    ResearchID string `json:"research_id"`
    EntityID   string `json:"entity_id"`
    Title      string `json:"title,omitempty"`
    SectionID  string `json:"section_id,omitempty"`
}
```

**Frontend использование:** ActivityIndicator показывает `event.title` вместо generic "Claude is working".

---

## BS-5: Research Tags — Auto-complete Endpoint

**Endpoint:** `GET /api/tags?q=prefix`

**Response:**
```json
{ "data": ["buggregator", "backend", "go", "php"] }
```

**SQL:**
```sql
-- Собрать все уникальные теги из всех researches
SELECT DISTINCT value as tag
FROM researches, json_each(researches.tags)
WHERE value LIKE ? || '%'
ORDER BY value
LIMIT 20
```

---

## BS-6: Research Stats Endpoint

**Endpoint:** `GET /api/researches/{id}/stats`

**Response:**
```json
{
  "data": {
    "entries_by_section": [
      { "section_id": "uuid", "section_name": "Market Analysis", "count": 12 }
    ],
    "questions_by_status": {
      "pending": 3, "answered": 47, "deferred": 2, "skipped": 1
    },
    "tasks_by_status": {
      "pending": 5, "in_progress": 2, "completed": 10, "failed": 1
    },
    "entries_timeline": [
      { "date": "2026-04-05", "count": 8 },
      { "date": "2026-04-04", "count": 15 }
    ],
    "completion_pct": 72
  }
}
```

Используется для Dashboard-виджета на странице research detail.
