package api

import (
	"github.com/danielgtaylor/huma/v2"

	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dovod-app/app/internal/api/handlers"
	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/service"
)

// shareRoutes builds the public read surface a share link opens onto.
//
// It is a mux of its own, mounted under `/api/shared/{token}/`, and that is the
// whole design. A share token is checked once, at the prefix; inside it there
// are only the routes listed here, and there is no path by which a share token
// reaches a route built for an owner. The alternative — teaching every existing
// route to also accept a share — would have made every future route a decision
// somebody had to remember to make correctly.
//
// The handlers are the same instances the authenticated API uses. They are safe
// here because service.Access resolves a share to "viewer on exactly one
// research": a read outside that research is ErrNotFound and a write is
// ErrForbidden, decided in one place rather than per route.
type shareDeps struct {
	shares   *service.ShareService
	research *handlers.ResearchHandler
	entry    *handlers.EntryHandler
	session  *handlers.SessionHandler
	task     *handlers.TaskHandler
	roadmap  *handlers.RoadmapHandler
	crossref *handlers.CrossRefHandler
	links    *handlers.ExternalLinkHandler
	export   *handlers.ExportHandler
	share    *handlers.ShareHandler
	// graph draws the research as nodes and edges. It is mounted ungated because
	// documents and sections are always in a link; the handler itself withholds
	// the session, question and task nodes the include flags do not cover.
	graph *handlers.GraphHandler
}

// shareReadLimit is generous because one page load is many fetches — the
// entries, their references, their links. It is here to blunt noise on a public
// URL, not to be the thing that keeps a link private.
var shareReadLimiter = newRateLimiter(600, time.Minute)

// shareUnlockLimiter is the one that matters: a password is guessable in a way
// a 256-bit token is not. It is keyed by the share, not by the caller, so
// spreading the attempts across addresses does not spread the budget.
//
// Its own instance, with its own map. Sharing one with the read limiter meant
// traffic to the public prefix could push the map to its ceiling and evict the
// password budget along with it.
var shareUnlockLimiter = newRateLimiter(10, time.Minute)

// shareUnlockIPLimiter is the second half of the same guard: the token key
// bounds attempts against one link, this one bounds a caller working through
// many links at once.
var shareUnlockIPLimiter = newRateLimiter(30, time.Minute)

// registerShareRoutes mounts both halves of the feature: the owner's management
// routes on the authenticated API, and the visitor's read surface under its own
// prefix.
func registerShareRoutes(rt *router, deps shareDeps, sShare *huma.Schema) {
	sOK := envelope(map[string]*huma.Schema{"status": {Type: "string"}})

	// --- Owner side: ordinary authenticated routes, no share involved ---
	rt.route(accessWrite, op("POST", "/api/researches/{id}/shares", "Create a share link",
		"Issues a read-only link to this research for people with no account here.\n\n"+
			"**The token is returned once, from this call, and cannot be recovered** — only its SHA-256 is stored. Choose what the link exposes with the `include` flags; everything not named stays private, and private skills, research memory, session notes and every document's revision history are never exposed by any link.").
		tag("Share links").
		body("What the link may show, and optionally a password and an expiry.", envelope(map[string]*huma.Schema{
			"include": envelope(map[string]*huma.Schema{
				"sessions": {Type: "boolean", Description: "Default `false`."},
				"tasks":    {Type: "boolean", Description: "Default `false`."},
				"roadmaps": {Type: "boolean", Description: "**Default `true`** — the one flag that is on unless you turn it off."},
				"export":   {Type: "boolean", Description: "Default `false`."},
			}),
			"password":   {Type: "string", Description: "When set, a visitor exchanges it at `/api/shared/{token}/unlock` before anything is readable."},
			"expires_at": {Type: "string", Format: "date-time"},
		})).
		returns("201", "The link, with its token shown for the only time.", envelope(map[string]*huma.Schema{
			"data": envelope(map[string]*huma.Schema{
				"share": sShare,
				"token": {Type: "string", Description: "The token itself. Store it now — only its SHA-256 is kept, and no route returns it again."},
				"url":   {Type: "string", Description: "The address to hand to a person, token included."},
			}),
		})).
		build(), deps.share.Create)

	rt.route(accessRead, op("GET", "/api/researches/{id}/shares", "List share links",
		"The links issued for this research, what each exposes and whether it is still live. The tokens themselves are not stored and are not listed.").
		tag("Share links").
		returns("200", "The links.", envelope(map[string]*huma.Schema{
			"data":  listOf(sShare),
			"count": {Type: "integer"},
		})).
		build(), deps.share.List)

	rt.route(accessWrite, op("PUT", "/api/shares/{id}", "Change what a share link shows",
		"Renames a live link or changes its `include` flags without reissuing it, so an address people have already saved keeps working.\n\n"+
			"**`include` is a complete replacement, not a patch**: send all four flags. A flag left out of the object is `false`. Widening a link shows more to everyone who already holds it, at once; the UI confirms that before calling this.\n\n"+
			"A revoked or expired link answers `409` — this is the owner's surface, where knowing the link died is the useful answer.").
		tag("Share links").
		body("The new label and/or the full set of include flags.", envelope(map[string]*huma.Schema{
			"label": {Type: "string"},
			"include": envelope(map[string]*huma.Schema{
				"sessions": {Type: "boolean"},
				"tasks":    {Type: "boolean"},
				"roadmaps": {Type: "boolean"},
				"export":   {Type: "boolean"},
			}),
		})).
		returns("200", "The updated link, in the same shape the list returns.", envelope(map[string]*huma.Schema{
			"data": envelope(map[string]*huma.Schema{"share": sShare}),
		})).
		responds("409", "The link is revoked or expired and cannot be changed.").
		build(), deps.share.Update)

	rt.route(accessWrite, op("DELETE", "/api/shares/{id}", "Revoke a share link",
		"The link stops working immediately. A revoked link, an expired one, an unknown one and one belonging to somebody else are all the same 404 with the same body — telling them apart would turn the prefix into an oracle for which ids are real.").
		tag("Share links").
		returns("200", "Revoked.", sOK).
		build(), deps.share.Revoke)

	// --- Visitor side ---
	scoped := shareScope(deps.shares)
	unlock := shareUnlockLimiter.limitByShareToken(http.HandlerFunc(deps.share.Unlock))
	payload := shareReadLimiter.limitByIP(scoped(http.HandlerFunc(deps.share.Payload)))

	// The password exchange sits outside the scope: a locked link has not been
	// resolved yet, and cannot be.
	rt.route(accessShare, op("POST", "/api/shared/{token}/unlock", "Unlock a password-protected link",
		"Exchanges the link's password for access to it. Rate limited per token and per caller.\n\n"+
			"This is the one route under the prefix that runs before the token is resolved, because a locked link cannot be resolved yet.").
		tag("Share links").
		body("The password.", envelope(map[string]*huma.Schema{
			"password": {Type: "string"},
		}, "password")).
		returns("200", "Unlocked.", sOK).
		responds("401", "Wrong password.").
		responds("429", "Too many attempts.").
		build(), unlock.ServeHTTP)

	rt.route(accessShare, op("GET", "/api/shared/{token}", "Read a shared research",
		"What the link shows: the research, its sections and its documents, with everything the owner did not share removed.\n\n"+
			"**The rest of the visitor's surface lives under `/api/shared/{token}/`**, and mirrors the authenticated paths with the `/api` prefix dropped: `/api/shared/<token>/researches/R1/entries` is served by the same handler as `/api/researches/R1/entries`, with the share in the context in place of a user.\n\n"+
			"That surface is a fixed list — the research, its sections, documents, tags, cross-references and external links, plus sessions, tasks, roadmaps and the export when the link's `include` flags allow them. Anything else under the prefix, and any method other than GET, is a 404. Provenance and revision history are never on it.").
		tag("Share links").
		queryBool("visit", "Record the visit. The page sets it once per load; a client polling the payload should not.").
		returns("200", "The shared research.", envelope(map[string]*huma.Schema{
			"data":     {Type: "object"},
			"sections": {Type: "array", Items: &huma.Schema{Type: "object"}},
		})).
		responds("401", "The link is password-protected and has not been unlocked.").
		responds("404", "Unknown, revoked or expired. The three are deliberately indistinguishable.").
		responds("429", "Too many requests.").
		build(), payload.ServeHTTP)

	rt.undocumented("GET /api/shared/{token}/",
		shareReadLimiter.limitByIP(scoped(shareRead(deps))))

	// Everything else under the prefix, by method. Without these the catch-all
	// at `/` takes a POST to `/api/shared/<token>/entries` and answers 200 with
	// the SPA's index page — no write happens, but a client cannot tell a
	// refusal from a success, and the prefix stops being a sealed surface that
	// can be reasoned about by reading this file.
	//
	// 404, not 405: a read-only surface has no business enumerating the methods
	// it does not have.
	notThere := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteShareError(w, service.ErrShareUnavailable)
	})
	rt.undocumented("/api/shared/{token}", notThere)
	rt.undocumented("/api/shared/{token}/", notThere)
}

// shareRead is the sub-mux, plus the path rewrite that feeds it.
//
// The rewrite is what lets the visitor's client speak the same API as the
// owner's: `/api/shared/<token>/researches/R1/entries` is served by the handler
// registered for `/api/researches/{id}/entries`, with the token stripped and
// the share in the context.
func shareRead(deps shareDeps) http.Handler {
	sub := http.NewServeMux()

	// needs gates a route on one of the creator's include flags. It is applied
	// at registration rather than checked inside the handler, so the answer to
	// "what does this link expose" is this list and nothing else.
	needs := func(pick func(domain.ShareInclude) bool, h http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sc := auth.ShareFromContext(r.Context())
			if sc == nil || !pick(sc.Include) {
				// The same 404 an unknown route would give. A link that does not
				// include sessions should look like a build that has none.
				handlers.WriteShareError(w, service.ErrShareUnavailable)
				return
			}
			h(w, r)
		})
	}

	// Content: the research itself, always.
	sub.HandleFunc("GET /api/researches/{id}", deps.research.Get)
	sub.HandleFunc("GET /api/researches/{id}/entries", deps.research.ListAllEntries)
	sub.HandleFunc("GET /api/researches/{id}/sections/{sectionId}/entries", deps.research.ListSectionEntries)
	sub.HandleFunc("GET /api/researches/{id}/tags", deps.research.ListTags)
	sub.HandleFunc("GET /api/researches/{id}/entries/{entryId}", deps.entry.GetByResearch)
	sub.HandleFunc("GET /api/researches/{id}/entries/by-code/{code}", deps.entry.ResolveCode)
	sub.HandleFunc("GET /api/entries/{id}", deps.entry.Get)
	// Related is served by the entry-id route only. The research-scoped form
	// collides with `/entries/by-code/{code}` in the pattern matcher, and the
	// shared view always has the uuid from the entry it just fetched.
	sub.HandleFunc("GET /api/entries/{id}/related", deps.entry.GetRelated)

	// References and links: the thing that makes this a research rather than a
	// document. Cross-references leaving the shared research are stripped to
	// inert text by Access.VisibleCrossRefs, which asks the same question this
	// prefix does and gets the same no.
	sub.HandleFunc("GET /api/researches/{id}/crossrefs", deps.crossref.ListForResearch)
	sub.HandleFunc("GET /api/entries/{id}/crossrefs", deps.crossref.GetForEntry)
	sub.HandleFunc("GET /api/researches/{id}/links", deps.links.ListByResearch)
	sub.HandleFunc("GET /api/entries/{id}/links", deps.links.ListByEntry)

	// The graph, ungated at the route: sections and documents are always in a
	// link, so the route always answers. It is the handler that leaves out the
	// session, question and task nodes a link's flags do not cover — the same
	// "a route that serves an optional part must gate itself" rule the export
	// and the roadmap `ref_data` follow. `needs` cannot express "serve the
	// route but not these nodes".
	sub.HandleFunc("GET /api/researches/{id}/graph", deps.graph.Get)

	// Opt-in parts.
	sub.Handle("GET /api/researches/{id}/roadmaps",
		needs(func(i domain.ShareInclude) bool { return i.Roadmaps }, deps.roadmap.ListByResearch))
	sub.Handle("GET /api/researches/{id}/roadmaps/{roadmapId}",
		needs(func(i domain.ShareInclude) bool { return i.Roadmaps }, deps.roadmap.GetByResearch))
	sub.Handle("GET /api/roadmaps/{id}",
		needs(func(i domain.ShareInclude) bool { return i.Roadmaps }, deps.roadmap.Get))
	sub.Handle("GET /api/researches/{id}/sessions",
		needs(func(i domain.ShareInclude) bool { return i.Sessions }, deps.session.ListByResearch))
	sub.Handle("GET /api/researches/{id}/sessions/{sessionId}",
		needs(func(i domain.ShareInclude) bool { return i.Sessions }, deps.session.Get))
	sub.Handle("GET /api/researches/{id}/tasks",
		needs(func(i domain.ShareInclude) bool { return i.Tasks }, deps.task.ListByResearch))
	sub.Handle("GET /api/researches/{id}/export",
		needs(func(i domain.ShareInclude) bool { return i.Export }, deps.export.Export))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		rest := strings.TrimPrefix(r.URL.Path, "/api/shared/"+token)
		if rest == "" || rest[0] != '/' {
			handlers.WriteShareError(w, service.ErrShareUnavailable)
			return
		}

		// A path that is not already clean does not get rewritten.
		//
		// The outer mux cleans the *escaped* path, so `%2e%2e` survives it and
		// arrives here as a literal `..`; the sub-mux then answers it with a
		// redirect to `/api/researches/{id}` — out of the public prefix
		// entirely. Nothing escalates (the redirected request carries no share
		// and no bearer token), but the property this file is built on is that
		// every path under the prefix is either a mounted route or the uniform
		// share 404, and a redirect is neither.
		target := "/api" + rest
		if path.Clean(target) != strings.TrimSuffix(target, "/") && path.Clean(target) != target {
			handlers.WriteShareError(w, service.ErrShareUnavailable)
			return
		}

		r2 := r.Clone(r.Context())
		r2.URL.Path = target
		// RawPath is the escaped form and would still name the token; leaving it
		// set makes the sub-mux match against the old path.
		r2.URL.RawPath = ""
		sub.ServeHTTP(w, r2)
	})
}

// shareScope resolves the token and puts the capability in the context.
//
// A caller who gets past it is a share visitor and nothing else: no user is
// added, so every handler that decides by reading the user still finds nobody
// and still refuses.
func shareScope(shares *service.ShareService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			share, err := shares.Resolve(r.Context(), r.PathValue("token"), shareUnlockValue(r))
			if err != nil {
				handlers.WriteShareError(w, err)
				return
			}
			ctx := auth.WithShare(r.Context(), service.Capability(share))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// shareUnlockValue reads what Unlock handed back. The query parameter is there
// for the WebSocket and for a download the browser navigates to, neither of
// which can set a header.
func shareUnlockValue(r *http.Request) string {
	if v := r.Header.Get("X-Share-Unlock"); v != "" {
		return v
	}
	return r.URL.Query().Get("unlock")
}

// limitByShareToken counts attempts against the link rather than the caller.
// An address is trivially changed; the token being attacked is not.
func (l *rateLimiter) limitByShareToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(r.PathValue("token")) || !shareUnlockIPLimiter.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
