package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Two properties this package now depends on, neither of which any other test
// could see.

// TestRouter_NoBareMuxRegistration is the guard CLAUDE.md promises.
//
// It used to be promised by the drift test, which cannot deliver it: that test
// walks `router.Routes()`, a slice only the router appends to, so a route
// registered straight on the mux is invisible to it — exactly the route that
// would be. This reads the source instead.
//
// The historic bug it guards against is a scoped read registered bare: without
// a wrapper nobody is put into the request context, `ResearchService.List`
// leaves `filter.UserID` nil, and it returns every user's researches.
func TestRouter_NoBareMuxRegistration(t *testing.T) {
	// router.go is where the one legitimate mux.Handle lives.
	const allowed = "router.go"

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var offenders []string

	for _, name := range files {
		if name == allowed || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name != "Handle" && sel.Sel.Name != "HandleFunc" {
				return true
			}
			recv, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			// `sub.HandleFunc` inside share_routes.go is the share visitor
			// sub-mux: a private mux mounted behind shareScope, not a route on
			// the server's own mux.
			if recv.Name == "sub" {
				return true
			}
			if recv.Name == "mux" {
				pos := fset.Position(call.Pos())
				offenders = append(offenders,
					pos.Filename+":"+itoa(pos.Line)+" — "+recv.Name+"."+sel.Sel.Name)
			}
			return true
		})
	}

	if len(offenders) > 0 {
		t.Fatalf("routes registered straight on the mux, bypassing the router and its wrapper:\n  %s\n\n"+
			"Register through rt.route(<accessKind>, op(...).build(), handler), or rt.undocumented(...) "+
			"for something that is not an HTTP operation. A scoped read registered bare puts nobody in "+
			"the request context and returns every user's data.",
			strings.Join(offenders, "\n  "))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestRouter_EveryScopedRouteRefusesAnonymously is the security property of the
// whole registration refactor, and nothing else asserts it.
//
// The document says, per route, which credential it wants. That claim is worth
// nothing unless the wrapper the access kind selects is actually applied. So:
// with accounts on, fire every registered route with no credential at all and
// require a refusal from every kind that is not deliberately open.
func TestRouter_EveryScopedRouteRefusesAnonymously(t *testing.T) {
	s := newSpecServer(t) // auth_enabled: true, api_token set

	var checked int
	for _, route := range s.router.Routes() {
		switch route.Access {
		case accessPublic, accessOptional, accessShare:
			// Open by decision, and each is covered by its own test.
			continue
		}
		if route.Method == "" {
			continue // a prefix mount, not an operation
		}

		// A concrete path that cannot exist. Reaching the handler at all would
		// answer 404; being refused first answers 401, which is the point —
		// authentication happens before the record is looked for.
		path := route.Pattern
		for _, name := range pathParamNames(route.Pattern) {
			path = strings.Replace(path, "{"+name+"}", "does-not-exist", 1)
		}

		r := httptest.NewRequest(route.Method, path, strings.NewReader("{}"))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s (%s) answered %d to a request with no credential, want 401\n  body: %s",
				route.Method, route.Pattern, route.Access, w.Code, truncate(w.Body.String(), 160))
		}
		checked++
	}

	if checked < 100 {
		t.Fatalf("only %d scoped routes were checked; the route table did not come through", checked)
	}
	t.Logf("%d scoped routes refuse an anonymous caller", checked)
}

// TestRouter_AccessKindMatchesTheDocument closes the loop the other way: what
// the document publishes about a route has to be what the router applied.
func TestRouter_AccessKindMatchesTheDocument(t *testing.T) {
	s := newSpecServer(t)
	doc := specOf(t, nil, s.mux)
	paths, _ := doc["paths"].(map[string]any)

	for _, route := range s.router.Routes() {
		item, ok := paths[route.Pattern].(map[string]any)
		if !ok {
			continue // undocumented on purpose; the drift test owns that list
		}
		opDoc, ok := item[strings.ToLower(route.Method)].(map[string]any)
		if !ok {
			continue
		}
		security, _ := opDoc["security"].([]any)
		wantOpen := len(route.Access.securityFor(true, true)) == 0

		if wantOpen && len(security) != 0 {
			t.Errorf("%s %s is %s — served without a wrapper — but the document declares %v",
				route.Method, route.Pattern, route.Access, security)
		}
		if !wantOpen && len(security) == 0 {
			t.Errorf("%s %s is %s — served behind a wrapper — but the document declares it open",
				route.Method, route.Pattern, route.Access)
		}
	}
}
