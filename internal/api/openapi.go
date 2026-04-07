package api

import (
	"net/http"

	"gopkg.in/yaml.v3"
)

// OpenAPI spec generated from registered routes.

func openAPISpec(_ bool) map[string]any {
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "MCP Research API",
			"description": "REST API for MCP Research server. Read endpoints are unauthenticated. Write endpoints require Bearer token.",
			"version":     "1.0.0",
		},
		"servers": []map[string]any{
			{"url": "http://localhost:8088", "description": "Local development"},
		},
	}

	paths := map[string]any{}

	// --- Read endpoints ---
	paths["/api/health"] = map[string]any{
		"get": endpoint("Health check", "Returns server status, database mode, and write API availability.", nil,
			response200(obj(
				field("status", "string", "Server status, always 'ok'"),
				field("in_memory", "boolean", "Whether using in-memory database"),
				field("write_api", "boolean", "Whether write API is enabled"),
			)),
		),
	}

	paths["/api/researches"] = map[string]any{
		"get": endpoint("List researches", "Returns all research projects. Optionally filter by status.",
			[]param{query("status", "Filter by status: active, completed, archived", false)},
			response200(obj(
				field("data", "array", "Array of research objects"),
				field("count", "integer", "Total count"),
			)),
		),
	}

	paths["/api/researches/{id}"] = map[string]any{
		"get": endpoint("Get research", "Returns research with sections (including entry counts) and active session.",
			[]param{path("id", "Research UUID")},
			response200(obj(
				field("data", "object", "Contains: research (object), sections (array with entries_count), active_session (object or null)"),
			)),
		),
	}

	paths["/api/researches/{id}/sections/{sectionId}/entries"] = map[string]any{
		"get": endpoint("List entries in section", "Returns entries WITHOUT content for efficiency. Use GET /api/entries/{id} for full content.",
			[]param{path("id", "Research UUID"), path("sectionId", "Section UUID")},
			response200(obj(
				field("data", "array", "Array of entry objects (id, code, title, description, status, tags, created_at)"),
				field("count", "integer", "Total count"),
			)),
		),
	}

	paths["/api/entries/{id}"] = map[string]any{
		"get": endpoint("Get entry", "Returns full entry including markdown content.",
			[]param{path("id", "Entry UUID")},
			response200(obj(field("data", "object", "Entry with id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at"))),
		),
	}

	paths["/api/researches/{id}/entries/by-code/{code}"] = map[string]any{
		"get": endpoint("Resolve entry by code", "Find entry by short code (e.g. E1, E2) within a research.",
			[]param{path("id", "Research UUID"), path("code", "Entry short code (e.g. E1)")},
			response200(obj(field("data", "object", "Full entry object"))),
		),
	}

	paths["/api/resolve/research/{code}"] = map[string]any{
		"get": endpoint("Resolve research by code", "Find research by short code (e.g. R1, R2).",
			[]param{path("code", "Research short code (e.g. R1)")},
			response200(obj(field("data", "object", "Contains: id, code, name"))),
		),
	}

	paths["/api/researches/{id}/crossrefs"] = map[string]any{
		"get": endpoint("List cross-references", "Returns all cross-references for a research, extracted from [[...]] patterns in entry content.",
			[]param{path("id", "Research UUID")},
			response200(obj(
				field("data", "array", "Array of crossref objects (source_entry_id, target_entry_id, target_ref, resolved)"),
				field("count", "integer", "Total count"),
			)),
		),
	}

	paths["/api/researches/{id}/tasks"] = map[string]any{
		"get": endpoint("List tasks", "Returns all tasks for a research.",
			[]param{path("id", "Research UUID")},
			response200(obj(field("data", "array", "Array of task objects"))),
		),
	}

	paths["/api/researches/{id}/sessions"] = map[string]any{
		"get": endpoint("List sessions", "Returns all sessions for a research.",
			[]param{path("id", "Research UUID")},
			response200(obj(field("data", "array", "Array of session objects"))),
		),
	}

	paths["/api/researches/{id}/sessions/{sessionId}"] = map[string]any{
		"get": endpoint("Get session", "Returns session with questions grouped by status and progress counters. Resolves session by UUID or short code (e.g. SS1) within the given research.",
			[]param{path("id", "Research UUID or short code"), path("sessionId", "Session UUID or short code (e.g. SS1)")},
			response200(obj(field("data", "object", "Contains: session, questions (grouped by status), progress (total, answered, pending, deferred, skipped)"))),
		),
	}

	paths["/api/researches/{id}/entries/{entryId}"] = map[string]any{
		"get": endpoint("Get entry by research", "Returns full entry including markdown content. Resolves entry by UUID or short code (e.g. E1) within the given research.",
			[]param{path("id", "Research UUID or short code"), path("entryId", "Entry UUID or short code (e.g. E1)")},
			response200(obj(field("data", "object", "Entry with id, code, research_id, section_id, title, content, description, status, tags, created_at, updated_at"))),
		),
	}

	paths["/api/researches/{id}/roadmaps"] = map[string]any{
		"get": endpoint("List roadmaps", "Returns all roadmaps for a research (without nodes/edges). Use GET /api/roadmaps/{id} for the full graph.",
			[]param{path("id", "Research UUID")},
			response200(obj(
				field("data", "array", "Array of roadmap objects (id, code, title, description, status, statuses)"),
				field("count", "integer", "Total count"),
			)),
		),
	}

	paths["/api/roadmaps/{id}"] = map[string]any{
		"get": endpoint("Get roadmap", "Returns a roadmap with all nodes and edges.",
			[]param{path("id", "Roadmap UUID")},
			response200(obj(field("data", "object", "Roadmap with id, code, research_id, title, description, status, statuses, nodes (array), edges (array)"))),
		),
	}

	// --- Write endpoints (always documented, require bearer auth at runtime) ---
	{
		addMethod(paths, "/api/researches", "post", writeEndpoint(
			"Create research",
			"Creates a new research project with optional initial sections. Auto-assigns short code (R1, R2...).",
			nil,
			body(obj(
				field("name", "string", "Research name (required)"),
				field("description", "string", "Research description"),
				field("goal", "string", "Research goal"),
				field("tags", "array", "Array of tag strings"),
				field("sections", "array", "Array of section objects: {name (slug, required), display_name, description, position}"),
			)),
			response201(obj(
				field("data", "object", "Contains: research_id, code, name, status, sections_created"),
			)),
		))

		addMethod(paths, "/api/researches/{id}", "put", writeEndpoint(
			"Update research",
			"Partial update. Only provided fields are changed. memory and add_memory are mutually exclusive.",
			[]param{path("id", "Research UUID")},
			body(obj(
				field("name", "string", "New name"),
				field("description", "string", "New description"),
				field("goal", "string", "New goal"),
				field("status", "string", "New status: active, completed, archived"),
				field("instruction", "string", "Working instructions for AI"),
				field("tags", "array", "Replace tags array"),
				field("memory", "array", "Replace entire memory array"),
				field("add_memory", "string", "Append single entry to memory"),
			)),
			response200(obj(field("data", "object", "Updated research object"))),
		))

		addMethod(paths, "/api/researches/{id}/sections", "post", writeEndpoint(
			"Add section",
			"Adds a new section to a research. Name must be a lowercase slug.",
			[]param{path("id", "Research UUID")},
			body(obj(
				field("name", "string", "Section slug name (required, lowercase)"),
				field("display_name", "string", "Human-readable name"),
				field("description", "string", "Section description"),
				field("position", "integer", "Sort position"),
			)),
			response201(obj(field("data", "object", "Created section object"))),
		))

		addMethod(paths, "/api/sections/{sectionId}", "put", writeEndpoint(
			"Update section",
			"Partial update. Setting status to 'completed' requires at least one entry.",
			[]param{path("sectionId", "Section UUID")},
			body(obj(
				field("display_name", "string", "New display name"),
				field("description", "string", "New description"),
				field("status", "string", "New status: draft, active, completed, archived"),
				field("position", "integer", "New sort position"),
			)),
			response200(obj(field("data", "object", "Updated section object"))),
		))

		addMethod(paths, "/api/entries", "post", writeEndpoint(
			"Create entry",
			"Creates entry with markdown content. Title and description are auto-generated from content if not provided. Auto-assigns short code (E1, E2...). Cross-references ([[E3]], [[R2:E5]]) in content are parsed and stored.",
			nil,
			body(obj(
				field("research_id", "string", "Research UUID (required)"),
				field("section_id", "string", "Section UUID (required)"),
				field("session_id", "string", "Session UUID (optional, links entry to a session)"),
				field("content", "string", "Markdown content (required)"),
				field("title", "string", "Title (auto-generated if empty)"),
				field("description", "string", "Description (auto-generated if empty)"),
				field("status", "string", "Status: draft, active, completed, archived"),
				field("tags", "array", "Array of tag strings"),
			)),
			response201(obj(field("data", "object", "Contains: entry_id, code, title, status"))),
		))

		addMethod(paths, "/api/entries/{id}", "put", writeEndpoint(
			"Update entry",
			"Partial update. Supports text_replace for surgical content edits. Cross-references re-parsed on update.",
			[]param{path("id", "Entry UUID")},
			body(obj(
				field("title", "string", "New title"),
				field("content", "string", "Replace entire content"),
				field("description", "string", "New description"),
				field("status", "string", "New status"),
				field("tags", "array", "Replace tags"),
				field("session_id", "string", "Link entry to a session (pass empty string to unlink)"),
				field("text_replace", "object", "Surgical edit: {from: string, to: string} replaces first occurrence"),
			)),
			response200(obj(field("data", "object", "Updated entry object"))),
		))

		addMethod(paths, "/api/tasks", "post", writeEndpoint(
			"Create task",
			"Creates a task in the research todo list.",
			nil,
			body(obj(
				field("research_id", "string", "Research UUID (required)"),
				field("title", "string", "Task title (required)"),
				field("description", "string", "Task description"),
				field("priority", "string", "Priority: high, medium, low (default: medium)"),
			)),
			response201(obj(field("data", "object", "Created task object"))),
		))

		addMethod(paths, "/api/tasks/{id}", "put", writeEndpoint(
			"Update task",
			"Partial update of task status, priority, or result.",
			[]param{path("id", "Task UUID")},
			body(obj(
				field("title", "string", "New title"),
				field("status", "string", "New status: pending, in_progress, blocked, completed, failed, deferred"),
				field("priority", "string", "New priority: high, medium, low"),
				field("result", "string", "Outcome description (when completing)"),
			)),
			response200(obj(field("data", "object", "Updated task object"))),
		))

		addMethod(paths, "/api/tasks/{id}", "delete", writeEndpoint(
			"Delete task",
			"Removes a task from the todo list.",
			[]param{path("id", "Task UUID")},
			nil,
			response200(obj(field("deleted", "boolean", "Always true"))),
		))

		addMethod(paths, "/api/sessions", "post", writeEndpoint(
			"Create session",
			"Creates a Q&A session with optional initial questions.",
			nil,
			body(obj(
				field("research_id", "string", "Research UUID (required)"),
				field("title", "string", "Session title (required)"),
				field("focus", "string", "Focus area for this session"),
				field("questions", "array", "Initial questions: [{text, area, rationale, priority, parent_id, position}]"),
			)),
			response201(obj(field("data", "object", "Created session object"))),
		))

		addMethod(paths, "/api/researches/{id}/crossrefs/rebuild", "post", writeEndpoint(
			"Rebuild cross-references",
			"Re-scans all entry content in the research and rebuilds the crossrefs table. Use when references become stale.",
			[]param{path("id", "Research UUID")},
			nil,
			response200(obj(
				field("rebuilt", "integer", "Number of entries scanned"),
				field("status", "string", "Always 'ok'"),
			)),
		))

		addMethod(paths, "/api/roadmaps", "post", writeEndpoint(
			"Create roadmap",
			"Creates a roadmap graph with optional initial nodes and edges. Use temp_id on nodes to reference them in edges.",
			nil,
			body(obj(
				field("research_id", "string", "Research UUID (required)"),
				field("title", "string", "Roadmap title (required)"),
				field("description", "string", "Roadmap description"),
				field("statuses", "array", "Available node statuses in order (e.g. ['not_started', 'in_progress', 'completed'])"),
				field("nodes", "array", "Array of node objects: {temp_id, title, description, node_type, status, position_x, position_y, parent_id}"),
				field("edges", "array", "Array of edge objects: {source (temp_id or node ID), target (temp_id or node ID), label, edge_type}"),
			)),
			response201(obj(field("data", "object", "Created roadmap with nodes and edges"))),
		))

		addMethod(paths, "/api/roadmaps/{id}", "put", writeEndpoint(
			"Update roadmap",
			"Partial update of roadmap metadata (title, description, statuses, status).",
			[]param{path("id", "Roadmap UUID")},
			body(obj(
				field("title", "string", "New title"),
				field("description", "string", "New description"),
				field("statuses", "array", "New list of available node statuses"),
				field("status", "string", "Roadmap status: active, completed, archived"),
			)),
			response200(obj(field("data", "object", "Updated roadmap object"))),
		))

		addMethod(paths, "/api/roadmaps/{id}", "delete", writeEndpoint(
			"Delete roadmap",
			"Deletes a roadmap and all its nodes and edges. Irreversible.",
			[]param{path("id", "Roadmap UUID")},
			nil,
			response200(obj(field("deleted", "boolean", "Always true"))),
		))

		addMethod(paths, "/api/roadmap-nodes/{nodeId}", "put", writeEndpoint(
			"Update roadmap node",
			"Partial update of a roadmap node (title, description, node_type, status, position, parent).",
			[]param{path("nodeId", "Node UUID")},
			body(obj(
				field("title", "string", "New title"),
				field("description", "string", "New description"),
				field("node_type", "string", "New type: step, milestone, decision, info, group"),
				field("status", "string", "New status (from the roadmap's statuses list)"),
				field("position_x", "number", "New X position"),
				field("position_y", "number", "New Y position"),
				field("parent_id", "string", "New parent node ID (empty to clear)"),
			)),
			response200(obj(field("data", "object", "Updated node object"))),
		))
	}

	spec["paths"] = paths

	// Security scheme
	spec["components"] = map[string]any{
		"securitySchemes": map[string]any{
			"bearerAuth": map[string]any{
				"type":        "http",
				"scheme":      "bearer",
				"description": "API token configured via --api-token flag or MCP_RESEARCH_API_TOKEN env var. Write endpoints return 401 if token is not configured on the server.",
			},
		},
	}

	return spec
}

func handleOpenAPI(writeEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		yaml.NewEncoder(w).Encode(openAPISpec(writeEnabled))
	}
}

// --- Helpers ---

type param struct {
	Name        string
	In          string
	Description string
	Required    bool
}

func path(name, desc string) param { return param{name, "path", desc, true} }
func query(name, desc string, required bool) param { return param{name, "query", desc, required} }

func endpoint(summary, desc string, params []param, resp map[string]any) map[string]any {
	e := map[string]any{"summary": summary, "description": desc, "responses": resp}
	if len(params) > 0 {
		var pp []map[string]any
		for _, p := range params {
			pp = append(pp, map[string]any{
				"name": p.Name, "in": p.In, "description": p.Description,
				"required": p.Required, "schema": map[string]any{"type": "string"},
			})
		}
		e["parameters"] = pp
	}
	return e
}

func writeEndpoint(summary, desc string, params []param, reqBody, resp map[string]any) map[string]any {
	e := endpoint(summary, desc, params, resp)
	e["security"] = []map[string]any{{"bearerAuth": []string{}}}
	if reqBody != nil {
		e["requestBody"] = reqBody
	}
	return e
}

func body(schema map[string]any) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{"schema": schema},
		},
	}
}

func response200(schema map[string]any) map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "Success",
			"content":     map[string]any{"application/json": map[string]any{"schema": schema}},
		},
	}
}

func response201(schema map[string]any) map[string]any {
	return map[string]any{
		"201": map[string]any{
			"description": "Created",
			"content":     map[string]any{"application/json": map[string]any{"schema": schema}},
		},
	}
}

func obj(fields ...map[string]any) map[string]any {
	props := map[string]any{}
	for _, f := range fields {
		props[f["_name"].(string)] = map[string]any{
			"type":        f["_type"],
			"description": f["_desc"],
		}
	}
	return map[string]any{"type": "object", "properties": props}
}

func field(name, typ, desc string) map[string]any {
	return map[string]any{"_name": name, "_type": typ, "_desc": desc}
}

// addMethod safely adds an HTTP method to a path, initializing the map if needed.
func addMethod(paths map[string]any, pathKey, method string, spec map[string]any) {
	if _, ok := paths[pathKey]; !ok {
		paths[pathKey] = map[string]any{}
	}
	paths[pathKey].(map[string]any)[method] = spec
}
