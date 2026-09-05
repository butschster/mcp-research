package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// The OpenAPI document is built from the routes, not written alongside them.
//
// It used to be written alongside them, by hand, in a 582-line file whose
// opening comment claimed it was generated. It described 24 of 99 paths and
// said nothing at all about authentication — so the one document an external
// client reads to learn this API was wrong about three quarters of it.
//
// Every route now goes through router.route. The path, the method, the path
// parameters and the credential the route demands are all read off the
// registration itself, and openapi_drift_test.go fails if a pattern reaches the
// mux any other way. What still has to be written by a person is what the
// endpoint is *for* — a summary, and the shape of what it returns.

// accessKind names the credential a route demands. It is the same set of
// wrappers NewServer has always built; naming them is what lets the document
// state, per route, what a caller has to present.
type accessKind int

const (
	// accessUnset is the zero value, and it is not a valid choice. It exists so
	// that a registration which never reached `prepare` — a typed huma
	// operation whose metadata was not stamped — cannot fall through to the
	// wrapper table and mount an unwrapped route. `serve` panics on it.
	accessUnset accessKind = iota
	// accessPublic is a route with no credential at all: health, the spec
	// itself, login, the OAuth endpoints.
	accessPublic
	// accessRead is a read scoped to the caller. With auth off there is nobody
	// in the context and no check; with auth on the credential is what puts the
	// user there, and without it the listing would not be filtered.
	accessRead
	// accessWrite is a write: a user credential when auth is on, the legacy
	// api_token when it is not.
	accessWrite
	// accessOperatorRead and accessOperatorWrite serve the routes that accept
	// either a person or whoever runs this server. The api_token is recognised
	// first and skips the user check, because the operator is not a user and is
	// in nobody's member list.
	accessOperatorRead
	accessOperatorWrite
	// accessOptional attaches a session when the request carries one and lets
	// it through when it does not.
	accessOptional
	// accessShare is a route under /api/shared/{token}/: the token in the path
	// is the whole credential, and it resolves to viewer on exactly one
	// research.
	accessShare
)

// securityFor says what a caller presents, in the vocabulary of the document's
// securitySchemes. An empty slice means the route is open; a slice containing
// an empty requirement means the credential is optional.
func (k accessKind) securityFor(authEnabled, hasAPIToken bool) []map[string][]string {
	bearer := map[string][]string{"bearerAuth": {}}
	operator := map[string][]string{"operatorToken": {}}
	switch k {
	case accessUnset, accessPublic, accessShare:
		return nil
	case accessOptional:
		return []map[string][]string{{}, bearer}
	case accessRead:
		if !authEnabled {
			return nil
		}
		return []map[string][]string{bearer}
	case accessWrite:
		if authEnabled {
			return []map[string][]string{bearer}
		}
		if hasAPIToken {
			return []map[string][]string{operator}
		}
		return nil
	case accessOperatorRead:
		// With accounts off, `wrapReadOperator` falls back to `wrapRead`, which
		// is the identity function — so these are open, api_token or not.
		// Declaring them protected told a reader of the published document that
		// the template library needs a credential on an instance where it does
		// not.
		if !authEnabled {
			return nil
		}
		return []map[string][]string{bearer, operator}
	case accessOperatorWrite:
		if authEnabled {
			return []map[string][]string{bearer, operator}
		}
		if hasAPIToken {
			// `wrapOperator` puts the api_token check in front; there is no
			// person to authenticate against.
			return []map[string][]string{operator}
		}
		return nil
	}
	return nil
}

func (k accessKind) String() string {
	switch k {
	case accessUnset:
		return "unset"
	case accessPublic:
		return "public"
	case accessRead:
		return "read"
	case accessWrite:
		return "write"
	case accessOperatorRead:
		return "operator-read"
	case accessOperatorWrite:
		return "operator-write"
	case accessOptional:
		return "optional"
	case accessShare:
		return "share"
	}
	return "unknown"
}

// registeredRoute is what the drift test compares against the document.
type registeredRoute struct {
	Method  string
	Pattern string
	Access  accessKind
}

// router registers a route once and gets three things from it: the handler on
// the mux behind the right wrapper, the operation in the OpenAPI document, and
// a record for the drift test.
type router struct {
	mux      *http.ServeMux
	api      huma.API
	wrappers map[accessKind]func(http.Handler) http.Handler
	routes   []registeredRoute

	authEnabled bool
	hasAPIToken bool
}

type routerConfig struct {
	Mux         *http.ServeMux
	BaseURL     string
	Version     string
	AuthEnabled bool
	APIToken    bool
	Wrappers    map[accessKind]func(http.Handler) http.Handler
}

func newRouter(cfg routerConfig) *router {
	r := &router{
		mux:         cfg.Mux,
		wrappers:    cfg.Wrappers,
		authEnabled: cfg.AuthEnabled,
		hasAPIToken: cfg.APIToken,
	}

	version := cfg.Version
	if version == "" {
		version = "0.0.0-dev"
	}
	doc := &huma.OpenAPI{
		OpenAPI: "3.1.0",
		Info: &huma.Info{
			Title:       "Dovod API",
			Version:     version,
			Description: apiDescription(cfg.AuthEnabled, cfg.APIToken),
		},
		Components: &huma.Components{
			// Entity schemas are generated from the domain structs themselves.
			// They are what the handlers serialise, so a field added there
			// appears in the document without anybody remembering to add it.
			Schemas:         huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer),
			SecuritySchemes: securitySchemes(),
		},
	}
	// A relative server is valid in OpenAPI 3.1 and resolves against wherever
	// the document was fetched from — which is right for a reverse proxy, a
	// tunnel, a different port, and a laptop, all without configuration. An
	// absolute one is published only when the operator has said what it is.
	if cfg.BaseURL != "" {
		doc.Servers = []*huma.Server{{URL: cfg.BaseURL, Description: "This instance"}}
	} else {
		doc.Servers = []*huma.Server{{URL: "/", Description: "Wherever this document was fetched from."}}
	}
	// Without a top-level list a renderer shows the tags alphabetically and
	// undescribed, which puts Admin first and Auth in the middle of a hundred
	// paths. This is the reading order.
	doc.Tags = []*huma.Tag{
		{Name: "Meta", Description: "What this server is and how it is configured. Answerable without a credential."},
		{Name: "Auth", Description: "Accounts, sessions and API keys."},
		{Name: "OAuth2", Description: "How an external client — ChatGPT, Claude.ai — authorises itself without anybody configuring it by hand."},
		{Name: "Research", Description: "The unit of work: its goal, its sections, and where it has got to."},
		{Name: "Documents", Description: "What the research has found. Every document carries a short code, and `[[E3]]` in any text is a link to one."},
		{Name: "Sessions", Description: "The interview loop — questions asked, answers recorded, and the documents a session produced."},
		{Name: "Tasks", Description: "The todo list an agent keeps for itself while working."},
		{Name: "Annotations", Description: "Marks a person leaves on a sentence they do not believe. Only a person creates one and only a person closes one."},
		{Name: "Revisions", Description: "Every write leaves a numbered snapshot: who wrote it, from which session, and what changed."},
		{Name: "Roadmaps", Description: "The graph of what is planned, with nodes that point at live tasks and questions."},
		{Name: "Cross-references", Description: "The `[[...]]` links between documents, and the outside URLs they cite."},
		{Name: "Teams", Description: "Who owns a research and who may do what to it. A role in the owning team is the whole of the permission model."},
		{Name: "Skills", Description: "A team's methodology, attached to the researches that work by it."},
		{Name: "Templates", Description: "Kickoff methodologies, belonging to a team or to the whole instance."},
		{Name: "Share links", Description: "Read-only capability over one research, for people with no account here."},
		{Name: "Export", Description: "Taking the work out: markdown, an Obsidian vault, or portable JSON."},
		{Name: "Admin", Description: "Maintenance a person runs once."},
	}

	config := huma.Config{
		OpenAPI: doc,
		Formats: huma.DefaultConfig("", "").Formats,
		// The spec is served by handlers registered in NewServer, on the paths
		// this product has always used. Leaving these empty stops huma from
		// claiming /openapi and /docs on its own.
		OpenAPIPath: "",
		DocsPath:    "",
	}
	config.DefaultFormat = "application/json"

	r.api = huma.NewAPI(config, &routerAdapter{r: r})
	registerSharedResponses(doc, r.schemaOf(apiError{}, "ApiError"))
	return r
}

// routerAdapter is the seam. huma hands it an operation and a handler; it puts
// the handler on the mux behind the wrapper the operation's access kind names,
// which is how a typed huma operation gets the same authentication as the
// hand-written handler next to it.
type routerAdapter struct{ r *router }

func (a *routerAdapter) Handle(op *huma.Operation, handler func(huma.Context)) {
	a.r.serve(op, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler(humago.NewContext(op, req, w))
	}))
}

func (a *routerAdapter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.r.mux.ServeHTTP(w, req)
}

const accessMetadataKey = "x-access"

// route documents an operation and serves it with an ordinary http handler.
//
// It is the path for the handlers this product already has, which write
// `map[string]any` and cannot be reflected over. When a resource is converted
// to a typed huma operation it will go through `huma.Register` against
// `r.api` instead and derive its schemas; both land in the same document and
// behind the same wrapper, because `routerAdapter` is what huma registers
// through. Such an operation must carry the access kind in its metadata —
// `serve` panics rather than mount one that does not.
func (r *router) route(access accessKind, op *huma.Operation, h http.HandlerFunc) {
	r.prepare(access, op)
	if op.Responses == nil {
		op.Responses = map[string]*huma.Response{}
	}
	if len(op.Responses) == 0 {
		op.Responses["200"] = jsonResponse("Success", nil)
	}
	r.errorResponses(op)
	r.api.OpenAPI().AddOperation(op)
	r.serve(op, h)
}

// typed prepares an operation for huma.Register, which is how a resource will
// be converted from a `map[string]any` handler to a typed one.
//
// It exists because `serve` panics on an operation with no access kind, and
// huma.Register does not populate Metadata — so without this the supported next
// step would be a crash at startup. Use it as:
//
//	huma.Register(rt.API(), *rt.typed(accessWrite, op(...).build()), handler)
func (r *router) typed(access accessKind, o *huma.Operation) *huma.Operation {
	r.prepare(access, o)
	if o.Responses == nil {
		o.Responses = map[string]*huma.Response{}
	}
	r.errorResponses(o)
	return o
}

// API is the huma API to register typed operations against.
func (r *router) API() huma.API { return r.api }

// prepare fills in everything the registration already knows: the operation id,
// the path parameters, and the credential.
func (r *router) prepare(access accessKind, op *huma.Operation) {
	if op.Metadata == nil {
		op.Metadata = map[string]any{}
	}
	op.Metadata[accessMetadataKey] = access
	if op.OperationID == "" {
		op.OperationID = operationID(op.Method, op.Path)
	}
	op.Security = access.securityFor(r.authEnabled, r.hasAPIToken)
	addPathParams(op)
}

// The refusals every route can produce, written once in components/responses
// and referenced from each operation. handlers.writeServiceError maps them in
// one place, so they genuinely are the same everywhere — and spelling them out
// 126 times made the document four times its necessary size.
var sharedErrorResponses = map[string]*huma.Response{
	"BadRequest": {
		Description: "The request could not be parsed, or a field failed validation.",
	},
	"Unauthorized": {
		Description: "No credential was presented, or the one presented is not valid. An OAuth access token that is past its hour reads the same as no credential; renew it with the `refresh_token` grant.",
	},
	"Forbidden": {
		Description: "The caller is a member of the team that owns this record but does not have the right this needs — a viewer attempting a write. Told apart from `404` on purpose: hiding a record from somebody who can already read it protects nothing.",
	},
	"NotFound": {
		Description: "No such record, **or** the caller is not a member of the team that owns it. The two are deliberately indistinguishable: confirming that a record exists is itself information.",
	},
	"ServerError": {
		Description: "The server failed to answer.",
	},
}

func registerSharedResponses(doc *huma.OpenAPI, errSchema *huma.Schema) {
	doc.Components.Responses = map[string]*huma.Response{}
	for name, resp := range sharedErrorResponses {
		doc.Components.Responses[name] = &huma.Response{
			Description: resp.Description,
			Content: map[string]*huma.MediaType{
				"application/json": {Schema: errSchema},
			},
		}
	}
}

// errorResponses attaches the refusals this route can actually produce.
func (r *router) errorResponses(op *huma.Operation) {
	access := op.Metadata[accessMetadataKey].(accessKind)
	add := func(status, name string) {
		if _, ok := op.Responses[status]; !ok {
			op.Responses[status] = &huma.Response{Ref: "#/components/responses/" + name}
		}
	}
	add("400", "BadRequest")
	if len(op.Security) > 0 && access != accessOptional {
		add("401", "Unauthorized")
	}
	switch access {
	case accessRead, accessWrite, accessOperatorRead, accessOperatorWrite:
		// A read can refuse with 403 too. Two do: listing a team's invitations
		// needs owner, and listing a research's share links goes through
		// Access.Write — a member with the wrong role is told so rather than
		// shown a 404, because hiding a record from somebody who can already
		// read it protects nothing.
		add("403", "Forbidden")
		// Not only the routes with a path parameter. The five creates that name
		// their parent in the body — entries, tasks, sessions, roadmaps and the
		// research import — answer 404 for an unknown or other-team parent,
		// which is the most common refusal on a create call and was undeclared.
		add("404", "NotFound")
	}
	add("500", "ServerError")
}

// serve puts the handler on the mux behind the right wrapper and records the
// route for the drift test.
func (r *router) serve(op *huma.Operation, h http.Handler) {
	access, ok := op.Metadata[accessMetadataKey].(accessKind)
	if !ok || access == accessUnset {
		// Failing closed here is the point. The alternative — treating a
		// missing key as the zero value — would mount the route with no
		// wrapper at all, which is the exact shape of every access bug this
		// product has had.
		panic("api: route " + op.Method + " " + op.Path + " was registered without an access kind")
	}
	if wrap, ok := r.wrappers[access]; ok && wrap != nil {
		h = wrap(h)
	}
	method := strings.ToUpper(op.Method)
	r.mux.Handle(method+" "+op.Path, h)
	r.routes = append(r.routes, registeredRoute{Method: method, Pattern: op.Path, Access: access})
}

// undocumented registers a route that has no place in an HTTP API document:
// the WebSocket upgrade, the MCP transport, the embedded frontend. It still
// goes through the router so the drift test knows the route exists and that
// somebody decided to leave it out.
func (r *router) undocumented(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
	method, path, found := strings.Cut(pattern, " ")
	if !found {
		method, path = "", pattern
	}
	r.routes = append(r.routes, registeredRoute{Method: method, Pattern: path, Access: accessPublic})
}

// Routes is what the drift test reads.
func (r *router) Routes() []registeredRoute {
	out := append([]registeredRoute(nil), r.routes...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pattern != out[j].Pattern {
			return out[i].Pattern < out[j].Pattern
		}
		return out[i].Method < out[j].Method
	})
	return out
}

// schemaOf generates the JSON schema of a Go type and registers it under
// components/schemas, returning a $ref to it.
//
// The types passed to it are the domain structs the handlers actually
// serialise. That is the whole point: a field added to domain.Entry appears in
// the document because it appears in the JSON, not because somebody remembered
// to describe it twice.
func (r *router) schemaOf(v any, hint string) *huma.Schema {
	return r.api.OpenAPI().Components.Schemas.Schema(reflect.TypeOf(v), true, hint)
}

// envelope wraps an entity schema in the response shape this API uses:
// `{"data": <entity>}`, sometimes with siblings alongside it.
func envelope(fields map[string]*huma.Schema, required ...string) *huma.Schema {
	return &huma.Schema{Type: "object", Properties: fields, Required: required}
}

// listOf is an array of the given schema.
func listOf(item *huma.Schema) *huma.Schema {
	return &huma.Schema{Type: "array", Items: item}
}

// --- OpenAPI helpers ---

var pathParamRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

func pathParamNames(path string) []string {
	var out []string
	for _, m := range pathParamRe.FindAllStringSubmatch(path, -1) {
		out = append(out, m[1])
	}
	return out
}

// addPathParams declares every {name} in the pattern. Deriving them is the
// point: a path parameter that is in the route and not in the document is the
// most common way a hand-written spec goes wrong.
func addPathParams(op *huma.Operation) {
	declared := map[string]bool{}
	for _, p := range op.Parameters {
		if p.In == "path" {
			declared[p.Name] = true
		}
	}
	for _, name := range pathParamNames(op.Path) {
		if declared[name] {
			continue
		}
		op.Parameters = append(op.Parameters, &huma.Param{
			Name:        name,
			In:          "path",
			Required:    true,
			Description: pathParamDescription(op.Path, name),
			Schema:      &huma.Schema{Type: "string"},
		})
	}
}

// pathParamDescription explains the ones whose name does not.
//
// It takes the path as well as the name, because `{id}` means something
// different on almost every route and the answer to "does a short code work
// here?" is not the same for all of them. Saying `R1` works everywhere — which
// one blanket sentence did — was wrong on 44 operations: codes resolve on
// `/api/researches/{id}/…` and on the nested `{sessionId}` and `{entryId}`, and
// nowhere else. `GET /api/entries/E1` is a 404, and a reader who believed
// otherwise went looking for a permissions problem.
func pathParamDescription(path, name string) string {
	switch name {
	case "id":
		switch {
		case strings.HasPrefix(path, "/api/researches/"):
			return "Research id (UUID) or short code, e.g. `R1`."
		case strings.HasPrefix(path, "/api/entries/"):
			return "Document id (UUID). A short code such as `E3` does **not** resolve here — use `/api/researches/{id}/entries/by-code/{code}`."
		case strings.HasPrefix(path, "/api/sessions/"):
			return "Session id (UUID). A short code such as `SS1` resolves only under `/api/researches/{id}/sessions/{sessionId}`."
		case strings.HasPrefix(path, "/api/teams/"):
			return "Team id (UUID)."
		case strings.HasPrefix(path, "/api/invites/"):
			return "Invitation id (UUID), as listed by GET /api/teams/{id}/invites."
		}
		return "Record id (UUID)."
	case "sectionId":
		return "Section id (UUID)."
	case "sessionId":
		return "Session id (UUID) or short code, e.g. `SS1`."
	case "entryId":
		return "Document id (UUID) or short code, e.g. `E3` — codes resolve here because the research that scopes them is in the path."
	case "roadmapId":
		return "Roadmap id (UUID)."
	case "nodeId":
		return "Roadmap node id (UUID)."
	case "questionId":
		return "Question id (UUID)."
	case "userId":
		return "User id (UUID), as listed by GET /api/teams/{id}/members."
	case "skillId":
		return "Skill id (UUID). Skills are managed by id because a built-in and a team's fork of it share a slug on purpose."
	case "templateId":
		return "Template id (UUID). Read by slug, write by id."
	case "slug":
		return "The slug, resolved in the caller's scope: a team's own first, then the server-wide library."
	case "code":
		return "Short code, e.g. `E3`."
	case "revision":
		return "Revision number, counting from 1."
	case "token":
		if strings.HasPrefix(path, "/api/invites/") {
			return "The invitation token from the link the person was sent. Not an id, and not a share token."
		}
		return "The share link token, as issued once by POST /api/researches/{id}/shares."
	}
	return ""
}

func operationID(method, path string) string {
	parts := []string{strings.ToLower(method)}
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, "{") {
			seg = "by-" + strings.Trim(seg, "{}")
		}
		parts = append(parts, seg)
	}
	id := strings.Join(parts, "-")
	id = strings.ReplaceAll(id, ".", "-")
	return id
}

func jsonResponse(description string, schema *huma.Schema) *huma.Response {
	resp := &huma.Response{Description: description}
	if schema != nil {
		resp.Content = map[string]*huma.MediaType{
			"application/json": {Schema: schema},
		}
	}
	return resp
}

func fileResponse(description, contentType string) *huma.Response {
	return &huma.Response{
		Description: description,
		Content: map[string]*huma.MediaType{
			contentType: {Schema: &huma.Schema{Type: "string", Format: "binary"}},
		},
	}
}

// errorSchemaRef is where the refusal body lives in the document. It is a
// constant because the operation builder runs before the router sees the
// operation, and a $ref is just a string.
//
// The name is what huma's DefaultSchemaNamer produces for the Go type below,
// which is why it is `ApiError` and not `Error` — a mismatch here is a $ref
// that resolves to nothing, and openapi_spec_test.go checks every one.
const errorSchemaRef = "#/components/schemas/ApiError"

// apiError is the body every refusal carries. Declaring it as a type rather
// than a literal is what lets it be generated once and referenced.
type apiError struct {
	Error string `json:"error" doc:"What went wrong, in one sentence."`
}

func securitySchemes() map[string]*huma.SecurityScheme {
	return map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:        "http",
			Scheme:      "bearer",
			Description: bearerDescription,
		},
		"operatorToken": {
			Type:        "http",
			Scheme:      "bearer",
			Description: operatorDescription,
		},
	}
}

const bearerDescription = `A credential belonging to a person, as ` + "`Authorization: Bearer <token>`" + `.

Three kinds are accepted and they are interchangeable on every route:

- a **JWT** from ` + "`POST /api/auth/login`" + ` or ` + "`POST /api/auth/register`" + ` — a browser session, which is also what marks a document edit as made by a human rather than an agent;
- an **API key** from ` + "`POST /api/auth/api-keys`" + `, shown once at creation — the credential for a long-lived MCP or script client;
- an **OAuth2 access token** from ` + "`POST /auth/token`" + `, valid for one hour and renewed with the ` + "`refresh_token`" + ` grant.

The WebSocket at ` + "`/ws`" + ` takes the same credential, as a header or as ` + "`?token=`" + `, because a browser cannot set headers on a WebSocket handshake.`

const operatorDescription = `The instance ` + "`api_token`" + ` from the server's configuration, as ` + "`Authorization: Bearer <token>`" + `.

It identifies whoever runs this server, not a person: it belongs to no team and appears in no member list. It is what enables the write API when multi-user auth is off, and on the template routes it is recognised ahead of the user check so that a server-wide template can be written and read back.`

func apiDescription(authEnabled, hasAPIToken bool) string {
	var b strings.Builder
	b.WriteString("REST API of the Dovod server.\n\n")
	b.WriteString("**Short codes work in a path, not in a body.** `/api/researches/R1` and `/api/researches/<uuid>` resolve the same record, and so do the nested `{sessionId}` and `{entryId}`. Everywhere else — `/api/entries/{id}`, and every `research_id` or `section_id` in a request body — a UUID is required, and a code is answered with `404`. Each parameter says which it takes.\n\n")
	switch {
	case authEnabled:
		b.WriteString("This instance has **multi-user authentication on**. Every route below except the ones marked otherwise needs a bearer credential, and what it returns is scoped to the teams that credential belongs to.\n")
	case hasAPIToken:
		b.WriteString("This instance has **no user accounts and a configured `api_token`**. Reads are open; writes need that token as a bearer credential.\n")
	default:
		b.WriteString("This instance has **no authentication configured**: it is a local single-user run, reads and writes are both open, and the write API is disabled unless an `api_token` is set.\n")
	}
	b.WriteString("\nErrors are `{\"error\": \"...\"}` with the status code carrying the meaning. A `404` on a record that exists but belongs to a team the caller is not in is deliberate: confirming it exists is itself information.\n")
	// The product's primary integration surface is not an HTTP operation and so
	// is not in `paths` — which meant a reader who arrived at this document
	// learned 126 REST routes and never learned that MCP exists at all, and
	// integrated over REST because that is what they were shown.
	b.WriteString("\n**This server also speaks MCP**, at `/mcp`, with the same bearer credential — that is what a Claude Code, Cursor or ChatGPT client connects to, and it is the shorter path if your client supports it. MCP is not described here: it is a protocol rather than a set of HTTP operations, and its tools are documented at `/llms.txt`. Use this REST API when you are writing a script, a webhook or an integration of your own.\n")
	return b.String()
}

// specHandlers serves the generated document. Both extensions are the same
// document; a client that cannot read YAML asks for the JSON.
func (r *router) specHandlers() (yamlHandler, jsonHandler http.HandlerFunc) {
	// The document is a build artefact of a running server: every route is
	// registered before the first request and nothing changes it afterwards. So
	// it is marshalled once rather than on every hit, and served with an ETag —
	// the reference page fetches it on each load, and a browser that already has
	// it should be told so.
	type document struct {
		body []byte
		etag string
	}
	var (
		once           sync.Once
		asYAML, asJSON document
		marshalErr     error
	)
	build := func() {
		asJSON.body, marshalErr = r.api.OpenAPI().MarshalJSON()
		if marshalErr != nil {
			return
		}
		asYAML.body, marshalErr = r.api.OpenAPI().YAML()
		if marshalErr != nil {
			return
		}
		// One digest per representation. Deriving both from the JSON gave the
		// two URLs the same validator, so a script that stored the ETag from one
		// and then asked for the other got a 304 with no body — and either
		// failed or quietly read JSON as YAML.
		asJSON.etag = etagOf(asJSON.body)
		asYAML.etag = etagOf(asYAML.body)
	}

	serve := func(contentType string, pick func() *document) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			once.Do(build)
			if marshalErr != nil {
				http.Error(w, fmt.Sprintf("openapi: %v", marshalErr), http.StatusInternalServerError)
				return
			}
			doc := pick()
			w.Header().Set("ETag", doc.etag)
			w.Header().Set("Cache-Control", "no-cache")
			if match := req.Header.Get("If-None-Match"); match == doc.etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(doc.body)
		}
	}

	return serve("application/yaml; charset=utf-8", func() *document { return &asYAML }),
		serve("application/json; charset=utf-8", func() *document { return &asJSON })
}

func etagOf(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

// --- operation builder ---
//
// A registration has to say four things a reader cannot derive: what the
// endpoint is called, what it is for, which query parameters it honours, and
// the shape of what comes back. Everything else — the path parameters, the
// credential, the refusals — is filled in by the router.

type opBuilder struct{ op *huma.Operation }

func op(method, path, summary, description string) *opBuilder {
	return &opBuilder{op: &huma.Operation{
		Method:      method,
		Path:        path,
		Summary:     summary,
		Description: description,
	}}
}

func (b *opBuilder) tag(t string) *opBuilder {
	b.op.Tags = append(b.op.Tags, t)
	return b
}

// query declares an optional query parameter.
func (b *opBuilder) query(name, description string) *opBuilder {
	b.op.Parameters = append(b.op.Parameters, &huma.Param{
		Name: name, In: "query", Description: description,
		Schema: &huma.Schema{Type: "string"},
	})
	return b
}

// queryBool declares an optional boolean query parameter.
func (b *opBuilder) queryBool(name, description string) *opBuilder {
	b.op.Parameters = append(b.op.Parameters, &huma.Param{
		Name: name, In: "query", Description: description,
		Schema: &huma.Schema{Type: "boolean"},
	})
	return b
}

// queryInt declares an optional integer query parameter.
func (b *opBuilder) queryInt(name, description string) *opBuilder {
	b.op.Parameters = append(b.op.Parameters, &huma.Param{
		Name: name, In: "query", Description: description,
		Schema: &huma.Schema{Type: "integer"},
	})
	return b
}

// body declares the request body.
func (b *opBuilder) body(description string, schema *huma.Schema) *opBuilder {
	b.op.RequestBody = &huma.RequestBody{
		Description: description,
		Required:    true,
		Content: map[string]*huma.MediaType{
			"application/json": {Schema: schema},
		},
	}
	return b
}

// bodyForm declares a body accepted as JSON *or* as a form. Most OAuth2 client
// libraries send a form, and declaring only JSON told them to do the one thing
// half of them cannot.
func (b *opBuilder) bodyForm(description string, schema *huma.Schema) *opBuilder {
	b.body(description, schema)
	b.op.RequestBody.Content["application/x-www-form-urlencoded"] = &huma.MediaType{Schema: schema}
	return b
}

// returns declares the success response.
func (b *opBuilder) returns(status, description string, schema *huma.Schema) *opBuilder {
	if b.op.Responses == nil {
		b.op.Responses = map[string]*huma.Response{}
	}
	b.op.Responses[status] = jsonResponse(description, schema)
	return b
}

// returnsFile declares a success response that is a download rather than JSON.
func (b *opBuilder) returnsFile(description, contentType string) *opBuilder {
	if b.op.Responses == nil {
		b.op.Responses = map[string]*huma.Response{}
	}
	b.op.Responses["200"] = fileResponse(description, contentType)
	return b
}

// respondsEmpty declares a status that carries no body — a 304, say, which is
// not a refusal and has nothing to describe.
func (b *opBuilder) respondsEmpty(status, description string) *opBuilder {
	if b.op.Responses == nil {
		b.op.Responses = map[string]*huma.Response{}
	}
	b.op.Responses[status] = &huma.Response{Description: description}
	return b
}

// responds declares an additional status this endpoint can answer with, one
// that is part of its contract rather than a generic refusal.
func (b *opBuilder) responds(status, description string) *opBuilder {
	if b.op.Responses == nil {
		b.op.Responses = map[string]*huma.Response{}
	}
	b.op.Responses[status] = jsonResponse(description, &huma.Schema{Ref: errorSchemaRef})
	return b
}

func (b *opBuilder) build() *huma.Operation { return b.op }

func ptrInt(v int) *int { return &v }
