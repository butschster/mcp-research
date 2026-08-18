package auth

import "context"

type operatorKey struct{}

// OperatorKind is *how* a request came to count as the operator, and the two
// ways are not interchangeable.
type OperatorKind int

const (
	// OperatorNone is every ordinary caller.
	OperatorNone OperatorKind = iota
	// OperatorByToken means the instance api_token was actually presented. It
	// proves who runs the server and nothing else — in particular it proves no
	// membership of any team, which is why a caller holding it reads the global
	// library and not other people's.
	OperatorByToken
	// OperatorNoBoundary is a local run with neither api_token nor accounts.
	// There is nothing to prove anything across, so every caller is the
	// operator — the same rule the rest of the product follows when there is
	// nobody in the context. It must stay distinguishable from the token,
	// because in that mode everything on the instance is visible and narrowing
	// the scope would hide a local user's own templates from them.
	OperatorNoBoundary
)

// WithOperator marks the request as coming from whoever runs this server.
//
// It is not a role and it is not a user: nobody can be promoted to it, no team
// grants it, and it never appears in a member list. The only thing that proves
// it is the instance's `api_token` — the credential from the config file, which
// by definition only the operator has. On a machine with neither `api_token` nor
// `auth_enabled` there is no boundary to prove anything across, and the HTTP
// layer stamps it for every caller, exactly as the rest of the product treats
// that mode.
//
// It exists for one act: writing a template every team on the instance will see.
// A team publishing to the whole server was refused deliberately — a template
// body steers a model, and lending one team's instructions to another team's
// kickoff needs a trust story this product does not have. The operator is
// outside that argument, because they already own the binary and the database.
func WithOperator(ctx context.Context, kind OperatorKind) context.Context {
	return context.WithValue(ctx, operatorKey{}, kind)
}

// OperatorFromContext says how this request qualifies, for the one caller that
// needs to tell the two apart.
func OperatorFromContext(ctx context.Context) OperatorKind {
	kind, _ := ctx.Value(operatorKey{}).(OperatorKind)
	return kind
}

// IsOperator answers whether this request may do an operator's work.
//
// A share is never the operator, and it cannot become one: the marker is only
// ever set by the HTTP middleware, from a header a share visitor does not carry.
func IsOperator(ctx context.Context) bool {
	return OperatorFromContext(ctx) != OperatorNone
}
