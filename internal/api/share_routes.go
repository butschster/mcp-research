package api

import (
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/butschster/mcp-research/internal/api/handlers"
	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
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
func registerShareRoutes(
	mux *http.ServeMux,
	deps shareDeps,
	wrap func(http.HandlerFunc) http.Handler,
	wrapRead func(http.HandlerFunc) http.Handler,
) {
	// --- Owner side: ordinary authenticated routes, no share involved ---
	mux.Handle("POST /api/researches/{id}/shares", wrap(deps.share.Create))
	mux.Handle("GET /api/researches/{id}/shares", wrapRead(deps.share.List))
	mux.Handle("DELETE /api/shares/{id}", wrap(deps.share.Revoke))

	// --- Visitor side ---
	scoped := shareScope(deps.shares)

	// The password exchange sits outside the scope: a locked link has not been
	// resolved yet, and cannot be.
	mux.Handle("POST /api/shared/{token}/unlock",
		shareUnlockLimiter.limitByShareToken(http.HandlerFunc(deps.share.Unlock)))

	mux.Handle("GET /api/shared/{token}",
		shareReadLimiter.limitByIP(scoped(http.HandlerFunc(deps.share.Payload))))

	mux.Handle("GET /api/shared/{token}/",
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
	mux.Handle("/api/shared/{token}", notThere)
	mux.Handle("/api/shared/{token}/", notThere)
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
