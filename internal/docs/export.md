# Export

How to export research data as documents.

## Export Page (Web UI)

Navigate to `/research/{code}/export` (or click the Export button on the research page).

The export page renders the full research as a single printable document:
- Research header (title, goal, description, tags)
- Table of contents with anchor links
- All sections with full entry content (rendered markdown)
- All sessions with questions and answers
- All tasks with statuses and results

### Download as Markdown

Click **"Download .md"** to download the complete research as a markdown file. The file includes all sections, entries, sessions, questions, and tasks in a structured format.

### Print to PDF

Click **"Print / PDF"** (or use Ctrl+P / Cmd+P) to open the browser print dialog. Select "Save as PDF" as the destination. The page is optimized for print — navigation, footer, and interactive elements are hidden automatically.

### Print Individual Entries

Open any entry page and use Ctrl+P / Cmd+P. The entry page has print-optimized CSS that hides navigation, action buttons, cross-references, and related entries, showing only the breadcrumb path and document content.

## Export API

### Full Export (JSON)

```
GET /api/researches/{id}/export
```

Returns all research data as structured JSON:

```json
{
  "research": { "name": "...", "goal": "...", ... },
  "sections": [
    {
      "name": "...",
      "display_name": "...",
      "entries": [
        { "title": "...", "content": "full markdown...", "tags": [...] }
      ]
    }
  ],
  "sessions": [
    {
      "title": "...",
      "questions": [
        { "text": "...", "answer": "...", "status": "answered" }
      ]
    }
  ],
  "tasks": [
    { "title": "...", "status": "completed", "result": "..." }
  ],
  "markdown": "# Full document as markdown string..."
}
```

### Download Markdown File

```
GET /api/researches/{id}/export?format=md
```

Returns the research as a downloadable `.md` file with `Content-Disposition: attachment` header.

The markdown document structure:
1. Title and metadata (goal, description, tags)
2. Table of contents
3. Sections with full entry content
4. Sessions with all questions and answers
5. Tasks with statuses and results

### Cross-References in Export

Cross-references (`[[E3]]`, `[[R2:E5]]`) are preserved as-is in markdown export. In the web export page, they are rendered as clickable links.
